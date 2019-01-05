package station

type Trigger struct {
	Parent   *Area
	Active   bool
	Position Vector3
	Action   func()
}
