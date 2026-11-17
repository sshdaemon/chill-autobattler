export type SimResult = { winner: number; ticks: number; checksum: string };

const worker = new Worker(new URL('./sim.worker.ts', import.meta.url), { type: 'module' });
const pending = new Map<number, (r: { version: string; result: SimResult }) => void>();
let nextId = 0;
worker.onmessage = (e) => { pending.get(e.data.id)?.(e.data); pending.delete(e.data.id); };

export function runBattle(seed: bigint, maxTicks = 4000) {
  const id = nextId++;
  return new Promise<{ version: string; result: SimResult }>((resolve) => {
    pending.set(id, resolve);
    worker.postMessage({ id, seed: seed.toString(), maxTicks });
  });
}
