package rpg

var playerSlots = []string{"head", "torso", "legs", "hands"}

type Player struct {
	Id        int            `json:"-"`
	Name      string         `json:"-"`
	Slots     map[string]int `json:"-"`
	Inventory map[int]bool   `json:"-"`

	CurrentZone int `json:"currentZone"`
	X           int `json:"x"`
	Y           int `json:"y"`

	HP     int        `json:"hp"`
	AP     int        `json:"ap"`
	Stats  StatBlock  `json:"stats"`
	Skills SkillBlock `json:"skills"`
	Timers Timers     `json:"timers"`

	Editing bool `json:"-"`
}

const (
	BASE_AP_REGEN = 1
	BASE_HP_REGEN = 8
)

type Timers struct {
	HP int
	AP int
}

func (p *Player) Rebuild(base *RPG) {
	base.Items.LoadIntoPlayer(p)
	p.BuildStats(base)
}

func (p *Player) BuildStats(base *RPG) {
	stats := p.Skills.BuildStats()
	for _, itemId := range p.Slots {
		if itemId < 0 {
			continue
		}
		item, ok := base.Items.Get(itemId)
		if !ok {
			continue
		}
		stats = stats.Add(item.Stats)
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

func (p *Player) GetSpells(defs *Definitions) map[string]SpellDef {
	spells := make(map[string]SpellDef)
	for n, s := range defs.Spells {
		req := p.Skills.GetSkillLevel(s.Skill)
		if req >= s.Level {
			spells[n] = s
		}
	}
	return spells
}
