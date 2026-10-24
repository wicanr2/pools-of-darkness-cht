#!/usr/bin/env python3
"""在 Docker 建立本作 VGA 首批請求包；來源資料只寫 ignored workplace/hd。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import struct

from PIL import Image, ImageChops

SCENES = {
    'bigpic1-240': 'This is a top-down fantasy geographical map, not a perspective landscape. Trace the exact coastline, islands, mountain chains, forest regions, roads and settlement symbols. Keep the original yellow-green land and blue sea. Do not invent or move settlements, roads, labels or decorative map furniture.',
    'bigpic1-241': 'Preserve the exact overhead volcanic fantasy map, purple and black terrain, orange-red lava channels and the small gray structures or symbols. Keep every major boundary, ridge and small symbol in its original location. Do not add a perspective sky, horizon, buildings, creatures or objects.',
    'bigpic1-242': 'Preserve this exact green fantasy landscape and colossal organic form, including every original silhouette, texture region, color mass and background boundary. Do not reinterpret the forms, invent anatomy or add creatures or buildings.',
    'bigpic1-244': 'Preserve the exact strange forest scene, twisting orange-red trunks, turquoise and green background, roots, vines and distant silhouettes. Keep their original positions, widths and shapes. No new subjects or objects.',
    'bigpic1-245': 'This is the original top-down fantasy map with a different set of settlement symbols. Preserve every coastline, island, mountain and town location exactly, including the exact roof colors and silhouettes of each source settlement. Do not reinterpret red roofs as fire. No new sites, roads, labels or decorations.',
    'bigpic1-243': 'Preserve the exact coastal city panorama: turbulent turquoise sea across the middle, storm-dark gray sky, pale distant mountains, green cliffs, cream buildings and red tiled roofs on both sides. Keep every coastline, tower, roof, tree group and horizon at its original relative position. Do not add people, ships, monsters, lettering or extra buildings.',
    'bigpic1-246': 'Preserve the exact fantasy battle composition: the large winged fiery orange and yellow creature dominates the left, reaching toward the right; the blue-white magical burst and robed spellcaster remain on the right. Preserve the wings, face, arms, hands, projectiles, purple ground and small green lights at their original relative positions. Do not add characters, weapons, lettering, frames or effects.',
    'pic1-131': 'Preserve this exact elderly wizard portrait: long flowing white beard and white hair, broad olive green pointed hat, wise elderly face, purple robe and warm golden patterned shoulder collar, warm brown-gold background. Keep the face, eyes, hat brim, beard silhouette and shoulder positions fixed. Keep his age and expression. No extra objects, letters or frame.',
    'pic1-043@0': 'Preserve the exact fiery orange-yellow creature, the black silhouette around it, facial expression, raised arms, flame shapes and the original orange-red glow. This is a painted fantasy creature portrait, not a new scene. Keep the original pose and every flame mass in place. Do not add objects, letters or frame.',
}


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--orig',type=Path,required=True)
    ap.add_argument('--inventory',type=Path,required=True)
    ap.add_argument('--catalog',default='inventory.json',help='盤點清單檔名，其他圖族使用 inventory-other.json')
    ap.add_argument('--out',type=Path,required=True)
    ap.add_argument('--issue',required=True)
    ap.add_argument('--batch',default='b01')
    ap.add_argument('--keys',nargs='+',default=['bigpic1-243','bigpic1-246'])
    ap.add_argument('--variants',nargs='*',default=[],help='明示狀態變體 key:base，使用來源 RGB 差異遮罩，不推論時間')
    a=ap.parse_args()
    a.out.mkdir(parents=True,exist_ok=True)
    if Path(a.catalog).name!=a.catalog:raise ValueError('catalog 須為清單檔名')
    source=json.loads((a.inventory/a.catalog).read_text())
    manifest=dict(batch=a.batch,scale=3,style='B',spec='010',issue=a.issue,items=[],
                  rights='local-original-reference; private-review-only; no-original-data-in-release')
    for it in source['items']:
        key=it['key']
        if key not in a.keys:continue
        original=a.orig/it['file']
        if sha(original)!=it['archive_sha256']:raise ValueError(f'{key}: archive changed')
        raw=original.read_bytes(); header=struct.unpack_from('<H',raw)[0]+2
        file_offset=None
        for p in range(2,header,9):
            bid,off,_,_=struct.unpack_from('<BIHH',raw,p)
            if bid==it['block']:file_offset=header+off
        w,h=it['width'],it['height']
        im=Image.open(a.inventory/f'{key}.png').convert('RGB')
        if im.size!=(w,h):raise ValueError(f'{key}: reference size changed')
        one=a.out/f'{key}-1x.png'; three=a.out/f'{key}-3x.png'
        im.save(one);im.resize((w*3,h*3),Image.Resampling.NEAREST).save(three)
        for ext,field in [('idx','indexed_sha256'),('pal','palette_sha256')]:
            path=a.inventory/f'{key}.{ext}'
            if sha(path)!=it[field]:raise ValueError(f'{key}: {ext} changed')
            shutil.copyfile(path,a.out/path.name)
        prompt=('Edit the supplied reference image. Use case: style-transfer. '
                'Create a faithful hand-painted fantasy illustration for a classic CRPG HD art asset. '
                f'Target canvas exactly {w*3} by {h*3} pixels, aspect ratio {w}:{h}, edge to edge with no borders or margins. '
                'Preserve the original composition, silhouettes, camera, proportions and original color relationships. '
                'Replace pixel stair-steps and dithering with deliberate brushwork, painted texture and clean readable forms. '
                'Style B: richly painted 1990s fantasy illustration, not photorealistic, no modern 3D rendering. '
                'The reference is the composition and palette authority. '+SCENES.get(key,
                'Preserve every character, object, silhouette, pose, clothing color, background and light source exactly in place. Do not add or remove subjects, objects, letters or frames.'))
        (a.out/f'{key}-prompt.txt').write_text(prompt+'\n')
        manifest['items'].append(dict(key=key,source_file=it['file'],source_file_sha256=it['archive_sha256'],
            block=it['block'],packed_file_offset=file_offset,decoded_sha256=it['decoded_sha256'],
            source_size=[w,h],output_size=[w*3,h*3],reference_1x=one.name,reference_3x=three.name,
            reference_1x_sha256=sha(one),reference_3x_sha256=sha(three),
            indexed=f'{key}.idx',indexed_sha256=it['indexed_sha256'],
            palette=f'{key}.pal',palette_sha256=it['palette_sha256'],transparent=False,
            animation_mask=None,runtime_observed=it['runtime_observed'],prompt=prompt))
        if 'transparent_index' in it:
            manifest['items'][-1]['transparent_index']=it['transparent_index']
    if {i['key'] for i in manifest['items']}!=set(a.keys):raise ValueError('缺指定來源')
    by_key={i['key']:i for i in manifest['items']}
    variants=dict(v.split(':',1)for v in a.variants)
    if not set(variants)<=set(by_key):raise ValueError('變體 key 未列入請求')
    for it in manifest['items']:
        key=it['key']
        if key in variants or key.startswith('pic1-') and '@' in key and not key.endswith('@0'):
            base=variants.get(key,key.split('@')[0]+'@0')
            if base not in by_key:raise ValueError(f'{key}: 缺基底 {base}')
            im=Image.open(a.out/it['reference_1x']).convert('RGB')
            first=Image.open(a.out/by_key[base]['reference_1x']).convert('RGB')
            diff=ImageChops.difference(im,first).split()
            mask=ImageChops.lighter(ImageChops.lighter(diff[0],diff[1]),diff[2]).point(lambda v:255 if v else 0)
            path=a.out/f'{key}-mask.png'
            mask.resize(tuple(it['output_size']),Image.Resampling.NEAREST).save(path)
            it['animation']={'base':base,'mask':path.name,'mask_sha256':sha(path),'meaning':'source state difference; no timing claim'}
    (a.out/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
    print(f'{a.batch}：{len(manifest["items"])} 件，原圖、3倍、色號、色盤、來源偏移、提示詞與雜湊齊全')


if __name__=='__main__':main()
