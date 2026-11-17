import { useEffect, useRef, useState } from 'react';
import { useFrame, type ThreeEvent } from '@react-three/fiber';
import { Billboard, Html } from '@react-three/drei';
import { LoopRepeat, type Group } from 'three';
import { useBattle } from '../store.ts';
import { UF } from '../sim/snapshot.ts';
import { BattleResult } from '../gen/chill/v1/battle_pb.ts';
import { alpha, lerp } from './interp.ts';
import { lookFor, makeCharacter, type Character } from './models.ts';
import { FOGGED_FIGHT } from './Fog.tsx';
import { feet, flinches, viewOf } from './effects.ts';
import { TEAM_RIM } from './models.ts';
import { Euler, Quaternion, Vector3 } from 'three';

const GOLD = TEAM_RIM[0].clone().setStyle('#ffd23f');
const UP = new Vector3(0, 1, 0), tmpQ = new Quaternion(), tmpAxis = new Vector3(), tmpE = new Euler();
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

// Which animation the sim state asks for, and at what speed.
function wanted(flags: number, cls: number, cheering: boolean): [string, number] {
  if (flags & UF.dead) return ['Death01', 1];
  if (flags & UF.staggered) return ['Hit_Chest', 1];
  if (flags & UF.swinging) return [cls === 0 ? 'Sword_Attack' : 'Spell_Simple_Shoot', 1];
  const slow = (flags & UF.wet ? 0.7 : 1) * (flags & UF.slowed ? 0.6 : 1); // legs slow down too
  if (flags & UF.dashing) return ['Jog_Fwd_Loop', 2.2];
  if (flags & UF.moving) return ['Jog_Fwd_Loop', slow];
  if (cheering) return ['Yes', 1];
  return ['Sword_Idle', 1];
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
  const anim = useRef({ name: 'Sword_Idle', flags: 0, deadAt: 0, faded: false, aim: 0 });

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
    const [name, rate] = wanted(b.flags, b.cls, won);
    const started = (flag: number) => (b.flags & flag) && !(st.flags & flag); // rising edge: replay a one-shot
    const next = ch.actions[name];
    if (name !== st.name || (name === 'Hit_Chest' && started(UF.staggered)) || (name !== 'Jog_Fwd_Loop' && name !== 'Sword_Idle' && started(UF.swinging))) {
      const cur = ch.actions[st.name];
      next.reset();
      if (name === 'Yes') next.setLoop(LoopRepeat, Infinity);
      next.play();
      // quick into attacks, hits and death; smooth between moving and standing
      if (cur && cur !== next) cur.crossFadeTo(next, ['Death01', 'Hit_Chest', 'Sword_Attack', 'Spell_Simple_Shoot'].includes(name) ? 0.12 : 0.25, false);
      st.name = name;
    }
    if (name === 'Sword_Attack' || name === 'Spell_Simple_Shoot') {
      // stretch the attack to the sim's swing time, so the blow lines up with the hit
      next.timeScale = ch.clips[name].duration / (data?.classes[b.cls].swing ?? 1);
    } else {
      // kiting archers back away facing the threat: play the jog backwards
      next.timeScale = name === 'Jog_Fwd_Loop' && b.flags & UF.kiting ? -rate : rate;
    }
    st.flags = b.flags;
    ch.mixer.update(delta * speed);

    // team rim, gold when selected
    const sel = bs.selected === index;
    ch.rim.color.value.copy(sel ? GOLD : TEAM_RIM[b.team]);
    ch.rim.k.value = sel ? 0.9 + 0.3 * Math.sin(t * 4) : 0.55;

    if (b.flags & UF.dead) return;
    // archers and mages turn their upper body toward the target (after the animation set the pose)
    const tgt = b.target >= 0 ? curr?.units[b.target] : undefined;
    let want = 0;
    if (b.cls !== 0 && tgt && !(tgt.flags & UF.dead)) {
      const face = Math.atan2(b.fx, b.fz), to = Math.atan2(tgt.x - b.x, tgt.z - b.z);
      want = Math.max(-AIM_MAX, Math.min(AIM_MAX, Math.atan2(Math.sin(to - face), Math.cos(to - face))));
    }
    st.aim += (want - st.aim) * (1 - Math.exp(-8 * delta));
    if (ch.spine2 && Math.abs(st.aim) > 0.01) ch.spine2.quaternion.premultiply(tmpQ.setFromAxisAngle(localUp(ch.spine2), st.aim));
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
