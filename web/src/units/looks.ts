// Unit looks: how each unit-roster entry looks and moves (presentation only).
// looks.json is matched to the roster by name; `wasm/looks_test.go` checks the
// two stay in sync. A missing look falls back to the warrior's (melee) or the
// archer's (ranged) and warns once in dev.
import raw from './looks.json';
import type { ClassInfo, SkillInfo } from '../gamedata.ts';

export type ClipInfo = { loop?: boolean; hit?: number; from?: string[] };
// action: played when the skill acts (gap-closer, escape); attack: replaces the
// unit's attacks while it has the skill; pitch: upper body tilts back while
// attacking (radians); rate: action playback speed; fx: effect preset name.
export type SkillLook = { action?: string; rate?: number; attack?: string; pitch?: number; fx?: string };
export type AnimSet = { idle: string; move: string; cheer: string; attack: string[]; hit: string[]; death: string[] };
export type UnitLook = { models: string[]; anims: AnimSet; skills: Record<string, SkillLook> };

export const CLIPS = raw.clips as Record<string, ClipInfo>;
const LOOKS = raw.units as Record<string, UnitLook>;

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

export const modelFor = (cls: number, nth: number) => { const m = lookOf(cls).models; return m[nth % m.length]; };

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
export function skillWith(cls: number, learned: number[], kind: string) {
  return skillLooks(cls, learned).find((s) => s.info.effects.some((e) => e.kind === kind));
}
