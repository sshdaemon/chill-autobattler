package sim

import "testing"

// unitWith is a copy of a roster entry whose first skill is the given effects
// (and optional traits), added to the roster for the test.
func unitWith(t *testing.T, base Class, name string, traits []Effect, effects ...Effect) Class {
	d := Roster[base]
	d.Name = name
	d.InRedPool = false
	d.Traits = traits
	d.Skills[0].Effects = effects
	withRoster(t, d)
	return Class(len(Roster) - 1)
}

// blow is the damage att deals v with one direct hit, in a fresh battle with
// a fixed random stream (so two calls roll the same).
func blow(b *Battle, att, v int) Fx {
	b.rng = NewRNG(5)
	b.computeMods()
	hp := b.Units[v].HP
	b.damage(b.Units[att], b.Units[v], b.Units[att].Pos, One)
	return hp - b.Units[v].HP
}

func near(a, c Fx, pct int) bool { return (a - c.MulDiv(pct, 100)).Abs() <= Ratio(2, 100) }

func TestDamageTakenAndBonus(t *testing.T) {
	shield := unitWith(t, Warrior, "shield", nil, Effect{Kind: DamageTaken, Pct: -30, Cond: CondRanged})
	mk := func(att Class, vSkills []Skill) *Battle {
		return testBattle(nil, tu{0, att, V(0, 0), nil}, tu{1, shield, V(0, I(-1)), vSkills})
	}
	if base, got := blow(mk(Archer, nil), 0, 1), blow(mk(Archer, []Skill{1}), 0, 1); !near(got, base, 70) {
		t.Errorf("shield wall vs an arrow: %v, want 70%% of %v", got, base)
	}
	if base, got := blow(mk(Warrior, nil), 0, 1), blow(mk(Warrior, []Skill{1}), 0, 1); got != base {
		t.Errorf("shield wall (ranged only) vs a sword: %v, want %v", got, base)
	}
	bodkin := unitWith(t, Archer, "bodkin", nil, Effect{Kind: BonusAgainst, Pct: 30, Tag: TagMelee})
	b0 := testBattle(nil, tu{0, bodkin, V(0, 0), nil}, tu{1, Warrior, V(0, I(-1)), nil})
	b1 := testBattle(nil, tu{0, bodkin, V(0, 0), []Skill{1}}, tu{1, Warrior, V(0, I(-1)), nil})
	if base, got := blow(b0, 0, 1), blow(b1, 0, 1); !near(got, base, 130) {
		t.Errorf("bodkin points vs melee: %v, want 130%% of %v", got, base)
	}
}

func TestAurasFuryPhalanx(t *testing.T) {
	song := unitWith(t, Mage, "singer", []Effect{{Kind: Aura, Mod: ModDamage, Pct: 10, Radius: I(6)}},
		Effect{Kind: Aura, Mod: ModAttackSpeed, Pct: -15, Radius: I(5), Foes: true})
	b := testBattle(nil, tu{0, song, V(0, 0), []Skill{1}}, tu{0, Warrior, V(I(3), 0), nil}, tu{0, Warrior, V(I(9), 0), nil}, tu{1, Warrior, V(0, I(-4)), nil})
	b.computeMods()
	if b.Units[1].mods[ModDamage] != 10 || b.Units[2].mods[ModDamage] != 0 || b.Units[0].mods[ModDamage] != 0 {
		t.Errorf("song trait: ally near %d, far %d, singer %d", b.Units[1].mods[ModDamage], b.Units[2].mods[ModDamage], b.Units[0].mods[ModDamage])
	}
	if b.Units[3].mods[ModAttackSpeed] != -15 {
		t.Errorf("mocking song on the enemy: %d", b.Units[3].mods[ModAttackSpeed])
	}
	fury := unitWith(t, Warrior, "fury", nil, Effect{Kind: Fury, Mod: ModDamage, Pct: 30, Threshold: 50})
	phal := unitWith(t, Warrior, "phalanx", nil, Effect{Kind: Phalanx, Mod: ModTaken, Pct: -8, Radius: I(2), Threshold: 3})
	b = testBattle(nil, tu{0, fury, V(0, 0), []Skill{1}}, tu{1, phal, V(0, I(-10)), []Skill{1}}, tu{1, phal, V(I(1), I(-10)), nil}, tu{1, phal, V(I(-1), I(-10)), nil})
	b.computeMods()
	if b.Units[0].mods[ModDamage] != 0 || b.Units[1].mods[ModTaken] != -16 {
		t.Errorf("full-health fury %d; phalanx with two neighbours %d", b.Units[0].mods[ModDamage], b.Units[1].mods[ModTaken])
	}
	b.Units[0].HP = b.Units[0].MaxHP / 3
	b.computeMods()
	if b.Units[0].mods[ModDamage] != 30 {
		t.Errorf("wounded fury: %d", b.Units[0].mods[ModDamage])
	}
}

