package main

import "log"

// import (
// 	"log"
// )

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
)

func handleChatMessage(source *Client, data []byte) (interface{}, error) {
	if source.State != AUTHENTICATED {
		return nil, ErrorUnauthenticated
	}

	var msg messageRecv

	if err := parseIncoming(data, &msg); err != nil {
		return nil, err
	}

	if len(msg.Content) <= 0 || len(msg.Content) > 256 {
		return nil, ErrorInvalidData
	}

	var err error
	err = BroadcastToAll(ACTION_CHAT_MESSAGE, messageSend{
		Content: msg.Content,
		From:    source.User.Name,
	})

	if err != nil {
		log.Printf("[chat] %s failed to send message: %v", source.User.NameNormal, err)
	} else {
		log.Printf("[chat] %s: %s", source.User.NameNormal, msg.Content)
	}

	return nil, err
}

// func handleListUsers(source *Client, target string, data []byte) (interface{}, error) {
// 	var list []string
// 	var err error

// 	room, exists := rooms[target]
// 	if !exists {
// 		err = ErrorRoomMissing
// 	} else {
// 		list = make([]string, len(room.Clients))
// 		for c := range room.Clients {
// 			// if c.Authenticated {
// 			list = append(list, c.User.Name)
// 			// }
// 		}
// 	}

// 	return list, err
// }
