package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestRoundTripAndReopen(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "t.db")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	big := uint64(1<<63 + 12345) // seeds use all 64 bits
	if err := db.CreateBattle(ctx, CreateBattleParams{ID: "b1", SimVersion: "v", MapSeed: int64(big), CombatSeed: -1, RedSkillCount: 2, CreatedAt: 10, DeployDeadline: 100}); err != nil {
		t.Fatal(err)
	}
	n, err := db.LockBattle(ctx, LockBattleParams{ID: "b1", BlueArmy: []byte{1, 2}, Now: sql.NullInt64{Int64: 50, Valid: true}, VerifyDeadline: sql.NullInt64{Int64: 500, Valid: true}})
	if err != nil || n != 1 {
		t.Fatalf("lock: %v %d", err, n)
	}
	if n, _ := db.LockBattle(ctx, LockBattleParams{ID: "b1", BlueArmy: []byte{9}, Now: sql.NullInt64{Int64: 60, Valid: true}}); n != 0 {
		t.Error("second lock should change nothing")
	}
	db.Close()

	db, err = Open(ctx, path) // reopening keeps everything
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	b, err := db.GetBattle(ctx, "b1")
	if err != nil {
		t.Fatal(err)
	}
	if uint64(b.MapSeed) != big || uint64(b.CombatSeed) != 1<<64-1 || string(b.BlueArmy) != "\x01\x02" || b.LockedAt.Int64 != 50 {
		t.Errorf("round trip: %+v", b)
	}
	if n, _ := db.RecordVerdict(ctx, RecordVerdictParams{ID: "b1", Now: sql.NullInt64{Int64: 600, Valid: true}}); n != 0 {
		t.Error("verdict after the verify deadline should change nothing")
	}
	if n, _ := db.RecordVerdict(ctx, RecordVerdictParams{ID: "b1", Now: sql.NullInt64{Int64: 200, Valid: true}, Fair: sql.NullInt64{Int64: 1, Valid: true}}); n != 1 {
		t.Error("verdict before the deadline should be recorded")
	}
	if n, _ := db.RecordVerdict(ctx, RecordVerdictParams{ID: "b1", Now: sql.NullInt64{Int64: 210, Valid: true}}); n != 0 {
		t.Error("second verdict should change nothing")
	}
}
