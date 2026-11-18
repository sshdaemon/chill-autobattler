package sim

// Shot is an arrow or magic bolt in flight. It homes in on its victim and
// hits if the victim is still alive when it arrives.
type Shot struct {
	Shooter, Victim int
	Bolt, Lob       bool
	Start, Aim      Vec
	T, Dur          int // ticks flown / total flight time
}

func (b *Battle) shoot(u, v *Unit) {
	start := u.Pos.Add(u.Face.Scale(ProjectileFwd))
	dur := Dist(start, v.Pos).Div(perTick(u.Stats.ShotSpeed)).Ceil()
	if dur < MinShotTicks {
		dur = MinShotTicks
	}
	b.Shots = append(b.Shots, &Shot{
		Shooter: u.Index, Victim: v.Index, Bolt: u.Def.Projectile == Bolt, Lob: u.kit.overCover,
		Start: start, Aim: v.Pos, Dur: dur,
	})
	b.emit(Event{Kind: EvShot, Unit: u.Index, Other: v.Index})
}

// Pos is where the projectile is now on the ground plane (for rendering).
func (s *Shot) Pos() Vec { return Lerp(s.Start, s.Aim, s.T, s.Dur) }

func (b *Battle) updateShots() {
	kept := b.Shots[:0]
	for _, s := range b.Shots {
		v := b.Units[s.Victim]
		if !v.Dead {
			s.Aim = v.Pos
		}
		s.T++
		if s.T < s.Dur {
			kept = append(kept, s)
			continue
		}
		if v.Dead { // missed: the arrow sticks or splashes, the bolt fizzles
			continue
		}
		u := b.Units[s.Shooter]
		b.damage(u, v, s.Start, One)
		for i := range u.kit.active { // on-hit effects, in kind order (slow, then splash)
			if a := &u.kit.active[i]; a.kind.onShotLands != nil {
				a.kind.onShotLands(b, u, v, s, a)
			}
		}
	}
	for i := len(kept); i < len(b.Shots); i++ {
		b.Shots[i] = nil
	}
	b.Shots = kept
}

// damage applies one hit: roll × flank × matchup × scale, then instant-kill
// skills. `from` is where the blow or projectile came from.
func (b *Battle) damage(att, v *Unit, from Vec, scale Fx) {
	if v.Dead || b.Result != Ongoing {
		return
	}
	amount := b.rng.Between(att.Stats.DmgLo, att.Stats.DmgHi)
	var flags uint8
	if v.Face.Dot(from.Sub(v.Pos).Norm()) < CosFront { // outside the ±60° front cone
		amount = amount.Mul(FlankMultiplier)
		flags |= FlagFlank
	}
	m := MatchupOf(att.Def, v.Def)
	amount = amount.Mul(m).Mul(scale)
	if m > One {
		flags |= FlagStrong
	} else if m < One {
		flags |= FlagWeak
	}
	for i := range att.kit.active { // effects that change a blow (finisher)
		if a := &att.kit.active[i]; a.kind.onBlow != nil {
			amount = a.kind.onBlow(b, att, v, scale, amount, a)
		}
	}
	att.Dealt += Min(amount, v.HP)
	v.HP -= amount
	b.emit(Event{Kind: EvHit, Unit: v.Index, Other: att.Index, Amount: amount, Flags: flags})
	if v.HP <= 0 {
		v.HP, v.Dead = 0, true
		v.moveTarget, v.swing, v.dash = nil, nil, nil
		att.Kills++
		b.emit(Event{Kind: EvDeath, Unit: v.Index, Other: att.Index})
	} else if v.swing == nil {
		v.busy = StaggerTicks
	}
}
