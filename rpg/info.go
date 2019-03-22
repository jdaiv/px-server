package rpg

type PlayerInfo struct {
	Id        int                  `json:"id"`
	Name      string               `json:"name"`
	Slots     map[string]ItemInfo  `json:"slots"`
	Inventory map[int]ItemInfo     `json:"inventory,omitempty"`
	Spells    map[string]SpellInfo `json:"spells,omitempty"`

	X int `json:"x"`
	Y int `json:"y"`

	HP     int        `json:"hp"`
	MaxHP  int        `json:"maxHP"`
	AP     int        `json:"ap,omitempty"`
	MaxAP  int        `json:"maxAP,omitempty"`
	Stats  StatBlock  `json:"stats,omitempty"`
	Level  int        `json:"level"`
	Skills SkillBlock `json:"skills,omitempty"`
}

func (p Player) GetInfo(base *RPG) PlayerInfo {
	inv := make(map[int]ItemInfo)
	for id, _ := range p.Inventory {
		if item, ok := base.Items.Get(id); ok {
			inv[id] = item.GetInfo()
		}
	}

	spells := make(map[string]SpellInfo)
	for id, s := range p.GetSpells(base.Defs) {
		spells[id] = s.GetInfo()
	}

	return PlayerInfo{
		Id:        p.Id,
		Name:      p.Name,
		Slots:     p.GetSlotInfo(base),
		Inventory: inv,
		Spells:    spells,
		X:         p.X,
		Y:         p.Y,
		HP:        p.HP,
		AP:        p.AP,
		MaxHP:     p.Stats.MaxHP(),
		MaxAP:     p.Stats.MaxAP(),
		Stats:     p.Stats,
		Skills:    p.Skills,
		Level:     p.Skills.TotalLevel(),
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
		Level: p.Skills.TotalLevel(),
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
	Id       int                    `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	X        int                    `json:"x"`
	Y        int                    `json:"y"`
	Usable   bool                   `json:"usable"`
	UseText  string                 `json:"useText"`
	Blocking bool                   `json:"-"`
	Fields   map[string]interface{} `json:"fields"`
}

func (e Entity) GetInfo() EntityInfo {
	exported := make(map[string]interface{})
	for _, f := range e.RootDef.Fields {
		if !f.Export {
			continue
		}
		if v, ok := e.Fields[f.Name]; ok {
			exported[f.Name] = v
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
		Fields:   exported,
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
	HP        int    `json:"hp"`
	MaxHP     int    `json:"maxHP"`
	Alignment string `json:"alignment"`
}

func (n NPC) GetInfo() NPCInfo {
	return NPCInfo{
		Id:        n.Id,
		Name:      n.Name,
		Type:      n.Type,
		X:         n.X,
		Y:         n.Y,
		HP:        n.HP,
		MaxHP:     n.MaxHP,
		Alignment: n.Alignment,
	}
}

type SpellInfo struct {
	Name  string `json:"name"`
	Skill string `json:"skill"`
	Level int    `json:"level"`
	Cost  int    `json:"cost"`
}

func (s SpellDef) GetInfo() SpellInfo {
	return SpellInfo{
		Name:  s.Name,
		Skill: s.Skill,
		Level: s.Level,
		Cost:  s.Cost,
	}
}
