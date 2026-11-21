# Battle plans roadmap

Status: November 2026. Steps 1–5 are done (#85, #87, #89, #91 and step 5's PR); step 6 is to decide.

The goal: give the player a say in **how** their army fights, not only what it
is. Today every unit fights by its own temperament (`sim/temperament.go`), so
battles between the same armies play out much the same way. A battle plan is
an army-wide stance the player picks before the battle. Later come tactics
(groups and triggers), then plans written in free text and turned into the same
data by a language model.

```
Step 1  Plans in the sim: built-in plans, Red's draw, goldens
Step 2  Plan picker in the client
Step 3  Balance pass: plan-vs-plan matrix
Step 4  Tactics: groups and one trigger
Step 5  Custom plans: the player builds a plan from chips
Step 6  Free-text orders through a language model   ── TO DECIDE LATER
```

Each step is its own OpenSpec change and PR.

## Ground rules

- **The sim stays deterministic.** A plan is data in the battle input, like
  the formation. Nothing outside the sim decides anything during a battle, so
  replays, goldens and the server's fairness check keep working.
- **Plan 0 is "Instinct"**: today's behaviour. A battle with plan 0 on both
  sides plays out exactly as it does now, and the golden replays prove it.
- **The server checks a plan as data**, the same way it checks an army. It
  never needs to know where the plan came from: a built-in preset, the chips
  editor or a language model.
- **Keep the vocabulary small.** A plan sets a few fields per role, in coarse
  steps (close, mid or far rather than raw distances), so it stays readable,
  balanceable and easy for a small model to fill in later.

## Decisions taken

- **Plan, then skills.** A unit's temperament is built as roster, then plan,
  then learned skills. The plan sets the army's stance; skills still add their
  own twist. Fields a plan leaves unset (a völva's `StayNear`, say) keep the
  unit's own value.
- **Red picks a plan too**, drawn from the seed like its formation. Red's plan
  stays hidden until the battle starts. Players draft and counter-pick units,
  not the enemy's strategy.
- **The plan is part of the rules**, so adding plans bumps `RulesVersion`.
- **No stalls.** Two waiting plans against each other must not reach the
  3-minute draw (the Rus vs Rus stall again). The golden tests include that
  matchup.

---

## Step 1: Plans in the sim (done, #85)

Built in OpenSpec change `battle-plans`. Red's plan is drawn from the combat
seed. The first plan-vs-plan matrix is in `docs/balance-findings.md`: All-out
charge dominates, and Skirmish and Shield wall are weak, which step 3 tunes.

- **Data:** `PlanDef { ID, Name, Melee, Ranged, Caster Stance }`. A `Stance` is
  optional overrides of the `Temperament` fields: target rule, engage rule,
  trigger, leash, fall-back and kite. Each plan lives in its own file
  (`sim/plan_<name>.go`) and registers itself under its id from `ids.go`, like
  factions and effects.
- **Input:** `Army.Plan` (proto `Army.plan` in `units.proto`). `Validate`
  rejects an unknown plan.
- **Red's draw:** a new stream `StreamRedPlan`, added at the **end** of the
  stream list so the existing streams (army, skills, formation, faction) keep
  their values.
- **Built-in plans.** Almost all of them use target and engage rules that exist
  already:

  | Plan | Melee | Ranged and caster | Feel |
  |---|---|---|---|
  | Instinct | as now | as now | the default |
  | Shield wall | Hold, Protect | Focus within reach, stay behind melee | patient, counters rushes |
  | All-out charge | Charge, Nearest | Charge, no kiting | fast and risky |
  | Hunt the archers | Charge, RangedFirst | Focus within reach | counters ranged-heavy armies |
  | Focus fire | Line, Focus | Focus | a slow, deadly block |
  | Skirmish | Hold, retreat at 40% health | Kite, keep distance | bleeds the enemy |

- **Stall tie-breaker:** for example, waiting units start to advance slowly
  after N seconds. To settle in the step's design.
- **Tests:** goldens for every plan, plan 0 byte-identical to today, a Shield
  wall vs Shield wall battle that ends before the time limit.
- **Balance tool:** `tools/balance` can fix a plan per side and report a
  plan-vs-plan win matrix.

## Step 2: Plan picker in the client (done, #87)

Built in OpenSpec change `battle-plan-picker`.

- A plan picker next to the formation in the deploy bar
  (`web/src/ui/DeployBar.tsx`), with a one-line description per plan. The
  chosen plan is remembered for the run.
- Red's plan is shown when the battle starts (a banner such as "Red: Shield
  wall"), not during the draft.
- Plan names and descriptions are presentation, in `web/src/`.

## Step 3: Balance pass (done, #89)

Built in OpenSpec change `plan-balance`: plans may grant effects for the
battle (an edge and a cost), and no plan dominates in any faction. Results in
`docs/balance-findings.md`. Step 5's custom plans must limit which effects a
player can pick, or they become a free buff.

- Run the plan-vs-plan matrix across the four factions.
- Aim: no plan wins against everything. Each plan has at least one plan it
  beats and one it loses to, roughly rock-paper-scissors.
- Write up the findings in `docs/balance-findings.md`.

## Step 4: Tactics (done)

Built in OpenSpec change `battle-tactics`. A tactic is a plan with groups, so
it travels in the same `Army.plan` field: three built-in tactics (Hammer and
anvil, Reserve, Hold, then charge), groups split left and right at the army's
mean x, three triggers (time, ally falls, other group fighting) and a Flank
engage rule. A trigger switches behaviour only; a tactic's effects are fixed
at spawn. Holding units wait for their own group. The deploy view tags Blue's
units L or R. Results in `docs/balance-findings.md`. Triggers on enemy
distance or group health are left for step 5's chips.

- **Groups:** the army splits into two or three groups, by default the
  formation's flanks. Each group gets its own stance. This makes classic
  moves possible, such as hammer and anvil: one flank holds while the other
  goes around.
- **One trigger per group:** the group switches to a second stance when
  something happens. For example: a set time passes, the first ally falls, an
  enemy reaches the line, or a group drops below some health.
- Same rules as Step 1: data in the input, goldens, balance matrix.

## Step 5: Custom plans (done)

Built in OpenSpec change `custom-plans`. Chips per role (target, engage, hold
trigger, keep distance, leash, fall back, stay near, kite) in coarse steps,
optional left and right groups with a trigger and a second stance, and one edge
per role from a five-edge menu (each an advantage with a cost). The battle
input carries it as `Army.custom_plan`; the sim checks it like a built-in plan.
The editor opens from the deploy bar and pauses the countdown. Balance:
`-exp edges` and `-exp custom` in the balance tool; results in
`docs/balance-findings.md` (the Rus come closest to a dominating custom plan).
For Step 6: a model fills the same JSON shape (`data().plans[i].chips`,
`chillSim.checkPlan`).

- An editor of chips: the player builds a plan from the Step 1 and Step 4
  vocabulary (per role or group: target, engage, retreat, trigger) instead of
  only picking a preset. Presets become starting points.
- The server checks the custom plan like any other input.
- This step is the fallback for Step 6 and its confirmation screen: whatever a
  model produces is shown back as these chips, and the player can fix it.

## Step 6: Free-text orders (to decide later)

The player types orders ("huskarls hold the line until their archers commit,
then go for the völva"), and a language model turns them into a Step 5 plan.
Optional: without text, the picker and chips work as before.

A model can't run inside the battle: a tick is 50 ms, model calls take
seconds, and model output isn't repeatable. It only fills in the plan before
the battle, so the sim stays deterministic.

**Options, cheapest first:**

| Option | Where it runs | Good | Bad |
|---|---|---|---|
| Chrome's built-in model (Prompt API, Gemini Nano) | browser | free, nothing to host, output forced to a JSON schema | Chrome desktop only; Chrome downloads the model on first use; heavy disk and GPU needs |
| WebLLM or transformers.js with a small model (Qwen3 0.6–1.7B, SmolLM2, Gemma) | browser (WebGPU) | any WebGPU browser, schema-forced output, free | one-off 0.5–1.6 GB download, slow on weak devices |
| Small embedding model (about 25 MB) matching text to a preset | browser | tiny, works everywhere | picks a preset only, can't write a custom plan |
| Self-hosted open model (llama.cpp or Ollama) | our server | free per call, any device | we run and pay for the machine |
| Hosted API (Claude Haiku 5.5) | our server | best understanding of the text, a fraction of a cent per plan | API key, cost and rate limits; the text leaves the device |

**Leaning towards** a tiered setup: chips always; Chrome's built-in model where
available; otherwise an opt-in WebLLM download; a server API only if
small models fail on real player phrasing. Running in the browser keeps the
player's text on their device.

**Open questions:**

- Which tier to start with, and whether the WebLLM download is worth it.
- A test set: 50–100 example orders with their expected plans, to score each
  model choice before shipping.
- Whether free text may also set tactics (Step 4) or only the plan.
- If a server option is chosen: rate limits per run, and what to do when the
  service is down (fall back to chips).
