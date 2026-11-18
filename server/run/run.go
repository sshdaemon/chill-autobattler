// Package run owns runs: players' chains of matches with veterans, skill picks,
// Red scaling, scores and the scoreboard. The battle package calls into it to
// start a run's battle, check an army and apply a verdict.
package run

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"time"

	"connectrpc.com/connect/v2"
	"google.golang.org/protobuf/proto"

	chillv1 "chill/pb/chill/v1"
	"chill/server/auth"
	"chill/server/store"
	"chill/sim"
)

const (
	Ready   = "ready"
	Picking = "picking"
	Over    = "over"

	WinPoints      = 100
	SurvivorPoints = 10
)

type Service struct {
	DB  *store.DB
	Now func() time.Time
}

func NewService(db *store.DB) *Service { return &Service{DB: db, Now: time.Now} }

func (s *Service) nowMs() int64 { return s.Now().UnixMilli() }

func NewID() string {
	var b [16]byte
	rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// The prototype's run names: adjective noun number.
var (
	adjectives = []string{"Iron", "Swift", "Grim", "Bold", "Silent", "Red", "Grey", "Wild", "Lucky", "Stern", "Old", "Young"}
	nouns      = []string{"Wolf", "Falcon", "Bear", "Raven", "Boar", "Lynx", "Oak", "Hammer", "Shield", "Spear", "Fox", "Stag"}
)

func pick(n int) int {
	v, _ := rand.Int(rand.Reader, big.NewInt(int64(n)))
	return int(v.Int64())
}

func NewName() string {
	return fmt.Sprintf("%s %s %d", adjectives[pick(len(adjectives))], nouns[pick(len(nouns))], 100+pick(900))
}

// Player returns the authenticated player or Unauthenticated.
func Player(ctx context.Context) (string, error) {
	if id := auth.PlayerID(ctx); id != "" {
		return id, nil
	}
	return "", connect.Errorf(connect.CodeUnauthenticated, "missing or unknown player token")
}

// Veterans decodes a run's veterans.
func Veterans(r store.Run) ([]*chillv1.UnitSpec, error) {
	if len(r.Veterans) == 0 {
		return nil, nil
	}
	var a chillv1.Army
	if err := proto.Unmarshal(r.Veterans, &a); err != nil {
		return nil, fmt.Errorf("run %s veterans: %w", r.ID, err)
	}
	return a.Units, nil
}

func encodeVeterans(v []*chillv1.UnitSpec) []byte {
	if len(v) == 0 {
		return nil
	}
	b, _ := proto.MarshalOptions{Deterministic: true}.Marshal(&chillv1.Army{Units: v})
	return b
}

// RedHealthFor is Red's health % for a run's match (0 = full); the curve
// lives in sim.RunRedHealth.
func RedHealthFor(match int64) int64 { return int64(sim.RunRedHealth(int(match))) }

// SkillTotal is Red's skill count for a run's next match: the abilities the
// veterans learned above level 1 (every unit, Red's too, starts with one).
func SkillTotal(v []*chillv1.UnitSpec) int {
	n := 0
	for _, u := range v {
		n += max(0, len(u.Skills)-1)
	}
	return n
}

// Owed lists the veterans that owe a skill pick (below the top level).
func Owed(v []*chillv1.UnitSpec) []uint32 {
	var out []uint32
	for i, u := range v {
		if len(u.Skills) < sim.MaxLevel {
			out = append(out, uint32(i))
		}
	}
	return out
}

// Get loads a run owned by player (NotFound otherwise) and settles abandonment.
func (s *Service) Get(ctx context.Context, player, id string) (store.Run, error) {
	r, err := s.DB.GetRun(ctx, id)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && r.PlayerID != player) {
		return r, connect.Errorf(connect.CodeNotFound, "no run %q", id)
	}
	if err != nil {
		return r, err
	}
	return s.settle(ctx, r)
}

// settle ends the run as abandoned if its current battle missed a deadline.
func (s *Service) settle(ctx context.Context, r store.Run) (store.Run, error) {
	if r.State == Over || !r.CurrentBattleID.Valid {
		return r, nil
	}
	b, err := s.DB.GetBattle(ctx, r.CurrentBattleID.String)
	if err != nil {
		return r, err
	}
	now := s.nowMs()
	expired := (!b.LockedAt.Valid && b.DeployDeadline <= now) ||
		(b.LockedAt.Valid && !b.VerifiedAt.Valid && b.VerifyDeadline.Int64 <= now)
	if !expired {
		return r, nil
	}
	if err := s.DB.EndRun(ctx, store.EndRunParams{ID: r.ID, EndReason: sql.NullString{String: "abandoned", Valid: true}, Now: sql.NullInt64{Int64: now, Valid: true}}); err != nil {
		return r, err
	}
	return s.DB.GetRun(ctx, r.ID)
}

