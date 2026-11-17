import { create } from 'zustand';
import { create as createMsg, toBinary, fromBinary } from '@bufbuild/protobuf';
import {
  BattleInputSchema, BattleOutcomeSchema, Formation, UnitClass, type BattleInput, type BattleOutcome,
} from './gen/chill/v1/battle_pb.ts';
import { decodeMap, decodeSnapshot, type Obstacle, type SimEvent, type Snapshot } from './sim/snapshot.ts';
import type { FromWorker, ToWorker } from './sim/sim.worker.ts';
import { battleApi } from './api.ts';

export const TICK_MS = 50;

export type Float = { id: number; x: number; z: number; text: string; color: string; born: number };

type State = {
  version?: string;
  input?: BattleInput;
  obstacles: Obstacle[];
  prev?: Snapshot; curr?: Snapshot; currAt: number; // two latest snapshots, for interpolation
  speed: number;
  outcome?: BattleOutcome;
  verdict?: string;
  error?: string;
  floats: Float[];
  start(cfg: { seed: bigint; formation: Formation; classes: UnitClass[]; redSkills: number }): void;
  setSpeed(s: number): void;
  verify(): Promise<void>;
};

const worker = new Worker(new URL('./sim/sim.worker.ts', import.meta.url), { type: 'module' });
const send = (m: ToWorker, transfer: Transferable[] = []) => worker.postMessage(m, transfer);
let floatId = 0;

export const useBattle = create<State>((set, get) => ({
  obstacles: [], currAt: 0, speed: 1, floats: [],
  start({ seed, formation, classes, redSkills }) {
    const version = get().version;
    if (!version) return;
    const input = createMsg(BattleInputSchema, {
      simVersion: version, seed, redSkillCount: redSkills,
      blue: { formation, units: classes.map((unitClass) => ({ unitClass })) },
    });
    set({ input, outcome: undefined, verdict: undefined, error: undefined, prev: undefined, curr: undefined, floats: [] });
    send({ type: 'start', input: toBinary(BattleInputSchema, input) });
  },
  setSpeed(speed) { set({ speed }); send({ type: 'speed', speed }); },
  async verify() {
    const { input, outcome } = get();
    if (!input || !outcome) return;
    set({ verdict: 'checking…' });
    try {
      const r = await battleApi.verifyBattle({ input, claimed: outcome });
      set({ verdict: r.fair ? `✅ fair (server checksum ${r.outcome?.checksum})` : `❌ ${r.reason}` });
    } catch (e) {
      set({ verdict: `⚠️ ${(e as Error).message}` });
    }
  },
}));

function addFloats(snap: Snapshot, events: SimEvent[]) {
  const now = performance.now();
  const floats = useBattle.getState().floats.filter((f) => now - f.born < 1000);
  for (const e of events) {
    const u = snap.units[e.unit];
    if (!u) continue;
    const label = ({ 3: ['💨 Dash', '#fff'], 4: ['🌀 Blink', '#c9a8ff'], 5: ['❄️ Slowed', '#9fd8ff'], 7: ['💀 Execute', '#ff5050'] } as Record<number, [string, string]>)[e.kind];
    if (e.kind === 1) {
      const flank = e.flags & 1, strong = e.flags & 2, weak = e.flags & 4;
      const text = `${Math.round(e.amount)}${flank ? ' x2' : ''}${strong ? ' ▲' : weak ? ' ▼' : ''}`;
      floats.push({ id: floatId++, x: u.x, z: u.z, text, color: flank ? '#ffd23f' : strong ? '#ff9a3d' : weak ? '#b9b9b9' : '#fff', born: now });
    } else if (label) {
      floats.push({ id: floatId++, x: u.x, z: u.z, text: label[0], color: label[1], born: now });
    }
  }
  return floats;
}

worker.onmessage = (e: MessageEvent<FromWorker>) => {
  const m = e.data;
  switch (m.type) {
    case 'ready': useBattle.setState({ version: m.version }); break;
    case 'error': useBattle.setState({ error: m.message }); break;
    case 'started': {
      const s = decodeSnapshot(m.snapshot);
      useBattle.setState({ obstacles: decodeMap(m.map), prev: s, curr: s, currAt: performance.now() });
      break;
    }
    case 'snapshot': {
      const s = decodeSnapshot(m.snapshot);
      useBattle.setState((st) => ({ prev: st.curr, curr: s, currAt: performance.now(), floats: addFloats(s, s.events) }));
      break;
    }
    case 'done':
      useBattle.setState({ outcome: fromBinary(BattleOutcomeSchema, m.outcome) });
      setTimeout(() => useBattle.setState({ floats: [] }), 1200); // let the last numbers fade
      break;
  }
};
