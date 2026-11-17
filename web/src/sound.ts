// Battle sound effects, synthesized with Web Audio at runtime: no files, no
// licences. Presentation only, driven by sim events and phase changes.
const KEY = 'sound';
type Prefs = { muted: boolean; volume: number };
let prefs: Prefs = { muted: false, volume: 0.6 };
try { prefs = { ...prefs, ...JSON.parse(localStorage.getItem(KEY) ?? '{}') }; } catch { /* first visit or storage blocked */ }

let ctx: AudioContext | undefined, master: GainNode | undefined;
const last: Record<string, number> = {};

// Browsers only allow audio after a user gesture.
addEventListener('pointerdown', () => {
  if (ctx) { void ctx.resume(); return; }
  ctx = new AudioContext();
  master = ctx.createGain();
  master.gain.value = prefs.muted ? 0 : prefs.volume;
  master.connect(ctx.destination);
}, { capture: true });

export const soundPrefs = () => prefs;
export function setSound(p: Partial<Prefs>) {
  prefs = { ...prefs, ...p };
  try { localStorage.setItem(KEY, JSON.stringify(prefs)); } catch { /* storage blocked */ }
  if (master) master.gain.value = prefs.muted ? 0 : prefs.volume;
}

function tone(freq: number, dur: number, type: OscillatorType, gain = 0.3, slideTo?: number, delay = 0) {
  if (!ctx || !master) return;
  const t = ctx.currentTime + delay, o = ctx.createOscillator(), g = ctx.createGain();
  o.type = type;
  o.frequency.setValueAtTime(freq, t);
  if (slideTo) o.frequency.exponentialRampToValueAtTime(slideTo, t + dur);
  g.gain.setValueAtTime(gain, t);
  g.gain.exponentialRampToValueAtTime(0.0001, t + dur);
  o.connect(g).connect(master);
  o.start(t);
  o.stop(t + dur + 0.02);
}

function noise(dur: number, freq: number, gain = 0.3, q = 1) {
  if (!ctx || !master) return;
  const t = ctx.currentTime, n = Math.max(1, Math.floor(ctx.sampleRate * dur));
  const buf = ctx.createBuffer(1, n, ctx.sampleRate), d = buf.getChannelData(0);
  for (let i = 0; i < n; i++) d[i] = Math.random() * 2 - 1; // sound only: never touches the battle
  const src = ctx.createBufferSource(), f = ctx.createBiquadFilter(), g = ctx.createGain();
  src.buffer = buf;
  f.type = 'bandpass'; f.frequency.value = freq; f.Q.value = q;
  g.gain.setValueAtTime(gain, t);
  g.gain.exponentialRampToValueAtTime(0.0001, t + dur);
  src.connect(f).connect(g).connect(master);
  src.start(t);
}

const RECIPES: Record<string, () => void> = {
  sword: () => { noise(0.09, 3200, 0.35, 2); tone(1400, 0.08, 'triangle', 0.08, 900); },
  arrowHit: () => noise(0.06, 900, 0.3, 1.5),
  boltHit: () => { tone(520, 0.18, 'sine', 0.18, 180); noise(0.12, 1800, 0.12, 3); },
  bow: () => tone(260, 0.12, 'triangle', 0.12, 120),
  cast: () => tone(420, 0.2, 'sine', 0.12, 900),
  death: () => { tone(220, 0.35, 'sawtooth', 0.12, 70); noise(0.2, 400, 0.15); },
  dash: () => noise(0.25, 700, 0.2, 0.7),
  blink: () => tone(900, 0.25, 'sine', 0.15, 2400),
  frost: () => { tone(1800, 0.15, 'sine', 0.08, 2600); noise(0.15, 6000, 0.08, 4); },
  execute: () => { tone(160, 0.3, 'square', 0.12, 60); noise(0.12, 2500, 0.2, 2); },
  fight: () => { tone(330, 0.15, 'square', 0.12); tone(440, 0.15, 'square', 0.12, undefined, 0.12); tone(660, 0.3, 'square', 0.12, undefined, 0.24); },
  victory: () => [523, 659, 784, 1047].forEach((f, i) => tone(f, 0.35, 'triangle', 0.16, undefined, i * 0.13)),
  defeat: () => [392, 330, 262, 196].forEach((f, i) => tone(f, 0.45, 'triangle', 0.14, undefined, i * 0.18)),
};

// Plays a sound, at most once per 40 ms per kind (so 4× playback doesn't machine-gun).
export function play(kind: keyof typeof RECIPES) {
  if (!ctx || prefs.muted || ctx.state !== 'running') return;
  const now = performance.now();
  if (now - (last[kind] ?? 0) < 40) return;
  last[kind] = now;
  RECIPES[kind]();
}
