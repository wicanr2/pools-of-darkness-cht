#!/usr/bin/env python3
"""攻略摘要的照抄檢查（CONTEXT 決策 2026-10-23：雜誌原文不照抄）。

把雜誌掃描的 OCR 草稿（workplace/walkthrough/ocr/軟體世界NN期/*.json 的 result.rec_texts）每列去空白，
取含 6 個以上連續漢字、長度 8 以上的列，切成 10 字窗口，統計有多少窗口原樣出現在摘要（去空白）裡。
OCR 有錯字，數字是下限估計。比例超過 --max（預設 0.05）或有整列 12 字以上原樣出現就以結束碼 1 結束。
用法（Docker sosb-tools 內）：python3 -I tools/walkthrough_overlap.py <摘要 .md> <OCR 期目錄> [--max 0.05]
"""
import argparse
import json
import re
from pathlib import Path


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('summary', type=Path)
    ap.add_argument('ocr', type=Path)
    ap.add_argument('--max', type=float, default=0.05)
    a = ap.parse_args()
    s = re.sub(r'\s', '', a.summary.read_text(encoding='utf-8'))
    lines = []
    for f in sorted(a.ocr.glob('*.json')):
        lines += json.loads(f.read_text(encoding='utf-8'))['result']['rec_texts']
    cand = [x for x in (re.sub(r'\s', '', l) for l in lines) if len(x) >= 8 and re.search(r'[一-鿿]{6}', x)]
    whole = [x for x in cand if len(x) >= 12 and x in s]
    win = {x[i:i + 10] for x in cand for i in range(len(x) - 9)}
    hit = sorted(w for w in win if w in s)
    ratio = len(hit) / len(win) if win else 0.0
    print(f'{a.summary.name}：OCR 列 {len(cand)}、整列照抄 {len(whole)}；10 字窗口 {len(win)}、出現 {len(hit)}（{ratio:.1%}）')
    for w in (whole or hit)[:10]:
        print('  ', w)
    raise SystemExit(1 if whole or ratio > a.max else 0)


if __name__ == '__main__':
    main()
