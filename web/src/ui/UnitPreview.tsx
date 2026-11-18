// Dev only (?preview=<unit>): one unit type's recipes on both teams, stepped
// through its animations, skills and effects, plus portraits, for review and
// for the screenshots /new-unit puts in its PR. window.previewShot(step) jumps
// to a step; window.previewSteps lists them. Uses the game's own character,
// effect and particle code, but not the store (that would start a run).
import { useEffect, useMemo, useState } from 'react';
import { Canvas } from '@react-three/fiber';
import type { ClassInfo, GameData } from '../gamedata.ts';
import type { FromWorker } from '../sim/sim.worker.ts';
import { CLIPS, lookOf, setRoster, uiOf } from '../units/looks.ts';
import { makeCharacter, type Character } from '../scene/models.ts';
import { Particles } from '../scene/Particles.tsx';
import { onSnapshot, play } from '../scene/effects.ts';
import { portrait } from '../scene/portraits.ts';
import type { Snapshot, UnitSnap } from '../sim/snapshot.ts';

type Step = { name: string; clip?: string; at?: number; fx?: string; note: string };

function stepsFor(cls: number, info: ClassInfo): Step[] {
  const look = lookOf(cls), a = look.anims;
  const steps: Step[] = [
    { name: 'idle', clip: a.idle, at: 0.5, note: 'idle' },
    { name: 'move', clip: a.move, at: 0.3, note: 'move' },
  ];
  for (const c of a.attack) steps.push({ name: `attack:${c}`, clip: c, at: CLIPS[c]?.hit, note: `attack ${c} at the blow` });
  for (const s of info.skills) {
    const l = look.skills[String(s.id)] ?? {};
    const clip = l.action ?? l.attack;
    if (!clip && !l.fx && !l.pitch) continue;
    steps.push({ name: `skill:${s.name}`, clip: clip ?? a.idle, at: clip ? (CLIPS[clip]?.hit ?? 0.35) : 0.5, fx: l.fx, note: `${s.name}: ${[l.action && `action ${l.action}`, l.attack && `attack ${l.attack}`, l.fx && `effect ${l.fx}`, l.pitch && `lean ${l.pitch}`].filter(Boolean).join(', ')}` });
  }
  steps.push({ name: 'hit', clip: a.hit[0], at: 0.15, note: 'hit reaction' });
  steps.push({ name: 'death', clip: a.death[0], at: 9, note: 'death (end)' });
  steps.push({ name: 'cheer', clip: a.cheer, at: 0.6, note: 'cheer' });
  steps.push({ name: 'portraits', note: 'portraits' });
  return steps;
}

const snapOf = (x: number, z: number): UnitSnap => ({ team: 0, cls: 0, x, z, fx: 0, fz: 1, hp: 1, maxHp: 1, flags: 0, target: -1, skills: 0, kills: 0, dealt: 0 });

