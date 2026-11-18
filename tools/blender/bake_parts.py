# Bakes the character kit: outfit parts, heads, hair and items for runtime
# assembly (web/src/scene/models.ts), plus parts.json and the team masks.
#
#   blender --background --factory-startup --python tools/blender/bake_parts.py [-- SRC]
#
# SRC (default ~/fun/assets-src) holds the unzipped CC0 packs:
#   outfits/   Quaternius "Modular Character Outfits - Fantasy" (Standard)
#   basechars/ Quaternius "Universal Base Characters" (Standard)
#   items/     Quaternius "Fantasy Props MegaKit" (Standard) and KayKit "Adventurers" (Free)
# Output: web/public/assets/parts/ (glTF + shared JPEG textures) and parts.json.
#
# Skinned parts keep their glTF nodes untouched (so the rest pose stays the
# one the clips play on); only meshes are rewritten: unused vertex colours
# and extra UV sets dropped, unreferenced vertices removed, and the base
# bodies cut down to the head (outfits bring their own bodies and arms).
# Blender is used only for images (resize, JPEG, masks) and for the few items
# modelled here (wizard hat, helmet).
import bpy, glob, json, math, os, struct, sys
import numpy as np

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(os.path.dirname(HERE))
ARGS = sys.argv[sys.argv.index('--') + 1:] if '--' in sys.argv else []
SRC = ARGS[0] if ARGS else os.path.expanduser('~/fun/assets-src')
OUT = os.path.join(ROOT, 'web', 'public', 'assets', 'parts')

OUTFITS = glob.glob(os.path.join(SRC, 'outfits', '**', 'glTF*', 'Modular Parts'), recursive=True)[0]
OUTFIT_TEX = glob.glob(os.path.join(SRC, 'outfits', '**', 'Textures'), recursive=True)[0]
BASES = glob.glob(os.path.join(SRC, 'basechars', '**', 'Base Characters', 'Godot - UE'), recursive=True)[0]
HAIRS = glob.glob(os.path.join(SRC, 'basechars', '**', 'Rigged to Head Bone', '*'), recursive=True)[0]
PROPS = os.path.dirname(glob.glob(os.path.join(SRC, 'items', 'FantasyProps*', '**', 'Sword_Bronze.gltf'), recursive=True)[0])
KAYKIT = os.path.dirname(glob.glob(os.path.join(SRC, 'items', 'KayKit*', '**', 'bow_withString.gltf'), recursive=True)[0])

# slot of each outfit part (body = torso)
SLOTS = {'Body': 'torso', 'Arms': 'arms', 'Legs': 'legs', 'Feet': 'feet', 'Feet_Boots': 'feet',
         'Head_Hood': 'headgear', 'Acc_Pauldron': 'accessory', 'Acc_Pauldrons': 'accessory'}
HAIR = {'male': ['Hair_SimpleParted', 'Hair_Beard', 'Hair_Buzzed', 'Eyebrows_Regular'],
        'female': ['Hair_Buns', 'Hair_Long', 'Hair_BuzzedFemale', 'Eyebrows_Female']}
# Items, one file per unit (or _shared.json for kit several units use) in
# tools/blender/items/*.json, so sessions working on different units edit
# different files. Each item: its attach bone and mount pose in the
# character's own space at rest (Y up, facing +Z, its left is +X): pos, offset
# from the bone's origin (metres); rot, radians (XYZ); scale. The client places
# the item there, then attaches it to the bone keeping that placement (tuned in
# the browser against idle and attack clips). Kinds:
#   "source": [pack, file]: a rigid prop from a pack ("props" or "kaykit");
#   "modelled": true: built from simple shapes in model_items() below;
#   "kind": "cloth": generated and simulated by the client (web/src/scene/cloth.ts),
#     no mesh here; width (top, hem) and length in metres; team: team-coloured.
PACKS = {'props': PROPS, 'kaykit': KAYKIT}


def load_items():
    items, owner = {}, {}
    for path in sorted(glob.glob(os.path.join(HERE, 'items', '*.json'))):
        for name, spec in json.load(open(path)).items():
            if name in owner:
                sys.exit(f'item {name} is in both {owner[name]} and {os.path.basename(path)}')
            items[name], owner[name] = spec, os.path.basename(path)
    return items


