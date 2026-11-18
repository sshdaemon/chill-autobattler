import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

const root = createRoot(document.getElementById('root')!);
const preview = import.meta.env.DEV ? new URLSearchParams(location.search).get('preview') : null;

// Dev only: ?preview=<unit> shows one unit type for review instead of the game
// (see .claude/skills/new-unit). The game itself is loaded only otherwise:
// importing it starts the store, which starts a run.
if (import.meta.env.DEV && preview) {
  void import('./ui/UnitPreview.tsx').then(({ UnitPreview }) => root.render(<StrictMode><UnitPreview name={preview} /></StrictMode>));
} else {
  void import('./App.tsx').then(({ App }) => root.render(<StrictMode><App /></StrictMode>));
}
