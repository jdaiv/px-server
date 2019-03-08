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
		log.Printf("[rpg/zone/%s/take_item] couldn't find npc %d", z.Name, npcId)
		return
	}

	// if !nextTo(p.X, p.Y, item.X, item.Y) {
	// 	log.Printf("[rpg/zone/%s/take_item] player %d tried to take item %d but was too far away", z.Name, p.Id, itemId)
	// 	return
	// }

	log.Printf("[rpg/zone/%s/attack] attacking %d", z.Name, npcId)

	if !z.CheckAPCost(p, 1) {
		return
	}

	p.Attack(npc)
	z.SendEffect("wood_ex", npc.X, npc.Y)
	z.SendEffect("screen_shake", 8, 8)
}
