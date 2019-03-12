package rpg

import (
	"errors"
	"log"
)

type Zone struct {
	Parent       *RPG
	Def          ZoneDef
	Name         string
	Width        int
	Height       int
	Map          []Tile
	CollisionMap []bool
	EntityCount  int
	NPCCount     int
	Entities     map[int]*Entity
	NPCs         map[int]*NPC
	Players      map[int]*Player
	Items        map[int]bool

	CombatInfo  *ZoneCombatData
	DisplayData ZoneDisplayData

	Dirty bool
}

type ZoneDisplayData struct {
	Width      int            `json:"width"`
	Height     int            `json:"height"`
	Map        []Tile         `json:"map"`
	Entities   []EntityInfo   `json:"entities"`
	Players    []PlayerInfo   `json:"players"`
	NPCs       []NPCInfo      `json:"npcs"`
	Items      []ItemInfo     `json:"items"`
	CombatInfo ZoneCombatData `json:"combatInfo"`
}

func NewZone(parent *RPG, name string, def ZoneDef) *Zone {
	tileMap := make([]Tile, def.Width*def.Height)
	for i, t := range def.Map {
		tileMap[i] = TileFromDef(t, parent.Defs)
	}

	zone := &Zone{
		Parent:       parent,
		Def:          def,
		Name:         name,
		Width:        def.Width,
		Height:       def.Height,
		Map:          tileMap,
		CollisionMap: make([]bool, def.Width*def.Height),
		Players:      make(map[int]*Player),
		NPCs:         make(map[int]*NPC),
		Entities:     make(map[int]*Entity),
		CombatInfo:   &ZoneCombatData{},
	}

	parent.Items.LoadIntoZone(zone)

	for _, e := range def.Entity {
		zone.AddEntity(e, false)
	}

	zone.BuildCollisionMap()

	for _, n := range def.NPC {
		zone.AddNPC(n, true)
	}

	return zone
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
					z.Dirty = true
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
					z.Dirty = true
				}
			} else {
				p.Timers.AP -= 1
			}
		}
	}
}

func (z *Zone) BuildCollisionMap() {
	for i, t := range z.Map {
		z.CollisionMap[i] = t.Blocking
	}
	for _, e := range z.Entities {
		if e.RootDef.Blocking {
			z.CollisionMap[e.X+e.Y*z.Width] = true
		}
	}
	for _, e := range z.NPCs {
		z.CollisionMap[e.X+e.Y*z.Width] = true
	}
}

func (z *Zone) BuildDisplayData() {
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
	for _, n := range z.NPCs {
		npcs[idx] = n.GetInfo()
		idx++
	}
	z.DisplayData = ZoneDisplayData{
		Width:      z.Width,
		Height:     z.Height,
		Map:        z.Map,
		Entities:   entities,
		Players:    players,
		Items:      items,
		NPCs:       npcs,
		CombatInfo: *z.CombatInfo,
	}
}

func (z *Zone) AddEntity(def ZoneEntityDef, updateCollisions bool) {
	id := z.EntityCount
	ent, err := NewEntity(z, id, def)
	if err != nil {
		log.Printf("[rpg/zone/%s/create] error creating entity '%s': %v", z.Name, def.Type, err)
		return
	}
	z.Entities[id] = ent
	z.EntityCount += 1

	if updateCollisions {
		z.BuildCollisionMap()
	}
}

func (z *Zone) RemoveEntity(entId int) {
	delete(z.Entities, entId)
	z.BuildCollisionMap()
}

func (z *Zone) AddNPC(def ZoneNPCDef, updateCollisions bool) {
	if z.CollisionMap[def.Position[0]+def.Position[1]*z.Width] {
		log.Printf("[rpg/zone/%s/createnpc] can't create npc '%s': coords %d,%d blocked", z.Name, def.Type, def.Position[0], def.Position[1])
		return
	}

	id := z.NPCCount
	npc, err := NewNPC(z, id, def)
	if err != nil {
		log.Printf("[rpg/zone/%s/createnpc] error creating npc '%s': %v", z.Name, def.Type, err)
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
	item.CurrentZone = z.Name

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
	item.CurrentZone = z.Name
	z.Items[item.Id] = true
	z.Parent.Items.Save(item)
}

func (z *Zone) RemoveItem(item *Item) {
	item.CurrentZone = ""
	delete(z.Items, item.Id)
}

func (z *Zone) SendMessage(player *Player, text string) {
	playerId := -1
	if player != nil {
		playerId = player.Id
	}
	z.Parent.Outgoing <- OutgoingMessage{
		PlayerId: playerId,
		Zone:     z.Name,
		Type:     ACTION_CHAT,
		Params: map[string]interface{}{
			"message": text,
		},
	}
}

func (z *Zone) SendEffect(effectType string, x, y int) {
	z.Parent.Outgoing <- OutgoingMessage{
		PlayerId: -1,
		Zone:     z.Name,
		Type:     ACTION_EFFECT,
		Params: map[string]interface{}{
			"type": effectType,
			"x":    x,
			"y":    y,
		},
	}
}

func (z *Zone) AddPlayer(player *Player, x, y int) {
	player.CurrentZone = z.Name

	if x >= 0 {
		player.X = x
	} else {
		player.X = z.Def.SpawnPoint[0]
	}

	if y >= 0 {
		player.Y = y
	} else {
		player.Y = z.Def.SpawnPoint[1]
	}

	z.Players[player.Id] = player
	z.CheckCombat()
}

func (z *Zone) RemovePlayer(player *Player) {
	if player.CurrentZone != z.Name {
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

	if _x < 0 {
		_x = 0
	}
	if _x >= z.Width {
		_x = z.Width - 1
	}
	if _y < 0 {
		_y = 0
	}
	if _y >= z.Height {
		_y = z.Height - 1
	}

	if z.CollisionMap[_x+_y*z.Width] {
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
