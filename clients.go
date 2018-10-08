package main

import (
	"fmt"
	"log"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
)

type Client struct {
	Authenticated bool
	SuperUser     bool
	User          User
	Conn          *websocket.Conn
	CurrentRoom   *Room
}

var clients = make(map[*Client]bool)
var authenticatedClients = make(map[string]*Client)

func MakeClient(conn *websocket.Conn) *Client {
	client := &Client{
		Conn: conn,
	}
	client.MakeAnon()
	clients[client] = true
	return client
}

func RemoveClient(client *Client) {
	RemoveClientFromAllRooms(client)
	// this is before broadcasting a user left so we don't enter an infinite loop
	delete(clients, client)
	if client.Authenticated {
		delete(authenticatedClients, client.User.NameNormal)
		BroadcastToAll("public", "chat", "new_message", messageSend{
			Content: fmt.Sprintf("%s left", client.User.Name),
			From:    "server",
			Class:   MESSAGE_CLASS_SERVER,
		})
	}
}

func BroadcastToAll(name, scope, action string, data interface{}) error {
	for c := range clients {
		err := c.Conn.WriteJSON(WSResponse{
			Error:  0,
			Action: WSAction{scope, action, name},
			Data:   data,
		})
		if err != nil {
			RemoveClient(c)
		}
	}
	return nil
}

func (c *Client) MakeAnon() {
	if c.CurrentRoom != nil {
		c.CurrentRoom.RemoveClient(c)
	}
	c.Authenticated = false
	anonName := fmt.Sprintf("anon (%s)", c.Conn.RemoteAddr())
	c.User = User{
		Name:       anonName,
		NameNormal: anonName,
	}
	if c.CurrentRoom != nil {
		c.CurrentRoom.AddClient(c)
	}
}

func (c *Client) Logout() error {
	if !c.Authenticated {
		// nothing to do
		return nil
	}
	BroadcastToAll("public", "chat", "new_message", messageSend{
		Content: fmt.Sprintf("%s logged out", c.User.Name),
		From:    "server",
		Class:   MESSAGE_CLASS_SERVER,
	})
	delete(authenticatedClients, c.User.NameNormal)
	c.MakeAnon()
	err := c.Conn.WriteJSON(WSResponse{
		Error:  0,
		Action: WSAction{"auth", "logout", "all"},
	})
	return err
}

func (c *Client) Authenticate(tokenStr string) error {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return JWTSecret, nil
	})

	if err != nil {
		return ErrorInvalidToken
	}

	var username string
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username = claims["name"].(string)
	} else {
		return ErrorInvalidToken
	}

	// check if user is already connected
	if client, exists := authenticatedClients[username]; exists {
		if err := client.Logout(); err != nil {
			log.Printf("[ws/auth] attempted to logout user but an error occurred (%v)", err)
		}
	}

	user, err := LoadUser(username)
	if err != nil {
		return ErrorUserMissing
	}

	if c.CurrentRoom != nil {
		c.CurrentRoom.RemoveClient(c)
	}

	c.User = user
	c.Authenticated = true

	authenticatedClients[username] = c

	if c.CurrentRoom != nil {
		c.CurrentRoom.AddClient(c)
		c.CurrentRoom.Broadcast("chat", "new_message", messageSend{
			Content: fmt.Sprintf("%s logged in", c.User.Name),
			From:    "server",
			Class:   MESSAGE_CLASS_SERVER,
		})
	}

	BroadcastToAll("public", "chat", "new_message", messageSend{
		Content: fmt.Sprintf("%s logged in", c.User.Name),
		From:    "server",
		Class:   MESSAGE_CLASS_SERVER,
	})

	return nil
}
