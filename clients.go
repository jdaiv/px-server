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
	clients[client] = true
	return client
}

func RemoveClient(client *Client) {
	if client.CurrentRoom != nil {
		RemoveClientFromAllRooms(client)
	}
	if client.Authenticated {
		delete(authenticatedClients, client.User.Name)
	}
	delete(clients, client)
}

func (c *Client) Logout() error {
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

	c.User = user
	c.Authenticated = true

	authenticatedClients[username] = c
	return nil
}
