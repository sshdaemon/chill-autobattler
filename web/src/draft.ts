// The draft before each battle (roadmap 1.3). Both sides pick five units in
// turns (sim DraftTurns: 1-2-2-2-2-1), the first side decided by a coin flip
// from the map seed. Red's picks are its generated army, revealed in a seeded
// order, so the whole draft comes from the sim (wasm draft()) and the server
// only checks the army Blue locks. Blue's veterans fill its first picks by
// themselves; every other Blue pick is a unit type of the run's faction and
// its level-1 ability, chosen within TURN_SECONDS per turn (or at random).

export const TURN_SECONDS = 15;
export const RED_PICK_MS = 900; // pause before each of Red's (and each veteran's) picks, to read them
export const BLUE = 0, RED = 1;

export type DraftUnit = { cls: number; skills: number[]; vet?: boolean };
// what wasm draft(seed, redSkills) returns
export type DraftPlan = { first: number; turns: number[]; redFaction: number; red: DraftUnit[]; reveal: number[] };

export type Draft = {
  plan: DraftPlan;
  steps: { side: number; turn: number }[]; // one per pick, in order
  at: number; // the next pick
  blue: DraftUnit[]; // in army order: veterans first, then recruits in pick order
  red: DraftUnit[]; // revealed so far
  vets: DraftUnit[]; // the run's veterans, placed as Blue's first picks
  turnEnds: number; // ms (Date.now) when Blue's current turn runs out
  nextAuto: number; // ms when the next automatic pick (Red's or a veteran) happens
};

export function stepsOf(plan: DraftPlan) {
  const steps: { side: number; turn: number }[] = [];
  plan.turns.forEach((n, turn) => {
    const side = (plan.first + turn) % 2;
    for (let k = 0; k < n; k++) steps.push({ side, turn });
  });
  return steps;
}

export function newDraft(plan: DraftPlan, vets: DraftUnit[], now: number): Draft {
  const d: Draft = { plan, steps: stepsOf(plan), at: 0, blue: [], red: [], vets, turnEnds: 0, nextAuto: now + RED_PICK_MS };
  startTurn(d, now);
  return d;
}

export const done = (d: Draft) => d.at >= d.steps.length;
export const current = (d: Draft) => d.steps[d.at];
// Blue picks by hand now (its turn, and no veteran left to place)
export const bluesChoice = (d: Draft) => !done(d) && current(d).side === BLUE && d.blue.length >= d.vets.length;
// picks left in the current turn
export const leftInTurn = (d: Draft) => d.steps.slice(d.at).filter((s) => s.turn === current(d)?.turn).length;

function startTurn(d: Draft, now: number) {
  if (!done(d) && (d.at === 0 || d.steps[d.at - 1].turn !== current(d).turn)) d.turnEnds = now + TURN_SECONDS * 1000;
}

function advance(d: Draft, now: number) {
  d.at++;
  d.nextAuto = now + RED_PICK_MS;
  startTurn(d, now);
}

// Blue picks a unit type and its level-1 ability. Returns false if it isn't Blue's choice.
export function pick(d: Draft, cls: number, skill: number, now: number): boolean {
  if (!bluesChoice(d)) return false;
  d.blue.push({ cls, skills: [skill] });
  advance(d, now);
  return true;
}

// Moves the draft along in time: Red's picks and Blue's veterans appear one
// by one, and when Blue's turn runs out its remaining picks are made at
// random from `fields` (the faction's unit types). Returns whether anything changed.
export function tick(d: Draft, now: number, fields: number[], abilities: number, rnd = Math.random): boolean {
  if (done(d)) return false;
  const s = current(d);
  if (s.side === RED || d.blue.length < d.vets.length) {
    if (now < d.nextAuto) return false;
    if (s.side === RED) d.red.push(d.plan.red[d.plan.reveal[d.red.length]]);
    else d.blue.push({ ...d.vets[d.blue.length], vet: true });
    advance(d, now);
    return true;
  }
  if (now < d.turnEnds) return false;
  while (!done(d) && current(d).side === BLUE && current(d).turn === s.turn) {
    d.blue.push({ cls: fields[Math.floor(rnd() * fields.length)], skills: [1 + Math.floor(rnd() * abilities)] });
    advance(d, now);
  }
  return true;
}
