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
	partsDir  = "../web/public/assets/parts/"

	clipsMaxRaw  = 1_500_000
	clipsMaxGzip = 400_000
	partsMax     = 10_000_000 // the whole parts library
	armyMax      = 3_000_000  // the parts and textures of every class's first recipe

	buildMaxGirth  = 0.04 // metres pushed out per bone
	buildMaxHeight = 0.08 // ±, overall height
	buildMaxScale  = 0.2  // ±, shoulders and head
)

type skillLook struct {
	Action string  `json:"action"`
	Rate   float64 `json:"rate"`
	Attack string  `json:"attack"`
	Pitch  float64 `json:"pitch"`
	Fx     string  `json:"fx"`
}

type recipe struct {
	Name, Sex, Build, Head string
	Parts, Hair, Items     []string
	Variant                int
	Team                   float64
	Tint                   string
}

type build struct {
	Girth                   map[string]float64
	Shoulders, Head, Height float64
}

type partsManifest struct {
	Parts map[string]struct {
		Slot, Sex string
		Materials []string
	}
	Items     map[string]struct{ Bone string }
	Materials map[string]struct {
		Variants []string
		Mask     string
	}
}

type unitLook struct {
	UI struct {
		Label, Icon, Blurb string
	} `json:"ui"`
	Recipes []recipe `json:"recipes"`
	Anims   struct {
		Idle, Walk, Move, Sprint, Cheer string
		Attack, Hit, Death              []string
	} `json:"anims"`
	Skills map[string]skillLook `json:"skills"`
}

