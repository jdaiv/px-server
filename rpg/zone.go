package rpg

import (
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
	Entities     map[int]*Entity
	Players      map[int]*Player

	DisplayData ZoneDisplayData
}

type ZoneDisplayData struct {
	Width    int                       `json:"width"`
	Height   int                       `json:"height"`
	Map      []Tile                    `json:"map"`
	Entities []EntityInfo              `json:"entities"`
	Players  map[int]PlayerDisplayData `json:"players"`
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
		Entities:     make(map[int]*Entity),
	}

	for _, e := range def.Entity {
		zone.AddEntity(e, false)
	}

	zone.BuildCollisionMap()

	return zone
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
	z.DisplayData = ZoneDisplayData{
		Width:    z.Width,
		Height:   z.Height,
		Map:      z.Map,
		Entities: entities,
		Players:  players,
	}
}

func (z *Zone) AddEntity(def ZoneEntityDef, updateCollisions bool) {
	id := z.EntityCount
	ent, err := NewEntity(z, id, def)
	if err != nil {
		log.Printf("[rpg/zone/create/%s] error creating entity '%s': %v", z.Name, def.Type, err)
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
}

func (z *Zone) RemovePlayer(player *Player) {
	if player.CurrentZone != z.Name {
		return
	}

	delete(z.Players, player.Id)
	player.CurrentZone = ""
}

func (z *Zone) MovePlayer(player *Player, direction string) {
	if player.CurrentZone != z.Name {
		return
	}

	x := player.X
	y := player.Y

	switch direction {
	case "N":
		y += 1
	case "S":
		y -= 1
	case "E":
		x += 1
	case "W":
		x -= 1
	}

	if x < 0 {
		x = 0
	}
	if x >= z.Width {
		x = z.Width - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= z.Height {
		y = z.Height - 1
	}

	if z.CollisionMap[x+y*z.Width] {
		return
	}

	for _, p := range z.Players {
		if x == p.X && y == p.Y {
			return
		}
	}

	player.X = x
	player.Y = y
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

	if intAbs(int64(ent.X-player.X)) > 1 || intAbs(int64(ent.X-player.X)) > 1 {
		log.Printf("[rpg/zone/%s/use] player %d tried to use ent %d, but was too far away", z.Name, player.Id, entId)
		return false
	}

	log.Printf("[rpg/zone/%s/use] using ent %d", z.Name, entId)

	needsUpdate, err := ent.Use(player)
	if err != nil {
		log.Printf("[rpg/zone/%s/use] failed to use ent %d (%s): %v", z.Name, entId, ent.Type, err)
	}

	return needsUpdate
}

// thanks http://cavaliercoder.com/blog/optimized-abs-for-int64-in-go.html
func intAbs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}
