## Why

Playing Kyivan Rus against Kyivan Rus, both armies took up position and never
fought. A test of every faction pairing found it in 5 of 6 Rus mirrors and in
every battle with the (placeholder) Romans. Three temperament rules could wait
on each other forever: a line unit waited for its allies' centre, which
included the supports that wait behind it; a bodyguard and its ward each stayed
near the other; and a holding unit with nothing in sight stood still.
The player also asked for a way to give up a run and start from scratch.

## What Changes

- **Line:** keeps its place in the formation relative to the allies that
  lead (not those that follow the front line), measured against where it stood
  at the start.
- **Stay near:** a support behind the front line follows its melee ally; a
  bodyguard advances normally and only turns back when more than 4 m from its ward.
- **Hold:** with nothing in sight it marches like a line; once it sees an
  enemy it waits at most 4 s for it to come within its trigger.
- A test plays every faction pairing and fails if any battle runs out of time.
- `RulesVersion` 0.8.0; golden cases re-recorded where Red draws Rus (new
  mutual-wipe seeds).
- **Surrender:** a `SurrenderRun` call ends the active run (reason
  "surrendered", score kept); a Surrender button in the run bar, then the
  faction screen.

## Capabilities

### New Capabilities

### Modified Capabilities
- `battle-rules`: temperament rules can't deadlock.
- `runs`: surrender.

## Impact

`sim/engage_rules.go`, `sim/temperament.go`, `sim/mirror_test.go`,
`proto/chill/v1/run.proto`, `server/run`, `web` (store, run bar).
