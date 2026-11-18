package sim

// Snapshot layout, all int32 (positions and HP are raw fixed-point, ×65536).
// The browser reads it as an Int32Array; keep web/src/sim/snapshot.ts in sync.
//
//	header: tick, result, nUnits, nShots, nEvents
//	unit:   team, class, x, z, faceX, faceZ, hp, maxHp, flags, target, skills, kills, dealt
//	shot:   flags, startX, startZ, aimX, aimZ, t, dur, victim
//	event:  kind, unit, other, amount, flags
const (
	SnapHeader = 5
	SnapUnit   = 13
	SnapShot   = 8
	SnapEvent  = 5
)

// Unit flags in a snapshot.
const (
	UFDead int32 = 1 << iota
	UFMoving
	UFSwinging
	UFWet
	UFSlowed
	UFKiting
	UFDashing
	UFStaggered
	UFSeen // Blue's team vision sees this unit (Blue's own units always)
)

// Shot flags in a snapshot.
const (
	SFBolt int32 = 1 << iota
	SFLob
	SFBurst // thrown by a Burst effect (spear, fire pot), not shot
)

// Snapshot encodes the current state plus the events since the last call.
func (b *Battle) Snapshot() []int32 {
	events := b.TakeEvents()
	out := make([]int32, 0, SnapHeader+len(b.Units)*SnapUnit+len(b.Shots)*SnapShot+len(events)*SnapEvent)
	out = append(out, int32(b.Tick), int32(b.Result), int32(len(b.Units)), int32(len(b.Shots)), int32(len(events)))
	for _, u := range b.Units {
		var f int32
		for _, bit := range []struct {
			on   bool
			flag int32
		}{
			{u.Dead, UFDead}, {u.Moving(), UFMoving}, {u.swing != nil, UFSwinging}, {u.Wet, UFWet},
			{u.slow > 0, UFSlowed}, {u.Kiting, UFKiting}, {u.dash != nil, UFDashing}, {u.busy > 0, UFStaggered},
			{b.SeenBy(0, u), UFSeen},
		} {
			if bit.on {
				f |= bit.flag
			}
		}
		out = append(out, int32(u.Team), int32(u.Class), int32(u.Pos.X), int32(u.Pos.Z), int32(u.Face.X), int32(u.Face.Z),
			int32(u.HP), int32(u.MaxHP), f, int32(u.Target), int32(len(u.Skills)), int32(u.Kills), int32(u.Dealt))
	}
	for _, s := range b.Shots {
		var f int32
		if s.Bolt {
			f |= SFBolt
		}
		if s.Lob {
			f |= SFLob
		}
		if s.burst {
			f |= SFBurst
		}
		out = append(out, f, int32(s.Start.X), int32(s.Start.Z), int32(s.Aim.X), int32(s.Aim.Z), int32(s.T), int32(s.Dur), int32(s.Victim))
	}
	for _, e := range events {
		out = append(out, int32(e.Kind), int32(e.Unit), int32(e.Other), int32(e.Amount), int32(e.Flags))
	}
	return out
}

// MapData encodes the obstacles as kind, x, z, r per obstacle (raw fixed-point).
func (b *Battle) MapData() []int32 {
	out := make([]int32, 0, len(b.Map.Obstacles)*4)
	for _, o := range b.Map.Obstacles {
		out = append(out, int32(o.Kind), int32(o.Pos.X), int32(o.Pos.Z), int32(o.R))
	}
	return out
}
