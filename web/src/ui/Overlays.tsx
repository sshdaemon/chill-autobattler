import { useBattle, resultText } from '../store.ts';
import { UF } from '../sim/snapshot.ts';
import { CLASS_ICON } from '../gamedata.ts';

export function Banner() {
  const phase = useBattle((s) => s.phase);
  if (phase === 'starting') return <div className="banner small">Preparing the battlefield…</div>;
  if (phase === 'countdown') return <div className="banner">Fight!</div>;
  return null;
}

// Top bar during deployment and the fight.
export function Status() {
  const { curr, phase } = useBattle();
  if (!curr || phase === 'picker') return null;
  const alive = (team: number) => curr.units.filter((u) => u.team === team && !(u.flags & UF.dead)).length;
  return (
    <div className="status">
      <span className="b">Blue {alive(0)}</span> vs <span className="r">{phase === 'deploying' || phase === 'locking' ? '?' : alive(1)} Red</span>
      {curr.tick > 0 && <span className="dim"> · {(curr.tick / 20).toFixed(1)} s</span>}
    </div>
  );
}

export function Result() {
  const { outcome, curr, verdict, verify, nextMatch } = useBattle();
  if (!outcome || !curr) return null;
  const text = resultText(outcome.result);
  const survivors = (team: number) => curr.units.filter((u) => u.team === team && !(u.flags & UF.dead));
  return (
    <div className="overlay clear">
      <div className={`panel result ${text.toLowerCase()}`}>
        <h1>{text}</h1>
        <div className="sub">after {(outcome.ticks / 20).toFixed(1)} s</div>
        <div className="surv">
          <div><span className="b">Blue</span> {survivors(0).map((u, i) => <span key={i} title={`${Math.ceil(u.hp)}/${u.maxHp}`}>{CLASS_ICON[u.cls]}</span>)}{survivors(0).length === 0 && '—'}</div>
          <div><span className="r">Red</span> {survivors(1).map((u, i) => <span key={i} title={`${Math.ceil(u.hp)}/${u.maxHp}`}>{CLASS_ICON[u.cls]}</span>)}{survivors(1).length === 0 && '—'}</div>
        </div>
        {verdict && (
          <div className={`verdict ${verdict.state}`} data-testid="verdict">
            {verdict.state === 'fair' ? '✅ ' : verdict.state === 'unfair' ? '❌ ' : verdict.state === 'error' ? '⚠️ ' : '⏳ '}{verdict.text}
            {verdict.state === 'error' && <button onClick={verify}>Retry</button>}
          </div>
        )}
        <button className="go" onClick={nextMatch}>Next match →</button>
      </div>
    </div>
  );
}

// Debug controls: playback speed and Red's skill count (client-chosen until runs exist).
export function DevPanel() {
  const { speed, setSpeed, redSkills, setRedSkills, version } = useBattle();
  return (
    <div className="dev">
      <span className="dim">dev · sim {version ?? '…'}</span>
      {[1, 2, 4].map((s) => <button key={s} className={speed === s ? 'on' : ''} onClick={() => setSpeed(s)}>{s}×</button>)}
      <label title="Red's total skills for the next match">Red ★ <input type="number" min={0} max={25} value={redSkills} onChange={(e) => setRedSkills(+e.target.value)} /></label>
    </div>
  );
}
