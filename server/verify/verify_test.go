package verify

import (
	"context"
	"strings"
	"testing"

	"connectrpc.com/connect/v2"
	"google.golang.org/protobuf/proto"

	chillv1 "chill/pb/chill/v1"
	"chill/pb/convert"
	"chill/sim"
)

func input() *chillv1.BattleInput {
	w, a, m := chillv1.UnitClass_UNIT_CLASS_WARRIOR, chillv1.UnitClass_UNIT_CLASS_ARCHER, chillv1.UnitClass_UNIT_CLASS_MAGE
	return &chillv1.BattleInput{SimVersion: sim.Version, Seed: 42, RedSkillCount: 1, Blue: &chillv1.Army{
		Formation: chillv1.Formation_FORMATION_3_2,
		Units:     []*chillv1.UnitSpec{{UnitClass: w}, {UnitClass: w}, {UnitClass: a}, {UnitClass: a}, {UnitClass: m}},
	}}
}

// honest client: plays the battle itself, as the WASM build does
func play(t *testing.T, in *chillv1.BattleInput) *chillv1.BattleOutcome {
	si, err := convert.Input(in)
	if err != nil {
		t.Fatal(err)
	}
	o, err := sim.Run(si)
	if err != nil {
		t.Fatal(err)
	}
	return convert.Outcome(o)
}

func TestHonestClaim(t *testing.T) {
	res, err := Server{}.VerifyBattle(context.Background(), &chillv1.VerifyBattleRequest{Input: input(), Claimed: play(t, input())})
	if err != nil || !res.Fair {
		t.Fatalf("honest claim: %v %v", err, res)
	}
	if res.Outcome.GetChecksum() == 0 {
		t.Error("server should return its outcome")
	}
}

func TestTamperedClaim(t *testing.T) {
	for name, tamper := range map[string]func(*chillv1.BattleOutcome){
		"winner":   func(o *chillv1.BattleOutcome) { o.Result = 3 - o.Result%3 },
		"checksum": func(o *chillv1.BattleOutcome) { o.Checksum++ },
		"ticks":    func(o *chillv1.BattleOutcome) { o.Ticks-- },
	} {
		claim := play(t, input())
		honest := proto.Clone(claim).(*chillv1.BattleOutcome)
		tamper(claim)
		res, err := Server{}.VerifyBattle(context.Background(), &chillv1.VerifyBattleRequest{Input: input(), Claimed: claim})
		if err != nil || res.Fair || res.Reason == "" {
			t.Errorf("%s: tampered claim passed: %v %v", name, err, res)
		}
		if !proto.Equal(res.Outcome, honest) {
			t.Errorf("%s: server should return its own outcome", name)
		}
	}
}

func TestTamperedInput(t *testing.T) {
	cases := map[string]func(*chillv1.BattleInput){
		"six units":   func(in *chillv1.BattleInput) { in.Blue.Units = append(in.Blue.Units, in.Blue.Units[0]) },
		"skill twice": func(in *chillv1.BattleInput) { in.Blue.Units[0].Skills = []chillv1.Skill{4, 4} },
		"outside zone": func(in *chillv1.BattleInput) {
			in.Blue.Formation = chillv1.Formation_FORMATION_CUSTOM
			for _, u := range in.Blue.Units {
				u.Position = &chillv1.Point{ZRaw: int64(sim.I(-10))} // in Red's zone
			}
		},
	}
	for name, mod := range cases {
		in := input()
		mod(in)
		_, err := Server{}.VerifyBattle(context.Background(), &chillv1.VerifyBattleRequest{Input: in, Claimed: &chillv1.BattleOutcome{}})
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: want InvalidArgument, got %v", name, err)
		}
	}
}

func TestUnknownVersion(t *testing.T) {
	in := input()
	in.SimVersion = "0.0.1-old"
	res, err := Server{}.VerifyBattle(context.Background(), &chillv1.VerifyBattleRequest{Input: in, Claimed: &chillv1.BattleOutcome{}})
	if err != nil || res.Fair || !strings.Contains(res.Reason, "0.0.1-old") || !strings.Contains(res.Reason, sim.Version) || res.Outcome != nil {
		t.Errorf("unknown version: %v %v", err, res)
	}
}
