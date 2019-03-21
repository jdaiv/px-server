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

func (d Zone) Value() (driver.Value, error) {
	j, err := json.Marshal(d)
	return j, err
}

func (d *Zone) Scan(src interface{}) error {
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

type ZoneDB struct {
	log      *log.Logger
	DB       *sql.DB
	AllZones map[int]*Zone
	dirty    map[int]bool
}

func NewZoneDB(db *sql.DB) *ZoneDB {
	zoneDB := ZoneDB{
		log:      log.New(os.Stdout, "[RPG/ZoneDB] ", log.LstdFlags),
		DB:       db,
		AllZones: make(map[int]*Zone),
		dirty:    make(map[int]bool),
	}

	zoneDB.log.Printf("Loading zones from DB...")
	rows, err := db.Query(`SELECT id, data FROM zones`)

	if err != nil {
		zoneDB.log.Fatalf("Couldn't load zones, SQL error: %v", err)
	}

	for rows.Next() {
		var id int
		var z Zone
		if err := rows.Scan(&id, &z); err != nil {
			zoneDB.log.Printf("SQL error: %v", err)
			continue
		}
		z.Id = id
		zoneDB.AllZones[id] = &z
		zoneDB.log.Printf("Loaded zone %d", id)
	}

	return &zoneDB
}

func (db *ZoneDB) Insert(zone *Zone) bool {
	db.log.Printf("Creating new zone")

	err := db.DB.QueryRow(`INSERT INTO zones (data) VALUES ($1) RETURNING id`, zone).Scan(&zone.Id)
	if err != nil {
		log.Printf("Failed to create new zone, SQL error: %v", err)
		return false
	}

	db.AllZones[zone.Id] = zone

	return true
}

func (db *ZoneDB) Get(id int) (*Zone, bool) {
	zone, ok := db.AllZones[id]
	return zone, ok
}

func (db *ZoneDB) SetDirty(id int) {
	db.dirty[id] = true
}

func (db *ZoneDB) IsDirty(id int) bool {
	dirty, ok := db.dirty[id]
	return ok && dirty
}

func (db *ZoneDB) Commit() {
	for id := range db.dirty {
		z := db.AllZones[id]
		_, err := db.DB.Exec(`UPDATE zones SET data = $1 WHERE id = $2`, z, id)
		if err != nil {
			db.log.Printf("Failed update zone[%v], SQL Error: %v", z, err)
		}
	}

	db.dirty = nil
	db.dirty = make(map[int]bool)
}
