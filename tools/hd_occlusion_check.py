#!/usr/bin/env python3
"""容器內核對030正常重播、029原版狀態與素材不變。"""
import hashlib
import json
import re
from pathlib import Path

import numpy as np
from PIL import Image

ROOT = Path('/project')
OUT = Path('/out')


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def read(p):
    return json.loads(p.read_text())


def tree(p):
    return {str(x.relative_to(p)): sha(x) for x in sorted(p.rglob('*')) if x.is_file()}


def state(log):
    a = re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)', log)
    b = re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)', log)
    assert a and b
    return dict(frames=int(a[1]), steps=int(a[2]), memory=a[3], indexed=a[4], cpu=b[1], palette=b[2])


baseline_path = ROOT / 'docs/re/029-invincible-runtime.json'
assert sha(baseline_path) == 'f5bc958eab9de574e8ae46ab0a5cc82bb8084605c3187fcca2befdf12042b4f3'
baseline = read(baseline_path)
old = ROOT / 'workplace/probe/invincible-v1/gui-audit'
gui = OUT / 'gui-final2'
log = (gui / 'run.log').read_text()
def gameplay_script(p):
    return [x for x in p.read_text().strip().split(',') if x.split(':', 1)[1] != 'shot']
assert gameplay_script(gui / 'script.txt') == gameplay_script(old / 'script.txt')
assert state(log) == baseline['normal']['states']['audit']
front = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt',
         'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
guest = {n: h for n, h in tree(gui / 'game').items() if n not in front}
assert guest == {n: h for n, h in tree(old / 'game').items() if n not in front}
assert len(guest) == 127
assert '中止 1 次、中止後返回 1 次' in log
assert '無敵 開啟true、還原 5 次、移除有害效果 0 筆、拒絕 0 次' in log
assert '字型缺字' not in log and '截圖失敗' not in log
raw = int(re.search(r'閃現 (\d+) 次', log)[1])
assert raw == 0, ('正常路線閃現', raw)
draw = read(gui / 'draw-report.json')
assert draw['NativeBytes'] == draw['HostBytes'] > 0
assert draw['FallbackBytes'] == draw['Mismatches'] == draw['LegacyRGBFrames'] == 0
sources = {}
for p in (gui / 'sources').iterdir():
    q = old / 'sources' / p.name
    if q.exists() and p.suffix in ('.idx', '.pal'):
        assert sha(p) == sha(q), p.name
        sources[p.name] = sha(p)
assert len(sources) >= 32
cells = []
for event in baseline['normal']['long_combat_cells']:
    first, last, row, col = (event[k] for k in ('first', 'last', 'row', 'col'))
    # actShot發生在Advance之前：量尺N必須用擷取N+1。
    before = np.array(Image.open(old / 'shots' / f'pod-f{first+1:06}.png').convert('RGBA'))
    after = np.array(Image.open(gui / 'shots' / f'pod-f{first+1:06}.png').convert('RGBA'))
    prefix = old / 'sources' / f'pod-f{first+1:06}'
    idx = np.frombuffer(prefix.with_suffix('.idx').read_bytes(), np.uint8).reshape(200, 320)
    pal = np.frombuffer(prefix.with_suffix('.pal').read_bytes(), np.uint8).reshape(256, 3)
    original = np.concatenate([pal[idx], np.full((200, 320, 1), 255, np.uint8)], axis=2).repeat(3, 0).repeat(3, 1)
    def cell(img):
        return img[row*24:(row+1)*24, col*24:(col+1)*24]
    assert np.array_equal(cell(before), cell(original)), ('舊格非原版', first, row, col)
    assert np.any(cell(after) != cell(original))
    cells.append(dict(advanced_frame=first, capture_label=first+1, row=row, col=col,
                      old_sha256=hashlib.sha256(cell(before).tobytes()).hexdigest(),
                      new_sha256=hashlib.sha256(cell(after).tobytes()).hexdigest()))
