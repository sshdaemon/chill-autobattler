package sim

import "fmt"

// StatBonus: Stat +Pct% (AttackSpeed: cooldown and swing ×Pct%). Folded into
// the unit's stats at spawn (foldStats), not a hook.
func init() {
	registerEffect(StatBonus, &effectKind{name: "statBonus", check: func(e Effect, _ *UnitDef) error {
		if e.Stat == StatAttackSpeed {
			return pctIn(e, 1, 100)
		}
		if e.Stat < StatHP || e.Stat > StatAuraRadius {
			return fmt.Errorf("stat bonus on unknown stat %d", e.Stat)
		}
		return pctIn(e, -90, 400)
	}})
}

// statOrder is the fixed order stat bonuses apply in. Within one stat they
// apply in the roster's skill order, each as its own MulDiv, so roundings
// (and replays) never depend on the order skills were learned.
var statOrder = [...]StatKind{StatHP, StatDamage, StatRange, StatSpeed, StatAttackSpeed, StatAuraRadius}

// foldStats applies the stat bonuses of the unit's skills to its base stats.
// Extra effects (its side's plan) apply last.
func foldStats(def *UnitDef, skills []Skill, extra ...Effect) Stats {
	st := def.Stats
	for _, stat := range statOrder {
		for k := -1; k <= len(def.Skills); k++ { // the type's traits, its learned skills, then the extras
			var effects []Effect
			if k < 0 {
				effects = def.Traits
			} else if k == len(def.Skills) {
				effects = extra
			} else if hasSkill(skills, def.Skills[k].Skill) {
				effects = def.Skills[k].Effects
			}
			for _, e := range effects {
				if e.Kind != StatBonus || e.Stat != stat {
					continue
				}
				switch stat {
				case StatHP:
					st.HP = st.HP.MulDiv(100+e.Pct, 100)
				case StatDamage:
					st.DmgLo = st.DmgLo.MulDiv(100+e.Pct, 100)
					st.DmgHi = st.DmgHi.MulDiv(100+e.Pct, 100)
				case StatRange:
					st.Range = st.Range.MulDiv(100+e.Pct, 100)
				case StatSpeed:
					st.Speed = st.Speed.MulDiv(100+e.Pct, 100)
				case StatAuraRadius:
					st.AuraPct += e.Pct
				case StatAttackSpeed: // faster blows: cooldown and swing ×Pct%
					st.Cooldown = st.Cooldown * e.Pct / 100
					st.SwingTicks = st.SwingTicks * e.Pct / 100
				}
			}
		}
	}
	return st
}
