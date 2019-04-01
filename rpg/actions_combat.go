package rpg

import "log"

func (g *RPG) PlayerAttack(p *Player, z *Zone, params ActionParams) {
	mode, ok := params.getString("mode")
	if !ok {
		log.Println("couldn't find mode param")
		return
	}

	switch mode {
	case "melee":
		npcId, ok := params.getInt("id")
		if !ok {
			log.Println("couldn't find npc id param")
			return
		}

		npc, ok := z.NPCs[npcId]
		if !ok {
			log.Printf("[rpg/zone/%s/attack] couldn't find npc %d", z.Name, npcId)
			return
		}

		if !nextTo(p.X, p.Y, npc.X, npc.Y) {
			log.Printf("[rpg/zone/%s/attack] player %d tried to attack %d but was too far away", z.Name, p.Id, npcId)
			return
		}

		log.Printf("[rpg/zone/%s/attack] attacking %d", z.Name, npcId)

		if !p.CheckAPCost(1) {
			return
		}

		g.DoMeleeAttack(z, p, npc)
		p.Skills.AttackMelee.AddXP(5)

		g.SendEffect(z, "wood_ex", effectParams{
			"x": npc.X,
			"y": npc.Y,
		})
		g.SendEffect(z, "screen_shake", effectParams{
			"x": 8,
			"y": 8,
		})
	case "spell":
		spellId, ok := params.getString("spell")
		if !ok {
			log.Println("couldn't find spell param")
			return
		}
		x, ok := params.getInt("x")
		if !ok {
			log.Println("couldn't find x param")
			return
		}
		y, ok := params.getInt("y")
		if !ok {
			log.Println("couldn't find y param")
			return
		}
		spell, ok := g.Defs.Spells[spellId]
		if !ok {
			log.Println("couldn't find spell")
			return
		}
		if !p.CheckAPCost(spell.Cost) {
			return
		}
		g.DoSpellAttack(z, p, spell, x, y)
	}
}
