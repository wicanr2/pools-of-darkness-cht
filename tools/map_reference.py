#!/usr/bin/env python3
"""在 Docker 產生 GEO 研究圖與索引；原版圖面只寫 ignored 工作區。"""
import argparse
import hashlib
import json
from pathlib import Path
import sys

from PIL import Image, ImageDraw

sys.path.insert(0, str(Path(__file__).resolve().parent))
from ecl_text_catalog import dax_blocks


def draw_map(key, body, events=False):
    cell, pad = 20, 24
    im = Image.new('RGB', (344, 364), 'white')
    d = ImageDraw.Draw(im)
    d.text((4, 2), key + (' event plane' if events else ' walls'), fill='black')
    for y in range(16):
        d.text((2, 30 + y * cell), f'{y:X}', fill='black')
        for x in range(16):
            if y == 0:
                d.text((pad + 6 + x * cell, 14), f'{x:X}', fill='black')
            a, c, i = pad + x * cell, pad + y * cell, y * 16 + x
            walls = body[i] >> 4, body[i] & 15, body[256+i] >> 4, body[256+i] & 15
            for k, line in enumerate(((a,c,a+cell,c), (a+cell,c,a+cell,c+cell),
                                      (a,c+cell,a+cell,c+cell), (a,c,a,c+cell))):
                if walls[k]:
                    d.line(line, fill='black', width=1)
            if events and body[512+i]:
                d.text((a+2,c+4), f'{body[512+i]:02X}', fill='blue')
    return im


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--orig', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    a = ap.parse_args()
    a.out.mkdir(parents=True, exist_ok=True)
    tiles, rows = [], []
    for path in sorted(a.orig.glob('GEO[0-9]*.DAX')):
        for bid, block in dax_blocks(path.read_bytes()):
            if len(block) != 0x402:
                raise ValueError(f'{path.name}:{bid:x} 長度 {len(block)}')
            key, body = f'{int(path.stem[3:])}/{bid:02x}', block[2:]
            tile = draw_map(key, body)
            name = f'{int(path.stem[3:])}-{bid:02x}'
            tile.save(a.out / f'{name}.png')
            draw_map(key, body, True).save(a.out / f'{name}-events.png')
            tiles.append(tile)
            rows.append(dict(key=key, archive=path.name, block=bid,
                             payload_sha256=hashlib.sha256(body).hexdigest()))
    for first in range(0, len(tiles), 8):
        sheet = Image.new('RGB', (344*4,364*2), '#bbb')
        for i, tile in enumerate(tiles[first:first+8]):
            sheet.paste(tile, (i%4*344,i//4*364))
        sheet.save(a.out / f'atlas-{first//8+1}.png')
    (a.out / 'geo-index.json').write_text(json.dumps(rows, indent=2)+'\n')
    print(f'{len(rows)} 塊 GEO，{(len(rows)+7)//8} 張 atlas')


if __name__ == '__main__':
    main()
