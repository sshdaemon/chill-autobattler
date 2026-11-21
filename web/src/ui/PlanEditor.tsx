import { useEffect, useMemo, useState } from 'react';
import { checkPlan, useBattle } from '../store.ts';
import { emptyRoles, type ChipsJSON, type CustomPlanJSON, type GameData, type Roles3 } from '../gamedata.ts';
import {
  CUSTOM_PLAN, EDGE_WORDS, ENGAGE_WORDS, KITE_WORDS, LEASH_WORDS, ROLE_NAMES, STAY_WORDS, STEP_WORDS, TARGET_WORDS, TRIGGER_WORDS,
  edgeEffects, planWords,
} from '../plans.ts';

// The custom-plan editor (step 5 of docs/battle-plans-roadmap.md): chips per
// role, optionally per group with a trigger and a second stance. The sim
// checks every change (chillSim.checkPlan); only a valid plan can be saved.

// The sim's reason, in players' words: "plan custom on strilets (…): a holding
// unit needs a trigger distance" → "strilets: a holding unit needs a trigger distance".
function friendly(e: string): string {
  const reason = e.slice(e.lastIndexOf(': ') + 2);
  const words: Record<string, string> = {
    'a holding unit needs a trigger distance': "units told to hold need a 'Hold until' distance",
    'only ranged units kite': 'only ranged units can kite',
  };
  const unit = /\bon ([\p{L} ]+?) \(/u.exec(e)?.[1];
  return (words[reason] ?? reason) + (unit ? ` (${unit})` : '');
}

const clone = (p: CustomPlanJSON): CustomPlanJSON => JSON.parse(JSON.stringify(p)) as CustomPlanJSON;
const metres = (m: number) => `${m} m`;

function Row({ label, options, value, onPick }: { label: string; options: [number, string, string?][]; value: number; onPick: (v: number) => void }) {
  return (
    <div className="chiprow">
      <span className="lbl">{label}</span>
      {options.map(([v, text, title]) => (
        <button key={v} className={`chip ${value === v ? 'on' : ''}`} title={title} onClick={() => onPick(v)}>{text}</button>
      ))}
    </div>
  );
}

// What a role's chips do, in a sentence.
function describe(c: ChipsJSON, data: GameData): string {
  const v = data.chips, parts: string[] = [];
  const name = (list: { id: number; name: string }[], id: number) => list.find((x) => x.id === id)?.name ?? '';
  if (c.target) parts.push(`targets ${(TARGET_WORDS[name(v.targets, c.target)] ?? name(v.targets, c.target)).toLowerCase()}`);
  if (c.engage) parts.push((ENGAGE_WORDS[name(v.engages, c.engage)] ?? name(v.engages, c.engage)).toLowerCase());
  if (c.trigger) parts.push(`go in when an enemy is within ${metres(v.trigger[c.trigger])}`);
  if (c.keep) parts.push(`keep ${metres(v.keep[c.keep])} away`);
  if (c.leash) parts.push(`stay within ${metres(v.leash[c.leash])} of the army`);
  if (c.fallBack) parts.push(`fall back below ${v.fallBack[c.fallBack]}% health`);
  if (c.stayNear) parts.push(STAY_WORDS[c.stayNear].toLowerCase());
  if (c.kite) parts.push(KITE_WORDS[c.kite].toLowerCase());
  const e = v.edges.find((x) => x.id === c.edge);
  if (e) parts.push(`${EDGE_WORDS[e.name] ?? e.name} (${edgeEffects(e.effects)})`);
  return parts.length ? parts.join(', ') : 'fight their own way';
}

function RoleChips({ role, chips, edges, data, set }: { role: number; chips: ChipsJSON; edges: boolean; data: GameData; set: (k: keyof ChipsJSON, v: number) => void }) {
  const v = data.chips;
  const pick = (k: keyof ChipsJSON) => (x: number) => set(k, x);
  const steps = (list: number[], words: string[]) => list.map((m, i): [number, string] => [i, i ? `${words[i]} ${metres(m)}` : 'Own']);
  return (
    <div className="role">
      <div className="rolehead"><b>{ROLE_NAMES[role]}</b> <span>{describe(chips, data)}</span></div>
      <Row label="Target" value={chips.target} onPick={pick('target')}
        options={[[0, 'Own'], ...v.targets.map((t): [number, string] => [t.id, TARGET_WORDS[t.name] ?? t.name])]} />
      <Row label="Engage" value={chips.engage} onPick={pick('engage')}
        options={[[0, 'Own'], ...v.engages.map((t): [number, string] => [t.id, ENGAGE_WORDS[t.name] ?? t.name])]} />
      <Row label="Hold until" value={chips.trigger} onPick={pick('trigger')} options={steps(v.trigger, STEP_WORDS)} />
      <Row label="Keep away" value={chips.keep} onPick={pick('keep')} options={steps(v.keep, STEP_WORDS)} />
      <Row label="Leash" value={chips.leash} onPick={pick('leash')} options={steps(v.leash, LEASH_WORDS)} />
      <Row label="Fall back" value={chips.fallBack} onPick={pick('fallBack')}
        options={v.fallBack.map((p, i): [number, string] => [i, i ? `below ${p}%` : 'Own'])} />
      <Row label="Stay near" value={chips.stayNear} onPick={pick('stayNear')} options={STAY_WORDS.map((w, i): [number, string] => [i, w])} />
      {role === 1 && <Row label="Kite" value={chips.kite} onPick={pick('kite')} options={KITE_WORDS.map((w, i): [number, string] => [i, w])} />}
      {edges && (
        <Row label="Edge" value={chips.edge} onPick={pick('edge')}
          options={[[0, 'None'], ...v.edges.filter((e) => e.roles[role]).map((e): [number, string, string] => [e.id, EDGE_WORDS[e.name] ?? e.name, edgeEffects(e.effects)])]} />
      )}
    </div>
  );
}

export function PlanEditor() {
  const { data, plan, custom, closeEditor, saveCustom } = useBattle();
  const [draft, setDraft] = useState<CustomPlanJSON>(() =>
    clone(plan === CUSTOM_PLAN && custom ? custom : data!.plans[plan]?.chips ?? { roles: emptyRoles(), groups: [] }));
  const [tab, setTab] = useState<-1 | 0 | 1>(-1); // -1: the whole army
  const [stage, setStage] = useState<'opening' | 'then'>('opening');
  const [error, setError] = useState<string | undefined>();
  useEffect(() => {
    let live = true;
    setError(undefined);
    void checkPlan(draft).then((e) => { if (live) setError(e); });
    return () => { live = false; };
  }, [draft]);
  const from = useMemo(() => (plan === CUSTOM_PLAN ? 'your custom plan' : planWords(data?.plans[plan]?.name).title), [plan, data]);
  if (!data) return null;
  const grouped = draft.groups.length === 2;
  const roles: Roles3 = tab === -1 ? draft.roles : draft.groups[tab][stage];
  const setChip = (i: number, k: keyof ChipsJSON, v: number) => setDraft((prev) => {
    const d = clone(prev);
    (tab === -1 ? d.roles : d.groups[tab][stage])[i][k] = v;
    return d;
  });
  const toggleGroups = () => {
    setDraft((prev) => ({ ...clone(prev), groups: prev.groups.length ? [] : [0, 1].map(() => ({ opening: emptyRoles(), then: emptyRoles(), trigger: 0, seconds: 0 })) }));
    setTab(-1);
  };
  const group = tab === -1 ? undefined : draft.groups[tab];
  const setTrigger = (trigger: number, seconds: number) => setDraft((prev) => {
    const d = clone(prev);
    d.groups[tab as 0 | 1].trigger = trigger;
    d.groups[tab as 0 | 1].seconds = seconds;
    return d;
  });
  const after = data.chips.triggers.find((t) => t.name === 'after')?.id ?? 1;
  return (
    <div className="planeditor" role="dialog" aria-label="Plan editor">
      <div className="pehead">
        <b>Custom plan</b> <span className="from">started from {from}</span>
        <label className="split"><input type="checkbox" checked={grouped} onChange={toggleGroups} /> Split into left and right groups</label>
      </div>
      <div className="petabs">
        <button className={`pbtn ${tab === -1 ? 'on' : ''}`} onClick={() => setTab(-1)}>Whole army</button>
        {grouped && <>
          <button className={`pbtn ${tab === 0 ? 'on' : ''}`} onClick={() => setTab(0)}>Left (L)</button>
          <button className={`pbtn ${tab === 1 ? 'on' : ''}`} onClick={() => setTab(1)}>Right (R)</button>
        </>}
      </div>
      {group && <>
        <Row label="Switch" value={group.trigger === after ? 100 + group.seconds : group.trigger}
          onPick={(x) => (x >= 100 ? setTrigger(after, x - 100) : setTrigger(x, 0))}
          options={[[0, 'Never'], ...data.chips.triggerSeconds.map((s): [number, string] => [100 + s, `After ${s} s`]),
            ...data.chips.triggers.filter((t) => t.id !== after).map((t): [number, string] => [t.id, TRIGGER_WORDS[t.name] ?? t.name])]} />
        <div className="petabs">
          <button className={`pbtn ${stage === 'opening' ? 'on' : ''}`} onClick={() => setStage('opening')}>Before the switch</button>
          <button className={`pbtn ${stage === 'then' ? 'on' : ''}`} disabled={!group.trigger} onClick={() => setStage('then')}>After the switch</button>
        </div>
      </>}
      <div className="peroles">
        {[0, 1, 2].map((i) => (
          <RoleChips key={`${tab}-${stage}-${i}`} role={i} chips={roles[i]} edges={tab === -1 || stage === 'opening'} data={data} set={(k, v) => setChip(i, k, v)} />
        ))}
      </div>
      <div className="pefoot">
        <span className={`peerror ${error ? 'bad' : ''}`}>{error === undefined ? 'Checking…' : error ? `Can't use this plan: ${friendly(error)}` : 'Ready to use.'}</span>
        <button className="pbtn" onClick={closeEditor}>Cancel</button>
        <button className="go" disabled={error !== ''} onClick={() => saveCustom(draft)}>Use this plan</button>
      </div>
    </div>
  );
}
