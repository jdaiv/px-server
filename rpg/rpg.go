package rpg

import (
	"database/sql"
	"log"
)

type RPG struct {
	Defs    *Definitions
	Zones   map[string]*Zone
	Players map[int]*Player
	Items   map[int]*Item

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
	Player PlayerInfo      `json:"player"`
}

func NewRPG(defDir string, db *sql.DB) (*RPG, error) {
	defs, err := LoadDefinitions(defDir)
	if err != nil {
		return nil, err
	}

	rpg := &RPG{
		Defs:     defs,
		Zones:    make(map[string]*Zone),
		Players:  make(map[int]*Player),
		Items:    make(map[int]*Item),
		Incoming: make(chan IncomingMessage),
		Outgoing: make(chan OutgoingMessage),
		DB:       db,
	}

	for k, v := range defs.Zones {
		if defs.RPG.Zones[k].Enabled {
			rpg.Zones[k] = NewZone(rpg, k, v)
		}
	}

	return rpg, nil
}

/*
	incoming action types:
		* Player join
		* Player leave
		* Player move
		* Player use
		* Player change zone

	outgoing action types:
		* Player update
		* State update
		* Chat message
*/

func (g *RPG) HandleMessages() {
	for {
		incoming := <-g.Incoming
		switch incoming.Data.Type {
		case ACTION_JOIN:
			g.PlayerJoin(incoming)
		case ACTION_LEAVE:
			g.PlayerLeave(incoming.PlayerId)
		case ACTION_MOVE:
			g.PlayerMove(incoming)
		case ACTION_USE:
			g.PlayerUse(incoming)
		case ACTION_TAKE_ITEM:
			g.PlayerTakeItem(incoming)
		case ACTION_EQUIP_ITEM:
			g.PlayerEquipItem(incoming)
		case ACTION_UNEQUIP_ITEM:
			g.PlayerUnequipItem(incoming)
		case ACTION_DROP_ITEM:
			g.PlayerDropItem(incoming)
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
		Player: p.GetInfo(),
		Zone:   zone.DisplayData,
	}
}

func (g *RPG) Tick(dt float64) {

}

func (g *RPG) PlayerJoin(msg IncomingMessage) {
	name := "ERROR"
	nameParam, ok := msg.Data.Params["name"]
	if ok {
		nameStr, ok := nameParam.(string)
		if ok {
			name = nameStr
		}
	}

	data, err := LoadPlayer(g.DB, msg.PlayerId)
	if err != nil {
		log.Printf("[rpg/player/join] error loading player: %v", err)
	}

	p := &Player{
		Id:   msg.PlayerId,
		Name: name,
		X:    data.X,
		Y:    data.Y,
	}

	g.LoadItemsForPlayer(p)

	g.Players[msg.PlayerId] = p

	if _, hasZone := g.Zones[data.CurrentZone]; hasZone {
		g.Zones[data.CurrentZone].AddPlayer(p, data.X, data.Y)
	} else {
		g.Zones[g.Defs.RPG.StartingZone].AddPlayer(p, -1, -1)
	}

	g.Outgoing <- OutgoingMessage{
		PlayerId: msg.PlayerId,
		Zone:     p.CurrentZone,
		Type:     ACTION_UPDATE,
	}
}

func (g *RPG) PlayerLeave(id int) {
	p, ok := g.Players[id]
	if !ok {
		return
	}
	zone := p.CurrentZone
	SavePlayer(g.DB, p)
	g.Zones[zone].RemovePlayer(p)

	g.Outgoing <- OutgoingMessage{
		PlayerId: id,
		Zone:     zone,
		Type:     ACTION_UPDATE,
	}
}

func (g *RPG) SaveAllPlayers() {
	for _, p := range g.Players {
		SavePlayer(g.DB, p)
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

func (g *RPG) PlayerUse(msg IncomingMessage) {
	p, ok := g.Players[msg.PlayerId]
	if !ok {
		log.Printf("couldn't find player %d", msg.PlayerId)
		return
	}

	zone, ok := g.Zones[p.CurrentZone]
	if !ok {
		log.Printf("couldn't find zone %s", p.CurrentZone)
		return
	}

	entIdParam, ok := msg.Data.Params["id"]
	if !ok {
		log.Println("couldn't find ent id param")
		return
	}

	entId, ok := entIdParam.(float64)
	if !ok {
		log.Println("ent id param not number")
		return
	}

	oldZone := p.CurrentZone

	if zone.UseItem(p, int(entId)) {
		g.Outgoing <- OutgoingMessage{
			PlayerId: msg.PlayerId,
			Zone:     p.CurrentZone,
			Type:     ACTION_UPDATE,
		}
		if p.CurrentZone != oldZone {
			g.Outgoing <- OutgoingMessage{
				Zone: oldZone,
				Type: ACTION_UPDATE,
			}
		}
	}
}

func (g *RPG) PlayerTakeItem(msg IncomingMessage) {
	p, ok := g.Players[msg.PlayerId]
	if !ok {
		log.Printf("couldn't find player %d", msg.PlayerId)
		return
	}

	zone, ok := g.Zones[p.CurrentZone]
	if !ok {
		log.Printf("couldn't find zone %s", p.CurrentZone)
		return
	}

	entIdParam, ok := msg.Data.Params["id"]
	if !ok {
		log.Println("couldn't find item id param")
		return
	}

	entId, ok := entIdParam.(float64)
	if !ok {
		log.Println("item id param not number")
		return
	}

	if zone.TakeItem(p, int(entId)) {
		g.Outgoing <- OutgoingMessage{
			PlayerId: msg.PlayerId,
			Zone:     p.CurrentZone,
			Type:     ACTION_UPDATE,
		}
	}
}

func (g *RPG) PlayerEquipItem(msg IncomingMessage) {
	p, ok := g.Players[msg.PlayerId]
	if !ok {
		log.Printf("couldn't find player %d", msg.PlayerId)
		return
	}

	itemIdParam, ok := msg.Data.Params["id"]
	if !ok {
		log.Println("couldn't find item id param")
		return
	}

	itemId, ok := itemIdParam.(float64)
	if !ok {
		log.Println("item id param not number")
		return
	}

	if p.EquipItem(int(itemId)) {
		g.Outgoing <- OutgoingMessage{
			PlayerId: msg.PlayerId,
			Zone:     p.CurrentZone,
			Type:     ACTION_UPDATE,
		}
	}
}

func (g *RPG) PlayerUnequipItem(msg IncomingMessage) {
	p, ok := g.Players[msg.PlayerId]
	if !ok {
		log.Printf("couldn't find player %d", msg.PlayerId)
		return
	}

	slotParam, ok := msg.Data.Params["slot"]
	if !ok {
		log.Println("couldn't find item slot param")
		return
	}

	slot, ok := slotParam.(string)
	if !ok {
		log.Println("item slot param not number")
		return
	}

	if p.UnequipItem(slot) {
		g.Outgoing <- OutgoingMessage{
			PlayerId: msg.PlayerId,
			Zone:     p.CurrentZone,
			Type:     ACTION_UPDATE,
		}
	}
}

func (g *RPG) PlayerDropItem(msg IncomingMessage) {
	p, ok := g.Players[msg.PlayerId]
	if !ok {
		log.Printf("couldn't find player %d", msg.PlayerId)
		return
	}

	zone, ok := g.Zones[p.CurrentZone]
	if !ok {
		log.Printf("couldn't find zone %s", p.CurrentZone)
		return
	}

	itemIdParam, ok := msg.Data.Params["id"]
	if !ok {
		log.Println("couldn't find ent id param")
		return
	}

	itemId, ok := itemIdParam.(float64)
	if !ok {
		log.Println("item id param not number")
		return
	}

	if p.DropItem(zone, int(itemId)) {
		g.Outgoing <- OutgoingMessage{
			PlayerId: msg.PlayerId,
			Zone:     p.CurrentZone,
			Type:     ACTION_UPDATE,
		}
	}
}
