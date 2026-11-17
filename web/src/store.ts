import { create } from 'zustand';
import { create as createMsg, toBinary, fromBinary } from '@bufbuild/protobuf';
import { BattleInputSchema, BattleOutcomeSchema, BattleResult, type BattleOutcome } from './gen/chill/v1/battle_pb.ts';
import { ArmySchema, type Army, type Skill, type UnitSpec } from './gen/chill/v1/units_pb.ts';
import { RunState, type Run } from './gen/chill/v1/runstate_pb.ts';
import type { GetScoreboardResponse, ScoreRow } from './gen/chill/v1/run_pb.ts';
import { decodeMap, decodeSnapshot, type Obstacle, type SimEvent, type Snapshot } from './sim/snapshot.ts';
import type { FromWorker, ToWorker } from './sim/sim.worker.ts';
import { battleApi, runApi } from './api.ts';
import { authed, ensurePlayer } from './session.ts';
import { CUSTOM, type GameData } from './gamedata.ts';

export const TICK_MS = 50;
export const DEPLOY_SECONDS = 15;
const COUNTDOWN_MS = 1000;
const FX = 65536;

export type Float = { id: number; x: number; z: number; text: string; color: string; born: number };
// loading → picker → starting → deploying → locking → countdown → fighting → done (result, picks or run over)
export type Phase = 'loading' | 'picker' | 'starting' | 'deploying' | 'locking' | 'countdown' | 'fighting' | 'done';
export type Counts = [number, number, number]; // warriors, archers, mages (veterans included)
type Pt = { x: number; z: number };
export type { Run, ScoreRow };

// The battle being deployed or played: what the server issued (+ the locked army).
type Battle = { id: string; seed: bigint; redSkillCount: number; combatSeed?: bigint; lockedBlue?: Army };

type State = {
  version?: string;
  data?: GameData;
  phase: Phase;
  run?: Run;
  board?: GetScoreboardResponse;
  counts: Counts;
  formation: number; // sim.Formation; 0 = Custom
  positions?: Pt[]; // Custom: Blue's positions
  slots: Record<number, Pt[]>; // preset → positions for the current army (from the sim)
  deployLeft: number;
  dragging?: { index: number; at: Pt };
  battle?: Battle;
  obstacles: Obstacle[];
  prev?: Snapshot; curr?: Snapshot; currAt: number; // two latest snapshots, for interpolation
  speed: number;
  outcome?: BattleOutcome;
  verdict?: { state: 'checking' | 'fair' | 'unfair' | 'error'; text: string };
  picks: Record<number, Skill>; // veteran index → chosen skill
  error?: string;
  floats: Float[];
  init(): Promise<void>;
  newRun(): Promise<void>;
  setCount(cls: number, delta: number): void;
  toDeploy(): Promise<void>;
  setFormation(f: number): void;
  beginDrag(index: number, at: Pt): void;
  dragTo(at: Pt): void;
  endDrag(): void;
  ready(): Promise<void>;
  setSpeed(s: number): void;
  verify(): Promise<void>;
  choosePick(veteran: number, skill: Skill): void;
  confirmPicks(): Promise<void>;
  nextMatch(): void;
};

const worker = new Worker(new URL('./sim/sim.worker.ts', import.meta.url), { type: 'module' });
const send = (m: ToWorker, transfer: Transferable[] = []) => worker.postMessage(m, transfer);
let floatId = 0;
let deployTimer: ReturnType<typeof setInterval> | undefined;
const errMsg = (e: unknown) => (e as Error).message;

const savedCounts = (): Counts => {
  try {
    const c = JSON.parse(localStorage.getItem('army') ?? '');
    if (Array.isArray(c) && c.length === 3 && c.every((n) => Number.isInteger(n) && n >= 0)) return c as Counts;
  } catch { /* first visit or storage blocked */ }
  return [2, 2, 1];
};

export const vetCounts = (run?: Run): Counts => {
  const c: Counts = [0, 0, 0];
  for (const v of run?.veterans ?? []) c[v.unitClass - 1]++;
  return c;
};

