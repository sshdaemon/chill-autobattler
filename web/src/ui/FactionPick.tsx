import { useBattle } from '../store.ts';
import { uiOf } from '../units/looks.ts';
import { FACTIONS, factionLook } from '../factions/factions.ts';

// The first screen of a run: pick the faction you play. Shows each faction's
// lore and its units; factions still being built are shown as coming soon.
export function FactionPick() {
  const { data, startRun, error } = useBattle();
  if (!data) return <div className="overlay"><div className="panel">{error ?? 'Loading…'}</div></div>;
  const order = (name: string) => factionLook(name).order;
  const playable = data.factions.filter((f) => f.playable).sort((a, b) => order(a.name) - order(b.name));
  const building = data.factions.filter((f) => !f.playable).sort((a, b) => order(a.name) - order(b.name)); // in the sim, units being built
  const known = new Set(data.factions.map((f) => f.name));
  const coming = Object.entries(FACTIONS).filter(([name, f]) => f.coming && !known.has(name)); // lore only
  return (
    <div className="overlay">
      <div className="panel factions">
        <h2>Choose your faction</h2>
        <div className="sub">You play it for the whole run. Red's faction is drawn for each battle.</div>
        <div className="cards">
          {playable.map((f) => {
            const look = factionLook(f.name);
            return (
              <div className="card faction" key={f.id} style={{ borderTopColor: look.colour }}>
                <div className="icon">{look.icon}</div>
                <div className="name">{look.label}</div>
                <div className="blurb">{look.blurb}</div>
                {look.lore.map((p, i) => <p className="lore" key={i}>{p}</p>)}
                <ul className="units">
                  {f.units.map((c) => <li key={c}>{uiOf(c).icon} <b>{uiOf(c).label}</b>: {uiOf(c).blurb}</li>)}
                </ul>
                <button className="go" onClick={() => void startRun(f.id)}>Play {look.label} →</button>
              </div>
            );
          })}
          {building.map((f) => {
            const look = factionLook(f.name);
            return (
              <div className="card faction coming" key={f.id} style={{ borderTopColor: look.colour }}>
                <div className="icon">{look.icon}</div>
                <div className="name">{look.label}</div>
                <div className="blurb">{look.blurb}</div>
                {look.lore.map((p, i) => <p className="lore" key={i}>{p}</p>)}
                <ul className="units">{f.units.map((c) => <li key={c}>{uiOf(c).icon} <b>{uiOf(c).label}</b>: {uiOf(c).blurb}</li>)}</ul>
                <button disabled>Coming soon</button>
              </div>
            );
          })}
          {coming.map(([name, look]) => (
            <div className="card faction coming" key={name} style={{ borderTopColor: look.colour }}>
              <div className="icon">{look.icon}</div>
              <div className="name">{look.label}</div>
              <div className="blurb">{look.blurb}</div>
              {look.lore.map((p, i) => <p className="lore" key={i}>{p}</p>)}
              <ul className="units">{(look.units ?? []).map((u) => <li key={u.label}><b>{u.label}</b>: {u.role}</li>)}</ul>
              <button disabled>Coming soon</button>
            </div>
          ))}
        </div>
        {error && <div className="err">{error}</div>}
      </div>
    </div>
  );
}