assert len(cells) == 19
lock = read(ROOT / 'review/hd/runtime/manifest.json')
theme = tree(ROOT / 'workplace/hd/theme')
assert theme == {x['path']: x['sha256'] for x in lock['theme']}
assert len(theme) == 1536 and len(lock['fonts']) == 10
for item in lock['fonts']:
    assert sha(ROOT / 'workplace/font' / item['path']) == item['sha256']
assert sha(OUT / 'pod-play') == sha(OUT / 'clean-pod-play')
engine = (OUT / 'engine-commit.txt').read_text().strip()
engine_tree = (OUT / 'engine-tree.txt').read_text().strip()
assert re.fullmatch(r'[0-9a-f]{40}', engine), '缺少完整引擎commit'
assert re.fullmatch(r'[0-9a-f]{40}', engine_tree), '缺少完整引擎tree'
for name in ('component-oracle.log', 'regression.log', 'frontend.log', 'cache.log'):
    assert '\nok ' in '\n' + (OUT / name).read_text(), name
assert not (OUT / 'vet.log').read_text().strip()
smoke_pixels = read(OUT / 'smoke-pixels.json')
smoke_verify = read(OUT / 'smoke/receipt.json')
assert smoke_pixels['baseline_frame_sha256'] == read(ROOT / 'docs/re/017-b43-play.json')['frames_sha256']['after']
assert smoke_pixels['frame_sha256'] == sha(OUT / 'smoke/frame.png') == lock['runtime_expected']['frame_sha256']
assert smoke_pixels['outside_approved_terrain_changes'] == smoke_pixels['pixel_mismatches'] == 0
assert smoke_verify['status'] == 'PASS' and smoke_verify['engine_commit'] == engine
assert smoke_verify['guest_files'] == 127 and smoke_verify['flash_count'] == 0
assert smoke_verify['binary_sha256'] == sha(OUT / 'pod-play')
receipt = dict(date='2026-11-01', issue=73, status='PASS within sampled normal DEMO route',
               base_engine=baseline['engine'], engine=engine,
               engine_tree=engine_tree, binary_sha256=sha(OUT / 'pod-play'),
               baseline_receipt_sha256=sha(baseline_path), capture_label_is_advanced_frame_plus_one=True,
               source_proof=dict(path='docs/re/030-hd-occlusion-source.json',
                                 sha256=sha(ROOT / 'docs/re/030-hd-occlusion-source.json')),
               normal=dict(state=state(log), guest_files=127, unchanged_guest=True,
                           non_screenshot_inputs_unchanged=True,
                           flash_before=19, flash_after=raw, corrected_cells=cells, source_hashes=sources,
                           demo_aborts=1, demo_returns=1, invincible_restores=5,
                           script_sha256=sha(gui / 'script.txt'), draw=draw,
                           screenshots={p.name: sha(p) for p in sorted((gui / 'window').glob('*.png'))}),
               assets=dict(theme_files=1536, fonts=10, unchanged=True), clean_git_build_matches=True,
               rebuild_smoke=dict(frames=7400, steps=712250000,
                                  scope='normal command rerun with clean-built engine and hash-locked current assets',
                                  pixels=smoke_pixels, verification=smoke_verify),
               logs={n: sha(OUT / n) for n in ('component-oracle.log', 'regression.log', 'frontend.log', 'cache.log', 'vet.log')},
               tools={p.name: sha(p) for p in sorted((ROOT / 'tools').glob('hd_occlusion*'))},
               limits=['sampled existing normal route only', 'ambiguous or unknown sources retain original',
                       'other original project work and old nine packages remain separate'])
(OUT / 'runtime.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(dict(flash_before=19, flash_after=raw, corrected_cells=len(cells),
                     guest_files=127, source_pairs=len(sources)//2, native_bytes=draw['NativeBytes'])))