type looksData struct {
	Clips map[string]struct {
		Loop  bool     `json:"loop"`
		Hit   float64  `json:"hit"`
		Speed float64  `json:"speed"` // move clips: m/s at 1×
		From  []string `json:"from"`
	} `json:"clips"`
	Builds map[string]build    `json:"builds"`
	Units  map[string]unitLook `json:"units"`
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
		if len(look.Recipes) == 0 {
			bad("%s: no recipes", d.Name)
		}
		if look.UI.Label == "" || look.UI.Icon == "" || look.UI.Blurb == "" {
			bad("%s: the look needs ui label, icon and blurb", d.Name)
		}
		a := look.Anims
		clip(d.Name+" idle", a.Idle, false)
		clip(d.Name+" move", a.Move, false)
		clip(d.Name+" walk", a.Walk, false)
		if a.Sprint != "" {
			clip(d.Name+" sprint", a.Sprint, false)
		}
		for _, m := range []string{a.Walk, a.Move, a.Sprint} {
			if c, ok := l.Clips[m]; m != "" && ok && c.Speed <= 0 {
				bad("%s: move clip %q has no speed", d.Name, m)
			}
		}
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

// checkRecipes returns every way a look's recipes disagree with the parts
// manifest or the build limits.
func checkRecipes(l looksData, m partsManifest) []string {
	var errs []string
	bad := func(format string, args ...any) { errs = append(errs, fmt.Sprintf(format, args...)) }
	for name, b := range l.Builds {
		for bone, g := range b.Girth {
			if g < -buildMaxGirth || g > buildMaxGirth {
				bad("build %s: girth %v on %s outside ±%v m", name, g, bone, buildMaxGirth)
			}
		}
		if b.Height < 1-buildMaxHeight || b.Height > 1+buildMaxHeight {
			bad("build %s: height %v outside ±%v", name, b.Height, buildMaxHeight)
		}
		for what, v := range map[string]float64{"shoulders": b.Shoulders, "head": b.Head} {
			if v < 1-buildMaxScale || v > 1+buildMaxScale {
				bad("build %s: %s %v outside ±%v", name, what, v, buildMaxScale)
			}
		}
	}
	for unit, look := range l.Units {
		names := map[string]bool{}
		for _, r := range look.Recipes {
			where := unit + " recipe " + r.Name
			if names[r.Name] {
				bad("%s: name used twice", where)
			}
			names[r.Name] = true
			if _, ok := l.Builds[r.Build]; !ok {
				bad("%s: unknown build %q", where, r.Build)
			}
			if len(r.Parts) == 0 {
				bad("%s: no parts", where)
			}
			slots := map[string]string{}
			for i, p := range append(append(append([]string{}, r.Parts...), r.Head), r.Hair...) {
				part, ok := m.Parts[p]
				if !ok {
					bad("%s: part %q is not in parts.json", where, p)
					continue
				}
				if part.Sex != r.Sex {
					bad("%s: part %s is %s, the recipe is %s (they have different skeletons)", where, p, part.Sex, r.Sex)
				}
				switch {
				case i == 0 && part.Slot != "torso":
					bad("%s: the first part must be a torso (its skeleton is the unit's), got %s (%s)", where, p, part.Slot)
				case p == r.Head && part.Slot != "head":
					bad("%s: head %s is a %s part", where, p, part.Slot)
				case slices.Contains(r.Hair, p) && part.Slot != "hair":
					bad("%s: hair %s is a %s part", where, p, part.Slot)
				case part.Slot != "hair" && part.Slot != "accessory" && slots[part.Slot] != "":
					bad("%s: %s and %s both fill slot %s", where, slots[part.Slot], p, part.Slot)
				}
				slots[part.Slot] = p
				for _, mat := range part.Materials {
					if spec, ok := m.Materials[mat]; ok && r.Variant >= len(spec.Variants) {
						bad("%s: colour variant %d, %s has %d", where, r.Variant, mat, len(spec.Variants))
					}
				}
			}
			for _, it := range r.Items {
				if _, ok := m.Items[it]; !ok {
					bad("%s: item %q is not in parts.json", where, it)
				}
			}
			if r.Team < 0 || r.Team > 1 {
				bad("%s: team strength %v outside 0..1", where, r.Team)
			}
		}
	}
	sort.Strings(errs)
	return errs
}

// size of a part (.gltf, .bin and its textures), counting textures once per set
func partFiles(t *testing.T, name string, seen map[string]bool) int64 {
	var g struct {
		Buffers []struct{ URI string }
		Images  []struct{ URI string }
	}
	readJSON(t, partsDir+name+".gltf", &g)
	files := []string{name + ".gltf"}
	for _, b := range g.Buffers {
		files = append(files, b.URI)
	}
	for _, i := range g.Images {
		files = append(files, i.URI)
	}
	var n int64
	for _, f := range files {
		if seen[f] {
			continue
		}
		seen[f] = true
		st, err := os.Stat(partsDir + f)
		if err != nil {
			t.Errorf("part %s: %v", name, err)
			continue
		}
		n += st.Size()
	}
	return n
}

func TestRecipes(t *testing.T) {
	var l looksData
	readJSON(t, looksFile, &l)
	var m partsManifest
	readJSON(t, partsDir+"parts.json", &m)
	for _, e := range checkRecipes(l, m) {
		t.Error(e)
	}
	// budgets
	entries, err := os.ReadDir(partsDir)
	if err != nil {
		t.Fatal(err)
	}
	var total int64
	for _, e := range entries {
		info, _ := e.Info()
		total += info.Size()
	}
	if total > partsMax {
		t.Errorf("parts library is %d bytes, budget %d", total, partsMax)
	}
	seen := map[string]bool{}
	var army int64
	for _, d := range sim.Roster {
		r := l.Units[d.Name].Recipes[0]
		for _, p := range append(append(append(append([]string{}, r.Parts...), r.Head), r.Hair...), r.Items...) {
			army += partFiles(t, p, seen)
		}
		for _, p := range r.Parts {
			for _, mat := range m.Parts[p].Materials {
				if spec, ok := m.Materials[mat]; ok {
					for _, f := range []string{spec.Variants[r.Variant], spec.Mask} {
						if !seen[f] {
							seen[f] = true
							st, _ := os.Stat(partsDir + f)
							army += st.Size()
						}
					}
				}
			}
		}
	}
	if army > armyMax {
		t.Errorf("a default army's parts are %d bytes, budget %d", army, armyMax)
	}
	t.Logf("parts library %.1f MB, default army %.1f MB", float64(total)/1e6, float64(army)/1e6)
}

func TestRecipeCheckCatches(t *testing.T) {
	var l looksData
	readJSON(t, looksFile, &l)
	var m partsManifest
	readJSON(t, partsDir+"parts.json", &m)
	edit := func(f func(r *recipe)) looksData {
		var c looksData
		b, _ := json.Marshal(l)
		json.Unmarshal(b, &c)
		f(&c.Units["warrior"].Recipes[0])
		return c
	}
	cases := map[string]looksData{
		"missing part":  edit(func(r *recipe) { r.Parts = append(r.Parts, "Male_Knight_Helmet") }),
		"missing item":  edit(func(r *recipe) { r.Items = append(r.Items, "Halberd") }),
		"torso first":   edit(func(r *recipe) { r.Parts[0], r.Parts[1] = r.Parts[1], r.Parts[0] }),
		"mixed sex":     edit(func(r *recipe) { r.Head = "Head_Female" }),
		"slot twice":    edit(func(r *recipe) { r.Parts = append(r.Parts, r.Parts[1]) }),
		"unknown build": edit(func(r *recipe) { r.Build = "giant" }),
		"variant":       edit(func(r *recipe) { r.Variant = 5 }),
	}
	big := edit(func(*recipe) {})
	big.Builds["bulky"] = build{Girth: map[string]float64{"spine_03": 0.2}, Shoulders: 1.5, Head: 1, Height: 1.3}
	cases["build limits"] = big
	for name, c := range cases {
		if errs := checkRecipes(c, m); len(errs) == 0 {
			t.Errorf("%s: not reported", name)
		}
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
		"move clip without speed": func(c *looksData) []string {
			jog := c.Clips["Jog_Fwd_Loop"]
			jog.Speed = 0
			c.Clips["Jog_Fwd_Loop"] = jog
			return packed
		},
		"no ui text": func(c *looksData) []string {
			u := c.Units["archer"]
			u.UI.Blurb = ""
			c.Units["archer"] = u
			return packed
		},
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
