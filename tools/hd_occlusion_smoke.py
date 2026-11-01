#!/usr/bin/env python3
"""容器內獨立核對7400格終點的新增HD地形，通過後才可更新圖面基準。"""
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image

ROOT = Path('/project')
OUT = Path('/out')
THEME = ROOT / 'workplace/hd/theme'


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    return json.loads(path.read_text())


def assets(name):
    for item in read(THEME / name)['items']:
        assert sha(THEME / item['indexed']) == item['indexed_sha256']
        assert sha(THEME / item['png']) == item['png_sha256']
        w, h = item['size']
        ref = np.frombuffer((THEME / item['indexed']).read_bytes(), np.uint8).reshape(h, w)
        png = np.array(Image.open(THEME / item['png']).convert('RGBA'))
        yield item, ref, png


old_path = ROOT / 'workplace/probe/g9-hd/b43-play-v1/hd/frame.png'
baseline = read(ROOT / 'docs/re/017-b43-play.json')
assert sha(old_path) == baseline['frames_sha256']['after']
old = np.array(Image.open(old_path).convert('RGBA'))
new_path = OUT / 'smoke/frame.png'
new = np.array(Image.open(new_path).convert('RGBA'))
assert old.shape == new.shape == (600, 1440, 4)
source = np.frombuffer((OUT / 'smoke/source.idx').read_bytes(), np.uint8).reshape(200, 320)
palette = OUT / 'smoke/source.pal'
expected = read(ROOT / 'review/hd/runtime/manifest.json')['runtime_expected']
assert sha(OUT / 'smoke/source.idx') == expected['source_indexed_sha256']
assert sha(palette) == expected['source_palette_sha256']
# 只以完整來源前景建立遮罩，不讀Go合成器的匹配結果。
cover = np.zeros((200, 320), bool)
foreground = []
for kind, manifest in [('actor', 'actors.json'), ('effect', 'effects.json')]:
    seen = set()
    for item, ref, png in assets(manifest):
        for flip in (False, True):
            r = ref[:, ::-1] if flip else ref
            signature = r.tobytes()
            if signature in seen:
                continue
            seen.add(signature)
            h, w = r.shape
            ink = r != 8
            step = 24 if kind == 'actor' else 8
            for y in range(8, 177-h, step):
                for x in range(8, 177-w, step):
                    if np.array_equal(source[y:y+h, x:x+w][ink], r[ink]):
                        cover[y:y+h, x:x+w] |= ink
                        foreground.append(dict(kind=kind, key=item['key'], x=x, y=y,
                                               flipped=flip, source_points=int(ink.sum())))
allowed = np.zeros((200, 320), bool)
terrain = [v for v in assets('combat.json') if v[0]['key'].startswith('dungcom-005@')]
layers = []
for y in range(8, 153, 24):
    for x in range(8, 153, 24):
        hidden = cover[y:y+24, x:x+24]
        visible = ~hidden
        raw = source[y:y+24, x:x+24]
        if not hidden.any() or visible.sum() < 144 or len(np.unique(raw[visible])) < 3:
            continue
        candidates = [(i, r, p) for i, r, p in terrain if np.array_equal(raw[visible], r[visible])]
        if len(candidates) != 1:
            continue
        item, ref, png = candidates[0]
        visible3 = visible.repeat(3, 0).repeat(3, 1)
        actual = new[y*3:(y+24)*3, 240+x*3:240+(x+24)*3]
        assert np.array_equal(actual[visible3], png[visible3]), (item['key'], x, y)
        allowed[y:y+24, x:x+24] |= visible
        layers.append(dict(key=item['key'], x=x, y=y, source_points=int(visible.sum()),
                           png_sha256=item['png_sha256']))
allowed_full = np.zeros(new.shape[:2], bool)
allowed_full[:, 240:1200] = allowed.repeat(3, 0).repeat(3, 1)
difference = np.any(old != new, axis=2)
assert difference.any()
assert not np.any(difference & ~allowed_full), '已證實地形外出現差異'
proof = dict(issue=73, kind='independent7400-frame endpoint terrain comparison',
             baseline_frame_sha256=sha(old_path), frame_sha256=sha(new_path),
             source_indexed_sha256=sha(OUT / 'smoke/source.idx'), source_palette_sha256=sha(palette),
             changed_pixels=int(difference.sum()), outside_approved_terrain_changes=0,
             pixel_mismatches=0, foreground=foreground, terrain=layers)
(OUT / 'smoke-pixels.json').write_text(json.dumps(proof, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(dict(changed_pixels=proof['changed_pixels'], terrain=len(layers),
                     outside_changes=0, pixel_mismatches=0)))
