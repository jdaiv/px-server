package rpg

type Item struct {
	Id           int          `json:"-"`
	Quality      int          `json:"quality"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	MaxQty       int          `json:"maxQty"`
	Durability   int          `json:"durability"`
	Price        int          `json:"price"`
	Stats        StatBlock    `json:"stats"`
	SpecialAttrs SpecialBlock `json:"specials"`

	Held     bool   `json:"held"`
	HeldBy   int    `json:"heldBy"`
	Equipped string `json:"equipped"`

	CurrentZone string `json:"currentZone"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
}

func (i *Item) ApplyMod(def ItemModDef) {
	i.Name = def.Name + " " + i.Name
	// i.ApplyStats(def.Stats)
}

func (i *Item) Give(player *Player) {
	i.Held = true
	i.HeldBy = player.Id
	i.Equipped = ""
	i.CurrentZone = ""
}
