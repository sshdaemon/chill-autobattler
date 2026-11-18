package sim

import "fmt"

// GapCloser: rush at Speed at a target up to Value away, stopping at weapon
// range; Ticks cooldown (Dash). Fires when the target is further than range +
// DashMinGap, within reach, and in sight.
func init() {
	registerEffect(GapCloser, &effectKind{
		name: "gapCloser",
		check: func(e Effect, _ *UnitDef) error {
			if e.Value <= 0 || e.Speed <= 0 || e.Ticks <= 0 {
				return fmt.Errorf("gap-closer needs reach, speed and cooldown")
			}
			return nil
		},
		approach: func(b *Battle, u, t *Unit, d Fx, a *active) bool {
			g := a.e
			if a.cd > 0 || d <= u.Stats.Range+DashMinGap || d > g.Value || !b.Map.LineOfSight(u.Pos, t.Pos) {
				return false
			}
			dir := t.Pos.Sub(u.Pos).Norm()
			reach := d - u.Stats.Range.MulDiv(8, 10)
			dur := reach.Div(perTick(g.Speed)).Ceil()
			if dur < 1 {
				dur = 1
			}
			u.dash = &dash{from: u.Pos, to: u.Pos.Add(dir.Scale(reach)), dur: dur}
			u.Face = dir
			a.cd = g.Ticks
			u.path = nil
			b.emit(Event{Kind: EvDash, Unit: u.Index, Other: -1})
			for i := range t.kit.active { // the target may brace against it
				if ta := &t.kit.active[i]; ta.kind.onCharged != nil {
					ta.kind.onCharged(b, t, u, ta)
				}
			}
			return true
		},
	})
}
