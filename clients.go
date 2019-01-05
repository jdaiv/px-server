package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

const (
	AWAITING_AUTH = iota
	AUTHENTICATED
	CLOSING
)

type Client struct {
	State       int
	SuperUser   bool
	User        User
	Conn        *websocket.Conn
	CurrentRoom *Room
}

var clients = make(map[*Client]bool)
var authenticatedClients = make(map[string]*Client)

func MakeClient(conn *websocket.Conn) *Client {
	client := &Client{
		Conn:  conn,
		State: AWAITING_AUTH,
	}
	clients[client] = true
	return client
}

func RemoveClient(client *Client) {
	RemoveClientFromAllRooms(client)
	// this is before broadcasting a user left so we don't enter an infinite loop
	delete(clients, client)
	if client.State == AUTHENTICATED {
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

func (c *Client) Authenticate(password string) error {
	// check if user is already connected
	// if client, exists := authenticatedClients[username]; exists {
	// 	if err := client.Logout(); err != nil {
	// 		log.Printf("[ws/auth] attempted to logout user but an error occurred (%v)", err)
	// 	}
	// }

	user, err := AuthenticateUser(password)
	if err != nil {
		return err
	}

	if c.CurrentRoom != nil {
		c.CurrentRoom.RemoveClient(c)
	}

	c.User = user
	c.State = AUTHENTICATED

	authenticatedClients[user.NameNormal] = c

	BroadcastToAll("public", "chat", "new_message", messageSend{
		Content: fmt.Sprintf("%s logged in", c.User.Name),
		From:    "server",
		Class:   MESSAGE_CLASS_SERVER,
	})

	return nil
}
