#!/usr/bin/env python3
"""核對地圖 A/B 收據，失配即以非零結束。"""
import argparse
import hashlib
import json
from pathlib import Path


def tree(path):
    return {str(p.relative_to(path)): hashlib.sha256(p.read_bytes()).hexdigest()
            for p in path.rglob('*') if p.is_file()}


ap = argparse.ArgumentParser(description=__doc__)
ap.add_argument('root', type=Path)
ap.add_argument('--map-key', default='1/10')
ap.add_argument('--position', type=int, nargs=3, default=[11, 2, 2])
args = ap.parse_args()
root = args.root
for scale in (2, 3):
    results = {m: json.loads((root / f'{scale}-{m}' / 'result.json').read_text())['shots'][0]
               for m in ('original', 'visited', 'all')}
    orig = results['original']
    for mode, result in results.items():
        for key in ('step', 'indexed_sha256', 'memory_sha256', 'registers'):
            if not result.get(key) or result[key] != orig[key]:
                raise SystemExit(f'{scale}-{mode}：原版 {key} 改變或未記錄')
        if result['diff_outside_stamps'] != 0 or result.get('missing_glyphs'):
            raise SystemExit(f'{scale}-{mode}：範圍外差異或缺字')
        if tree(root / f'{scale}-{mode}' / 'scratch') != tree(root / f'{scale}-original' / 'scratch'):
            raise SystemExit(f'{scale}-{mode}：原版存檔／暫存層不同')
        if mode != 'original' and (result.get('map_key') != args.map_key or result.get('map_position') != args.position):
            raise SystemExit(f'{scale}-{mode}：地圖或座標方向不符')
    if len({r['overlay_sha256'] for r in results.values()}) != 3:
        raise SystemExit(f'{scale} 倍三個模式畫面沒有區分')
    print(f'{scale} 倍：原版／走過／全圖，色號、記憶體、暫存器及存檔相同；面板外差異 0、缺字 0')
