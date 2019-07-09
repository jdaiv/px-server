package rpg

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

func (d Item) Value() (driver.Value, error) {
	j, err := json.Marshal(d)
	return j, err
}

func (d *Item) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	source, ok := src.([]byte)
	if !ok {
		return errors.New(fmt.Sprintf("Type assertion .([]byte) failed. Actual: %T", src))
	}

	err := json.Unmarshal(source, d)
	if err != nil {
		return err
	}

	return nil
}

type ItemDB struct {
	log   *log.Logger
	DB    *sql.DB
	items map[int]Item
}

func NewItemDB(db *sql.DB) *ItemDB {
	itemDB := ItemDB{
		log:   log.New(os.Stdout, "[RPG/ItemDB] ", log.LstdFlags),
		DB:    db,
		items: make(map[int]Item),
	}

	itemDB.log.Printf("Loading items from DB...")
	rows, err := db.Query(`SELECT id, data FROM items`)

	if err != nil {
		itemDB.log.Fatalf("Couldn't load items, SQL Error: %v", err)
	}

	for rows.Next() {
		var id int
		var item Item
		if err := rows.Scan(&id, &item); err != nil {
			itemDB.log.Printf("SQL Error: %v", err)
			continue
		}
		item.Id = id
		itemDB.items[id] = item
		itemDB.log.Printf("Loaded item %d: %s", id, item.Name)
	}

	return &itemDB
}

func (db *ItemDB) New(def ItemDef) (Item, bool) {
	db.log.Printf("Creating new item %s", def.Name)
	item := Item{
		Quality: def.Quality,
		Name:    def.Name,
		Type:    def.Type,
		// PowerLevel: def.PowerLevel,
	}

	err := db.DB.QueryRow(`INSERT INTO items (data) VALUES ($1) RETURNING id`, item).Scan(&item.Id)
	if err != nil {
		log.Printf("Failed to create new item, SQL error: %v", err)
		return item, false
	}

	db.items[item.Id] = item

	return item, true
}

func (db *ItemDB) Get(id int) (Item, bool) {
	// db.log.Printf("Getting item %d", id)
	item, ok := db.items[id]
	return item, ok
}

// Gets an item by ID, while checking it's in a zone.
func (db *ItemDB) GetInZone(id int, zone int) (item Item, ok bool) {
	// db.log.Printf("Getting item %d", id)
	item, ok = db.items[id]
	ok = ok && item.CurrentZone == zone
	return
}

func (db *ItemDB) GetList(ids []int) map[int]Item {
	// db.log.Printf("Getting items %v", ids)
	items := make(map[int]Item)
	for _, id := range ids {
		if item, ok := db.items[id]; ok {
			items[id] = item
		}
	}
	return items
}

func (db *ItemDB) LoadIntoPlayer(player *Player) {
	db.log.Printf("Loading items into player %d:%s", player.Id, player.Name)

	player.Inventory = make(map[int]bool)
	player.Slots = make(map[string]int)
	for _, s := range playerSlots {
		player.Slots[s] = -1
	}

	for id, item := range db.items {
		if !item.Held || item.HeldBy != player.Id {
			continue
		}
		if item.Equipped != "" {
			if _, ok := player.Slots[item.Equipped]; ok {
				player.Slots[item.Equipped] = id
			} else {
				player.Inventory[id] = true
			}
		} else {
			player.Inventory[id] = true
		}
	}
}

func (db *ItemDB) LoadIntoZone(zone *Zone) {
	db.log.Printf("Loading items into zone %s", zone.Name)

	zone.Items = make(map[int]bool)

	for id, item := range db.items {
		if item.Held || item.CurrentZone != zone.Id {
			continue
		}
		zone.Items[id] = true
	}
}

func (db *ItemDB) Save(new Item) {
	id := new.Id
	oldItem := db.items[id]
	db.log.Printf("Saving item %v", new)
	_, err := db.DB.Exec(`UPDATE items SET data = $1 WHERE id = $2`, new, id)
	if err != nil {
		db.items[id] = oldItem
		db.log.Printf("Failed to commit item old[%v] new[%v] changes, SQL Error: %v", oldItem, new, err)
	} else {
		db.items[id] = new
	}
}
