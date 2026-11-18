// Locomotion tuning (presentation only): the numbers behind legs matching
// ground speed and feet staying on the floor. Defaults are the shipped values;
// in dev builds the unit lab edits them live and saves them in this browser,
// and battles use the saved values too, so they can be felt in real fights.
// Production always uses the defaults.

export type Tuning = {
  walkSpeed: number; jogSpeed: number; sprintSpeed: number; // m/s each clip covers at 1× on our characters
  speedSmooth: number; // s: smoothing of the unit's measured ground speed
  walkMin: number; jogMax: number; sprintMax: number; // playback limits
  blendBias: number; // -0.5..0.5: shifts the walk/jog mix (cadence adjusts so feet still don't slide)
  ground: boolean; groundSmooth: number; // keep soles on the floor while moving; s
};

export const DEFAULTS: Tuning = {
  walkSpeed: 1.04, jogSpeed: 6.24, sprintSpeed: 9.57,
  speedSmooth: 0.12,
  walkMin: 0.5, jogMax: 1.4, sprintMax: 1.6,
  blendBias: 0,
  ground: true, groundSmooth: 0.02,
};

const KEY = 'lab.tuning';

function saved(): Partial<Tuning> {
  if (!import.meta.env.DEV) return {};
  try { return JSON.parse(localStorage.getItem(KEY) ?? '{}') as Partial<Tuning>; } catch { return {}; }
}

export const tuning: Tuning = { ...DEFAULTS, ...saved() };

export function setTuning(patch: Partial<Tuning>) {
  Object.assign(tuning, patch);
  if (!import.meta.env.DEV) return;
  try { localStorage.setItem(KEY, JSON.stringify(changedOnly())); } catch { /* storage blocked */ }
}

export function resetTuning() { setTuning({ ...DEFAULTS }); }

// what differs from the defaults (what's worth reporting)
export function changedOnly(): Partial<Tuning> {
  return Object.fromEntries(Object.entries(tuning).filter(([k, v]) => DEFAULTS[k as keyof Tuning] !== v)) as Partial<Tuning>;
}

export const isTuned = () => Object.keys(changedOnly()).length > 0;
