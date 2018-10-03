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

func handleActivityAction(source *Client, target string, data []byte) (interface{}, error) {
	if !source.Authenticated {
		return nil, ErrorUnauthenticated
	}

	BroadcastToRoom(target, "activity", "launch", fireworkLaunch{
		Room:     target,
		Hue:      rand.Intn(360),
		Position: rand.Intn(100),
		Lifetime: rand.Float32()*2 + 1,
	})

	log.Printf("[ws/activity] %s launched a firework", source.User.Name)
	return nil, nil
}
