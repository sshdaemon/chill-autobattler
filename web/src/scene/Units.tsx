import { useEffect, useRef, useState } from 'react';
import { useFrame, type ThreeEvent } from '@react-three/fiber';
import { Billboard, Html } from '@react-three/drei';
import type { Group } from 'three';
import { useBattle } from '../store.ts';
import { UF } from '../sim/snapshot.ts';
import { BattleResult } from '../gen/chill/v1/battle_pb.ts';
import { alpha, lerp } from './interp.ts';
import { lookFor, makeCharacter, type Character } from './models.ts';
import { FOGGED_FIGHT } from './Fog.tsx';
import { feet, flinches, skillActions, viewOf } from './effects.ts';
import { CLIPS, lookOf, skillLooks, type UnitLook } from '../units/looks.ts';
import { TEAM_RIM } from './models.ts';
import { Euler, Quaternion, Vector3 } from 'three';

const GOLD = TEAM_RIM[0].clone().setStyle('#ffd23f');
const UP = new Vector3(0, 1, 0), RIGHT = new Vector3(1, 0, 0), tmpQ = new Quaternion(), tmpAxis = new Vector3(), tmpE = new Euler();
const AIM_MAX = 0.96, FLINCH_FOR = 0.3, FLINCH_MAX = 0.21;

const TEAM = ['#3a8cff', '#ff4040'];
const FOGGED = ['starting', 'deploying', 'locking'];

// A direction in the character's own frame (Y up) as seen from a bone's parent
// space, so a rotation about it can be premultiplied onto the bone's pose.
const tmpM = new Quaternion();
function toBone(bone: import('three').Bone, dir: Vector3) {
  const root = bone.parent ? bone.parent.getWorldQuaternion(tmpM) : tmpM.identity();
  const charQ = new Quaternion();
  let o: import('three').Object3D | null = bone;
  while (o && !o.userData.isCharacterRoot && o.parent) o = o.parent;
  if (o) o.getWorldQuaternion(charQ);
  return dir.clone().applyQuaternion(charQ).applyQuaternion(root.invert()).normalize();
}
const localUp = (bone: import('three').Bone) => toBone(bone, UP);
const FADE_AFTER = 1.2, FADE_FOR = 1.5; // a body lies, then fades into the ground

// Stable per-unit pick among variants (attack, hit, death), so a replay looks the same.
const variant = (list: string[], index: number, k: number) => list[((Math.imul(index + 1, 2654435761) + Math.imul(k, 40503)) >>> 16) % list.length];
const QUICK = new Set(['death', 'hit', 'attack', 'action']); // fade into these fast

type Pick = { name: string; kind: 'death' | 'hit' | 'attack' | 'action' | 'move' | 'sprint' | 'cheer' | 'idle'; rate: number; restart: boolean };

// Which clip the sim state asks for (see unit-looks): death, hit, attack
// (the skill's attack if it has one, else a variant), a skill action started
// by an event, move, cheer, idle.
function choose(look: UnitLook, sk: ReturnType<typeof skillLooks>, flags: number, prevFlags: number, st: AnimState, cheering: boolean, index: number, action: { running: boolean }): Pick {
  const started = (f: number) => !!(flags & f) && !(prevFlags & f);
  const a = look.anims;
  if (flags & UF.dead) return { name: variant(a.death, index, 0), kind: 'death', rate: 1, restart: false };
  if (flags & UF.staggered) return { name: variant(a.hit, index, st.hits), kind: 'hit', rate: 1, restart: started(UF.staggered) };
  if (flags & UF.swinging) {
    const name = sk.find((s) => s.look.attack)?.look.attack ?? variant(a.attack, index, st.swings);
    return { name, kind: 'attack', rate: 1, restart: started(UF.swinging) };
  }
  if (st.action && (st.actionNew || action.running)) return { name: st.action.clip, kind: 'action', rate: st.action.rate, restart: st.actionNew };
  // legs: rates are set from the unit's ground speed (locomotion below)
  if (flags & UF.dashing) return { name: a.sprint ?? a.move, kind: a.sprint ? 'sprint' : 'move', rate: 1, restart: false };
  if (flags & UF.moving) return { name: a.move, kind: 'move', rate: 1, restart: false };
  if (cheering) return { name: a.cheer, kind: 'cheer', rate: 1, restart: false };
  return { name: a.idle, kind: 'idle', rate: 1, restart: false };
}

