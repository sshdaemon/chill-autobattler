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
};

export const planWords = (name: string | undefined) =>
  (name && WORDS[name]) || { title: name ?? '?', desc: '' };

// A side's plan title (team 0 Blue, 1 Red) once Red's plan is real: from the
// countdown on (before the lock the preview's Red plan is a placeholder).
// Returns a string, so it is safe as a store selector.
export const shownPlan = (s: { phase: string; battlePlans?: [number, number]; data?: { plans: { name: string }[] } }, team: 0 | 1) =>
  s.battlePlans && ['countdown', 'fighting', 'done'].includes(s.phase)
    ? planWords(s.data?.plans[s.battlePlans[team]]?.name).title
    : undefined;
