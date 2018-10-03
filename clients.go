package main

import "github.com/gorilla/websocket"

type Client struct {
	Authenticated bool
	SuperUser     bool
	User          User
	Conn          *websocket.Conn
	CurrentRoom   *Room
}

var clients = make(map[*Client]bool)

func MakeClient(conn *websocket.Conn) *Client {
	client := &Client{
		Conn: conn,
	}
	clients[client] = true
	return client
}

func RemoveClient(client *Client) {
	if client.CurrentRoom != nil {
		client.CurrentRoom.RemoveClient(client)
	}
	delete(clients, client)
}