// Dev only: every clip start (window.animLog), for checking choices and timing.
const animLog: { tick: number; unit: number; clip: string; kind: string; rate: number; from: number }[] = [];
if (import.meta.env.DEV) Object.assign(window, { animLog });

type AnimState = {
  name: string; flags: number; deadAt: number; faded: boolean; aim: number; pitch: number;
  swings: number; hits: number; action?: { clip: string; rate: number }; actionNew: boolean;
  speed: number; // smoothed ground speed, m/s of sim time
};

const TICK = 0.05; // sim seconds per tick

// Locomotion (battle-presentation "Animation follows the sim"): walk and jog
// blended and kept in step so the stride covers the unit's ground speed v.
// Each clip covers speed·duration metres per cycle at 1× (scaled by the
// unit's size); a blend k mixes stride and cycle length linearly, and
// S(k)/D(k) = v has a closed-form k. Below walking speed the walk slows down
// (to 0.5×); above jogging speed the jog speeds up (to 1.4×).
export function locomotion(v: number, size: number, walk: { speed: number; duration: number }, jog: { speed: number; duration: number }) {
  const Sw = walk.speed * walk.duration * size, Sj = jog.speed * jog.duration * size, Dw = walk.duration, Dj = jog.duration;
  if (v <= Sw / Dw) return { k: 0, walkRate: Math.max(0.5, v / (Sw / Dw)), jogRate: 1 };
  if (v >= Sj / Dj) return { k: 1, walkRate: 1, jogRate: Math.min(1.4, v / (Sj / Dj)) };
  const k = (v * Dw - Sw) / ((Sj - Sw) - v * (Dj - Dw)), D = Dw + k * (Dj - Dw);
  return { k, walkRate: Dw / D, jogRate: Dj / D };
}

