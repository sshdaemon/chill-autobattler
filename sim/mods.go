package sim

// Modifiers: what auras, timed buffs and debuffs and conditional effects do
// to a unit this tick. They are recomputed at the start of every tick (no
// state carries over except the timed buffs) and applied only when non-zero,
// so units without such effects play exactly as before.

// ModKind is what a modifier changes, in percent.
type ModKind uint8

const (
	ModDamage      ModKind = iota + 1 // damage dealt
	ModTaken                          // damage taken
	ModAttackSpeed                    // attacks per second (+15: 15% faster)
	ModEvade                          // chance to dodge a blow
	ModSpeed                          // movement speed
)

var modNames = [...]string{"", "damage", "taken", "attackSpeed", "evade", "speed"}

func (m ModKind) String() string {
	if int(m) < len(modNames) {
		return modNames[m]
	}
	return "mod?"
}

type mods [len(modNames)]int

// buff is a timed modifier on a unit (rally, curse).
type buff struct {
	mod   ModKind
	pct   int
	ticks int
}

// Cond is when a DamageTaken effect applies.
type Cond uint8

const (
	CondAlways   Cond = iota
	CondRanged        // against arrows, bolts and thrown attacks
	CondFront         // against blows from the front
	CondStanding      // while not moving
)

// addBuff gives a unit a timed modifier; a second one of the same kind and
// size refreshes the first instead of stacking.
func (u *Unit) addBuff(m ModKind, pct, ticks int) {
	for i := range u.buffs {
		if u.buffs[i].mod == m && u.buffs[i].pct == pct {
			u.buffs[i].ticks = max(u.buffs[i].ticks, ticks)
			return
		}
	}
	u.buffs = append(u.buffs, buff{m, pct, ticks})
}

// computeMods sets every living unit's modifiers for this tick: timed buffs,
// then auras (in unit order), then each unit's own conditional effects.
func (b *Battle) computeMods() {
	for _, u := range b.Units {
		u.mods = mods{}
		if u.Dead {
			continue
		}
		kept := u.buffs[:0]
		for _, f := range u.buffs {
			if f.ticks--; f.ticks >= 0 {
				u.mods[f.mod] += f.pct
				kept = append(kept, f)
			}
		}
		u.buffs = kept
		if u.fear > 0 {
			u.fear--
		}
	}
	for _, u := range b.Units {
		if u.Dead {
			continue
		}
		for i := range u.kit.active {
			if a := &u.kit.active[i]; a.kind.aura != nil {
				a.kind.aura(b, u, a)
			}
		}
	}
	for _, u := range b.Units {
		if u.Dead {
			continue
		}
		for i := range u.kit.active {
			if a := &u.kit.active[i]; a.kind.self != nil {
				a.kind.self(b, u, a)
			}
		}
	}
}

// auraRadius is an aura's reach for its source, with StatAuraRadius applied.
func auraRadius(src *Unit, r Fx) Fx {
	if src.Stats.AuraPct == 0 {
		return r
	}
	return r.MulDiv(100+src.Stats.AuraPct, 100)
}

// eachNear calls f for every living unit within r of p, allies of team or
// (foes) its enemies, in unit order, never `skip`.
func (b *Battle) eachNear(p Vec, r Fx, team int, foes bool, skip *Unit, f func(*Unit)) {
	for _, o := range b.Units {
		if o.Dead || o == skip || (o.Team != team) != foes || !Within(p, o.Pos, r) {
			continue
		}
		f(o)
	}
}

// pctOf applies a percentage change to an amount (nothing when pct is 0).
func pctOf(v Fx, pct int) Fx {
	if pct == 0 {
		return v
	}
	return v.MulDiv(100+pct, 100)
}

// cooldownOf scales an attack cooldown by the attack-speed modifier.
func cooldownOf(cd, attackSpeedPct int) int {
	if attackSpeedPct == 0 {
		return cd
	}
	return cd * 100 / max(10, 100+attackSpeedPct)
}

// zone is burning ground: enemies of team standing in it take Pct% of the
// source's blow every second for its remaining ticks.
type zone struct {
	src        int
	team       int
	pos        Vec
	radius     Fx
	pct, ticks int
}

func (b *Battle) updateZones() {
	kept := b.zones[:0]
	for _, z := range b.zones {
		if z.ticks--; z.ticks < 0 {
			continue
		}
		if z.ticks%TicksPerSecond == 0 {
			src := b.Units[z.src]
			b.eachNear(z.pos, z.radius, z.team, true, nil, func(v *Unit) { b.damage(src, v, z.pos, Ratio(z.pct, 100)) })
		}
		kept = append(kept, z)
	}
	b.zones = kept
}
