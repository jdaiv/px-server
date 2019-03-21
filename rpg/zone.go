package rpg

import (
	"errors"
	"fmt"
	"log"
)

type Zone struct {
	Id          int             `json:"-"`
	Parent      *RPG            `json:"-"`
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
	Name       string          `json:"name"`
	Width      int             `json:"width"`
	Height     int             `json:"height"`
	Map        map[string]tile `json:"map"`
	Entities   []EntityInfo    `json:"entities"`
	Players    []PlayerInfo    `json:"players"`
	NPCs       []NPCInfo       `json:"npcs"`
	Items      []ItemInfo      `json:"items"`
	CombatInfo ZoneCombatData  `json:"combatInfo"`
}

func (z *Zone) Init(rpg *RPG) {
	z.Parent = rpg
	if z.Map == nil {
		z.Map = NewZoneMap(10, 10, rpg.Defs.Tiles[2])
	}
	rpg.Items.LoadIntoZone(z)
	z.Players = make(map[int]*Player)
	if z.Entities == nil {
		z.Entities = make(map[int]*Entity)
	} else {
		invalidEnts := make([]int, 0)
		for id, e := range z.Entities {
			entityDef, ok := z.Parent.Defs.Entities[e.Type]
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
	z.BuildCollisionMap()
	z.CombatInfo = &ZoneCombatData{}
}

func (z *Zone) Tick() {
	if z.CombatInfo.InCombat {
		z.CombatTick()
	} else {
		for _, p := range z.Players {
			maxHP := p.Stats.MaxHP()
			maxAP := p.Stats.MaxAP()

			if p.Timers.HP <= 0 {
				if p.HP < maxHP {
					p.HP += 1
					if p.HP > maxHP {
						p.HP = maxHP
					}
					p.Timers.HP = BASE_HP_REGEN
					z.Parent.Players.SetDirty(p.Id)
					z.Parent.Zones.SetDirty(z.Id)
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
					z.Parent.Players.SetDirty(p.Id)
					z.Parent.Zones.SetDirty(z.Id)
				}
			} else {
				p.Timers.AP -= 1
			}
		}
	}
}

func (z *Zone) BuildCollisionMap() {
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

func (z *Zone) BuildDisplayData() {
	tiles := make(map[string]tile)
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
		players[idx] = p.GetInfo(z.Parent)
		idx++
	}
	items := make([]ItemInfo, len(z.Items))
	idx = 0
	for id, _ := range z.Items {
		if item, ok := z.Parent.Items.Get(id); ok {
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
	z.DisplayData = ZoneDisplayData{
		Name:       z.Name,
		Map:        tiles,
		Entities:   entities,
		Players:    players,
		Items:      items,
		NPCs:       npcs,
		CombatInfo: *z.CombatInfo,
	}
}

func (z *Zone) AddEntity(entType string, x, y int, updateCollisions bool) (*Entity, error) {
	id := z.EntityCount
	ent, err := NewEntity(z, id, entType, x, y)
	if err != nil {
		log.Printf("[rpg/zone/%s/create] error creating entity '%s': %v", z.Name, entType, err)
		return nil, err
	}
	z.Entities[id] = ent
	z.EntityCount += 1

	if updateCollisions {
		z.BuildCollisionMap()
	}

	return ent, nil
}

func (z *Zone) RemoveEntity(entId int) {
	delete(z.Entities, entId)
	z.BuildCollisionMap()
}

func (z *Zone) AddNPC(npcType string, x, y int, updateCollisions bool) {
	if z.Map.IsBlocking(x, y) {
		log.Printf("[rpg/zone/%s/createnpc] can't create npc '%s': coords %d,%d blocked", z.Name, npcType, x, y)
		return
	}

	id := z.NPCCount
	npc, err := NewNPC(z, id, npcType, x, y)
	if err != nil {
		log.Printf("[rpg/zone/%s/createnpc] error creating npc '%s': %v", z.Name, npcType, err)
		return
	}
	z.NPCs[id] = npc
	z.NPCCount += 1

	if updateCollisions {
		z.BuildCollisionMap()
	}

	z.CheckCombat()
}

func (z *Zone) RemoveNPC(npcId int) {
	delete(z.NPCs, npcId)
	z.BuildCollisionMap()
	z.CheckCombat()
}

func (z *Zone) AddItem(itemType string, x, y int) (Item, error) {
	def, ok := z.Parent.Defs.Items[itemType]
	if !ok {
		log.Printf("[rpg/zone/%s/createitem] item doesn't exist '%s'", z.Name, itemType)
		return Item{}, errors.New("item doesn't exist")
	}

	item, ok := z.Parent.Items.New(def)
	if !ok {
		log.Printf("[rpg/zone/%s/createitem] error creating item '%s'", z.Name, itemType)
		return Item{}, nil
	}

	item.X = x
	item.Y = y
	item.CurrentZone = z.Id

	z.Items[item.Id] = true
	z.Parent.Items.Save(item)

	return item, nil
}

func (z *Zone) AddExistingItem(itemId int, x, y int) {
	item, ok := z.Parent.Items.Get(itemId)
	if !ok {
		log.Printf("[rpg/zone/%s/additem] item %d doesn't exist", z.Name, itemId)
		return
	}
	item.Held = false
	item.X = x
	item.Y = y
	item.CurrentZone = z.Id
	z.Items[item.Id] = true
	z.Parent.Items.Save(item)
}

func (z *Zone) RemoveItem(item *Item) {
	item.CurrentZone = -1
	delete(z.Items, item.Id)
}

func (z *Zone) SendMessage(player *Player, text string) {
	playerId := -1
	if player != nil {
		playerId = player.Id
	}
	z.Parent.Outgoing <- OutgoingMessage{
		PlayerId: playerId,
		Zone:     z.Id,
		Type:     ACTION_CHAT,
		Params: map[string]interface{}{
			"message": text,
		},
	}
}

func (z *Zone) SendEffect(effectType string, x, y int) {
	z.Parent.Outgoing <- OutgoingMessage{
		PlayerId: -1,
		Zone:     z.Id,
		Type:     ACTION_EFFECT,
		Params: map[string]interface{}{
			"type": effectType,
			"x":    x,
			"y":    y,
		},
	}
}

func (z *Zone) AddPlayer(player *Player, x, y int) {
	player.CurrentZone = z.Id

	if x >= 0 {
		player.X = x
	} else {
		player.X = z.SpawnPoint[0]
	}

	if y >= 0 {
		player.Y = y
	} else {
		player.Y = z.SpawnPoint[1]
	}

	z.Players[player.Id] = player
	z.CheckCombat()
}

func (z *Zone) RemovePlayer(player *Player) {
	if player.CurrentZone != z.Id {
		return
	}

	delete(z.Players, player.Id)
	z.CheckCombat()
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

// thanks http://cavaliercoder.com/blog/optimized-abs-for-int64-in-go.html
func intAbs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

func nextTo(x1, y1, x2, y2 int) bool {
	return intAbs(int64(x2-x1)) <= 1 && intAbs(int64(y2-y1)) <= 1
}
