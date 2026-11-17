import { create } from 'zustand';
import { create as createMsg, toBinary, fromBinary } from '@bufbuild/protobuf';
import {
  ArmySchema, BattleInputSchema, BattleOutcomeSchema, Formation, UnitClass,
  type Army, type BattleOutcome, type StartBattleResponse,
} from './gen/chill/v1/battle_pb.ts';
import { decodeMap, decodeSnapshot, type Obstacle, type SimEvent, type Snapshot } from './sim/snapshot.ts';
import type { FromWorker, ToWorker } from './sim/sim.worker.ts';
import { battleApi } from './api.ts';

export const TICK_MS = 50;

export type Float = { id: number; x: number; z: number; text: string; color: string; born: number };

// idle → (New map) deploying → (Fight) fighting → done → (Verify)
export type Phase = 'idle' | 'starting' | 'deploying' | 'locking' | 'fighting' | 'done';

type State = {
  version?: string;
  phase: Phase;
  battle?: StartBattleResponse; // issued by the server
  combatSeed?: bigint; // revealed by LockDeployment
  formation: Formation;
  classes: UnitClass[];
  redSkills: number;
  obstacles: Obstacle[];
  prev?: Snapshot; curr?: Snapshot; currAt: number; // two latest snapshots, for interpolation
  speed: number;
  outcome?: BattleOutcome;
  verdict?: string;
  error?: string;
  floats: Float[];
  newMap(): Promise<void>;
  setArmy(a: { formation?: Formation; classes?: UnitClass[] }): void;
  setRedSkills(n: number): void;
  fight(): Promise<void>;
  setSpeed(s: number): void;
  verify(): Promise<void>;
};

const worker = new Worker(new URL('./sim/sim.worker.ts', import.meta.url), { type: 'module' });
const send = (m: ToWorker, transfer: Transferable[] = []) => worker.postMessage(m, transfer);
let floatId = 0;

const blueArmy = (s: State): Army => createMsg(ArmySchema, { formation: s.formation, units: s.classes.map((unitClass) => ({ unitClass })) });

// The input the browser plays: what the server issued plus Blue's army.
// Before locking, the combat seed is unknown; previews use 0, which only
// changes combat rolls, never the map or the starting positions.
function inputBytes(s: State, combatSeed: bigint) {
  const b = s.battle!;
  return toBinary(BattleInputSchema, createMsg(BattleInputSchema, {
    simVersion: s.version, seed: b.seed, redSkillCount: b.redSkillCount, combatSeed, blue: blueArmy(s),
  }));
}

const errMsg = (e: unknown) => (e as Error).message;

export const useBattle = create<State>((set, get) => ({
  phase: 'idle', formation: Formation.FORMATION_3_2, redSkills: 0,
  classes: [UnitClass.WARRIOR, UnitClass.WARRIOR, UnitClass.ARCHER, UnitClass.ARCHER, UnitClass.MAGE],
  obstacles: [], currAt: 0, speed: 1, floats: [],

  async newMap() {
    const { version, redSkills } = get();
    if (!version) return;
    set({ phase: 'starting', error: undefined, outcome: undefined, verdict: undefined, combatSeed: undefined, floats: [] });
    try {
      const battle = await battleApi.startBattle({ simVersion: version, redSkillCount: redSkills });
      set({ battle, phase: 'deploying' });
      send({ type: 'preview', input: inputBytes(get(), 0n) });
    } catch (e) {
      set({ phase: 'idle', error: `StartBattle: ${errMsg(e)}` });
    }
  },
  setArmy(a) {
    set(a);
    if (get().phase === 'deploying') send({ type: 'preview', input: inputBytes(get(), 0n) });
  },
  setRedSkills(redSkills) { set({ redSkills }); },
  async fight() {
    const s = get();
    if (s.phase !== 'deploying' || !s.battle) return;
    set({ phase: 'locking', error: undefined });
    try {
      const r = await battleApi.lockDeployment({ battleId: s.battle.battleId, blue: blueArmy(s) });
      set({ combatSeed: r.combatSeed, phase: 'fighting' });
      send({ type: 'start', input: inputBytes(get(), r.combatSeed) });
    } catch (e) {
      set({ phase: 'deploying', error: `LockDeployment: ${errMsg(e)}` });
    }
  },
  setSpeed(speed) { set({ speed }); send({ type: 'speed', speed }); },
  async verify() {
    const { battle, outcome } = get();
    if (!battle || !outcome) return;
    set({ verdict: 'checking…' });
    try {
      const r = await battleApi.verifyBattle({ battleId: battle.battleId, claimed: outcome });
      set({ verdict: r.fair ? `✅ fair (server checksum ${r.outcome?.checksum})` : `❌ ${r.reason}` });
    } catch (e) {
      set({ verdict: `⚠️ ${errMsg(e)}` });
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
      useBattle.setState({ outcome: fromBinary(BattleOutcomeSchema, m.outcome), phase: 'done' });
      setTimeout(() => useBattle.setState({ floats: [] }), 1200); // let the last numbers fade
      break;
  }
};
