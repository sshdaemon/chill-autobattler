package sim

// Brace: an enemy that dashes at the unit is stopped short and takes a
// counter-blow of Pct% (Braced spear).
func init() {
	registerEffect(Brace, &effectKind{
		name:  "brace",
		check: func(e Effect, _ *UnitDef) error { return pctIn(e, 1, 400) },
		onCharged: func(b *Battle, t, charger *Unit, a *active) {
			if charger.dash != nil { // stop a metre further out
				d := charger.dash
				d.to = Lerp(d.from, d.to, max(1, d.dur*3/4), d.dur)
			}
			b.emit(Event{Kind: EvBrace, Unit: t.Index, Other: charger.Index})
			b.damage(t, charger, t.Pos, Ratio(a.e.Pct, 100))
		},
	})
}
