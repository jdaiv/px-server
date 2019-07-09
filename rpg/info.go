package rpg

type PlayerInfo struct {
	Id        int                 `json:"id"`
	Name      string              `json:"name"`
	Slots     map[string]ItemInfo `json:"slots"`
	Inventory map[int]ItemInfo    `json:"inventory,omitempty"`

	X      int    `json:"x"`
	Y      int    `json:"y"`
	Facing string `json:"facing"`

	HP    int `json:"hp"`
	MaxHP int `json:"maxHP"`
	MP    int `json:"mp,omitempty"`
	MaxMP int `json:"maxMP,omitempty"`
	Level int `json:"level"`
}

func (p *Player) GetInfo(base *RPG) PlayerInfo {
	inv := make(map[int]ItemInfo)
	for id, _ := range p.Inventory {
		if item, ok := base.Items.Get(id); ok {
			inv[id] = item.GetInfo()
		}
	}

	slots, powerLevel := p.GetSlotInfo(base)
	return PlayerInfo{
		Id:        p.Id,
		Name:      p.Name,
		Slots:     slots,
		Inventory: inv,
		X:         p.X,
		Y:         p.Y,
		Facing:    p.Facing,
		Level:     powerLevel,
	}
}

func (p Player) GetInfoPublic(base *RPG) PlayerInfo {
	slots, powerLevel := p.GetSlotInfo(base)
	return PlayerInfo{
		Id:     p.Id,
		Name:   p.Name,
		Slots:  slots,
		X:      p.X,
		Y:      p.Y,
		Facing: p.Facing,
		Level:  powerLevel,
	}
}

func (p Player) GetSlotInfo(base *RPG) (map[string]ItemInfo, int) {
	slots := make(map[string]ItemInfo)
	powerLevel := 0
	for slot, id := range p.Slots {
		if p.Slots[slot] != -1 {
			if item, ok := base.Items.Get(id); ok {
				slots[slot] = item.GetInfo()
				powerLevel += slots[slot].PowerLevel
			}
		} else {
			slots[slot] = ItemInfo{Type: "empty"}
		}
	}
	return slots, powerLevel
}

type EntityInfo struct {
	Id       int                    `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	X        int                    `json:"x"`
	Y        int                    `json:"y"`
	Rotation int                    `json:"rotation"`
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
		Rotation: e.Rotation,
		Usable:   e.RootDef.Usable,
		UseText:  e.RootDef.UseText,
		Blocking: e.RootDef.Blocking,
		Fields:   exported,
	}
}

type ItemInfo struct {
	Id         int    `json:"id"`
	Quality    int    `json:"quality"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	PowerLevel int    `json:"powerLevel"`

	X int `json:"x"`
	Y int `json:"y"`
}

func (i Item) GetInfo() ItemInfo {
	return ItemInfo{
		Id:         i.Id,
		Name:       i.Name,
		Quality:    i.Quality,
		Type:       i.Type,
		PowerLevel: i.PowerLevel,
		X:          i.X,
		Y:          i.Y,
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
		Id:   n.Id,
		Name: n.Name,
		Type: n.Type,
		X:    n.X,
		Y:    n.Y,
	}
}
