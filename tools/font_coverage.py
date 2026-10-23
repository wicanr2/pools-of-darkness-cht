#!/usr/bin/env python3
"""字型覆蓋檢查（AGENTS.md 第 6 節）：譯文表要顯示的字都要在 GOLEMFNT 字型裡。在 sosb-tools 容器內以 python3 -I 執行。

用法：font_coverage.py --font <GOLEMFNT> <譯文 TSV …>
要顯示的欄與字集規則同 tools/eten_font.py 的 charset（translation、title、body 等欄，加可列印 ASCII）。
GOLEMFNT：magic 'GOLEMFNT'、LE u16 W、u16 H、u32 字數，每字 u32 碼點＋u8 來源標記＋字模（tools/rounded_font.py）。
缺字逐一列出（字、碼點、第一個出現的檔案）並以結束碼 1 結束。
"""
import argparse
import csv
import struct
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from eten_font import SHOWN  # noqa: E402


def font_chars(path):
    b = Path(path).read_bytes()
    if b[:8] != b'GOLEMFNT':
        raise SystemExit(f'{path}: 不是 GOLEMFNT')
    w, h, n = struct.unpack('<HHI', b[8:16])
    size = 5 + ((w + 7) // 8) * h
    return {struct.unpack('<I', b[16 + i * size:20 + i * size])[0] for i in range(n)}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--font', required=True)
    ap.add_argument('catalogs', nargs='+')
    a = ap.parse_args()
    have = font_chars(a.font)
    missing = {}
    for p in a.catalogs:
        rows = list(csv.reader(open(p, encoding='utf-8'), delimiter='\t', quoting=csv.QUOTE_NONE))
        cols = [i for i, h in enumerate(rows[0] if rows else []) if h in SHOWN]
        for r in rows[1:]:
            for c in cols:
                for ch in (r[c] if c < len(r) else ''):
                    if ch != '\t' and ord(ch) not in have:
                        missing.setdefault(ch, p)
    print(f'字型 {len(have)} 字；譯文表 {len(a.catalogs)} 個；缺 {len(missing)} 字')
    if missing:
        for ch, p in sorted(missing.items()):
            print(f'  {ch} U+{ord(ch):04X} {p}', file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()
