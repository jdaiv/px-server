package rpg

import (
	"fmt"
	"log"
)

type Zone struct {
	Id          int             `json:"-"`
	Name        string          `json:"name"`
	SpawnPoint  Position        `json:"-"`
	Map         *ZoneMap        `json:"map"`
	EntityCount int             `json:"entCount"`
	NPCCount    int             `json:"npcCount"`
	Entities    map[int]*Entity `json:"entities"`
	NPCs        map[int]*NPC    `json:"npcs"`
	Players     map[int]*Player `json:"-"`
	Items       map[int]bool    `json:"-"`

	CombatInfo  *ZoneCombatData `json:"-"`
	DisplayData ZoneDisplayData `json:"-"`
}

type ZoneDisplayData struct {
	Name              string           `json:"name"`
	Width             int              `json:"width"`
	Height            int              `json:"height"`
	Map               map[string]*tile `json:"map"`
	Entities          []EntityInfo     `json:"entities"`
	Players           []PlayerInfo     `json:"players"`
	NPCs              []NPCInfo        `json:"npcs"`
	Items             []ItemInfo       `json:"items"`
	InCombat          bool             `json:"inCombat"`
	CurrentInitiative int              `json:"currentInitiative"`
	Combatants        []CombatInfo     `json:"combatants"`
}

func (g *RPG) InitZone(z *Zone) {
	if z.Map == nil {
		z.Map = NewZoneMap(10, 10, g.Defs.Tiles[2])
	}
	g.Items.LoadIntoZone(z)
	z.Players = make(map[int]*Player)
	if z.Entities == nil {
		z.Entities = make(map[int]*Entity)
	} else {
		invalidEnts := make([]int, 0)
		for id, e := range z.Entities {
			entityDef, ok := g.Defs.Entities[e.Type]
			if !ok {
				log.Printf("[rpg/zone/%s] can't create entity due to missing definition %v", z.Name, e)
				invalidEnts = append(invalidEnts, id)
				continue
			}
			e.RootDef = entityDef
		}
		for _, id := range invalidEnts {
			delete(z.Entities, id)
		}
	}
	if z.NPCs == nil {
		z.NPCs = make(map[int]*NPC)
	}
	z.Items = make(map[int]bool)
	g.BuildCollisionMap(z)
	z.CombatInfo = &ZoneCombatData{}
}

func (g *RPG) ZoneTick(z *Zone) {
	if z.CombatInfo.InCombat {
		g.CombatTick(z)
	} else {
		for _, p := range z.Players {
			maxHP := p.Stats.MaxHP
			maxAP := p.Stats.MaxAP

			if p.Timers.HP <= 0 {
				if p.HP < maxHP {
					p.HP += 1
					if p.HP > maxHP {
						p.HP = maxHP
					}
					p.Timers.HP = BASE_HP_REGEN
					g.Players.SetDirty(p.Id)
					g.Zones.SetDirty(z.Id)
				}
			} else {
				p.Timers.HP -= 1
			}

			if p.Timers.AP <= 0 {
				if p.AP < maxAP {
					p.AP += 1
					if p.AP > maxAP {
						p.AP = maxAP
					}
					p.Timers.AP = BASE_AP_REGEN
					g.Players.SetDirty(p.Id)
					g.Zones.SetDirty(z.Id)
				}
			} else {
				p.Timers.AP -= 1
			}
		}
	}
}

func (g *RPG) BuildCollisionMap(z *Zone) {
	for _, t := range z.Map.Tiles {
		t.Blocking = g.Defs.Tiles[t.Tile].Blocking
		t.BlockingEnt = false
	}
	for _, e := range z.Entities {
		if e.RootDef.Blocking {
			size := e.RootDef.Size
			for x := 0; x < size[0]; x++ {
				for y := 0; y < size[1]; y++ {
					_x := e.X + x
					_y := e.Y + y
					z.Map.SetBlocking(_x, _y, true)
				}
			}
		}
	}
	for _, e := range z.NPCs {
		z.Map.SetBlocking(e.X, e.Y, true)
	}
}

