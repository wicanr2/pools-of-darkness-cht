"""把 cmd/probe -shots 存的色號陣列（.raw）與同名 .pal 轉成 PNG。

用法（Docker sosb-img 內）：python3 -I tools/indexed_png.py <寬> <高> <x.raw> [...]
色盤值都 ≤ 63 時視為 VGA DAC 的 6 位元值，乘 255/63。輸出同名 .png。
"""
import sys
from pathlib import Path
from PIL import Image

w, h = int(sys.argv[1]), int(sys.argv[2])
for raw in map(Path, sys.argv[3:]):
    data = raw.read_bytes()[: w * h]
    pal = raw.with_suffix(".pal").read_bytes()
    if pal and max(pal) <= 63:
        pal = bytes(v * 255 // 63 for v in pal)
    im = Image.frombytes("P", (w, h), data)
    im.putpalette(pal[:768].ljust(768, b"\0"))
    out = raw.with_suffix(".png")
    im.convert("RGB").save(out)
    print(out)
