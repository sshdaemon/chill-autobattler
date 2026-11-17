# Bakes the battle units into ready-to-load glTF files.
#   blender --background --python tools/blender/bake_units.py            -> tools/blender/out/*.gltf
#   blender --background --python tools/blender/bake_units.py -- preview -> also renders previews
#   python3 tools/assets/pack_units.py tools/blender/out                -> web/public/assets/units
# Each file = outfit + head (cut at the neck) + hair, all skinned to the outfit's
# skeleton, hair tinted, team texture applied; mages get purple robes and the hat.
# Weapons stay in code (battle.html) because they swap per class at spawn.
import bpy, math, os, sys

# Inputs: the Quaternius CC0 source packs (not in this repo; set QUATERNIUS_DIR,
# default: the prototype's copy). Output: raw bake (BAKE_OUT), which
# tools/assets/pack_units.py then packs into web/public/assets/units.
HERE = os.path.dirname(os.path.abspath(__file__))
SRC = os.environ.get('QUATERNIUS_DIR', os.path.join(HERE, '..', '..', '..', 'web3d-demo', 'assets', 'quaternius'))
OUT = os.environ.get('BAKE_OUT', os.path.join(HERE, 'out'))
ROOT = HERE  # previews go to tools/blender/previews
PREVIEW = 'preview' in sys.argv[sys.argv.index('--') + 1:] if '--' in sys.argv else False

ROSTER = [
    dict(name='Mirko', outfit='Male_Peasant.gltf', head='Superhero_Male_FullBody.gltf', extras=['Hair_SimpleParted.gltf', 'Hair_Beard.gltf'], hair=0xb8b2a8),
    dict(name='Vesna', outfit='Female_Peasant.gltf', head='Superhero_Female_FullBody.gltf', extras=['Hair_Buns.gltf'], hair=0x6b4325),
    dict(name='Bogdan', outfit='Male_Ranger.gltf', head='Superhero_Male_FullBody.gltf', extras=['Hair_Beard.gltf'], hair=0x3a2a1c),
    dict(name='Lada', outfit='Female_Ranger.gltf', head='Superhero_Female_FullBody.gltf', extras=[], hair=0x7a3b1e),
    dict(name='Radim', outfit='Male_Peasant.gltf', head='Superhero_Male_FullBody.gltf', extras=['Hair_SimpleParted.gltf'], hair=0xc8a165),
]
by = {d['name']: d for d in ROSTER}
# mages reuse the peasant bodies (no mage outfit in the free packs)
UNITS = ROSTER + [dict(by[b], name=n, mage=True) for n, b in [('Zoran', 'Radim'), ('Milena', 'Vesna'), ('Velimir', 'Mirko')]]
ALT_TEXTURES = {'MI_Peasant': 'T_Peasant_2_BaseColor.jpg', 'MI_Ranger': 'T_Ranger_3_BaseColor.jpg'}
ROBE = 0x8f74ff
HAT = 0x4a2a8a


def lin(hexc):  # sRGB hex -> linear RGBA (what three's Color.setHex stores)
    c = [((hexc >> s) & 255) / 255 for s in (16, 8, 0)]
    return tuple(x / 12.92 if x <= 0.04045 else ((x + 0.055) / 1.055) ** 2.4 for x in c) + (1.0,)


def import_gltf(f):
    before = set(bpy.data.objects)
    bpy.ops.import_scene.gltf(filepath=os.path.join(SRC, f), bone_heuristic='TEMPERANCE')
    return [o for o in bpy.data.objects if o not in before]


def base_color_link(mat):
    bsdf = next(n for n in mat.node_tree.nodes if n.type == 'BSDF_PRINCIPLED')
    return bsdf, bsdf.inputs['Base Color']


def tint(mat, rgba):  # texture x colour, exported as baseColorFactor
    nt = mat.node_tree
    bsdf, sock = base_color_link(mat)
    if not sock.links:
        sock.default_value = rgba
        return
    src = sock.links[0].from_socket
    mix = nt.nodes.new('ShaderNodeMix'); mix.data_type = 'RGBA'; mix.blend_type = 'MULTIPLY'
    mix.inputs['Factor'].default_value = 1.0
    nt.links.new(src, mix.inputs[6]); mix.inputs[7].default_value = rgba
    nt.links.new(mix.outputs[2], sock)


