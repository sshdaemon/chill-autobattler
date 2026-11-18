// Factions' lore and looks (presentation only): one file per faction,
// factions/<name>.json, matched to the sim's factions (wasm data()) by name.
// A file with "coming": true is a faction not playable yet, shown as a teaser
// with its planned units.
export type FactionLook = {
  order: number; label: string; icon: string; colour: string; blurb: string; lore: string[];
  coming?: boolean; units?: { label: string; role: string }[];
};

const files = import.meta.glob<FactionLook>('./*.json', { eager: true, import: 'default' });
// by name, in display order
export const FACTIONS: Record<string, FactionLook> = Object.fromEntries(
  Object.entries(files).map(([path, f]) => [path.slice(2, -'.json'.length), f] as const).sort(([, a], [, b]) => a.order - b.order));

export const factionLook = (name: string): FactionLook =>
  FACTIONS[name] ?? { order: 99, label: name, icon: '🏳️', colour: '#667', blurb: '', lore: [] };
