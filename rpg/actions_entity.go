package rpg

import "log"

func (g *RPG) PlayerUse(p *Player, z *Zone, params ActionParams) {
	entId, ok := params.getInt("id")
	if !ok {
		log.Println("couldn't find ent id param")
		return
	}

	ent, ok := z.Entities[entId]
	if !ok {
		log.Printf("[rpg/zone/%s/use] couldn't find ent %d", z.Name, entId)
		return
	}

	if !nextTo(p.X, p.Y, ent.X, ent.Y) {
		log.Printf("[rpg/zone/%s/use] player %d tried to use ent %d, but was too far away", z.Name, p.Id, entId)
		return
	}

	log.Printf("[rpg/zone/%s/use] using ent %d", z.Name, entId)

	if !p.CheckAPCost(1) {
		return
	}

	needsUpdate, err := g.UseEntity(z, ent, p)
	if err != nil {
		log.Printf("[rpg/zone/%s/use] failed to use ent %d (%s): %v", z.Name, entId, ent.Type, err)
	}

	if needsUpdate {
		g.Zones.SetDirty(z.Id)
	}
}
