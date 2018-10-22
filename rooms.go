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

// TODO: convert to interface?
type Room struct {
	Permissions  *Permissions
	Name         string
	FriendlyName string
	Clients      map[*Client]int
	ClientsEnd   int
	Owner        string
	State        *RoomState
	Activity     Activity
}

type RoomPublicData struct {
	Owner         string      `json:"owner"`
	Name          string      `json:"name"`
	FriendlyName  string      `json:"friendly_name"`
	State         *RoomState  `json:"state"`
	Activity      string      `json:"activity"`
	ActivityState interface{} `json:"activity_state"`
}

type RoomState struct {
	Players map[string]*PlayerState `json:"players"`
}

type PlayerState struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

var Station *Room

func configureRooms() {
	Station, _ = NewRoom("station")
	Station.Permissions.TakeOwnership = false
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
		State:        &RoomState{make(map[string]*PlayerState)},
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
	room.Broadcast(scope, action, data)
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
	r.BroadcastState()
}

func (r *Room) AddClient(c *Client) error {
	// if !c.Authenticated {
	// 	return ErrorUnauthenticated
	// }
	r.State.Players[c.User.NameNormal] = new(PlayerState)
	r.Clients[c] = len(r.Clients)
	r.ClientsEnd++
	if r.Owner == "" && c.Authenticated && r.Permissions.TakeOwnership {
		r.AssignOwnership(c)
	}
	return nil
}

func (r *Room) GetFirstClient() (client *Client) {
	lowest := r.ClientsEnd
	// this might be quicker as a standard loop from 0 -> r.ClientsEnd
	// but I think this might handle a very sparse list better
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

	delete(r.State.Players, username)
	delete(r.Clients, c)

	if username == r.Owner && r.Permissions.TakeOwnership {
		r.AssignOwnership(r.GetFirstClient())
	}
}

func (r *Room) Broadcast(scope, action string, data interface{}) {
	for c := range r.Clients {
		// if !c.Authenticated {
		// 	continue
		// }
		// log.Printf("[chat] sending to %s", c.User.NameNormal)
		err := c.Conn.WriteJSON(WSResponse{
			Error:  0,
			Action: WSAction{scope, action, r.Name},
			Data:   data,
		})
		if err != nil {
			log.Printf("[chat/%s] error sending to %s, removing from room (%v)",
				r.Name, c.User.NameNormal, err)
			r.RemoveClient(c)
		}
	}
}

func (r *Room) GetPublicData() RoomPublicData {
	data := RoomPublicData{
		Owner:        r.Owner,
		Name:         r.Name,
		FriendlyName: r.FriendlyName,
		State:        r.State,
	}
	if r.Activity != nil {
		data.Activity = r.Activity.Name()
		data.ActivityState = r.Activity.PublicState()
	}
	return data
}

func (r *Room) BroadcastState() {
	r.Broadcast("chat", "update_room", r.GetPublicData())
}
