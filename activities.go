package main

import (
	"fmt"
	"log"
	"math/rand"
)

type Activity interface {
	Name() string
	Init(owner *Client, room *Room) error
	Tick(source *Client, action string, data []byte) (interface{}, error)
	PublicState() interface{}
}

type ActivityBase struct {
	Owner *Client `json:"-"`
	Room  *Room   `json:"-"`
}

func MakeActivity(actType string, owner *Client, room *Room) (act Activity, err error) {
	switch actType {
	case "fireworks":
		act = &FireworksActivity{}
	case "tictactoe":
		act = &TictactoeActivity{}
	case "cards":
		act = &FireworksActivity{}
	default:
		return nil, ErrorActMissing
	}
	err = act.Init(owner, room)
	return
}

func handleActivityList(source *Client, target string, data []byte) (interface{}, error) {
	return map[string]string{
		"fireworks": "fireworks",
		"tictactoe": "tic-tac-toe",
		// "cards":     "bootleg hearthstone",
	}, nil
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

	if room.Activity == nil {
		return nil, ErrorActMissing
	}

	return room.Activity.Tick(source, action, data)
}

// --- fireworks ---

type FireworksActivity struct {
	ActivityBase
}

func (f *FireworksActivity) Name() string {
	return "fireworks"
}

func (f *FireworksActivity) Init(owner *Client, room *Room) error {
	f.Owner = owner
	f.Room = room
	return nil
}

func (f *FireworksActivity) Tick(source *Client, action string, data []byte) (interface{}, error) {
	if action != "launch" {
		return nil, ErrorActInvalidAction
	}

	f.Room.Broadcast("activity", "launch", struct {
		Hue      int     `json:"hue"`
		Position float32 `json:"position"`
		Lifetime float32 `json:"lifetime"`
		Velocity float32 `json:"velocity"`
	}{rand.Intn(360), rand.Float32(), rand.Float32()*2 + 1, rand.Float32()})

	log.Printf("[ws/activity] %s launched a firework", source.User.Name)

	return nil, nil
}

func (f *FireworksActivity) PublicState() interface{} {
	return nil
}

// --- tic tac toe ---

type TictactoeActivity struct {
	ActivityBase
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

func (a *TictactoeActivity) Name() string {
	return "tictactoe"
}

func (a *TictactoeActivity) Init(owner *Client, room *Room) error {
	a.Owner = owner
	a.Room = room
	a.CurrentPlayer = rand.Intn(50) % 2
	return nil
}

func (a *TictactoeActivity) Tick(source *Client, action string, data []byte) (interface{}, error) {
	if action != "move" {
		return nil, ErrorActInvalidAction
	}

	if a.Winner > 0 {
		return nil, nil
	}

	if (a.CurrentPlayer != 0 && source.User.NameNormal == a.Room.Owner) ||
		(a.CurrentPlayer == 0 && source.User.NameNormal != a.Room.Owner) {
		log.Println("player tried to take turn out of order")
		return nil, ErrorActInvalidAction
	}

	var turnData tictactoeTurn

	if err := parseIncoming(data, &turnData); err != nil {
		return nil, err
	}

	idx := turnData.X + turnData.Y*3
	if idx < 0 || idx > 9 || a.Board[idx] != 0 {
		return nil, ErrorActInvalidAction
	}

	a.Board[idx] = a.CurrentPlayer + 1

	for i := 0; i < len(winConditions); i += 3 {
		x := winConditions[i+0]
		y := winConditions[i+1]
		z := winConditions[i+2]
		if a.Board[x]+a.Board[y]+a.Board[z] == 0 {
			continue
		}
		if a.Board[x] == a.Board[y] &&
			a.Board[y] == a.Board[z] {
			a.Winner = a.Board[x]
			break
		}
	}

	if a.Winner == 0 {
		draw := true
		for _, v := range a.Board {
			if v == 0 {
				draw = false
			}
		}
		if draw {
			a.Winner = 3
		}
	}

	switch a.Winner {
	case 1:
		a.Room.Broadcast("chat", "new_message", messageSend{
			Content: fmt.Sprintf("%s wins!", a.Owner.User.Name),
			From:    "activity",
			Class:   MESSAGE_CLASS_SERVER,
		})
		break
	case 2:
		a.Room.Broadcast("chat", "new_message", messageSend{
			Content: fmt.Sprintf("%s wins!", source.User.Name),
			From:    "activity",
			Class:   MESSAGE_CLASS_SERVER,
		})
		break
	case 3:
		a.Room.Broadcast("chat", "new_message", messageSend{
			Content: "draw",
			From:    "activity",
			Class:   MESSAGE_CLASS_SERVER,
		})
		break
	}

	if a.CurrentPlayer == 0 {
		a.CurrentPlayer = 1
	} else {
		a.CurrentPlayer = 0
	}

	a.Room.BroadcastState()

	return nil, nil
}

func (a *TictactoeActivity) PublicState() interface{} {
	return a
}

// --- bootleg hearthstone ---
