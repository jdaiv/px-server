package rpg

type tile struct {
	Tile        int  `json:"id"`
	Blocking    bool `json:"-"`
	BlockingEnt bool `json:"-"`
}

type ZoneMap struct {
	Width  int
	Height int
	Tiles  []*tile
}

func NewZoneMap(width, height int, fill TileDef) *ZoneMap {
	newMap := ZoneMap{
		width, height, make([]*tile, width*height),
	}

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			newMap.SetTile(x, y, fill)
		}
	}

	return &newMap
}

func (m *ZoneMap) ClampedCoords(x, y int) int {
	if x < 0 {
		x = 0
	}
	if x >= m.Width {
		x = m.Width - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= m.Height {
		y = m.Height - 1
	}
	return y*m.Width + x
}

func (m *ZoneMap) SetTile(x, y int, t TileDef) {
	i := m.ClampedCoords(x, y)
	m.Tiles[i] = &tile{t.Id, t.Blocking, false}
}

func (m *ZoneMap) SetBlocking(x, y int, blocking bool) {
	i := m.ClampedCoords(x, y)
	m.Tiles[i].BlockingEnt = true
}

func (m *ZoneMap) IsBlocking(x, y int) bool {
	if x < 0 || x >= m.Width ||
		y < 0 || y >= m.Height {
		return true
	}
	i := m.ClampedCoords(x, y)
	t := m.Tiles[i]
	return t.Blocking || t.BlockingEnt
}

func (m *ZoneMap) GetTile(x, y int) *tile {
	return m.Tiles[m.ClampedCoords(x, y)]
}
