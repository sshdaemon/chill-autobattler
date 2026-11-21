import { lazy, Suspense } from 'react';
import { battleView, loadBattleView } from './battleView.ts';
import { useBattle, BATTLE_PHASES } from './store.ts';
import { DraftScreen } from './ui/Draft.tsx';
import { FactionPick } from './ui/FactionPick.tsx';
import { DeployBar } from './ui/DeployBar.tsx';
import { Loading } from './ui/Loading.tsx';
import { Banner, DevPanel, Result, RunBar, Status } from './ui/Overlays.tsx';

// The 3D view loads only for battles (openspec change asset-loading); faction
// pick and the draft sit on a light 2D backdrop.
const LazyBattleView = lazy(loadBattleView);

export function App() {
  const phase = useBattle((s) => s.phase);
  const battle = BATTLE_PHASES.includes(phase);
  return (
    <>
      {!battle ? <div className="backdrop" />
        : battleView ? <battleView.default />
          : <Suspense fallback={<Loading chunk />}><LazyBattleView /></Suspense>}
      <RunBar />
      <Status />
      <Banner />
      <Loading />
      {phase === 'faction' && <FactionPick />}
      {phase === 'drafting' && <DraftScreen />}
      {(phase === 'deploying' || phase === 'locking') && <DeployBar />}
      {phase === 'done' && <Result />}
      <DevPanel />
    </>
  );
}
