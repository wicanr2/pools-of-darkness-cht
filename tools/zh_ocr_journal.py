#!/usr/bin/env python3
"""由珍157 手札跨頁掃描的 PaddleOCR 結果組出各則草稿（docs/re/005）。改寫自銀色匕首之謎 tools/zh_ocr_journal.py。

輸入 tools/ocr_batch.py 的逐張 JSON（每張是一個跨頁，左頁頁碼較小），依頁序給 `<左頁碼>=<JSON 路徑>`。
書溝取畫面中央 500–780 像素內被文字框覆蓋最少的位置；文字框中心在書溝左邊算左頁。
頁首（「黑暗之池」「冒險者手冊」）與頁尾（頁碼）依 y 位置略過。「旅行手札 N」開新的一則；
其後第一行若不是以「開頭、也沒有縮排就當標題；縮排（x 明顯大於該頁左緣）起新段落，段落以兩個字元 \\n 相接。
OCR 草稿未校對，要對頁面圖校讀。輸出 TSV：entry、page、title、body、note。
"""
import argparse
import json
import re
import sys
from pathlib import Path

HEAD = re.compile(r'^旅行手[札扎礼紮]\s*(\d+)')


def boxes(path):
    d = json.loads(Path(path).read_text(encoding='utf-8'))['result']
    return [(t.strip(), b) for t, b in zip(d['rec_texts'], d['rec_boxes']) if t.strip()]


def gutter(bs):
    cover = {x: 0 for x in range(500, 781)}
    for _, (x0, _, x1, _) in bs:
        for x in range(max(500, x0), min(780, x1) + 1):
            cover[x] += 1
    return min(cover, key=lambda x: (cover[x], abs(x - 640)))


def page_lines(bs, h):
    out = []
    for t, (x0, y0, x1, y1) in bs:
        if y1 < 0.09 * h or y0 > 0.92 * h or re.fullmatch(r'[\d\sA-Z]*', t):
            continue
        out.append((y0, x0, t))
    out.sort()
    left = min((x for _, x, _ in out), default=0)
    return [(x - left, t) for _, x, t in out]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--spreads', required=True, nargs='+', help='左頁碼=JSON 路徑，依頁序')
    ap.add_argument('--out', required=True, type=Path)
    a = ap.parse_args()
    ents, cur = {}, None
    for spec in a.spreads:
        page, path = spec.split('=', 1)
        bs = boxes(path)
        h = max(b[3] for _, b in bs)
        g = gutter(bs)
        halves = ([x for x in bs if (x[1][0] + x[1][2]) / 2 < g], [x for x in bs if (x[1][0] + x[1][2]) / 2 >= g])
        for side, part in enumerate(halves):
            pno = int(page) + side
            for dx, t in page_lines(part, h):
                m = HEAD.match(t.replace(' ', ''))
                if m:
                    cur = int(m.group(1))
                    ents[cur] = {'page': pno, 'title': None, 'paras': []}
                    continue
                if cur is None:
                    continue
                e = ents[cur]
                if e['title'] is None and not e['paras']:
                    if not t.startswith('「') and dx < 20:
                        e['title'] = t
                        continue
                    e['title'] = ''
                if dx >= 20 or not e['paras']:
                    e['paras'].append(t)
                else:
                    e['paras'][-1] += t
    with open(a.out, 'w', encoding='utf-8', newline='') as f:
        f.write('entry\tpage\ttitle\tbody\tnote\n')
        for n in sorted(ents):
            e = ents[n]
            body = '\\n'.join(e['paras'])
            f.write(f"{n}\t{e['page']}\t{e['title'] or ''}\t{body}\tOCR 草稿未校對\n")
    missing = sorted(set(range(1, 101)) - set(ents))
    print(f'{len(ents)} 則；缺 {missing}', file=sys.stderr)


if __name__ == '__main__':
    main()
