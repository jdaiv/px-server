package main

import (
	"log"

	"bitbucket.org/panicexpress/backend/station"
	"github.com/google/uuid"
)

var rooms = make(map[string]*Room)

// TODO: convert to interface?
type Room struct {
	Name         string
	FriendlyName string
	Clients      map[*Client]int
	ClientsEnd   int
	Owner        string
	Area         *station.Area
}

type RoomPublicData struct {
	Name          string        `json:"name"`
	FriendlyName  string        `json:"friendly_name"`
	Area          *station.Area `json:"state"`
	Activity      string        `json:"activity"`
	ActivityState interface{}   `json:"activity_state"`
}

var Public *Room

func configureRooms() {
	Public, _ = NewRoom("public")
}

func NewRoom(name string) (*Room, error) {
	if _, exists := rooms[name]; exists {
		return nil, ErrorRoomExists
	}
	r := &Room{
		Name:         name,
		FriendlyName: name,
		Clients:      make(map[*Client]int, 0),
		Area:         station.NewArea(),
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

/* func (r *Room) AssignOwnership(c *Client) {
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
} */

func (r *Room) AddClient(c *Client) error {
	// if !c.Authenticated {
	// 	return ErrorUnauthenticated
	// }
	r.Area.Handle(c.User.NameNormal, "create", "player", []float64{0, 0, 0, 0})
	r.Clients[c] = len(r.Clients)
	r.ClientsEnd++
	return nil
}

func (r *Room) GetFirstClient() (client *Client) {
	lowest := r.ClientsEnd
	// this might be quicker as a standard loop from 0 -> r.ClientsEnd
	// but I think this might handle a very sparse list better
	for c, i := range r.Clients {
		if i < lowest && c.State == AUTHENTICATED {
			client = c
			lowest = i
		}
	}
	return
}

func (r *Room) RemoveClient(c *Client) {
	r.Area.Handle(c.User.NameNormal, "remove", "player", []float64{})
	delete(r.Clients, c)
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
		Name:         r.Name,
		FriendlyName: r.FriendlyName,
		Area:         r.Area,
	}
	return data
}

func (r *Room) BroadcastState() {
	r.Broadcast("chat", "update_room", r.GetPublicData())
}
