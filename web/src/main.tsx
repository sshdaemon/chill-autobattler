import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

const root = createRoot(document.getElementById('root')!);
const params = new URLSearchParams(location.search);
const preview = import.meta.env.DEV ? params.get('preview') : null;
const lab = import.meta.env.DEV ? params.get('lab') : null; // '' for ?lab

// Dev only: ?lab[=<unit>] is the unit lab (inspect units, animations, effects);
// ?preview=<unit> shows one unit type's fixed preview steps for /new-unit's screenshots
// (see .claude/skills/new-unit). The game itself is loaded only otherwise:
// importing it starts the store, which starts a run.
if (import.meta.env.DEV && lab !== null) {
  void import('./ui/UnitLab.tsx').then(({ UnitLab }) => root.render(<StrictMode><UnitLab initial={lab} /></StrictMode>));
} else if (import.meta.env.DEV && preview) {
  void import('./ui/UnitPreview.tsx').then(({ UnitPreview }) => root.render(<StrictMode><UnitPreview name={preview} /></StrictMode>));
} else {
  void import('./App.tsx').then(({ App }) => root.render(<StrictMode><App /></StrictMode>));
}
