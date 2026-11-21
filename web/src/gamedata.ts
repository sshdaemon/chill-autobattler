// Balance data as exported by the sim (wasm data()). Never hard-code these numbers.
// One effect of a skill (sim EffectKind), e.g. { kind: 'gapCloser', value: 10, speed: 16, secs: 10 }.
export type EffectInfo = {
  kind: string; stat?: string; pct?: number; value?: number; radius?: number; speed?: number; factor?: number; secs?: number;
  duration?: number; mod?: string; tags?: string[]; cond?: number; threshold?: number; foes?: boolean;
};
export type SkillInfo = { id: number; name: string; pct: number; value: number; effects: EffectInfo[] };
// A unit-roster entry (classes[i] is sim class i; a retired id's slot is null).
export type ClassInfo = {
  name: string; hp: number; dmgLo: number; dmgHi: number; cooldown: number; range: number; speed: number; sight: number; swing: number;
  ranged: boolean; kite: boolean; target: string; engage: string; /* temperament */ projectile?: 'arrow' | 'bolt'; tags: string[]; bonus: { tag: string; mult: number }[];
  matchup: number[]; skills: SkillInfo[]; traits: EffectInfo[];
};
// A unit type's entry, for ids that are live by construction (an army's or a
// faction's units): a retired id here is a bug, not data.
export function classOf(data: { classes: (ClassInfo | null)[] }, cls: number): ClassInfo {
  const c = data.classes[cls];
  if (!c) throw new Error(`unit type ${cls} is retired or unknown`);
  return c;
}
// A unit's role, as the sim's plans see it (roleOf): 0 melee, 1 ranged, 2 caster.
export const roleOf = (c: ClassInfo): 0 | 1 | 2 => (c.tags.includes('caster') ? 2 : c.ranged ? 1 : 0);
// A faction (sim.Faction): the unit types a side fields. Lore and looks: factions/<name>.json.
export type FactionInfo = { id: number; name: string; units: number[]; playable: boolean };
export type GameData = {
  version: string; classes: (ClassInfo | null)[]; formations: string[]; armySize: number; deployZone: [number, number, number, number];
  factions: FactionInfo[];
  plans: { id: number; name: string; tactic: boolean; chips: CustomPlanJSON }[]; // battle plans, index = sim.Plan; tactics split the army in two groups; chips: the plan as its nearest chips; words in plans.ts
  chips: ChipVocab; // the custom-plan vocabulary (sim/custom.go)
  hitAt: number; // share of a swing at which the blow lands or the shot leaves
};
// Custom plans (sim/custom.go). Every value is the sim's: 0 is "own".
export type ChipsJSON = { target: number; engage: number; trigger: number; keep: number; leash: number; fallBack: number; stayNear: number; kite: number; edge: number };
export type Roles3 = [ChipsJSON, ChipsJSON, ChipsJSON]; // melee, ranged, caster
export type GroupJSON = { opening: Roles3; then: Roles3; trigger: number; seconds: number };
export type CustomPlanJSON = { roles: Roles3; groups: GroupJSON[] }; // groups: none, or left and right
export type ChipVocab = {
  targets: { id: number; name: string }[]; engages: { id: number; name: string }[];
  trigger: number[]; keep: number[]; leash: number[]; fallBack: number[]; // per step; index 0 is "own"
  triggers: { id: number; name: string }[]; triggerSeconds: number[];
  edges: { id: number; name: string; roles: [boolean, boolean, boolean]; effects: EffectInfo[] }[];
};
export const NO_CHIPS: ChipsJSON = { target: 0, engage: 0, trigger: 0, keep: 0, leash: 0, fallBack: 0, stayNear: 0, kite: 0, edge: 0 };
export const emptyRoles = (): Roles3 => [{ ...NO_CHIPS }, { ...NO_CHIPS }, { ...NO_CHIPS }];
export const PLAN_CUSTOM = 255; // sim.PlanCustom: a side fighting by a custom plan

// Every unit of the current battle (wasm units()), Red's generated army too.
export type UnitLoadout = { cls: number; skills: number[]; group: number }; // group: 0 left, 1 right in a tactic

// Unit labels, icons and blurbs come from each unit's look (units/looks.ts uiOf).

// sim.Formation values (wasm/data().formations index); pb Formation = sim + 1
export const PRESETS = [1, 2, 3, 4, 5, 6, 7];
export const CUSTOM = 0;
