import { useRef } from 'react';
import { useBattle } from '../store.ts';
import { UF, type UnitSnap } from '../sim/snapshot.ts';
import { CLASS_ICON, CLASS_LABEL } from '../gamedata.ts';
import { lookFor } from '../scene/models.ts';
import { FOGGED_FIGHT } from '../scene/Fog.tsx';
import { nthOfClass } from './Portraits.tsx';

export function status(f: number): string {
  if (f & UF.dead) return 'Dead';
  const base = f & UF.staggered ? 'Staggered' : f & UF.dashing ? 'Dashing' : f & UF.swinging ? 'Attacking'
    : f & UF.kiting ? 'Falling back from melee' : f & UF.moving ? 'Advancing' : 'Holding';
  const extra = [f & UF.wet ? 'wading, 70% speed' : '', f & UF.slowed ? 'slowed by frost' : ''].filter(Boolean);
  return extra.length ? `${base} · ${extra.join(' · ')}` : base;
}

// The selected unit. An unseen Red unit keeps showing what was last seen.
export function InfoPanel() {
  const { selected, curr, data, run, fogOn, phase, select, follow } = useBattle();
  const last = useRef<{ index: number; u: UnitSnap }>(undefined);
  if (selected === undefined || !curr || !data) return null;
  const now = curr.units[selected];
  if (!now) return null;
  const unseen = now.team === 1 && fogOn && FOGGED_FIGHT.includes(phase) && !(now.flags & UF.seen);
  if (!unseen || last.current?.index !== selected) last.current = { index: selected, u: now };
  const u = unseen ? last.current.u : now;
  const cls = data.classes[u.cls];
  const blueIndex = curr.units.slice(0, selected).filter((x) => x.team === 0).length;
  // veterans' skill names; after the result the run's veterans are already the new ones
  const skills = u.team === 0 && phase !== 'done' && blueIndex < (run?.veterans.length ?? 0)
    ? run!.veterans[blueIndex].skills.map((id) => cls.skills.find((s) => s.id === id)?.name).join(', ')
    : '';
  return (
    <div className="info">
      <button className="x" onClick={() => select(undefined)} aria-label="close">×</button>
      <div className="title">
        <span className={u.team ? 'r' : 'b'}>{lookFor(u.cls, nthOfClass(curr.units, selected))}</span>
        {u.skills > 0 && <span className="vetstars"> {'★'.repeat(u.skills)}</span>}
        <span className="dim"> · {CLASS_ICON[u.cls]} {CLASS_LABEL[u.cls]}{u.team ? ' (Red)' : ''}</span>
      </div>
      <div className="hp"><i style={{ width: `${(u.hp / u.maxHp) * 100}%` }} /><span>{Math.ceil(u.hp)} / {u.maxHp}</span></div>
      <div className="st">{unseen ? 'Out of sight' : status(u.flags)}</div>
      <table><tbody>
        <tr><td>Kills</td><td>{u.kills}</td><td>Damage dealt</td><td>{Math.round(u.dealt)}</td></tr>
        <tr><td>Damage</td><td>{cls.dmgLo}–{cls.dmgHi}</td><td>Range</td><td>{cls.range} m</td></tr>
      </tbody></table>
      <div className="takes">Takes from: {data.classes.map((a, j) => `${CLASS_ICON[j]} ${Math.round(a.matchup[u.cls] * 100)}%`).join(' · ')}</div>
      <div className="skills">{u.skills === 0 ? 'No skills' : skills || `${u.skills} skill${u.skills === 1 ? '' : 's'}`}</div>
      <div className="dim keys">F {follow ? 'stop following' : 'follow'} · Space centre · double-click a portrait to centre</div>
    </div>
  );
}
