package rpg

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Position [2]int

type Definitions struct {
	Tiles    []TileDef
	Entities map[string]EntityDef
	NPCs     map[string]NPCDef
	Items    map[string]ItemDef
	ItemMods map[string]ItemModDef
	Skills   map[string]SkillDef
	Spells   map[string]SpellDef
}

type TileDef struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Blocking bool   `json:"blocking"`
}

type EntityDef struct {
	DefaultName string
	Draw        string
	Size        Position
	ServerOnly  bool
	Blocking    bool
	Usable      bool
	UseFunc     string
	UseText     string
	Fields      []EntityField
}

type EntityField struct {
	Name    string
	Type    string
	Export  bool
	Default interface{}
}

type NPCDef struct {
	DefaultName string
	Alignment   string
	Logic       string
	HP          int
	Slots       map[string]string
	Skills      SkillBlock
}

type ItemDef struct {
	Name       string
	Type       string
	Quality    int
	MaxQty     int
	Durability int
	Price      int
	Special    SpecialBlock
	Stats      StatBlock
}

type ItemModDef struct {
	Name  string
	Stats StatBlock
}

type SkillDef struct {
	Name string
}

type SpellDef struct {
	Name    string
	Skill   string
	Level   int
	Cost    int
	Effects []SpellEffectDef
}

type SpellEffectDef struct {
	Type     string
	Effect   string
	Duration int
	Origin   string
	Target   string
	Range    int
	Damage   int
}

func LoadDefinitions(dir string) (*Definitions, error) {
	def := Definitions{}

	tiles := struct{ Tile []TileDef }{}
	if _, err := toml.DecodeFile(dir+"tiles.toml", &tiles); err != nil {
		log.Printf("[rpg/definitions] error loading tile definitions: %v", err)
		return nil, err
	}
	for i := range tiles.Tile {
		tiles.Tile[i].Id = i
	}
	log.Printf("%v", tiles.Tile)
	def.Tiles = tiles.Tile

	if _, err := toml.DecodeFile(dir+"entities.toml", &def.Entities); err != nil {
		log.Printf("[rpg/definitions] error loading tile definitions: %v", err)
		return nil, err
	}

	if _, err := toml.DecodeFile(dir+"npcs.toml", &def.NPCs); err != nil {
		log.Printf("[rpg/definitions] error loading npc definitions: %v", err)
		return nil, err
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

	if _, err := toml.DecodeFile(dir+"spells.toml", &def.Spells); err != nil {
		log.Printf("[rpg/definitions] error loading spell definitions: %v", err)
		return nil, err
	}

	return &def, nil
}
