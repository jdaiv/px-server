package rpg

type Player struct {
	Id        int
	Name      string
	Money     int
	Slots     PlayerSlots
	Inventory []*Item

	CurrentZone string
	X           int
	Y           int
}

type PlayerSlots struct {
	Head      *Item
	Body      *Item
	Legs      *Item
	RightHand *Item
	LeftHand  *Item
}

type Item struct {
	Data      ItemData
	Qty       int
	Durablity int
}

type ItemData struct {
	Id          string
	Name        string
	Type        string
	MaxQty      int
	Durablity   int
	Stats       map[string]float32
	SpecialAttr string
}
