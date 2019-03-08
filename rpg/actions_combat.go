package rpg

import "log"

func (g *RPG) PlayerAttack(p *Player, z *Zone, params ActionParams) {
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

	if !z.CheckAPCost(p, 1) {
		return
	}

	p.Attack(npc)
	z.SendEffect("wood_ex", npc.X, npc.Y)
	z.SendEffect("screen_shake", 8, 8)
}
