// What the store asks of the 3D scene, registered by the scene when it mounts
// (scene/BattleView.tsx), so the store, part of the startup bundle, needs no
// three.js. With no scene mounted there is nothing to show, and the store
// skips these.
import type { SimEvent, Snapshot } from '../sim/snapshot.ts';

export const sceneHooks: {
  onSnapshot?: (s: Snapshot, events: SimEvent[]) => void; // effects for a new snapshot
  lastKill?: () => void; // the battle's final blow
} = {};
