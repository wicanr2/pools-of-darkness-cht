#!/usr/bin/env python3
"""由珍157 跨頁掃描的 PaddleOCR 結果組出逐頁轉錄草稿（docs/re/006；改寫自銀色匕首之謎 tools/zh_manual_draft.py）。
在 sosb-tools 容器內以 python3 -I 執行。

用法：zh_manual_draft.py --book <冊名> --spreads <左頁碼>=<OCR JSON> … --out <TSV> [--append]
書溝與左右頁沿用 tools/zh_ocr_journal.py；頁首（書名花邊）與頁尾（頁碼）依 y 位置略過；行的組法同 tools/zh157_names.py。
輸出 TSV：book、page、text（行以字面 \\n 相接）、status（固定「OCR 草稿未校對」）。草稿含原書文字，只放 workplace/。
"""
import argparse
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from zh_ocr_journal import boxes, gutter  # noqa: E402
from zh157_names import lines  # noqa: E402


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--book', required=True)
    ap.add_argument('--spreads', required=True, nargs='+')
    ap.add_argument('--out', required=True, type=Path)
    ap.add_argument('--append', action='store_true')
    a = ap.parse_args()
    rows = [] if a.append and a.out.exists() else ['book\tpage\ttext\tstatus']
    for spec in a.spreads:
        left, path = spec.split('=', 1)
        bs = boxes(path)
        h = max(b[3] for _, b in bs)
        g = gutter(bs)
        bs = [x for x in bs if 0.09 * h < x[1][3] and x[1][1] < 0.92 * h]
        for side in (0, 1):
            ls = lines([x for x in bs if ((x[1][0] + x[1][2]) / 2 < g) == (side == 0)])
            text = '\\n'.join(t.replace('\t', ' ') for t in ls)
            rows.append(f'{a.book}\t{int(left) + side}\t{text}\tOCR 草稿未校對')
    with a.out.open('a' if a.append and a.out.exists() else 'w', encoding='utf-8') as f:
        f.write('\n'.join(rows) + '\n')
    print(f'{a.book}：{len(a.spreads) * 2} 頁', file=sys.stderr)


if __name__ == '__main__':
    main()
