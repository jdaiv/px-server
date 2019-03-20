package rpg

func (g *RPG) PlayerMove(p *Player, z *Zone, params ActionParams) {
	direction, ok := params["direction"].(string)
	if !ok {
		return
	}

	x, y, ok := z.Move(p.X, p.Y, direction)
	if !ok {
		return
	}

	if !z.CheckAPCost(p, 1) {
		return
	}
	p.X = x
	p.Y = y
	g.Zones.SetDirty(z.Id)
}