ITEMS = load_items()
POSE_KEYS = ('pos', 'rot', 'scale')


# team masks: outfit -> (variant base colours, rule); 'cloth' also drops pixels
# that are brownish in every variant (leather and wood recolour between variants too)
MASKS = {'Ranger': (['T_Ranger_BaseColor', 'T_Ranger_3_BaseColor'], 'cloth'),
         'Peasant': (['T_Peasant_BaseColor', 'T_Peasant_2_BaseColor'], 'diff')}
TEX_SIZE = {'BaseColor': 1024, 'Dark': 1024}  # anything else: 512 (props' trim sheets: see bake_textures)
KEEP_ATTRS = ('POSITION', 'NORMAL', 'TEXCOORD_0', 'JOINTS_0', 'WEIGHTS_0')
COMP = {5120: 'b', 5121: 'B', 5122: 'h', 5123: 'H', 5125: 'I', 5126: 'f'}
WIDTH = {'SCALAR': 1, 'VEC2': 2, 'VEC3': 3, 'VEC4': 4, 'MAT4': 16}

textures = {}  # source image path -> output name


def read(path):
    g = json.load(open(path))
    data = open(os.path.join(os.path.dirname(path), g['buffers'][0]['uri']), 'rb').read()
    return g, data


def acc(g, data, i):
    a = g['accessors'][i]; v = g['bufferViews'][a['bufferView']]
    n, c = WIDTH[a['type']], COMP[a['componentType']]
    size = struct.calcsize('<' + c) * n
    stride = v.get('byteStride', size)
    off = v.get('byteOffset', 0) + a.get('byteOffset', 0)
    return [struct.unpack_from('<%d%s' % (n, c), data, off + k * stride) for k in range(a['count'])], a


def mat4(vals):  # column-major glTF -> row lists
    return [[vals[c * 4 + r] for c in range(4)] for r in range(4)]


def inv_translation(ibm):  # joint origin in mesh space = inverse(IBM) applied to the origin
    m = np.array(mat4(ibm), dtype=np.float64)
    return np.linalg.inv(m)[:3, 3]


def rewrite(g, data, keep_tri=None):
    """New buffer with only KEEP_ATTRS, referenced vertices, and keep_tri(prim, tri) filtering."""
    out = bytearray(); views, accessors = [], []
    quantized = [False]

    def push(values, comp, typ, target, minmax=False, normalized=False):
        nonlocal out
        while len(out) % 4: out.append(0)
        fmt = '<%d%s' % (WIDTH[typ], COMP[comp])
        size = struct.calcsize(fmt)
        stride = (size + 3) // 4 * 4 if target == 34962 else size  # vertex elements start on 4-byte boundaries
        start = len(out)
        for v in values:
            out += struct.pack(fmt, *v) + bytes(stride - size)
        view = {'buffer': 0, 'byteOffset': start, 'byteLength': len(out) - start, **({'target': target} if target else {})}
        if stride != size:
            view['byteStride'] = stride
        views.append(view)
        a = {'bufferView': len(views) - 1, 'componentType': comp, 'count': len(values), 'type': typ, **({'normalized': True} if normalized else {})}
        if minmax:
            a['min'] = [min(v[k] for v in values) for k in range(WIDTH[typ])]
            a['max'] = [max(v[k] for v in values) for k in range(WIDTH[typ])]
        if comp == 5121 and typ == 'VEC4' and target == 34962 and values and False:
            pass
        accessors.append(a)
        return len(accessors) - 1

    for mesh in g['meshes']:
        prims = []
        for p in mesh['primitives']:
            idx = [t[0] for t in acc(g, data, p['indices'])[0]]
            tris = [idx[k:k + 3] for k in range(0, len(idx), 3)]
            attrs = {k: acc(g, data, v) for k, v in p['attributes'].items() if k in KEEP_ATTRS}
            if keep_tri:
                pos = attrs['POSITION'][0]
                tris = [t for t in tris if keep_tri(mesh, [pos[i] for i in t])]
            if not tris:
                continue
            used = sorted({i for t in tris for i in t})
            remap = {o: n for n, o in enumerate(used)}
            np_ = {'attributes': {}, **({'material': p['material']} if 'material' in p else {})}
            for k, (vals, a) in attrs.items():
                vs = [vals[i] for i in used]
                comp, norm = a['componentType'], a.get('normalized', False)
                if comp == 5126 and k == 'WEIGHTS_0':  # 4 bytes instead of 16; re-normalised so they still sum to 1
                    vs = [quant_weights(v) for v in vs]; comp, norm = 5121, True
                elif comp == 5126 and k == 'NORMAL':  # KHR_mesh_quantization
                    vs = [tuple(max(-127, min(127, round(x * 127))) for x in v) for v in vs]; comp, norm = 5120, True
                    quantized[0] = True
                elif comp == 5126 and k == 'TEXCOORD_0' and all(0 <= x <= 1 for v in vs for x in v):
                    vs = [tuple(round(x * 65535) for x in v) for v in vs]; comp, norm = 5123, True
                np_['attributes'][k] = push(vs, comp, a['type'], 34962, minmax=(k == 'POSITION'), normalized=norm)
            flat = [(remap[i],) for t in tris for i in t]
            np_['indices'] = push(flat, 5125 if len(used) > 65535 else 5123, 'SCALAR', 34963)
            prims.append(np_)
        mesh['primitives'] = prims
    # skins' inverse bind matrices
    for skin in g.get('skins', []):
        vals, a = acc(g, data, skin['inverseBindMatrices'])
        skin['inverseBindMatrices'] = push(vals, 5126, 'MAT4', None)
    # drop meshes that lost every primitive (and nodes pointing at them)
    for n in g['nodes']:
        if 'mesh' in n and not g['meshes'][n['mesh']]['primitives']:
            del n['mesh']; n.pop('skin', None)
    g['bufferViews'], g['accessors'] = views, accessors
    g.pop('animations', None)  # a few source files carry a stray pose clip; the game uses clips.json
    if quantized[0]:
        g['extensionsUsed'] = g['extensionsRequired'] = ['KHR_mesh_quantization']
    return g, bytes(out)


