package rpg

import (
	"database/sql"
	"log"
)

type RPG struct {
	Defs    *Definitions
	Zones   map[string]*Zone
	Players *PlayerDB
	Items   *ItemDB

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
		Players:  NewPlayerDB(db),
		Items:    NewItemDB(db),
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
		if incoming.Data.Type == ACTION_TICK {
			g.Tick()
		} else if incoming.Data.Type == ACTION_JOIN {
			g.PlayerJoin(incoming)
		} else if incoming.Data.Type == ACTION_LEAVE {
			g.PlayerLeave(incoming.PlayerId)
		} else {
			p, ok := g.Players.Get(incoming.PlayerId)
			if !ok {
				log.Printf("couldn't find player %d", incoming.PlayerId)
				continue
			}
			zone, ok := g.Zones[p.CurrentZone]
			if !ok {
				log.Printf("couldn't find zone %s for player %d (%s), placing at default", p.CurrentZone, p.Id, p.Name)
				p.CurrentZone = ""
				g.Zones[g.Defs.RPG.StartingZone].AddPlayer(p, -1, -1)
				g.Outgoing <- OutgoingMessage{
					PlayerId: p.Id,
					Zone:     g.Defs.RPG.StartingZone,
					Type:     ACTION_UPDATE,
				}
				continue
			}
			if !zone.CanAct(p) {
				log.Printf("player tried to act out of order %s", p.Name)
				continue
			}

			oldZone := p.CurrentZone

			g.Players.SetDirty(p.Id)

			switch incoming.Data.Type {
			case ACTION_MOVE:
				g.PlayerMove(p, zone, incoming.Data.Params)
			case ACTION_USE:
				g.PlayerUse(p, zone, incoming.Data.Params)
			case ACTION_TAKE_ITEM:
				g.PlayerTakeItem(p, zone, incoming.Data.Params)
			case ACTION_EQUIP_ITEM:
				g.PlayerEquipItem(p, zone, incoming.Data.Params)
			case ACTION_UNEQUIP_ITEM:
				g.PlayerUnequipItem(p, zone, incoming.Data.Params)
			case ACTION_DROP_ITEM:
				g.PlayerDropItem(p, zone, incoming.Data.Params)
			case ACTION_ATTACK:
				g.PlayerAttack(p, zone, incoming.Data.Params)
			}

			zone.PostPlayerAction(p)
			zone.CheckCombat()
			zone.BuildCollisionMap()

			g.Players.Commit()

			g.Outgoing <- OutgoingMessage{
				PlayerId: p.Id,
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
}

func (g *RPG) PrepareDisplay() {
	for _, z := range g.Zones {
		z.BuildDisplayData()
	}
}

func (g *RPG) BuildDisplayFor(pId int) DisplayData {
	p, ok := g.Players.Get(pId)
	if !ok {
		return DisplayData{}
	}

	zone, ok := g.Zones[p.CurrentZone]
	if !ok {
		return DisplayData{}
	}

	return DisplayData{
		Player: p.GetInfo(g),
		Zone:   zone.DisplayData,
	}
}

func (g *RPG) Tick() {
	for _, z := range g.Zones {
		z.Tick()
	}
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

	p, ok := g.Players.Get(msg.PlayerId)
	if !ok {
		log.Printf("[rpg/player/join] error loading player %d:%s", msg.PlayerId, name)
	}

	p.Name = name
	p.Rebuild(g)

	if p.HP <= 0 {
		p.HP = p.Stats.MaxHP()
	}

	if _, hasZone := g.Zones[p.CurrentZone]; hasZone {
		g.Zones[p.CurrentZone].AddPlayer(p, p.X, p.Y)
	} else {
		g.Zones[g.Defs.RPG.StartingZone].AddPlayer(p, -1, -1)
		g.Players.SetDirty(p.Id)
	}

	g.Outgoing <- OutgoingMessage{
		PlayerId: msg.PlayerId,
		Zone:     p.CurrentZone,
		Type:     ACTION_UPDATE,
	}
}

func (g *RPG) PlayerLeave(id int) {
	p, ok := g.Players.Get(id)
	if !ok {
		return
	}
	zone := p.CurrentZone
	g.Players.SetDirty(id)
	g.Zones[zone].RemovePlayer(p)

	g.Outgoing <- OutgoingMessage{
		PlayerId: id,
		Zone:     zone,
		Type:     ACTION_UPDATE,
	}
}

func (g *RPG) SaveAllPlayers() {
	log.Printf("saving all players")
	g.Players.Commit()
}

func (g *RPG) KillPlayer(p *Player) {
	g.Zones[p.CurrentZone].AddEntity(ZoneEntityDef{
		Name:     "corpse of " + p.Name,
		Position: Position{p.X, p.Y},
		Type:     "corpse",
		Strings:  map[string]string{"type": "player"},
	}, false)
	g.PlayerReset(p)
}

func (g *RPG) KillNPC(z *Zone, n *NPC) {
	delete(z.NPCs, n.Id)
	z.AddEntity(ZoneEntityDef{
		Name:     "corpse of " + n.Name,
		Position: Position{n.X, n.Y},
		Type:     "corpse",
		Strings:  map[string]string{"type": n.Type},
	}, false)
}

func (g *RPG) PlayerReset(p *Player) {
	g.Zones[p.CurrentZone].RemovePlayer(p)
	p.HP = p.Stats.MaxHP()
	g.Zones[g.Defs.RPG.StartingZone].AddPlayer(p, -1, -1)

	g.Outgoing <- OutgoingMessage{
		PlayerId: p.Id,
		Zone:     p.CurrentZone,
		Type:     ACTION_UPDATE,
	}
}
