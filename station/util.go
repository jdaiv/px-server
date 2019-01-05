package station

type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func (a Vector3) Clone() Vector3 {
	return Vector3{a.X, a.Y, a.Z}
}

func (a Vector3) Add(b Vector3) Vector3 {
	a.X += b.X
	a.Y += b.Y
	a.Z += b.Z
	return a
}

func (a Vector3) Sub(b Vector3) Vector3 {
	a.X -= b.X
	a.Y -= b.Y
	a.Z -= b.Z
	return a
}

func (a Vector3) Mul(b Vector3) Vector3 {
	a.X *= b.X
	a.Y *= b.Y
	a.Z *= b.Z
	return a
}

func (a Vector3) Div(b Vector3) Vector3 {
	a.X /= b.X
	a.Y /= b.Y
	a.Z /= b.Z
	return a
}

type AABB struct {
	Center  Vector3
	Extents Vector3
	Min     Vector3
	Max     Vector3
}

func NewAABB(center, extents Vector3) AABB {
	return AABB{
		center, extents,
		center.Clone().Sub(extents),
		center.Clone().Add(extents),
	}
}

func (a AABB) Intersects(b AABB) bool {
	return (a.Min.X <= b.Max.X && a.Max.X >= b.Min.X) &&
		(a.Min.Y <= b.Max.Y && a.Max.Y >= b.Min.Y) &&
		(a.Min.Z <= b.Max.Z && a.Max.Z >= b.Min.Z)
}

func (a AABB) Intersects2D(b AABB) bool {
	return (a.Min.X <= b.Max.X && a.Max.X >= b.Min.X) &&
		(a.Min.Y <= b.Max.Y && a.Max.Y >= b.Min.Y)
}
