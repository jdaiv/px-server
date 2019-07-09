package rpg

import (
	"errors"
	"fmt"
)

var dirMap = map[int]string{
	0: "N",
	1: "S",
	2: "E",
	3: "W",
}

var npcLogicFuncs = map[string]func(*RPG, *NPC, *Zone) bool{
	"blob": BlobIdle,
}

var npcCombatLogicFuncs = map[string]func(*RPG, *NPC, *Zone) bool{
	"blob": BlobCombat,
}

type NPC struct {
	Id    int
	Name  string
	Type  string
	X     int
	Y     int
	Logic string
}

func (g *RPG) NewNPC(z *Zone, npcType string, x, y int) (*NPC, error) {
	if z.Map.IsBlocking(int(x), int(y)) {
		return nil, fmt.Errorf("coords %d,%d blocked", x, y)
	}

	id := z.NPCCount
	npcDef, ok := g.Defs.NPCs[npcType]
	if !ok {
		return nil, errors.New("npc missing")
	}

	npc := &NPC{
		Id:    id,
		Name:  npcDef.DefaultName,
		Type:  npcType,
		X:     x,
		Y:     y,
		Logic: npcDef.Logic,
	}

	z.NPCs[id] = npc
	z.NPCCount += 1

	g.BuildCollisionMap(z)

	return npc, nil
}

func (g *RPG) RemoveNPC(z *Zone, npcId int) {
	delete(z.NPCs, npcId)
	g.BuildCollisionMap(z)
}

func BlobIdle(g *RPG, self *NPC, zone *Zone) bool {
	return false
}

func BlobCombat(g *RPG, self *NPC, zone *Zone) bool {
	// for _, p := range zone.Players {
	// 	if intAbs(int64(self.X-p.X)) <= 1 && intAbs(int64(self.Y-p.Y)) <= 1 {
	// 		g.SendEffect(zone, "wood_ex", effectParams{
	// 			"x": p.X,
	// 			"y": p.Y,
	// 		})
	// 		g.SendEffect(zone, "screen_shake", effectParams{
	// 			"x": 8,
	// 			"y": 8,
	// 		})
	// 		return true
	// 	}
	// }
	// x, y, ok := zone.Move(self.X, self.Y, dirMap[rand.Intn(4)])
	// self.X = x
	// self.Y = y
	// return ok
	return true
}
