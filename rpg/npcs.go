package rpg

import (
	"errors"
)

var npcLogicFuncs = map[string]func(*Entity, *Player) (bool, error){
	"use_sign":     UseSign,
	"use_door":     UseDoor,
	"spawn_item":   SpawnItem,
	"take_item":    TakeItem,
	"attack_dummy": AttackDummy,
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
