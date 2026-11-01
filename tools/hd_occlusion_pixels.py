#!/usr/bin/env python3
"""容器內獨立像素核對，不使用 Go 比對器匯出的匹配清單。"""
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image

ROOT = Path('/project')
OUT = Path('/out')
THEME = ROOT / 'workplace/hd/theme'
CAPTURE = ROOT / 'workplace/probe/invincible-v1/gui-audit/sources'


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def assets(manifest):
    result = []
    for item in json.loads((THEME / manifest).read_text())['items']:
        assert sha(THEME / item['indexed']) == item['indexed_sha256']
        assert sha(THEME / item['png']) == item['png_sha256']
        w, h = item['size']
        ref = np.frombuffer((THEME / item['indexed']).read_bytes(), np.uint8).reshape(h, w)
        png = np.array(Image.open(THEME / item['png']).convert('RGBA'))
        result.append((item, ref, png))
    return result


effects = assets('effects.json')
actors = assets('actors.json')
terrain = [v for v in assets('combat.json') if v[0]['key'].startswith('dungcom-005@')]
proof = []
for label, key, x, y in [(8192, 'comspr-019', 16, 24), (8193, 'comspr-019', 24, 32),
                         (8196, 'comspr-019', 40, 48), (8899, 'comspr-133', 104, 80)]:
    source_path = CAPTURE / f'pod-f{label:06}.idx'
    source = np.frombuffer(source_path.read_bytes(), np.uint8).reshape(200, 320)
    palette = CAPTURE / f'pod-f{label:06}.pal'
    assert hashlib.sha256(palette.read_bytes()[32*3:64*3]).hexdigest() == '072944b66eed642a372f2e70e094649cad7b53b7574b0cad471dc7e26164dd63'
    got = np.array(Image.open(OUT / f'component-{label:06}.png').convert('RGBA'))
    old = np.array(Image.open(OUT / f'component-old-{label:06}.png').convert('RGBA'))
    allowed = np.zeros((200, 320), bool)
    cover = allowed.copy()
    checked = []

    def check(item, ref, png, px, py, mask, kind, flip=False):
        h, w = ref.shape
        mask3 = mask.repeat(3, 0).repeat(3, 1)
        actual = got[py*3:(py+h)*3, px*3:(px+w)*3]
        assert np.array_equal(actual[mask3], png[mask3]), (label, kind, item['key'])
        allowed[py:py+h, px:px+w] |= mask
        checked.append(dict(kind=kind, key=item['key'], x=px, y=py, flipped=flip,
                            source_points=int(mask.sum()), png_sha256=item['png_sha256']))

    item, ref, png = next(v for v in effects if v[0]['key'] == key)
    matches = []
    for flip in [False, True]:
        r = ref[:, ::-1] if flip else ref
        p = png[:, ::-1] if flip else png
        mask = r != 8
        if np.array_equal(source[y:y+24, x:x+24][mask], r[mask]):
            if not any(np.array_equal(r, previous[1]) for previous in matches):
                matches.append((flip, r, p, mask))
    assert len(matches) == 1, (label, '特效來源歧義')
    flip, ref, png, fx = matches[0]
    check(item, ref, png, x, y, fx, 'effect', flip)
    cover[y:y+24, x:x+24] |= fx
    effect_mask = cover.copy()
    # 獨立列舉所有完整怪物，以及僅被已證實特效遮住的怪物。
    masked_actors = 0
    for item, ref, png in actors:
        h, w = ref.shape
        for flip in [False, True]:
            r = ref[:, ::-1] if flip else ref
            p = png[:, ::-1] if flip else png
            ink = r != 8
            for py in range(8, 177-h, 24):
                for px in range(8, 177-w, 24):
                    hidden = effect_mask[py:py+h, px:px+w]
                    visible = ink & ~hidden
                    raw = source[py:py+h, px:px+w]
                    if visible.sum() < 144 or len(np.unique(r[visible])) < 3:
                        continue
                    if not np.array_equal(raw[visible], r[visible]):
                        continue
                    if np.any(ink & hidden):
                        assert np.any(raw[ink & hidden] != r[ink & hidden])
                        check(item, r, p, px, py, visible, 'occluded_actor', flip)
                        masked_actors += 1
                    cover[py:py+h, px:px+w] |= ink
    assert masked_actors == (1 if label == 8899 else 0)
    for py in range(8, 153, 24):
        for px in range(8, 153, 24):
            hidden = cover[py:py+24, px:px+24]
            visible = ~hidden
            raw = source[py:py+24, px:px+24]
            if not hidden.any() or visible.sum() < 144 or len(np.unique(raw[visible])) < 3:
                continue
            candidates = [(i, r, p) for i, r, p in terrain if np.array_equal(raw[visible], r[visible])]
            if len(candidates) == 1:
                item, ref, png = candidates[0]
                check(item, ref, png, px, py, visible, 'terrain')
    difference = np.any(got != old, axis=2)
    assert not np.any(difference & ~allowed.repeat(3, 0).repeat(3, 1)), (label, '核准範圍外變更')
    proof.append(dict(capture_label=label, advanced_frame=label-1, source_sha256=sha(source_path),
                      component_sha256=sha(OUT / f'component-{label:06}.png'),
                      baseline_component_sha256=sha(OUT / f'component-old-{label:06}.png'),
                      changed_pixels=int(difference.sum()), outside_changes=0, pixel_mismatches=0,
                      layers=checked))

