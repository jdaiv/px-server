package station

import "math/rand"

type entity struct {
	Parent    *Area     `json:"-"`
	Owner     string    `json:"-"`
	Type      string    `json:"type"`
	Id        string    `json:"id"`
	Transform Transform `json:"transform"`
	Velocity  Vector3   `json:"velocity"`
}

type Delta []float64

type Entity interface {
	Create(area *Area, id, source string, data Delta)
	Recv(source string, data Delta)
	Tick(dt float64)
	Send() Delta
}

type EntPlayer struct {
	entity
	User           string  `json:"user"`
	AnimationState float64 `json:"animation_state"`
}

func (e *EntPlayer) Create(area *Area, id, source string, data Delta) {
	e.Parent = area
	e.Id = id
	e.Type = "player"
	e.Owner = source

	if len(data) != 7 {
		return
	}

	e.Transform.Position.X = data[0]
	e.Transform.Position.Y = data[1]
	e.Transform.Position.Z = data[2]
	e.Velocity.X = data[3]
	e.Velocity.Y = data[4]
	e.Velocity.Z = data[5]
	e.AnimationState = data[6]
}

func (e *EntPlayer) Recv(source string, data Delta) {
	if e.Owner != source {
		return
	}

	if len(data) != 7 {
		return
	}

	e.Transform.Position.X = data[0]
	e.Transform.Position.Y = data[1]
	e.Transform.Position.Z = data[2]
	e.Velocity.X = data[3]
	e.Velocity.Y = data[4]
	e.Velocity.Z = data[5]
	e.AnimationState = data[6]
}

func (e *EntPlayer) Tick(dt float64) {
}

func (e *EntPlayer) Send() Delta {
	return Delta{
		e.Transform.Position.X,
		e.Transform.Position.Y,
		e.Transform.Position.Z,
		e.AnimationState,
	}
}

type EntFirework struct {
	entity
	Timer float64 `json:"timer"`
}

func (e *EntFirework) Create(area *Area, id, source string, data Delta) {
	e.Parent = area
	e.Id = id
	e.Type = "firework"
	e.Owner = source
	e.Timer = rand.Float64() + 0.5

	e.Transform.Position.X = rand.Float64()*256 - 128
	e.Velocity.X = rand.Float64()*2 - 1
	e.Velocity.Y = 5
}

func (e *EntFirework) Recv(source string, data Delta) {

}

func (e *EntFirework) Tick(dt float64) {
	e.Timer -= dt

	if e.Timer < -10 {
		delete(e.Parent.Entities, e.Id)
	} else if e.Timer < 0 {
		e.Velocity = Vector3{0, 0, 0}
	} else {
		e.Transform.Position = e.Transform.Position.Add(e.Velocity)
	}
}

func (e *EntFirework) Send() Delta {
	return Delta{
		e.Transform.Position.X,
		e.Transform.Position.Y,
		e.Transform.Position.Z,
	}
}

type EntFireworkLauncher struct {
	entity
	PlayerTimers map[string]float64 `json:"-"`
	Collider     AABB               `json:"-"`
}

func (e *EntFireworkLauncher) Create(area *Area, id, source string, data Delta) {
	e.Parent = area
	e.Id = id
	e.Type = "firework_launcher"
	e.Owner = source

	e.Transform.Position.X = data[0]
	e.Transform.Position.Y = data[1]
	e.Transform.Position.Z = data[2]

	e.PlayerTimers = make(map[string]float64)
	e.Collider = NewAABB(e.Transform.Position, Vector3{8, 8, 8})
}

func (e *EntFireworkLauncher) Recv(source string, data Delta) {

}

func (e *EntFireworkLauncher) Tick(dt float64) {
	for k, v := range e.PlayerTimers {
		if v > 0 {
			e.PlayerTimers[k] = v - dt
		}
	}

	for id, ent := range e.Parent.Entities {
		if player, ok := ent.(*EntPlayer); ok {
			if t, ok := e.PlayerTimers[id]; !ok || (ok && t <= 0) {
				aabb := NewAABB(player.Transform.Position, Vector3{4, 4, 4})
				if aabb.Intersects(e.Collider) {
					firework := &EntFirework{}
					e.Parent.AddEntity(firework, "", nil)
					e.PlayerTimers[id] = 1
				}
			}
		}
	}
}

func (e *EntFireworkLauncher) Send() Delta {
	return nil
}
