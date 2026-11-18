// Simulated cloth for capes (presentation only). A grid of points hangs from
// the shoulders: the top row follows the spine bone, the rest is Verlet
// integration with gravity and damping, distance constraints (structure,
// shear, bending) and collisions with capsules around the torso and legs, so
// the cape swings with the unit's motion and drapes over its back instead of
// passing through it. Simulated in world space, drawn in the character's space.
import * as THREE from 'three';

export type ClothSpec = { bone: string; width: [number, number]; length: number; team?: boolean };

const COLS = 7, ROWS = 10;
const STEP = 1 / 60, MAX_STEPS = 4, ITER = 4;
const GRAVITY = -9.8, DAMP = 0.99, DRAG = 2.2; // drag: per second, towards the air's velocity

type Capsule = { a: THREE.Bone; b: THREE.Bone; r: number };

// Builds a cape for a character. bones: the character's bones by name; girth:
// how much the build pushes the torso and thighs out (metres), so the colliders
// fit the body. Returns its mesh (child of root) and a tick to call after the
// animation has posed the bones.
export function makeCape(root: THREE.Object3D, bones: Record<string, THREE.Bone>, spec: ClothSpec, color: THREE.Color, girth: (bone: string) => number) {
  const anchor = bones[spec.bone];
  const n = COLS * ROWS;
  // rest shape in the character's space (T-pose rest: faces +Z, up +Y)
  root.updateMatrixWorld(true);
  const toRoot = new THREE.Matrix4().copy(root.matrixWorld).invert();
  const top = anchor.getWorldPosition(new THREE.Vector3()).applyMatrix4(toRoot);
  const backZ = top.z - 0.12 - girth('spine_03');
  const rest: THREE.Vector3[] = [];
  for (let r = 0; r < ROWS; r++) {
    const t = r / (ROWS - 1), w = spec.width[0] + (spec.width[1] - spec.width[0]) * t;
    for (let c = 0; c < COLS; c++) {
      const u = c / (COLS - 1) - 0.5;
      // the top edge curves round the shoulders, the hem hangs straight
      rest.push(new THREE.Vector3(u * w, top.y + 0.1 - spec.length * t, backZ - 0.03 * t + (1 - t) * 0.09 * (2 * u) ** 2));
    }
  }
  // anchors: the top row, held in the spine bone's space
  const anchorLocal = rest.slice(0, COLS).map((p) => anchor.worldToLocal(p.clone().applyMatrix4(root.matrixWorld)));

  // constraints between point indices, with rest lengths
  const links: [number, number, number][] = [];
  const at = (r: number, c: number) => r * COLS + c;
  const link = (i: number, j: number) => links.push([i, j, rest[i].distanceTo(rest[j])]);
  for (let r = 0; r < ROWS; r++) for (let c = 0; c < COLS; c++) {
    if (c + 1 < COLS) link(at(r, c), at(r, c + 1));
    if (r + 1 < ROWS) link(at(r, c), at(r + 1, c));
    if (r + 1 < ROWS && c + 1 < COLS) { link(at(r, c), at(r + 1, c + 1)); link(at(r, c + 1), at(r + 1, c)); }
    if (r + 2 < ROWS) link(at(r, c), at(r + 2, c)); // bending: keeps it from folding like paper
  }
  const b = (name: string) => bones[name];
  const capsules: Capsule[] = [
    { a: b('spine_03'), b: b('pelvis'), r: 0.17 + girth('spine_03') },
    { a: b('pelvis'), b: b('pelvis'), r: 0.17 + girth('pelvis') },
    ...(['l', 'r'] as const).flatMap((s) => [
      { a: b(`thigh_${s}`), b: b(`calf_${s}`), r: 0.1 + girth('thigh') },
      { a: b(`calf_${s}`), b: b(`foot_${s}`), r: 0.07 + girth('calf') },
    ]),
  ].filter((k) => k.a && k.b);

  // geometry: positions rewritten each frame, colours mark a darker hem and collar
  const geo = new THREE.BufferGeometry();
  const pos = new Float32Array(n * 3), uv = new Float32Array(n * 2), col = new Float32Array(n * 3);
  for (let r = 0; r < ROWS; r++) for (let c = 0; c < COLS; c++) {
    const i = at(r, c), t = r / (ROWS - 1);
    uv[i * 2] = c / (COLS - 1); uv[i * 2 + 1] = 1 - t;
    const shade = t > 0.88 ? 0.45 : r === 0 ? 0.6 : 0.92 + 0.08 * Math.sin(c * 1.7); // trim, collar, slight weave
    col.set([shade, shade, shade], i * 3);
  }
  const index: number[] = [];
  for (let r = 0; r + 1 < ROWS; r++) for (let c = 0; c + 1 < COLS; c++) index.push(at(r, c), at(r + 1, c), at(r, c + 1), at(r, c + 1), at(r + 1, c), at(r + 1, c + 1));
  geo.setIndex(index);
  geo.setAttribute('position', new THREE.BufferAttribute(pos, 3).setUsage(THREE.DynamicDrawUsage));
  geo.setAttribute('uv', new THREE.BufferAttribute(uv, 2));
  geo.setAttribute('color', new THREE.BufferAttribute(col, 3));
  const mat = new THREE.MeshStandardMaterial({ color, vertexColors: true, roughness: 0.85, side: THREE.DoubleSide, name: 'MI_Cape' });
  const mesh = new THREE.Mesh(geo, mat);
  mesh.castShadow = true;
  mesh.frustumCulled = false;
  root.add(mesh);

  const p = rest.map(() => new THREE.Vector3()), prev = rest.map(() => new THREE.Vector3());
  let ready = false, carry = 0;
  const lastTop = new THREE.Vector3();
  const A = new THREE.Vector3(), B = new THREE.Vector3(), AB = new THREE.Vector3(), Q = new THREE.Vector3(), D = new THREE.Vector3();

  function reset() { // hang straight from the current pose
    root.updateMatrixWorld(true);
    rest.forEach((r, i) => { p[i].copy(r).applyMatrix4(root.matrixWorld); prev[i].copy(p[i]); });
    ready = true;
  }

  const air = new THREE.Vector3(); // wind, m/s (0 in battle: moving through still air already drags the cape back)
  function step(h: number) {
    for (let c = 0; c < COLS; c++) { prev[c].copy(p[c]); anchor.localToWorld(p[c].copy(anchorLocal[c])); }
    const k = Math.min(1, DRAG * h);
    for (let i = COLS; i < n; i++) {
      const x = p[i], v = D.copy(x).sub(prev[i]).multiplyScalar(DAMP); // displacement this step
      v.lerp(Q.copy(air).multiplyScalar(h), k); // air drag
      prev[i].copy(x);
      x.add(v);
      x.y += GRAVITY * h * h;
    }
    for (let k = 0; k < ITER; k++) {
      for (const [i, j, len] of links) {
        const d = D.copy(p[j]).sub(p[i]), l = d.length() || 1e-6, f = (l - len) / l;
        const wi = i < COLS ? 0 : 1, wj = j < COLS ? 0 : 1, w = wi + wj;
        if (!w) continue;
        if (wi) p[i].addScaledVector(d, (f * wi) / w);
        if (wj) p[j].addScaledVector(d, (-f * wj) / w);
      }
      for (const cap of capsules) {
        cap.a.getWorldPosition(A); cap.b.getWorldPosition(B); AB.copy(B).sub(A);
        const len2 = AB.lengthSq();
        for (let i = COLS; i < n; i++) {
          const t = len2 > 1e-8 ? Math.min(1, Math.max(0, D.copy(p[i]).sub(A).dot(AB) / len2)) : 0;
          Q.copy(A).addScaledVector(AB, t);
          const d = D.copy(p[i]).sub(Q), l = d.length();
          if (l < cap.r && l > 1e-6) p[i].addScaledVector(d, (cap.r - l) / l);
        }
      }
    }
  }

  // dt: seconds of game time since the last call (0 just re-poses); wind: m/s
  function tick(dt: number, wind?: THREE.Vector3) {
    air.set(0, 0, 0);
    if (wind) air.copy(wind);
    root.updateMatrixWorld(true);
    anchor.localToWorld(D.copy(anchorLocal[(COLS - 1) >> 1]));
    if (!ready || D.distanceTo(lastTop) > 0.8) reset(); // first frame, a teleport (Blink) or a drag
    lastTop.copy(D);
    carry = Math.min(carry + Math.max(0, dt), STEP * MAX_STEPS);
    while (carry >= STEP) { step(STEP); carry -= STEP; }
    if (dt === 0) for (let c = 0; c < COLS; c++) anchor.localToWorld(p[c].copy(anchorLocal[c]));
    const inv = toRoot.copy(root.matrixWorld).invert();
    for (let i = 0; i < n; i++) { Q.copy(p[i]).applyMatrix4(inv); pos[i * 3] = Q.x; pos[i * 3 + 1] = Q.y; pos[i * 3 + 2] = Q.z; }
    geo.attributes.position.needsUpdate = true;
    geo.computeVertexNormals();
  }

  // settles the cape over the current pose (for still renders: preview, portraits)
  function settle(seconds = 1) { reset(); for (let t = 0; t < seconds; t += STEP) step(STEP); tick(0); }

  return { mesh, mat, tick, settle };
}
