package rpg

import (
	"math/rand"
)

type StatBlock struct {
	AttackPhys     int `json:"attack_phys"`
	AttackMagic    int `json:"attack_magic"`
	DefensePhys    int `json:"defense_phys"`
	DefenseMagic   int `json:"defense_magic"`
	CriticalChance int `json:"critical_chance"`
	Speed          int `json:"speed"`
	Dodge          int `json:"dodge"`
}

type SpecialBlock struct {
	Sunglasses bool `json:"sunglasses,omitempty"`
	Consumable bool `json:"consumable,omitempty"`
}

type SkillBlock struct {
	AttackMelee  Skill `json:"attack_melee"`
	AttackRanged Skill `json:"attack_ranged"`
	DefensePhys  Skill `json:"defense_phys"`
	DefenseMagic Skill `json:"defense_magic"`
	Dodge        Skill `json:"dodge"`
	MagicFire    Skill `json:"magic_fire"`
	MagicIce     Skill `json:"magic_ice"`
	MagicStone   Skill `json:"magic_stone"`
}

type Skill struct {
	Level int `json:"level"`
	XP    int `json:"xp"`
}

func (s SkillBlock) BuildStats() StatBlock {
	return StatBlock{
		AttackPhys:     s.AttackMelee.Level + s.AttackRanged.Level,
		AttackMagic:    s.MagicFire.Level + s.MagicIce.Level + s.MagicStone.Level,
		DefensePhys:    s.DefensePhys.Level * 2,
		DefenseMagic:   s.DefenseMagic.Level * 2,
		CriticalChance: 5,
		Speed:          5,
		Dodge:          s.Dodge.Level,
	}
}

func (s *Skill) AddXP(amt int) {
	s.XP += amt
	for s.XP >= 100 {
		s.Level += 1
		s.XP -= 100
	}
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
	return 10 + s.DefensePhys/2
}

func (s StatBlock) MaxAP() int {
	return 1 + s.Speed
}

func (s SkillBlock) TotalLevel() int {
	return s.AttackMelee.Level +
		s.AttackRanged.Level +
		s.DefensePhys.Level +
		s.DefenseMagic.Level +
		s.Dodge.Level +
		s.MagicFire.Level +
		s.MagicIce.Level +
		s.MagicStone.Level
}

func (s SkillBlock) GetSkillLevel(name string) int {
	switch name {
	case "attack_melee":
		return s.AttackMelee.Level
	case "attack_ranged":
		return s.AttackRanged.Level
	case "defense_phys":
		return s.DefensePhys.Level
	case "defense_magic":
		return s.DefenseMagic.Level
	case "dodge":
		return s.Dodge.Level
	case "magic_fire":
		return s.MagicFire.Level
	case "magic_ice":
		return s.MagicIce.Level
	case "magic_stone":
		return s.MagicStone.Level
	}
	return 0
}

func (s StatBlock) RollPhysDamage() DamageInfo {
	crit := rand.Intn(100) <= s.CriticalChance
	variance := (s.AttackPhys / 4)
	if variance < 2 {
		variance = 2
	}
	dmg := 1 + s.AttackPhys - variance + rand.Intn(variance*2)
	if crit {
		dmg *= 2
	}
	return DamageInfo{dmg, crit, "slash"}
}