// Proto renders a run for the client, including what it needs to resume.
func (s *Service) Proto(ctx context.Context, r store.Run) (*chillv1.Run, error) {
	vets, err := Veterans(r)
	if err != nil {
		return nil, err
	}
	out := &chillv1.Run{
		Id: r.ID, Name: r.Name, Match: uint32(r.MatchNo), Wins: uint32(r.Wins), Score: uint32(r.Score),
		Veterans: vets, EndReason: r.EndReason.String, RedSkills: uint32(SkillTotal(vets)),
	}
	switch r.State {
	case Ready:
		out.State = chillv1.RunState_RUN_STATE_READY
	case Picking:
		out.State = chillv1.RunState_RUN_STATE_PICKING
		out.OwedPicks = Owed(vets)
	case Over:
		out.State = chillv1.RunState_RUN_STATE_OVER
	}
	if r.CurrentBattleID.Valid {
		b, err := s.DB.GetBattle(ctx, r.CurrentBattleID.String)
		if err != nil {
			return nil, err
		}
		cb := &chillv1.CurrentBattle{BattleId: b.ID, Seed: uint64(b.MapSeed), RedSkillCount: uint32(b.RedSkillCount),
			DeployDeadlineMs: b.DeployDeadline, Locked: b.LockedAt.Valid, RedHealthPct: uint32(b.RedHealthPct)}
		if b.LockedAt.Valid {
			cb.CombatSeed, cb.VerifyDeadlineMs = uint64(b.CombatSeed), b.VerifyDeadline.Int64
			cb.Blue = &chillv1.Army{}
			if err := proto.Unmarshal(b.BlueArmy, cb.Blue); err != nil {
				return nil, err
			}
		}
		out.Battle = cb
	}
	return out, nil
}

func (s *Service) CreatePlayer(ctx context.Context, _ *chillv1.CreatePlayerRequest) (*chillv1.CreatePlayerResponse, error) {
	token, hash := auth.NewToken()
	if err := s.DB.CreatePlayer(ctx, store.CreatePlayerParams{ID: NewID(), TokenHash: hash, CreatedAt: s.nowMs()}); err != nil {
		return nil, err
	}
	return &chillv1.CreatePlayerResponse{PlayerToken: token}, nil
}

func (s *Service) StartRun(ctx context.Context, _ *chillv1.StartRunRequest) (*chillv1.StartRunResponse, error) {
	player, err := Player(ctx)
	if err != nil {
		return nil, err
	}
	id := NewID()
	now := s.nowMs()
	err = s.DB.Tx(ctx, func(q *store.Queries) error {
		if err := q.EndActiveRuns(ctx, store.EndActiveRunsParams{PlayerID: player, EndReason: sql.NullString{String: "restarted", Valid: true}, Now: sql.NullInt64{Int64: now, Valid: true}}); err != nil {
			return err
		}
		return q.CreateRun(ctx, store.CreateRunParams{ID: id, PlayerID: player, Name: NewName(), CreatedAt: now})
	})
	if err != nil {
		return nil, err
	}
	r, err := s.DB.GetRun(ctx, id)
	if err != nil {
		return nil, err
	}
	pr, err := s.Proto(ctx, r)
	return &chillv1.StartRunResponse{Run: pr}, err
}

func (s *Service) GetRun(ctx context.Context, _ *chillv1.GetRunRequest) (*chillv1.GetRunResponse, error) {
	player, err := Player(ctx)
	if err != nil {
		return nil, err
	}
	r, err := s.DB.LatestRun(ctx, player)
	if errors.Is(err, sql.ErrNoRows) {
		return &chillv1.GetRunResponse{}, nil
	}
	if err != nil {
		return nil, err
	}
	if r, err = s.settle(ctx, r); err != nil {
		return nil, err
	}
	pr, err := s.Proto(ctx, r)
	return &chillv1.GetRunResponse{Run: pr}, err
}