def quant_weights(w):
    q = [round(x * 255) for x in w]
    q[q.index(max(q))] += 255 - sum(q)  # keep the sum at exactly 255
    return tuple(q)


def tex_name(src):
    base = os.path.splitext(os.path.basename(src))[0].replace('_png', '')
    return base + '.jpg'


def remap_images(g, src_dir, extra_dirs=()):
    for img in g.get('images', []):
        uri = img['uri']
        path = next((p for p in [os.path.join(d, uri) for d in (src_dir, *extra_dirs)] if os.path.exists(p)), None)
        for cand in (os.path.basename(uri), os.path.basename(uri).replace('_png.png', '.png')):
            if path is None:  # some files name textures that sit elsewhere, or as X_png.png for X.png
                hits = glob.glob(os.path.join(SRC, '**', cand), recursive=True)
                path = hits[0] if hits else None
        if path is None:
            raise SystemExit(f'texture not found: {uri}')
        name = tex_name(path)
        textures[path] = name
        img['uri'] = name
        img.pop('mimeType', None)


def write(name, g, data):
    g['buffers'] = [{'uri': name + '.bin', 'byteLength': len(data)}]
    for k in ('extensionsUsed', 'extensionsRequired'):
        g[k] = [e for e in g.get(k, []) if e == 'KHR_mesh_quantization']
        if not g[k]:
            del g[k]
    open(os.path.join(OUT, name + '.bin'), 'wb').write(data)
    json.dump(g, open(os.path.join(OUT, name + '.gltf'), 'w'), separators=(',', ':'))


def skinned_part(src, name, keep_tri=None):
    g, data = read(src)
    g, data = rewrite(g, data, keep_tri)
    remap_images(g, os.path.dirname(src), [os.path.join(OUTFIT_TEX, d) for d in os.listdir(OUTFIT_TEX)])
    write(name, g, data)
    return [m['name'] for m in g.get('materials', [])]


def head_part(src, name):
    """A base body cut to the head: drops every triangle with a vertex more than
    4 cm below the neck joint (measured along pelvis -> head)."""
    g, data = read(src)
    skin = g['skins'][0]
    ibms, _ = acc(g, data, skin['inverseBindMatrices'])
    pos = {g['nodes'][j]['name']: inv_translation(ibms[k]) for k, j in enumerate(skin['joints'])}
    up = pos['Head'] - pos['pelvis']; up /= np.linalg.norm(up)
    neck = pos['neck_01']

    def keep(mesh, verts):
        if mesh['name'].lower().startswith(('face', 'eye')):
            return True
        return all(float(np.dot(np.array(v) - neck, up)) > -0.04 for v in verts)
    return skinned_part(src, name, keep)


