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
	Items        map[int]*Item

	CombatInfo  *ZoneCombatData
	DisplayData ZoneDisplayData
}

type ZoneDisplayData struct {
	Width      int                       `json:"width"`
	Height     int                       `json:"height"`
	Map        []Tile                    `json:"map"`
	Entities   []EntityInfo              `json:"entities"`
	Players    map[int]PlayerDisplayData `json:"players"`
	NPCs       map[int]NPCInfo           `json:"npcs"`
	Items      map[int]ItemInfo          `json:"items"`
	CombatInfo ZoneCombatData            `json:"combatInfo"`
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
		Items:        parent.GetItemsForZone(name),
		CombatInfo:   &ZoneCombatData{},
	}

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
	entities := make([]EntityInfo, 0)
	for _, e := range z.Entities {
		entities = append(entities, e.GetInfo())
	}
	players := make(map[int]PlayerDisplayData)
	for _, p := range z.Players {
		p.UpdateDisplay()
		players[p.Id] = p.DisplayData
	}
	items := make(map[int]ItemInfo)
	for id, i := range z.Items {
		items[id] = i.GetInfo()
	}
	npcs := make(map[int]NPCInfo)
	for id, n := range z.NPCs {
		npcs[id] = n.GetInfo()
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

func (z *Zone) AddItem(itemType string, x, y int) (*Item, error) {
	def, ok := z.Parent.Defs.Items[itemType]
	if !ok {
		log.Printf("[rpg/zone/%s/createitem] item doesn't exist '%s'", z.Name, itemType)
		return nil, errors.New("item doesn't exist")
	}

	item, err := z.Parent.NewItem(def)
	if err != nil {
		log.Printf("[rpg/zone/%s/createitem] error creating item '%s': %v", z.Name, itemType, err)
		return nil, errors.New("db error")
	}

	item.X = x
	item.Y = y
	item.CurrentZone = z.Name
	if err := item.Save(); err != nil {
		log.Printf("[rpg/zone/%s/createitem] error updating item %d: %v", z.Name, item.Id, err)
		return item, errors.New("db error")
	}

	z.Items[item.Id] = item
	return item, nil
}

func (z *Zone) AddExistingItem(item *Item, x, y int) error {
	item.Held = false
	item.X = x
	item.Y = y
	item.CurrentZone = z.Name
	if err := item.Save(); err != nil {
		log.Printf("[rpg/zone/%s/createitem] error updating item %d: %v", z.Name, item.Id, err)
		return errors.New("db error")
	}

	z.Items[item.Id] = item
	return nil
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
	if player.CurrentZone != "" {
		return
	}
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
	player.CurrentZone = ""
	z.CheckCombat()
}

func (z *Zone) MovePlayer(player *Player, direction string) {
	if player.CurrentZone != z.Name {
		return
	}

	x, y, ok := z.Move(player.X, player.Y, direction)
	if !ok {
		return
	}

	if !z.CheckAPCost(player, 1) {
		return
	}
	player.X = x
	player.Y = y
	z.PostPlayerAction(player)
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

func (z *Zone) UseItem(player *Player, entId int) bool {
	if player.CurrentZone != z.Name {
		return false
	}

	ent, ok := z.Entities[entId]
	if !ok {
		log.Printf("[rpg/zone/%s/use] couldn't find ent %d", z.Name, entId)
		return false
	}

	if intAbs(int64(ent.X-player.X)) > 1 || intAbs(int64(ent.Y-player.Y)) > 1 {
		log.Printf("[rpg/zone/%s/use] player %d tried to use ent %d, but was too far away", z.Name, player.Id, entId)
		return false
	}

	log.Printf("[rpg/zone/%s/use] using ent %d", z.Name, entId)

	if !z.CheckAPCost(player, 1) {
		return false
	}

	needsUpdate, err := ent.Use(player)
	if err != nil {
		log.Printf("[rpg/zone/%s/use] failed to use ent %d (%s): %v", z.Name, entId, ent.Type, err)
	}

	z.PostPlayerAction(player)

	return needsUpdate
}

func (z *Zone) TakeItem(player *Player, itemId int) bool {
	if player.CurrentZone != z.Name {
		return false
	}

	item, ok := z.Items[itemId]
	if !ok {
		log.Printf("[rpg/zone/%s/take_item] couldn't find item %d", z.Name, itemId)
		return false
	}

	if intAbs(int64(item.X-player.X)) > 1 || intAbs(int64(item.X-player.X)) > 1 {
		log.Printf("[rpg/zone/%s/take_item] player %d tried to take item %d but was too far away", z.Name, player.Id, itemId)
		return false
	}

	log.Printf("[rpg/zone/%s/take_item] grabbing item %d", z.Name, itemId)

	if !z.CheckAPCost(player, 1) {
		return false
	}

	err := item.Give(player)
	if err != nil {
		log.Printf("[rpg/zone/%s/take_item] failed to update item %d: %v", z.Name, itemId, err)
	}
	delete(z.Items, itemId)

	z.PostPlayerAction(player)

	return true
}

// thanks http://cavaliercoder.com/blog/optimized-abs-for-int64-in-go.html
func intAbs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}
