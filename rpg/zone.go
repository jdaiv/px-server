package rpg

import "math/rand"

type Zone struct {
	Width  int
	Height int
	Map    []Tile
}

func NewZone(width int, height int) *Zone {
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
		Width:  width,
		Height: height,
		Map:    sampleMap,
	}
}
