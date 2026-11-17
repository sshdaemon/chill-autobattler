// Runs the Go sim (WASM) off the main thread at a fixed 20 Hz (× speed), so
// the battle never depends on the frame rate. Posts a snapshot every tick.
declare const Go: new () => { importObject: WebAssembly.Imports; run(i: WebAssembly.Instance): Promise<void> };
type SimApi = {
  version: string;
  create(input: Uint8Array): string;
  map(): Uint8Array;
  step(n: number): Uint8Array;
  outcome(): Uint8Array;
};
// preview: set the battle up and show tick 0 (deployment) without playing it
export type ToWorker = { type: 'start' | 'preview'; input: Uint8Array } | { type: 'speed'; speed: number } | { type: 'stop' };
export type FromWorker =
  | { type: 'ready'; version: string }
  | { type: 'started'; map: ArrayBuffer; snapshot: ArrayBuffer }
  | { type: 'snapshot'; snapshot: ArrayBuffer }
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
  post({ type: 'ready', version: sim.version });
  return sim;
})();

let timer: ReturnType<typeof setInterval> | undefined;
let speed = 1;

function loop(sim: SimApi) {
  clearInterval(timer);
  timer = setInterval(() => {
    const snap = buf(sim.step(1));
    const finished = new Int32Array(snap, 4, 1)[0] !== 0; // header: tick, result, ...
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
    post({ type: 'started', map, snapshot }, [map, snapshot]);
    if (msg.type === 'start') loop(sim);
  } else if (msg.type === 'speed') {
    speed = msg.speed;
    if (timer !== undefined) loop(sim);
  } else if (msg.type === 'stop') {
    clearInterval(timer);
    timer = undefined;
  }
};