func (g *RPG) BuildDisplayData(z *Zone) {
	tiles := make(map[string]*tile)
	for i, t := range z.Map.Tiles {
		x, y := uncompactCoords(i)
		tiles[fmt.Sprintf("%d,%d", x, y)] = t
	}
	entities := make([]EntityInfo, len(z.Entities))
	idx := 0
	for _, e := range z.Entities {
		entities[idx] = e.GetInfo()
		idx++
	}
	players := make([]PlayerInfo, len(z.Players))
	idx = 0
	for _, p := range z.Players {
		players[idx] = p.GetInfo(g)
		idx++
	}
	items := make([]ItemInfo, len(z.Items))
	idx = 0
	for id, _ := range z.Items {
		if item, ok := g.Items.Get(id); ok {
			items[idx] = item.GetInfo()
			idx++
		}
	}
	npcs := make([]NPCInfo, len(z.NPCs))
	idx = 0
	for _, n := range z.NPCs {
		npcs[idx] = n.GetInfo()
		idx++
	}
	combatants := make([]CombatInfo, 0)
	if z.CombatInfo.InCombat {
		for _, info := range z.CombatInfo.Combatants {
			combatants = append(combatants, *info)
		}
	}
	z.DisplayData = ZoneDisplayData{
		Name:              z.Name,
		Map:               tiles,
		Entities:          entities,
		Players:           players,
		Items:             items,
		NPCs:              npcs,
		InCombat:          z.CombatInfo.InCombat,
		CurrentInitiative: z.CombatInfo.CurrentInitiative,
		Combatants:        combatants,
	}
}

func (g *RPG) AddEntity(z *Zone, entType string, x, y int, updateCollisions bool) (*Entity, error) {
	id := z.EntityCount
	ent, err := g.NewEntity(z, id, entType, x, y)
	if err != nil {
		log.Printf("[rpg/zone/%s/createent] error creating entity '%s': %v", z.Name, entType, err)
		return nil, err
	}
	z.Entities[id] = ent
	z.EntityCount += 1

	if updateCollisions {
		g.BuildCollisionMap(z)
	}

	return ent, nil
}

func (g *RPG) RemoveEntity(z *Zone, entId int) {
	delete(z.Entities, entId)
	g.BuildCollisionMap(z)
}

func (g *RPG) SendMessage(z *Zone, player *Player, text string) {
	playerId := -1
	if player != nil {
		playerId = player.Id
	}
	g.Outgoing <- OutgoingMessage{
		PlayerId: playerId,
		Zone:     z.Id,
		Type:     ACTION_CHAT,
		Params: map[string]interface{}{
			"message": text,
		},
	}
}

type effectParams map[string]interface{}

func (g *RPG) SendEffect(z *Zone, effectType string, params effectParams) {
	params["type"] = effectType
	g.Outgoing <- OutgoingMessage{
		PlayerId: -1,
		Zone:     z.Id,
		Type:     ACTION_EFFECT,
		Params:   params,
	}
}

func (g *RPG) AddPlayer(z *Zone, player *Player, x, y int) {
	player.CurrentZone = z.Id

	player.X = x
	player.Y = y

	z.Players[player.Id] = player
	g.CheckCombat(z)
}

func (g *RPG) RemovePlayer(z *Zone, player *Player) {
	if player.CurrentZone != z.Id {
		return
	}

	delete(z.Players, player.Id)
	g.CheckCombat(z)
}

func (z *Zone) Move(x, y int, direction string) (int, int, bool) {
	_x := x
	_y := y

	switch direction {
	case "N":
		_y += 1
	case "S":
		_y -= 1
	case "E":
		_x += 1
	case "W":
		_x -= 1
	}

	if z.Map.IsBlocking(_x, _y) {
		_x = x
		_y = y
	}

	for _, p := range z.Players {
		if _x == p.X && _y == p.Y {
			_x = x
			_y = y
		}
	}

	return _x, _y, (x != _x || y != _y)
}

func (z *Zone) MoveNoclip(x, y int, direction string) (int, int, bool) {
	_x := x
	_y := y

	switch direction {
	case "N":
		_y += 1
	case "S":
		_y -= 1
	case "E":
		_x += 1
	case "W":
		_x -= 1
	}

	return _x, _y, true
}

// thanks http://cavaliercoder.com/blog/optimized-abs-for-int64-in-go.html
func intAbs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

func nextTo(x1, y1, x2, y2 int) bool {
	return intAbs(int64(x2-x1)) <= 1 && intAbs(int64(y2-y1)) <= 1
}
