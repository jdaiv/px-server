package rpg

import "math/rand"

type Zone struct {
	Name        string
	Width       int
	Height      int
	Map         []Tile
	EntityCount int
	Entities    map[int]Entity
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

func NewZone(name string, width int, height int) *Zone {
	sampleMap := make([]Tile, width*height)
	for i := range sampleMap {
		tileType := "flat"
		if rand.Intn(2) == 1 {
			tileType = "grass"
		}
		sampleMap[i] = Tile{
			Type:  tileType,
			Flags: 0,
		}
	}

	return &Zone{
		Name:     name,
		Width:    width,
		Height:   height,
		Map:      sampleMap,
		Players:  make(map[int]*Player),
		Entities: make(map[int]Entity),
	}
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

func (z *Zone) AddEntity(ent Entity, name string, x, y int) {
	id := z.EntityCount
	z.EntityCount += 1
	ent.Init(z, id, name, x, y)
	z.Entities[id] = ent
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
		if info.Collision && x == info.X && y == info.Y {
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
