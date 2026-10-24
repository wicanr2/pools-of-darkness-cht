#!/usr/bin/env python3
"""核對 pod-play 正常讀檔、查驗題、地圖與攻略面板的收據。"""
import argparse
import csv
import hashlib
import json
from pathlib import Path

from PIL import Image


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


ap = argparse.ArgumentParser(description=__doc__)
ap.add_argument('--out', required=True, type=Path)
ap.add_argument('--seed', required=True, type=Path)
ap.add_argument('--binary', required=True, type=Path)
ap.add_argument('--text', required=True, type=Path)
ap.add_argument('--font', required=True, type=Path)
a = ap.parse_args()
log = (a.out / 'run.log').read_text()
if '缺字' in log or '截圖失敗' in log or 'frames=4800' not in log:
    raise SystemExit('前端未正常結束、有缺字或截圖失敗')
with (a.out / 'inv.tsv').open() as f:
    rows = list(csv.DictReader(f, delimiter='\t', quoting=csv.QUOTE_NONE))
texts = {r['text'] for r in rows}
if not {'load which game: ', 'ROLAND', 'CHARACTERS AND PARTIES', ' (1)         ', '1     '} <= texts:
    raise SystemExit('讀檔路徑或查驗題不是預期的第 1 頁、第 1 字、CHARACTERS AND PARTIES')
if not any(int(r['step']) > 387_000_000 and 'Move Area Cast View Encamp Search Look' in r['text'] for r in rows):
    raise SystemExit('未在作答 YOU 後回到冒險畫面')
visited = (a.out / 'save/pod-map-visited.txt').read_text().splitlines()
if visited != ['1/10 11 2']:
    raise SystemExit(f'走過格錯誤或混入未初始化資料：{visited}')
shots = {}
for frame, name in ((3900, 'quiz'), (4200, 'adventure'), (4320, 'map'), (4420, 'guide'), (4520, 'map-again'), (4620, 'closed')):
    p = a.out / f'shots/pod-f{frame:06d}.png'
    if Image.open(p).size != (640, 400):
        raise SystemExit(f'{name} 尺寸不符')
    shots[name] = sha(p)
if shots['map'] != shots['map-again'] or len({shots['map'], shots['guide'], shots['closed']}) != 3:
    raise SystemExit('地圖／攻略切換結果不符')
Image.frombytes('RGBA', (640, 400), (a.out / 'final.rgba').read_bytes()).save(a.out / 'final.png')
manifest = dict(source='pod-play cold boot; load game A; no state injection',
                program_sha256=sha(a.binary), font_sha256=sha(a.font),
                catalogs={n: sha(a.text / n) for n in ('map-geo.tsv', 'map-guide.zh-TW.tsv', 'map-titles.tsv', 'walkthrough.zh-TW.tsv', 'host-text.zh-TW.tsv')},
                quiz=dict(page=1, word=1, heading='CHARACTERS AND PARTIES', answer='YOU'),
                rng='dosgolem 固定 DOS 時間全零，固定開機輸入時序；未修改或重擲原版 RNG，不比較骰子規則',
                save_inputs={p.name: sha(p) for p in sorted(a.seed.iterdir()) if p.is_file()},
                shots=shots, visited=visited)
(a.out / 'receipt.json').write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + '\n')
print('冷開機讀槽 A → 原版查驗題 YOU → 冒險 → 地圖 → 攻略 → 地圖 → 關閉：通過；走過格只有 1/10 11 2')
