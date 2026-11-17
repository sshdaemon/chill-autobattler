import { armySpecs, useBattle, vetCounts } from '../store.ts';
import { CLASS_BLURB, CLASS_ICON, CLASS_LABEL } from '../gamedata.ts';

const pct = (v: number) => `${Math.round(v * 100)}%`;

// Before each match: pick 5 units. Stats come from the sim (wasm data()).
export function Picker() {
  const { data, counts, setCount, toDeploy, error, run } = useBattle();
  if (!data || !run) return <div className="overlay"><div className="panel">{error ?? 'Loading…'}</div></div>;
  const total = counts[0] + counts[1] + counts[2], size = data.armySize;
  const vets = vetCounts(run), army = armySpecs({ run, counts });
  return (
    <div className="overlay">
      <div className="panel picker">
        <h2>{run.match === 1 ? 'Choose your army' : `Match ${run.match}: reinforce your army`}</h2>
        <div className="sub">{run.veterans.length
          ? `Your ${run.veterans.length} veteran${run.veterans.length === 1 ? '' : 's'} (★) stay with their skills. Fill the rest.`
          : `Pick ${size} units. Next you arrange them, then they fight on their own.`}</div>
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
                <button aria-label={`fewer ${CLASS_LABEL[i]}s`} disabled={counts[i] <= vets[i]}
                  title={vets[i] ? `${vets[i]} veteran${vets[i] === 1 ? '' : 's'} can't be dismissed` : undefined} onClick={() => setCount(i, -1)}>−</button>
                <b>{counts[i]}</b>
                <button aria-label={`more ${CLASS_LABEL[i]}s`} disabled={total >= size} onClick={() => setCount(i, 1)}>+</button>
              </div>
            </div>
          ))}
        </div>
        <div className="slots">
          {Array.from({ length: size }, (_, k) => {
            const u = army[k];
            if (!u) return <span key={k} className="empty" />;
            const names = u.skills.map((id) => data.classes[u.cls].skills.find((s) => s.id === id)?.name).join(', ');
            return <span key={k} className={u.vet ? 'vet' : ''} title={u.vet ? `Veteran: ${names}` : 'Recruit'}>
              {CLASS_ICON[u.cls]}{u.vet && <i>{'★'.repeat(u.skills.length)}</i>}
            </span>;
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
