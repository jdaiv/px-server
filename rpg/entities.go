package rpg

import (
	"errors"
	"fmt"
)

const MISSING_ENT_STR = "!MISSING STRING!"

var entityUseFuncs = map[string]func(*Entity, *Player) (bool, error){
	"use_sign":     UseSign,
	"use_door":     UseDoor,
	"spawn_item":   SpawnItem,
	"spawn_npc":    SpawnNPC,
	"take_item":    TakeItem,
	"attack_dummy": AttackDummy,
}

type EntityInfo struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Usable   bool   `json:"usable"`
	UseText  string `json:"useText"`
	Blocking bool   `json:"-"`
}

type Entity struct {
	Def     ZoneEntityDef
	RootDef EntityDef
	Zone    *Zone
	Id      int
	Name    string
	Type    string
	X       int
	Y       int
}

func NewEntity(zone *Zone, id int, def ZoneEntityDef) (*Entity, error) {
	entityDef, ok := zone.Parent.Defs.Entities[def.Type]
	if !ok {
		return nil, errors.New("entity missing")
	}
	name := def.Name
	if len(name) <= 0 {
		name = entityDef.DefaultName
	}
	return &Entity{
		Def:     def,
		RootDef: entityDef,
		Zone:    zone,
		Id:      id,
		Name:    name,
		Type:    def.Type,
		X:       def.Position[0],
		Y:       def.Position[1],
	}, nil
}

func (e *Entity) GetInfo() EntityInfo {
	return EntityInfo{
		Id:       e.Id,
		Name:     e.Name,
		Type:     e.Type,
		X:        e.X,
		Y:        e.Y,
		Usable:   e.RootDef.Usable,
		UseText:  e.RootDef.UseText,
		Blocking: e.RootDef.Blocking,
	}
}

func (e *Entity) Use(player *Player) (bool, error) {
	if !e.RootDef.Usable {
		return false, nil
	}
	fn, ok := entityUseFuncs[e.RootDef.UseFunc]
	if !ok {
		return false, errors.New("entity use func missing")
	}
	return fn(e, player)
}

func UseSign(ent *Entity, player *Player) (bool, error) {
	str, ok := ent.Def.Strings["message"]
	if !ok {
		str = MISSING_ENT_STR
	}
	ent.Zone.SendMessage(player, fmt.Sprintf("the sign says: %s", str))
	return false, nil
}

func UseDoor(ent *Entity, player *Player) (bool, error) {
	targetZone, ok := ent.Def.Strings["target_zone"]
	if !ok {
		return false, errors.New("target zone not found")
	}
	targetX, ok := ent.Def.Ints["x"]
	if !ok {
		targetX = -1
	}
	targetY, ok := ent.Def.Ints["y"]
	if !ok {
		targetY = -1
	}
	root := ent.Zone.Parent
	newZone, ok := root.Zones[targetZone]
	if !ok {
		return false, errors.New("target zone doesn't exist")
	}
	ent.Zone.RemovePlayer(player)
	newZone.AddPlayer(player, targetX, targetY)
	return true, nil
}

func SpawnItem(ent *Entity, player *Player) (bool, error) {
	itemType, ok := ent.Def.Strings["item_id"]
	if !ok {
		return false, errors.New("target item not found")
	}
	x, ok := ent.Def.Ints["x"]
	if !ok {
		return false, errors.New("x not found")
	}
	y, ok := ent.Def.Ints["y"]
	if !ok {
		return false, errors.New("y not found")
	}

	ent.Zone.AddItem(itemType, x, y)

	return true, nil
}

func SpawnNPC(ent *Entity, player *Player) (bool, error) {
	npcType, ok := ent.Def.Strings["npc_id"]
	if !ok {
		return false, errors.New("target npc not found")
	}
	x, ok := ent.Def.Ints["x"]
	if !ok {
		return false, errors.New("x not found")
	}
	y, ok := ent.Def.Ints["y"]
	if !ok {
		return false, errors.New("y not found")
	}

	ent.Zone.AddNPC(ZoneNPCDef{
		Name:     "SPAWNED_NPC",
		Position: Position{x, y},
		Type:     npcType,
	}, true)

	return true, nil
}

func TakeItem(ent *Entity, player *Player) (bool, error) {
	ent.Zone.RemoveEntity(ent.Id)
	return true, nil
}

func AttackDummy(ent *Entity, player *Player) (bool, error) {
	ent.Zone.SendEffect("wood_ex", ent.X, ent.Y)
	ent.Zone.SendEffect("screen_shake", 16, 16)
	return true, nil
}
