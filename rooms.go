package main

import (
	"log"

	"github.com/google/uuid"
)

var rooms = make(map[string]*Room)

type Permissions struct {
	Read  bool
	Write bool
}

type Room struct {
	Permissions     *Permissions
	Name            string
	FriendlyName    string
	Clients         map[string]*Client
	Owner           string
	CurrentActivity int
}

func AddDefaultRooms() {
	sr, _ := NewRoom("system")
	sr.Permissions.Write = false
	NewRoom("public")
}

func NewRoom(name string) (*Room, error) {
	if _, exists := rooms[name]; exists {
		return nil, ErrorRoomExists
	}
	r := &Room{
		Permissions:  new(Permissions),
		Name:         name,
		FriendlyName: name,
		Clients:      make(map[string]*Client),
	}
	rooms[name] = r
	return r, nil
}

func NewRoomRandom() (*Room, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return NewRoom(id.String())
}

func DeleteRoom(name string) {
	if _, exists := rooms[name]; exists {
		delete(rooms, name)
	}
}

func JoinRoom(name string, client *Client) error {
	room, exists := rooms[name]
	if !exists {
		return ErrorRoomMissing
	}
	return room.AddClient(client)
}

func BroadcastToRoom(name, scope, action string, data interface{}) error {
	room, exists := rooms[name]
	if !exists {
		return ErrorRoomMissing
	}
	for _, c := range room.Clients {
		if !c.Authenticated {
			continue
		}
		// log.Printf("[chat] sending to %s", c.User.Name)
		err := c.Conn.WriteJSON(WSResponse{
			Error:   0,
			Message: "success",
			Scope:   scope,
			Action:  action,
			Data:    data,
		})
		if err != nil {
			log.Printf("[chat] error sending to %s, removing from room (%v)", c.User.Name, err)
			room.RemoveClient(c)
		}
	}
	return nil
}

func (r *Room) AddClient(c *Client) error {
	if !c.Authenticated {
		return ErrorUnauthenticated
	}
	// if _, exists := r.Clients[c.User.Name]; exists {
	// 	return AlreadyInRoom
	// }
	r.Clients[c.User.Name] = c
	return nil
}

func (r *Room) RemoveClient(c *Client) {
	if _, exists := r.Clients[c.User.Name]; exists {
		delete(r.Clients, c.User.Name)
	}
}
