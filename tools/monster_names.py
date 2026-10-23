#!/usr/bin/env python3
"""列出怪物資料檔 MON1CHA.DAX 各區塊的怪物名（G5；AGENTS.md 第 11 節：怪物譯名先查怪物資料檔）。在 sosb-tools 容器內以 python3 -I 執行。

用法：monster_names.py <MON1CHA.DAX> [--out TSV]
DAX 解壓沿用 tools/ecl_text_catalog.py 的 dax_blocks。區塊開頭是 Pascal 字串（長度 byte＋最多 15 字元，同 Gold Box 角色紀錄的名字欄）；
長度超出範圍或含非可列印字元的區塊照列並標「格式不符」。輸出 TSV：block、size、name、status。
"""
import argparse
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from ecl_text_catalog import dax_blocks  # noqa: E402


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('dax', type=Path)
    ap.add_argument('--out', type=Path)
    a = ap.parse_args()
    lines = ['block\tsize\tname\tstatus']
    for bid, blk in dax_blocks(a.dax.read_bytes()):
        n = blk[0] if blk else 0
        raw = blk[1:1 + n]
        ok = 0 < n <= 15 and all(0x20 <= c < 0x7F for c in raw)
        name = raw.decode('ascii') if ok else ''
        lines.append(f'{bid}\t{len(blk)}\t{name}\t{"" if ok else "格式不符"}')
    text = '\n'.join(lines) + '\n'
    if a.out:
        a.out.write_text(text, encoding='utf-8')
    sys.stdout.write(text)


if __name__ == '__main__':
    main()
