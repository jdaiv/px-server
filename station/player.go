package station

type Player struct {
	Name      string
	Money     int
	Slots     PlayerSlots
	Inventory []*Item
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
	Id         string
	Name       string
	MaxQty     int
	Equippable bool
	Durablity  int
}
