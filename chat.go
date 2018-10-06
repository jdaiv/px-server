package main

import (
	"fmt"
	"log"
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
		Activity int    `json:"activity"`
	}
	roomData struct {
		Name         string `json:"name"`
		FriendlyName string `json:"friendly_name"`
		Activity     int    `json:"activity"`
	}
)

func handleChatMessage(source *Client, target string, data []byte) (interface{}, error) {
	if !source.Authenticated {
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

	if target == "public" {
		list = make([]string, len(authenticatedClients))
		for _, c := range authenticatedClients {
			if c.Authenticated {
				list = append(list, c.User.Name)
			}
		}
	} else {
		room, exists := rooms[target]
		if !exists {
			err = ErrorRoomMissing
		} else {
			list = make([]string, len(room.Clients))
			for c := range room.Clients {
				if c.Authenticated {
					list = append(list, c.User.Name)
				}
			}
		}
	}

	return WSResponse{
		Error:   0,
		Message: "success",
		Action:  WSAction{"chat", "list_users", target},
		Data:    list,
	}, err
}

func handleListRooms(source *Client, target string, data []byte) (interface{}, error) {
	if !source.Authenticated {
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
	// if !source.Authenticated {
	// 	return nil, ErrorUnauthenticated
	// }

	if source.CurrentRoom != nil {
		return nil, ErrorClientHasRoom
	}

	if len(target) <= 0 || len(target) > 64 {
		return nil, ErrorInvalidData
	}

	room, exists := rooms[target]
	if !exists {
		log.Printf("[chat] %s tried to join %s", source.User.NameNormal, target)
		return WSResponse{
			Error:   ErrorRoomMissing.Code(),
			Message: ErrorRoomMissing.ExternalMessage(),
			Action:  WSAction{"chat", "join_room", target},
		}, nil
	}

	log.Printf("[chat/%s] %s joined", target, source.User.NameNormal)

	if source.Authenticated {
		BroadcastToRoom(target, "chat", "new_message", messageSend{
			Content: fmt.Sprintf("%s joined", source.User.Name),
			From:    "server",
		})
	}

	err := room.AddClient(source)
	if err != nil {
		return nil, err
	}
	source.CurrentRoom = room

	return WSResponse{
		Error:   0,
		Message: "success",
		Action:  WSAction{"chat", "join_room", room.Name},
		Data: roomData{
			Name:         room.Name,
			FriendlyName: room.FriendlyName,
			Activity:     room.CurrentActivity,
		},
	}, nil
}

func handleCreateRoom(source *Client, target string, data []byte) (interface{}, error) {
	if !source.Authenticated {
		return nil, ErrorUnauthenticated
	}

	if source.CurrentRoom != nil {
		return nil, ErrorClientHasRoom
	}

	var roomInfo roomCreate

	if err := parseIncoming(data, &roomInfo); err != nil {
		return nil, err
	}

	if len(roomInfo.Name) > 255 || roomInfo.Activity < 0 || roomInfo.Activity >= ACT_MAX {
		return nil, ErrorInvalidData
	}

	room, err := NewRoomRandom()
	if err != nil {
		return nil, err
	}
	room.FriendlyName = roomInfo.Name
	room.CurrentActivity = roomInfo.Activity
	room.AssignOwnership(source)
	source.CurrentRoom = room
	log.Printf("[chat/%s] %s created", room.Name, source.User.Name)
	return WSResponse{
		Error:   0,
		Message: "success",
		Action:  WSAction{"chat", "join_room", room.Name},
		Data: roomData{
			Name:         room.Name,
			FriendlyName: room.FriendlyName,
			Activity:     room.CurrentActivity,
		},
	}, room.AddClient(source)
}
