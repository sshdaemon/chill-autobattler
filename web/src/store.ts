import { lastKill, onSnapshot, viewOf } from './scene/effects.ts';
import { create } from 'zustand';
import { create as createMsg, toBinary, fromBinary } from '@bufbuild/protobuf';
import { BattleInputSchema, BattleOutcomeSchema, BattleResult, type BattleOutcome } from './gen/chill/v1/battle_pb.ts';
import { ArmySchema, type Army, type UnitSpec } from './gen/chill/v1/units_pb.ts';

// an ability's number within its unit's roster entry (1..5)
type Skill = number;
import { RunState, type Run } from './gen/chill/v1/runstate_pb.ts';
import type { GetScoreboardResponse, ScoreRow } from './gen/chill/v1/run_pb.ts';
import { decodeMap, decodeSnapshot, type Obstacle, type SimEvent, type Snapshot } from './sim/snapshot.ts';
import type { FromWorker, ToWorker } from './sim/sim.worker.ts';
import { battleApi, runApi } from './api.ts';
import { authed, ensurePlayer } from './session.ts';
import { CUSTOM, type GameData, type UnitLoadout } from './gamedata.ts';
import { setRoster } from './units/looks.ts';
import { done as draftDone, newDraft, pick as draftPick, tick as draftTick, type Draft, type DraftPlan } from './draft.ts';
import { play as playSound } from './sound.ts';

export const TICK_MS = 50;
export const DEPLOY_SECONDS = 15;
const COUNTDOWN_MS = 1000;
const FX = 65536;

export type Float = { id: number; x: number; z: number; text: string; color: string; born: number };
// loading → faction (new run) → starting → drafting → deploying → locking → countdown → fighting → done (result, picks or run over)
export type Phase = 'loading' | 'faction' | 'starting' | 'drafting' | 'deploying' | 'locking' | 'countdown' | 'fighting' | 'done';
// A unit of Blue's army: sim class, abilities, and whether it's a veteran.
export type ArmyUnit = { cls: number; skills: Skill[]; vet: boolean };
type Pt = { x: number; z: number };
export type { Run, ScoreRow };

// The battle being deployed or played: what the server issued (+ the locked army).
type Battle = { id: string; seed: bigint; redSkillCount: number; redHealthPct: number; combatSeed?: bigint; lockedBlue?: Army };

