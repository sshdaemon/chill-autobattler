package sim

import "fmt"

// A trigger switches a tactic's group to its second stance (plan.go). Each
// kind lives in its own trigger_<name>.go and registers itself under its id
// from ids.go. Triggers are checked once a tick, before units think.
type triggerKind struct {
	name string
	// check validates the trigger's number.
	check func(t Trigger) error
	// fired reports whether the trigger fires for team's group g this tick.
	fired func(b *Battle, team, g int, t Trigger) bool
}

var triggerKinds [256]*triggerKind

func registerTrigger(id TriggerKind, k *triggerKind) {
	if id == 0 || triggerKinds[id] != nil {
		panic(fmt.Sprintf("trigger kind %d registered twice or as 0: allocate ids in ids.go", id))
	}
	triggerKinds[id] = k
}

func (k TriggerKind) String() string {
	if t := triggerKinds[k]; t != nil {
		return t.name
	}
	return fmt.Sprintf("trigger(%d)", k)
}

// checkTrigger validates a group's trigger; kind 0 (never) needs no number.
func checkTrigger(t Trigger) error {
	if t.Kind == 0 {
		if t != (Trigger{}) {
			return fmt.Errorf("a group without a trigger has a number: %+v", t)
		}
		return nil
	}
	k := triggerKinds[t.Kind]
	if k == nil {
		return fmt.Errorf("unknown trigger kind %d", t.Kind)
	}
	if k.check != nil {
		return k.check(t)
	}
	if t.Seconds != 0 {
		return fmt.Errorf("trigger %s takes no number", k.name)
	}
	return nil
}

// checkTriggers switches every tactic group whose trigger fires, teams and
// groups in order, and rebuilds its living units' temperaments.
func (b *Battle) checkTriggers() {
	for team := 0; team < 2; team++ {
		p := PlanOf(b.Plans[team])
		if p == nil || !p.IsTactic() {
			continue
		}
		for g := range p.Groups {
			t := p.Groups[g].When
			if b.switched[team][g] || t.Kind == 0 || !triggerKinds[t.Kind].fired(b, team, g, t) {
				continue
			}
			b.switched[team][g] = true
			for _, u := range b.Units {
				if u.Team == team && u.Group == g && !u.Dead {
					u.Temper = temperOf(u.Def, p, g, true, &u.kit)
					u.held, u.lineSet = 0, false
				}
			}
		}
	}
}
