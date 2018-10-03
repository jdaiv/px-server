package main

import (
	"fmt"
	"log"
)

type (
	messageRecv struct {
		Content string `json:"content"`
	}
	messageSend struct {
		From    string `json:"from"`
		Content string `json:"content"`
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

	err := BroadcastToRoom(target, "chat", "new_message", messageSend{
		Content: msg.Content,
		From:    source.User.Name,
	})

	if err != nil {
		log.Printf("[chat] %s tried to send to %s", source.User.Name, target)
	} else {
		log.Printf("[chat/%s] %s: %s", target, source.User.Name, msg.Content)
	}

	return nil, err
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
	if !source.Authenticated {
		return nil, ErrorUnauthenticated
	}

	if len(target) <= 0 || len(target) > 64 {
		return nil, ErrorInvalidData
	}

	room, exists := rooms[target]
	if !exists {
		log.Printf("[chat] %s tried to join %s", source.User.Name, target)
		return WSResponse{
			Error:   ErrorRoomMissing.Code(),
			Message: ErrorRoomMissing.ExternalMessage(),
			Action:  WSAction{"chat", "join_room", target},
		}, nil
	}

	log.Printf("[chat/%s] %s joined", target, source.User.Name)

	BroadcastToRoom(target, "chat", "new_message", messageSend{
		Content: fmt.Sprintf("%s joined", source.User.Name),
		From:    "server",
	})

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
	room.Owner = source.User.Name
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
