#!/usr/bin/env python3
"""並列六張正常截圖的工具列原始像素，不縮放畫面。"""
import argparse
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

ap = argparse.ArgumentParser()
ap.add_argument('root', type=Path)
ap.add_argument('out', type=Path)
ap.add_argument('--label-font', type=Path, required=True)
a = ap.parse_args()
rows = [(1011, 3, '小'), (1014, 3, '中'), (1017, 3, '大'),
        (1024, 2, '小'), (1028, 2, '中'), (1032, 2, '大')]
font = ImageFont.truetype(str(a.label_font), 18)
sheet = Image.new('RGB', (960, sum(48 * scale + 30 for _, scale, _ in rows)), '#202020')
draw = ImageDraw.Draw(sheet)
y = 0
for frame, scale, size in rows:
    source = Image.open(a.root / f'pod-f{frame:06d}.png').convert('RGB')
    assert source.size == (480 * scale, 200 * scale)
    draw.text((4, y + 2), f'{scale} 倍，{size}', font=font, fill='white')
    crop = source.crop((80 * scale, 0, 400 * scale, 48 * scale))
    sheet.paste(crop, (0, y + 26))
    y += 48 * scale + 30
sheet.save(a.out)
