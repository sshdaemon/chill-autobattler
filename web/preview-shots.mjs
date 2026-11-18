// Dev server only: the unit preview (?preview=<unit>) posts its screenshots
// here, and they land in design/units/<unit>/preview/ for /new-unit's PR.
// Names are checked so nothing can be written outside that folder.
import { mkdirSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '..');

export function previewShots() {
  return {
    name: 'preview-shots',
    apply: 'serve',
    configureServer(server) {
      server.middlewares.use('/__preview-shot', (req, res) => {
        const m = /^\/([a-z0-9-]+)\/([a-z0-9._-]+\.png)$/i.exec(req.url ?? '');
        if (req.method !== 'POST' || !m) { res.statusCode = 400; res.end('POST /__preview-shot/<unit>/<file>.png'); return; }
        const chunks = [];
        req.on('data', (c) => chunks.push(c));
        req.on('end', () => {
          const file = resolve(ROOT, 'design', 'units', m[1], 'preview', m[2]);
          mkdirSync(dirname(file), { recursive: true });
          writeFileSync(file, Buffer.concat(chunks));
          res.end(file);
        });
      });
    },
  };
}
