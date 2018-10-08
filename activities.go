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
	"tictactoe": Activity{
		Name:    "tic-tac-toe",
		Handler: tictactoeHandler,
		Init:    tictactoeInit,
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

type tictactoeState struct {
	Board         [9]int `json:"board"`
	CurrentPlayer int    `json:"current_player"`
	Winner        int    `json:"winner"`
}
type tictactoeTurn struct {
	X int `json:"x"`
	Y int `json:"y"`
}

var winConditions = []int{
	0, 1, 2,
	3, 4, 5,
	6, 7, 8,
	0, 3, 6,
	1, 4, 7,
	2, 5, 8,
	0, 4, 8,
	2, 4, 6,
}

func tictactoeHandler(source *Client, room *Room, action string, data []byte) (interface{}, error) {
	if action != "move" {
		return nil, ErrorActInvalidAction
	}

	state, stateOk := room.ActivityState.(*tictactoeState)
	if !stateOk {
		log.Println("activity state not configured correctly!")
		return nil, ErrorInternal
	}

	if state.Winner > 0 {
		return nil, nil
	}

	if (state.CurrentPlayer != 0 && source.User.NameNormal == room.Owner) ||
		(state.CurrentPlayer == 0 && source.User.NameNormal != room.Owner) {
		log.Println("player tried to take turn out of order")
		return nil, ErrorActInvalidAction
	}

	var turnData tictactoeTurn

	if err := parseIncoming(data, &turnData); err != nil {
		return nil, err
	}

	idx := turnData.X + turnData.Y*3
	if idx < 0 || idx > 9 || state.Board[idx] != 0 {
		return nil, ErrorActInvalidAction
	}

	state.Board[idx] = state.CurrentPlayer + 1

	for i := 0; i < len(winConditions); i += 3 {
		x := winConditions[i+0]
		y := winConditions[i+1]
		z := winConditions[i+2]
		if state.Board[x]+state.Board[y]+state.Board[z] == 0 {
			continue
		}
		if state.Board[x] == state.Board[y] &&
			state.Board[y] == state.Board[z] {
			state.Winner = state.Board[x]
			break
		}
	}

	if state.Winner == 0 {
		draw := true
		for _, v := range state.Board {
			if v == 0 {
				draw = false
			}
		}
		if draw {
			state.Winner = 3
		}
	}

	if state.CurrentPlayer == 0 {
		state.CurrentPlayer = 1
	} else {
		state.CurrentPlayer = 0
	}

	room.BroadcastState()

	return nil, nil
}

func tictactoeInit(owner *Client, room *Room) error {
	room.ActivityState = &tictactoeState{
		CurrentPlayer: rand.Intn(2),
	}
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
