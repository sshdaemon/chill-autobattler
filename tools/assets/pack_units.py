#!/usr/bin/env python3
"""Packs baked units (tools/blender/bake_units.py output) for the web app.

    python3 tools/assets/pack_units.py [BAKE_DIR] [CLIPS_JSON]

Defaults: BAKE_DIR = the prototype's ../web3d-demo/assets/units,
          CLIPS_JSON = $QUATERNIUS_DIR/clips.json (default: the prototype's copy).
Output:   web/public/assets/units/ and web/public/assets/clips.json

- Blue and Red of a look have byte-identical geometry: one <Look>.bin is kept
  and both .gltf files point at it.
- PNG textures (roughness maps) become JPEG via macOS `sips` (dev tool only).
- clips.json keeps only the clips the game plays.
"""
import hashlib
import json
import os
import shutil
import subprocess
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
PROTO = os.path.join(ROOT, '..', 'web3d-demo', 'assets')
BAKE = sys.argv[1] if len(sys.argv) > 1 else os.path.join(PROTO, 'units')
CLIPS = sys.argv[2] if len(sys.argv) > 2 else os.path.join(os.environ.get('QUATERNIUS_DIR', os.path.join(PROTO, 'quaternius')), 'clips.json')
OUT = os.path.join(ROOT, 'web', 'public', 'assets', 'units')
CLIPS_OUT = os.path.join(ROOT, 'web', 'public', 'assets', 'clips.json')
KEEP_CLIPS = ['Sword_Idle', 'Jog_Fwd_Loop', 'Sword_Attack', 'Spell_Simple_Shoot', 'Hit_Chest', 'Death01', 'Yes']


def sha(path):
    return hashlib.sha256(open(path, 'rb').read()).hexdigest()


def main():
    if os.path.exists(OUT):
        shutil.rmtree(OUT)
    os.makedirs(OUT)
    looks = sorted(f[:-len('_blue.gltf')] for f in os.listdir(BAKE) if f.endswith('_blue.gltf'))
    images = set()
    for look in looks:
        gl = {t: json.load(open(os.path.join(BAKE, f'{look}_{t}.gltf'))) for t in ('blue', 'red')}
        bins = [os.path.join(BAKE, g['buffers'][0]['uri']) for g in gl.values()]
        if sha(bins[0]) != sha(bins[1]):
            sys.exit(f'{look}: blue and red geometry differ; cannot share')
        shutil.copyfile(bins[0], os.path.join(OUT, f'{look}.bin'))
        for team, g in gl.items():
            g['buffers'][0]['uri'] = f'{look}.bin'
            for img in g.get('images', []):
                images.add(img['uri'])
                if img['uri'].endswith('.png'):
                    img['uri'] = img['uri'][:-4] + '.jpg'
                    img.pop('mimeType', None)
            json.dump(g, open(os.path.join(OUT, f'{look}_{team}.gltf'), 'w'), separators=(',', ':'))
    for uri in sorted(images):
        src = os.path.join(BAKE, uri)
        if uri.endswith('.png'):
            dst = os.path.join(OUT, uri[:-4] + '.jpg')
            subprocess.run(['sips', '-s', 'format', 'jpeg', '-s', 'formatOptions', '85', src, '--out', dst], check=True, capture_output=True)
        else:
            shutil.copyfile(src, os.path.join(OUT, uri))
    clips = [c for c in json.load(open(CLIPS)) if c['name'] in KEEP_CLIPS]
    missing = set(KEEP_CLIPS) - {c['name'] for c in clips}
    if missing:
        sys.exit(f'clips missing: {missing}')
    json.dump(clips, open(CLIPS_OUT, 'w'), separators=(',', ':'))
    size = sum(os.path.getsize(os.path.join(OUT, f)) for f in os.listdir(OUT))
    print(f'packed {len(looks)} looks, {len(images)} textures, {len(clips)} clips: {size / 1e6:.1f} MB + clips {os.path.getsize(CLIPS_OUT) / 1e6:.1f} MB')


if __name__ == '__main__':
    main()
