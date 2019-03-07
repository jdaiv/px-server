package rpg

type PlayerInfo struct {
	Id        int                 `json:"id"`
	Name      string              `json:"name"`
	Slots     map[string]ItemInfo `json:"slots"`
	Inventory map[int]ItemInfo    `json:"inventory"`

	X int `json:"x"`
	Y int `json:"y"`

	HP    int       `json:"hp"`
	MaxHP int       `json:"maxHP"`
	AP    int       `json:"ap"`
	MaxAP int       `json:"maxAP"`
	Stats StatBlock `json:"stats"`
}

func (p Player) GetInfo(base *RPG) PlayerInfo {
	inv := make(map[int]ItemInfo)
	for id, _ := range p.Inventory {
		if item, ok := base.Items.Get(id); ok {
			inv[id] = item.GetInfo()
		}
	}

	return PlayerInfo{
		Id:        p.Id,
		Name:      p.Name,
		Slots:     p.GetSlotInfo(base),
		Inventory: inv,
		X:         p.X,
		Y:         p.Y,
		HP:        p.HP,
		AP:        p.AP,
		MaxHP:     p.Stats.MaxHP(),
		MaxAP:     p.Stats.MaxAP(),
		Stats:     p.Stats,
	}
}

func (p Player) GetInfoPublic(base *RPG) PlayerInfo {
	return PlayerInfo{
		Id:    p.Id,
		Name:  p.Name,
		Slots: p.GetSlotInfo(base),
		X:     p.X,
		Y:     p.Y,
		HP:    p.HP,
		MaxHP: p.Stats.MaxHP(),
	}
}

func (p Player) GetSlotInfo(base *RPG) map[string]ItemInfo {
	slots := make(map[string]ItemInfo)
	for slot, id := range p.Slots {
		if p.Slots[slot] != -1 {
			if item, ok := base.Items.Get(id); ok {
				slots[slot] = item.GetInfo()
			}
		} else {
			slots[slot] = ItemInfo{Type: "empty"}
		}
	}
	return slots
}

type EntityInfo struct {
	Id       int               `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	X        int               `json:"x"`
	Y        int               `json:"y"`
	Usable   bool              `json:"usable"`
	UseText  string            `json:"useText"`
	Blocking bool              `json:"-"`
	Strings  map[string]string `json:"strings"`
}

func (e Entity) GetInfo() EntityInfo {
	exportedStrings := make(map[string]string)
	for _, key := range e.RootDef.ExportStrings {
		if v, ok := e.Def.Strings[key]; ok {
			exportedStrings[key] = v
		}
	}
	return EntityInfo{
		Id:       e.Id,
		Name:     e.Name,
		Type:     e.Type,
		X:        e.X,
		Y:        e.Y,
		Usable:   e.RootDef.Usable,
		UseText:  e.RootDef.UseText,
		Blocking: e.RootDef.Blocking,
		Strings:  exportedStrings,
	}
}

type ItemInfo struct {
	Id           int          `json:"id"`
	Quality      int          `json:"quality"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Durability   int          `json:"durability"`
	Price        int          `json:"price"`
	Stats        StatBlock    `json:"stats"`
	SpecialAttrs SpecialBlock `json:"specials"`

	X int `json:"x"`
	Y int `json:"y"`
}

func (i Item) GetInfo() ItemInfo {
	return ItemInfo{
		Id:           i.Id,
		Name:         i.Name,
		Quality:      i.Quality,
		Type:         i.Type,
		Durability:   i.Durability,
		Price:        i.Price,
		SpecialAttrs: i.SpecialAttrs,
		Stats:        i.Stats,
		X:            i.X,
		Y:            i.Y,
	}
}

type NPCInfo struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Alignment string `json:"alignment"`
}

func (n NPC) GetInfo() NPCInfo {
	return NPCInfo{
		Id:        n.Id,
		Name:      n.Name,
		Type:      n.Type,
		X:         n.X,
		Y:         n.Y,
		Alignment: n.Alignment,
	}
}