def rigid_part(src, name):
    """An item: base colour only (normal and roughness maps don't show on something this small)."""
    g, data = read(src)
    g, data = rewrite(g, data)
    keep = set()
    for m in g.get('materials', []):
        m.pop('normalTexture', None); m.pop('occlusionTexture', None)
        pbr = m.get('pbrMetallicRoughness', {})
        pbr.pop('metallicRoughnessTexture', None)
        if 'baseColorTexture' in pbr:
            keep.add(pbr['baseColorTexture']['index'])
    # drop unused textures and images, renumbering the base colour references
    order = sorted(keep)
    tex = [g['textures'][i] for i in order]
    img_order = sorted({t['source'] for t in tex})
    for t in tex:
        t['source'] = img_order.index(t['source'])
    g['images'] = [g['images'][i] for i in img_order]
    g['textures'] = tex
    for m in g.get('materials', []):
        pbr = m.get('pbrMetallicRoughness', {})
        if 'baseColorTexture' in pbr:
            pbr['baseColorTexture']['index'] = order.index(pbr['baseColorTexture']['index'])
    for k in ('textures', 'images', 'samplers'):
        if k in g and not g[k]:
            del g[k]
    remap_images(g, os.path.dirname(src))
    write(name, g, data)


def model_items():
    """Items modelled here: exported as static glTF in the attach bone's frame."""
    def mat(name, rgb, rough=0.7, metal=0.0):
        m = bpy.data.materials.new(name); m.use_nodes = True
        b = m.node_tree.nodes['Principled BSDF']
        b.inputs['Base Color'].default_value = (*rgb, 1); b.inputs['Roughness'].default_value = rough; b.inputs['Metallic'].default_value = metal
        return m

    def export(name, objs):
        bpy.ops.object.select_all(action='DESELECT')
        for o in objs:
            o.select_set(True)
        bpy.context.view_layer.objects.active = objs[0]
        if len(objs) > 1:
            bpy.ops.object.join()
        bpy.ops.export_scene.gltf(filepath=os.path.join(OUT, name + '.gltf'), export_format='GLTF_SEPARATE', use_selection=True, export_yup=True)

    def lin(h):
        c = [((h >> s) & 255) / 255 for s in (16, 8, 0)]
        return tuple(x / 12.92 if x <= 0.04045 else ((x + 0.055) / 1.055) ** 2.4 for x in c)

    bpy.ops.wm.read_factory_settings(use_empty=True)
    hat = mat('MI_WizardHat', lin(0x4a2a8a))
    bpy.ops.mesh.primitive_cylinder_add(vertices=24, radius=0.2, depth=0.02, location=(0, 0, 0))
    brim = bpy.context.object
    bpy.ops.mesh.primitive_cone_add(vertices=20, radius1=0.12, depth=0.44, location=(0, 0.03, 0.22), rotation=(math.radians(-14), 0, 0))
    cone = bpy.context.object
    for o in (brim, cone):
        o.data.materials.append(hat)
    export('Wizard_Hat', [brim, cone])

    bpy.ops.wm.read_factory_settings(use_empty=True)
    steel = mat('MI_Helmet', lin(0x9aa0a8), 0.35, 0.8)
    bpy.ops.mesh.primitive_uv_sphere_add(segments=24, ring_count=12, radius=0.13, location=(0, 0, 0))
    dome = bpy.context.object
    bpy.ops.object.mode_set(mode='EDIT'); import bmesh
    bm = bmesh.from_edit_mesh(dome.data)
    bmesh.ops.delete(bm, geom=[v for v in bm.verts if v.co.z < -0.02], context='VERTS')
    bmesh.update_edit_mesh(dome.data); bpy.ops.object.mode_set(mode='OBJECT')
    bpy.ops.mesh.primitive_torus_add(major_radius=0.13, minor_radius=0.012, location=(0, 0, -0.015))
    rim = bpy.context.object
    bpy.ops.mesh.primitive_cube_add(size=1, location=(0, -0.13, -0.05), scale=(0.012, 0.01, 0.06))
    nose = bpy.context.object
    for o in (dome, rim, nose):
        o.data.materials.append(steel)
    export('Helmet', [dome, rim, nose])
    # spear: a long shaft (grip at the origin, point along +Y like the sword), steel head
    bpy.ops.wm.read_factory_settings(use_empty=True)
    wood = mat('MI_SpearShaft', lin(0x6b4a2a), 0.75)
    steel2 = mat('MI_SpearHead', lin(0xb8bec6), 0.3, 0.85)
    bpy.ops.mesh.primitive_cylinder_add(vertices=8, radius=0.018, depth=2.0, location=(0, 0, 0.55))
    shaft = bpy.context.object
    bpy.ops.mesh.primitive_cone_add(vertices=8, radius1=0.045, depth=0.24, location=(0, 0, 1.67))
    head = bpy.context.object
    head.scale = (1, 0.35, 1)
    bpy.ops.mesh.primitive_cylinder_add(vertices=8, radius=0.028, depth=0.06, location=(0, 0, 1.53))
    collar = bpy.context.object
    shaft.data.materials.append(wood)
    for o in (head, collar):
        o.data.materials.append(steel2)
    export('Spear', [shaft, head, collar])
    # great helm: a closed barrel helm with a flat top, an eye slit and a short crest
    bpy.ops.wm.read_factory_settings(use_empty=True)
    plate = mat('MI_GreatHelm', lin(0x8f969e), 0.32, 0.85)
    dark = mat('MI_GreatHelmSlit', lin(0x15171a), 0.9)
    bpy.ops.mesh.primitive_cylinder_add(vertices=20, radius=0.135, depth=0.26, location=(0, 0, 0.0))
    barrel = bpy.context.object
    bpy.ops.mesh.primitive_uv_sphere_add(segments=20, ring_count=8, radius=0.135, location=(0, 0, 0.13))
    top = bpy.context.object
    top.scale = (1, 1, 0.45)
    bpy.ops.mesh.primitive_cube_add(size=1, location=(0, -0.132, 0.03), scale=(0.2, 0.012, 0.022))
    slit = bpy.context.object
    bpy.ops.mesh.primitive_cube_add(size=1, location=(0, 0.0, 0.2), scale=(0.018, 0.2, 0.06))
    crest = bpy.context.object
    for o in (barrel, top, crest):
        o.data.materials.append(plate)
    slit.data.materials.append(dark)
    export('Great_Helm', [barrel, top, slit, crest])

    for name in ('Wizard_Hat', 'Helmet', 'Spear', 'Great_Helm'):
        g = json.load(open(os.path.join(OUT, name + '.gltf')))
        json.dump(g, open(os.path.join(OUT, name + '.gltf'), 'w'), separators=(',', ':'))


