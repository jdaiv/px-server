package rpg

import (
	"math"
)

func (g *RPG) DoSpellAttack(z *Zone, origin *Player, spell SpellDef, x, y int) {
	seq := NewSequence()
	for _, effect := range spell.Effects {
		switch effect.Type {
		case "effect":
			seq.AddSpellEffect(effect.Effect, origin.Id, false, x, y, effect.Duration)
		case "aoe":
			dmg := DamageInfo{effect.Damage, false, true}
			for _, n := range z.NPCs {
				dist := math.Sqrt(math.Pow(float64(n.X-x), 2) + math.Pow(float64(n.Y-y), 2))
				if dist <= float64(effect.Range) {
					seq.AddDamage(dmg, n.Id, true)
				}
			}
			if effect.Effect != "" {
				r := float64(effect.Range)
				for _x := -r; _x <= r; _x += 1 {
					for _y := -r; _y <= r; _y += 1 {
						dist := math.Sqrt(math.Pow(_x, 2) + math.Pow(_y, 2))
						if dist <= r {
							seq.AddEffect(effect.Effect, x+int(_x), y+int(_y), effect.Duration)
						}
					}
				}
			}
		}
	}

	z.CombatInfo.CurrentSequence = seq
}
