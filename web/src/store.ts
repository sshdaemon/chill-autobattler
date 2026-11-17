import { create } from 'zustand';
import { create as createMsg, toBinary, fromBinary } from '@bufbuild/protobuf';
import {
  ArmySchema, BattleInputSchema, BattleOutcomeSchema, BattleResult,
  type Army, type BattleOutcome, type StartBattleResponse,
} from './gen/chill/v1/battle_pb.ts';
import { decodeMap, decodeSnapshot, type Obstacle, type SimEvent, type Snapshot } from './sim/snapshot.ts';
import type { FromWorker, ToWorker } from './sim/sim.worker.ts';
import { battleApi } from './api.ts';
import { CUSTOM, type GameData } from './gamedata.ts';

export const TICK_MS = 50;
export const DEPLOY_SECONDS = 15;
const COUNTDOWN_MS = 1000;
const FX = 65536;

export type Float = { id: number; x: number; z: number; text: string; color: string; born: number };
export type Phase = 'picker' | 'starting' | 'deploying' | 'locking' | 'countdown' | 'fighting' | 'done';
export type Counts = [number, number, number]; // warriors, archers, mages
type Pt = { x: number; z: number };

type State = {
  version?: string;
  data?: GameData;
  phase: Phase;
  counts: Counts;
  formation: number; // sim.Formation; 0 = Custom
  positions?: Pt[]; // Custom: Blue's positions
  slots: Record<number, Pt[]>; // preset → positions for the current army (from the sim)
  deployLeft: number;
  dragging?: { index: number; at: Pt };
  battle?: StartBattleResponse;
  combatSeed?: bigint;
  obstacles: Obstacle[];
  prev?: Snapshot; curr?: Snapshot; currAt: number; // two latest snapshots, for interpolation
  speed: number;
  redSkills: number;
  outcome?: BattleOutcome;
  verdict?: { state: 'checking' | 'fair' | 'unfair' | 'error'; text: string };
  error?: string;
  floats: Float[];
  setCount(cls: number, delta: number): void;
  toDeploy(): Promise<void>;
  setFormation(f: number): void;
  beginDrag(index: number, at: Pt): void;
  dragTo(at: Pt): void;
  endDrag(): void;
  ready(): Promise<void>;
  setSpeed(s: number): void;
  setRedSkills(n: number): void;
  verify(): Promise<void>;
  nextMatch(): void;
};

const worker = new Worker(new URL('./sim/sim.worker.ts', import.meta.url), { type: 'module' });
const send = (m: ToWorker, transfer: Transferable[] = []) => worker.postMessage(m, transfer);
let floatId = 0;
let deployTimer: ReturnType<typeof setInterval> | undefined;

const savedCounts = (): Counts => {
  try {
    const c = JSON.parse(localStorage.getItem('army') ?? '');
    if (Array.isArray(c) && c.length === 3 && c.every((n) => Number.isInteger(n) && n >= 0)) return c as Counts;
  } catch { /* first visit or storage blocked */ }
  return [2, 2, 1];
};

// sim classes (0 warrior, 1 archer, 2 mage) in army order: warriors, archers, mages
export const armyClasses = (c: Counts) => c.flatMap((n, cls) => Array<number>(n).fill(cls));

function blueArmy(s: State): Army {
  const classes = armyClasses(s.counts);
  return createMsg(ArmySchema, {
    formation: s.formation + 1, // pb Formation = sim Formation + 1
    units: classes.map((cls, i) => ({
      unitClass: cls + 1, // pb UnitClass = sim Class + 1
      position: s.formation === CUSTOM && s.positions
        ? { xRaw: BigInt(Math.round(s.positions[i].x * FX)), zRaw: BigInt(Math.round(s.positions[i].z * FX)) }
        : undefined,
    })),
  });
}

// What the browser plays: what the server issued plus Blue's army. Before the
// lock the combat seed is unknown; previews use 0, which changes only combat
// rolls, never the map or the starting positions.
function inputBytes(s: State, combatSeed: bigint) {
  const b = s.battle!;
  return toBinary(BattleInputSchema, createMsg(BattleInputSchema, {
    simVersion: s.version, seed: b.seed, redSkillCount: b.redSkillCount, combatSeed, blue: blueArmy(s),
  }));
}

const preview = () => send({ type: 'preview', input: inputBytes(useBattle.getState(), 0n) });
const errMsg = (e: unknown) => (e as Error).message;

