// Runs the Go sim (WASM) off the main thread so rendering never waits on it.
declare const Go: new () => { importObject: WebAssembly.Imports; run(i: WebAssembly.Instance): Promise<void> };
type SimApi = { version: string; run(seed: string, maxTicks: number): { winner: number; ticks: number; checksum: string } };

const ready = (async () => {
  await import(/* @vite-ignore */ `${self.location.origin}/wasm_exec.js`); // defines globalThis.Go
  const go = new Go();
  const { instance } = await WebAssembly.instantiateStreaming(fetch('/sim.wasm'), go.importObject);
  void go.run(instance); // never resolves; main() blocks to keep the API alive
  return (globalThis as unknown as { chillSim: SimApi }).chillSim;
})();

self.onmessage = async (e: MessageEvent<{ id: number; seed: string; maxTicks: number }>) => {
  const sim = await ready;
  const { id, seed, maxTicks } = e.data;
  self.postMessage({ id, version: sim.version, result: sim.run(seed, maxTicks) });
};