type State = {
  version?: string;
  data?: GameData;
  phase: Phase;
  run?: Run;
  board?: GetScoreboardResponse;
  army: ArmyUnit[]; // Blue's army for the current battle, in the order the server requires (veterans first)
  draft?: Draft; // the draft in progress (phase drafting)
  formation: number; // sim.Formation; 0 = Custom
  plan: number; // Blue's battle plan (sim.Plan; 0 = Instinct), kept for the run
  battlePlans?: [number, number]; // the battle's [Blue, Red] plans from the sim; Red's is real only after the lock
  positions?: Pt[]; // Custom: Blue's positions
  slots: Record<number, Pt[]>; // preset → positions for the current army (from the sim)
  deployLeft: number;
  dragging?: { index: number; at: Pt };
  battle?: Battle;
  obstacles: Obstacle[];
  units: UnitLoadout[]; // every unit's class and skills (Red's too), from the sim at battle start
  prev?: Snapshot; curr?: Snapshot; currAt: number; // two latest snapshots, for interpolation
  speed: number;
  outcome?: BattleOutcome;
  verdict?: { state: 'checking' | 'fair' | 'unfair' | 'error'; text: string };
  picks: Record<number, Skill>; // veteran index → chosen skill
  error?: string;
  floats: Float[];
  fog?: Uint8Array; // Blue's visible cells (from the sim)
  fogOn: boolean;
  quality: 'high' | 'low'; // graphics; remembered per browser
  autoLowered: boolean; // the game dropped to Low because High was too slow
  lowDpr: boolean; // High renders at 1x pixel density: the first step when it runs slow
  qualityChosen: boolean; // the player picked the quality by hand this session: never override it
  tiltShift: boolean; // experimental depth-of-field (dev)
  selected?: number; // unit index
  follow: boolean; // camera follows the selected unit
  centerOn?: { index?: number; x?: number; z?: number; n: number }; // a request for the camera to jump to a unit or point
  init(): Promise<void>;
  select(i?: number): void;
  center(i: number): void;
  centerAt(x: number, z: number): void;
  toggleFog(): void;
  setQuality(q: 'high' | 'low', auto?: boolean): void;
  toggleTiltShift(): void;
  toggleFollow(): void;
  newRun(): Promise<void>; // to the faction screen
  startRun(faction: number): Promise<void>;
  surrender(): Promise<void>; // end the active run (score kept) and go to the faction screen
  beginDraft(plan: DraftPlan): void; // the sim's draft for the battle arrived
  draftPick(cls: number, skill: Skill): void;
  retryDraft(): void;
  setFormation(f: number): void;
  setPlan(p: number): void;
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

const savedQuality = (): 'high' | 'low' => {
  try { return localStorage.getItem('quality') === 'low' ? 'low' : 'high'; } catch { return 'high'; }
};

// The army in the order the server requires: veterans first (exactly as
// stored), then the drafted recruits (level 1: one ability each).
export const armySpecs = (s: { army: ArmyUnit[] }) => s.army;
export const armyClasses = (s: { army: ArmyUnit[] }) => s.army.map((u) => u.cls);

function blueArmy(s: State): Army {
  if (s.battle?.lockedBlue) return s.battle.lockedBlue; // replaying a locked battle after a reload
  return createMsg(ArmySchema, {
    formation: s.formation + 1, // pb Formation = sim Formation + 1
    plan: s.plan,
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
    simVersion: s.version, seed: b.seed, redSkillCount: b.redSkillCount, redHealthPct: b.redHealthPct, combatSeed, blue: blueArmy(s),
    blueFaction: s.run?.faction ?? 0,
  }));
}

const preview = () => send({ type: 'preview', input: inputBytes(useBattle.getState(), 0n) });

// Counts the picker should show: the remembered army, raised to cover the veterans.
// The unit types the run's faction fields.
export const factionUnits = (data: GameData | undefined, run: Run | undefined) =>
  data?.factions.find((x) => x.id === run?.faction)?.units ?? [];
