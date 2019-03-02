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
	Id    int                 `json:"id"`
	Name  string              `json:"name"`
	Slots map[string]ItemInfo `json:"slots"`
	X     int                 `json:"x"`
	Y     int                 `json:"y"`
}

func (p *Player) UpdateDisplay() {
	p.DisplayData = PlayerDisplayData{
		Id:    p.Id,
		Name:  p.Name,
		Slots: p.GetInfo().Slots,
		X:     p.X,
		Y:     p.Y,
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

func (p *Player) EquipItem(itemId int) bool {
	item, ok := p.Inventory[itemId]
	if !ok || item == nil {
		return false
	}

	targetSlot := ""
	switch item.Type {
	case "helmet":
		targetSlot = "head"
	}

	if targetSlot == "" {
		return false
	}

	if _, equipped := p.Slots[targetSlot]; equipped {
		p.UnequipItem(targetSlot)
	}

	item.Equipped = targetSlot
	delete(p.Inventory, item.Id)
	p.Slots[targetSlot] = item
	item.Save()

	return true
}

func (p *Player) UnequipItem(slot string) bool {
	item, ok := p.Slots[slot]
	if !ok || item == nil {
		return false
	}

	item.Equipped = ""
	p.Inventory[item.Id] = item
	p.Slots[slot] = nil
	item.Save()

	return true
}

func (p *Player) DropItem(zone *Zone, itemId int) bool {
	item, ok := p.Inventory[itemId]
	if !ok {
		return false
	}

	delete(p.Inventory, itemId)
	zone.AddExistingItem(item, p.X, p.Y)

	return true
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
