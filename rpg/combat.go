package rpg

import (
	"fmt"
	"log"
)

const MAX_PLAYER_TURN_TIME = 120

type ZoneCombatData struct {
	InCombat          bool
	Waiting           bool
	Delay             int
	Turn              int
	Current           Combatant
	CurrentInitiative int
	SpellData         *CombatSpellData
	Combatants        map[Combatant]*CombatInfo
}

type CombatInfo struct {
	Initiative int  `json:"initiative"`
	IsPlayer   bool `json:"isPlayer"`
	Id         int  `json:"id"`
	Timer      int  `json:"timer"`
}

type DamageInfo struct {
	Amount int    `json:"amount"`
	Crit   bool   `json:"crit"`
	Type   string `json:"type"`
}

type Combatant interface {
	GetName() string
	InitCombat() CombatInfo
	Attack() DamageInfo
	Damage(DamageInfo)
	NewTurn(ci *CombatInfo)
	Tick(g *RPG, z *Zone, ci *CombatInfo)
	IsTurnOver(ci *CombatInfo) bool
}

func (g *RPG) CheckCombat(z *Zone) {
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
		g.Zones.SetDirty(z.Id)
		g.StartCombat(z)
	} else if !z.CombatInfo.InCombat && oldVal {
		log.Printf("zone %s exiting combat", z.Name)
		g.Zones.SetDirty(z.Id)
	}
}

func (g *RPG) StartCombat(z *Zone) {
	ci := z.CombatInfo
	ci.SpellData = &CombatSpellData{}
	ci.Current = nil
	ci.Turn = 1
	ci.Combatants = nil
	ci.Combatants = make(map[Combatant]*CombatInfo)
	for _, p := range z.Players {
		ci.AddCombatant(p, false)
	}
	for _, n := range z.NPCs {
		ci.AddCombatant(n, false)
	}
	if ci.Current != nil {
		ci.Current.NewTurn(ci.Combatants[ci.Current])
	}
	ci.AddDelay(2)
}

func (g *RPG) CheckCombatants(z *Zone) {
	toRemove := make([]Combatant, 0)
	currentLeft := false
	for c, info := range z.CombatInfo.Combatants {
		exists := false
		if info.IsPlayer {
			_, exists = z.Players[info.Id]
		} else {
			_, exists = z.NPCs[info.Id]
		}

		if !exists {
			if c == z.CombatInfo.Current {
				currentLeft = true
				log.Printf("current combatant in zone %s left", z.Name)
			}
			toRemove = append(toRemove, c)
		}
	}

	for _, c := range toRemove {
		delete(z.CombatInfo.Combatants, c)
	}

	if currentLeft {
		z.CombatInfo.NextCombatant()
		z.CombatInfo.Current.NewTurn(z.CombatInfo.Combatants[z.CombatInfo.Current])
	}

	for _, p := range z.Players {
		z.CombatInfo.AddCombatant(p, true)
	}
	for _, n := range z.NPCs {
		z.CombatInfo.AddCombatant(n, true)
	}
}

func (g *RPG) CheckAlive(z *Zone) {
	for c, info := range z.CombatInfo.Combatants {
		if info.IsPlayer {
			p := c.(*Player)
			if p.HP <= 0 {
				g.KillPlayer(p)
			}
		} else {
			p := c.(*NPC)
			if p.HP <= 0 {
				g.KillNPC(z, p)
			}
		}
	}
}

func (g *RPG) CanAct(z *Zone, player *Player) bool {
	ci := z.CombatInfo
	return !ci.InCombat || (!ci.Waiting &&
		len(ci.Combatants) > 0 &&
		ci.Combatants[ci.Current].IsPlayer &&
		ci.Combatants[ci.Current].Id == player.Id)
}

func (p *Player) CheckAPCost(cost int) bool {
	if p.Editing {
		return true
	}
	if p.AP >= cost {
		p.AP -= cost
		return true
	}
	return false
}

func (g *RPG) CombatTick(z *Zone) bool {
	ci := z.CombatInfo

	if len(ci.Combatants) <= 0 {
		return false
	}

	g.Zones.SetDirty(z.Id)

	spellRunning := ci.SpellData.Tick(g, z)
	if spellRunning {
		return true
	}

	if ci.Waiting {
		if ci.Delay > 0 {
			ci.Delay -= 1
			return true
		} else {
			ci.Waiting = false
		}
	}

	g.PostCombatAction(z)

	return true
}

func (g *RPG) PostPlayerAction(z *Zone, player *Player) {
	if !z.CombatInfo.InCombat {
		return
	}
	g.PostCombatAction(z)
}

func (g *RPG) PostCombatAction(z *Zone) {
	ci := z.CombatInfo
	current, ok := ci.Combatants[ci.Current]
	if !ok {
		log.Printf("missing combatant")
		g.CheckAlive(z)
		g.CheckCombat(z)
		ci.NextCombatant()
		return
	}

	g.CheckAlive(z)
	ci.Current.Tick(g, z, current)
	g.CheckAlive(z)
	g.CheckCombat(z)

	if ci.Current.IsTurnOver(current) {
		ci.NextCombatant()
	}
}

func (g *RPG) DoMeleeAttack(z *Zone, origin Combatant, target Combatant) {
	dmg := origin.Attack()
	target.Damage(dmg)
	var msg string
	if dmg.Crit {
		msg = "%s attacked %s for %d damage (CRITICAL)"
	} else {
		msg = "%s attacked %s for %d damage"
	}
	g.SendMessage(z, nil, fmt.Sprintf(msg,
		origin.GetName(), target.GetName(), dmg.Amount))
}

func (g *RPG) DoSpellAttack(z *Zone, origin *Player, spell SpellDef, x, y int) {
	delay := z.CombatInfo.SpellData.RunSpell(g, z, spell, origin.X, origin.Y, x, y)
	if delay {
		z.CombatInfo.Waiting = true
	}
}

func (d *ZoneCombatData) AddDelay(ticks int) {
	d.Delay += ticks
	d.Waiting = true
}

func (d *ZoneCombatData) AddCombatant(c Combatant, late bool) {
	info := c.InitCombat()
	if _, exists := d.Combatants[c]; exists {
		return
	}
	if late {
		lowest := d.CurrentInitiative
		for _, info := range d.Combatants {
			if info.Initiative < lowest {
				lowest = info.Initiative
			}
		}
		info.Initiative = lowest - 1
	} else {
		info.Initiative *= 1000
	}
	inserted := false
	for !inserted {
		overlap := false
		for _, i := range d.Combatants {
			if i.Initiative == info.Initiative {
				overlap = true
				break
			}
		}
		if overlap {
			info.Initiative -= 1
		} else {
			inserted = true
		}
	}
	d.Combatants[c] = &info
}

func (d *ZoneCombatData) NextCombatant() {
	hasNext := false
	highest := -9999999
	for c, info := range d.Combatants {
		if info.Initiative < d.CurrentInitiative && info.Initiative > highest {
			d.Current = c
			highest = info.Initiative
			hasNext = true
		}
	}
	if !hasNext {
		highest = -9999999
		for c, info := range d.Combatants {
			if info.Initiative > highest {
				d.Current = c
				highest = info.Initiative
			}
		}
		d.Turn += 1
	}
	info := d.Combatants[d.Current]
	d.CurrentInitiative = info.Initiative
	d.Current.NewTurn(info)
	d.AddDelay(4)
}
