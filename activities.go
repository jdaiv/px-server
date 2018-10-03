package main

import (
	"log"
	"math/rand"
)

const (
	ACT_NOTHING = iota
	ACT_FIREWORKS
	ACT_MAX
)

type fireworkLaunch struct {
	Room     string  `json:"room"`
	Hue      int     `json:"hue"`
	Position int     `json:"position"`
	Lifetime float32 `json:"lifetime"`
}

func handleActivityAction(source *Client, action string, data []byte) (interface{}, error) {
	if !source.Authenticated {
		return nil, ErrorUnauthenticated
	}

	switch action {
	case "launch":
		var msg messageRecv

		if err := parseIncoming(data, &msg); err != nil {
			return nil, err
		}

		BroadcastToRoom(msg.Room, "activity", "launch", fireworkLaunch{
			Room:     msg.Room,
			Hue:      rand.Intn(360),
			Position: rand.Intn(100),
			Lifetime: rand.Float32()*2 + 1,
		})

		log.Printf("[ws/activity] %s launched a firework", source.User.Name)
		return nil, nil
	}
	return nil, ErrorMissingAction
}