function Unit({ index }: { index: number }) {
  const g = useRef<Group>(null), bar = useRef<Group>(null), sink = useRef<Group>(null);
  const info = useBattle((s) => s.curr?.units[index]);
  const nth = useBattle((s) => s.curr?.units.slice(0, index).filter((u) => u.team === s.curr!.units[index].team && u.cls === s.curr!.units[index].cls).length ?? 0);
  // Red is hidden while deploying, and during the fight unless Blue's vision sees it
  const hidden = useBattle((s) => {
    const u = s.curr?.units[index];
    if (!u || u.team !== 1) return false;
    if (FOGGED.includes(s.phase)) return true;
    return s.fogOn && FOGGED_FIGHT.includes(s.phase) && !(u.flags & UF.seen);
  });
  const selected = useBattle((s) => s.selected === index);
  const draggable = useBattle((s) => info?.team === 0 && s.phase === 'deploying');
  const [ch, setCh] = useState<Character>();
  const anim = useRef<AnimState>({ name: '', flags: 0, deadAt: 0, faded: false, aim: 0, pitch: 0, swings: 0, hits: 0, actionNew: false, speed: 0 });

  const team = info?.team, cls = info?.cls;
  useEffect(() => {
    if (team === undefined || cls === undefined) return;
    let live = true;
    makeCharacter(cls, lookFor(cls, nth), team, nth).then((c) => { if (live) setCh(c); }); // capsule until loaded
    return () => { live = false; };
  }, [team, cls, nth]);

  useFrame((state, delta) => {
    const bs = useBattle.getState();
    const { prev, curr, dragging, speed, data, outcome, phase } = bs;
    const a = prev?.units[index], b = curr?.units[index];
    if (!a || !b || !g.current) return;
    if (dragging?.index === index) { // follows the pointer until the sim places it
      g.current.position.set(dragging.at.x, 0.25, dragging.at.z);
    } else {
      const k = alpha();
      g.current.position.set(lerp(a.x, b.x, k), b.flags & UF.wet && !(b.flags & UF.dead) ? -0.12 : 0, lerp(a.z, b.z, k));
      g.current.rotation.y = Math.atan2(lerp(a.fx, b.fx, k), lerp(a.fz, b.fz, k));
      if (phase === 'fighting') feet(index, g.current.position.x, g.current.position.z, b.flags, delta, viewOf(bs), b);
    }
    if (bar.current) bar.current.scale.x = Math.max(0.001, b.hp / b.maxHp);

    const st = anim.current;
    const t = state.clock.elapsedTime;
    if (b.flags & UF.dead) {
      st.deadAt ||= t;
      const p = Math.min(1, Math.max(0, (t - st.deadAt - FADE_AFTER) / FADE_FOR));
      if (p > 0 && ch && !st.faded) { // the unit's own materials (models.ts), so only this body fades
        for (const m of ch.mats) m.transparent = true;
        st.faded = true;
      }
      if (ch && st.faded) for (const m of ch.mats) m.opacity = 1 - p;
      if (sink.current) { sink.current.position.y = -0.35 * p; sink.current.visible = p < 1; }
    } else {
      st.deadAt = 0;
      if (sink.current) { sink.current.position.y = 0; sink.current.visible = true; }
    }
    if (!ch) return;
    const won = phase === 'done' && ((outcome?.result === BattleResult.BLUE_WIN && b.team === 0) || (outcome?.result === BattleResult.RED_WIN && b.team === 1));
    const look = lookOf(b.cls), learned = bs.units[index]?.skills ?? [], sk = skillLooks(b.cls, learned);
    if ((b.flags & UF.swinging) && !(st.flags & UF.swinging)) st.swings++;
    if ((b.flags & UF.staggered) && !(st.flags & UF.staggered)) st.hits++;
    const queued = skillActions.get(index); // a skill acted this tick (dash, blink): play its action
    if (queued) { skillActions.delete(index); st.action = queued; st.actionNew = true; }
    const p = choose(look, sk, b.flags, st.flags, st, won, index, { running: !!st.action && !!ch.actions[st.action.clip]?.isRunning() });
    if (p.kind !== 'action') st.action = undefined;
    st.actionNew = false;
    // ground speed from the sim's positions (Swift, water, frost and turns included), smoothed
    const ticks = (curr?.tick ?? 0) - (prev?.tick ?? 0);
    if (ticks > 0) {
      const raw = Math.hypot(b.x - a.x, b.z - a.z) / (ticks * TICK);
      st.speed += (raw - st.speed) * (1 - Math.exp(-(delta * speed) / 0.12));
    }
    const walkAct = ch.actions[look.anims.walk], jogAct = ch.actions[look.anims.move];
    const next = ch.actions[p.name] ?? ch.actions[look.anims.idle];
    if (next && (p.name !== st.name || p.restart)) {
      const cur = ch.actions[st.name];
      const fade = QUICK.has(p.kind) ? 0.12 : 0.25;
      next.reset();
      next.play();
      if (cur && cur !== next) cur.crossFadeTo(next, fade, false);
      // the walk rides along with the jog: in with it, out with it
      if (p.kind === 'move' && walkAct && walkAct !== next) { walkAct.reset(); walkAct.play(); walkAct.fadeIn(fade); }
      if (cur === jogAct && p.kind !== 'move' && walkAct?.isRunning()) walkAct.fadeOut(fade);
      st.name = p.name;
      if (p.kind === 'attack') {
        // time the clip so its blow lands with the sim's hit (unit-looks):
        // swing S (after attack-speed skills), blow at hitAt·S
        let swing = data?.classes[b.cls]?.swing ?? 1;
        for (const s of sk) for (const e of s.info.effects) if (e.kind === 'statBonus' && e.stat === 'attackSpeed') swing *= (e.pct ?? 100) / 100;
        const at = (data?.hitAt ?? 0.45) * swing, h = CLIPS[p.name]?.hit ?? next.getClip().duration * (data?.hitAt ?? 0.45);
        const rate = Math.min(1.8, Math.max(0.6, h / at));
        next.timeScale = rate;
        next.time = Math.max(0, h - rate * at); // when clamped, start later so the blow still lands on time
      }
      if (import.meta.env.DEV) animLog.push({ tick: b.flags & UF.dead ? -1 : (bs.curr?.tick ?? 0), unit: index, clip: p.name, kind: p.kind, rate: next.timeScale, from: next.time });
    }
    if (next && p.kind === 'move' && walkAct && jogAct) {
      const sz = ch.root.scale.x, back = b.flags & UF.kiting && !(b.flags & UF.dashing) ? -1 : 1; // kiting archers back away facing the threat
      const wc = walkAct.getClip(), jc = jogAct.getClip();
      const l = locomotion(st.speed, sz, { speed: CLIPS[wc.name]?.speed ?? 1, duration: wc.duration }, { speed: CLIPS[jc.name]?.speed ?? 5, duration: jc.duration });
      jogAct.weight = l.k;
      walkAct.weight = 1 - l.k;
      jogAct.timeScale = back * l.jogRate;
      walkAct.timeScale = back * l.walkRate;
      if (l.k > 0 && l.k < 1) walkAct.time = (jogAct.time / jc.duration) * wc.duration; // same phase: feet land together
    } else if (next && p.kind === 'sprint') {
      next.timeScale = Math.min(1.6, st.speed / ((CLIPS[next.getClip().name]?.speed ?? 8) * ch.root.scale.x));
    } else if (next && p.kind !== 'attack') {
      next.timeScale = p.rate;
    }
    if (p.kind !== 'move') { if (jogAct) jogAct.weight = 1; if (walkAct && !walkAct.isRunning()) walkAct.weight = 1; }
    // arcing shots: lean back while attacking
    const wantPitch = p.kind === 'attack' ? sk.find((s) => s.look.pitch)?.look.pitch ?? 0 : 0;
    st.pitch += (wantPitch - st.pitch) * (1 - Math.exp(-10 * delta * speed));
    st.flags = b.flags;
    ch.mixer.update(delta * speed);
    ch.ground(p.kind === 'move' || p.kind === 'sprint', delta * speed); // feet on the floor while walking
    ch.tick?.(delta * speed); // capes follow the posed body

    // team rim, gold when selected
    const sel = bs.selected === index;
    ch.rim.color.value.copy(sel ? GOLD : TEAM_RIM[b.team]);
    ch.rim.k.value = sel ? 0.9 + 0.3 * Math.sin(t * 4) : 0.55;

    if (b.flags & UF.dead) return;
    // ranged units turn their upper body toward the target (after the animation set the pose)
    const tgt = b.target >= 0 ? curr?.units[b.target] : undefined;
    let want = 0;
    if (data?.classes[b.cls]?.ranged && tgt && !(tgt.flags & UF.dead)) {
      const face = Math.atan2(b.fx, b.fz), to = Math.atan2(tgt.x - b.x, tgt.z - b.z);
      want = Math.max(-AIM_MAX, Math.min(AIM_MAX, Math.atan2(Math.sin(to - face), Math.cos(to - face))));
    }
    st.aim += (want - st.aim) * (1 - Math.exp(-8 * delta));
    if (ch.spine2 && Math.abs(st.aim) > 0.01) ch.spine2.quaternion.premultiply(tmpQ.setFromAxisAngle(localUp(ch.spine2), st.aim));
    if (ch.spine2 && st.pitch > 0.01) ch.spine2.quaternion.premultiply(tmpQ.setFromAxisAngle(toBone(ch.spine2, RIGHT), -st.pitch));
    // a short flinch away from the attacker
    const f = flinches.get(index);
    if (f && ch.spine1) {
      const k = (performance.now() / 1000 - f.t) / FLINCH_FOR;
      if (k >= 1) flinches.delete(index);
      else if (k >= 0) {
        const amt = FLINCH_MAX * Math.sin(Math.PI * Math.min(1, k * 1.6)) * (1 - k);
        // tilt about the horizontal axis perpendicular to the hit, in the unit's own frame
        const yaw = g.current.rotation.y;
        tmpAxis.set(f.dz, 0, -f.dx).applyEuler(tmpE.set(0, -yaw, 0));
        ch.spine1.quaternion.premultiply(tmpQ.setFromAxisAngle(toBone(ch.spine1, tmpAxis), amt));
      }
    }
  });

  if (!info) return null;
  const dead = !!(info.flags & UF.dead);
  const grab = (e: ThreeEvent<PointerEvent>) => {
    if (!draggable) return;
    e.stopPropagation();
    useBattle.getState().beginDrag(index, { x: e.point.x, z: e.point.z });
  };
  const pick = (e: ThreeEvent<MouseEvent>) => {
    if (draggable || hidden) return;
    e.stopPropagation();
    useBattle.getState().select(index);
  };
  // hidden units stay mounted (their animation keeps its state), just invisible
  return (
    <group ref={g} visible={!hidden} onPointerDown={grab} onClick={pick}
      onPointerOver={() => { if (draggable) document.body.style.cursor = 'grab'; }}
      onPointerOut={() => { document.body.style.cursor = ''; }}>
      <group ref={sink}>
        {ch ? <primitive object={ch.root} /> : (
          <mesh position-y={0.8} castShadow>
            <capsuleGeometry args={[0.32, 0.9]} />
            <meshStandardMaterial color={TEAM[info.team]} transparent opacity={0.6} />
          </mesh>
        )}
      </group>
      {selected && !hidden && (
        <mesh position-y={0.04} rotation-x={-Math.PI / 2}><ringGeometry args={[0.55, 0.66, 32]} /><meshBasicMaterial color="#ffd23f" /></mesh>
      )}
      {!dead && !hidden && <>
        {/* drawn over the ground, grass and water, like a selection circle */}
        <mesh position-y={0.06} rotation-x={-Math.PI / 2} renderOrder={6}>
          <ringGeometry args={[0.4, 0.5, 28]} />
          <meshBasicMaterial color={TEAM[info.team]} transparent opacity={0.85} depthWrite={false} depthTest={false} />
        </mesh>
        <Billboard position-y={2.15}>
          {/* always on top: drawn last, never hidden by trees, grass or units */}
          <mesh renderOrder={20}><planeGeometry args={[0.96, 0.16]} /><meshBasicMaterial color="#fff" transparent depthTest={false} depthWrite={false} /></mesh>
          <mesh position-z={0.001} renderOrder={21}><planeGeometry args={[0.92, 0.12]} /><meshBasicMaterial color="#1a1a1a" transparent depthTest={false} depthWrite={false} /></mesh>
          <group ref={bar} position={[-0.45, 0, 0.002]}>{/* pinned at the left edge, scaled by health */}
            <mesh position-x={0.45} renderOrder={22}><planeGeometry args={[0.9, 0.1]} /><meshBasicMaterial color={TEAM[info.team]} transparent depthTest={false} depthWrite={false} /></mesh>
          </group>
        </Billboard>
        {info.skills > 0 && (
          <Html position={[0, 2.45, 0]} center style={{ pointerEvents: 'none' }}>
            <div className="stars">{'★'.repeat(info.skills)}</div>
          </Html>
        )}
      </>}
      {!!(info.flags & UF.slowed) && !dead && (
        <mesh position-y={0.05} rotation-x={-Math.PI / 2}><ringGeometry args={[0.52, 0.62, 24]} /><meshBasicMaterial color="#9fd8ff" /></mesh>
      )}
    </group>
  );
}

// Invisible ground that catches pointer moves while a unit is being dragged.
function DragPlane() {
  const dragging = useBattle((s) => !!s.dragging);
  if (!dragging) return null;
  return (
    <mesh rotation-x={-Math.PI / 2} position-y={0.01}
      onPointerMove={(e) => useBattle.getState().dragTo({ x: e.point.x, z: e.point.z })}
      onPointerUp={() => useBattle.getState().endDrag()}>
      <planeGeometry args={[80, 80]} />
      <meshBasicMaterial visible={false} />
    </mesh>
  );
}

export function Units() {
  const n = useBattle((s) => s.curr?.units.length ?? 0);
  const battleId = useBattle((s) => s.battle?.id); // fresh units (and characters) per battle
  return <>{Array.from({ length: n }, (_, i) => <Unit key={`${battleId}-${i}`} index={i} />)}<DragPlane /></>;
}
