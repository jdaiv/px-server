package rpg

import "log"

func (g *RPG) PlayerTakeItem(p *Player, z *Zone, params ActionParams) {
	itemId, ok := params.getInt("id")
	if !ok {
		log.Println("couldn't find item id param")
		return
	}

	item, ok := z.Parent.Items.GetInZone(itemId, z.Name)
	if !ok {
		log.Printf("[rpg/zone/%s/take_item] couldn't find item %d", z.Name, itemId)
		return
	}

	if !nextTo(p.X, p.Y, item.X, item.Y) {
		log.Printf("[rpg/zone/%s/take_item] player %d tried to take item %d but was too far away", z.Name, p.Id, itemId)
		return
	}

	log.Printf("[rpg/zone/%s/take_item] grabbing item %d", z.Name, itemId)

	if !z.CheckAPCost(p, 1) {
		return
	}

	item.Give(p)
	g.Items.Save(item)
	delete(z.Items, itemId)
	p.Rebuild(g)

	z.Dirty = true
}

func (g *RPG) PlayerEquipItem(p *Player, zone *Zone, params ActionParams) {
	itemId, ok := params.getInt("id")
	if !ok {
		log.Println("couldn't find item id param")
		return
	}

	if !zone.CheckAPCost(p, 1) {
		return
	}
	p.EquipItem(g, itemId)
	p.Rebuild(g)
}

func (g *RPG) PlayerUnequipItem(p *Player, zone *Zone, params ActionParams) {
	slot, ok := params.getString("slot")
	if !ok {
		log.Println("couldn't find item slot param")
		return
	}

	if !zone.CheckAPCost(p, 1) {
		return
	}
	p.UnequipItem(g, slot)
	p.Rebuild(g)
}

func (g *RPG) PlayerDropItem(p *Player, zone *Zone, params ActionParams) {
	itemId, ok := params.getInt("id")
	if !ok {
		log.Println("couldn't find item id param")
		return
	}

	dropped := p.DropItem(zone, itemId)
	zone.Dirty = zone.Dirty || dropped
	p.Rebuild(g)
}
