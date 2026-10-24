#!/usr/bin/env python3
"""以已證實街景與已驗收素材建立既有格式walls.json，不改其他圖層。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import sys

from PIL import Image

sys.path.insert(0, str(Path(__file__).resolve().parent))
from dax_vga import block_vga
from ecl_text_catalog import dax_blocks


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for name in ('orig', 'capture', 'proof', 'req', 'out', 'sky-req', 'sky-out', 'theme'):
        ap.add_argument('--' + name, type=Path, required=True)
    a = ap.parse_args()
    proof = json.loads(a.proof.read_text())
    capture = (a.capture / 'source.idx').read_bytes()
    palette = (a.capture / 'source.pal').read_bytes()
    assert sha(a.capture / 'source.idx') == proof['inputs']['indexed_sha256']
    assert sha(a.capture / 'source.pal') == proof['inputs']['palette_sha256']
    for name in ('8X8D1.DAX', 'SKYGRND.DAX', 'WALLDEF1.DAX'):
        assert sha(a.orig / name) == proof['inputs'][name]
    wallblock = dict(dax_blocks((a.orig / '8X8D1.DAX').read_bytes()))[1]
    skyblock = dict(dax_blocks((a.orig / 'SKYGRND.DAX').read_bytes()))[1]
    _, _, skyframes, _, _, _ = block_vga(skyblock, palette)
    expected_keys = {f'8x8d1-001@{n}' for n in proof['unique_wall_tiles']}
    assets, refs, files, provenance = [], {}, [], []
    for reqdir, outdir, kind in ((a.req, a.out, 'wall'), (a.sky_req, a.sky_out, 'sky')):
        request = json.loads((reqdir / 'manifest.json').read_text())
        valid = json.loads((outdir / 'validation.json').read_text())
        visual = json.loads((outdir / 'visual-review.json').read_text())
        assert valid['all_pass'] and valid['scope'] == 'all'
        assert valid['request_manifest_sha256'] == sha(reqdir / 'manifest.json')
        numbers = {r['key']: r for r in valid['items']}
        selected = request['items'] if kind == 'wall' else [r for r in request['items'] if r['key'] == 'skygrnd-001']
        assert {r['key'] for r in selected} == (expected_keys if kind == 'wall' else {'skygrnd-001'})
        for it in selected:
            key = it['key']
            png, idx, pal = outdir / (key + '.png'), reqdir / it['indexed'], reqdir / it['palette']
            assert sha(png) == numbers[key]['png_sha256'] == visual[key]
            assert sha(idx) == it['indexed_sha256'] and sha(pal) == it['palette_sha256']
            ref = idx.read_bytes()
            sourcefile = '8X8D1.DAX' if kind == 'wall' else 'SKYGRND.DAX'
            assert it['source_file'] == sourcefile and it['source_file_sha256'] == sha(a.orig / sourcefile)
            if kind == 'wall':
                n = int(key.split('@')[1])
                assert ref == wallblock[434+n*64:434+(n+1)*64] and it['source_size'] == [8, 8]
                assert it.get('transparent_index') == 13
            else:
                assert ref == skyframes[0] and it['source_size'] == [88, 88]
            w, h = it['source_size']
            with Image.open(png) as im:
                assert im.mode == 'RGBA' and im.size == (w*3, h*3)
                assert im.getextrema()[3] == (255, 255)
                if kind == 'wall':
                    for y in range(24):
                        for x in range(24):
                            c = ref[y//3*8+x//3]
                            want = (85, 85, 85, 255) if c == 13 else tuple(palette[c*3:c*3+3]) + (255,)
                            if c == 13 or x in (0, 23) or y in (0, 23):
                                assert im.getpixel((x, y)) == want, f'{key}: 留底或接縫改動'
            refs[key] = ref
            assets.append(dict(key=key, kind=kind, size=[w, h], indexed=key+'.idx', indexed_sha256=sha(idx), png=key+'.png', png_sha256=sha(png)))
            files.extend(((idx, key+'.idx'), (png, key+'.png')))
        provenance.append(dict(kind=kind, request_sha256=sha(reqdir/'manifest.json'), validation_sha256=sha(outdir/'validation.json'), visual_sha256=sha(outdir/'visual-review.json')))
    cells = []
    assert len(proof['cells']) == 121 and {r['position'] for r in proof['cells']} == set(range(121))
    for row in proof['cells']:
        pos = row['position']
        cx, cy = pos % 11, pos // 11
        x, y = 24+cx*8, 24+cy*8
        assert row['rect'] == [x, y, 8, 8] and row['sky'] == 'skygrnd-001'
        ref = bytearray().join(refs['skygrnd-001'][(cy*8+dy)*88+cx*8:(cy*8+dy)*88+cx*8+8] for dy in range(8))
        layers = [dict(key='skygrnd-001', at=[-cx*8, -cy*8])]
        if 'wall' in row:
            tile = refs[row['wall']]
            for i, c in enumerate(tile):
                if c != 13:
                    ref[i] = c
            layers.append(dict(key=row['wall'], at=[0, 0]))
        actual = b''.join(capture[(y+dy)*320+x:(y+dy)*320+x+8] for dy in range(8))
        assert ref == actual and hashlib.sha256(ref).hexdigest() == row['indexed_sha256']
        cells.append(dict(key=f'vga-phlan-{pos:03d}', positions=[pos], base_color=8, layers=layers, indexed_sha256=row['indexed_sha256'], paint=True))
    manifest = dict(version=1, scale=3, kind='wall-grid', viewport=[24, 24, 11, 11, 8], assets=assets, cells=cells)
    a.theme.mkdir(parents=True, exist_ok=True)
    for src, name in files:
        dst = a.theme / name
        assert not dst.exists() or sha(dst) == sha(src), f'{name}已有不同素材，請使用乾淨測試主題'
    for src, name in files:
        shutil.copyfile(src, a.theme / name)
    temp = a.theme / 'walls.json.tmp'
    temp.write_text(json.dumps(manifest, ensure_ascii=False, indent=2)+'\n')
    temp.replace(a.theme / 'walls.json')
    (a.theme/'vga-walls-provenance.json').write_text(json.dumps(dict(proof_sha256=sha(a.proof), scope='one observed viewport; runtime gate separate', items=provenance), indent=2)+'\n')
    print('既有格式：18個來源，121個完整組合格；其他圖層保留。')


if __name__ == '__main__':
    main()
