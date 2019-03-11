package rpg

import (
	"log"
	"math/rand"
)

const MAX_PLAYER_TURN_TIME = 120

type ZoneCombatData struct {
	InCombat   bool          `json:"inCombat"`
	Waiting    bool          `json:"waiting"`
	Delay      int           `json:"delay"`
	Turn       int           `json:"turn"`
	Current    int           `json:"current"`
	Combatants []*CombatInfo `json:"combatants"`
}

type CombatInfo struct {
	Initiative int       `json:"initiative"`
	IsPlayer   bool      `json:"isPlayer"`
	Id         int       `json:"id"`
	Timer      int       `json:"timer"`
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
		log.Printf("zone %s entering combat!", z.Name)
		z.Dirty = true
		z.StartCombat()
	} else if !z.CombatInfo.InCombat && oldVal {
		log.Printf("zone %s exiting combat", z.Name)
		z.Dirty = true
	}

	if z.CombatInfo.InCombat {
		z.CheckCombatants()
	}

	return z.CombatInfo.InCombat || z.CombatInfo.InCombat != oldVal
}

func (z *Zone) StartCombat() {
	z.CombatInfo.Current = 0
	z.CombatInfo.Turn = 1
	z.CombatInfo.Combatants = nil
	z.CombatInfo.Combatants = make([]*CombatInfo, 0)
	for _, p := range z.Players {
		z.AddCombatant(p, false)
	}
	for _, n := range z.NPCs {
		z.AddCombatant(n, false)
	}
	if len(z.CombatInfo.Combatants) > 0 {
		z.CombatInfo.Combatants[0].Actor.NewTurn(z.CombatInfo.Combatants[0])
	}
	z.AddDelay(2)
}

func (z *Zone) AddDelay(ticks int) {
	z.CombatInfo.Waiting = true
	z.CombatInfo.Delay += ticks
}

func (z *Zone) CheckCombatants() {
	updatedCombatants := make([]*CombatInfo, 0)
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
			z.CombatInfo.Combatants[currentIdx].Actor.NewTurn(z.CombatInfo.Combatants[currentIdx])
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

	z.CombatInfo.Combatants = append(z.CombatInfo.Combatants, &info)
}

func (z *Zone) CanAct(player *Player) bool {
	ci := z.CombatInfo
	return !ci.InCombat || (!ci.Waiting &&
		len(ci.Combatants) > 0 &&
		ci.Combatants[ci.Current].IsPlayer &&
		ci.Combatants[ci.Current].Id == player.Id)
}

func (z *Zone) NextCombatant() {
	ci := z.CombatInfo
	next := ci.Current + 1
	if next >= len(ci.Combatants) {
		next = 0
		ci.Turn += 1
	}
	ci.Current = next
	ci.Combatants[next].Actor.NewTurn(ci.Combatants[next])
	z.AddDelay(4)
	z.Dirty = true
}

func (z *Zone) CombatTick() bool {
	ci := z.CombatInfo

	if len(ci.Combatants) <= 0 {
		return false
	}

	z.Dirty = true

	if ci.Waiting {
		if ci.Delay > 0 {
			ci.Delay -= 1
			return true
		} else {
			ci.Waiting = false
		}
	}

	z.PostCombatAction()

	return true
}

func (z *Zone) CheckAPCost(player *Player, cost int) bool {
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
	z.PostCombatAction()
}

func (z *Zone) PostCombatAction() {
	ci := z.CombatInfo
	current := ci.Combatants[ci.Current]

	current.Actor.Tick(current)
	for _, c := range ci.Combatants {
		if c.IsPlayer {
			p := c.Actor.(*Player)
			if p.HP <= 0 {
				z.Parent.KillPlayer(p)
			}
		} else {
			p := c.Actor.(*NPC)
			if p.HP <= 0 {
				z.Parent.KillNPC(z, p)
			}
		}
	}

	if current.Actor.IsTurnOver(current) {
		z.NextCombatant()
	}
}

type Combatant interface {
	InitCombat() CombatInfo
	Attack(Combatant)
	Damage(int)
	NewTurn(ci *CombatInfo)
	Tick(ci *CombatInfo)
	IsTurnOver(ci *CombatInfo) bool
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
	enemy.Damage(5)
}

func (n *NPC) Damage(amt int) {
	n.HP -= amt
}

func (n *NPC) NewTurn(ci *CombatInfo) {

}

func (n *NPC) Tick(ci *CombatInfo) {
	n.CombatTick()
}

func (n *NPC) IsTurnOver(ci *CombatInfo) bool {
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
	enemy.Damage(5)
}

func (p *Player) Damage(amt int) {
	p.HP -= amt
}

func (p *Player) NewTurn(ci *CombatInfo) {
	p.AP = 5
	ci.Timer = MAX_PLAYER_TURN_TIME
}

func (p *Player) Tick(ci *CombatInfo) {
	ci.Timer -= 1
}

func (p *Player) IsTurnOver(ci *CombatInfo) bool {
	return ci.Timer <= 0 || p.AP <= 0
}
