package rpg

import (
	"errors"
	"log"
)

type Item struct {
	Id           int          `json:"-"`
	Quality      int          `json:"quality"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	MaxQty       int          `json:"maxQty"`
	Durability   int          `json:"durability"`
	Price        int          `json:"price"`
	Stats        StatBlock    `json:"stats"`
	SpecialAttrs SpecialBlock `json:"specials"`
	Modded       bool         `json:"modded"`

	Held     bool   `json:"held"`
	HeldBy   int    `json:"heldBy"`
	Equipped string `json:"equipped"`

	CurrentZone int `json:"currentZone"`
	X           int `json:"x"`
	Y           int `json:"y"`
}

func (i *Item) ApplyMod(def ItemModDef) {
	i.Name = def.Name + " " + i.Name
	i.Modded = true
	i.Stats = i.Stats.Add(def.Stats)
}

func (i *Item) Give(player *Player) {
	i.Held = true
	i.HeldBy = player.Id
	i.Equipped = ""
	i.CurrentZone = -1
}

func (g *RPG) AddItem(z *Zone, itemType string, x, y int) (Item, error) {
	def, ok := g.Defs.Items[itemType]
	if !ok {
		log.Printf("[rpg/zone/%s/createitem] item doesn't exist '%s'", z.Name, itemType)
		return Item{}, errors.New("item doesn't exist")
	}

	item, ok := g.Items.New(def)
	if !ok {
		log.Printf("[rpg/zone/%s/createitem] error creating item '%s'", z.Name, itemType)
		return Item{}, nil
	}

	item.X = x
	item.Y = y
	item.CurrentZone = z.Id

	z.Items[item.Id] = true
	g.Items.Save(item)

	return item, nil
}

func (g *RPG) AddExistingItem(z *Zone, itemId int, x, y int) {
	item, ok := g.Items.Get(itemId)
	if !ok {
		log.Printf("[rpg/zone/%s/additem] item %d doesn't exist", z.Name, itemId)
		return
	}
	item.Held = false
	item.X = x
	item.Y = y
	item.CurrentZone = z.Id
	z.Items[item.Id] = true
	g.Items.Save(item)
}

func (g *RPG) RemoveItem(z *Zone, item *Item) {
	item.CurrentZone = -1
	delete(z.Items, item.Id)
}
