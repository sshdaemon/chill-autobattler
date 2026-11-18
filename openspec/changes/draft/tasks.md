## 1. Sim and server

- [x] 1.1 `sim.Draft` (first picker, Red's reveal order) and `DraftTurns`, from the map seed
- [x] 1.2 wasm `draft(seed, redSkills)`: order, Red's faction and army
- [x] 1.3 Deploy window 3 minutes

## 2. Client

- [x] 2.1 Draft logic (`draft.ts`): turns, veterans first, Red's reveals, 15 s per turn with random fill
- [x] 2.2 Store: start the battle, then draft, then deploy; resume re-drafts an unlocked battle
- [x] 2.3 Draft screen; the picker removed
