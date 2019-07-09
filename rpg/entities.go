package rpg

import (
	"errors"
	"fmt"
)

const MISSING_ENT_STR = "!MISSING STRING!"

var entityUseFuncs = map[string]func(*RPG, *Zone, *Entity, *Player) (bool, error){
	"use_sign":     UseSign,
	"use_door":     UseDoor,
	"spawn_item":   SpawnItem,
	"spawn_npc":    SpawnNPC,
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

func (g *RPG) NewEntity(zone *Zone, id int, entType string, x, y int) (*Entity, error) {
	entityDef, ok := g.Defs.Entities[entType]
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

func (g *RPG) UseEntity(zone *Zone, e *Entity, player *Player) (bool, error) {
	if !e.RootDef.Usable {
		return false, nil
	}
	fn, ok := entityUseFuncs[e.RootDef.UseFunc]
	if !ok {
		return false, errors.New("entity use func missing")
	}
	return fn(g, zone, e, player)
}

func UseSign(g *RPG, zone *Zone, ent *Entity, player *Player) (bool, error) {
	str, ok := ent.Fields.GetString("message")
	if !ok {
		str = MISSING_ENT_STR
	}
	g.SendMessage(zone, player, fmt.Sprintf("the sign says: %s", str))
	return false, nil
}

func UseDoor(g *RPG, zone *Zone, ent *Entity, player *Player) (bool, error) {
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
	newZone, ok := g.Zones.Get(int(targetZone))
	if !ok {
		return false, errors.New("target zone doesn't exist")
	}
	g.RemovePlayer(zone, player)
	g.AddPlayer(newZone, player, int(targetX), int(targetY))
	return true, nil
}

func SpawnItem(g *RPG, zone *Zone, ent *Entity, player *Player) (bool, error) {
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

	g.AddItem(zone, itemType, int(x), int(y))

	return true, nil
}

func SpawnNPC(g *RPG, zone *Zone, ent *Entity, player *Player) (bool, error) {
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

	if _, err := g.NewNPC(zone, npcType, int(x), int(y)); err != nil {
		return false, err
	}

	return true, nil
}

func AttackDummy(g *RPG, zone *Zone, ent *Entity, player *Player) (bool, error) {
	g.SendEffect(zone, "wood_ex", effectParams{
		"x": ent.X,
		"y": ent.Y,
	})
	g.SendEffect(zone, "screen_shake", effectParams{
		"x": 16,
		"y": 16,
	})
	return true, nil
}
