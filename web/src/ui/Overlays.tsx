import { classOf } from '../gamedata.ts';
import { useBattle, resultText } from '../store.ts';
import { UF } from '../sim/snapshot.ts';
import { uiOf } from '../units/looks.ts';
import { isTuned } from '../scene/tuning.ts';
import { RunState } from '../gen/chill/v1/runstate_pb.ts';
import { useState } from 'react';
import { setSound, soundPrefs } from '../sound.ts';
import { useEffect } from 'react';
import { fps } from '../scene/perfStats.ts';
import { shownPlan } from '../plans.ts';

export function Banner() {
  const phase = useBattle((s) => s.phase);
  const error = useBattle((s) => s.error);
  const red = useBattle((s) => shownPlan(s, 1));
  // the sim can refuse a battle (an invalid input): say so instead of a frozen field
  if (error && (phase === 'starting' || phase === 'countdown' || phase === 'fighting')) return <div className="banner small err">{error}</div>;
  if (phase === 'starting') return <div className="banner small">Starting the battle…</div>;
  if (phase === 'countdown') return <div className="banner">Fight!{red && <small className="plan">Red: {red}</small>}</div>;
  return null;
}

// Run name · match · wins · points · Red's skills.
export function RunBar() {
  const run = useBattle((s) => s.run);
  const phase = useBattle((s) => s.phase);
  const surrender = useBattle((s) => s.surrender);
  // the server eases a run's first matches; shown only while that battle is on
  const eased = useBattle((s) => (['deploying', 'locking', 'countdown', 'fighting'].includes(s.phase) ? s.battle?.redHealthPct ?? 0 : 0));
  if (!run) return null;
  return (
    <div className="runbar">
      <b>{run.name}</b> · match {run.match} · {run.wins} win{run.wins === 1 ? '' : 's'} · <b>{run.score}</b> pts
      {run.redSkills > 0 && <span className="r"> · Red ★×{run.redSkills}</span>}
      {eased > 0 && <span className="dim" title="Red is weaker in a run's first matches"> · Red at {eased}% strength</span>}
      {run.state !== RunState.OVER && phase !== 'faction' && phase !== 'loading' && (
        <button className="surrender" title="End this run (its score stays on the board) and start a new one"
          onClick={() => { if (window.confirm('Surrender this run and start a new one? Its score stays on the board.')) void surrender(); }}>Surrender</button>
      )}
    </div>
  );
}

// Unit counts during deployment and the fight.
export function Status() {
  const { curr, phase } = useBattle();
  const bluePlan = useBattle((s) => shownPlan(s, 0));
  const redPlan = useBattle((s) => shownPlan(s, 1));
  if (!curr || phase === 'drafting' || phase === 'faction' || phase === 'loading') return null;
  const alive = (team: number) => curr.units.filter((u) => u.team === team && !(u.flags & UF.dead)).length;
  return (
    <div className="status">
      <span className="b">Blue {alive(0)}{bluePlan && <i> ({bluePlan})</i>}</span> vs <span className="r">{phase === 'deploying' || phase === 'locking' ? '?' : alive(1)} Red{redPlan && <i> ({redPlan})</i>}</span>
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
        const known = classOf(data, cls).skills.filter((s) => vet.skills.includes(s.id)).map((s) => s.name);
        return (
          <div className="pickrow" key={v}>
            <span className="who" title={known.length ? `Knows: ${known.join(', ')}` : 'No skills yet'}>
              {uiOf(cls).icon} {uiOf(cls).label} <i>{'★'.repeat(vet.skills.length)}</i>
            </span>
            {classOf(data, cls).skills.filter((s) => !vet.skills.includes(s.id)).map((s) => (
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
  abandoned: 'The battle was abandoned', restarted: 'Restarted', surrendered: 'You surrendered',
  'rules changed': 'The rules changed', 'faction retired': 'Its faction was retired',
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
            <div><span className="b">Blue</span> {survivors(0).map((u, i) => <span key={i}>{uiOf(u.cls).icon}</span>)}{survivors(0).length === 0 && '—'}</div>
            <div><span className="r">Red</span> {survivors(1).map((u, i) => <span key={i}>{uiOf(u.cls).icon}</span>)}{survivors(1).length === 0 && '—'}</div>
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

// Playback speed and sound.
export function DevPanel() {
  const { speed, setSpeed, version, quality, setQuality, autoLowered, tiltShift, toggleTiltShift } = useBattle();
  const [snd, setSnd] = useState(soundPrefs());
  const [rate, setRate] = useState(0);
  useEffect(() => { const t = setInterval(() => setRate(fps.value), 1000); return () => clearInterval(t); }, []);
  const change = (p: Partial<typeof snd>) => { setSound(p); setSnd(soundPrefs()); };
  return (
    <div className="dev">
      <span className="dim">dev · sim {version ?? '…'}{import.meta.env.DEV && <> · <a href="/?lab" target="_blank" rel="noreferrer">unit lab</a>{isTuned() && <b title="walk/run tuning from the unit lab is active in this browser"> · tuned</b>}</>}</span>
      {[1, 2, 4].map((s) => <button key={s} className={speed === s ? 'on' : ''} onClick={() => setSpeed(s)}>{s}×</button>)}
      <button aria-label={snd.muted ? 'unmute' : 'mute'} title={snd.muted ? 'Sound off' : 'Sound on'} onClick={() => change({ muted: !snd.muted })}>{snd.muted ? '🔇' : '🔊'}</button>
      <input type="range" min={0} max={1} step={0.05} value={snd.volume} aria-label="volume" style={{ width: 60 }}
        onChange={(e) => change({ volume: +e.target.value, muted: false })} />
      <select aria-label="graphics quality" value={quality} onChange={(e) => setQuality(e.target.value as 'high' | 'low')}>
        <option value="high">High</option><option value="low">Low</option>
      </select>
      <button className={tiltShift ? 'on' : ''} title="Tilt-shift depth of field (experimental)" onClick={toggleTiltShift}>DoF</button>
      <span className="dim" data-testid="fps">{rate} fps</span>
      {autoLowered && <span className="err" style={{ margin: 0 }} title="High couldn't keep up with this display's refresh rate">lowered for smoothness</span>}
    </div>
  );
}
