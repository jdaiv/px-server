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
	NPCs     map[string]NPCDef
	Items    map[string]ItemDef
	ItemMods map[string]ItemModDef
	Skills   map[string]SkillDef
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
	NPC        []ZoneNPCDef
}

type TileDef struct {
	Blocking bool
}

type EntityDef struct {
	DefaultName   string
	Draw          string
	Size          Position
	Blocking      bool
	Usable        bool
	UseFunc       string
	UseText       string
	Strings       []string
	ExportStrings []string
	Ints          []string
}

type NPCDef struct {
	Alignment string
	Logic     string
}

type ZoneEntityDef struct {
	Ref      string
	Name     string
	Type     string
	Position Position
	Strings  map[string]string
	Ints     map[string]int
}

type ZoneNPCDef struct {
	Ref      string
	Name     string
	Type     string
	Position Position
}

type ItemDef struct {
	Name       string
	Type       string
	Quality    int
	MaxQty     int
	Durability int
	Special    []string
	Stats      map[string]int
}

type ItemModDef struct {
	Name  string
	Stats map[string]int
}

type SkillDef struct {
	Name string
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

	if _, err := toml.DecodeFile(dir+"npcs.toml", &def.NPCs); err != nil {
		log.Printf("[rpg/definitions] error loading npc definitions: %v", err)
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

	if _, err := toml.DecodeFile(dir+"items.toml", &def.Items); err != nil {
		log.Printf("[rpg/definitions] error loading item definitions: %v", err)
		return nil, err
	}

	if _, err := toml.DecodeFile(dir+"item_mods.toml", &def.ItemMods); err != nil {
		log.Printf("[rpg/definitions] error loading item mod definitions: %v", err)
		return nil, err
	}

	if _, err := toml.DecodeFile(dir+"skills.toml", &def.Skills); err != nil {
		log.Printf("[rpg/definitions] error loading skill definitions: %v", err)
		return nil, err
	}

	return &def, nil
}