def save_jpeg(px, w, h, path, quality=85):
    im = bpy.data.images.new('out', w, h, alpha=False)
    im.pixels[:] = px.ravel().tolist()
    s = bpy.context.scene
    s.render.image_settings.file_format = 'JPEG'; s.render.image_settings.quality = quality
    s.render.image_settings.color_mode = 'RGB'
    im.save_render(path, scene=s)
    bpy.data.images.remove(im)


def load_px(path, size):
    im = bpy.data.images.load(path)
    im.colorspace_settings.name = 'Non-Color'  # keep the stored values as they are
    if im.size[0] != size:
        im.scale(size, size)
    px = np.array(im.pixels[:], dtype=np.float32).reshape(size, size, 4)
    bpy.data.images.remove(im)
    return px


def bake_textures():
    for src, name in sorted(textures.items(), key=lambda kv: kv[1]):
        size = next((s for k, s in TEX_SIZE.items() if k in name), 512)
        if 'Trim' in name:
            size = 512
        if 'Eye' in name or 'Hair' in name or 'texture' in name:
            size = 256 if 'texture' in name else 512
        px = load_px(src, size)
        save_jpeg(np.dstack([px[..., :3], np.ones(px.shape[:2], np.float32)]), size, size, os.path.join(OUT, name))
    # variant base colours not referenced by any part file
    for outfit, (variants, _) in MASKS.items():
        for v in variants:
            path = os.path.join(OUTFIT_TEX, outfit, v + '.png')
            if path not in textures:
                px = load_px(path, 1024)
                save_jpeg(np.dstack([px[..., :3], np.ones((1024, 1024), np.float32)]), 1024, 1024, os.path.join(OUT, v + '.jpg'))


