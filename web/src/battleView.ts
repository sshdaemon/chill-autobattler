// The 3D battle view's module (scene/BattleView.tsx), loaded once and shared:
// the store loads it while preparing a battle, and App renders it directly
// once it's here, so deployment doesn't flash a loading screen the way
// React.lazy's first render would.
type BattleViewModule = typeof import('./scene/BattleView.tsx');

let loading: Promise<BattleViewModule> | undefined;
export let battleView: BattleViewModule | undefined;

export const loadBattleView = () => (loading ??= import('./scene/BattleView.tsx').then((m) => (battleView = m)));
