package rpg

import (
	"log"
)

type Zone struct {
	Parent      *RPG
	Def         ZoneDef
	Name        string
	Width       int
	Height      int
	Map         []Tile
	EntityCount int
	Entities    map[int]*Entity
	Players     map[int]*Player

	DisplayData ZoneDisplayData
}

type ZoneDisplayData struct {
	Width    int                 `json:"width"`
	Height   int                 `json:"height"`
	Map      []Tile              `json:"map"`
	Entities []EntityInfo        `json:"entities"`
	Players  []PlayerDisplayData `json:"players"`
}

func NewZone(parent *RPG, name string, def ZoneDef) *Zone {
	tileMap := make([]Tile, def.Width*def.Height)
	for i, t := range def.Map {
		tileMap[i] = TileFromDef(t, parent.Defs)
	}

	zone := &Zone{
		Parent:   parent,
		Def:      def,
		Name:     name,
		Width:    def.Width,
		Height:   def.Height,
		Map:      tileMap,
		Players:  make(map[int]*Player),
		Entities: make(map[int]*Entity),
	}

	for _, e := range def.Entity {
		id := zone.EntityCount
		ent, err := NewEntity(zone, id, e)
		if err != nil {
			log.Printf("[rpg/zone/create/%s] error creating entity '%s': %v", name, e.Type, err)
			continue
		}
		zone.Entities[id] = ent
		zone.EntityCount += 1
	}

	return zone
}

func (z *Zone) BuildDisplayData() {
	entities := make([]EntityInfo, 0)
	for _, e := range z.Entities {
		entities = append(entities, e.GetInfo())
	}
	players := make([]PlayerDisplayData, 0)
	for _, p := range z.Players {
		p.UpdateDisplay()
		players = append(players, p.DisplayData)
	}
	z.DisplayData = ZoneDisplayData{
		Width:    z.Width,
		Height:   z.Height,
		Map:      z.Map,
		Entities: entities,
		Players:  players,
	}
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

func (z *Zone) AddPlayer(player *Player) {
	if player.CurrentZone != "" {
		return
	}
	player.CurrentZone = z.Name
	z.Players[player.Id] = player
}

func (z *Zone) RemovePlayer(player *Player) {
	if player.CurrentZone != z.Name {
		return
	}

	delete(z.Players, player.Id)
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

	for _, e := range z.Entities {
		info := e.GetInfo()
		if info.Blocking && x == info.X && y == info.Y {
			return
		}
	}

	for _, p := range z.Players {
		if x == p.X && y == p.Y {
			return
		}
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

	log.Printf("[rpg/zone/%s/use] using ent %d", z.Name, entId)

	needsUpdate, err := ent.Use(player)
	if err != nil {
		log.Printf("[rpg/zone/%s/use] failed to use ent %d (%s): %v", z.Name, entId, ent.Type, err)
	}

	return needsUpdate
}
