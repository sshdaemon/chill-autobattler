import { armyClasses, DEPLOY_SECONDS, useBattle } from '../store.ts';
import { CUSTOM, PRESETS } from '../gamedata.ts';
import { planWords } from '../plans.ts';

// Mini-diagram of where your own units would stand (top = toward the enemy).
function Diagram({ f }: { f: number }) {
  const { slots, army, data } = useBattle();
  const frame = (inner: React.ReactNode, dashed = false) => (
    <svg viewBox="0 0 96 56"><rect x="1" y="1" width="94" height="54" rx="5" fill="rgba(58,140,255,.12)" stroke="#3a8cff" strokeDasharray={dashed ? '4 3' : undefined} />{inner}</svg>
  );
  if (f === CUSTOM) return frame(<text x="48" y="37" fontSize="24" textAnchor="middle">✋</text>, true);
  const z = data?.deployZone, pos = slots[f];
  if (!z || !pos) return frame(null);
  const sx = (x: number) => 4 + ((x - z[0]) / (z[1] - z[0])) * 88, sy = (v: number) => 4 + ((v - z[2]) / (z[3] - z[2])) * 48;
  const classes = armyClasses({ army });
  return frame(pos.map((p, i) => {
    const x = sx(p.x), y = sy(p.z);
    return classes[i] === 0 ? <circle key={i} cx={x} cy={y} r="4.2" fill="#3a8cff" stroke="#bfe0ff" />
      : classes[i] === 1 ? <circle key={i} cx={x} cy={y} r="3.6" fill="none" stroke="#fff" strokeWidth="1.8" />
        : <rect key={i} x={x - 3.4} y={y - 3.4} width="6.8" height="6.8" fill="#b48cff" transform={`rotate(45 ${x} ${y})`} />;
  }));
}

export function DeployBar() {
  const { formation, setFormation, plan, setPlan, deployLeft, ready, data, phase } = useBattle();
  if (!data) return null;
  const locking = phase === 'locking';
  return (
    <div className="deploybar">
      <div className="deploymain">
      <div className="plans">
        <span className="lbl">Plan</span>
        {data.plans.map((p) => {
          const w = planWords(p.name);
          return (
            <button key={p.id} className={`pbtn ${plan === p.id ? 'on' : ''}`} disabled={locking} title={w.desc} onClick={() => setPlan(p.id)}>{w.title}</button>
          );
        })}
      </div>
      <div className="plandesc">{planWords(data.plans[plan]?.name).desc}</div>
      <div className="formations">
        {[...PRESETS, CUSTOM].map((f) => (
          <button key={f} className={`fbtn ${formation === f ? 'on' : ''}`} disabled={f === CUSTOM || locking}
            title={f === CUSTOM ? 'Drag a unit to place it yourself' : undefined} onClick={() => setFormation(f)}>
            <Diagram f={f} />
            {f === 1 ? 'Fully spread' : data.formations[f]}
          </button>
        ))}
      </div>
      </div>
      <div className="deployside">
        <div className="timer"><b>{Math.ceil(deployLeft)}</b> s</div>
        <div className="bar"><i style={{ width: `${(deployLeft / DEPLOY_SECONDS) * 100}%` }} /></div>
        <button className="go" disabled={locking} onClick={ready}>{locking ? 'Locking…' : 'Ready ▶'}</button>
        <div className="hint">Drag your units inside the blue zone</div>
      </div>
    </div>
  );
}
