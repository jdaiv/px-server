package rpg

import (
	"database/sql"
	"errors"
	"log"
)

type Item struct {
	Parent *RPG `json:"-"`

	Id           int            `json:"id"`
	Quality      int            `json:"quality"`
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	MaxQty       int            `json:"maxQty"`
	Durability   int            `json:"durability"`
	Stats        map[string]int `json:"stats"`
	SpecialAttrs []string       `json:"specials"`

	Held     bool   `json:"held"`
	HeldBy   int    `json:"heldBy"`
	Equipped string `json:"equipped"`

	CurrentZone string `json:"currentZone"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
}

type ItemInfo struct {
	Id           int            `json:"id"`
	Quality      int            `json:"quality"`
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	Durability   int            `json:"durability"`
	Stats        map[string]int `json:"stats"`
	SpecialAttrs []string       `json:"specials"`

	X int `json:"x"`
	Y int `json:"y"`
}

func (g *RPG) NewItem(def ItemDef) (*Item, error) {
	item := &Item{
		Parent:       g,
		Quality:      def.Quality,
		Name:         def.Name,
		Type:         def.Type,
		MaxQty:       def.MaxQty,
		Durability:   def.Durability,
		SpecialAttrs: make([]string, len(def.Special)),
		Stats:        make(map[string]int),
	}
	copy(item.SpecialAttrs, def.Special)
	item.ApplyStats(def.Stats)

	var id int

	err := g.DB.QueryRow(`INSERT INTO items (data) VALUES ($1) RETURNING id`, item).Scan(&id)
	if err != nil {
		log.Printf("SQL Error: %v", err)
		return nil, err
	}

	item.Id = id
	g.Items[id] = item

	return item, nil
}

func (g *RPG) GetItem(id int) (*Item, error) {
	if item, hasItem := g.Items[id]; hasItem {
		return item, nil
	}

	item := &Item{}

	err := g.DB.QueryRow(`SELECT data FROM items WHERE id = $1`, id).Scan(item)
	if err != nil {
		log.Printf("SQL Error: %v", err)
		if err == sql.ErrNoRows {
			return item, errors.New("item not found")
		}
		return item, err
	}

	item.Parent = g
	g.Items[id] = item
	return item, nil
}

func (g *RPG) GetItemsForZone(name string) map[int]*Item {
	log.Printf("[rpg/zone/%s/loaditems] loading items from DB", name)

	items := make(map[int]*Item)

	rows, err := g.DB.Query(`SELECT id, data FROM items WHERE data->>'currentZone' = $1`, name)

	if err != nil {
		log.Printf("SQL Error: %v", err)
		return items
	}

	for rows.Next() {
		var id int
		var item Item
		if err := rows.Scan(&id, &item); err != nil {
			log.Printf("SQL Error: %v", err)
			continue
		}
		item.Parent = g
		items[id] = &item
		log.Printf("[rpg/zone/%s/loaditems] loaded item %d: %s", name, id, item.Name)
	}

	return items
}

func (g *RPG) LoadItemsForPlayer(player *Player) {
	log.Printf("[rpg/player/%d/loaditems] loading items from DB", player.Id)

	player.Inventory = make(map[int]*Item)
	player.Slots = make(map[string]*Item)
	for _, s := range playerSlots {
		player.Slots[s] = nil
	}

	rows, err := g.DB.Query(`SELECT id, data FROM items WHERE data->>'held' = 'true' AND data->>'heldBy' = $1`, player.Id)

	if err != nil {
		log.Printf("SQL Error: %v", err)
		return
	}

	for rows.Next() {
		var itemId int
		var item Item
		if err := rows.Scan(&itemId, &item); err != nil {
			log.Printf("SQL Error: %v", err)
			continue
		}
		item.Parent = g
		if item.Equipped != "" {
			if _, ok := player.Slots[item.Equipped]; ok {
				player.Slots[item.Equipped] = &item
			} else {
				player.Inventory[itemId] = &item
			}
		} else {
			player.Inventory[itemId] = &item
		}
		log.Printf("[rpg/player/%d/loaditems] loaded item %d: %s", player.Id, itemId, item.Name)
	}
}

func (i *Item) Save() error {
	_, err := i.Parent.DB.Exec(`UPDATE items SET data = $1 WHERE id = $2`, i, i.Id)
	if err != nil {
		log.Printf("SQL Error: %v", err)
		return err
	}

	return nil
}

func (i *Item) ApplyStats(stats map[string]int) error {
	for new, val := range stats {
		if _, hasNew := i.Stats[new]; hasNew {
			i.Stats[new] += val
		} else {
			i.Stats[new] = val
		}
	}
	return i.Save()
}

func (i *Item) ApplyMod(def ItemModDef) error {
	i.Name = def.Name + " " + i.Name
	i.ApplyStats(def.Stats)
	return i.Save()
}

func (i *Item) Give(player *Player) error {
	i.Held = true
	i.HeldBy = player.Id
	i.Equipped = ""
	i.CurrentZone = ""
	player.Inventory[i.Id] = i
	return i.Save()
}

func (i *Item) GetInfo() ItemInfo {
	return ItemInfo{
		Id:           i.Id,
		Name:         i.Name,
		Quality:      i.Quality,
		Type:         i.Type,
		Durability:   i.Durability,
		SpecialAttrs: i.SpecialAttrs,
		Stats:        i.Stats,
		X:            i.X,
		Y:            i.Y,
	}
}
