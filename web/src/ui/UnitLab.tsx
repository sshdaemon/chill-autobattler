// Dev only (?lab or ?lab=<unit>): inspect units up close. Any unit type,
// recipe (or the lineup), team, build and variety seed, assembled by the game's
// own code; an orbit camera all the way round; any clip played, paused, looped,
// slowed or scrubbed; the battle's walk–jog blend on a moving floor with a
// foot-slide readout; each skill's animation and effect, any effect preset;
// the portrait. Runs without the game store (that would start a run).
import { useEffect, useMemo, useRef, useState } from 'react';
import { Canvas, useFrame, useThree } from '@react-three/fiber';
import { Grid, OrbitControls } from '@react-three/drei';
import * as THREE from 'three';
import type { GameData } from '../gamedata.ts';
import type { FromWorker } from '../sim/sim.worker.ts';
import { BUILDS, CLIPS, lookOf, setRoster, uiOf } from '../units/looks.ts';
import { loadClips, makeCharacter, type Character } from '../scene/models.ts';
import { Particles } from '../scene/Particles.tsx';
import { onSnapshot, play, PRESETS } from '../scene/effects.ts';
import { clipSpeed, locomotion } from '../scene/Units.tsx';
import { changedOnly, DEFAULTS, resetTuning, setTuning, tuning, type Tuning } from '../scene/tuning.ts';
import { portrait } from '../scene/portraits.ts';
import type { Snapshot, UnitSnap } from '../sim/snapshot.ts';

type Mode = { kind: 'clip'; clip: string; loop: boolean } | { kind: 'walk' };
type Unit = { ch: Character; x: number; z: number; recipe: string; team: number };
type Live = { // what the frame loop reads, without re-rendering React every frame
  units: Unit[]; mode: Mode; playing: boolean; rate: number; scrub: number | null; speed: number; spin: boolean;
  time: number; duration: number; blend: string; slide: string; floor: number; restart: number;
  wobble: boolean; clock: number; legSpeed: number; // speed changes like turns in battle; the legs' smoothed speed
};

const ALL = '*';
const snapOf = (x: number, z: number): UnitSnap => ({ team: 0, cls: 0, x, z, fx: 0, fz: 1, hp: 1, maxHp: 1, flags: 0, target: -1, skills: 0, kills: 0, dealt: 0 });
const panel: React.CSSProperties = { position: 'absolute', top: 8, left: 8, bottom: 8, width: 300, overflow: 'auto', background: '#fffffff0', borderRadius: 10, padding: '10px 12px', font: '12px system-ui', boxShadow: '0 4px 20px #0003' };
const row: React.CSSProperties = { display: 'flex', gap: 6, alignItems: 'center', flexWrap: 'wrap', margin: '4px 0' };
const h: React.CSSProperties = { margin: '10px 0 4px', fontSize: 12, textTransform: 'uppercase', letterSpacing: 0.5, color: '#556' };

