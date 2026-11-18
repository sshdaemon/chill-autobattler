// Runs the Go sim (WASM) off the main thread at a fixed 20 Hz (× speed), so
// the battle never depends on the frame rate. Posts a snapshot every tick.
declare const Go: new () => { importObject: WebAssembly.Imports; run(i: WebAssembly.Instance): Promise<void> };
type SimApi = {
  version: string;
  create(input: Uint8Array): string;
  map(): Uint8Array;
  step(n: number): Uint8Array;
  outcome(): Uint8Array;
  data(): string;
  units(): string;
  formationSlots(formation: number, classes: number[]): Uint8Array | null;
  fog(): Uint8Array | null;
};
// preview: set the battle up and show tick 0 (deployment) without playing it
export type ToWorker =
  | { type: 'start' | 'preview'; input: Uint8Array }
  | { type: 'speed'; speed: number }
  | { type: 'stop' }
  | { type: 'slots'; classes: number[] }; // sim classes (0..2); replies with every preset's positions
export type FromWorker =
  | { type: 'ready'; version: string; data: string }
  | { type: 'slots'; slots: Record<number, { x: number; z: number }[]> } // keyed by sim formation (1..7)
  | { type: 'started'; map: ArrayBuffer; snapshot: ArrayBuffer; units: string } // units: JSON UnitLoadout[]
  | { type: 'snapshot'; snapshot: ArrayBuffer }
  | { type: 'fog'; seen: ArrayBuffer } // Blue's visible cells, 56×56, every 2 ticks
  | { type: 'done'; outcome: Uint8Array }
  | { type: 'error'; message: string };

const TICK_MS = 50;
const post = (m: FromWorker, transfer: Transferable[] = []) => (self as unknown as Worker).postMessage(m, transfer);
const buf = (u: Uint8Array) => u.buffer.slice(u.byteOffset, u.byteOffset + u.byteLength) as ArrayBuffer;

const ready = (async () => {
  await import(/* @vite-ignore */ `${self.location.origin}/wasm_exec.js`); // defines globalThis.Go
  const go = new Go();
  const { instance } = await WebAssembly.instantiateStreaming(fetch('/sim.wasm'), go.importObject);
  void go.run(instance); // never resolves; main() blocks to keep the API alive
  const sim = (globalThis as unknown as { chillSim: SimApi }).chillSim;
  post({ type: 'ready', version: sim.version, data: sim.data() });
  return sim;
})();

let timer: ReturnType<typeof setInterval> | undefined;
let speed = 1;

function loop(sim: SimApi) {
  clearInterval(timer);
  timer = setInterval(() => {
    const snap = buf(sim.step(1));
    const header = new Int32Array(snap, 0, 2), finished = header[1] !== 0; // header: tick, result, ...
    if (header[0] % 2 === 0 || finished) { const f = sim.fog(); if (f) { const seen = buf(f); post({ type: 'fog', seen }, [seen]); } }
    post({ type: 'snapshot', snapshot: snap }, [snap]); // transfers (detaches) snap
    if (finished) {
      clearInterval(timer);
      timer = undefined;
      post({ type: 'done', outcome: sim.outcome() });
    }
  }, TICK_MS / speed);
}

self.onmessage = async (e: MessageEvent<ToWorker>) => {
  const sim = await ready;
  const msg = e.data;
  if (msg.type === 'start' || msg.type === 'preview') {
    clearInterval(timer);
    timer = undefined;
    const err = sim.create(msg.input);
    if (err) return post({ type: 'error', message: err });
    const map = buf(sim.map()), snapshot = buf(sim.step(0));
    post({ type: 'started', map, snapshot, units: sim.units() }, [map, snapshot]);
    if (msg.type === 'start') loop(sim);
  } else if (msg.type === 'slots') {
    const slots: Record<number, { x: number; z: number }[]> = {};
    for (let f = 1; f <= 7; f++) {
      const raw = sim.formationSlots(f, msg.classes);
      if (!raw) continue;
      const a = new Int32Array(raw.buffer.slice(raw.byteOffset, raw.byteOffset + raw.byteLength));
      slots[f] = Array.from({ length: a.length / 2 }, (_, i) => ({ x: a[2 * i] / 65536, z: a[2 * i + 1] / 65536 }));
    }
    post({ type: 'slots', slots });
  } else if (msg.type === 'speed') {
    speed = msg.speed;
    if (timer !== undefined) loop(sim);
  } else if (msg.type === 'stop') {
    clearInterval(timer);
    timer = undefined;
  }
};
