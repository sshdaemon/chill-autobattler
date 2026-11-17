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
		Shooter: u.Index, Victim: v.Index, Bolt: u.Class == Mage, Lob: u.Has(SkillLob),
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
		if !v.Dead && u.Has(SkillFrost) && b.rng.IntN(100) < u.skill(SkillFrost).Pct {
			v.slow = FrostTicks
			b.emit(Event{Kind: EvFrost, Unit: v.Index, Other: u.Index})
		}
		if u.Has(SkillBlast) {
			b.emit(Event{Kind: EvBlast, Unit: v.Index, Other: u.Index})
			for _, e := range b.Units {
				if e != v && !e.Dead && e.Team == v.Team && Within(e.Pos, v.Pos, BlastRadius) {
					b.damage(u, e, s.Start, Ratio(u.skill(SkillBlast).Pct, 100))
				}
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
	m := Matchup[att.Class][v.Class]
	amount = amount.Mul(m).Mul(scale)
	if m > One {
		flags |= FlagStrong
	} else if m < One {
		flags |= FlagWeak
	}
	if scale == One && att.Has(SkillExecute) && b.rng.IntN(100) < att.skill(SkillExecute).Pct {
		amount = Max(amount, v.HP)
		b.emit(Event{Kind: EvExecute, Unit: v.Index, Other: att.Index})
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