export function UnitLab({ initial }: { initial: string }) {
  const [data, setData] = useState<GameData>();
  const [clipNames, setClipNames] = useState<string[]>([]);
  const [cls, setCls] = useState(0);
  const [recipe, setRecipe] = useState(ALL);
  const [team, setTeam] = useState(0); // 0 Blue, 1 Red, 2 both
  const [build, setBuild] = useState('');
  const [seed, setSeed] = useState(0);
  const [mode, setMode] = useState<Mode>({ kind: 'clip', clip: '', loop: true });
  const [playing, setPlaying] = useState(true);
  const [rate, setRate] = useState(1);
  const [speed, setSpeed] = useState(3.2);
  const [spin, setSpin] = useState(false);
  const [units, setUnits] = useState<Unit[]>([]);
  const [face, setFace] = useState<string>();
  const [fx, setFx] = useState('hit.flank');
  const [, tick] = useState(0);
  const [wobble, setWobble] = useState(false);
  const [, retune] = useState(0);
  const [copied, setCopied] = useState('');
  const live = useRef<Live>({ units: [], mode, playing, rate, scrub: null, speed, spin, time: 0, duration: 0, blend: '', slide: '', floor: 0, restart: 0, wobble, clock: 0, legSpeed: speed });

  useEffect(() => { // the sim worker only for data(); no battle is created
    const w = new Worker(new URL('../sim/sim.worker.ts', import.meta.url), { type: 'module' });
    w.onmessage = (e: MessageEvent<FromWorker>) => {
      if (e.data.type !== 'ready') return;
      const d = JSON.parse(e.data.data) as GameData;
      setRoster(d.classes);
      setData(d);
      const i = d.classes.findIndex((c) => c.name === initial);
      if (i >= 0) setCls(i);
      w.terminate();
    };
    void loadClips().then((c) => setClipNames(Object.keys(c).sort()));
    return () => w.terminate();
  }, [initial]);

  const look = data ? lookOf(cls) : undefined;
  const info = data?.classes[cls];
  useEffect(() => { if (look) setMode({ kind: 'clip', clip: look.anims.idle, loop: true }); }, [look]);

  useEffect(() => { // (re)assemble the characters
    if (!look) return;
    let live_ = true;
    const recipes = recipe === ALL ? look.recipes : look.recipes.filter((r) => r.name === recipe);
    const teams = team === 2 ? [0, 1] : [team];
    void Promise.all(recipes.flatMap((r, i) => teams.map(async (t) => {
      const ch = await makeCharacter(cls, r.name, t, seed + i, build ? { build } : {});
      ch.root.userData.cls = cls;
      return { ch, x: (i - (recipes.length - 1) / 2) * 1.4, z: teams.length > 1 && t === 1 ? -1.6 : 0, recipe: r.name, team: t };
    }))).then((u) => { if (live_) { setUnits(u); live.current.restart++; } });
    void portrait(cls, (recipe === ALL ? look.recipes[0] : look.recipes.find((r) => r.name === recipe)!).name, team === 1 ? 1 : 0, seed).then((f) => { if (live_) setFace(f); });
    return () => { live_ = false; };
  }, [look, cls, recipe, team, build, seed]);

  Object.assign(live.current, { units, mode, playing, rate, speed, spin, wobble });
  useEffect(() => { live.current.restart++; live.current.scrub = null; }, [mode]);
  useEffect(() => { const t = setInterval(() => tick((n) => n + 1), 150); return () => clearInterval(t); }, []); // readouts

  const fire = (preset: string) => {
    onSnapshot({ tick: 0, result: 0, units: [], shots: [], events: [] } as Snapshot, [], { fogOn: false, fogged: false, speed: 1, high: true, units: [], bolt: () => false });
    for (const u of units) play(preset, snapOf(u.x, u.z));
  };
  const presetSpeeds = useMemo(() => info ? [
    ['stand', 0], ['frozen', info.speed * 0.6], ['water', info.speed * 0.7], ['normal', info.speed], ['Swift', info.speed * 1.4], ['sprint', 9],
  ] as [string, number][] : [], [info]);

  if (!data || !look || !info) return <div style={{ padding: 20, font: '14px system-ui' }}>Loading the roster…</div>;
  const L = live.current;
  const unitClips = [...new Set([look.anims.idle, look.anims.walk, look.anims.move, look.anims.sprint, ...look.anims.attack, ...look.anims.hit, ...look.anims.death, look.anims.cheer,
    ...Object.values(look.skills).flatMap((s) => [s.action, s.attack])].filter(Boolean) as string[])];
  const clipBtn = (c: string) => (
    <button key={c} onClick={() => setMode({ kind: 'clip', clip: c, loop: !!CLIPS[c]?.loop || mode.kind !== 'clip' || mode.loop })}
      style={{ fontWeight: mode.kind === 'clip' && mode.clip === c ? 700 : 400 }}>{c}</button>
  );
  return (
    <div style={{ position: 'fixed', inset: 0, background: '#cfd8e0' }}>
      <div style={{ position: 'absolute', top: 0, bottom: 0, right: 0, left: 316 }}>
      <Canvas shadows camera={{ position: [0, 1.6, 4.2], fov: 38 }}>
        <hemisphereLight args={['#ffffff', '#667766', 1.8]} />
        <directionalLight position={[3, 6, 4]} intensity={2.4} castShadow shadow-mapSize={[2048, 2048]} />
        <Grid position={[0, 0.001, 0]} args={[30, 30]} cellSize={0.5} sectionSize={2} cellColor="#8a9a7a" sectionColor="#6a7a5a" fadeDistance={24} infiniteGrid />
        <Floor live={live} />
        <OrbitControls makeDefault target={[0, 0.9, 0]} minDistance={1.2} maxDistance={40} />
        <Driver live={live} />
        {units.map((u, i) => <primitive key={i} object={u.ch.root} position={[u.x, 0, u.z]} />)}
        <Particles />
        <Framing />
      </Canvas>
      </div>

      <div style={panel}>
        <b style={{ fontSize: 14 }}>{uiOf(cls).icon} Unit lab</b> <span style={{ color: '#778' }}>· dev only · <a href="/">back to the game</a></span>
        <div style={h}>Unit</div>
        <div style={row}>
          <select value={cls} onChange={(e) => { setCls(+e.target.value); setRecipe(ALL); setBuild(''); }}>
            {data.classes.map((c, i) => <option key={c.name} value={i}>{uiOf(i).label}</option>)}
          </select>
          <select value={recipe} onChange={(e) => setRecipe(e.target.value)}>
            <option value={ALL}>all recipes</option>
            {look.recipes.map((r) => <option key={r.name} value={r.name}>{r.name} ({r.build})</option>)}
          </select>
        </div>
        <div style={row}>
          team <select value={team} onChange={(e) => setTeam(+e.target.value)}><option value={0}>Blue</option><option value={1}>Red</option><option value={2}>both</option></select>
          build <select value={build} onChange={(e) => setBuild(e.target.value)}><option value="">recipe's</option>{Object.keys(BUILDS).map((b) => <option key={b}>{b}</option>)}</select>
        </div>
        <div style={row}>variety <input type="range" min={0} max={20} value={seed} onChange={(e) => setSeed(+e.target.value)} /> {seed}</div>
        <div style={{ color: '#667' }}>{info.hp} HP · {info.dmgLo}–{info.dmgHi} dmg · {info.range} m · {info.speed} m/s · {info.ranged ? info.projectile : 'melee'}</div>

        <div style={h}>View</div>
        <div style={row}>
          <label><input type="checkbox" checked={spin} onChange={(e) => setSpin(e.target.checked)} /> turntable</label>
          <button onClick={() => window.dispatchEvent(new CustomEvent('lab-frame', { detail: 'close' }))}>close-up</button>
          <button onClick={() => window.dispatchEvent(new CustomEvent('lab-frame', { detail: 'face' }))}>face</button>
          <button onClick={() => window.dispatchEvent(new CustomEvent('lab-frame', { detail: 'battle' }))}>battle camera</button>
        </div>
        <div style={{ color: '#889' }}>drag to orbit · scroll to zoom · right-drag to pan</div>

        <div style={h}>Animation</div>
        <div style={row}>
          <button onClick={() => setPlaying(!playing)}>{playing ? '⏸ pause' : '▶ play'}</button>
          {mode.kind === 'clip' && <label><input type="checkbox" checked={mode.loop} onChange={(e) => setMode({ ...mode, loop: e.target.checked })} /> loop</label>}
          speed <input type="range" min={0.05} max={2} step={0.05} value={rate} onChange={(e) => setRate(+e.target.value)} style={{ width: 70 }} /> {rate.toFixed(2)}×
        </div>
        {mode.kind === 'clip' && (
          <div style={row}>
            <input type="range" min={0} max={L.duration || 1} step={0.005} value={L.scrub ?? L.time} style={{ flex: 1 }}
              onChange={(e) => { L.scrub = +e.target.value; setPlaying(false); }} />
            {(L.scrub ?? L.time).toFixed(2)} / {L.duration.toFixed(2)} s
          </div>
        )}
        <div style={{ ...row, marginTop: 6 }}><i>this unit:</i>{unitClips.map(clipBtn)}</div>
        <details><summary>all clips ({clipNames.length})</summary><div style={row}>{clipNames.map(clipBtn)}</div></details>

        <div style={h}>Walking (battle blend)</div>
        <div style={row}>
          <button onClick={() => setMode({ kind: 'walk' })} style={{ fontWeight: mode.kind === 'walk' ? 700 : 400 }}>walk on the floor</button>
          <input type="range" min={0} max={10} step={0.05} value={speed} onChange={(e) => { setSpeed(+e.target.value); setMode({ kind: 'walk' }); }} style={{ width: 90 }} /> {speed.toFixed(2)} m/s
        </div>
        <div style={row}>{presetSpeeds.map(([n, v]) => <button key={n} onClick={() => { setSpeed(+v.toFixed(2)); setMode({ kind: 'walk' }); }}>{n}</button>)}</div>
        {mode.kind === 'walk' && <div style={{ color: '#445' }}>{L.blend}<br />{L.slide}</div>}

        <div style={h}>Tuning (walk & run)</div>
        <div style={{ color: '#667' }}>Live; saved in this browser and used by battles too (dev only).</div>
        <TuneSlider k="walkSpeed" label="walk clip speed" min={0.5} max={2} step={0.01} unit="m/s" onChange={() => retune((n) => n + 1)} />
        <TuneSlider k="jogSpeed" label="jog clip speed" min={3} max={9} step={0.02} unit="m/s" onChange={() => retune((n) => n + 1)} />
        <TuneSlider k="sprintSpeed" label="sprint clip speed" min={5} max={13} step={0.05} unit="m/s" onChange={() => retune((n) => n + 1)} />
        <TuneSlider k="blendBias" label="walk ← mix → jog" min={-0.5} max={0.5} step={0.01} onChange={() => retune((n) => n + 1)} />
        <TuneSlider k="walkMin" label="slowest walk" min={0.2} max={1} step={0.05} unit="×" onChange={() => retune((n) => n + 1)} />
        <TuneSlider k="jogMax" label="fastest jog" min={1} max={2.5} step={0.05} unit="×" onChange={() => retune((n) => n + 1)} />
        <TuneSlider k="sprintMax" label="fastest sprint" min={1} max={3} step={0.05} unit="×" onChange={() => retune((n) => n + 1)} />
        <TuneSlider k="speedSmooth" label="speed smoothing" min={0} max={0.5} step={0.01} unit="s" onChange={() => retune((n) => n + 1)} />
        <div style={row}>
          <label><input type="checkbox" checked={tuning.ground} onChange={(e) => { setTuning({ ground: e.target.checked }); retune((n) => n + 1); }} /> keep feet on the floor</label>
        </div>
        <TuneSlider k="groundSmooth" label="floor smoothing" min={0} max={0.2} step={0.005} unit="s" onChange={() => retune((n) => n + 1)} />
        <div style={row}>
          <label title="speed drops to 55% and back every 0.6 s, like sharp turns in battle: shows what speed smoothing does"><input type="checkbox" checked={wobble} onChange={(e) => { setWobble(e.target.checked); setMode({ kind: 'walk' }); }} /> wobble speed (turns)</label>
        </div>
        <div style={row}>
          <button onClick={() => { const txt = JSON.stringify(changedOnly()); void navigator.clipboard?.writeText(txt).catch(() => undefined); setCopied(txt === '{}' ? 'all defaults' : txt); }}>copy values</button>
          <button onClick={() => { resetTuning(); retune((n) => n + 1); setCopied(''); }}>reset</button>
        </div>
        {copied && <div style={{ font: '11px ui-monospace, monospace', background: '#f2f5fa', padding: 4, borderRadius: 4, wordBreak: 'break-all' }}>{copied}</div>}

        <div style={h}>Skills</div>
        <div style={row}>
          {info.skills.map((s) => {
            const sl = look.skills[String(s.id)] ?? {};
            const clip = sl.action ?? sl.attack;
            return <button key={s.id} title={[clip && `clip ${clip}`, sl.fx && `effect ${sl.fx}`, sl.pitch && `lean ${sl.pitch}`].filter(Boolean).join(', ') || 'stats only'}
              onClick={() => { if (clip) setMode({ kind: 'clip', clip, loop: false }); if (sl.fx) setTimeout(() => fire(sl.fx!), 50); }}>{s.name}</button>;
          })}
        </div>
        <div style={h}>Effects</div>
        <div style={row}>
          <select value={fx} onChange={(e) => setFx(e.target.value)}>{Object.keys(PRESETS).map((p) => <option key={p}>{p}</option>)}</select>
          <button onClick={() => fire(fx)}>fire</button>
        </div>
        <div style={h}>Portrait</div>
        {face && <img src={face} width={108} height={108} style={{ borderRadius: 8, border: `3px solid ${team === 1 ? '#d63a2e' : '#2f6fe8'}` }} alt="" />}
      </div>
    </div>
  );
}