// The army in the order the server requires: veterans first (exactly as
// stored), then recruits by class. cls is the sim class (0..2).
export function armySpecs(s: { run?: Run; counts: Counts }): { cls: number; skills: Skill[]; vet: boolean }[] {
  const vets = s.run?.veterans ?? [];
  const out = vets.map((v: UnitSpec) => ({ cls: v.unitClass - 1, skills: [...v.skills], vet: true }));
  const vc = vetCounts(s.run);
  s.counts.forEach((n, cls) => { for (let i = vc[cls]; i < n; i++) out.push({ cls, skills: [], vet: false }); });
  return out;
}
export const armyClasses = (s: { run?: Run; counts: Counts }) => armySpecs(s).map((u) => u.cls);

function blueArmy(s: State): Army {
  if (s.battle?.lockedBlue) return s.battle.lockedBlue; // replaying a locked battle after a reload
  return createMsg(ArmySchema, {
    formation: s.formation + 1, // pb Formation = sim Formation + 1
    units: armySpecs(s).map((u, i) => ({
      unitClass: u.cls + 1, // pb UnitClass = sim Class + 1
      skills: u.skills,
      position: s.formation === CUSTOM && s.positions
        ? { xRaw: BigInt(Math.round(s.positions[i].x * FX)), zRaw: BigInt(Math.round(s.positions[i].z * FX)) }
        : undefined,
    })),
  });
}

// What the browser plays: what the server issued plus Blue's army. Before the
// lock the combat seed is unknown; previews use 0 (only combat rolls differ).
function inputBytes(s: State, combatSeed: bigint) {
  const b = s.battle!;
  return toBinary(BattleInputSchema, createMsg(BattleInputSchema, {
    simVersion: s.version, seed: b.seed, redSkillCount: b.redSkillCount, combatSeed, blue: blueArmy(s),
  }));
}

const preview = () => send({ type: 'preview', input: inputBytes(useBattle.getState(), 0n) });

// Counts the picker should show: the remembered army, raised to cover the veterans.
function pickerCounts(run?: Run): Counts {
  const c = savedCounts(), v = vetCounts(run);
  const out = c.map((n, i) => Math.max(n, v[i])) as Counts;
  for (let i = 2; i >= 0; i--) while (out[i] > v[i] && out[0] + out[1] + out[2] > 5) out[i]--;
  return out;
}

