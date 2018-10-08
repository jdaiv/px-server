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
		Activity string `json:"activity"`
	}
	roomData struct {
		Owner         string      `json:"owner"`
		Name          string      `json:"name"`
		FriendlyName  string      `json:"friendly_name"`
		Activity      string      `json:"activity"`
		ActivityState interface{} `json:"activity_state"`
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

	return list, err
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
		return nil, ErrorRoomMissing
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

	return roomData{
		Owner:        room.Owner,
		Name:         room.Name,
		FriendlyName: room.FriendlyName,
		Activity:     room.Activity,
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

	if len(roomInfo.Name) < 2 || len(roomInfo.Name) > 255 {
		return nil, ErrorInvalidData
	}

	room, err := NewRoomRandom()
	if err != nil {
		return nil, err
	}
	room.FriendlyName = roomInfo.Name
	room.AssignOwnership(source)
	source.CurrentRoom = room
	log.Printf("[chat/%s] %s created with name %s", room.Name, source.User.Name, roomInfo.Name)
	return WSResponse{
		Error:   0,
		Message: "success",
		Action:  WSAction{"chat", "join_room", room.Name},
		Data: roomData{
			Owner:        room.Owner,
			Name:         room.Name,
			FriendlyName: room.FriendlyName,
			Activity:     room.Activity,
		},
	}, room.AddClient(source)
}

func handleModifyRoom(source *Client, target string, data []byte) (interface{}, error) {
	if !source.Authenticated {
		return nil, ErrorUnauthenticated
	}

	room, exists := rooms[target]
	if !exists {
		return nil, ErrorRoomMissing
	}

	if source.CurrentRoom != room {
		return nil, ErrorWrongRoom
	}

	if room.Owner != source.User.NameNormal {
		return nil, ErrorNotOwner
	}

	var _data roomCreate

	if err := parseIncoming(data, &_data); err != nil {
		return nil, err
	}

	if len(_data.Name) < 2 || len(_data.Name) > 255 {
		return nil, ErrorInvalidData
	}

	if len(_data.Activity) > 0 && (len(_data.Activity) < 2 || len(_data.Activity) > 255) {
		return nil, ErrorInvalidData
	}

	if _data.Activity != room.Activity {
		if len(_data.Activity) > 0 {
			act, exists := activities[_data.Activity]
			if !exists {
				return nil, ErrorActMissing
			}
			room.Activity = _data.Activity
			act.Init(source, room)
			log.Printf("[chat/%s] %s changed activity to %s", room.Name, source.User.Name, _data.Activity)
			room.Broadcast("chat", "new_message", messageSend{
				Content: fmt.Sprintf("changed activity to %s", _data.Activity),
				From:    "server",
				Class:   MESSAGE_CLASS_SERVER,
			})
		} else {
			room.Activity = ""
			room.ActivityState = nil
			log.Printf("[chat/%s] %s cleared activity", room.Name, source.User.Name)
			room.Broadcast("chat", "new_message", messageSend{
				Content: "stopped the current activity",
				From:    "server",
				Class:   MESSAGE_CLASS_SERVER,
			})
		}
	}

	if room.FriendlyName != _data.Name {
		room.FriendlyName = _data.Name
		log.Printf("[chat/%s] %s changed name of room to %s", room.Name, source.User.Name, _data.Name)
		room.Broadcast("chat", "new_message", messageSend{
			Content: fmt.Sprintf("changed room name to %s", _data.Name),
			From:    "server",
			Class:   MESSAGE_CLASS_SERVER,
		})
	}

	room.BroadcastState()

	return nil, nil
}
