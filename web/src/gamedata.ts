// Balance data as exported by the sim (wasm data()). Never hard-code these numbers.
export type ClassInfo = {
  name: string; hp: number; dmgLo: number; dmgHi: number; cooldown: number; range: number; speed: number; sight: number;
  ranged: boolean; kite: boolean; matchup: number[]; skills: { id: number; name: string; pct: number; value: number }[];
};
export type GameData = { version: string; classes: ClassInfo[]; formations: string[]; armySize: number; deployZone: [number, number, number, number] };

export const CLASS_ICON = ['⚔️', '🏹', '✨'];
export const CLASS_LABEL = ['Warrior', 'Archer', 'Mage'];
export const CLASS_BLURB = ['Sturdy melee fighter', 'Fragile, shoots from range, backs off from melee', 'Magic bolts, strong vs warriors, holds ground'];

// sim.Formation values (wasm/data().formations index); pb Formation = sim + 1
export const PRESETS = [1, 2, 3, 4, 5, 6, 7];
export const CUSTOM = 0;
