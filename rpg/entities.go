package rpg

type EntityInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Usable    bool   `json:"usable"`
	Collision bool   `json:"-"`
}

type Entity interface {
	GetInfo() EntityInfo
	Init(*Zone, int, string, int, int)
	Use(Player) bool
}

type ent struct {
	Zone *Zone
	Id   int
	Name string
	Type string
	X    int
	Y    int
}

func (e *ent) Init(zone *Zone, id int, name string, x, y int) {
	e.Id = id
	e.Name = name
	e.X = x
	e.Y = y
}

type EntSign struct {
	ent
	Text string
}

func NewSign(text string) *EntSign {
	return &EntSign{ent: ent{Type: "sign"}, Text: text}
}

func (s *EntSign) GetInfo() EntityInfo {
	return EntityInfo{
		Name:      "SIGN",
		Type:      "sign",
		X:         5,
		Y:         5,
		Usable:    true,
		Collision: true,
	}
}

func (s *EntSign) Use(player Player) bool {
	return false
}
