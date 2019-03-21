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

var npcLogicFuncs = map[string]func(*NPC, *Zone) bool{
	"blob": BlobIdle,
}

var npcCombatLogicFuncs = map[string]func(*NPC, *Zone) bool{
	"blob": BlobCombat,
}

type NPC struct {
	Id        int
	Name      string
	Type      string
	X         int
	Y         int
	HP        int
	MaxHP     int
	Alignment string
	Logic     string
	Skills    SkillBlock
	Stats     StatBlock
	Slots     map[string]NPCItem
}

type NPCItem struct {
	Name    string
	Stats   StatBlock
	Special SpecialBlock
}

func NewNPC(zone *Zone, id int, npcType string, x, y int) (*NPC, error) {
	npcDef, ok := zone.Parent.Defs.NPCs[npcType]
	if !ok {
		return nil, errors.New("npc missing")
	}

	items := make(map[string]NPCItem)
	for slot, itemName := range npcDef.Slots {
		if itemName == "" {
			continue
		}
		if itemDef, ok := zone.Parent.Defs.Items[itemName]; ok {
			items[slot] = NPCItem{
				itemDef.Name,
				itemDef.Stats,
				itemDef.Special,
			}
		}
	}

	stats := npcDef.Skills.BuildStats()

	return &NPC{
		Id:        id,
		Name:      npcDef.DefaultName,
		Type:      npcType,
		X:         x,
		Y:         y,
		HP:        npcDef.HP,
		MaxHP:     npcDef.HP,
		Alignment: npcDef.Alignment,
		Logic:     npcDef.Logic,
		Skills:    npcDef.Skills,
		Stats:     stats,
		Slots:     items,
	}, nil
}

func (n *NPC) CombatTick(zone *Zone) {
	fn, ok := npcCombatLogicFuncs[n.Logic]
	if !ok {
		log.Printf("entity '%s' combat logic missing", n.Type)
		return
	}
	fn(n, zone)
}

func BlobIdle(self *NPC, zone *Zone) bool {
	return false
}

func BlobCombat(self *NPC, zone *Zone) bool {
	for _, p := range zone.Players {
		if intAbs(int64(self.X-p.X)) <= 1 && intAbs(int64(self.Y-p.Y)) <= 1 {
			zone.DoMeleeAttack(self, p)
			zone.SendEffect("wood_ex", effectParams{
				"x": p.X,
				"y": p.Y,
			})
			zone.SendEffect("screen_shake", effectParams{
				"x": 8,
				"y": 8,
			})
			return true
		}
	}
	x, y, ok := zone.Move(self.X, self.Y, dirMap[rand.Intn(4)])
	self.X = x
	self.Y = y
	return ok
}
