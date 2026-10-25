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
    args = parser.parse_args()
    selections = [('b18', '044'), ('b18', '048'), ('b21', '069'), ('b21', '077'),
                  ('b23', '094'), ('b23', '095'), ('b23', '105'), ('b24', '108')]
    canvas = Image.new('RGB', (600, 300), (85, 85, 85))
    rows = []
    for i, (batch, number) in enumerate(selections):
        key = 'cpic1-' + number
        directory = args.root / 'review/hd' / (batch + '-cpic')
        proof = json.loads((directory / 'manifest.json').read_text())
        item = next(row for row in proof['items'] if row['key'] == key)
        path = directory / (key + '.png')
        digest = sha(path)
        assert item['numeric']['target_met'] and item['visual']['png_sha256'] == digest
        with Image.open(path) as image:
            assert image.size == tuple(item['source']['output_size']) and image.mode == 'RGBA'
            x = (i % 4) * 150 + (150 - image.width) // 2
            y = (i // 4) * 150 + (150 - image.height) // 2
            canvas.paste(image.convert('RGB'), (x, y))
            rows.append(dict(key=key, source=str(path.relative_to(args.root)), png_sha256=digest,
                             size=list(image.size), position=[x, y]))
    target = args.root / 'docs/screenshots/monster-hd-preview.png'
    assert target.parent.is_dir() and target.parent.stat().st_uid == 1000
    canvas.save(target)
    report = dict(scope='AI-redrawn asset preview; not a gameplay screenshot',
                  composition='native 3x PNG; no resizing or cropping', size=[600, 300],
                  output_sha256=sha(target), items=rows,
                  rights='AI redraws within existing private HD patch scope; no original reference pixels')
    target.with_suffix('.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
    print(target, report['output_sha256'])


if __name__ == '__main__':
    main()
