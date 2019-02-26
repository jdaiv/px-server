package rpg

type Tile struct {
	Type     string `json:"type"`
	Blocking bool   `json:"blocking"`
}

func TileFromDef(t string, defs *Definitions) Tile {
	tileDef, ok := defs.Tiles[t]
	if ok {
		return Tile{t, tileDef.Blocking}
	}
	return Tile{"MISSING", false}
}
