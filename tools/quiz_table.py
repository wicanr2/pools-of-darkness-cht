#!/usr/bin/env python3
"""由 GAME.EXE 解開後的映像解出手冊查驗題題目表與答案（docs/re/008）。

題目表在資料段 `DS:0x60 + 0x62*k`（k = 1..21，`5509:0124` 取 Random(21)+1），每筆 0x62 bytes：
  +0x00 標題（string[30]，加密）、+0x1F 英文手札頁碼、+0x20..0x22 三個字序（0xFF 表示沒有）、
  +0x23、+0x38、+0x4D 三個答案（string[20]，加密）。
加密：Pascal 字串長度 L 不變，內容倒轉、每個字元加 L（解法見 `5509:0000`）。
以實機答對的「第 4 頁 ALIGNMENT 第 7 字 LIVES」（tools/g6-saveload-receipt.sh）定位，前後以 0x62 延伸到欄位不合為止，
不寫死檔案位置；筆數不是 21 就報錯。
用法（Docker sosb-tools 內）：python3 -I tools/quiz_table.py workplace/text-src/game-image.bin <輸出 TSV>
輸出 TSV：page、word、heading、answer，依頁、字序排序。
"""
import re
import sys
from pathlib import Path

REC = 0x62
COUNT = 21
ANSWER_AT = (0x23, 0x38, 0x4D)


def enc(s):
    n = len(s)
    return bytes([n]) + bytes((ord(c) + n) & 0xFF for c in reversed(s))


def dec(d, o, cap):
    n = d[o]
    if not 1 <= n <= cap:
        return None
    return ''.join(chr((d[o + i] - n) & 0xFF) for i in range(n, 0, -1))


def record(d, o):
    """解一筆；欄位不合就回 None。"""
    if o < 0 or o + REC > len(d):
        return None
    head = dec(d, o, 30)
    page = d[o + 0x1F]
    if not head or not re.fullmatch(r"[A-Z0-9][A-Z0-9 ,.'&/-]*", head) or not 1 <= page <= 99:
        return None
    qs = []
    for i, at in enumerate(ANSWER_AT):
        w = d[o + 0x20 + i]
        if w == 0xFF:
            continue
        a = dec(d, o + at, 20)
        if not 1 <= w <= 200 or not a or not re.fullmatch(r"[A-Z0-9'-]+", a):
            return None
        qs.append((w, a))
    return (page, head, qs) if qs else None


def main():
    d = Path(sys.argv[1]).read_bytes()
    key = enc('ALIGNMENT')
    hits = [i for i in range(len(d)) if d.startswith(key, i)]
    hits = [i for i in hits if record(d, i) and record(d, i)[0] == 4 and (7, 'LIVES') in record(d, i)[2]]
    if len(hits) != 1:
        sys.exit(f'題目表錨點命中 {len(hits)} 處')
    s = hits[0]
    while record(d, s - REC):
        s -= REC
    recs, o = [], s
    while record(d, o):
        recs.append(record(d, o))
        o += REC
    if len(recs) != COUNT:
        sys.exit(f'題目表 {len(recs)} 筆，預期 {COUNT} 筆（起點 0x{s:X}）')
    rows = sorted((p, w, h, a) for p, h, qs in recs for w, a in qs)
    with open(sys.argv[2], 'w', encoding='utf-8', newline='') as f:
        f.write('page\tword\theading\tanswer\n')
        for r in rows:
            f.write('\t'.join(map(str, r)) + '\n')
    print(f'題目表起點 0x{s:X}，{len(recs)} 個標題、{len(rows)} 題 → {sys.argv[2]}')


if __name__ == '__main__':
    main()
