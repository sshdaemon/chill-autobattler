#!/usr/bin/env python3
"""Packs the CC0 nature assets for the web app.

    python3 tools/assets/pack_nature.py

Sources (not in this repo; download them yourself):
- NATURE_DIR, default ../assets-src/quaternius-nature: Quaternius "Stylized
  Nature MegaKit [Standard]" (quaternius.itch.io/stylized-nature-megakit), unzipped.
- AMBIENTCG_DIR, default ../assets-src/ambientcg: Grass004, Ground037 and
  Ground054 at 1K-JPG from ambientcg.com, each unzipped into a folder of its name.

Output: web/public/assets/nature/
- nature.glb: the chosen models, one named root node each, textures downscaled.
- <Name>_color.jpg (1024 px) / <Name>_normal.jpg (512 px): ground textures.

Steps: copy the chosen glTF files to a temp folder with downscaled textures
(macOS `sips`), then run Blender headless (this same file) to export one GLB.
"""
import json
import os
import shutil
import subprocess
import sys
import tempfile

MODELS = [
    'Rock_Medium_1', 'Rock_Medium_2', 'Rock_Medium_3',
    'CommonTree_1', 'CommonTree_2', 'CommonTree_3',
    'Pine_1', 'Pine_2', 'Pine_5',
    'Bush_Common', 'Fern_1', 'Flower_3_Group', 'Pebble_Round_2',
]
# texture -> max size in px; leaves keep PNG for alpha, the rest become JPEG
TEX_SIZE = {'Bark': 512, 'Rocks': 512, 'PathRocks': 256}
GROUND = ['Grass004', 'Ground037', 'Ground054']


def in_blender():
    import bpy  # noqa: F401  (only available inside Blender)
    src, out = sys.argv[sys.argv.index('--') + 1:]
    bpy.ops.wm.read_factory_settings(use_empty=True)
    for name in MODELS:
        before = set(bpy.data.objects)
        bpy.ops.import_scene.gltf(filepath=os.path.join(src, name + '.gltf'))
        new = [o for o in bpy.data.objects if o not in before]
        for o in new:
            o.name = name + '_part'  # free the model's name for its root node
        root = bpy.data.objects.new(name, None)
        bpy.context.scene.collection.objects.link(root)
        for o in new:
            if o.parent is None:
                o.parent = root
    bpy.ops.export_scene.gltf(filepath=out, export_format='GLB', export_image_format='AUTO',
                              export_animations=False, export_apply=True)


def main():
    root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    src_dir = os.environ.get('NATURE_DIR', os.path.join(root, '..', 'assets-src', 'quaternius-nature'))
    acg_dir = os.environ.get('AMBIENTCG_DIR', os.path.join(root, '..', 'assets-src', 'ambientcg'))
    out = os.path.join(root, 'web', 'public', 'assets', 'nature')
    shutil.rmtree(out, ignore_errors=True)
    os.makedirs(out)

    with tempfile.TemporaryDirectory() as tmp:
        images = set()
        for name in MODELS:
            g = json.load(open(os.path.join(src_dir, 'glTF', name + '.gltf')))
            for b in g['buffers']:
                shutil.copyfile(os.path.join(src_dir, 'glTF', b['uri']), os.path.join(tmp, b['uri']))
            for img in g.get('images', []):
                base = os.path.basename(img['uri'])
                images.add(base)
                if not base.startswith('Leaf') and not base.startswith('Leaves') and not base.startswith('Grass') and not base.startswith('Flowers'):
                    img['uri'] = base[:-4] + '.jpg'
                    img['mimeType'] = 'image/jpeg'
                else:
                    img['uri'] = base
            json.dump(g, open(os.path.join(tmp, name + '.gltf'), 'w'))
        for base in sorted(images):
            src = os.path.join(src_dir, 'Textures', base)
            if not os.path.exists(src):
                src = os.path.join(src_dir, 'glTF', base)
            size = next((s for k, s in TEX_SIZE.items() if base.startswith(k)), 512)
            alpha = base.startswith(('Leaf', 'Leaves', 'Grass', 'Flowers'))
            dst = os.path.join(tmp, base if alpha else base[:-4] + '.jpg')
            cmd = ['sips', '-Z', str(size)] + ([] if alpha else ['-s', 'format', 'jpeg', '-s', 'formatOptions', '80'])
            subprocess.run(cmd + [src, '--out', dst], check=True, capture_output=True)
        subprocess.run(['blender', '--background', '--python', os.path.abspath(__file__), '--', tmp,
                        os.path.join(out, 'nature.glb')], check=True, capture_output=True)

    for name in GROUND:
        for kind, suffix, size, q in (('color', 'Color', 1024, 75), ('normal', 'NormalGL', 512, 75)):
            src = os.path.join(acg_dir, name, f'{name}_1K-JPG_{suffix}.jpg')
            subprocess.run(['sips', '-Z', str(size), '-s', 'format', 'jpeg', '-s', 'formatOptions', str(q), src,
                            '--out', os.path.join(out, f'{name}_{kind}.jpg')], check=True, capture_output=True)

    total = sum(os.path.getsize(os.path.join(out, f)) for f in os.listdir(out))
    for f in sorted(os.listdir(out)):
        print(f'{os.path.getsize(os.path.join(out, f)) / 1e6:6.2f} MB  {f}')
    print(f'{total / 1e6:6.2f} MB  total')


if __name__ == '__main__':
    try:
        import bpy  # noqa: F401
        in_blender()
    except ImportError:
        main()
