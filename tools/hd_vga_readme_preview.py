#!/usr/bin/env python3
"""以正式三倍HD成品原尺寸拼成README素材預覽，不裁切、不縮放。"""
import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--root', type=Path, default=Path('.'))
    parser.add_argument('--kind', choices=('monster', 'terrain', 'effects', 'background', 'encounter', 'intro'), default='monster')
    parser.add_argument('--terrain-batch', default='b26-terrain', help='review/hd內的地形批次目錄')
    parser.add_argument('--name', help='輸出PNG檔名；須為單一檔名')
    args = parser.parse_args()
    if args.kind in ('terrain', 'effects', 'background', 'encounter', 'intro'):
        assert Path(args.terrain_batch).name == args.terrain_batch
        proof = json.loads((args.root / 'review/hd' / args.terrain_batch / 'manifest.json').read_text())
        selections = [(args.terrain_batch, row['key']) for row in proof['items']]
        assert selections
        columns, cell, name = 6, 80, 'dungeon-terrain-hd-preview.png'
        if args.kind in ('background', 'encounter'):
            columns, cell, name = 3, 272, 'background-hd-preview.png'
        row_height = cell
        if args.kind == 'intro':
            columns, cell, row_height, name = 2, 960, 504, 'intro-keyframes-hd-preview.png'
        size = (columns * cell, ((len(selections) + columns - 1) // columns) * row_height)
    else:
        selections = [('b18-cpic', 'cpic1-044'), ('b18-cpic', 'cpic1-048'),
                      ('b21-cpic', 'cpic1-069'), ('b21-cpic', 'cpic1-077'),
                      ('b23-cpic', 'cpic1-094'), ('b23-cpic', 'cpic1-095'),
                      ('b23-cpic', 'cpic1-105'), ('b24-cpic', 'cpic1-108')]
        size, columns, cell, name = (600, 300), 4, 150, 'monster-hd-preview.png'
        row_height = cell
    if args.name:
        assert Path(args.name).name == args.name and args.name.endswith('.png')
        name = args.name
    canvas = Image.new('RGB', size, (85, 85, 85))
    rows = []
    for i, (batch, key) in enumerate(selections):
        directory = args.root / 'review/hd' / batch
        proof = json.loads((directory / 'manifest.json').read_text())
        item = next(row for row in proof['items'] if row['key'] == key)
        path = directory / (key + '.png')
        digest = sha(path)
        assert item['numeric']['target_met'] and item['visual']['png_sha256'] == digest
        with Image.open(path) as image:
            assert image.size == tuple(item['source']['output_size']) and image.mode == 'RGBA'
            assert image.width <= cell and image.height <= row_height
            x = (i % columns) * cell + (cell - image.width) // 2
            y = (i // columns) * row_height + (row_height - image.height) // 2
            canvas.paste(image.convert('RGB'), (x, y))
            rows.append(dict(key=key, source=str(path.relative_to(args.root)), png_sha256=digest,
                             size=list(image.size), position=[x, y]))
    target = args.root / 'docs/screenshots' / name
    assert target.parent.is_dir() and target.parent.stat().st_uid == 1000
    canvas.save(target)
    report = dict(scope='AI-redrawn asset preview; not a gameplay screenshot',
                  composition='native 3x PNG; no resizing or cropping', size=list(size),
                  output_sha256=sha(target), items=rows,
                  rights='AI redraws within existing private HD patch scope; no original reference pixels')
    target.with_suffix('.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
    print(target, report['output_sha256'])


if __name__ == '__main__':
    main()