// One tuning value as a slider, with its default marked.
function TuneSlider({ k, label, min, max, step, unit = '', onChange }: { k: keyof Tuning; label: string; min: number; max: number; step: number; unit?: string; onChange: () => void }) {
  const v = tuning[k] as number, d = DEFAULTS[k] as number;
  return (
    <div style={{ ...row, margin: '2px 0' }}>
      <span style={{ width: 108 }}>{label}</span>
      <input type="range" min={min} max={max} step={step} value={v} style={{ width: 90 }}
        onChange={(e) => { setTuning({ [k]: +e.target.value } as Partial<Tuning>); onChange(); }} />
      <span style={{ width: 52, fontWeight: v !== d ? 700 : 400 }} title={`default ${d}`}>{v.toFixed(step < 0.01 ? 3 : 2)}{unit}</span>
      {v !== d && <button style={{ padding: '0 4px' }} title={`back to ${d}`} onClick={() => { setTuning({ [k]: d } as Partial<Tuning>); onChange(); }}>↺</button>}
    </div>
  );
}

// Moves the floor under walking units (they stay put; the ground slides back).
function Floor({ live }: { live: React.MutableRefObject<Live> }) {
  const ref = useRef<THREE.Mesh>(null);
  const tex = useMemo(() => {
    const c = document.createElement('canvas');
    c.width = c.height = 128;
    const g = c.getContext('2d')!;
    g.fillStyle = '#9fb08a'; g.fillRect(0, 0, 128, 128);
    g.fillStyle = '#8fa07a'; g.fillRect(0, 0, 64, 64); g.fillRect(64, 64, 64, 64);
    const t = new THREE.CanvasTexture(c);
    t.wrapS = t.wrapT = THREE.RepeatWrapping;
    t.repeat.set(30, 30); // 1 m tiles over 30 m
    t.colorSpace = THREE.SRGBColorSpace;
    return t;
  }, []);
  useFrame(() => { tex.offset.y = -(live.current.floor % 1); });
  return <mesh ref={ref} rotation-x={-Math.PI / 2} receiveShadow><planeGeometry args={[30, 30]} /><meshStandardMaterial map={tex} /></mesh>;
}

