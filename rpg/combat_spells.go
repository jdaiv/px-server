package rpg

import (
	"math"
)

type CombatSpellData struct {
	SpellCasting bool     `json:"-"`
	CurrentSpell SpellDef `json:"-"`
	SpellTimer   int      `json:"-"`
	SpellStage   int      `json:"-"`
	SourceX      int      `json:"-"`
	SourceY      int      `json:"-"`
	TargetX      int      `json:"-"`
	TargetY      int      `json:"-"`
}

func (d *CombatSpellData) RunSpell(z *Zone, spell SpellDef, sX, sY, tX, tY int) bool {
	if len(spell.Effects) <= 0 {
		return false
	}
	d.SpellCasting = true
	d.CurrentSpell = spell
	d.SpellTimer = 0
	d.SpellStage = 0
	d.SourceX = sX
	d.SourceY = sY
	d.TargetX = tX
	d.TargetY = tY
	return d.Tick(z)
}

func (d *CombatSpellData) Tick(z *Zone) bool {
	if !d.SpellCasting {
		return false
	}

	if d.SpellTimer <= 0 {
		for d.SpellTimer <= 0 {
			if d.SpellStage >= len(d.CurrentSpell.Effects) {
				d.SpellCasting = false
				return false
			}
			effect := d.CurrentSpell.Effects[d.SpellStage]
			switch effect.Type {
			case "effect":
				eParams := effectParams{
					"sourceX": d.SourceX,
					"sourceY": d.SourceY,
					"targetX": d.TargetX,
					"targetY": d.TargetY,
				}
				z.SendEffect(effect.Effect, eParams)
			case "aoe":
				for _, n := range z.NPCs {
					dist := math.Sqrt(math.Pow(float64(n.X-d.TargetX), 2) +
						math.Pow(float64(n.Y-d.TargetY), 2))
					if dist <= float64(effect.Range) {
						n.Damage(DamageInfo{effect.Damage, false, "spell"})
					}
				}
				if effect.Effect != "" {
					r := float64(effect.Range)
					for x := -r; x <= r; x += 1 {
						for y := -r; y <= r; y += 1 {
							dist := math.Sqrt(math.Pow(float64(x), 2) + math.Pow(float64(y), 2))
							if dist <= r {
								z.SendEffect(effect.Effect, effectParams{
									"x": d.TargetX + int(x),
									"y": d.TargetY + int(y),
								})
							}
						}
					}
				}
			}
			d.SpellTimer = effect.Duration
			d.SpellStage += 1
		}
	} else {
		d.SpellTimer -= 1
	}

	return true
}
