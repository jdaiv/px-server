package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"bitbucket.org/panicexpress/backend/rpg"

	"github.com/gorilla/websocket"
)

const (
	AWAITING_AUTH = iota
	READY
	CLOSING
	TIMED_OUT
	CLOSED
)

type Client struct {
	State         int
	Authenticated bool
	User          User
	Conn          *websocket.Conn
	LastPing      int64
}

var clientsMutex = &sync.Mutex{}
var clients = make(map[*Client]bool)
var authenticatedClients = make(map[int]*Client)

func MakeClient(conn *websocket.Conn) *Client {
	client := &Client{
		Conn:     conn,
		State:    AWAITING_AUTH,
		LastPing: time.Now().UnixNano(),
	}
	clientsMutex.Lock()
	clients[client] = true
	clientsMutex.Unlock()
	return client
}

func RemoveClient(client *Client) {
	client.State = CLOSING
}

func BroadcastToAll(action ActionStr, data interface{}) {
	clientsMutex.Lock()
	for c := range clients {
		c.Write(WSResponse{
			Error:  0,
			Action: action,
			Data:   data,
		})
	}
	clientsMutex.Unlock()
}

func BroadcastToAllNoLock(action ActionStr, data interface{}) {
	for c := range clients {
		c.Write(WSResponse{
			Error:  0,
			Action: action,
			Data:   data,
		})
	}
}

func ClientMaintenace() {
	for {
		clientsMutex.Lock()

		// check if clients have timed out
		now := time.Now().UnixNano()
		for c := range clients {
			if now-c.LastPing > 10*1e+9 {
				c.State = TIMED_OUT
			}
		}

		// find closed or timed out clients
		toRemove := make([]*Client, 0)
		for c := range clients {
			if c.State >= CLOSING {
				toRemove = append(toRemove, c)
			}
		}

		// and remove
		for _, c := range toRemove {
			delete(clients, c)
			if c.Authenticated {
				game.Incoming <- rpg.IncomingMessage{
					PlayerId: c.User.Id,
					Data: rpg.IncomingMessageData{
						Type: rpg.ACTION_LEAVE,
					},
				}
				delete(authenticatedClients, c.User.Id)
				message := "%s disconnected"
				if c.State == TIMED_OUT {
					message = "%s timed out"
				}
				BroadcastToAllNoLock(ACTION_CHAT_MESSAGE, messageSend{
					Content: fmt.Sprintf(message, c.User.Name),
					From:    "server",
					Class:   MESSAGE_CLASS_SERVER,
				})

				log.Printf("[ws] "+message, c.User.Name)
			}
		}

		clientsMutex.Unlock()

		// sleep
		time.Sleep(100 * time.Millisecond)
	}
}

func (c *Client) Write(data interface{}) {
	if c.State < CLOSING {
		outgoing <- outgoingMessage{Data: data, Dest: c}
	}
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

	c.User = user
	c.Authenticated = true
	c.State = READY

	authenticatedClients[user.Id] = c

	game.Incoming <- rpg.IncomingMessage{
		PlayerId: user.Id,
		Data: rpg.IncomingMessageData{
			Type: rpg.ACTION_JOIN,
		},
	}

	BroadcastToAll(ACTION_CHAT_MESSAGE, messageSend{
		Content: fmt.Sprintf("%s logged in", c.User.Name),
		From:    "server",
		Class:   MESSAGE_CLASS_SERVER,
	})

	return nil
}
