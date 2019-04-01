package rpg

import (
	"log"
	"strings"
)

func (g *RPG) HandleEdit(player *Player, zone *Zone, params ActionParams) {
	log.Printf("START EDIT IN %s", zone.Name)

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
		updated = true
	case "disable":
		player.Editing = false
		log.Printf("%s DISABLED", player.Name)
		updated = true
		return
	case "zone_create":
		log.Printf("EDIT TYPE: CREATE ZONE")
		newZone := &Zone{Name: "unnamed"}
		g.InitZone(newZone)
		ok := g.Zones.Insert(newZone)
		if !ok {
			log.Printf("EDIT FAILED: CAN'T CREATE ZONE")
			return
		}
		g.RemovePlayer(zone, player)
		g.AddPlayer(newZone, player, 0, 0)
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
		g.RemovePlayer(zone, player)
		g.AddPlayer(newZone, player, -1, -1)
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
		g.BuildCollisionMap(zone)
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
		if _, err := g.AddEntity(zone, entType, x, y, true); err != nil {
			log.Printf("EDIT FAILED: %v", err)
		} else {
			updated = true
		}
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
		rotation, ok := params.getInt("rotation")
		if !ok {
			log.Printf("EDIT FAILED: INVALID ROTATION")
			return
		}

		if len(name) > 0 {
			ent.Name = name
		} else {
			ent.Name = ent.RootDef.DefaultName
		}
		ent.X = x
		ent.Y = y
		ent.Rotation = rotation
		if ent.Fields == nil {
			ent.Fields = make(EntityFields)
		}
		for k, v := range params {
			if strings.HasPrefix(k, "f_") {
				key := strings.TrimPrefix(k, "f_")
				ent.Fields[key] = v
			}
		}
		g.BuildCollisionMap(zone)

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

		g.RemoveEntity(zone, entId)

		updated = true
	case "clear_corpses":
		log.Printf("EDIT TYPE: CLEAR CORPSES")

		toRemove := make(map[int]bool)
		for id, e := range zone.Entities {
			if e.Type == "corpse" {
				toRemove[id] = true
			}
		}
		for id := range toRemove {
			g.RemoveEntity(zone, id)
		}

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
