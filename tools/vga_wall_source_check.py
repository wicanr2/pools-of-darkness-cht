#!/usr/bin/env python3
"""核對既有正常街景的天空／8×8牆面來源，只產生本機研究資料。"""
import argparse
import hashlib
import json
from pathlib import Path
import struct
import sys

import numpy as np
from PIL import Image

sys.path.insert(0, str(Path(__file__).resolve().parent))
from dax_vga import block_vga
from ecl_text_catalog import dax_blocks


def sha(data):
    return hashlib.sha256(data).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for name in ('orig', 'capture', 'binary', 'out'):
        ap.add_argument('--' + name, type=Path, required=True)
    a = ap.parse_args()
    assert not a.out.exists(), '保留舊證據，請指定新的輸出目錄'
    indexed = (a.capture / 'source.idx').read_bytes()
    palette = (a.capture / 'source.pal').read_bytes()
    assert len(indexed) == 64000 and len(palette) == 768
    screen = np.frombuffer(indexed, np.uint8).reshape(200, 320)
    actual_rgb = np.frombuffer(palette, np.uint8).reshape(256, 3)
    archives = {name: (a.orig / name).read_bytes() for name in ('8X8D1.DAX', 'WALLDEF1.DAX', 'SKYGRND.DAX')}
    banks = {}
    for bid, data in dax_blocks(archives['8X8D1.DAX']):
        count = struct.unpack_from('<H', data, 6)[0]
        first, colors = data[8], data[9] + 1
        start = len(data) - count * 64
        assert data[:2] == bytes((8, 1)) and count == 256
        assert first == 32 and colors == 112 and len(data) == 16818 and start == 434
        dac = np.frombuffer(data[10:10 + colors * 3], np.uint8).reshape(colors, 3)
        assert np.all(dac < 64)
        rgb = (dac << 2) | (dac >> 4)
        banks[bid] = dict(data=data, tiles=np.frombuffer(data[start:], np.uint8).reshape(count, 8, 8),
                          rgb=rgb, start=start, palette_equal=bool(np.array_equal(rgb, actual_rgb[32:144])))
    assert set(banks) == set(range(1, 7))
    sky_matches = []
    skies = {}
    for bid, data in dax_blocks(archives['SKYGRND.DAX']):
        w, h, frames, pal, _, _ = block_vga(data, palette)
        assert (w, h, len(frames)) == (88, 88, 1) and data[8:10] == bytes((144, 31))
        skies[bid] = (frames[0], pal)
        if pal[144*3:176*3] == palette[144*3:176*3]:
            sky_matches.append(bid)
    assert sky_matches == [1] and [k for k, b in banks.items() if b['palette_equal']] == [1]
    sky = np.frombuffer(skies[1][0], np.uint8).reshape(88, 88)
    tiles = banks[1]['tiles']
    cells, used = [], set()
    for cy in range(11):
        for cx in range(11):
            at = [24 + cx * 8, 24 + cy * 8, 8, 8]
            background = sky[cy*8:cy*8+8, cx*8:cx*8+8]
            observed = screen[at[1]:at[1]+8, at[0]:at[0]+8]
            row = dict(position=cy*11+cx, rect=at, indexed_sha256=sha(observed.tobytes()), sky='skygrnd-001')
            if not np.array_equal(background, observed):
                hits = [i for i, tile in enumerate(tiles)
                        if np.array_equal(np.where(tile == 13, background, tile), observed)]
                assert len(hits) == 1, f'{at}: 來源缺少或歧義：{hits}'
                tile = hits[0]
                used.add(tile)
                row['wall'] = f'8x8d1-001@{tile}'
                row['wall_decoded_pixel_offset'] = banks[1]['start'] + tile * 64
            cells.append(row)
    assert len(cells) == 121 and sum('wall' in c for c in cells) == 31 and len(used) == 17
    a.out.mkdir(parents=True)
    def save_source(key, pixels, w, h, transparent=None):
        (a.out / (key + '.idx')).write_bytes(pixels)
        (a.out / (key + '.pal')).write_bytes(palette)
        reference = actual_rgb[np.frombuffer(pixels, np.uint8).reshape(h, w)].copy()
        if transparent is not None:
            reference[np.frombuffer(pixels, np.uint8).reshape(h, w) == transparent] = (85, 85, 85)
        Image.fromarray(reference).save(a.out / (key + '.png'))
        Image.fromarray(reference).resize((w*3, h*3), Image.Resampling.NEAREST).save(a.out / (key + '-3x.png'))
    save_source('skygrnd-001', skies[1][0], 88, 88)
    for tile in sorted(used):
        save_source(f'8x8d1-001@{tile}', tiles[tile].tobytes(), 8, 8, 13)
    inventory = [dict(key=f'8x8d1-001@{tile}', file='8X8D1.DAX',
                      archive_sha256=sha(archives['8X8D1.DAX']), block=1, frame=tile,
                      width=8, height=8, decoded_sha256=sha(banks[1]['data']),
                      indexed_sha256=sha(tiles[tile].tobytes()), palette_sha256=sha(palette),
                      pixel_offset=434+tile*64, palette_end=346,
                      evidence_level='confirmed in captured viewport', runtime_observed=True, transparent_index=13)
                 for tile in sorted(used)]
    (a.out / 'inventory-walls.json').write_text(json.dumps(dict(items=inventory), ensure_ascii=False, indent=2) + '\n')
    report = dict(scope='同一既有G8冷啟動讀檔4800格終點；只證明該88×88街景的完整來源組合',
                  evidence='已證實：本次色號、來源區段RGB及13留底組合逐像素吻合；未知：其他位置與WALLDEF內部繪製規則',
                  tool='Python 3.13.16 / Pillow 11.0 / NumPy; sosb-img:r1; dosgolem pod-play-readme f3d019e來源',
                  inputs={**{name: sha(data) for name, data in archives.items()},
                          'indexed_sha256': sha(indexed), 'palette_sha256': sha(palette),
                          'binary_sha256': sha(a.binary.read_bytes()), 'run_log_sha256': sha((a.capture / 'run.log').read_bytes())},
                  viewport=[24, 24, 88, 88], wall_bank=1, sky_block=1, transparent_wall_index=13,
                  wall_palette=dict(start=32, count=112, sha256=sha(palette[32*3:144*3])),
                  sky_palette=dict(start=144, count=32, sha256=sha(palette[144*3:176*3])),
                  wall_block_size=16818, tile_pixel_offset=434, tile_count_per_bank=256,
                  palette_end_offset=346, unknown_gap=[346, 434],
                  sky_only_cells=90, wall_cells=31, unique_wall_tiles=sorted(used), cells=cells)
    (a.out / 'source-proof.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
    print('街景121格完全吻合：90格天空／地面、31格牆面，17個不同牆圖塊。沒有接入產品路徑。')


if __name__ == '__main__':
    main()
