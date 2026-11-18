// Unit looks: how each unit-roster entry looks and moves (presentation only).
// One file per unit, looks/<unit>.json, matched to the roster by name, plus
// looks/_shared.json (body builds and the clips several units use); each unit
// file lists the clips only it uses. `wasm/looks_test.go` checks they stay in
// sync with the roster. A missing look falls back to the warrior's (melee) or
// the archer's (ranged) and warns once in dev.
import type { ClassInfo, SkillInfo } from '../gamedata.ts';

// speed: ground speed (m/s) a move clip covers at 1× (from the packs' root motion)
export type ClipInfo = { loop?: boolean; hit?: number; speed?: number; from?: string[] };
// action: played when the skill acts (gap-closer, escape); attack: replaces the
// unit's attacks while it has the skill; pitch: upper body tilts back while
// attacking (radians); rate: action playback speed; fx: effect preset name.
export type SkillLook = { action?: string; rate?: number; attack?: string; pitch?: number; fx?: string };
// walk and move (a jog) are blended by ground speed; sprint plays for dashes without a skill action
export type AnimSet = { idle: string; walk: string; move: string; sprint?: string; cheer: string; attack: string[]; hit: string[]; death: string[] };
// A character recipe (unit-looks): outfit parts (the first one's skeleton is the
// unit's), head, hair, the outfit colour variant, items on bones, team-colour
// strength on cloth, an optional cloth tint, and a body build.
export type Recipe = {
  name: string; sex: 'male' | 'female'; build: string; parts: string[]; head: string; hair: string[];
  variant: number; items: string[]; team: number; tint?: string;
};
// Body builds: girth in metres per bone (pushed out along the surface), shoulder
// width, head size and height factors.
export type Build = { girth: Record<string, number>; shoulders: number; head: number; height: number };
// What the UI shows for the unit type: name, a short glyph (emoji) and a one-line description.
// placeholder: the unit's look is borrowed until its unit session dresses it (roadmap Phase 4)
export type UnitUI = { label: string; icon: string; blurb: string; placeholder?: boolean };
// shot: how its bolts look: 'smoke' (dark smoke and the `.smoke` hit and shot
// presets, for curses) instead of the default glowing orb
export type UnitLook = { ui: UnitUI; recipes: Recipe[]; anims: AnimSet; skills: Record<string, SkillLook>; shot?: 'smoke' };

type LookFile = UnitLook & { clips?: Record<string, ClipInfo> };
type SharedFile = { clips: Record<string, ClipInfo>; builds: Record<string, Build> };
const files = import.meta.glob<LookFile | SharedFile>('./looks/*.json', { eager: true, import: 'default' });
const fileName = (path: string) => path.slice(path.lastIndexOf('/') + 1, -'.json'.length);
const shared = files['./looks/_shared.json'] as SharedFile;
const LOOKS: Record<string, UnitLook> = {};
export const CLIPS: Record<string, ClipInfo> = { ...shared.clips };
// clip bundles to load (web/public/assets/clips/<name>.json): shared, then one per unit
export const CLIP_BUNDLES = ['_shared'];
for (const [path, f] of Object.entries(files)) {
  const name = fileName(path);
  if (name === '_shared') continue;
  const { clips, ...look } = f as LookFile;
  LOOKS[name] = look;
  CLIP_BUNDLES.push(name);
  Object.assign(CLIPS, clips);
}
export const BUILDS = shared.builds;

let roster: ClassInfo[] = [];
const warned = new Set<string>();
// called once the sim's data() arrives
export function setRoster(classes: ClassInfo[]) { roster = classes; }

export function lookOf(cls: number): UnitLook {
  const info = roster[cls];
  const look = info && LOOKS[info.name];
  if (look) return look;
  const fallback = info?.ranged ? LOOKS.archer : LOOKS.warrior;
  if (import.meta.env.DEV && !warned.has(`u${cls}`)) {
    warned.add(`u${cls}`);
    console.warn(`no look for unit ${info?.name ?? cls}; using the ${info?.ranged ? 'archer' : 'warrior'}'s`);
  }
  return fallback;
}

const recipeFor = (cls: number, nth: number) => { const r = lookOf(cls).recipes; return r[nth % r.length]; };
// the recipe's name doubles as the unit's display name
export const uiOf = (cls: number): UnitUI => lookOf(cls).ui;
export const modelFor = (cls: number, nth: number) => recipeFor(cls, nth).name;
export const recipeNamed = (cls: number, name: string) => lookOf(cls).recipes.find((r) => r.name === name) ?? recipeFor(cls, 0);

// The unit's learned skills, each with its sim definition and its look.
export function skillLooks(cls: number, learned: number[]): { info: SkillInfo; look: SkillLook }[] {
  const info = roster[cls], look = lookOf(cls);
  if (!info) return [];
  // roster order, so "the first skill with an attack" is stable
  return info.skills.filter((s) => learned.includes(s.id)).map((s) => {
    const l = look.skills[String(s.id)];
    if (!l && import.meta.env.DEV && !warned.has(`s${cls}.${s.id}`)) {
      warned.add(`s${cls}.${s.id}`);
      console.warn(`no look for ${info.name}'s skill ${s.name}`);
    }
    return { info: s, look: l ?? {} };
  });
}

// The look of the unit's skill with an effect of this kind (e.g. 'gapCloser'), if it has one.
// Its bolts are drawn as smoke (the look's `shot`).
export const smokeShots = (cls: number) => { const info = roster[cls]; return !!info && LOOKS[info.name]?.shot === 'smoke'; };

export function skillWith(cls: number, learned: number[], kind: string) {
  return skillLooks(cls, learned).find((s) => s.info.effects.some((e) => e.kind === kind));
}
