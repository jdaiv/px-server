package rpg

import (
	"log"
	"strings"
)

func (g *RPG) HandleEdit(player *Player, zone *Zone, params ActionParams) {
	log.Printf("START EDIT")

	editType, ok := params.getString("type")
	if !ok {
		log.Printf("EDIT FAILED: MISSING TYPE")
		return
	}

	if !player.Editing && editType != "enable" {
		return
	}

	updated := false

	log.Printf("EDIT DATA: %v", params)
	switch editType {
	case "enable":
		player.Editing = true
		log.Printf("%s ENABLED", player.Name)
	case "disable":
		player.Editing = false
		log.Printf("%s DISABLED", player.Name)
		return
	case "zone_create":
		log.Printf("EDIT TYPE: CREATE ZONE")
		newZone := &Zone{Name: "unnamed"}
		newZone.Init(g)
		ok := g.Zones.Insert(newZone)
		if !ok {
			log.Printf("EDIT FAILED: CAN'T CREATE ZONE")
			return
		}
		zone.RemovePlayer(player)
		newZone.AddPlayer(player, -1, -1)
		g.Zones.SetDirty(newZone.Id)
		g.Outgoing <- OutgoingMessage{
			Zone: newZone.Id,
			Type: ACTION_UPDATE,
		}
		updated = true
	case "zone_goto":
		log.Printf("EDIT TYPE: GOTO ZONE")
		to, ok := params.getInt("zone")
		if !ok {
			log.Printf("EDIT FAILED: INVALID ZONE ID")
			return
		}
		newZone, ok := g.Zones.Get(to)
		if !ok {
			log.Printf("EDIT FAILED: INVALID ZONE")
			return
		}
		zone.RemovePlayer(player)
		newZone.AddPlayer(player, -1, -1)
		g.Zones.SetDirty(newZone.Id)
		g.Outgoing <- OutgoingMessage{
			Zone: newZone.Id,
			Type: ACTION_UPDATE,
		}
		updated = true
	case "tile":
		log.Printf("EDIT TYPE: TILE")
		x, ok := params.getInt("x")
		if !ok {
			log.Printf("EDIT FAILED: INVALID X")
			return
		}
		y, ok := params.getInt("y")
		if !ok {
			log.Printf("EDIT FAILED: INVALID Y")
			return
		}
		to, ok := params.getInt("to")
		if !ok || to < 0 || to >= len(g.Defs.Tiles) {
			log.Printf("EDIT FAILED: INVALID TILE ID")
			return
		}
		tile := g.Defs.Tiles[to]
		zone.Map.SetTile(x, y, tile)
		zone.BuildCollisionMap()
		updated = true
	case "entity_create":
		log.Printf("EDIT TYPE: CREATE ENTITY")
		entType, ok := params.getString("ent")
		if !ok {
			log.Printf("EDIT FAILED: INVALID TYPE")
			return
		}
		x, ok := params.getInt("x")
		if !ok {
			log.Printf("EDIT FAILED: INVALID X")
			return
		}
		y, ok := params.getInt("y")
		if !ok {
			log.Printf("EDIT FAILED: INVALID Y")
			return
		}
		zone.AddEntity(entType, x, y, true)
		updated = true
	case "entity_edit":
		log.Printf("EDIT TYPE: EDIT ENTITY")
		entId, ok := params.getInt("ent")
		if !ok {
			log.Printf("EDIT FAILED: INVALID ID")
			return
		}
		ent, ok := zone.Entities[entId]
		if !ok {
			log.Printf("EDIT FAILED: INVALID ENTITY")
			return
		}
		name, ok := params.getString("name")
		if !ok {
			log.Printf("EDIT FAILED: INVALID NAME")
			return
		}
		x, ok := params.getInt("x")
		if !ok {
			log.Printf("EDIT FAILED: INVALID X")
			return
		}
		y, ok := params.getInt("y")
		if !ok {
			log.Printf("EDIT FAILED: INVALID Y")
			return
		}

		if len(name) > 0 {
			ent.Name = name
		} else {
			ent.Name = ent.RootDef.DefaultName
		}
		ent.X = x
		ent.Y = y
		if ent.Fields == nil {
			ent.Fields = make(EntityFields)
		}
		for k, v := range params {
			if strings.HasPrefix(k, "f_") {
				key := strings.TrimPrefix(k, "f_")
				ent.Fields[key] = v
			}
		}
		zone.BuildCollisionMap()

		updated = true
	case "entity_delete":
		log.Printf("EDIT TYPE: DELETE ENTITY")
		entId, ok := params.getInt("ent")
		if !ok {
			log.Printf("EDIT FAILED: INVALID ID")
			return
		}
		if _, ok := zone.Entities[entId]; !ok {
			log.Printf("EDIT FAILED: INVALID ENTITY")
			return
		}

		zone.RemoveEntity(entId)

		updated = true
	}

	log.Printf("EDIT SUCCESS: %v", updated)

	if updated {
		g.Zones.SetDirty(zone.Id)
		g.Outgoing <- OutgoingMessage{
			Zone: zone.Id,
			Type: ACTION_UPDATE,
		}
	}
}
