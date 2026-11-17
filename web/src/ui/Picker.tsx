import { useBattle } from '../store.ts';
import { CLASS_BLURB, CLASS_ICON, CLASS_LABEL } from '../gamedata.ts';

const pct = (v: number) => `${Math.round(v * 100)}%`;

// Before each match: pick 5 units. Stats come from the sim (wasm data()).
export function Picker() {
  const { data, counts, setCount, toDeploy, error } = useBattle();
  if (!data) return <div className="overlay"><div className="panel">Loading the battle engine…</div></div>;
  const total = counts[0] + counts[1] + counts[2], size = data.armySize;
  return (
    <div className="overlay">
      <div className="panel picker">
        <h2>Choose your army</h2>
        <div className="sub">Pick {size} units. Next you arrange them, then they fight on their own.</div>
        <div className="cards">
          {data.classes.map((c, i) => (
            <div className="card" key={c.name}>
              <div className="icon">{CLASS_ICON[i]}</div>
              <div className="name">{CLASS_LABEL[i]}</div>
              <div className="blurb">{CLASS_BLURB[i]}</div>
              <table><tbody>
                <tr><td>Health</td><td>{c.hp}</td></tr>
                <tr><td>Damage</td><td>{c.dmgLo}–{c.dmgHi}</td></tr>
                <tr><td>Range</td><td>{c.range} m</td></tr>
                <tr><td>Cooldown</td><td>{c.cooldown} s</td></tr>
                <tr><td>Speed</td><td>{c.speed} m/s</td></tr>
                <tr><td>Sight</td><td>{c.sight} m</td></tr>
              </tbody></table>
              <div className="takes">Damage taken<br />{data.classes.map((a, j) => `${CLASS_ICON[j]} ${pct(a.matchup[i])}`).join(' · ')}</div>
              <div className="count">
                <button aria-label={`fewer ${CLASS_LABEL[i]}s`} disabled={counts[i] === 0} onClick={() => setCount(i, -1)}>−</button>
                <b>{counts[i]}</b>
                <button aria-label={`more ${CLASS_LABEL[i]}s`} disabled={total >= size} onClick={() => setCount(i, 1)}>+</button>
              </div>
            </div>
          ))}
        </div>
        <div className="slots">
          {Array.from({ length: size }, (_, k) => {
            const cls = k < counts[0] ? 0 : k < counts[0] + counts[1] ? 1 : k < total ? 2 : -1;
            return <span key={k} className={cls < 0 ? 'empty' : ''}>{cls < 0 ? '' : CLASS_ICON[cls]}</span>;
          })}
        </div>
        <button className="go" disabled={total !== size} onClick={toDeploy}>
          {total === size ? 'Continue to deployment →' : `Pick ${size - total} more`}
        </button>
        {error && <div className="err">{error}</div>}
      </div>
    </div>
  );
}
