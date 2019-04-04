package rpg

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
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

func (g *RPG) NewNPC(z *Zone, npcType string, x, y int) (*NPC, error) {
	if z.Map.IsBlocking(x, y) {
		return nil, fmt.Errorf("coords %d,%d blocked", x, y)
	}

	id := z.NPCCount
	npcDef, ok := g.Defs.NPCs[npcType]
	if !ok {
		return nil, errors.New("npc missing")
	}

	items := make(map[string]NPCItem)
	for slot, itemName := range npcDef.Slots {
		if itemName == "" {
			continue
		}
		if itemDef, ok := g.Defs.Items[itemName]; ok {
			items[slot] = NPCItem{
				itemDef.Name,
				itemDef.Stats,
				itemDef.Special,
			}
		}
	}

	stats := npcDef.Skills.BuildStats()

	npc := &NPC{
		Id:        id,
		Name:      npcDef.DefaultName,
		Type:      npcType,
		X:         x,
		Y:         y,
		HP:        stats.MaxHP,
		MaxHP:     stats.MaxHP,
		Alignment: npcDef.Alignment,
		Logic:     npcDef.Logic,
		Skills:    npcDef.Skills,
		Stats:     stats,
		Slots:     items,
	}

	z.NPCs[id] = npc
	z.NPCCount += 1

	g.BuildCollisionMap(z)

	return npc, nil
}

func (g *RPG) RemoveNPC(z *Zone, npcId int) {
	delete(z.NPCs, npcId)
	g.BuildCollisionMap(z)
	g.CheckCombat(z)
}

func (n *NPC) GetName() string {
	return n.Name
}

func (n *NPC) InitCombat() CombatInfo {
	return CombatInfo{
		Initiative: rand.Intn(20),
		IsPlayer:   false,
		Id:         n.Id,
	}
}

func (n *NPC) Attack() DamageInfo {
	return n.Stats.RollPhysDamage()
}

func (n *NPC) Damage(dmg DamageInfo) DamageInfo {
	dmg = n.Stats.RollDefence(dmg)
	n.HP -= dmg.Amount
	return dmg
}

func (n *NPC) NewTurn(ci *CombatInfo) {

}

func (n *NPC) Tick(g *RPG, z *Zone, ci *CombatInfo) {
	fn, ok := npcCombatLogicFuncs[n.Logic]
	if !ok {
		log.Printf("entity '%s' combat logic missing", n.Type)
		return
	}
	fn(g, n, z)
}

func (n *NPC) IsTurnOver(ci *CombatInfo) bool {
	return true
}

func BlobIdle(g *RPG, self *NPC, zone *Zone) bool {
	return false
}

func BlobCombat(g *RPG, self *NPC, zone *Zone) bool {
	for _, p := range zone.Players {
		if intAbs(int64(self.X-p.X)) <= 1 && intAbs(int64(self.Y-p.Y)) <= 1 {
			g.DoMeleeAttack(zone, self, p)
			g.SendEffect(zone, "wood_ex", effectParams{
				"x": p.X,
				"y": p.Y,
			})
			g.SendEffect(zone, "screen_shake", effectParams{
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
