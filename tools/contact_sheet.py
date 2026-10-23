"""掃描縮圖拼版：判讀頁序、缺頁與冊別用。

用法（Docker sosb-img 內）：python3 -I tools/contact_sheet.py <輸出目錄> <每張欄數> <縮圖寬> <檔案 ...>
依參數順序排版，每張最多 欄數×4 格，每格下方標檔名。輸出 sheet-NN.png。
"""
import sys
from pathlib import Path
from PIL import Image, ImageDraw

out, cols, w = Path(sys.argv[1]), int(sys.argv[2]), int(sys.argv[3])
files = [Path(f) for f in sys.argv[4:]]
rows, label = 4, 18
out.mkdir(parents=True, exist_ok=True)
per = cols * rows
for n in range(0, len(files), per):
    chunk = files[n:n + per]
    thumbs = []
    for f in chunk:
        im = Image.open(f).convert("RGB")
        im.thumbnail((w, w))
        thumbs.append((f.name, im))
    h = max(t.height for _, t in thumbs)
    sheet = Image.new("RGB", (cols * w, ((len(thumbs) + cols - 1) // cols) * (h + label)), "white")
    d = ImageDraw.Draw(sheet)
    for i, (name, t) in enumerate(thumbs):
        x, y = (i % cols) * w, (i // cols) * (h + label)
        sheet.paste(t, (x, y))
        d.text((x + 2, y + h + 2), name, fill="red")
    p = out / f"sheet-{n // per + 1:02d}.png"
    sheet.save(p)
    print(p, len(thumbs))
