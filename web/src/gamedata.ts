// Balance data as exported by the sim (wasm data()). Never hard-code these numbers.
// One effect of a skill (sim EffectKind), e.g. { kind: 'gapCloser', value: 10, speed: 16, secs: 10 }.
export type EffectInfo = { kind: string; stat?: string; pct?: number; value?: number; radius?: number; speed?: number; factor?: number; secs?: number };
export type SkillInfo = { id: number; name: string; pct: number; value: number; effects: EffectInfo[] };
// A unit-roster entry (classes[i] is sim class i).
export type ClassInfo = {
  name: string; hp: number; dmgLo: number; dmgHi: number; cooldown: number; range: number; speed: number; sight: number; swing: number;
  ranged: boolean; kite: boolean; target: string; engage: string; /* temperament */ projectile?: 'arrow' | 'bolt'; tags: string[]; bonus: { tag: string; mult: number }[];
  matchup: number[]; skills: SkillInfo[];
};
export type GameData = {
  version: string; classes: ClassInfo[]; formations: string[]; armySize: number; deployZone: [number, number, number, number];
  hitAt: number; // share of a swing at which the blow lands or the shot leaves
};
// Every unit of the current battle (wasm units()), Red's generated army too.
export type UnitLoadout = { cls: number; skills: number[] };

// Unit labels, icons and blurbs come from each unit's look (units/looks.ts uiOf).

// sim.Formation values (wasm/data().formations index); pb Formation = sim + 1
export const PRESETS = [1, 2, 3, 4, 5, 6, 7];
export const CUSTOM = 0;
