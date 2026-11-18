import { useBattle, factionUnits } from '../store.ts';
import { uiOf } from '../units/looks.ts';
import { factionLook } from '../factions/factions.ts';
import { BLUE, bluesChoice, current, done, leftInTurn, type DraftUnit } from '../draft.ts';
import type { GameData } from '../gamedata.ts';

const pct = (v: number) => `${Math.round(v * 100)}%`;

// The draft before each battle (roadmap 1.3): both sides pick five units in
// turns, Blue a unit type of its faction and its level-1 ability per pick.
export function DraftScreen() {
  const { data, run, draft, draftPick, retryDraft, error } = useBattle();
  if (!data || !run) return <div className="overlay"><div className="panel">{error ?? 'Loading…'}</div></div>;
  if (!draft) return (
    <div className="overlay"><div className="panel">
      <div className="err">{error ?? 'Setting up the draft…'}</div>
      {error && <button className="go" onClick={retryDraft}>Try again</button>}
    </div></div>
  );
  const mine = factionLook(data.factions.find((f) => f.id === run.faction)?.name ?? '');
  const theirs = factionLook(data.factions.find((f) => f.id === draft.plan.redFaction)?.name ?? '');
  const yours = bluesChoice(draft), left = Math.max(0, Math.ceil((draft.turnEnds - Date.now()) / 1000));
  const status = done(draft) ? 'Draft complete: to deployment…'
    : yours ? `Your pick: ${leftInTurn(draft)} more this turn · ${left} s`
      : current(draft).side === BLUE ? 'Your veterans take their places…' : `${theirs.label} (Red) is picking…`;
  return (
    <div className="overlay">
      <div className="panel draft">
        <h2>Draft · match {run.match}</h2>
        <div className="sub">{draft.plan.first === BLUE ? 'You pick first.' : 'Red picks first.'} Turns of 1-2-2-2-2-1 picks; each pick is a unit and its first ability.</div>
        <div className={`turn ${yours ? 'yours' : ''}`}>{status}</div>
        <div className="sides">
          <Side label={`${mine.icon} You · ${mine.label}`} units={draft.blue} data={data} team="b" />
          <Side label={`${theirs.icon} Red · ${theirs.label}`} units={draft.red} data={data} team="r" />
        </div>
        <div className="cards">
          {factionUnits(data, run).map((i) => {
            const c = data.classes[i];
            return (
              <div className={`card ${yours ? '' : 'waiting'}`} key={c.name}>
                <div className="icon">{uiOf(i).icon}</div>
                <div className="name">{uiOf(i).label}</div>
                <div className="blurb">{uiOf(i).blurb}</div>
                <div className="stats">{c.hp} HP · {c.dmgLo}–{c.dmgHi} dmg · {c.range} m · {c.speed} m/s</div>
                <div className="takes">Damage taken: {factionUnits(data, run).map((j) => `${uiOf(j).icon} ${pct(data.classes[j].matchup[i])}`).join(' · ')}</div>
                <div className="abilities">
                  {c.skills.map((s) => (
                    <button key={s.id} disabled={!yours} title={s.pct ? `${s.name}: ${s.pct}%` : s.name} onClick={() => draftPick(i, s.id)}>{s.name}</button>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
        {error && <div className="err">{error}</div>}
      </div>
    </div>
  );
}

function Side({ label, units, data, team }: { label: string; units: DraftUnit[]; data: GameData; team: 'b' | 'r' }) {
  return (
    <div className={`side ${team}`}>
      <div className="label">{label}</div>
      <div className="slots">
        {Array.from({ length: data.armySize }, (_, k) => {
          const u = units[k];
          if (!u) return <span key={k} className="empty" />;
          const names = u.skills.map((id) => data.classes[u.cls].skills.find((s) => s.id === id)?.name).join(', ');
          return <span key={k} className={u.vet ? 'vet' : ''} title={`${uiOf(u.cls).label}${u.vet ? ' (veteran)' : ''}: ${names}`}>
            {uiOf(u.cls).icon}<i>{'★'.repeat(u.skills.length)}</i>
          </span>;
        })}
      </div>
    </div>
  );
}
