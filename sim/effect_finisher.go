package sim

// Finisher: Pct% chance per direct hit to kill outright a victim below
// ExecuteBelowPct of its max health (Executioner, Headshot).
func init() {
	registerEffect(Finisher, &effectKind{
		name:  "finisher",
		check: func(e Effect, _ *UnitDef) error { return pctIn(e, 1, 100) },
		onBlow: func(b *Battle, att, v *Unit, scale, amount Fx, a *active) Fx {
			// instant kills only finish off the wounded, and only with a full hit (not splash)
			if scale == One && v.HP.MulDiv(100, ExecuteBelowPct) <= v.MaxHP && b.rng.IntN(100) < a.e.Pct {
				b.emit(Event{Kind: EvExecute, Unit: v.Index, Other: att.Index})
				return Max(amount, v.HP)
			}
			return amount
		},
	})
}
