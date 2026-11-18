package sim

import "fmt"

// Escape: teleport Value away when melee enemies are within Radius; Ticks
// cooldown (Blink). Away from the threats' centre, turning (BlinkTurns) until
// a free spot inside the field is found.
func init() {
	registerEffect(Escape, &effectKind{
		name: "escape",
		check: func(e Effect, _ *UnitDef) error {
			if e.Value <= 0 || e.Radius <= 0 || e.Ticks <= 0 {
				return fmt.Errorf("escape needs distance, threat radius and cooldown")
			}
			return nil
		},
		beforeAct: func(b *Battle, u *Unit, a *active) bool {
			if a.cd > 0 {
				return false
			}
			esc := a.e
			var sum Vec
			n := 0
			for _, e := range b.Units {
				if !e.Dead && e.Team != u.Team && !e.Stats.Ranged && Within(u.Pos, e.Pos, esc.Radius) {
					sum = sum.Add(e.Pos)
					n++
				}
			}
			if n == 0 {
				return false
			}
			centre := V(sum.X/Fx(n), sum.Z/Fx(n))
			base := u.Pos.Sub(centre).Norm()
			if base.IsZero() {
				base = V(0, One)
			}
			for _, turn := range BlinkTurns {
				p := u.Pos.Add(base.Rot(turn).Scale(esc.Value))
				if p.X.Abs() > EdgeLimit || p.Z.Abs() > EdgeLimit || !b.Map.FreeAt(p) {
					continue
				}
				b.emit(Event{Kind: EvBlink, Unit: u.Index, Other: -1, Amount: 0})
				u.Pos = p
				u.moveTarget, u.path, u.swing = nil, nil, nil
				a.cd = esc.Ticks
				return true
			}
			return false
		},
	})
}
