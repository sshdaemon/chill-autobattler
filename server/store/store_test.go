package store

import (
	"context"
	"database/sql"
	"io/fs"
	"path/filepath"
	"sort"
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

func userVersion(t *testing.T, db *DB) int {
	var v int
	if err := db.conn.QueryRow("PRAGMA user_version").Scan(&v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestMigrations(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	// fresh database: every migration applied
	fresh, err := Open(ctx, filepath.Join(dir, "fresh.db"))
	if err != nil {
		t.Fatal(err)
	}
	if v := userVersion(t, fresh); v != 7 {
		t.Errorf("fresh db at version %d, want 7", v)
	}
	if err := fresh.CreatePlayer(ctx, CreatePlayerParams{ID: "p", TokenHash: []byte{1}, CreatedAt: 1}); err != nil {
		t.Fatalf("players table missing: %v", err)
	}
	fresh.Close()

	// reopening is a no-op
	again, err := Open(ctx, filepath.Join(dir, "fresh.db"))
	if err != nil {
		t.Fatalf("rerun: %v", err)
	}
	if id, err := again.PlayerByTokenHash(ctx, []byte{1}); err != nil || id != "p" {
		t.Errorf("data lost on reopen: %v %q", err, id)
	}
	again.Close()

	// a database from before runs (battles table only, user_version 0) upgrades in place
	oldPath := filepath.Join(dir, "old.db")
	raw, err := sql.Open("sqlite", "file:"+oldPath)
	if err != nil {
		t.Fatal(err)
	}
	first, _ := migrations.ReadFile("migrations/0001_battles.sql")
	if _, err := raw.Exec(string(first)); err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(`INSERT INTO battles (id, sim_version, map_seed, combat_seed, red_skill_count, created_at, deploy_deadline) VALUES ('old', 'v', 1, 2, 0, 1, 2)`); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	up, err := Open(ctx, oldPath)
	if err != nil {
		t.Fatalf("upgrade: %v", err)
	}
	defer up.Close()
	b, err := up.GetBattle(ctx, "old")
	if err != nil || b.MapSeed != 1 || b.RunID.Valid {
		t.Errorf("old battle after upgrade: %v %+v", err, b)
	}
	if v := userVersion(t, up); v != 7 {
		t.Errorf("upgraded db at version %d", v)
	}
}

// Retiring Classic (0007) ends its unfinished runs and leaves every other run alone.
func TestRetireClassicMigration(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "v6.db")
	raw, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	files, _ := fs.Glob(migrations, "migrations/*.sql")
	sort.Strings(files)
	for _, f := range files[:6] {
		q, _ := migrations.ReadFile(f)
		if _, err := raw.Exec(string(q)); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
	}
	if _, err := raw.Exec(`PRAGMA user_version = 6;
		INSERT INTO players (id, token_hash, created_at) VALUES ('p', x'01', 1);
		INSERT INTO runs (id, player_id, name, state, match_no, wins, score, current_battle_id, created_at, faction) VALUES
			('classic-active', 'p', 'a', 'picking', 4, 3, 120, 'b', 1, 1),
			('classic-over', 'p', 'b', 'over', 2, 1, 40, NULL, 1, 1),
			('rus-active', 'p', 'c', 'ready', 2, 1, 30, NULL, 1, 2);`); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	want := map[string]struct{ state, reason string }{
		"classic-active": {"over", "faction retired"},
		"classic-over":   {"over", ""},
		"rus-active":     {"ready", ""},
	}
	for id, w := range want {
		r, err := db.GetRun(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if r.State != w.state || r.EndReason.String != w.reason {
			t.Errorf("%s: state %q reason %q, want %q %q", id, r.State, r.EndReason.String, w.state, w.reason)
		}
		if id == "classic-active" && (r.CurrentBattleID.Valid || !r.EndedAt.Valid || r.Score != 120) {
			t.Errorf("ended classic run: %+v", r)
		}
	}
}
