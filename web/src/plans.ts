// Battle plans' words for players. The sim owns the list (wasm data().plans,
// index = sim.Plan); this maps a plan's sim name to its title and a one-line
// description. A plan without words here shows its sim name.
const WORDS: Record<string, { title: string; desc: string }> = {
  instinct: { title: 'Instinct', desc: 'Every unit fights its own way.' },
  'shield wall': { title: 'Shield wall', desc: 'Melee hold, braced against blows from the front; ranged shoot from behind them. Gives up the first move.' },
  'all-out charge': { title: 'All-out charge', desc: 'Everyone rushes the nearest enemy, faster than usual. No kiting, a little reckless.' },
  'hunt the archers': { title: 'Hunt the archers', desc: "Melee sprint past the front for the enemy's archers and casters, exposed while they do." },
  'focus fire': { title: 'Focus fire', desc: 'Advance together, gang up on one enemy at a time and hit harder. Slow to close in.' },
  skirmish: { title: 'Skirmish', desc: 'Melee screen the back line; ranged kite, quicker and with longer reach.' },
  'hammer and anvil': { title: 'Hammer and anvil', desc: 'Left (L) holds as a braced shield wall. Right (R) waits, then swings round the enemy\'s side once the left is fighting.' },
  reserve: { title: 'Reserve', desc: 'Left (L) advances and takes the first blows. Right (R) holds back and comes in a little stronger when the first ally falls.' },
  'hold, then charge': { title: 'Hold, then charge', desc: 'Everyone holds, braced, for 10 seconds while the enemy comes, then charges the nearest enemy.' },
};

export const planWords = (name: string | undefined) =>
  (name && WORDS[name]) || { title: name ?? '?', desc: '' };

// The client's id for "the custom plan" in the picker (the sim's is 255).
export const CUSTOM_PLAN = -1;
export const CUSTOM_WORDS = { title: 'Custom', desc: 'Your own plan, built from chips.' };

// Chip words (sim names → what players read).
export const TARGET_WORDS: Record<string, string> = {
  nearest: 'Nearest', wounded: 'Most wounded', rangedFirst: 'Archers first', strongest: 'Hardest hitter',
  focus: 'Gang up', focusReach: 'Gang up in reach', protect: 'Protect back line',
};
export const ENGAGE_WORDS: Record<string, string> = { charge: 'Charge', line: 'Advance in line', hold: 'Hold', flank: 'Go round' };
export const TRIGGER_WORDS: Record<string, string> = { after: 'After', 'ally falls': 'When an ally falls', 'other group fighting': 'When the other group fights' };
export const EDGE_WORDS: Record<string, string> = {
  'braced shields': 'Braced shields', reckless: 'Reckless', drilled: 'Drilled', 'long reach': 'Long reach', 'light feet': 'Light feet',
};
export const ROLE_NAMES = ['Melee', 'Ranged', 'Casters'] as const;
export const STAY_WORDS = ['Own', 'Behind the melee', 'Guard the back line'];
export const KITE_WORDS = ['Own', 'Kite', "Don't kite"];
export const STEP_WORDS = ['Own', 'Close', 'Mid', 'Far'];
export const LEASH_WORDS = ['Own', 'Short', 'Long'];

// An edge's effects in words, e.g. "+10% damage, takes 8% more damage".
export function edgeEffects(effects: { kind: string; stat?: string; pct?: number; cond?: number }[]): string {
  return effects.map((e) => {
    const p = e.pct ?? 0, sign = p > 0 ? '+' : '−', n = Math.abs(p);
    if (e.kind === 'damageTaken') return `takes ${n}% ${p < 0 ? 'less' : 'more'} damage${e.cond ? ' from the front' : ''}`;
    if (e.kind === 'statBonus') return `${sign}${n}% ${e.stat === 'hp' ? 'health' : e.stat === 'attackSpeed' ? 'attack speed' : e.stat}`;
    return e.kind;
  }).join(', ');
}

// A side's plan title (team 0 Blue, 1 Red) once Red's plan is real: from the
// countdown on (before the lock the preview's Red plan is a placeholder).
// Returns a string, so it is safe as a store selector.
export const shownPlan = (s: { phase: string; battlePlans?: [number, number]; data?: { plans: { name: string }[] } }, team: 0 | 1) =>
  s.battlePlans && ['countdown', 'fighting', 'done'].includes(s.phase)
    ? (s.battlePlans[team] === 255 ? CUSTOM_WORDS.title : planWords(s.data?.plans[s.battlePlans[team]]?.name).title)
    : undefined;
