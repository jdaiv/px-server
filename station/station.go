package station

import "github.com/google/uuid"

type Area struct {
	Entities   map[string]Entity `json:"entities"`
	LastUpdate map[string]Delta  `json:"-"`
}

func NewArea() *Area {
	a := &Area{
		make(map[string]Entity),
		make(map[string]Delta),
	}

	launcher := &EntFireworkLauncher{}
	a.AddEntity(launcher, "", []float64{-64, 0, 4})

	return a
}

func (a *Area) Handle(source, action, desc string, data Delta) {
	switch action {
	case "create":
		var e Entity
		switch desc {
		case "player":
			e = &EntPlayer{User: source}
			break
		}
		if e != nil {
			a.AddEntity(e, source, data)
		}
		break
	case "remove":
		switch desc {
		case "player":
			for id, e := range a.Entities {
				if player, ok := e.(*EntPlayer); ok {
					if player.Owner == source {
						delete(a.Entities, id)
					}
				}
			}
			break
		}
		break
	}
}

func (a *Area) AddEntity(e Entity, source string, data Delta) {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	idStr := id.String()
	e.Create(a, idStr, source, data)
	a.Entities[idStr] = e
}

func (a *Area) Recv(source string, data map[string]Delta) {
	for k, v := range data {
		if e, ok := a.Entities[k]; ok {
			e.Recv(source, v)
		}
	}
}

func (a *Area) Tick(dt float64) {
	for _, e := range a.Entities {
		e.Tick(dt)
	}
}

func (a *Area) LateTick(dt float64) {

}

func (a *Area) Send() map[string]Entity {
	updates := 0
	data := make(map[string]Entity)
	for id, e := range a.Entities {
		delta := e.Send()
		// if !CompareDelta(delta, a.LastUpdate[id]) {
		data[id] = e
		a.LastUpdate[id] = delta
		updates++
		// } else {
		// 	data[id] = nil
		// }
	}

	if updates <= 0 {
		return nil
	}

	return data
}

func CompareDelta(a, b Delta) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil || len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
