package rpg

import "log"

func (g *RPG) PlayerTakeItem(p *Player, z *Zone, params ActionParams) {
	itemId, ok := params.getInt("id")
	if !ok {
		log.Println("couldn't find item id param")
		return
	}

	item, ok := g.Items.GetInZone(itemId, z.Id)
	if !ok {
		log.Printf("[rpg/zone/%s/take_item] couldn't find item %d", z.Name, itemId)
		return
	}

	if !nextTo(p.X, p.Y, item.X, item.Y) {
		log.Printf("[rpg/zone/%s/take_item] player %d tried to take item %d but was too far away", z.Name, p.Id, itemId)
		return
	}

	log.Printf("[rpg/zone/%s/take_item] grabbing item %d", z.Name, itemId)

	if !p.CheckAPCost(1) {
		return
	}

	item.Give(p)
	g.Items.Save(item)
	delete(z.Items, itemId)
	g.BuildPlayer(p)

	g.Zones.SetDirty(z.Id)
}

func (g *RPG) PlayerEquipItem(p *Player, zone *Zone, params ActionParams) {
	itemId, ok := params.getInt("id")
	if !ok {
		log.Println("couldn't find item id param")
		return
	}

	if !p.CheckAPCost(1) {
		return
	}
	g.EquipItem(p, itemId)
	g.BuildPlayer(p)
}

func (g *RPG) PlayerUnequipItem(p *Player, zone *Zone, params ActionParams) {
	slot, ok := params.getString("slot")
	if !ok {
		log.Println("couldn't find item slot param")
		return
	}

	if !p.CheckAPCost(1) {
		return
	}
	g.UnequipItem(p, slot)
	g.BuildPlayer(p)
}

func (g *RPG) PlayerDropItem(p *Player, zone *Zone, params ActionParams) {
	itemId, ok := params.getInt("id")
	if !ok {
		log.Println("couldn't find item id param")
		return
	}

	dropped := g.DropItem(zone, p, itemId)
	if dropped {
		g.Zones.SetDirty(zone.Id)
	}
	g.BuildPlayer(p)
}
