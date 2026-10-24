#!/usr/bin/env python3
"""由實際 GEO archive 清單產生地圖雜湊表，檢查攻略點；Docker 內執行。

沿用銀色匕首之謎 map_geo.py 的資料契約，不假定有 GEO2/3.DAX。
--check 只比對重建結果，不覆寫；原版解碼 bytes 不寫入正式表。
"""
import argparse
import csv
import hashlib
from pathlib import Path
import sys

sys.path.insert(0, str(Path(__file__).resolve().parent))
from ecl_text_catalog import dax_blocks


def read(path, header):
    with path.open(encoding='utf-8', newline='') as f:
        reader = csv.DictReader(f, delimiter='\t')
        if reader.fieldnames != header:
            raise ValueError(f'{path}: 表頭應為 {header}')
        rows = list(reader)
    if any(None in r or None in r.values() for r in rows):
        raise ValueError(f'{path}: 欄數錯誤')
    return rows


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--orig', type=Path, required=True)
    ap.add_argument('--titles', type=Path, required=True)
    ap.add_argument('--guide', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--check', action='store_true')
    a = ap.parse_args()
    titles = {}
    for r in read(a.titles, ['map', 'title', 'evidence']):
        if r['map'] in titles or not r['title'] or not r['evidence']:
            raise ValueError(f'名稱重複或缺證據：{r}')
        titles[r['map']] = r['title']
    archives = sorted(a.orig.glob('GEO[0-9]*.DAX'))
    if not archives:
        raise ValueError('缺 GEO archive')
    rows = ['map\tpayload_sha256\ttitle']
    keys, sums = set(), set()
    for path in archives:
        n = int(path.stem[3:])
        for bid, block in dax_blocks(path.read_bytes()):
            key = f'{n}/{bid:02x}'
            if len(block) != 0x402 or key in keys:
                raise ValueError(f'{key}: 長度 {len(block)} 或重複鍵')
            digest = hashlib.sha256(block[2:]).hexdigest()
            if digest in sums:
                raise ValueError(f'{key}: 重用 GEO 需先研究 ECL 區分，不可覆蓋先前地圖')
            keys.add(key)
            sums.add(digest)
            rows.append(f'{key}\t{digest}\t{titles.get(key, "")}')
    if titles.keys() - keys:
        raise ValueError(f'名稱找不到 GEO：{titles.keys() - keys}')
    points = read(a.guide, ['map', 'n', 'x', 'y', 'label', 'summary', 'source'])
    seen = set()
    for r in points:
        ident = (r['map'], int(r['n']), int(r['x']), int(r['y']))
        key, n, x, y = ident
        if ident in seen or key not in titles or not 1 <= n <= 99 or not (0 <= x < 16 and 0 <= y < 16):
            raise ValueError(f'攻略點重複、未命名或越界：{r}')
        if not r['label'] or len(r['label']) > 8 or not r['summary'] or not r['source']:
            raise ValueError(f'攻略點缺文、名稱過長或缺出處：{r}')
        seen.add(ident)
    result = '\n'.join(rows) + '\n'
    if a.check:
        if a.out.read_text(encoding='utf-8') != result:
            raise ValueError('地圖表與原版 GEO／名稱表不一致，請重建')
    else:
        a.out.write_text(result, encoding='utf-8')
    print(f'{len(archives)} 個 GEO archive、{len(keys)} 塊、{len(titles)} 個已命名地圖、{len(points)} 個攻略格：通過')


if __name__ == '__main__':
    main()
