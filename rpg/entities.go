package rpg

import (
	"errors"
	"fmt"
)

const MISSING_ENT_STR = "!MISSING STRING!"

var entityUseFuncs = map[string]func(*Zone, *Entity, *Player) (bool, error){
	"use_sign":     UseSign,
	"use_door":     UseDoor,
	"spawn_item":   SpawnItem,
	"modify_item":  ModifyItem,
	"spawn_npc":    SpawnNPC,
	"take_item":    TakeItem,
	"attack_dummy": AttackDummy,
}

type Entity struct {
	RootDef  EntityDef    `json:"-"`
	Id       int          `json:"id"`
	Name     string       `json:"name"`
	Type     string       `json:"type"`
	X        int          `json:"x"`
	Y        int          `json:"y"`
	Rotation int          `json:"rotation"`
	Fields   EntityFields `json:"fields"`
}

type EntityFields map[string]interface{}

func (ef EntityFields) GetNumber(name string) (float64, bool) {
	f, ok := ef[name]
	if !ok {
		return 0, false
	}

	v, ok := f.(float64)
	return v, ok
}

func (ef EntityFields) GetString(name string) (string, bool) {
	f, ok := ef[name]
	if !ok {
		return "", false
	}

	v, ok := f.(string)
	return v, ok
}

func NewEntity(zone *Zone, id int, entType string, x, y int) (*Entity, error) {
	entityDef, ok := zone.Parent.Defs.Entities[entType]
	if !ok {
		return nil, errors.New("entity missing")
	}
	fields := make(EntityFields)
	for _, f := range entityDef.Fields {
		fields[f.Name] = f.Default
	}
	return &Entity{
		RootDef: entityDef,
		Id:      id,
		Name:    entityDef.DefaultName,
		Type:    entType,
		X:       x,
		Y:       y,
		Fields:  fields,
	}, nil
}

func (e *Entity) Use(zone *Zone, player *Player) (bool, error) {
	if !e.RootDef.Usable {
		return false, nil
	}
	fn, ok := entityUseFuncs[e.RootDef.UseFunc]
	if !ok {
		return false, errors.New("entity use func missing")
	}
	return fn(zone, e, player)
}

func UseSign(zone *Zone, ent *Entity, player *Player) (bool, error) {
	str, ok := ent.Fields.GetString("message")
	if !ok {
		str = MISSING_ENT_STR
	}
	zone.SendMessage(player, fmt.Sprintf("the sign says: %s", str))
	return false, nil
}

func UseDoor(zone *Zone, ent *Entity, player *Player) (bool, error) {
	targetZone, ok := ent.Fields.GetNumber("target_zone")
	if !ok {
		return false, errors.New("target zone not found")
	}
	targetX, ok := ent.Fields.GetNumber("x")
	if !ok {
		targetX = -1
	}
	targetY, ok := ent.Fields.GetNumber("y")
	if !ok {
		targetY = -1
	}
	root := zone.Parent
	newZone, ok := root.Zones.Get(int(targetZone))
	if !ok {
		return false, errors.New("target zone doesn't exist")
	}
	zone.RemovePlayer(player)
	newZone.AddPlayer(player, int(targetX), int(targetY))
	return true, nil
}

func SpawnItem(zone *Zone, ent *Entity, player *Player) (bool, error) {
	itemType, ok := ent.Fields.GetString("item")
	if !ok {
		return false, errors.New("target item not found")
	}
	x, ok := ent.Fields.GetNumber("x")
	if !ok {
		return false, errors.New("x not found")
	}
	y, ok := ent.Fields.GetNumber("y")
	if !ok {
		return false, errors.New("y not found")
	}

	zone.AddItem(itemType, int(x), int(y))

	return true, nil
}

func ModifyItem(zone *Zone, ent *Entity, player *Player) (bool, error) {
	modId, ok := ent.Fields.GetString("item_mod")
	if !ok {
		return false, errors.New("target item mod not found")
	}
	slot, ok := ent.Fields.GetString("target_slot")
	if !ok {
		return false, errors.New("target slot not found")
	}
	modDef, ok := zone.Parent.Defs.ItemMods[modId]
	if !ok {
		return false, errors.New("item mod not found")
	}

	if itemId, ok := player.Slots[slot]; ok && itemId > 0 {
		item, exists := zone.Parent.Items.Get(itemId)
		if !exists {
			return false, errors.New("item not found")
		}
		if item.Modded {
			return false, nil
		}
		item.ApplyMod(modDef)
		zone.Parent.Items.Save(item)
		return true, nil
	}

	return false, nil
}

func SpawnNPC(zone *Zone, ent *Entity, player *Player) (bool, error) {
	npcType, ok := ent.Fields.GetString("npc")
	if !ok {
		return false, errors.New("target npc not found")
	}
	x, ok := ent.Fields.GetNumber("x")
	if !ok {
		return false, errors.New("x not found")
	}
	y, ok := ent.Fields.GetNumber("y")
	if !ok {
		return false, errors.New("y not found")
	}

	zone.AddNPC(npcType, int(x), int(y), true)

	return true, nil
}

func TakeItem(zone *Zone, ent *Entity, player *Player) (bool, error) {
	zone.RemoveEntity(ent.Id)
	return true, nil
}

func AttackDummy(zone *Zone, ent *Entity, player *Player) (bool, error) {
	zone.SendEffect("wood_ex", effectParams{
		"x": ent.X,
		"y": ent.Y,
	})
	zone.SendEffect("screen_shake", effectParams{
		"x": 16,
		"y": 16,
	})
	return true, nil
}
