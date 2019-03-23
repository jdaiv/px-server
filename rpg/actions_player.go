package rpg

func (g *RPG) PlayerMove(p *Player, z *Zone, params ActionParams) {
	direction, ok := params["direction"].(string)
	if !ok {
		return
	}

	var x int
	var y int
	if p.Editing {
		x, y, ok = z.MoveNoclip(p.X, p.Y, direction)
	} else {
		x, y, ok = z.Move(p.X, p.Y, direction)
	}

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

func (g *RPG) PlayerFace(p *Player, z *Zone, params ActionParams) {
	direction, ok := params.getString("direction")
	if !ok || !ValidFace(direction) {
		return
	}
	p.Facing = direction
	g.Zones.SetDirty(z.Id)
}
