package rpg

import (
	"log"
	"math/rand"
)

type ZoneCombatData struct {
	InCombat   bool         `json:"inCombat"`
	Turn       int          `json:"turn"`
	Current    int          `json:"current"`
	Combatants []CombatInfo `json:"combatants"`
}

type CombatInfo struct {
	Initiative int       `json:"initiative"`
	IsPlayer   bool      `json:"isPlayer"`
	Id         int       `json:"id"`
	Actor      Combatant `json:"-"`
}

func (z *Zone) CheckCombat() bool {
	oldVal := z.CombatInfo.InCombat

	hostiles := false
	for _, n := range z.NPCs {
		if n.Alignment == "hostile" {
			hostiles = true
		}
	}
	z.CombatInfo.InCombat = len(z.Players) > 0 && hostiles

	// if we've just entered combat, i.e. previously false now true
	if z.CombatInfo.InCombat && !oldVal {
		z.StartCombat()
	}

	if z.CombatInfo.InCombat {
		z.CheckCombatants()
	}

	return z.CombatInfo.InCombat || z.CombatInfo.InCombat != oldVal
}

func (z *Zone) StartCombat() {
	z.CombatInfo.Combatants = nil
	z.CombatInfo.Combatants = make([]CombatInfo, 0)
	for _, p := range z.Players {
		z.AddCombatant(p, false)
	}
	for _, n := range z.NPCs {
		z.AddCombatant(n, false)
	}
	if len(z.CombatInfo.Combatants) > 0 {
		z.CombatInfo.Combatants[0].Actor.NewTurn()
	}
}

func (z *Zone) CheckCombatants() {
	updatedCombatants := make([]CombatInfo, 0)
	changed := false
	currentChanged := false
	current := z.CombatInfo.Combatants[z.CombatInfo.Current]
	currentIdx := z.CombatInfo.Current
	needNewCurrent := false
	count := 0
	for _, info := range z.CombatInfo.Combatants {
		exists := false
		isCurrent := info.IsPlayer == current.IsPlayer && info.Id == current.Id
		if info.IsPlayer {
			for id := range z.Players {
				if id == info.Id {
					exists = true
					break
				}
			}
		} else {
			for id := range z.NPCs {
				if id == info.Id {
					exists = true
					break
				}
			}
		}

		if exists {
			if isCurrent || needNewCurrent {
				currentChanged = needNewCurrent
				needNewCurrent = false
				currentIdx = count
			}
			updatedCombatants = append(updatedCombatants, info)
			count += 1
		} else {
			if isCurrent {
				needNewCurrent = true
				log.Printf("current combatant in zone %s left", z.Name)
			}
			changed = true
		}
	}

	if changed {
		// actor was last in list
		if needNewCurrent {
			currentIdx = 0
			currentChanged = true
		}
		z.CombatInfo.Current = currentIdx
		z.CombatInfo.Combatants = updatedCombatants
		if currentChanged {
			z.CombatInfo.Combatants[currentIdx].Actor.NewTurn()
		}
	}

	for _, p := range z.Players {
		z.AddCombatant(p, true)
	}
	for _, n := range z.NPCs {
		z.AddCombatant(n, true)
	}
}

func (z *Zone) AddCombatant(c Combatant, late bool) {
	info := c.InitCombat()
	exists := false
	for _, i := range z.CombatInfo.Combatants {
		if i.IsPlayer == info.IsPlayer && i.Id == info.Id {
			exists = true
			break
		}
	}
	// if were're already added, just fail silently
	if exists {
		return
	}

	// if late {
	// } else {
	// }

	z.CombatInfo.Combatants = append(z.CombatInfo.Combatants, info)
}

func (z *Zone) CanAct(player *Player) bool {
	ci := z.CombatInfo
	return !ci.InCombat || (len(ci.Combatants) > 0 && ci.Combatants[ci.Current].IsPlayer && ci.Combatants[ci.Current].Id == player.Id)
}

func (z *Zone) NextCombatant() {
	ci := z.CombatInfo
	next := ci.Current + 1
	if next >= len(ci.Combatants) {
		next = 0
		ci.Turn += 1
	}
	ci.Current = next
	ci.Combatants[next].Actor.NewTurn()

	z.Parent.Outgoing <- OutgoingMessage{
		Zone: z.Name,
		Type: ACTION_UPDATE,
	}
}

func (z *Zone) CombatTick() bool {
	ci := z.CombatInfo
	if len(ci.Combatants) <= 0 {
		return false
	}

	current := ci.Combatants[ci.Current]
	// if we're waiting on a player to act do nothing
	if current.IsPlayer {
		return false
	}

	current.Actor.Tick()
	if current.Actor.IsTurnOver() {
		z.NextCombatant()
	}

	return true
}

func (z *Zone) CheckAPCost(player *Player, cost int) bool {
	if !z.CombatInfo.InCombat {
		return true
	}
	if player.AP >= cost {
		player.AP -= cost
		return true
	}
	return false
}

func (z *Zone) PostPlayerAction(player *Player) {
	if !z.CombatInfo.InCombat {
		return
	}
	if player.IsTurnOver() {
		z.NextCombatant()
	}
}

type Combatant interface {
	InitCombat() CombatInfo
	Attack(Combatant)
	Damage(int)
	NewTurn()
	Tick()
	IsTurnOver() bool
}

func (n *NPC) InitCombat() CombatInfo {
	return CombatInfo{
		Initiative: rand.Intn(20),
		IsPlayer:   false,
		Id:         n.Id,
		Actor:      n,
	}
}

func (n *NPC) Attack(enemy Combatant) {
	enemy.Damage(1)
}

func (n *NPC) Damage(amt int) {
	n.HP -= amt
}

func (n *NPC) NewTurn() {

}

func (n *NPC) Tick() {
	n.CombatTick()
}

func (n *NPC) IsTurnOver() bool {
	return true
}

func (p *Player) InitCombat() CombatInfo {
	return CombatInfo{
		Initiative: rand.Intn(20),
		IsPlayer:   true,
		Id:         p.Id,
		Actor:      p,
	}
}

func (p *Player) Attack(enemy Combatant) {
	enemy.Damage(1)
}

func (p *Player) Damage(amt int) {
	p.HP -= amt
}

func (p *Player) NewTurn() {
	p.AP = 5
}

func (p *Player) Tick() {

}

func (p *Player) IsTurnOver() bool {
	return p.AP <= 0
}
