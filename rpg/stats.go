package rpg

type StatBlock struct {
	AttackPhys     int `json:"attack_phys"`
	AttackMagic    int `json:"attack_magic"`
	DefensePhys    int `json:"defense_phys"`
	DefenseMagic   int `json:"defense_magic"`
	CriticalChance int `json:"critical_chance"`
	Speed          int `json:"speed"`
	Dodge          int `json:"dodge"`
}

func (s *StatBlock) ApplyStat(stat string, value int) {
	switch stat {
	case "attack_phys":
		s.AttackPhys += value
	case "attack_magic":
		s.AttackMagic += value
	case "defense_phys":
		s.DefensePhys += value
	case "defense_magic":
		s.DefenseMagic += value
	case "critical_chance":
		s.CriticalChance += value
	case "speed":
		s.Speed += value
	case "dodge":
		s.Dodge += value
	}
}

func (p *Player) BuildStats() {
	stats := StatBlock{}
	for _, item := range p.Slots {
		if item == nil {
			continue
		}
		for stat, value := range item.Stats {
			stats.ApplyStat(stat, value)
		}
	}
	p.Stats = stats
}
