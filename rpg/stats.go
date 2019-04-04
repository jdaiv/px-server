package rpg

import (
	"log"
	"math/rand"
)

type StatBlock struct {
	AttackPhys     int `json:"attack_phys"`
	AttackMagic    int `json:"attack_magic"`
	Defence        int `json:"defence"`
	CriticalChance int `json:"critical_chance"`
	Speed          int `json:"speed"`
	MaxHP          int `json:"maxHP"`
	MaxAP          int `json:"maxAP"`
	MaxMP          int `json:"maxMP"`
}

type SpecialBlock struct {
	Sunglasses bool `json:"sunglasses,omitempty"`
	Consumable bool `json:"consumable,omitempty"`
}

type SkillBlock struct {
	Attack  Skill `json:"attack"`
	Defence Skill `json:"defence"`
	Speed   Skill `json:"speed"`
	Magic   Skill `json:"magic"`
}

type Skill struct {
	Level int `json:"level"`
	XP    int `json:"xp"`
}

func (s SkillBlock) BuildStats() StatBlock {
	return StatBlock{
		AttackPhys:     5 + s.Attack.Level*2,
		AttackMagic:    1 + s.Magic.Level*2,
		Defence:        4 + s.Defence.Level*2,
		CriticalChance: 5 + s.Attack.Level/10,
		Speed:          5 + s.Speed.Level,
		MaxHP:          10 + s.Defence.Level*4,
		MaxAP:          6 + s.Speed.Level/2,
		MaxMP:          5 + s.Magic.Level*5,
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
	s.Defence += b.Defence
	s.CriticalChance += b.CriticalChance
	s.Speed += b.Speed
	s.MaxHP += b.MaxHP
	s.MaxAP += b.MaxAP
	s.MaxMP += b.MaxMP
	return s
}

func (s SkillBlock) TotalLevel() int {
	return s.Attack.Level +
		s.Defence.Level +
		s.Speed.Level +
		s.Magic.Level
}

func (s SkillBlock) GetSkillLevel(name string) int {
	switch name {
	case "attack":
		return s.Attack.Level
	case "Defence":
		return s.Defence.Level
	case "speed":
		return s.Speed.Level
	case "magic":
		return s.Magic.Level
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
	return DamageInfo{dmg, crit, false}
}

func (s StatBlock) RollDefence(dmg DamageInfo) DamageInfo {
	if dmg.Crit {
		return dmg
	}
	if dmg.Amount < s.Defence/4 {
		dmg.Amount = 0
	} else {
		if dmg.Magic {
			dmg.Amount -= s.AttackMagic / 10
		}
		dmg.Amount -= s.Defence / 6
	}
	return dmg
}

type XPEvent int

const (
	EVENT_PHYS_DAMAGE = iota
	EVENT_MAGIC_DAMAGE
	EVENT_PHYS_DEFENCE
	EVENT_MAGIC_DEFENCE
	EVENT_DODGE
)

func (g *RPG) RecordEvent(p *Player, opposingLevel int, actionType XPEvent, amt int) {
	mod := float64(opposingLevel-p.Skills.TotalLevel())/10.0 + 1
	if mod < 0 {
		return
	}
	total := mod * float64(amt)
	log.Printf("Awarding %s %f XP", p.Name, total)
	switch actionType {
	case EVENT_PHYS_DAMAGE:
		p.Skills.Attack.AddXP(int(total))
	case EVENT_MAGIC_DAMAGE:
		p.Skills.Magic.AddXP(int(total))
	case EVENT_PHYS_DEFENCE:
		p.Skills.Defence.AddXP(int(total))
	case EVENT_MAGIC_DEFENCE:
		p.Skills.Defence.AddXP(int(total * 0.7))
		p.Skills.Magic.AddXP(int(total * 0.4))
	case EVENT_DODGE:
		p.Skills.Defence.AddXP(int(total * 0.2))
		p.Skills.Speed.AddXP(int(total * 0.9))
	}
}