export const useBattle = create<State>((set, get) => ({
  phase: 'picker', counts: savedCounts(), formation: 2 /* 3-2 */, slots: {}, deployLeft: DEPLOY_SECONDS,
  obstacles: [], currAt: 0, speed: 1, redSkills: 0, floats: [],

  setCount(cls, delta) {
    const counts = [...get().counts] as Counts;
    const total = counts[0] + counts[1] + counts[2];
    if (delta > 0 && total >= (get().data?.armySize ?? 5)) return;
    counts[cls] = Math.max(0, counts[cls] + delta);
    set({ counts });
  },

  async toDeploy() {
    const { version, redSkills, counts } = get();
    if (!version) return;
    try { localStorage.setItem('army', JSON.stringify(counts)); } catch { /* storage blocked */ }
    send({ type: 'slots', classes: armyClasses(counts) });
    set({ phase: 'starting', error: undefined, outcome: undefined, verdict: undefined, combatSeed: undefined,
      floats: [], positions: undefined, formation: get().formation === CUSTOM ? 2 : get().formation, curr: undefined, prev: undefined });
    try {
      const battle = await battleApi.startBattle({ simVersion: version, redSkillCount: redSkills });
      set({ battle, phase: 'deploying', deployLeft: DEPLOY_SECONDS });
      preview();
      clearInterval(deployTimer);
      deployTimer = setInterval(() => {
        const left = get().deployLeft - 0.25;
        if (get().phase !== 'deploying') return clearInterval(deployTimer);
        if (left <= 0) { clearInterval(deployTimer); void get().ready(); } else set({ deployLeft: left });
      }, 250);
    } catch (e) {
      set({ phase: 'picker', error: `Couldn't start a battle: ${errMsg(e)}` });
    }
  },

  setFormation(formation) {
    if (get().phase !== 'deploying' || formation === CUSTOM) return;
    set({ formation, positions: undefined });
    preview();
  },

  beginDrag(index, at) { if (get().phase === 'deploying') set({ dragging: { index, at } }); },
  dragTo(at) {
    const d = get().dragging, z = get().data?.deployZone;
    if (!d || !z) return;
    const m = 0.35; // unit radius: stay fully inside the zone
    set({ dragging: { index: d.index, at: { x: Math.min(z[1] - m, Math.max(z[0] + m, at.x)), z: Math.min(z[3] - m, Math.max(z[2] + m, at.z)) } } });
  },
  endDrag() {
    const { dragging, curr, phase } = get();
    if (!dragging) return;
    set({ dragging: undefined });
    if (phase !== 'deploying' || !curr) return;
    // Custom: every Blue unit where it stands now, the dragged one at its drop point.
    // The sim snaps it onto a walkable spot when the preview is rebuilt.
    const blue = curr.units.filter((u) => u.team === 0).map((u) => ({ x: u.x, z: u.z }));
    blue[dragging.index] = dragging.at;
    set({ formation: CUSTOM, positions: blue });
    preview();
  },

  async ready() {
    const s = get();
    if (s.phase !== 'deploying' || !s.battle) return;
    clearInterval(deployTimer);
    set({ phase: 'locking', dragging: undefined, error: undefined });
    try {
      const r = await battleApi.lockDeployment({ battleId: s.battle.battleId, blue: blueArmy(s) });
      set({ combatSeed: r.combatSeed, phase: 'countdown' });
      setTimeout(() => {
        set({ phase: 'fighting' });
        send({ type: 'start', input: inputBytes(get(), r.combatSeed) });
      }, COUNTDOWN_MS);
    } catch (e) {
      set({ phase: 'picker', error: `Couldn't lock the deployment: ${errMsg(e)}` });
    }
  },

  setSpeed(speed) { set({ speed }); send({ type: 'speed', speed }); },
  setRedSkills(redSkills) { set({ redSkills: Math.max(0, Math.min(25, redSkills | 0)) }); },

  async verify() {
    const { battle, outcome } = get();
    if (!battle || !outcome) return;
    set({ verdict: { state: 'checking', text: 'verifying with the server…' } });
    try {
      const r = await battleApi.verifyBattle({ battleId: battle.battleId, claimed: outcome });
      set({ verdict: r.fair ? { state: 'fair', text: 'verified fair by the server' } : { state: 'unfair', text: `not fair: ${r.reason}` } });
    } catch (e) {
      set({ verdict: { state: 'error', text: `couldn't verify: ${errMsg(e)}` } });
    }
  },

  nextMatch() { set({ phase: 'picker', outcome: undefined, verdict: undefined, battle: undefined }); },
}));

export const resultText = (r?: BattleResult) =>
  r === BattleResult.BLUE_WIN ? 'Victory' : r === BattleResult.RED_WIN ? 'Defeat' : r === BattleResult.DRAW ? 'Draw' : '';

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
    case 'ready': useBattle.setState({ version: m.version, data: JSON.parse(m.data) as GameData }); break;
    case 'slots': useBattle.setState({ slots: m.slots }); break;
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
      void useBattle.getState().verify();
      break;
  }
};

addEventListener('pointerup', () => useBattle.getState().endDrag()); // drops outside the canvas
