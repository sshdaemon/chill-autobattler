package sim

// Shot is an arrow or magic bolt in flight. It homes in on its victim and
// hits if the victim is still alive when it arrives.
type Shot struct {
	Shooter, Victim int
	Bolt, Lob       bool
	Start, Aim      Vec
	T, Dur          int // ticks flown / total flight time
	// a thrown burst (spear, fire pot): its damage share and splash
	burst     bool
	scale     Fx
	splash    Fx
	splashPct int
}

// throw launches a burst attack (Burst effect) at v: thrown at e.Speed, ×e.Factor
// damage, splashing e.Pct% within e.Radius if set.
func (b *Battle) throw(u, v *Unit, e *Effect) {
	start := u.Pos.Add(u.Face.Scale(ProjectileFwd))
	dur := max(MinShotTicks, Dist(start, v.Pos).Div(perTick(e.Speed)).Ceil())
	b.Shots = append(b.Shots, &Shot{Shooter: u.Index, Victim: v.Index, Lob: true, Start: start, Aim: v.Pos, Dur: dur,
		burst: true, scale: e.Factor, splash: e.Radius, splashPct: e.Pct})
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
		if s.burst {
			b.damage(u, v, s.Start, s.scale)
			if s.splash > 0 {
				b.emit(Event{Kind: EvBlast, Unit: v.Index, Other: u.Index})
				b.eachNear(v.Pos, s.splash, v.Team, false, v, func(o *Unit) { b.damage(u, o, s.Start, s.scale.MulDiv(s.splashPct, 100)) })
			}
			continue
		}
		b.damage(u, v, s.Start, One)
		b.landed(u, v, s.Start)
	}
	for i := len(kept); i < len(b.Shots); i++ {
		b.Shots[i] = nil
	}
	b.Shots = kept
}

// landed runs the attacker's on-hit effects after a direct hit, in kind order.
func (b *Battle) landed(att, v *Unit, from Vec) {
	for i := range att.kit.active {
		if a := &att.kit.active[i]; a.kind.onHit != nil {
			a.kind.onHit(b, att, v, from, a)
		}
	}
}

// damage applies one hit: a dodge, roll × flank × matchup × scale, the
// modifiers, the victim's own effects, the attacker's effects (instant kills
// last) and guards. `from` is where the blow or projectile came from.
func (b *Battle) damage(att, v *Unit, from Vec, scale Fx) {
	if v.Dead || b.Result != Ongoing {
		return
	}
	if ev := v.mods[ModEvade]; ev > 0 && b.rng.IntN(100) < ev {
		b.emit(Event{Kind: EvMiss, Unit: v.Index, Other: att.Index})
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
	amount = pctOf(amount, att.mods[ModDamage])
	for i := range v.kit.active { // the victim's own protection (shield wall)
		if a := &v.kit.active[i]; a.kind.onHurt != nil {
			amount = a.kind.onHurt(b, v, att, from, amount, a)
		}
	}
	amount = pctOf(amount, v.mods[ModTaken])
	for i := range att.kit.active { // effects that change a blow (bonus against, finisher)
		if a := &att.kit.active[i]; a.kind.onBlow != nil {
			amount = a.kind.onBlow(b, att, v, scale, amount, a)
		}
	}
	for _, g := range b.Units { // allies that take part of it (guard), in unit order
		if g == v || g.Dead || g.Team != v.Team {
			continue
		}
		for i := range g.kit.active {
			if a := &g.kit.active[i]; a.kind.intercept != nil {
				if share := a.kind.intercept(b, g, v, att, amount, a); share > 0 {
					amount -= share
					b.emit(Event{Kind: EvGuard, Unit: g.Index, Other: v.Index, Amount: share})
					b.hurt(att, g, share, 0)
				}
			}
		}
	}
	b.hurt(att, v, amount, flags)
}

// hurt takes an amount of health from v (already computed), with the events,
// kill credit, death and stagger.
func (b *Battle) hurt(att, v *Unit, amount Fx, flags uint8) {
	if v.Dead {
		return
	}
	att.Dealt += Min(amount, v.HP)
	v.HP -= amount
	b.emit(Event{Kind: EvHit, Unit: v.Index, Other: att.Index, Amount: amount, Flags: flags})
	if v.HP <= 0 {
		v.HP, v.Dead = 0, true
		v.moveTarget, v.swing, v.dash = nil, nil, nil
		att.Kills++
		b.emit(Event{Kind: EvDeath, Unit: v.Index, Other: att.Index})
		for _, o := range b.Units { // effects that answer a death (rally), in unit order
			if o.Dead {
				continue
			}
			for i := range o.kit.active {
				if a := &o.kit.active[i]; a.kind.onDeath != nil {
					a.kind.onDeath(b, o, v, a)
				}
			}
		}
	} else if v.swing == nil {
		v.busy = StaggerTicks
	}
}
