"""依 OCR 文字框裁出掃描局部（原解析度），校讀難辨字用。

用法（Docker sosb-img 內）：python3 -I tools/ocr_crop.py <OCR json 目錄> <掃描目錄> <輸出目錄> <標籤>=<檔名>:<要找的字串> ...
在該張掃描的 rec_texts 裡找含字串的文字框（找不到就用前兩字再找），上下各擴 40 像素裁出，存 <標籤>.png，並印出 OCR 讀到的那一行。
"""
import json
import sys
from pathlib import Path
from PIL import Image

ocr, scans, out = map(Path, sys.argv[1:4])
out.mkdir(parents=True, exist_ok=True)
for spec in sys.argv[4:]:
    label, rest = spec.split('=', 1)
    name, needle = rest.split(':', 1)
    d = json.loads((ocr / f'{name}.json').read_text(encoding='utf-8'))['result']
    hits = [(t, b) for t, b in zip(d['rec_texts'], d['rec_boxes']) if needle in t]
    if not hits:
        hits = [(t, b) for t, b in zip(d['rec_texts'], d['rec_boxes']) if needle[:2] in t]
    if not hits:
        print(label, '找不到', needle)
        continue
    t, (x0, y0, x1, y1) = hits[0]
    im = Image.open(scans / name)
    im.crop((max(0, x0 - 20), max(0, y0 - 40), min(im.width, x1 + 20), min(im.height, y1 + 40))).save(out / f'{label}.png')
    print(label, name, (x0, y0, x1, y1), t)