def swap_texture(mat, file):
    for node in mat.node_tree.nodes:
        if node.type == 'TEX_IMAGE' and node.image and 'BaseColor' in node.image.name:
            node.image = bpy.data.images.load(os.path.join(SRC, file), check_existing=True)


def adopt(objs, arm, cut_z=None):
    """Re-skin imported meshes onto `arm`; optionally keep only the part above cut_z."""
    for o in objs:
        if o.type != 'MESH':
            continue
        if cut_z is not None and o.name.lower().startswith('superhero_'):
            mw, me = o.matrix_world, o.data
            below = [(mw @ v.co).z <= cut_z for v in me.vertices]
            import bmesh
            bm = bmesh.new(); bm.from_mesh(me)
            bmesh.ops.delete(bm, geom=[f for f in bm.faces if any(below[v.index] for v in f.verts)], context='FACES')
            bmesh.ops.delete(bm, geom=[v for v in bm.verts if not v.link_faces], context='VERTS')
            bm.to_mesh(me); bm.free()
        mw = o.matrix_world.copy()
        o.parent = arm; o.matrix_world = mw
        for m in o.modifiers:
            if m.type == 'ARMATURE':
                m.object = arm
    for o in objs:
        if o.type in ('ARMATURE', 'EMPTY'):
            bpy.data.objects.remove(o)


def add_hat(arm):
    mat = bpy.data.materials.new('MI_MageHat'); mat.use_nodes = True
    bsdf, sock = base_color_link(mat); sock.default_value = lin(HAT); bsdf.inputs['Roughness'].default_value = 0.7
    head = arm.matrix_world @ arm.data.bones['Head'].head_local
    bpy.ops.mesh.primitive_cylinder_add(vertices=20, radius=0.19, depth=0.02)
    brim = bpy.context.object
    bpy.ops.mesh.primitive_cone_add(vertices=16, radius1=0.11, depth=0.42, location=(0, 0, 0.22), rotation=(0, 0.15, 0))
    cone = bpy.context.object
    for o in (brim, cone):
        o.data.materials.append(mat)
    bpy.ops.object.select_all(action='DESELECT')
    brim.select_set(True); cone.select_set(True); bpy.context.view_layer.objects.active = brim
    bpy.ops.object.join()
    hat = bpy.context.object; hat.name = 'MageHat'
    hat.location = (head.x, head.y + 0.02, head.z + 0.24)  # three: y+0.24, z-0.02 (Blender is Z-up, -Y forward)
    bpy.context.view_layer.update()
    mw = hat.matrix_world.copy()
    hat.parent = arm; hat.parent_type = 'BONE'; hat.parent_bone = 'Head'
    hat.matrix_world = mw


def build(u, team):
    bpy.ops.wm.read_factory_settings(use_empty=True)
    outfit = import_gltf(u['outfit'])
    arm = next(o for o in outfit if o.type == 'ARMATURE')
    arm.name = 'Armature'  # NOT 'Rig': clips.json has a 'Rig.quaternion' track that would tip the unit over
    neck = (arm.matrix_world @ arm.data.bones['neck_01'].head_local).z
    adopt(import_gltf(u['head']), arm, cut_z=neck - 0.04)
    for e in u['extras']:
        adopt(import_gltf(e), arm)
    for mat in bpy.data.materials:
        if not mat.use_nodes:
            continue
        key = mat.name.split('.')[0]
        if key.startswith('MI_Hair'):
            tint(mat, lin(u['hair']))
        if team == 'red' and key in ALT_TEXTURES:
            swap_texture(mat, ALT_TEXTURES[key])
        if u.get('mage') and key == 'MI_Peasant':
            tint(mat, lin(ROBE))
    if u.get('mage'):
        add_hat(arm)
    return arm


def export(path):
    bpy.ops.export_scene.gltf(filepath=path, export_format='GLTF_SEPARATE', export_texture_dir='',
                              export_animations=False, export_image_format='AUTO', export_yup=True)