func (s *Service) PickSkills(ctx context.Context, req *chillv1.PickSkillsRequest) (*chillv1.PickSkillsResponse, error) {
	player, err := Player(ctx)
	if err != nil {
		return nil, err
	}
	r, err := s.Get(ctx, player, req.RunId)
	if err != nil {
		return nil, err
	}
	if r.State != Picking {
		return nil, connect.Errorf(connect.CodeFailedPrecondition, "run %s owes no skill picks", r.ID)
	}
	vets, err := Veterans(r)
	if err != nil {
		return nil, err
	}
	owed := Owed(vets)
	if len(req.Picks) != len(owed) {
		return nil, connect.Errorf(connect.CodeInvalidArgument, "%d picks owed, got %d", len(owed), len(req.Picks))
	}
	seen := map[uint32]bool{}
	for _, p := range req.Picks {
		if !slices.Contains(owed, p.Veteran) || seen[p.Veteran] {
			return nil, connect.Errorf(connect.CodeInvalidArgument, "veteran %d owes no pick here", p.Veteran)
		}
		seen[p.Veteran] = true
		v := vets[p.Veteran]
		if _, ok := sim.SkillDefOf(sim.Class(v.UnitClass-1), sim.Skill(p.Skill)); !ok || slices.Contains(v.Skills, p.Skill) {
			return nil, connect.Errorf(connect.CodeInvalidArgument, "veteran %d can't learn skill %v", p.Veteran, p.Skill)
		}
	}
	for _, p := range req.Picks {
		vets[p.Veteran].Skills = append(vets[p.Veteran].Skills, p.Skill)
	}
	n, err := s.DB.FinishPicks(ctx, store.FinishPicksParams{ID: r.ID, Veterans: encodeVeterans(vets)})
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, connect.Errorf(connect.CodeFailedPrecondition, "run %s owes no skill picks", r.ID)
	}
	if r, err = s.DB.GetRun(ctx, r.ID); err != nil {
		return nil, err
	}
	pr, err := s.Proto(ctx, r)
	return &chillv1.PickSkillsResponse{Run: pr}, err
}

func (s *Service) GetScoreboard(ctx context.Context, _ *chillv1.GetScoreboardRequest) (*chillv1.GetScoreboardResponse, error) {
	player, err := Player(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.DB.SweepAbandoned(ctx, sql.NullInt64{Int64: s.nowMs(), Valid: true}); err != nil {
		return nil, err
	}
	rows, err := s.DB.TopRuns(ctx, player)
	if err != nil {
		return nil, err
	}
	out := &chillv1.GetScoreboardResponse{}
	for i, r := range rows {
		out.Top = append(out.Top, &chillv1.ScoreRow{Rank: uint32(i + 1), Name: r.Name, Score: uint32(r.Score), Wins: uint32(r.Wins), Mine: r.Mine})
	}
	best, err := s.DB.BestFinishedScore(ctx, player)
	if err != nil {
		return nil, err
	}
	if best >= 0 {
		above, err := s.DB.CountFinishedAbove(ctx, best)
		if err != nil {
			return nil, err
		}
		out.MyBestRank, out.MyBestScore = uint32(above+1), uint32(best)
	}
	return out, nil
}

// ---------- used by the battle service ----------

// CheckArmy enforces that Blue's army is the run's veterans (in order, exactly)
// followed by level-1 recruits (one ability each), 5 units in all.
func CheckArmy(vets []*chillv1.UnitSpec, blue *chillv1.Army) error {
	units := blue.GetUnits()
	if len(units) != sim.ArmySize {
		return fmt.Errorf("a run's army has %d units, got %d", sim.ArmySize, len(units))
	}
	for i, u := range units {
		if i < len(vets) {
			v := vets[i]
			if u.UnitClass != v.UnitClass || !sameSkills(u.Skills, v.Skills) {
				return fmt.Errorf("unit %d must be veteran %d (%v with %v)", i, i, v.UnitClass, v.Skills)
			}
		} else if len(u.Skills) != 1 {
			return fmt.Errorf("unit %d is a recruit: it starts at level 1 with one ability, got %d", i, len(u.Skills))
		}
	}
	return nil
}

func sameSkills(a, b []uint32) bool {
	x, y := slices.Clone(a), slices.Clone(b)
	slices.Sort(x)
	slices.Sort(y)
	return slices.Equal(x, y)
}

// ApplyResult updates the run after a verdict, inside the caller's transaction.
func ApplyResult(ctx context.Context, q *store.Queries, r store.Run, fair bool, blue *chillv1.Army, o sim.Outcome, now int64) error {
	if fair && o.Result == sim.BlueWin {
		var vets []*chillv1.UnitSpec
		for _, sv := range o.Survivors {
			if sv.Index < len(blue.Units) { // Blue's units come first
				u := blue.Units[sv.Index]
				vets = append(vets, &chillv1.UnitSpec{UnitClass: u.UnitClass, Skills: slices.Clone(u.Skills)})
			}
		}
		state := Ready
		if len(Owed(vets)) > 0 {
			state = Picking
		}
		return q.WinMatch(ctx, store.WinMatchParams{ID: r.ID, State: state, Points: int64(WinPoints + SurvivorPoints*len(vets)), Veterans: encodeVeterans(vets)})
	}
	reason := "defeat"
	switch {
	case !fair:
		reason = "not fair"
	case o.Result == sim.Draw:
		reason = "draw"
	}
	return q.EndRun(ctx, store.EndRunParams{ID: r.ID, EndReason: sql.NullString{String: reason, Valid: true}, Now: sql.NullInt64{Int64: now, Valid: true}})
}
