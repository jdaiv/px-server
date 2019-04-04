package rpg

import "math/rand"

var playerSlots = []string{"head", "torso", "legs", "hands"}

type Player struct {
	Id        int            `json:"-"`
	Name      string         `json:"-"`
	Slots     map[string]int `json:"-"`
	Inventory map[int]bool   `json:"-"`

	CurrentZone int    `json:"currentZone"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Facing      string `json:"facing"`

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

func ValidFace(f string) bool {
	return f == "N" || f == "S" || f == "E" || f == "W"
}

type Timers struct {
	HP int
	AP int
}

func (g *RPG) BuildPlayer(p *Player) {
	g.Items.LoadIntoPlayer(p)
	stats := p.Skills.BuildStats()
	for _, itemId := range p.Slots {
		if itemId < 0 {
			continue
		}
		item, ok := g.Items.Get(itemId)
		if !ok {
			continue
		}
		stats = stats.Add(item.Stats)
	}
	p.Stats = stats
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

func (g *RPG) GetSpellsFor(p *Player) map[string]SpellDef {
	spells := make(map[string]SpellDef)
	for n, s := range g.Defs.Spells {
		req := p.Skills.Magic.Level
		if req >= s.Level {
			spells[n] = s
		}
	}
	return spells
}

func (p *Player) GetName() string {
	return p.Name
}

func (p *Player) InitCombat() CombatInfo {
	return CombatInfo{
		Initiative: rand.Intn(20) + p.Stats.Speed/2,
		IsPlayer:   true,
		Id:         p.Id,
	}
}

func (p *Player) Attack() DamageInfo {
	return p.Stats.RollPhysDamage()
}

func (p *Player) Damage(dmg DamageInfo) DamageInfo {
	def := p.Stats.RollDefence(dmg)
	p.HP -= def.Amount
	return def
}

func (p *Player) NewTurn(ci *CombatInfo) {
	p.AP = p.Stats.MaxAP
	ci.Timer = MAX_PLAYER_TURN_TIME
}

func (p *Player) Tick(g *RPG, z *Zone, ci *CombatInfo) {
	ci.Timer -= 1
}

func (p *Player) IsTurnOver(ci *CombatInfo) bool {
	return ci.Timer <= 0 || p.AP <= 0
}
