package rpg

import "database/sql"

type RPG struct {
	Zones   map[string]*Zone
	Players map[int]*Player

	Incoming chan string
	DB       *sql.DB
}

func NewRPG(db *sql.DB) *RPG {
	return &RPG{
		Zones:    make(map[string]*Zone),
		Players:  make(map[int]*Player),
		Incoming: make(chan string),
		DB:       db,
	}
}

func (g *RPG) Tick(dt float64) {

}