func TestRallyGuardCurse(t *testing.T) {
	lament := unitWith(t, Mage, "lament", nil, Effect{Kind: Rally, Mod: ModDamage, Pct: 30, Radius: I(6), Ticks: 100})
	b := testBattle(nil, tu{0, lament, V(0, 0), []Skill{1}}, tu{0, Warrior, V(I(2), 0), nil}, tu{0, Warrior, V(I(-2), 0), nil}, tu{1, Warrior, V(0, I(-1)), nil})
	b.hurt(b.Units[3], b.Units[2], b.Units[2].HP, 0) // an ally dies
	b.computeMods()
	if b.Units[1].mods[ModDamage] != 30 {
		t.Errorf("rally after an ally's death: %d", b.Units[1].mods[ModDamage])
	}
	guard := unitWith(t, Warrior, "guard", nil, Effect{Kind: Guard, Pct: 30, Radius: Ratio(25, 10), Tag: TagRanged})
	b = testBattle(nil, tu{0, guard, V(I(1), 0), []Skill{1}}, tu{0, Archer, V(0, 0), nil}, tu{1, Warrior, V(0, I(-1)), nil})
	b0 := testBattle(nil, tu{0, guard, V(I(1), 0), nil}, tu{0, Archer, V(0, 0), nil}, tu{1, Warrior, V(0, I(-1)), nil})
	full, kept := blow(b0, 2, 1), blow(b, 2, 1)
	if !near(kept, full, 70) || b.Units[0].HP >= b.Units[0].MaxHP {
		t.Errorf("guard: the archer kept %v of %v, the guard took nothing", kept, full)
	}
	curse := unitWith(t, Warrior, "curse", nil, Effect{Kind: Curse, Pct: 20, Ticks: 100})
	b = testBattle(nil, tu{0, curse, V(0, 0), []Skill{1}}, tu{1, Warrior, V(0, I(-1)), nil})
	b.landed(b.Units[0], b.Units[1], V(0, 0))
	b.computeMods()
	if b.Units[1].mods[ModTaken] != 20 {
		t.Errorf("curse: %d", b.Units[1].mods[ModTaken])
	}
}

func TestBraceFearBurst(t *testing.T) {
	brace := unitWith(t, Warrior, "brace", nil, Effect{Kind: Brace, Pct: 150})
	b := testBattle(nil, tu{0, Warrior, V(0, Ratio(65, 10)), []Skill{WarriorDash}}, tu{1, brace, V(0, 0), []Skill{1}})
	b.steps(3)
	braced := false
	for _, e := range b.TakeEvents() {
		braced = braced || e.Kind == EvBrace
	}
	if !braced || b.Units[0].HP >= b.Units[0].MaxHP {
		t.Errorf("a dash into a braced unit: braced %v, charger at full health %v", braced, b.Units[0].HP >= b.Units[0].MaxHP)
	}
	dread := unitWith(t, Mage, "dread", nil, Effect{Kind: Fear, Radius: I(4), Ticks: 240, Duration: 30})
	b = testBattle(nil, tu{0, dread, V(0, 0), []Skill{1}}, tu{1, Warrior, V(0, I(-3)), nil})
	b.steps(10)
	if b.Units[1].fear <= 0 || b.Units[1].Pos.Z > I(-3) {
		t.Errorf("dread: fear %d, the warrior at z %v (should flee)", b.Units[1].fear, b.Units[1].Pos.Z)
	}
	spear := unitWith(t, Archer, "spear", nil, Effect{Kind: Burst, Value: I(8), Ticks: 160, Factor: I(2), Speed: I(15)})
	b = testBattle(nil, tu{0, spear, V(0, 0), []Skill{1}}, tu{1, Warrior, V(0, I(-6)), nil})
	b.steps(3)
	threw := false
	for _, e := range b.TakeEvents() {
		threw = threw || e.Kind == EvBurst
	}
	if !threw || len(b.Shots) == 0 || !b.Shots[0].burst || b.Shots[0].scale != I(2) {
		t.Errorf("throwing spear: thrown %v, shots %d", threw, len(b.Shots))
	}
}

func TestConeBurningFocusKite(t *testing.T) {
	jet := unitWith(t, Warrior, "jet", []Effect{{Kind: Cone, Factor: Half}}, Effect{Kind: BurningGround, Pct: 30, Radius: I(1), Duration: 60})
	b := testBattle(nil, tu{0, jet, V(0, 0), []Skill{1}}, tu{1, Warrior, V(0, I(-1)), nil}, tu{1, Warrior, V(Ratio(5, 10), I(-1)), nil}, tu{1, Warrior, V(0, I(1)), nil})
	b.Units[0].Face = V(0, -One)
	b.computeMods()
	b.strike(b.Units[0], b.Units[1])
	if b.Units[1].HP == b.Units[1].MaxHP || b.Units[2].HP == b.Units[2].MaxHP || b.Units[3].HP != b.Units[3].MaxHP {
		t.Errorf("cone: front %v %v, behind %v", b.Units[1].HP, b.Units[2].HP, b.Units[3].HP)
	}
	if len(b.zones) == 0 {
		t.Fatal("burning ground: no fire left")
	}
	hp := b.Units[1].HP
	for i := 0; i < TicksPerSecond+1; i++ {
		b.updateZones()
	}
	if b.Units[1].HP >= hp {
		t.Error("burning ground: no damage over a second")
	}
	volley := unitWith(t, Archer, "volley", nil, Effect{Kind: Focus, Pct: 15, Threshold: 2})
	mk := func(skills []Skill) *Battle {
		b := testBattle(nil, tu{0, volley, V(0, 0), skills}, tu{0, Archer, V(I(1), 0), nil}, tu{0, Archer, V(I(2), 0), nil}, tu{1, Warrior, V(0, I(-6)), nil})
		b.Units[1].Target, b.Units[2].Target = 3, 3
		return b
	}
	if base, got := blow(mk(nil), 0, 3), blow(mk([]Skill{1}), 0, 3); !near(got, base, 115) {
		t.Errorf("volley with two allies on the target: %v, want 115%% of %v", got, base)
	}
	steppe := unitWith(t, Archer, "steppe", nil, Effect{Kind: KiteShoot})
	b = testBattle(nil, tu{0, steppe, V(0, 0), []Skill{1}}, tu{1, Warrior, V(0, I(-2)), nil})
	b.steps(6)
	if len(b.Shots) == 0 && b.Units[1].HP == b.Units[1].MaxHP {
		t.Error("steppe shooting: no shot while backing away")
	}
}