export function UnitPreview({ name }: { name: string }) {
  const [data, setData] = useState<GameData>();
  const [chars, setChars] = useState<{ ch: Character; x: number; z: number; team: number; recipe: string }[]>([]);
  const [step, setStep] = useState(0);
  const [faces, setFaces] = useState<string[]>([]);
  useEffect(() => { // the sim worker only for data(); no battle is created
    const w = new Worker(new URL('../sim/sim.worker.ts', import.meta.url), { type: 'module' });
    w.onmessage = (e: MessageEvent<FromWorker>) => {
      if (e.data.type !== 'ready') return;
      const d = JSON.parse(e.data.data) as GameData;
      setRoster(d.classes);
      setData(d);
      w.terminate();
    };
    return () => w.terminate();
  }, []);
  const cls = data ? data.classes.findIndex((c) => c.name === name) : -1;
  const steps = useMemo(() => (data && cls >= 0 ? stepsFor(cls, data.classes[cls]) : []), [data, cls]);

  useEffect(() => {
    if (cls < 0) return;
    let live = true;
    const recipes = lookOf(cls).recipes;
    void Promise.all(recipes.flatMap((r, i) => [0, 1].map(async (team) => {
      const ch = await makeCharacter(cls, r.name, team, i);
      return { ch, x: (i - (recipes.length - 1) / 2) * 1.6, z: team ? -1.8 : 0, team, recipe: r.name };
    }))).then((c) => { if (live) setChars(c); });
    void Promise.all(recipes.flatMap((r, i) => [portrait(cls, r.name, 0, i), portrait(cls, r.name, 1, i)])).then((f) => { if (live) setFaces(f); });
    return () => { live = false; };
  }, [cls]);

  useEffect(() => { // pose every character for the step, and fire its effect
    const s = steps[step];
    if (!s || !chars.length) return;
    for (const { ch } of chars) {
      ch.mixer.stopAllAction();
      const act = s.clip ? ch.actions[s.clip] : undefined;
      if (act) { act.reset().play(); act.time = Math.min(s.at ?? 0, act.getClip().duration - 1e-3); }
      ch.mixer.update(0);
      ch.settle?.(); // capes drape over the pose
    }
    if (s.fx) {
      onSnapshot({ tick: 0, result: 0, units: [], shots: [], events: [] } as Snapshot, [], { fogOn: false, fogged: false, speed: 1, high: true, units: [], bolt: () => false });
      for (const c of chars) play(s.fx, snapOf(c.x, c.z));
    }
  }, [step, steps, chars]);

  useEffect(() => {
    Object.assign(window, {
      previewSteps: steps.map((s) => s.name),
      previewShot: (n: string | number) => { const i = typeof n === 'number' ? n : steps.findIndex((s) => s.name === n); if (i >= 0) setStep(i); return steps[i]?.note ?? 'no such step'; },
      previewReady: () => chars.length > 0 && faces.length > 0,
      // saves the current step to design/units/<unit>/preview/<nn>-<step>.png (dev server)
      previewSave: async () => {
        const s = steps[step], file = `${String(step + 1).padStart(2, '0')}-${s.name.replace(/[^a-z0-9]+/gi, '-').toLowerCase()}.png`;
        let canvas = document.querySelector('canvas') as HTMLCanvasElement;
        if (s.name === 'portraits') { // the faces are HTML, not in the 3D canvas: draw them onto one
          canvas = document.createElement('canvas');
          canvas.width = faces.length * 118; canvas.height = 118;
          const g = canvas.getContext('2d')!;
          g.fillStyle = '#cfd8e0'; g.fillRect(0, 0, canvas.width, canvas.height);
          await Promise.all(faces.map((f, i) => new Promise<void>((done) => { const im = new Image(); im.onload = () => { g.drawImage(im, i * 118 + 5, 5, 108, 108); done(); }; im.src = f; })));
        }
        const blob = await new Promise<Blob>((ok) => canvas.toBlob((b) => ok(b!), 'image/png'));
        const r = await fetch(`/__preview-shot/${name}/${file}`, { method: 'POST', body: blob });
        return r.ok ? file : `failed: ${r.status}`;
      },
    });
  }, [steps, chars, faces, step, name]);

  if (!data) return <div style={{ padding: 20 }}>Loading the roster…</div>;
  if (cls < 0) return <div style={{ padding: 20 }}>No roster entry “{name}”. Units: {data.classes.map((c) => c.name).join(', ')}</div>;
  const s = steps[step];
  return (
    <div style={{ position: 'fixed', inset: 0, background: '#cfd8e0' }}>
      <Canvas shadows gl={{ preserveDrawingBuffer: true }} camera={{ position: [0, 2.2, 6.2], fov: 38 }} onCreated={({ camera }) => camera.lookAt(0, 0.9, -0.9)}>
        <hemisphereLight args={['#ffffff', '#667766', 1.8]} />
        <directionalLight position={[3, 6, 4]} intensity={2.4} castShadow />
        <mesh rotation-x={-Math.PI / 2} receiveShadow position-z={-0.9}><planeGeometry args={[14, 7]} /><meshStandardMaterial color="#9fb08a" /></mesh>
        {chars.map(({ ch, x, z }, i) => <primitive key={i} object={ch.root} position={[x, 0, z]} />)}
        <Particles />
      </Canvas>
      <div style={{ position: 'absolute', top: 10, left: 12, font: '13px system-ui', background: '#fffd', padding: '6px 10px', borderRadius: 8 }}>
        <b>{uiOf(cls).icon} {uiOf(cls).label}</b> · {lookOf(cls).recipes.map((r) => r.name).join(', ')} · front Blue, back Red
        <div>Step {step + 1}/{steps.length}: <b>{s?.note}</b></div>
        <div style={{ marginTop: 4, display: 'flex', flexWrap: 'wrap', gap: 4, maxWidth: 520 }}>
          {steps.map((x, i) => <button key={x.name} onClick={() => setStep(i)} style={{ fontWeight: i === step ? 700 : 400 }}>{x.name}</button>)}
        </div>
      </div>
      {s?.name === 'portraits' && (
        <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 10, background: '#cfd8e0ee', flexWrap: 'wrap' }}>
          {faces.map((f, i) => <img key={i} src={f} width={108} height={108} style={{ borderRadius: 10, border: `3px solid ${i % 2 ? '#d63a2e' : '#2f6fe8'}` }} alt="" />)}
        </div>
      )}
    </div>
  );
}
