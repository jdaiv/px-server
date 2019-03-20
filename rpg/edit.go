package rpg

import (
	"log"
	"strings"
)

func (g *RPG) HandleEdit(zone *Zone, params ActionParams) {
	log.Printf("START EDIT")

	editType, ok := params.getString("type")
	if !ok {
		log.Printf("EDIT FAILED: MISSING TYPE")
		return
	}

	log.Printf("EDIT DATA: %v", params)

	updated := false
	switch editType {
	case "tile":
		log.Printf("EDIT TYPE: TILE")
		idx, ok := params.getInt("idx")
		if !ok || idx < 0 || idx >= len(zone.Map) {
			log.Printf("EDIT FAILED: INVALID INDEX")
			return
		}
		to, ok := params.getInt("to")
		if !ok || to < 0 || to >= len(g.Defs.Tiles) {
			log.Printf("EDIT FAILED: INVALID TILE ID")
			return
		}
		zone.Map[idx] = to
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
		if !ok || x < 0 || x >= zone.Width {
			log.Printf("EDIT FAILED: INVALID X")
			return
		}
		y, ok := params.getInt("y")
		if !ok || y < 0 || y >= zone.Height {
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
		if !ok || x < 0 || x >= zone.Width {
			log.Printf("EDIT FAILED: INVALID X")
			return
		}
		y, ok := params.getInt("y")
		if !ok || y < 0 || y >= zone.Height {
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
