package rpg

var playerSlots = []string{"hat", "core", "book"}

type Player struct {
	Id        int            `json:"-"`
	Name      string         `json:"-"`
	Slots     map[string]int `json:"-"`
	Inventory map[int]bool   `json:"-"`

	CurrentZone int    `json:"currentZone"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Facing      string `json:"facing"`

	HeldPower int    `json:"heldPower"`
	Timers    Timers `json:"timers"`

	Editing bool `json:"-"`
}

const (
	BASE_MP_REGEN = 1
	BASE_HP_REGEN = 8
	MOVE_TIMER    = 1
)

func ValidFace(f string) bool {
	return f == "N" || f == "S" || f == "E" || f == "W"
}

type Timers struct {
	HP   int
	MP   int
	Move int
}

func (g *RPG) BuildPlayer(p *Player) {
	g.Items.LoadIntoPlayer(p)
}

func (g *RPG) EquipItem(p *Player, itemId int) bool {
	_, ok := p.Inventory[itemId]
	if !ok {
		return false
	}

	item, ok := g.Items.Get(itemId)
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
		g.UnequipItem(p, targetSlot)
	}

	item.Equipped = targetSlot
	g.Items.Save(item)
	return true
}

func (g *RPG) UnequipItem(p *Player, slot string) bool {
	id, ok := p.Slots[slot]
	if !ok || id < 0 {
		return false
	}

	item, ok := g.Items.Get(id)
	if !ok {
		return false
	}

	item.Equipped = ""
	g.Items.Save(item)
	return true
}

func (g *RPG) DropItem(zone *Zone, p *Player, itemId int) bool {
	_, ok := p.Inventory[itemId]
	if !ok {
		return false
	}

	g.AddExistingItem(zone, itemId, p.X, p.Y)
	return true
}
