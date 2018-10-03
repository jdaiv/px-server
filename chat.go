package main

import (
	"fmt"
	"log"
)

type (
	messageRecv struct {
		Room    string `json:"room"`
		Content string `json:"content"`
	}
	messageSend struct {
		From    string `json:"from"`
		Content string `json:"content"`
		Room    string `json:"room"`
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

func handleChatAction(source *Client, action string, data []byte) (interface{}, error) {
	if !source.Authenticated {
		return nil, ErrorUnauthenticated
	}

	switch action {
	case "message":
		var msg messageRecv

		if err := parseIncoming(data, &msg); err != nil {
			return nil, err
		}

		if len(msg.Room) <= 0 || len(msg.Room) > 64 {
			return nil, ErrorInvalidData
		}

		if len(msg.Content) <= 0 || len(msg.Content) > 256 {
			return nil, ErrorInvalidData
		}

		err := BroadcastToRoom(msg.Room, "chat", "new_message", messageSend{
			Content: msg.Content,
			Room:    msg.Room,
			From:    source.User.Name,
		})

		if err != nil {
			log.Printf("[chat] %s tried to send to %s", source.User.Name, msg.Room)
		} else {
			log.Printf("[chat/%s] %s: %s", msg.Room, source.User.Name, msg.Content)
		}

		return nil, err

	case "list_rooms":
		list := roomList{
			Rooms: make([]string, len(rooms)),
		}
		for name, _ := range rooms {
			list.Rooms = append(list.Rooms, name)
		}
		return list, nil

	case "join_room":
		var roomName string

		if err := parseIncoming(data, &roomName); err != nil {
			return nil, err
		}

		if len(roomName) <= 0 || len(roomName) > 64 {
			return nil, ErrorInvalidData
		}

		room, exists := rooms[roomName]
		if !exists {
			log.Printf("[chat] %s tried to join %s", source.User.Name, roomName)
			return WSResponse{
				Error:   ErrorRoomMissing.Code(),
				Message: ErrorRoomMissing.ExternalMessage(),
				Scope:   "chat",
				Action:  action,
				Data:    roomName,
			}, nil
		}

		log.Printf("[chat/%s] %s joined", roomName, source.User.Name)

		BroadcastToRoom(roomName, "chat", "new_message", messageSend{
			Content: fmt.Sprintf("%s joined", source.User.Name),
			Room:    roomName,
			From:    "server",
		})

		return WSResponse{
			Error:   0,
			Message: "success",
			Scope:   "chat",
			Action:  "join_room",
			Data: roomData{
				Name:         room.Name,
				FriendlyName: room.FriendlyName,
				Activity:     room.CurrentActivity,
			},
		}, room.AddClient(source)

	case "create_room":
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
		room.Owner = source.User.Name
		source.CurrentRoom = room
		log.Printf("[chat/%s] %s created", room.Name, source.User.Name)
		return WSResponse{
			Error:   0,
			Message: "success",
			Scope:   "chat",
			Action:  "join_room",
			Data: roomData{
				Name:         room.Name,
				FriendlyName: room.FriendlyName,
				Activity:     room.CurrentActivity,
			},
		}, room.AddClient(source)
	}

	return nil, ErrorMissingAction
}
