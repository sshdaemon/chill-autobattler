import { useBattle } from '../store.ts';
import type { Progress } from '../assets.ts';

// The loading screen (openspec change asset-loading): at start while the sim
// downloads, and before deployment while the battle's assets do. With chunk
// set it stands in for the 3D view's code (App's Suspense fallback).
const mb = (n: number) => (n / 1048576).toFixed(1);

function Bar({ title, what, p }: { title: string; what: string; p?: Progress }) {
  const share = p && p.total > 0 ? Math.min(1, p.done / p.total) : undefined;
  return (
    <div className="loading" role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={share === undefined ? undefined : Math.round(share * 100)}>
      <b>{title}</b>
      <div className="lbar"><i style={{ width: share === undefined ? '15%' : `${share * 100}%` }} className={share === undefined ? 'busy' : undefined} /></div>
      <span>{what}{p && p.total > 0 ? ` · ${mb(p.done)} of ${mb(p.total)} MB` : ''}</span>
    </div>
  );
}

export function Loading({ chunk = false }: { chunk?: boolean }) {
  const phase = useBattle((s) => s.phase);
  const sim = useBattle((s) => s.simLoad);
  const battle = useBattle((s) => s.battleLoad?.progress);
  if (chunk) return <Bar title="Preparing the battlefield" what="Loading the battlefield" />;
  if (phase === 'loading') return <Bar title="Chill Autobattler" what="Loading the game" p={sim} />;
  if (phase === 'preparing') return <Bar title="Preparing the battlefield" what="Loading the armies and the field" p={battle} />;
  return null;
}