export const useBattle = create<State>((set, get) => {
  const startDeployTimer = () => {
    clearInterval(deployTimer);
    set({ deployLeft: DEPLOY_SECONDS });
    deployTimer = setInterval(() => {
      if (get().phase !== 'deploying') return clearInterval(deployTimer);
      const left = get().deployLeft - 0.25;
      if (left <= 0) { clearInterval(deployTimer); void get().ready(); } else set({ deployLeft: left });
    }, 250);
  };
  const fight = (combatSeed: bigint) => {
    set({ phase: 'countdown' });
    setTimeout(() => {
      set({ phase: 'fighting' });
      send({ type: 'start', input: inputBytes(get(), combatSeed) });
    }, COUNTDOWN_MS);
  };
  const toPicker = (run: Run) => {
    const counts = pickerCounts(run);
    send({ type: 'slots', classes: armyClasses({ run, counts }) });
    set({ run, phase: 'picker', counts, outcome: undefined, verdict: undefined, battle: undefined, picks: {}, board: undefined });
  };
  const showRunOver = async (run: Run) => {
    set({ run, phase: 'done' });
    try { set({ board: await authed(() => runApi.getScoreboard({})) }); } catch (e) { set({ error: `Scoreboard: ${errMsg(e)}` }); }
  };
  // Continue a run from whatever state the server has it in (load and reload).
  const resume = async (run: Run) => {
    const b = run.battle;
    if (run.state === RunState.OVER) return showRunOver(run);
    if (run.state === RunState.PICKING) return set({ run, phase: 'done', picks: {} });
    if (!b) return toPicker(run);
    const battle: Battle = { id: b.battleId, seed: b.seed, redSkillCount: b.redSkillCount };
    if (b.locked) { // replay the locked battle from tick 0, then verify it
      set({ run, battle: { ...battle, combatSeed: b.combatSeed, lockedBlue: b.blue } });
      return fight(b.combatSeed);
    }
    set({ run, battle, counts: pickerCounts(run), formation: 2, positions: undefined, phase: 'deploying' });
    send({ type: 'slots', classes: armyClasses(get()) });
    preview();
    startDeployTimer();
  };

  return {
    phase: 'loading', counts: savedCounts(), formation: 2 /* 3-2 */, slots: {}, deployLeft: DEPLOY_SECONDS,
    obstacles: [], currAt: 0, speed: 1, floats: [], picks: {},

    async init() {
      try {
        await ensurePlayer();
        const { run } = await authed(() => runApi.getRun({}));
        if (!run || run.state === RunState.OVER) return get().newRun();
        await resume(run);
      } catch (e) {
        set({ phase: 'picker', error: `Couldn't reach the server: ${errMsg(e)}` });
      }
    },

    async newRun() {
      try {
        const { run } = await authed(() => runApi.startRun({}));
        if (run) toPicker(run);
      } catch (e) {
        set({ error: `Couldn't start a run: ${errMsg(e)}` });
      }
    },

    setCount(cls, delta) {
      const counts = [...get().counts] as Counts;
      const total = counts[0] + counts[1] + counts[2];
      if (delta > 0 && total >= (get().data?.armySize ?? 5)) return;
      counts[cls] = Math.max(vetCounts(get().run)[cls], counts[cls] + delta); // veterans can't be dismissed
      set({ counts });
      send({ type: 'slots', classes: armyClasses(get()) });
    },

    async toDeploy() {
      const { version, run, counts } = get();
      if (!version || !run) return;
      try { localStorage.setItem('army', JSON.stringify(counts)); } catch { /* storage blocked */ }
      send({ type: 'slots', classes: armyClasses(get()) });
      set({ phase: 'starting', error: undefined, outcome: undefined, verdict: undefined, floats: [], positions: undefined,
        formation: get().formation === CUSTOM ? 2 : get().formation, curr: undefined, prev: undefined });
      try {
        const st = await authed(() => battleApi.startBattle({ simVersion: version, runId: run.id }));
        set({ battle: { id: st.battleId, seed: st.seed, redSkillCount: st.redSkillCount }, phase: 'deploying' });
        preview();
        startDeployTimer();
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
        const r = await authed(() => battleApi.lockDeployment({ battleId: s.battle!.id, blue: blueArmy(s) }));
        set({ battle: { ...s.battle!, combatSeed: r.combatSeed } });
        fight(r.combatSeed);
      } catch (e) {
        set({ phase: 'picker', error: `Couldn't lock the deployment: ${errMsg(e)}` });
      }
    },

    setSpeed(speed) { set({ speed }); send({ type: 'speed', speed }); },

    async verify() {
      const { battle, outcome } = get();
      if (!battle || !outcome) return;
      set({ verdict: { state: 'checking', text: 'verifying with the server…' } });
      try {
        const r = await authed(() => battleApi.verifyBattle({ battleId: battle.id, claimed: outcome }));
        set({ verdict: r.fair ? { state: 'fair', text: 'verified fair by the server' } : { state: 'unfair', text: `not fair: ${r.reason}` },
          run: r.run, picks: {} });
        if (r.run?.state === RunState.OVER) void showRunOver(r.run);
      } catch (e) {
        set({ verdict: { state: 'error', text: `couldn't verify: ${errMsg(e)}` } });
      }
    },

    choosePick(veteran, skill) { set({ picks: { ...get().picks, [veteran]: skill } }); },

    async confirmPicks() {
      const { run, picks } = get();
      if (!run) return;
      try {
        const r = await authed(() => runApi.pickSkills({ runId: run.id, picks: run.owedPicks.map((v) => ({ veteran: v, skill: picks[v] })) }));
        if (r.run) set({ run: r.run, picks: {} });
      } catch (e) {
        set({ error: `Couldn't save the skills: ${errMsg(e)}` });
      }
    },

    nextMatch() {
      const run = get().run;
      if (!run) return;
      if (run.state === RunState.OVER) void get().newRun();
      else if (run.state === RunState.READY) toPicker(run);
    },
  };
});

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
    case 'ready':
      useBattle.setState({ version: m.version, data: JSON.parse(m.data) as GameData });
      void useBattle.getState().init();
      break;
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
