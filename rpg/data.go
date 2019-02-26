package rpg

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Position [2]int

type Definitions struct {
	RPG      RPGDef
	Zones    map[string]ZoneDef
	Tiles    map[string]TileDef
	Entities map[string]EntityDef
}

type RPGDef struct {
	StartingZone string
	Zones        map[string]ZoneInfoDef
}

type ZoneInfoDef struct {
	Enabled bool
}

type ZoneDef struct {
	SpawnPoint Position
	Width      int
	Height     int
	Map        []string
	Entity     []ZoneEntityDef
}

type TileDef struct {
	Blocking bool
}

type EntityDef struct {
	DefaultName string
	Draw        string
	Size        Position
	Blocking    bool
	Usable      bool
	UseFunc     string
	Strings     []string
}

type ZoneEntityDef struct {
	Ref      string
	Name     string
	Type     string
	Position Position
	Strings  map[string]string
}

func LoadDefinitions(dir string) (*Definitions, error) {
	def := Definitions{}

	if _, err := toml.DecodeFile(dir+"game.toml", &def.RPG); err != nil {
		log.Printf("[rpg/definitions] error loading root definitions: %v", err)
		return nil, err
	}

	if _, err := toml.DecodeFile(dir+"tiles.toml", &def.Tiles); err != nil {
		log.Printf("[rpg/definitions] error loading tile definitions: %v", err)
		return nil, err
	}

	if _, err := toml.DecodeFile(dir+"entities.toml", &def.Entities); err != nil {
		log.Printf("[rpg/definitions] error loading tile definitions: %v", err)
		return nil, err
	}

	def.Zones = make(map[string]ZoneDef)
	for z := range def.RPG.Zones {
		zone := ZoneDef{}
		if _, err := toml.DecodeFile(dir+"zones/"+z+".toml", &zone); err != nil {
			log.Printf("[rpg/definitions] error loading zone (%s) definition: %v", z, err)
			return nil, err
		}
		def.Zones[z] = zone
	}

	return &def, nil
}
