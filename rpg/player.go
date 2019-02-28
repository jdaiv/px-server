package rpg

type Player struct {
	Id        int         `json:"id"`
	Name      string      `json:"name"`
	Slots     PlayerSlots `json:"slots"`
	Inventory []*Item     `json:"inventory"`

	CurrentZone string `json:"-"`
	X           int    `json:"-"`
	Y           int    `json:"-"`

	DisplayData PlayerDisplayData `json:"-"`
}

type PlayerSlots struct {
	Head      *Item `json:"head"`
	Body      *Item `json:"body"`
	Legs      *Item `json:"legs"`
	RightHand *Item `json:"rightHand"`
	LeftHand  *Item `json:"leftHand"`
}

type Item struct {
	Id         int      `json:"-"`
	Data       ItemData `json:"data"`
	Qty        int      `json:"qty"`
	Durability int      `json:"durability"`
}

type ItemData struct {
	Id          string             `json:"id"`
	Name        string             `json:"name"`
	Type        string             `json:"type"`
	MaxQty      int                `json:"maxQty"`
	Durability  int                `json:"durability"`
	Stats       map[string]float32 `json:"stats"`
	SpecialAttr string             `json:"special"`
}

type PlayerDisplayData struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

func (p *Player) UpdateDisplay() {
	p.DisplayData = PlayerDisplayData{
		Id:   p.Id,
		Name: p.Name,
		X:    p.X,
		Y:    p.Y,
	}
}