// Camera presets (close-up, face, battle distance).
function Framing() {
  const { camera, controls } = useThree() as unknown as { camera: THREE.PerspectiveCamera; controls: { target: THREE.Vector3; update(): void } | null };
  useEffect(() => {
    const on = (e: Event) => {
      const kind = (e as CustomEvent<string>).detail;
      const [pos, target] = kind === 'face' ? [[0, 1.6, 1.2], [0, 1.5, 0]] : kind === 'battle' ? [[0, 14, 11], [0, 0.5, 0]] : [[0, 1.6, 4.2], [0, 0.9, 0]];
      camera.position.set(pos[0], pos[1], pos[2]);
      controls?.target.set(target[0], target[1], target[2]);
      controls?.update();
    };
    window.addEventListener('lab-frame', on);
    return () => window.removeEventListener('lab-frame', on);
  }, [camera, controls]);
  return null;
}

// Plays the chosen clip or the walk blend on every unit, each frame.
function Driver({ live }: { live: React.MutableRefObject<Live> }) {
  const seen = useRef({ restart: -1, feet: [] as { y: number; z: number }[], samples: [] as number[], minY: 9 });
  useFrame((_, dt) => {
    const L = live.current, S = seen.current;
    if (!L.units.length) return;
    const fresh = S.restart !== L.restart;
    S.restart = L.restart;
    const step = L.playing ? dt * L.rate : 0;
    // the true ground speed (wobbling like turns, if asked) and the legs' smoothed one, as in battle
    L.clock += step;
    const groundSpeed = L.wobble ? L.speed * (Math.floor(L.clock / 0.6) % 2 ? 0.55 : 1) : L.speed;
    L.legSpeed += (groundSpeed - L.legSpeed) * (tuning.speedSmooth > 0 ? 1 - Math.exp(-step / tuning.speedSmooth) : 1);
    for (const u of L.units) {
      const ch = u.ch;
      if (L.spin) ch.root.rotation.y += dt * 0.6;
      if (L.mode.kind === 'clip') {
        const a = ch.actions[L.mode.clip];
        if (!a) continue;
        if (fresh) {
          ch.mixer.stopAllAction();
          a.reset();
          a.setLoop(L.mode.loop ? THREE.LoopRepeat : THREE.LoopOnce, Infinity);
          a.clampWhenFinished = true;
          a.weight = 1;
          a.timeScale = 1;
          a.play();
        }
        if (L.scrub !== null) a.time = Math.min(L.scrub, a.getClip().duration - 1e-4);
        ch.mixer.update(L.scrub !== null ? 0 : step);
        ch.ground(false, 1 / 30);
        ch.tick?.(L.scrub !== null ? 1 / 60 : step); // capes keep settling while scrubbing
        L.time = a.time; L.duration = a.getClip().duration;
      } else {
        const look = lookOf(clsOf(ch));
        const w = ch.actions[look.anims.walk], j = ch.actions[look.anims.move];
        if (!w || !j) continue;
        if (fresh) { ch.mixer.stopAllAction(); w.reset().play(); j.reset().play(); }
        const wc = w.getClip(), jc = j.getClip();
        const l = locomotion(L.legSpeed, ch.root.scale.x, { speed: clipSpeed(wc.name, 'walk'), duration: wc.duration }, { speed: clipSpeed(jc.name, 'jog'), duration: jc.duration });
        j.weight = l.k; w.weight = 1 - l.k; j.timeScale = l.jogRate; w.timeScale = l.walkRate;
        ch.mixer.update(step);
        if (l.k > 0 && l.k < 1) { w.time = (j.time / jc.duration) * wc.duration; ch.mixer.update(0); }
        ch.ground(tuning.ground, step || 1 / 60); // as in battle: feet on the floor
        // the unit stays put on a moving floor: the air streams past it instead
        const ry = ch.root.rotation.y;
        ch.tick?.(step, new THREE.Vector3(-Math.sin(ry) * L.legSpeed, 0, -Math.cos(ry) * L.legSpeed));
        L.blend = `blend ${(100 * (1 - l.k)).toFixed(0)}% walk / ${(100 * l.k).toFixed(0)}% jog · walk ${l.walkRate.toFixed(2)}× · jog ${l.jogRate.toFixed(2)}×` + (L.wobble ? ` · legs at ${L.legSpeed.toFixed(2)} m/s` : '');
      }
    }
    if (L.mode.kind === 'walk') {
      L.floor += groundSpeed * step;
      // foot slide on the first unit: a planted foot should move back exactly as fast as the floor
      const ch = L.units[0].ch;
      ch.root.updateMatrixWorld(true);
      const feet: THREE.Object3D[] = [];
      ch.root.traverse((o) => { if (o.name === 'ball_l' || o.name === 'ball_r') feet.push(o); });
      if (fresh) { S.feet = []; S.samples = []; S.minY = 9; }
      feet.forEach((f, k) => {
        const p = f.getWorldPosition(new THREE.Vector3());
        S.minY = Math.min(S.minY, p.y);
        const prev = S.feet[k];
        if (prev && step > 0 && p.y < S.minY + 0.015 && prev.y < S.minY + 0.015) {
          const footBack = -(p.z - prev.z) / step; // m/s, same clock as the floor
          S.samples.push(footBack);
          if (S.samples.length > 120) S.samples.shift();
        }
        S.feet[k] = { y: p.y, z: p.z };
      });
      if (S.samples.length > 10 && L.speed > 0.2 && !L.spin) {
        const s = [...S.samples].sort((a, b) => a - b), m = s[Math.floor(s.length / 2)];
        L.slide = `planted foot slides ${Math.abs(m - groundSpeed).toFixed(2)} m/s over the floor (${Math.round((100 * Math.abs(m - groundSpeed)) / Math.max(0.1, groundSpeed))}% of the speed)`;
      } else L.slide = L.spin ? 'foot slide: turn the turntable off' : '';
    }
  });
  return null;
}

// the unit type a character was built for (set by UnitLab when assembling)
const clsOf = (ch: Character) => (ch.root.userData.cls as number | undefined) ?? 0;
