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

func (d Player) Value() (driver.Value, error) {
	j, err := json.Marshal(d)
	return j, err
}

func (d *Player) Scan(src interface{}) error {
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

type PlayerDB struct {
	log     *log.Logger
	DB      *sql.DB
	players map[int]*Player
	dirty   map[int]bool
}

func NewPlayerDB(db *sql.DB) *PlayerDB {
	playerDB := PlayerDB{
		log:     log.New(os.Stdout, "[RPG/PlayerDB] ", log.LstdFlags),
		DB:      db,
		players: make(map[int]*Player),
		dirty:   make(map[int]bool),
	}

	playerDB.log.Printf("Loading players from DB...")
	rows, err := db.Query(`SELECT id, data FROM players`)

	if err != nil {
		playerDB.log.Fatalf("Couldn't load players, SQL error: %v", err)
	}

	for rows.Next() {
		var id int
		var p Player
		if err := rows.Scan(&id, &p); err != nil {
			if err == sql.ErrNoRows {
				// if we don't find any rows, player must be new
				p = Player{}
			} else {
				playerDB.log.Printf("SQL error: %v", err)
				continue
			}
		}
		p.Id = id
		playerDB.players[id] = &p
		playerDB.log.Printf("Loaded player %d", id)
	}

	return &playerDB
}

func (db *PlayerDB) Get(id int) *Player {
	// db.log.Printf("Getting player %d", id)
	player, ok := db.players[id]
	if !ok {
		db.log.Printf("player %d is new!", id)
		player = &Player{Id: id}
		db.players[id] = player
	}
	return player
}

func (db *PlayerDB) SetDirty(id int) {
	db.dirty[id] = true
}

func (db *PlayerDB) Commit() {
	for id := range db.dirty {
		p := db.players[id]
		// db.log.Printf("Saving player %d:%s", id, p.Name)
		_, err := db.DB.Exec(`UPDATE players SET data = $1 WHERE id = $2`, p, id)
		if err != nil {
			db.log.Printf("Failed update player[%v], SQL Error: %v", p, err)
		}
	}

	db.dirty = nil
	db.dirty = make(map[int]bool)
}