def brownish(a):
    mx, mn = a.max(axis=2), a.min(axis=2)
    sat = np.where(mx > 0, (mx - mn) / np.maximum(mx, 1e-6), 0)
    r, g, b = a[..., 0], a[..., 1], a[..., 2]
    h = np.degrees(np.arctan2(np.sqrt(3) * (g - b), 2 * r - g - b)) % 360
    return ((h > 8) & (h < 55) & (sat > 0.25)) | (mx < 0.06)


def bake_masks():
    n = 256
    for outfit, (variants, rule) in MASKS.items():
        imgs = [load_px(os.path.join(OUTFIT_TEX, outfit, v + '.png'), n)[..., :3] for v in variants]
        d = np.zeros((n, n), np.float32)
        for i in range(len(imgs)):
            for j in range(i + 1, len(imgs)):
                d = np.maximum(d, np.abs(imgs[i] - imgs[j]).max(axis=2))
        m = np.clip((d - 0.04) * 6, 0, 1)
        if rule == 'cloth':
            leather = np.ones((n, n), bool)
            for a in imgs:
                leather &= brownish(a)
            m = m * (~leather)
        for _ in range(2):
            p = np.pad(m, 1, mode='edge'); m = sum(p[y:y + n, x:x + n] for y in range(3) for x in range(3)) / 9
        save_jpeg(np.dstack([m, m, m, np.ones_like(m)]), n, n, os.path.join(OUT, f'T_{outfit}_Mask.jpg'), 90)
        print(f'mask {outfit}: {float((m > 0.5).mean()):.0%} cloth')


def main():
    os.makedirs(OUT, exist_ok=True)
    for f in glob.glob(os.path.join(OUT, '*')):
        os.remove(f)
    manifest = {'parts': {}, 'items': {}, 'materials': {}}
    for path in sorted(glob.glob(os.path.join(glob.escape(OUTFITS), '*.gltf'))):  # pack folders are named '...[Standard]'
        name = os.path.splitext(os.path.basename(path))[0]  # Male_Ranger_Body
        sex, outfit, piece = name.split('_', 2)
        mats = skinned_part(path, name)
        manifest['parts'][name] = {'slot': SLOTS[piece], 'sex': sex.lower(), 'outfit': outfit, 'materials': mats}
    for sex in ('male', 'female'):
        src = os.path.join(BASES, f'Superhero_{sex.title()}_FullBody.gltf')
        name = f'Head_{sex.title()}'
        manifest['parts'][name] = {'slot': 'head', 'sex': sex, 'materials': head_part(src, name)}
        for h in HAIR[sex]:
            mats = skinned_part(os.path.join(HAIRS, h + '.gltf'), h)
            manifest['parts'][h] = {'slot': 'hair', 'sex': sex, 'materials': mats}
    pose = lambda spec: {k: spec[k] for k in POSE_KEYS if k in spec}
    for name, spec in ITEMS.items():
        if 'source' in spec:
            pack, file = spec['source']
            rigid_part(os.path.join(PACKS[pack], file), name)
            manifest['items'][name] = {'bone': spec['bone'], **pose(spec)}
    model_items()
    for name, spec in ITEMS.items():
        if spec.get('modelled'):
            manifest['items'][name] = {'bone': spec['bone'], **pose(spec)}
        elif spec.get('kind') == 'cloth':
            manifest['items'][name] = spec
    bake_textures()
    bake_masks()
    for outfit, (variants, _) in MASKS.items():
        manifest['materials'][f'MI_{outfit}'] = {'variants': [v + '.jpg' for v in variants], 'mask': f'T_{outfit}_Mask.jpg'}
    json.dump(manifest, open(os.path.join(OUT, 'parts.json'), 'w'), indent=1)
    total = sum(os.path.getsize(f) for f in glob.glob(os.path.join(OUT, '*')))
    print(f'baked {len(manifest["parts"])} parts, {len(manifest["items"])} items: {total / 1e6:.1f} MB')


main()
