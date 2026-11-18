#!/usr/bin/env python3
"""Packs the animation clips the game plays (web/public/assets/clips/*.json) from
Quaternius's Universal Animation Library 1 and 2 (CC0, Standard editions).

    python3 -I tools/assets/ual_clips.py [UAL_DIR]

UAL_DIR (default ~/fun/assets-src) holds ual1/ and ual2/ with the unzipped
packs. Which clips to keep comes from the looks (web/src/units/looks/*.json,
each file's "clips"); each look file gets its own bundle of the clips it lists
(_shared.json: clips several units use), so units' clips change separately. An
entry with "from": [A, B] is made by playing A then B (a combo strike and
its recovery, so the attack starts and ends at rest).

The UAL rig is the one our characters are baked on (same bones, same rest
pose: the clips we had match the UAL ones value for value); only 'root' is
our 'Rig'. Kept, like before: every bone's rotation, the pelvis position and
the root rotation; scale, other translations and leaf bones are dropped. Keys
are resampled to 30 fps, keys a straight line reproduces are removed and
values are rounded to 4 decimals.
"""
import glob, json, os, struct, sys

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
LOOKS = os.path.join(ROOT, 'web', 'src', 'units', 'looks')
OUT = os.path.join(ROOT, 'web', 'public', 'assets', 'clips')

FPS = 30
DEC = 4


def load(path):
    b = open(path, 'rb').read()
    n = struct.unpack('<I', b[12:16])[0]
    g = json.loads(b[20:20 + n])
    off = 20 + n
    m = struct.unpack('<I', b[off:off + 4])[0]
    return g, b[off + 8:off + 8 + m]


def accessor(g, binary, i):
    a = g['accessors'][i]
    v = g['bufferViews'][a['bufferView']]
    width = {'SCALAR': 1, 'VEC3': 3, 'VEC4': 4}[a['type']]
    assert a['componentType'] == 5126, 'float accessors only'
    start = v.get('byteOffset', 0) + a.get('byteOffset', 0)
    stride = v.get('byteStride', 4 * width)
    out = []
    for k in range(a['count']):
        out.append(struct.unpack_from('<%df' % width, binary, start + k * stride))
    return out


def rename(name):
    return 'Rig' if name == 'root' else name


def sample(times, values, t, quat):
    if t <= times[0]:
        return values[0]
    if t >= times[-1]:
        return values[-1]
    lo = 0
    while times[lo + 1] < t:
        lo += 1
    a, b = values[lo], values[lo + 1]
    k = (t - times[lo]) / (times[lo + 1] - times[lo])
    if quat:
        dot = sum(x * y for x, y in zip(a, b))
        if dot < 0:
            b = [-x for x in b]
        q = [x + (y - x) * k for x, y in zip(a, b)]
        n = sum(x * x for x in q) ** 0.5
        return [x / n for x in q]
    return [x + (y - x) * k for x, y in zip(a, b)]


def dense(g, binary, anim):
    """Every kept channel resampled at FPS: {track name: (quat, [values per frame])}."""
    nodes = g['nodes']
    out, dur = {}, 0
    for ch in anim['channels']:
        dur = max(dur, accessor(g, binary, anim['samplers'][ch['sampler']]['input'])[-1][0])
    n = max(1, round(dur * FPS))
    for ch in anim['channels']:
        node = nodes[ch['target']['node']]['name']
        path = ch['target']['path']
        if path == 'scale' or (path == 'translation' and node != 'pelvis') or '_leaf' in node:
            continue  # leaf bones carry no skin; our rig names them differently anyway
        s = anim['samplers'][ch['sampler']]
        times = [t[0] for t in accessor(g, binary, s['input'])]
        values = accessor(g, binary, s['output'])
        quat = path == 'rotation'
        name = rename(node) + ('.quaternion' if quat else '.position')
        out[name] = (quat, [sample(times, values, i / FPS, quat) for i in range(n + 1)])
    return out, n


def concat(parts):
    """Plays dense clips back to back (each next part starts on the last frame of the previous)."""
    names = set(parts[0][0])
    for d, _ in parts[1:]:
        names &= set(d)
    out, total = {}, 0
    for k in names:
        quat, vs = parts[0][0][k][0], list(parts[0][0][k][1])
        for d, _ in parts[1:]:
            vs += d[k][1][1:]
        out[k] = (quat, vs)
    total = sum(n for _, n in parts)
    return out, total


def simplify(name, tracks, n):
    out = []
    for track, (quat, vs) in sorted(tracks.items()):
        ts = [i / FPS for i in range(n + 1)]
        keep = [0]  # drop keys a straight line between neighbours reproduces
        for i in range(1, n):
            mid = [p + (q - p) * 0.5 for p, q in zip(vs[keep[-1]], vs[i + 1])] if i - keep[-1] == 1 else None
            if mid is None or max(abs(x - y) for x, y in zip(mid, vs[i])) > 2e-4:
                keep.append(i)
        if n > 0:
            keep.append(n)
        if all(max(abs(x - y) for x, y in zip(vs[0], vs[i])) < 2e-4 for i in keep):
            keep = [0]  # constant: one key
        out.append({
            'name': track, 'type': 'quaternion' if quat else 'vector',
            'times': [round(ts[i], DEC) for i in keep],
            'values': [round(x, DEC) for i in keep for x in vs[i]],
        })
    return {'name': name, 'duration': round(n / FPS, DEC), 'tracks': out, 'uuid': 'ual-' + name, 'blendMode': 2500}


def main():
    src = sys.argv[1] if len(sys.argv) > 1 else os.path.expanduser('~/fun/assets-src')
    sources = sorted(glob.glob(os.path.join(src, 'ual1', '**', 'Unreal-Godot', '*_Standard.glb'), recursive=True)) + \
        sorted(glob.glob(os.path.join(src, 'ual2', '**', 'Unreal-Godot', '*_Standard.glb'), recursive=True))
    if len(sources) != 2:
        sys.exit(f'expected UAL1 and UAL2 Standard .glb under {src}/ual1 and ual2, found {sources}')
    raw = {}
    for path in sources:
        g, binary = load(path)
        for anim in g.get('animations', []):
            raw.setdefault(anim['name'], (g, binary, anim))
    cache = {}

    def get(name):
        if name not in raw:
            sys.exit(f'clip not found in the packs: {name}')
        if name not in cache:
            cache[name] = dense(*raw[name])
        return cache[name]

    os.makedirs(OUT, exist_ok=True)
    seen = {}
    for look in sorted(glob.glob(os.path.join(LOOKS, '*.json'))):
        bundle = os.path.basename(look)[:-len('.json')]
        clips = []
        for name, spec in sorted(json.load(open(look)).get('clips', {}).items()):
            if name in seen:
                sys.exit(f'clip {name} is listed in both {seen[name]} and {bundle}: keep it in one (shared clips in _shared.json)')
            seen[name] = bundle
            parts = spec.get('from', [name])
            tracks, n = concat([get(p) for p in parts]) if len(parts) > 1 else get(parts[0])
            clips.append(simplify(name, tracks, n))
        out = os.path.join(OUT, bundle + '.json')
        json.dump(clips, open(out, 'w'), separators=(',', ':'))
        print(f'{len(clips)} clips -> {out}: {os.path.getsize(out) / 1e6:.2f} MB')


if __name__ == '__main__':
    main()