def restore_rest(path, outfit):
    """Blender re-orients bones on import, so the exported joints no longer match the
    rest pose the shared animation clips (clips.json) were made for. Put the outfit's
    original joint transforms back and fix inverse bind matrices and bone-attached
    children so the mesh stays exactly where it was."""
    import json, struct
    from mathutils import Matrix, Quaternion, Vector
    src = {n['name']: n for n in json.load(open(os.path.join(SRC, outfit)))['nodes']}
    g = json.load(open(path)); N = g['nodes']
    parent = {c: i for i, n in enumerate(N) for c in n.get('children', [])}

    def local(n):
        t = Vector(n.get('translation', (0, 0, 0)))
        r = n.get('rotation', (0, 0, 0, 1)); q = Quaternion((r[3], r[0], r[1], r[2]))
        s = n.get('scale', (1, 1, 1))
        return Matrix.Translation(t) @ q.to_matrix().to_4x4() @ Matrix.Diagonal((*s, 1))

    def worlds():
        out = {}
        def w(i):
            if i not in out:
                out[i] = (w(parent[i]) @ local(N[i])) if i in parent else local(N[i])
            return out[i]
        for i in range(len(N)): w(i)
        return out

    old = worlds()
    joints = set(g['skins'][0]['joints'])
    for i in joints:
        o = src[N[i]['name']]
        for k in ('translation', 'rotation', 'scale'):
            N[i].pop(k, None)
            if k in o: N[i][k] = o[k]
    new = worlds()
    for i, n in enumerate(N):  # bone-attached props (mage hat): keep world placement
        if i not in joints and parent.get(i) in joints:
            m = new[parent[i]].inverted() @ old[i]
            t, q, s = m.decompose()
            n['translation'] = list(t); n['rotation'] = [q.x, q.y, q.z, q.w]; n['scale'] = list(s)
    bin_path = os.path.join(os.path.dirname(path), g['buffers'][0]['uri'])
    data = bytearray(open(bin_path, 'rb').read())
    for skin in g['skins']:
        acc = g['accessors'][skin['inverseBindMatrices']]; bv = g['bufferViews'][acc['bufferView']]
        off = bv.get('byteOffset', 0) + acc.get('byteOffset', 0)
        for k, j in enumerate(skin['joints']):
            p = off + k * 64
            vals = struct.unpack_from('<16f', data, p)
            ibm = Matrix([vals[c::4] for c in range(4)])  # column-major -> rows
            ibm = new[j].inverted() @ old[j] @ ibm
            struct.pack_into('<16f', data, p, *[ibm[r][c] for c in range(4) for r in range(4)])
    open(bin_path, 'wb').write(data)
    json.dump(g, open(path, 'w'), indent=1)


def render(path):
    s = bpy.context.scene
    cam = bpy.data.objects.new('cam', bpy.data.cameras.new('cam')); s.collection.objects.link(cam)
    cam.data.lens = 70
    cam.location = (1.9, -4.6, 1.35); cam.rotation_euler = (math.radians(86), 0, math.radians(22))
    s.camera = cam
    sun = bpy.data.objects.new('sun', bpy.data.lights.new('sun', 'SUN')); sun.data.energy = 4
    sun.rotation_euler = (0.7, 0.2, 0.5); s.collection.objects.link(sun)
    w = bpy.data.worlds.new('w'); w.color = (0.45, 0.55, 0.65); s.world = w
    s.render.resolution_x, s.render.resolution_y = 300, 420
    s.render.filepath = path
    bpy.ops.render.render(write_still=True)


os.makedirs(OUT, exist_ok=True)
for u in UNITS:
    for team in ('blue', 'red'):
        build(u, team)
        name = f"{u['name']}_{team}"
        export(os.path.join(OUT, name + '.gltf'))
        restore_rest(os.path.join(OUT, name + '.gltf'), u['outfit'])
        if PREVIEW:
            build(u, team)  # export can leave the scene in a converted state; rebuild for the shot
            render(os.path.join(ROOT, 'previews', name + '.png'))
        print('BAKED', name)