prefix_path = OUT / 'gui-diagnose/sources/pod-f008751.idx'
raw = np.frombuffer(prefix_path.read_bytes(), np.uint8).reshape(200, 320)[64:72, 64:88]
prefix_candidates = []
seen = set()
for item, ref, png in effects:
    for flip in [False, True]:
        r = ref[:, ::-1] if flip else ref
        if r.tobytes() in seen:
            continue
        seen.add(r.tobytes())
        if np.array_equal(raw, np.where(r[:8] == 8, 32, r[:8])):
            prefix_candidates.append(dict(key=item['key'], flipped=flip,
                                          foreground_points=int((r[:8] != 8).sum())))
assert prefix_candidates == [dict(key='comspr-005', flipped=False, foreground_points=37)]
assert len(np.unique(raw)) == 7
receipt = dict(issue=73, method='Pillow/numpy independent source and approved PNG pixel comparison',
               capture_label_is_advanced_frame_plus_one=True, samples=proof,
               small_prefix=dict(capture_label=8751, advanced_frame=8750,
                                 source_sha256=sha(prefix_path), x=64, y=64, width=24, height=8,
                                 complete_source_points=192, colors=7, candidates=prefix_candidates,
                                 purpose='eligibility for existing four-frame hold only; no partial HD'))
(OUT / 'pixels.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + '\n')
log = (OUT / 'component-oracle.log').read_text()
for name in ['TestHDCompositeRealSources', 'TestHDCompositeForegroundIndependent',
             'TestHDCompositeSmallPrefixGuards', 'TestHDCompositeHoldRemainsFourFrames']:
    assert '--- PASS: ' + name in log
source_receipt = dict(date='2026-11-01', issue=73,
                      kind='normal captured source component verification; not a substitute for continuous GUI',
                      base_engine='33094bda183f9656c4be0f821a831b8523319dda',
                      base_binary_sha256='139279d04a43467564599c3e7095dda705781498a7c6f3ebff7920ff7dd8075f',
                      component=json.loads((OUT / 'component.json').read_text()), independent_pixels=receipt,
                      prefix_guards_pass=True, hold_limit_frames=4, fifth_frame_returns_original=True,
                      environment=dict(go='1.24.13', go_image='dosgolem-go-ebiten:1.24-r1',
                                       image_tools='sosb-img:r1'),
                      component_log_sha256=sha(OUT / 'component-oracle.log'))
(OUT / 'source.json').write_text(json.dumps(source_receipt, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(dict(samples=len(proof), outside_changes=0, pixel_mismatches=0)))