const ABILITIES = 5; // abilities each unit type offers (sim MaxSkills)
let draftTimer: ReturnType<typeof setInterval> | undefined;

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
    // set the battle up with the real combat seed now, so the countdown can name Red's plan
    set({ phase: 'countdown', battlePlans: undefined });
    send({ type: 'preview', input: inputBytes(get(), combatSeed) });
    playSound('fight');
    setTimeout(() => {
      set({ phase: 'fighting' });
      send({ type: 'start', input: inputBytes(get(), combatSeed) });
    }, COUNTDOWN_MS);
  };
  // Next match: start the battle (its map seed fixes Red's army and the draft
  // order), then draft.
  const toDraft = async (run: Run) => {
    const { version } = get();
    if (!version) return;
    set({ run, phase: 'starting', army: [], draft: undefined, error: undefined, outcome: undefined, verdict: undefined, battle: undefined, picks: {}, board: undefined,
      floats: [], positions: undefined, fog: undefined, selected: undefined, follow: false, curr: undefined, prev: undefined,
      formation: get().formation === CUSTOM ? 2 : get().formation });
    try {
      const st = await authed(() => battleApi.startBattle({ simVersion: version, runId: run.id }));
      requestDraft({ id: st.battleId, seed: st.seed, redSkillCount: st.redSkillCount, redHealthPct: st.redHealthPct });
    } catch (e) {
      set({ phase: 'drafting', error: `Couldn't start a battle: ${errMsg(e)}` });
    }
  };
  // Ask the sim for the battle's draft (worker replies 'draft' → beginDraft).
  const requestDraft = (battle: Battle) => {
    set({ battle, phase: 'starting' });
    send({ type: 'draft', seed: battle.seed.toString(), redSkills: battle.redSkillCount });
  };
  const finishDraft = () => {
    clearInterval(draftTimer);
    const d = get().draft;
    if (!d) return;
    set({ army: d.blue.map((u) => ({ cls: u.cls, skills: u.skills, vet: !!u.vet })), phase: 'deploying', positions: undefined });
    send({ type: 'slots', classes: armyClasses(get()) });
    preview();
    startDeployTimer();
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
    if (!b) return toDraft(run);
    const battle: Battle = { id: b.battleId, seed: b.seed, redSkillCount: b.redSkillCount, redHealthPct: b.redHealthPct };
    if (b.locked) { // replay the locked battle from tick 0, then verify it
      set({ run, battle: { ...battle, combatSeed: b.combatSeed, lockedBlue: b.blue }, fog: undefined, selected: undefined });
      return fight(b.combatSeed);
    }
    set({ run, formation: 2, positions: undefined }); // not locked yet: draft it again
    requestDraft(battle);
  };

  return {
    phase: 'loading', army: [], formation: 2 /* 3-2 */, plan: 0, slots: {}, deployLeft: DEPLOY_SECONDS,
    obstacles: [], units: [], currAt: 0, speed: 1, floats: [], picks: {}, fogOn: true, follow: false,
    quality: savedQuality(), autoLowered: false, lowDpr: false, qualityChosen: false, tiltShift: false,
    setQuality(quality, auto = false) {
      if (auto && (get().qualityChosen || get().autoLowered)) return; // once per session, never over a manual choice
      set({ quality, autoLowered: auto, qualityChosen: get().qualityChosen || !auto });
      if (!auto) try { localStorage.setItem('quality', quality); } catch { /* storage blocked */ }
    },
    toggleTiltShift() { set({ tiltShift: !get().tiltShift }); },

    select(selected) { set({ selected, follow: selected === undefined ? false : get().follow }); },
    center(index) { set({ selected: index, centerOn: { index, n: (get().centerOn?.n ?? 0) + 1 } }); },
    centerAt(x, z) { set({ centerOn: { x, z, n: (get().centerOn?.n ?? 0) + 1 } }); },
    toggleFog() { set({ fogOn: !get().fogOn }); },
    toggleFollow() { if (get().selected !== undefined) set({ follow: !get().follow }); },

    async init() {
      try {
        await ensurePlayer();
        const { run } = await authed(() => runApi.getRun({}));
        if (!run || run.state === RunState.OVER) return set({ phase: 'faction' });
        await resume(run);
      } catch (e) {
        set({ phase: 'faction', error: `Couldn't reach the server: ${errMsg(e)}` });
      }
    },

    async newRun() {
      set({ phase: 'faction', run: undefined, outcome: undefined, verdict: undefined, battle: undefined, board: undefined, error: undefined });
    },

    async surrender() {
      const run = get().run;
      if (!run || run.state === RunState.OVER) return get().newRun();
      clearInterval(deployTimer);
      clearInterval(draftTimer);
      send({ type: 'stop' });
      try {
        await authed(() => runApi.surrenderRun({ runId: run.id }));
        await get().newRun();
      } catch (e) {
        set({ error: `Couldn't surrender: ${errMsg(e)}` });
      }
    },

    async startRun(faction) {
      try {
        const { run } = await authed(() => runApi.startRun({ faction }));
        set({ plan: 0 }); // a new run starts on Instinct
        if (run) await toDraft(run);
      } catch (e) {
        set({ error: `Couldn't start a run: ${errMsg(e)}` });
      }
    },

    beginDraft(plan: DraftPlan) {
      const run = get().run;
      if (get().phase !== 'starting' || !run) return;
      const vets = (run.veterans ?? []).map((v: UnitSpec) => ({ cls: v.unitClass - 1, skills: [...v.skills] }));
      set({ draft: newDraft(plan, vets, Date.now()), phase: 'drafting' });
      clearInterval(draftTimer);
      draftTimer = setInterval(() => {
        const d = get().draft;
        if (get().phase !== 'drafting' || !d) return clearInterval(draftTimer);
        if (draftTick(d, Date.now(), factionUnits(get().data, get().run), ABILITIES)) set({ draft: { ...d } });
        else if (draftDone(d)) finishDraft();
        else set({ draft: { ...d } }); // the turn's countdown
      }, 250);
    },

    draftPick(cls, skill) {
      const d = get().draft;
      if (d && draftPick(d, cls, skill, Date.now())) set({ draft: { ...d } });
    },

    retryDraft() { // the battle if it was started, else a new one
      const { run, battle } = get();
      if (battle && !battle.combatSeed) requestDraft(battle);
      else if (run) void toDraft(run);
    },

    setFormation(formation) {
      if (get().phase !== 'deploying' || formation === CUSTOM) return;
      set({ formation, positions: undefined });
      preview();
    },

    setPlan(plan) {
      if (get().phase !== 'deploying') return;
      set({ plan });
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
        set({ phase: 'drafting', draft: undefined, error: `Couldn't lock the deployment: ${errMsg(e)}` });
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
      else if (run.state === RunState.READY) void toDraft(run);
    },
  };
});

