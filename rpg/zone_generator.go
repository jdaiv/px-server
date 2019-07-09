package rpg

import (
	simplex "github.com/ojrac/opensimplex-go"
)

const (
	SEED                = 6942069
	OVERWORLD_ZONE_SIZE = 32
)

var noise = simplex.New(SEED)

func (g *RPG) GenerateOverworldTile(x, y int, m *ZoneMap) {
	grass := g.Defs.Tiles[3]
	water := g.Defs.Tiles[4]
	tree := g.Defs.Tiles[5]
	// rock := g.Defs.Tiles[6]

	for x := 0; x < OVERWORLD_ZONE_SIZE; x++ {
		for y := 0; y < OVERWORLD_ZONE_SIZE; y++ {
			_x := float64(x) / float64(OVERWORLD_ZONE_SIZE/8)
			_y := float64(y) / float64(OVERWORLD_ZONE_SIZE/8)
			value := noise.Eval2(_x, _y)

			if value < -0.4 {
				m.SetTile(x, y, water)
			} else if value > 0.4 {
				m.SetTile(x, y, tree)
			} else {
				m.SetTile(x, y, grass)
			}
		}
	}
}
