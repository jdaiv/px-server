package rpg

import (
	"database/sql"
	"log"
)

type RPG struct {
	Defs    *Definitions
	Zones   *ZoneDB
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
	Zone     int                    `json:"-"`
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

	// EDITOR VALUES
	Defs      *Definitions   `json:"defs,omitempty"`
	DebugZone *Zone          `json:"debugZone,omitempty"`
	AllZones  map[int]string `json:"allZones,omitempty"`
}

func NewRPG(defDir string, db *sql.DB) (*RPG, error) {
	defs, err := LoadDefinitions(defDir)
	if err != nil {
		return nil, err
	}

	rpg := &RPG{
		Defs:     defs,
		Players:  NewPlayerDB(db),
		Items:    NewItemDB(db),
		Zones:    NewZoneDB(db),
		Incoming: make(chan IncomingMessage),
		Outgoing: make(chan OutgoingMessage),
		DB:       db,
	}

	for _, z := range rpg.Zones.AllZones {
		z.Init(rpg)
	}

	return rpg, nil
}

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
			p := g.Players.Get(incoming.PlayerId)
			zone, ok := g.Zones.Get(p.CurrentZone)
			if !ok {
				log.Printf("couldn't find zone %d for player %d (%s), placing at default", p.CurrentZone, p.Id, p.Name)
				p.CurrentZone = 1
				zone, _ := g.Zones.Get(1)
				zone.AddPlayer(p, -1, -1)
				g.Outgoing <- OutgoingMessage{
					PlayerId: p.Id,
					Zone:     1,
					Type:     ACTION_UPDATE,
				}
				continue
			}

			if incoming.Data.Type == ACTION_EDIT {
				g.HandleEdit(p, zone, incoming.Data.Params)
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
			p.BuildStats(g)
			zone.CheckCombat()
			zone.BuildCollisionMap()

			g.Players.Commit()

			if g.Zones.IsDirty(p.CurrentZone) {
				g.Outgoing <- OutgoingMessage{
					PlayerId: p.Id,
					Zone:     p.CurrentZone,
					Type:     ACTION_UPDATE,
				}
			}
			if p.CurrentZone != oldZone {
				g.Outgoing <- OutgoingMessage{
					Zone: oldZone,
					Type: ACTION_UPDATE,
				}
				g.Zones.SetDirty(p.CurrentZone)
			}
		}
	}
}

func (g *RPG) PrepareDisplay() {
	for _, z := range g.Zones.AllZones {
		z.BuildDisplayData()
	}
}

func (g *RPG) BuildDisplayFor(pId int) DisplayData {
	p := g.Players.Get(pId)

	zone, ok := g.Zones.Get(p.CurrentZone)
	if !ok {
		return DisplayData{}
	}

	d := DisplayData{
		Player: p.GetInfo(g),
		Zone:   zone.DisplayData,
	}

	if p.Editing {
		d.Defs = g.Defs
		d.DebugZone = zone
		allZones := make(map[int]string)
		for id, z := range g.Zones.AllZones {
			allZones[id] = z.Name
		}
		d.AllZones = allZones
	}

	return d
}

func (g *RPG) Tick() {
	for id, z := range g.Zones.AllZones {
		z.Tick()
		if g.Zones.IsDirty(id) {
			g.Outgoing <- OutgoingMessage{
				Zone: id,
				Type: ACTION_UPDATE,
			}
		}
	}
	g.Zones.Commit()
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

	p := g.Players.Get(msg.PlayerId)
	p.Name = name
	p.Rebuild(g)

	if p.HP <= 0 {
		p.HP = p.Stats.MaxHP()
	}

	if z, hasZone := g.Zones.Get(p.CurrentZone); hasZone {
		z.AddPlayer(p, p.X, p.Y)
	} else {
		zone, _ := g.Zones.Get(1)
		zone.AddPlayer(p, 0, 0)
		g.Players.SetDirty(p.Id)
	}

	g.Outgoing <- OutgoingMessage{
		PlayerId: msg.PlayerId,
		Zone:     p.CurrentZone,
		Type:     ACTION_UPDATE,
	}
}

func (g *RPG) PlayerLeave(id int) {
	p := g.Players.Get(id)
	zone := p.CurrentZone
	g.Players.SetDirty(id)
	if z, ok := g.Zones.Get(p.CurrentZone); ok {
		z.RemovePlayer(p)

		g.Outgoing <- OutgoingMessage{
			PlayerId: id,
			Zone:     zone,
			Type:     ACTION_UPDATE,
		}
	}
}

func (g *RPG) SaveAll() {
	log.Printf("saving all")
	g.Players.Commit()
	g.Zones.Commit()
}

func (g *RPG) KillPlayer(p *Player) {
	zone, ok := g.Zones.Get(p.CurrentZone)
	if ok {
		ent, err := zone.AddEntity("corpse", p.X, p.Y, false)
		if err == nil {
			ent.Name = "corpse of " + p.Name
			ent.Fields["type"] = "player"
		}
		g.Zones.SetDirty(p.CurrentZone)
	}
	g.PlayerReset(p)
}

func (g *RPG) KillNPC(z *Zone, n *NPC) {
	delete(z.NPCs, n.Id)
	ent, err := z.AddEntity("corpse", n.X, n.Y, false)
	if err == nil {
		ent.Name = "corpse of " + n.Name
		ent.Fields["type"] = n.Type
	}
	g.Zones.SetDirty(z.Id)
}

func (g *RPG) PlayerReset(p *Player) {
	if z, ok := g.Zones.Get(p.CurrentZone); ok {
		z.RemovePlayer(p)
	}
	p.HP = p.Stats.MaxHP()
	zone, _ := g.Zones.Get(1)
	zone.AddPlayer(p, -1, -1)

	g.Outgoing <- OutgoingMessage{
		PlayerId: p.Id,
		Zone:     p.CurrentZone,
		Type:     ACTION_UPDATE,
	}
}