export const resultText = (r?: BattleResult) =>
  r === BattleResult.BLUE_WIN ? 'Victory' : r === BattleResult.RED_WIN ? 'Defeat' : r === BattleResult.DRAW ? 'Draw' : '';

// Sound for one sim event (presentation only).
function sound(snap: Snapshot, e: SimEvent) {
  // by the roster's projectile: none (melee), arrow or bolt
  const shot = (u?: { cls: number }) => (u ? useBattle.getState().data?.classes[u.cls]?.projectile : undefined);
  const by = shot(snap.units[e.other]); // attacker / shooter
  switch (e.kind) {
    case 1: playSound(by === 'arrow' ? 'arrowHit' : by === 'bolt' ? 'boltHit' : 'sword'); break;
    case 2: playSound('death'); break;
    case 3: playSound('dash'); break;
    case 4: playSound('blink'); break;
    case 5: playSound('frost'); break;
    case 7: playSound('execute'); break;
    case 8: playSound(shot(snap.units[e.unit]) === 'bolt' ? 'cast' : 'bow'); break;
  }
}

function addFloats(snap: Snapshot, events: SimEvent[]) {
  const now = performance.now();
  const floats = useBattle.getState().floats.filter((f) => now - f.born < 1000);
  for (const e of events) {
    const u = snap.units[e.unit];
    if (!u) continue;
    sound(snap, e);
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
      setRoster(useBattle.getState().data!.classes);
      void useBattle.getState().init();
      break;
    case 'slots': useBattle.setState({ slots: m.slots }); break;
    case 'draft': {
      const plan = m.plan ? JSON.parse(m.plan) as DraftPlan : undefined;
      if (plan) useBattle.getState().beginDraft(plan);
      else useBattle.setState({ phase: 'drafting', error: 'The sim could not set up the draft' });
      break;
    }
    case 'fog': useBattle.setState({ fog: new Uint8Array(m.seen) }); break;
    case 'error': useBattle.setState({ error: m.message }); break;
    case 'started': {
      const s = decodeSnapshot(m.snapshot);
      useBattle.setState({ obstacles: decodeMap(m.map), units: JSON.parse(m.units) as UnitLoadout[], prev: s, curr: s, currAt: performance.now(),
        battlePlans: m.plans ? JSON.parse(m.plans) as [number, number] : undefined });
      break;
    }
    case 'snapshot': {
      const s = decodeSnapshot(m.snapshot);
      onSnapshot(s, s.events, viewOf(useBattle.getState()));
      useBattle.setState((st) => ({ prev: st.curr, curr: s, currAt: performance.now(), floats: addFloats(s, s.events) }));
      break;
    }
    case 'done': {
      const outcome = fromBinary(BattleOutcomeSchema, m.outcome);
      playSound(outcome.result === BattleResult.BLUE_WIN ? 'victory' : 'defeat');
      lastKill();
      useBattle.setState({ outcome, phase: 'done' });
      setTimeout(() => useBattle.setState({ floats: [] }), 1200); // let the last numbers fade
      void useBattle.getState().verify();
      break;
    }
  }
};

addEventListener('pointerup', () => useBattle.getState().endDrag()); // drops outside the canvas

// dev: the store from the console (window.battle.getState()), whether or not the 3D view runs
if (import.meta.env.DEV) (window as unknown as { battle: typeof useBattle }).battle = useBattle;
