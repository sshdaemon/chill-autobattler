import { useBattle, resultText } from '../store.ts';
import { UF } from '../sim/snapshot.ts';
import { CLASS_ICON, CLASS_LABEL } from '../gamedata.ts';
import { RunState } from '../gen/chill/v1/runstate_pb.ts';

export function Banner() {
  const phase = useBattle((s) => s.phase);
  if (phase === 'loading') return <div className="banner small">Loading…</div>;
  if (phase === 'starting') return <div className="banner small">Preparing the battlefield…</div>;
  if (phase === 'countdown') return <div className="banner">Fight!</div>;
  return null;
}

// Run name · match · wins · points · Red's skills.
export function RunBar() {
  const run = useBattle((s) => s.run);
  // the server eases a run's first matches; shown only while that battle is on
  const eased = useBattle((s) => (['deploying', 'locking', 'countdown', 'fighting'].includes(s.phase) ? s.battle?.redHealthPct ?? 0 : 0));
  if (!run) return null;
  return (
    <div className="runbar">
      <b>{run.name}</b> · match {run.match} · {run.wins} win{run.wins === 1 ? '' : 's'} · <b>{run.score}</b> pts
      {run.redSkills > 0 && <span className="r"> · Red ★×{run.redSkills}</span>}
      {eased > 0 && <span className="dim" title="Red is weaker in a run's first matches"> · Red at {eased}% strength</span>}
    </div>
  );
}

// Unit counts during deployment and the fight.
export function Status() {
  const { curr, phase } = useBattle();
  if (!curr || phase === 'picker' || phase === 'loading') return null;
  const alive = (team: number) => curr.units.filter((u) => u.team === team && !(u.flags & UF.dead)).length;
  return (
    <div className="status">
      <span className="b">Blue {alive(0)}</span> vs <span className="r">{phase === 'deploying' || phase === 'locking' ? '?' : alive(1)} Red</span>
      {curr.tick > 0 && <span className="dim"> · {(curr.tick / 20).toFixed(1)} s</span>}
    </div>
  );
}

function SkillPicks() {
  const { run, data, picks, choosePick, confirmPicks } = useBattle();
  if (!run || !data) return null;
  const done = run.owedPicks.every((v) => picks[v] !== undefined);
  return (
    <div className="picks">
      <h3>Your survivors learn a skill</h3>
      {run.owedPicks.map((v) => {
        const vet = run.veterans[v], cls = vet.unitClass - 1;
        const known = data.classes[cls].skills.filter((s) => vet.skills.includes(s.id)).map((s) => s.name);
        return (
          <div className="pickrow" key={v}>
            <span className="who" title={known.length ? `Knows: ${known.join(', ')}` : 'No skills yet'}>
              {CLASS_ICON[cls]} {CLASS_LABEL[cls]} <i>{'★'.repeat(vet.skills.length)}</i>
            </span>
            {data.classes[cls].skills.filter((s) => !vet.skills.includes(s.id)).map((s) => (
              <button key={s.id} className={picks[v] === s.id ? 'on' : ''} onClick={() => choosePick(v, s.id)}
                title={s.pct ? `${s.name}: ${s.pct}%` : s.name}>{s.name}</button>
            ))}
          </div>
        );
      })}
      <button className="go" disabled={!done} onClick={confirmPicks}>{done ? 'Learn skills' : 'Pick one skill for each survivor'}</button>
    </div>
  );
}

function Scoreboard() {
  const board = useBattle((s) => s.board);
  if (!board) return <div className="dim">Loading the scoreboard…</div>;
  return (
    <table className="board">
      <thead><tr><th>#</th><th>Run</th><th>Wins</th><th>Score</th></tr></thead>
      <tbody>
        {board.top.map((r) => (
          <tr key={r.rank} className={r.mine ? 'mine' : ''}><td>{r.rank}</td><td>{r.name}</td><td>{r.wins}</td><td>{r.score}</td></tr>
        ))}
      </tbody>
      {board.myBestRank > 10 && <tfoot><tr className="mine"><td>{board.myBestRank}</td><td>your best</td><td /><td>{board.myBestScore}</td></tr></tfoot>}
    </table>
  );
}

const REASON: Record<string, string> = {
  defeat: 'Defeated', draw: 'A draw ends the run', 'not fair': 'The server could not confirm the result',
  abandoned: 'The battle was abandoned', restarted: 'Restarted',
};

export function Result() {
  const { outcome, curr, verdict, verify, nextMatch, run, error } = useBattle();
  if (!run) return null;
  const text = outcome ? resultText(outcome.result) : run.state === RunState.OVER ? 'Run over' : 'Victory';
  const survivors = (team: number) => curr?.units.filter((u) => u.team === team && !(u.flags & UF.dead)) ?? [];
  const over = run.state === RunState.OVER;
  return (
    <div className="overlay clear">
      <div className={`panel result ${text.toLowerCase().replace(' ', '-')}`}>
        <h1>{text}</h1>
        {outcome && curr && <>
          <div className="sub">after {(outcome.ticks / 20).toFixed(1)} s</div>
          <div className="surv">
            <div><span className="b">Blue</span> {survivors(0).map((u, i) => <span key={i}>{CLASS_ICON[u.cls]}</span>)}{survivors(0).length === 0 && '—'}</div>
            <div><span className="r">Red</span> {survivors(1).map((u, i) => <span key={i}>{CLASS_ICON[u.cls]}</span>)}{survivors(1).length === 0 && '—'}</div>
          </div>
        </>}
        {verdict && (
          <div className={`verdict ${verdict.state}`} data-testid="verdict">
            {verdict.state === 'fair' ? '✅ ' : verdict.state === 'unfair' ? '❌ ' : verdict.state === 'error' ? '⚠️ ' : '⏳ '}{verdict.text}
            {verdict.state === 'error' && <button onClick={verify}>Retry</button>}
          </div>
        )}
        {(!verdict || verdict.state !== 'checking') && run.state === RunState.PICKING && <SkillPicks />}
        {run.state === RunState.READY && <>
          <div className="sub"><b>{run.score}</b> points · {run.wins} win{run.wins === 1 ? '' : 's'}</div>
          <button className="go" onClick={nextMatch}>Next match →</button>
        </>}
        {over && <>
          <div className="runover">
            <b>{run.name}</b>: {REASON[run.endReason] ?? run.endReason} after {run.wins} win{run.wins === 1 ? '' : 's'} · <b>{run.score}</b> points
          </div>
          <Scoreboard />
          <button className="go" onClick={nextMatch}>New run →</button>
        </>}
        {error && <div className="err">{error}</div>}
      </div>
    </div>
  );
}

// Debug controls: playback speed.
export function DevPanel() {
  const { speed, setSpeed, version } = useBattle();
  return (
    <div className="dev">
      <span className="dim">dev · sim {version ?? '…'}</span>
      {[1, 2, 4].map((s) => <button key={s} className={speed === s ? 'on' : ''} onClick={() => setSpeed(s)}>{s}×</button>)}
    </div>
  );
}
