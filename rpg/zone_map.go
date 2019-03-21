package rpg

const OFFSET = int(int32(^uint32(0)>>1) / 2)

func compactCoords(x, y int) uint64 {
	_x := uint64(x + OFFSET)
	_y := uint64(y + OFFSET)
	return (_x << 32) ^ _y
}

func uncompactCoords(c uint64) (int, int) {
	x := int(c >> 32)
	y := int(c & 0xFFFFFFFF)
	return (x - OFFSET), (y - OFFSET)
}

type tile struct {
	Tile     int  `json:"id"`
	Blocking bool `json:"blocking"`
}

type ZoneMap struct {
	MinX  int
	MaxX  int
	MinY  int
	MaxY  int
	Tiles map[uint64]tile
}

func NewZoneMap(width, height int, fill TileDef) *ZoneMap {
	newMap := ZoneMap{
		Tiles: make(map[uint64]tile),
	}

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			newMap.SetTile(x, y, fill)
		}
	}

	return &newMap
}

func (m *ZoneMap) SetTile(x, y int, t TileDef) {
	if x < m.MinX {
		m.MinX = x
	}
	if x > m.MaxX {
		m.MaxX = x
	}
	if y < m.MinY {
		m.MinY = y
	}
	if y > m.MaxY {
		m.MaxY = y
	}
	m.Tiles[compactCoords(x, y)] = tile{t.Id, t.Blocking}
}

func (m *ZoneMap) SetBlocking(x, y int, blocking bool) {
	if t, ok := m.Tiles[compactCoords(x, y)]; ok {
		t.Blocking = true
	}
}

func (m *ZoneMap) IsBlocking(x, y int) bool {
	t, ok := m.Tiles[compactCoords(x, y)]
	return !ok || t.Blocking
}

func (m *ZoneMap) GetTile(x, y int) (tile, bool) {
	t, ok := m.Tiles[compactCoords(x, y)]
	return t, ok
}
