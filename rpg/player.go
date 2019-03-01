package rpg

var playerSlots = []string{"head", "torso", "legs", "rhand", "lhand"}

type Player struct {
	Id        int              `json:"id"`
	Name      string           `json:"name"`
	Slots     map[string]*Item `json:"slots"`
	Inventory map[int]*Item    `json:"inventory"`

	CurrentZone string `json:"-"`
	X           int    `json:"-"`
	Y           int    `json:"-"`

	DisplayData PlayerDisplayData `json:"-"`
}

type PlayerDisplayData struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

func (p *Player) UpdateDisplay() {
	p.DisplayData = PlayerDisplayData{
		Id:   p.Id,
		Name: p.Name,
		X:    p.X,
		Y:    p.Y,
	}
}

type PlayerInfo struct {
	Id        int                 `json:"id"`
	Name      string              `json:"name"`
	Slots     map[string]ItemInfo `json:"slots"`
	Inventory map[int]ItemInfo    `json:"inventory"`

	X int `json:"x"`
	Y int `json:"y"`
}

func (p *Player) GetInfo() PlayerInfo {
	slots := make(map[string]ItemInfo)
	for slot, i := range p.Slots {
		if p.Slots[slot] != nil {
			slots[slot] = i.GetInfo()
		} else {
			slots[slot] = ItemInfo{Type: "empty"}
		}
	}

	inv := make(map[int]ItemInfo)
	for id, i := range p.Inventory {
		inv[id] = i.GetInfo()
	}

	return PlayerInfo{
		Id:        p.Id,
		Name:      p.Name,
		Slots:     slots,
		Inventory: inv,
		X:         p.X,
		Y:         p.Y,
	}
}
