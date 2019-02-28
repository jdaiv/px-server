package rpg

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type PersistantPlayerData struct {
	CurrentZone string `json:"currentZone"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
}

func (d PersistantPlayerData) Value() (driver.Value, error) {
	j, err := json.Marshal(d)
	return j, err
}

func (d *PersistantPlayerData) Scan(src interface{}) error {
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

func LoadPlayer(db *sql.DB, id int) (PersistantPlayerData, error) {
	player := PersistantPlayerData{}

	err := db.QueryRow(`SELECT data FROM players
        WHERE id = $1`, id).Scan(&player)
	if err != nil {
		log.Printf("SQL Error: %v", err)
		if err == sql.ErrNoRows {
			return player, errors.New("player not found")
		}
		return player, err
	}

	return player, nil
}

func SavePlayer(db *sql.DB, player *Player) error {
	data := PersistantPlayerData{
		CurrentZone: player.CurrentZone,
		X:           player.X,
		Y:           player.Y,
	}
	_, err := db.Exec(`UPDATE players SET data = $1 WHERE id = $2`, data, player.Id)
	if err != nil {
		log.Printf("SQL Error: %v", err)
		return err
	}

	return nil
}
