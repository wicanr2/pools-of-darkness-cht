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
    parser.add_argument('--kind', choices=('monster', 'terrain'), default='monster')
    args = parser.parse_args()
    if args.kind == 'terrain':
        proof = json.loads((args.root / 'review/hd/b26-terrain/manifest.json').read_text())
        selections = [('b26-terrain', row['key']) for row in proof['items']]
        assert len(selections) == 23
        size, columns, cell, name = (480, 320), 6, 80, 'dungeon-terrain-hd-preview.png'
    else:
        selections = [('b18-cpic', 'cpic1-044'), ('b18-cpic', 'cpic1-048'),
                      ('b21-cpic', 'cpic1-069'), ('b21-cpic', 'cpic1-077'),
                      ('b23-cpic', 'cpic1-094'), ('b23-cpic', 'cpic1-095'),
                      ('b23-cpic', 'cpic1-105'), ('b24-cpic', 'cpic1-108')]
        size, columns, cell, name = (600, 300), 4, 150, 'monster-hd-preview.png'
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
            x = (i % columns) * cell + (cell - image.width) // 2
            y = (i // columns) * cell + (cell - image.height) // 2
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
