import { useEffect, useState } from 'react';
import { useBattle } from '../store.ts';
import { UF, type UnitSnap } from '../sim/snapshot.ts';
import { CLASS_ICON } from '../gamedata.ts';
import { lookFor } from '../scene/models.ts';
import { portrait } from '../scene/portraits.ts';
import { FOGGED_FIGHT } from '../scene/Fog.tsx';

export const nthOfClass = (units: UnitSnap[], i: number) =>
  units.slice(0, i).filter((u) => u.team === units[i].team && u.cls === units[i].cls).length;

function Card({ index }: { index: number }) {
  const u = useBattle((s) => s.curr?.units[index]);
  const nth = useBattle((s) => (s.curr ? nthOfClass(s.curr.units, index) : 0));
  const unseen = useBattle((s) => {
    const x = s.curr?.units[index];
    return !!x && x.team === 1 && s.fogOn && FOGGED_FIGHT.includes(s.phase) && !(x.flags & UF.seen);
  });
  const selected = useBattle((s) => s.selected === index);
  const [img, setImg] = useState<string>();
  const cls = u?.cls, team = u?.team;
  useEffect(() => {
    if (cls === undefined || team === undefined) return;
    let live = true;
    portrait(cls, lookFor(cls, nth), team).then((url) => { if (live) setImg(url); }).catch(() => undefined);
    return () => { live = false; };
  }, [cls, team, nth]);
  if (!u) return null;
  const dead = !!(u.flags & UF.dead);
  const { select, center } = useBattle.getState();
  return (
    <div className={`pcard ${dead ? 'dead' : ''} ${selected ? 'sel' : ''} t${u.team}`} title={lookFor(u.cls, nth)}
      onClick={() => select(index)} onDoubleClick={() => center(index)}>
      {u.skills > 0 && <span className="vet">{'★'.repeat(u.skills)}</span>}
      {img ? <img src={img} alt="" /> : <div className="ph">{CLASS_ICON[u.cls]}</div>}
      {dead && <span className="cross">✝</span>}
      <div className="bar"><i style={{ width: unseen || dead ? 0 : `${(u.hp / u.maxHp) * 100}%` }} /></div>
      <div className="num">{dead ? '—' : unseen ? '?' : Math.ceil(u.hp)}</div>
    </div>
  );
}

// Blue along the bottom-left, Red along the bottom-right.
export function Portraits() {
  const units = useBattle((s) => s.curr?.units.map((u) => u.team).join(','));
  const phase = useBattle((s) => s.phase);
  if (!units || !['countdown', 'fighting', 'done'].includes(phase)) return null;
  const teams = units.split(',').map(Number);
  return <>
    {[0, 1].map((team) => (
      <div key={team} className={`portraits ${team ? 'red' : 'blue'}`}>
        {teams.map((t, i) => (t === team ? <Card key={i} index={i} /> : null))}
      </div>
    ))}
  </>;
}
