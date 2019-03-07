package rpg

type StatBlock struct {
	AttackPhys     int `json:"attack_phys,omitempty"`
	AttackMagic    int `json:"attack_magic,omitempty"`
	DefensePhys    int `json:"defense_phys,omitempty"`
	DefenseMagic   int `json:"defense_magic,omitempty"`
	CriticalChance int `json:"critical_chance,omitempty"`
	Speed          int `json:"speed,omitempty"`
	Dodge          int `json:"dodge,omitempty"`
}

type SpecialBlock struct {
	Sunglasses int `json:"sunglasses,omitempty"`
	Consumable int `json:"consumable,omitempty"`
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

func ConvertStatMap(stats map[string]int) StatBlock {
	block := StatBlock{}
	for stat, value := range stats {
		block.ApplyStat(stat, value)
	}
	return block
}

func (s StatBlock) Add(b StatBlock) StatBlock {
	s.AttackPhys += b.AttackPhys
	s.AttackMagic += b.AttackMagic
	s.DefensePhys += b.DefensePhys
	s.DefenseMagic += b.DefenseMagic
	s.CriticalChance += b.CriticalChance
	s.Speed += b.Speed
	s.Dodge += b.Dodge
	return s
}

func (s StatBlock) MaxHP() int {
	return 10
}

func (s StatBlock) MaxAP() int {
	return 5 + s.Speed
}
