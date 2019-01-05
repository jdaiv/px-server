package main

import (
	"fmt"
	"log"

	"bitbucket.org/panicexpress/backend/station"
)

const (
	MESSAGE_CLASS_NORMAL = ""
	MESSAGE_CLASS_SERVER = "server"
)

type (
	messageRecv struct {
		Content string `json:"content"`
	}
	messageSend struct {
		From    string `json:"from"`
		Content string `json:"content"`
		Class   string `json:"class"`
	}
	roomList struct {
		Rooms []string `json:"rooms"`
	}
	roomCreate struct {
		Name     string `json:"name"`
		Activity string `json:"activity"`
	}
)

func handlePlayerMove(source *Client, target string, data []byte) (interface{}, error) {
	if source.CurrentRoom == nil {
		return nil, ErrorRoomMissing
	}

	var msg map[string]station.Delta

	if err := parseIncoming(data, &msg); err != nil {
		return nil, err
	}

	source.CurrentRoom.Area.Recv(source.User.NameNormal, msg)

	return nil, nil
}

func handleChatMessage(source *Client, target string, data []byte) (interface{}, error) {
	if source.State != AUTHENTICATED {
		return nil, ErrorUnauthenticated
	}

	var msg messageRecv

	if err := parseIncoming(data, &msg); err != nil {
		return nil, err
	}

	if len(target) <= 0 || len(target) > 64 {
		return nil, ErrorInvalidData
	}

	if len(msg.Content) <= 0 || len(msg.Content) > 256 {
		return nil, ErrorInvalidData
	}

	var err error
	if target == "public" {
		err = BroadcastToAll(target, "chat", "new_message", messageSend{
			Content: msg.Content,
			From:    source.User.Name,
		})
	} else {
		err = BroadcastToRoom(target, "chat", "new_message", messageSend{
			Content: msg.Content,
			From:    source.User.Name,
		})
	}

	if err != nil {
		log.Printf("[chat] %s tried to send to %s", source.User.NameNormal, target)
	} else {
		log.Printf("[chat/%s] %s: %s", target, source.User.NameNormal, msg.Content)
	}

	return nil, err
}

func handleListUsers(source *Client, target string, data []byte) (interface{}, error) {
	var list []string
	var err error

	room, exists := rooms[target]
	if !exists {
		err = ErrorRoomMissing
	} else {
		list = make([]string, len(room.Clients))
		for c := range room.Clients {
			// if c.Authenticated {
			list = append(list, c.User.Name)
			// }
		}
	}

	return list, err
}

func handleListRooms(source *Client, target string, data []byte) (interface{}, error) {
	if source.State != AUTHENTICATED {
		return nil, ErrorUnauthenticated
	}

	list := roomList{
		Rooms: make([]string, len(rooms)),
	}
	for name, _ := range rooms {
		list.Rooms = append(list.Rooms, name)
	}
	return list, nil
}

func handleJoinRoom(source *Client, target string, data []byte) (interface{}, error) {
	if source.State != AUTHENTICATED {
		return nil, ErrorUnauthenticated
	}

	if source.CurrentRoom != nil {
		// return nil, ErrorClientHasRoom
		source.CurrentRoom.RemoveClient(source)
	}

	if len(target) <= 0 || len(target) > 64 {
		return nil, ErrorInvalidData
	}

	room, exists := rooms[target]
	if !exists {
		log.Printf("[chat] %s tried to join %s", source.User.NameNormal, target)
		return nil, ErrorRoomMissing
	}

	log.Printf("[chat/%s] %s joined", target, source.User.NameNormal)

	if source.State == AUTHENTICATED {
		BroadcastToRoom(target, "chat", "new_message", messageSend{
			Content: fmt.Sprintf("%s joined", source.User.Name),
			From:    "server",
			Class:   MESSAGE_CLASS_SERVER,
		})
	}

	err := room.AddClient(source)
	if err != nil {
		return nil, err
	}
	source.CurrentRoom = room

	return room.GetPublicData(), nil
}
