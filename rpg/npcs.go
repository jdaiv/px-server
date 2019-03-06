package rpg

import (
	"errors"
	"log"
	"math/rand"
)

var dirMap = map[int]string{
	0: "N",
	1: "S",
	2: "E",
	3: "W",
}

var npcLogicFuncs = map[string]func(*NPC) bool{
	"blob": BlobIdle,
}

var npcCombatLogicFuncs = map[string]func(*NPC) bool{
	"blob": BlobCombat,
}

type NPCInfo struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Alignment string `json:"alignment"`
}

type NPC struct {
	Zone      *Zone
	Id        int
	Name      string
	Type      string
	X         int
	Y         int
	HP        int
	Alignment string
	Logic     string
}

func NewNPC(zone *Zone, id int, def ZoneNPCDef) (*NPC, error) {
	npcDef, ok := zone.Parent.Defs.NPCs[def.Type]
	if !ok {
		return nil, errors.New("npc missing")
	}
	name := def.Name
	// if len(name) <= 0 {
	// 	name = npcDef.DefaultName
	// }
	return &NPC{
		Zone:      zone,
		Id:        id,
		Name:      name,
		Type:      def.Type,
		X:         def.Position[0],
		Y:         def.Position[1],
		HP:        10,
		Alignment: npcDef.Alignment,
		Logic:     npcDef.Logic,
	}, nil
}

func (n *NPC) GetInfo() NPCInfo {
	return NPCInfo{
		Id:        n.Id,
		Name:      n.Name,
		Type:      n.Type,
		X:         n.X,
		Y:         n.Y,
		Alignment: n.Alignment,
	}
}

func (n *NPC) CombatTick() {
	fn, ok := npcCombatLogicFuncs[n.Logic]
	if !ok {
		log.Printf("entity '%s' combat logic missing", n.Type)
		return
	}
	fn(n)
}

func BlobIdle(self *NPC) bool {
	return false
}

func BlobCombat(self *NPC) bool {
	for _, p := range self.Zone.Players {
		if intAbs(int64(self.X-p.X)) <= 1 && intAbs(int64(self.Y-p.Y)) <= 1 {
			self.Attack(p)
			self.Zone.SendEffect("wood_ex", p.X, p.Y)
			self.Zone.SendEffect("screen_shake", 8, 8)
			return true
		}
	}
	x, y, ok := self.Zone.Move(self.X, self.Y, dirMap[rand.Intn(4)])
	self.X = x
	self.Y = y
	return ok
}
