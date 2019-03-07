package rpg

var playerSlots = []string{"head", "torso", "legs", "hands"}

type Player struct {
	Id        int            `json:"-"`
	Name      string         `json:"-"`
	Slots     map[string]int `json:"-"`
	Inventory map[int]bool   `json:"-"`

	CurrentZone string `json:"currentZone"`
	X           int    `json:"x"`
	Y           int    `json:"y"`

	HP    int       `json:"hp"`
	AP    int       `json:"ap"`
	Stats StatBlock `json:"-"`
}

func (p *Player) Rebuild(base *RPG) {
	base.Items.LoadIntoPlayer(p)
	p.BuildStats(base)
}

func (p *Player) BuildStats(base *RPG) {
	stats := StatBlock{}
	for _, itemId := range p.Slots {
		if itemId <= 0 {
			continue
		}
		item, ok := base.Items.Get(itemId)
		if !ok {
			continue
		}
		stats.Add(item.Stats)
	}
	p.Stats = stats
}

func (p *Player) EquipItem(base *RPG, itemId int) bool {
	_, ok := p.Inventory[itemId]
	if !ok {
		return false
	}

	item, ok := base.Items.Get(itemId)
	if !ok {
		return false
	}

	targetSlot := ""
	switch item.Type {
	case "helmet":
		targetSlot = "head"
	case "melee":
		fallthrough
	case "ranged":
		targetSlot = "hands"
	}

	if targetSlot == "" {
		return false
	}

	if _, equipped := p.Slots[targetSlot]; equipped {
		p.UnequipItem(base, targetSlot)
	}

	item.Equipped = targetSlot
	base.Items.Save(item)
	return true
}

func (p *Player) UnequipItem(base *RPG, slot string) bool {
	id, ok := p.Slots[slot]
	if !ok || id < 0 {
		return false
	}

	item, ok := base.Items.Get(id)
	if !ok {
		return false
	}

	item.Equipped = ""
	base.Items.Save(item)
	return true
}

func (p *Player) DropItem(zone *Zone, itemId int) bool {
	_, ok := p.Inventory[itemId]
	if !ok {
		return false
	}

	zone.AddExistingItem(itemId, p.X, p.Y)
	return true
}
