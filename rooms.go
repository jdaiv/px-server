package main

import (
	"log"

	"github.com/google/uuid"
)

var rooms = make(map[string]*Room)

type Permissions struct {
	Read          bool
	Write         bool
	TakeOwnership bool
}

type Room struct {
	Permissions     *Permissions
	Name            string
	FriendlyName    string
	Clients         map[*Client]int
	ClientsEnd      int
	Owner           string
	CurrentActivity int
}

func NewRoom(name string) (*Room, error) {
	if _, exists := rooms[name]; exists {
		return nil, ErrorRoomExists
	}
	r := &Room{
		Permissions:  &Permissions{true, true, true},
		Name:         name,
		FriendlyName: name,
		Clients:      make(map[*Client]int, 0),
		ClientsEnd:   1,
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
	for c := range room.Clients {
		// if !c.Authenticated {
		// 	continue
		// }
		// log.Printf("[chat] sending to %s", c.User.NameNormal)
		err := c.Conn.WriteJSON(WSResponse{
			Error:  0,
			Action: WSAction{scope, action, name},
			Data:   data,
		})
		if err != nil {
			log.Printf("[chat] error sending to %s, removing from room (%v)", c.User.NameNormal, err)
			room.RemoveClient(c)
		}
	}
	return nil
}

func RemoveClientFromAllRooms(c *Client) {
	for _, room := range rooms {
		room.RemoveClient(c)
	}
}

func (r *Room) AssignOwnership(c *Client) {
	if c == nil {
		r.Owner = ""
		log.Printf("[chat/%s] cleared ownership", r.Name)
		return
	}
	if !c.Authenticated {
		log.Printf("[chat/%s] tried to assign ownership to unauthenticated client", r.Name)
		return
	}
	r.Owner = c.User.NameNormal
	log.Printf("[chat/%s] assigned new owner: %s", r.Name, r.Owner)
}

func (r *Room) AddClient(c *Client) error {
	// if !c.Authenticated {
	// 	return ErrorUnauthenticated
	// }
	if r.Owner == "" && c.Authenticated && r.Permissions.TakeOwnership {
		r.AssignOwnership(c)
	}
	r.Clients[c] = len(r.Clients)
	r.ClientsEnd++
	return nil
}

func (r *Room) GetFirstClient() (client *Client) {
	lowest := r.ClientsEnd
	for c, i := range r.Clients {
		if i < lowest && c.Authenticated {
			client = c
			lowest = i
		}
	}
	return
}

func (r *Room) RemoveClient(c *Client) {
	username := c.User.NameNormal

	delete(r.Clients, c)

	if username == r.Owner && r.Permissions.TakeOwnership {
		r.AssignOwnership(r.GetFirstClient())
	}
}
