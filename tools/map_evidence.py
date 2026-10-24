#!/usr/bin/env python3
"""檢查 G8 正常按鍵收據的地圖指標、座標與原版 GEO，輸出可重生 JSON。"""
import argparse
import hashlib
import json
from pathlib import Path
import struct
import sys

from PIL import Image

sys.path.insert(0, str(Path(__file__).resolve().parent))
from ecl_text_catalog import dax_blocks


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--orig', type=Path, required=True)
    ap.add_argument('--states', type=Path, required=True)
    a = ap.parse_args()
    blocks = dict(dax_blocks((a.orig / 'GEO1.DAX').read_bytes()))
    assert len(blocks) == 32 and all(len(b) == 0x402 for b in blocks.values())
    body = blocks[0x10][2:]
    expected = {
        'start': (11, 2, 4), 'left': (11, 2, 2), 'east1': (12, 2, 2),
        'east2': (13, 2, 2), 'south': (13, 2, 4), 'west': (13, 2, 6),
        'north': (13, 2, 0), 'gate': (0, 4, 6),
    }
    results = []
    for name, pos in expected.items():
        path = a.out / f'{name}.mem'
        mem = path.read_bytes()
        assert len(mem) == 0x100000, (name, len(mem))
        off, seg = struct.unpack_from('<HH', mem, 0x20960 + 0xA9EA)
        ptr = seg * 16 + off
        actual = tuple(mem[0x20960 + 0xA9F3:0x20960 + 0xA9F6])
        assert actual == pos, (name, actual, pos)
        assert mem[ptr:ptr + 0x400] == body, (name, 'GEO mismatch')
        results.append(dict(name=name, memory_sha256=sha(path), pointer=f'{seg:04X}:{off:04X}',
                            position=actual, payload_sha256=hashlib.sha256(body).hexdigest()))
        raw = a.out / f'{name}.raw'
        if raw.exists():
            im = Image.frombytes('P', (320, 200), raw.read_bytes())
            im.putpalette(raw.with_suffix('.pal').read_bytes())
            im.convert('RGB').resize((640, 400), Image.Resampling.NEAREST).save(a.out / f'{name}.png')
    inputs = {f: sha(a.orig / f) for f in ('GAME.EXE', 'GAME.OVR', 'ADLIB.COM', 'GEO1.DAX')}
    states = {f: sha(a.states / f) for f in ('g6-town/town.state', 'g6-explore/town1/end.state')}
    result = dict(address_space='dosgolem real-mode; data segment 2096h',
                  geo_pointer_offset='A9EA', position_offset='A9F3',
                  tool=(a.out / 'tool-version.txt').read_text().splitlines(),
                  inputs=inputs, states=states, samples=results,
                  rng='從既有正常路徑狀態重播；不重設或重擲亂數；本收據不驗證亂數規則')
    (a.out / 'evidence.json').write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
    print(f'GEO1.DAX {len(blocks)} 塊；{len(results)} 個取樣的地圖、座標與方向一致')


if __name__ == '__main__':
    main()
