package rpg

import (
	"database/sql"
)

type RPG struct {
	Zones   map[string]*Zone
	Players map[int]*Player

	Incoming chan IncomingMessage
	Outgoing chan OutgoingMessage
	DB       *sql.DB
}

type IncomingMessage struct {
	PlayerId int
	Data     IncomingMessageData
}

type OutgoingMessage struct {
	PlayerId int                    `json:"-"`
	Zone     string                 `json:"-"`
	Type     string                 `json:"type"`
	Params   map[string]interface{} `json:"params"`
}

type IncomingMessageData struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}

type DisplayData struct {
	Zone   ZoneDisplayData `json:"zone"`
	Player Player          `json:"player"`
}

func NewRPG(db *sql.DB) *RPG {
	return &RPG{
		Zones:    make(map[string]*Zone),
		Players:  make(map[int]*Player),
		Incoming: make(chan IncomingMessage),
		Outgoing: make(chan OutgoingMessage),
		DB:       db,
	}
}

/*
	incoming action types:
		* Player join
		* Player leave
		* Player move
		* Player use

	outgoing action types:
		* Player update
		* State update
		* Chat message

*/
const (
	ACTION_JOIN  = "join"
	ACTION_LEAVE = "leave"
	ACTION_MOVE  = "move"
	ACTION_USE   = "use"

	ACTION_UPDATE        = "state_update"
	ACTION_UPDATE_PLAYER = "player_update"
	ACTION_CHAT          = "chat_message"
)

func (g *RPG) HandleMessages() {
	for {
		incoming := <-g.Incoming
		switch incoming.Data.Type {
		case ACTION_JOIN:
			g.PlayerJoin(incoming.PlayerId)
		case ACTION_LEAVE:
			g.PlayerLeave(incoming.PlayerId)
		case ACTION_MOVE:
			g.PlayerMove(incoming)
		}
	}
}

func (g *RPG) PrepareDisplay() {
	for _, z := range g.Zones {
		z.BuildDisplayData()
	}
}

func (g *RPG) BuildDisplayFor(pId int) DisplayData {
	p, ok := g.Players[pId]
	if !ok {
		return DisplayData{}
	}

	zone, ok := g.Zones[p.CurrentZone]
	if !ok {
		return DisplayData{}
	}

	return DisplayData{
		Player: *p,
		Zone:   zone.DisplayData,
	}
}

func (g *RPG) Tick(dt float64) {

}

func (g *RPG) PlayerJoin(id int) {
	p := &Player{
		Id:          id,
		Name:        string(id),
		CurrentZone: "",
		X:           0,
		Y:           0,
	}

	g.Players[id] = p
	g.Zones["start"].AddPlayer(p)

	g.Outgoing <- OutgoingMessage{
		PlayerId: id,
		Zone:     p.CurrentZone,
		Type:     ACTION_UPDATE,
	}
}

func (g *RPG) PlayerLeave(id int) {
	p, ok := g.Players[id]
	if !ok {
		return
	}
	g.Zones[p.CurrentZone].RemovePlayer(p)

	g.Outgoing <- OutgoingMessage{
		PlayerId: id,
		Zone:     p.CurrentZone,
		Type:     ACTION_UPDATE,
	}
}

func (g *RPG) PlayerMove(msg IncomingMessage) {
	p, ok := g.Players[msg.PlayerId]
	if !ok {
		return
	}

	zone, ok := g.Zones[p.CurrentZone]
	if !ok {
		return
	}

	direction, ok := msg.Data.Params["direction"].(string)
	if !ok {
		return
	}

	zone.MovePlayer(p, direction)

	g.Outgoing <- OutgoingMessage{
		PlayerId: msg.PlayerId,
		Zone:     p.CurrentZone,
		Type:     ACTION_UPDATE,
	}
}
