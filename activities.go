package main

import (
	"log"
	"math/rand"
)

type ActivityHandler func(source *Client, room *Room, action string, data []byte) (interface{}, error)
type ActivityInit func(owner *Client, room *Room) error
type Activity struct {
	Name    string
	Handler ActivityHandler
	Init    ActivityInit
}

var activities = map[string]Activity{
	"fireworks": Activity{
		Name:    "fireworks",
		Handler: fireworksHandler,
		Init:    fireworksInit,
	},
	"cards": Activity{
		Name:    "bootleg hearthstone",
		Handler: cardsHandler,
		Init:    cardsInit,
	},
}

func handleActivityList(source *Client, target string, data []byte) (interface{}, error) {
	list := make(map[string]string)
	for k, a := range activities {
		list[k] = a.Name
	}

	return list, nil
}

func handleActivityAction(source *Client, target, action string, data []byte) (interface{}, error) {
	if !source.Authenticated {
		return nil, ErrorUnauthenticated
	}

	room, exists := rooms[target]
	if !exists {
		return nil, ErrorRoomMissing
	}

	if room != source.CurrentRoom {
		return nil, ErrorWrongRoom
	}

	act, exists := activities[room.Activity]
	if !exists {
		return nil, ErrorActMissing
	}

	return act.Handler(source, room, action, data)
}

// --- fireworks ---

type fireworkLaunch struct {
	Hue      int     `json:"hue"`
	Position int     `json:"position"`
	Lifetime float32 `json:"lifetime"`
}

func fireworksHandler(source *Client, room *Room, action string, data []byte) (interface{}, error) {
	if action != "launch" {
		return nil, ErrorActInvalidAction
	}

	room.Broadcast("activity", "launch", fireworkLaunch{
		Hue:      rand.Intn(360),
		Position: rand.Intn(100),
		Lifetime: rand.Float32()*2 + 1,
	})

	log.Printf("[ws/activity] %s launched a firework", source.User.Name)

	return nil, nil
}

func fireworksInit(owner *Client, room *Room) error {
	return nil
}

// --- bootleg hearthstone ---

type cardsState struct {
	Hue      int     `json:"hue"`
	Position int     `json:"position"`
	Lifetime float32 `json:"lifetime"`
}

func cardsHandler(source *Client, room *Room, action string, data []byte) (interface{}, error) {
	if action != "launch" {
		return nil, ErrorActInvalidAction
	}

	room.Broadcast("activity", "launch", fireworkLaunch{
		Hue:      rand.Intn(360),
		Position: rand.Intn(100),
		Lifetime: rand.Float32()*2 + 1,
	})

	log.Printf("[ws/activity] %s launched a firework", source.User.Name)

	return nil, nil
}

func cardsInit(owner *Client, room *Room) error {
	return nil
}
