import { Canvas } from '@react-three/fiber';
import { OrbitControls } from '@react-three/drei';
import { Formation, UnitClass } from './gen/chill/v1/battle_pb.ts';
import { useBattle } from './store.ts';
import { RESULT, UF } from './sim/snapshot.ts';
import { Arena } from './scene/Arena.tsx';
import { Units } from './scene/Units.tsx';
import { Shots } from './scene/Shots.tsx';
import { Floats } from './scene/Floats.tsx';

const FORMATIONS: [Formation, string][] = [
  [Formation.FORMATION_SPREAD, 'Spread'], [Formation.FORMATION_3_2, '3-2'], [Formation.FORMATION_2_3, '2-3'],
  [Formation.FORMATION_4_1, '4-1'], [Formation.FORMATION_1_4, '1-4'], [Formation.FORMATION_0_5, '0-5'], [Formation.FORMATION_5_0, '5-0'],
];
const CLASSES: [UnitClass, string][] = [[UnitClass.WARRIOR, '⚔️'], [UnitClass.ARCHER, '🏹'], [UnitClass.MAGE, '✨']];
const RESULT_TEXT = { [RESULT.blue]: 'Blue wins', [RESULT.red]: 'Red wins', [RESULT.draw]: 'Draw' } as Record<number, string>;

function Hud() {
  const { version, phase, battle, curr, outcome, verdict, error, speed, formation, classes, redSkills,
    newMap, setArmy, setRedSkills, fight, setSpeed, verify } = useBattle();
  const alive = (team: number) => curr?.units.filter((u) => u.team === team && !(u.flags & UF.dead)).length ?? 0;
  const deploying = phase === 'deploying';
  const busy = phase === 'starting' || phase === 'locking' || phase === 'fighting';

  return (
    <div className="hud">
      <b>Chill Autobattler</b> <span className="dim">sim {version ?? 'loading…'} · debug viewer</span>
      <div className="row">
        <button className={phase === 'idle' || phase === 'done' ? 'primary' : ''} disabled={!version || busy} onClick={newMap}>🗺 New map</button>
        <label>Red ★ <input type="number" min={0} max={25} value={redSkills} disabled={busy} onChange={(e) => setRedSkills(+e.target.value)} style={{ width: 44 }} /></label>
        {battle && <span className="dim" title={battle.battleId}>map #{String(battle.seed).slice(0, 6)}…</span>}
      </div>
      <div className="row">
        Blue {classes.map((c, i) => (
          <button key={i} title="click to change class" disabled={busy} onClick={() => setArmy({ classes: classes.map((x, j) => (j === i ? ((x % 3) + 1) as UnitClass : x)) })}>
            {CLASSES.find(([k]) => k === c)?.[1]}
          </button>
        ))}
        <select value={formation} disabled={busy} onChange={(e) => setArmy({ formation: +e.target.value })}>
          {FORMATIONS.map(([f, n]) => <option key={f} value={f}>{n}</option>)}
        </select>
      </div>
      <div className="row">
        <button className={deploying ? 'primary' : ''} disabled={!deploying} onClick={fight}>▶ Fight</button>
        {[1, 2, 4].map((s) => <button key={s} className={speed === s ? 'on' : ''} onClick={() => setSpeed(s)}>{s}×</button>)}
        <span className="dim">{{ idle: 'press New map', starting: 'starting…', deploying: 'deploying: pick a formation', locking: 'locking…', fighting: 'fighting', done: 'done' }[phase]}</span>
      </div>
      {curr && (
        <div className="row" data-testid="status">
          tick {curr.tick} ({(curr.tick / 20).toFixed(1)} s) · <span style={{ color: '#3a8cff' }}>Blue {alive(0)}</span> vs <span style={{ color: '#ff4040' }}>{alive(1)} Red</span>
        </div>
      )}
      {outcome && (
        <div className="row" data-testid="outcome">
          <b>{RESULT_TEXT[outcome.result]}</b> after {outcome.ticks} ticks
          <button onClick={verify}>Verify on server</button>
        </div>
      )}
      {verdict && <div data-testid="verdict">{verdict}</div>}
      {error && <div className="err">{error}</div>}
    </div>
  );
}

export function App() {
  return (
    <>
      <Canvas shadows camera={{ position: [0, 26, 22], fov: 45 }}>
        <color attach="background" args={['#8fb6d8']} />
        <hemisphereLight args={['#dfefff', '#445533', 1.5]} />
        <directionalLight position={[10, 25, 8]} intensity={2.4} castShadow shadow-mapSize={[2048, 2048]}
          shadow-camera-left={-18} shadow-camera-right={18} shadow-camera-top={18} shadow-camera-bottom={-18} />
        <Arena />
        <Units />
        <Shots />
        <Floats />
        <OrbitControls maxPolarAngle={1.35} target={[0, 0, 2]} />
      </Canvas>
      <Hud />
    </>
  );
}
