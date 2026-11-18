package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"testing"

	"chill/sim"
)

// The unit looks (web/src/units/looks.json) are presentation data matched to
// the sim's unit roster by name; this keeps the two, the effect presets
// (web/src/scene/fx.json) and the packed clips (web/public/assets/clips.json)
// in sync. Runs under node with GOOS=js (pnpm test:wasm), which can read files.

const (
	looksFile = "../web/src/units/looks.json"
	fxFile    = "../web/src/scene/fx.json"
	clipsFile = "../web/public/assets/clips.json"

	clipsMaxRaw  = 1_500_000
	clipsMaxGzip = 400_000
)

type skillLook struct {
	Action string  `json:"action"`
	Rate   float64 `json:"rate"`
	Attack string  `json:"attack"`
	Pitch  float64 `json:"pitch"`
	Fx     string  `json:"fx"`
}

type unitLook struct {
	Models []string `json:"models"`
	Anims  struct {
		Idle, Move, Cheer  string
		Attack, Hit, Death []string
	} `json:"anims"`
	Skills map[string]skillLook `json:"skills"`
}

type looksData struct {
	Clips map[string]struct {
		Loop bool     `json:"loop"`
		Hit  float64  `json:"hit"`
		From []string `json:"from"`
	} `json:"clips"`
	Units map[string]unitLook `json:"units"`
}

func readJSON(t *testing.T, path string, v any) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return raw
}

// checkLooks returns every way the looks disagree with the roster, the
// presets or the packed clips (sorted, so failures read the same each run).
func checkLooks(roster []sim.UnitDef, l looksData, fx map[string]json.RawMessage, packed []string) []string {
	var errs []string
	bad := func(format string, args ...any) { errs = append(errs, fmt.Sprintf(format, args...)) }
	clip := func(where, name string, attack bool) {
		c, ok := l.Clips[name]
		switch {
		case name == "":
			bad("%s: no clip", where)
		case !ok:
			bad("%s: clip %q is not in looks.json clips", where, name)
		case !slices.Contains(packed, name):
			bad("%s: clip %q is not packed in clips.json (run tools/assets/ual_clips.py)", where, name)
		case attack && c.Hit <= 0:
			bad("%s: attack clip %q has no hit time", where, name)
		}
	}
	for _, d := range roster {
		look, ok := l.Units[d.Name]
		if !ok {
			bad("unit %s has no look", d.Name)
			continue
		}
		if len(look.Models) == 0 {
			bad("%s: no models", d.Name)
		}
		a := look.Anims
		clip(d.Name+" idle", a.Idle, false)
		clip(d.Name+" move", a.Move, false)
		clip(d.Name+" cheer", a.Cheer, false)
		for set, list := range map[string][]string{"attack": a.Attack, "hit": a.Hit, "death": a.Death} {
			if len(list) == 0 {
				bad("%s: no %s animation", d.Name, set)
			}
			for _, c := range list {
				clip(d.Name+" "+set, c, set == "attack")
			}
		}
		var want []string
		for _, s := range d.Skills {
			want = append(want, strconv.Itoa(int(s.Skill)))
			sl, ok := look.Skills[strconv.Itoa(int(s.Skill))]
			if !ok {
				bad("%s: look misses skill %s (%d)", d.Name, s.Name, s.Skill)
				continue
			}
			if sl.Action != "" {
				clip(d.Name+" "+s.Name+" action", sl.Action, false)
			}
			if sl.Attack != "" {
				clip(d.Name+" "+s.Name+" attack", sl.Attack, true)
			}
			if _, ok := fx[sl.Fx]; sl.Fx != "" && !ok {
				bad("%s: skill %s names effect preset %q, which fx.json lacks", d.Name, s.Name, sl.Fx)
			}
		}
		for id := range look.Skills {
			if !slices.Contains(want, id) {
				bad("%s: look has skill %s, which the unit can't learn", d.Name, id)
			}
		}
	}
	for name, c := range l.Clips {
		if !slices.Contains(packed, name) {
			bad("clip %q is in looks.json but not packed", name)
		}
		if len(c.From) == 1 {
			bad("clip %q: from needs two or more parts", name)
		}
	}
	sort.Strings(errs)
	return errs
}

func loadLooks(t *testing.T) (looksData, map[string]json.RawMessage, []string, []byte) {
	var l looksData
	readJSON(t, looksFile, &l)
	var fx map[string]json.RawMessage
	readJSON(t, fxFile, &fx)
	var clips []struct{ Name string }
	raw := readJSON(t, clipsFile, &clips)
	var names []string
	for _, c := range clips {
		names = append(names, c.Name)
	}
	return l, fx, names, raw
}

func TestLooksMatchRoster(t *testing.T) {
	l, fx, packed, raw := loadLooks(t)
	for _, e := range checkLooks(sim.Roster, l, fx, packed) {
		t.Error(e)
	}
	if len(packed) != len(l.Clips) {
		t.Errorf("clips.json has %d clips, looks.json names %d: re-run tools/assets/ual_clips.py", len(packed), len(l.Clips))
	}
	var z bytes.Buffer
	w := gzip.NewWriter(&z)
	w.Write(raw)
	w.Close()
	if len(raw) > clipsMaxRaw || z.Len() > clipsMaxGzip {
		t.Errorf("clips.json is %d bytes (%d gzipped), budget %d (%d)", len(raw), z.Len(), clipsMaxRaw, clipsMaxGzip)
	}
}

// One broken look per rule; each must be reported.
func TestLooksCheckCatches(t *testing.T) {
	l, fx, packed, _ := loadLooks(t)
	clone := func() looksData {
		var c looksData
		b, _ := json.Marshal(l)
		json.Unmarshal(b, &c)
		return c
	}
	cases := map[string]func(c *looksData) []string{
		"no look":       func(c *looksData) []string { delete(c.Units, "mage"); return packed },
		"missing skill": func(c *looksData) []string { delete(c.Units["warrior"].Skills, "4"); return packed },
		"foreign skill": func(c *looksData) []string { c.Units["warrior"].Skills["9"] = skillLook{}; return packed },
		"unknown clip": func(c *looksData) []string {
			u := c.Units["archer"]
			u.Anims.Idle = "Nope"
			c.Units["archer"] = u
			return packed
		},
		"unknown preset": func(c *looksData) []string { c.Units["mage"].Skills["8"] = skillLook{Fx: "nope"}; return packed },
		"attack without hit": func(c *looksData) []string {
			u := c.Units["warrior"]
			u.Anims.Attack = []string{"Jog_Fwd_Loop"}
			c.Units["warrior"] = u
			return packed
		},
		"not packed": func(c *looksData) []string {
			return slices.DeleteFunc(slices.Clone(packed), func(n string) bool { return n == "Yes" })
		},
	}
	for name, breakIt := range cases {
		c := clone()
		p := breakIt(&c)
		if errs := checkLooks(sim.Roster, c, fx, p); len(errs) == 0 {
			t.Errorf("%s: not reported", name)
		}
	}
}
