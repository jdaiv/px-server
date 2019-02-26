package rpg

import (
	"errors"
	"fmt"
)

const MISSING_ENT_STR = "!MISSING STRING!"

var entityUseFuncs = map[string]func(*Entity, *Player) bool{
	"use_sign": UseSign,
	// "use_door": UseDoor,
}

type EntityInfo struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Usable   bool   `json:"usable"`
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
	return &Entity{
		Def:     def,
		RootDef: entityDef,
		Zone:    zone,
		Id:      id,
		Name:    def.Name,
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
	return fn(e, player), nil
}

func UseSign(ent *Entity, player *Player) bool {
	str, ok := ent.Def.Strings["message"]
	if !ok {
		str = MISSING_ENT_STR
	}
	ent.Zone.SendMessage(player, fmt.Sprintf("the sign says: %s", str))
	return false
}
