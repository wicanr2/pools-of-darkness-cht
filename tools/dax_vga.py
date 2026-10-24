#!/usr/bin/env python3
"""本作 VGA BIGPIC／TITLE／PIC 研究圖。只寫 ignored 工作區；執行期另驗。

偏移契約與外部線索見 docs/re/014-vga-hd.md。PIC 使用獨立的第二層解碼；SPRIT 尚未支援。
"""
import argparse
import hashlib
import json
from pathlib import Path
import struct
import sys

from PIL import Image, ImageDraw

sys.path.insert(0, str(Path(__file__).resolve().parent))
from ecl_text_catalog import dax_blocks


def sha(data):
    return hashlib.sha256(data).hexdigest()


def block_vga(data, base_palette=None):
    if len(data) < 10:
        raise ValueError('短於表頭')
    h, w = data[0], data[1] * 8
    count = struct.unpack_from('<H', data, 6)[0]
    base, colors = data[8], data[9] + 1
    end_palette = 10 + colors * 3
    start_pixels = len(data) - w * h * count
    if not (0 < w <= 320 and 0 < h <= 200 and 0 < count <= 64 and
            base + colors <= 256 and end_palette <= start_pixels):
        raise ValueError(f'尺寸／色盤／像素範圍不合：{w}×{h}，{count} 幀')
    dac = data[10:end_palette]
    if max(dac) > 63:
        raise ValueError('色盤超過 6 bit')
    # dosgolem 色值轉換另以執行期圖核對；此處用 6-bit 位元展開。
    if base_palette is not None and len(base_palette) != 768:
        raise ValueError('執行期 RGB 色盤須為 768 bytes')
    palette = list(base_palette) if base_palette is not None else [0] * 768
    palette[base*3:(base+colors)*3] = [(v << 2) | (v >> 4) for v in dac]
    frames = []
    for i in range(count):
        px = data[start_pixels + i*w*h:start_pixels + (i+1)*w*h]
        uncovered = sorted(set(px) - set(range(base, base+colors)))
        if uncovered and base_palette is None:
            raise ValueError(f'影像用到未定義色盤 {uncovered[:12]}，不得猜色')
        frames.append(px)
    return w, h, frames, bytes(palette), start_pixels, end_palette


def sprite_vga(data, base_palette=None):
    """PIC 的第二層 RLE 及 XOR 候選，來源 VgaSpriteBlock.cs；執行期另驗。"""
    if len(data)<11:
        raise ValueError('PIC 表頭不足')
    h,units=struct.unpack_from('<HH',data)
    w=units*8
    extra,base,colors=data[8:11]
    start=colors*3+132+3*extra
    end_palette=11+colors*3
    if not (0<w<=320 and 0<h<=200 and colors>0 and base+colors<=256 and end_palette<start<len(data)):
        raise ValueError('PIC 尺寸／色盤／資料範圍錯誤')
    dac=data[11:end_palette]
    if max(dac)>63 or base_palette is None or len(base_palette)!=768:
        raise ValueError('PIC 色盤或執行期保留色不足')
    palette=bytearray(base_palette)
    palette[base*3:(base+colors)*3]=bytes((v<<2)|(v>>4) for v in dac)
    out=bytearray();p=start
    while p<len(data):
        c=data[p];p+=1
        if c<128:
            n=c+1
            if p+n>len(data):raise ValueError('PIC literal 越界')
            out.extend(data[p:p+n]);p+=n
        else:
            if p==len(data):raise ValueError('PIC repeat 越界')
            out.extend([data[p]]*(257-c));p+=1
        if len(out)>64*w*h:raise ValueError('PIC 解壓超過上限')
    size=w*h
    if len(out)%size or not 0<len(out)//size<=64:
        raise ValueError(f'PIC 解壓長度 {len(out)} 不合 {w}×{h}')
    frame_count=len(out)//size
    main=data[colors*3+127]
    if main>=frame_count:raise ValueError('PIC 基底幀越界')
    frames=[]
    baseline=out[main*size:(main+1)*size]
    for i in range(frame_count):
        frame=bytes(out[i*size:(i+1)*size])
        frames.append(frame if i==main else bytes(x^y for x,y in zip(frame,baseline)))
    return w,h,frames,bytes(palette),start,end_palette


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--orig', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--base-palette', type=Path, help='dosgolem 原版收據的 768-byte RGB 色盤，供資產外保留色使用')
    a = ap.parse_args()
    a.out.mkdir(parents=True, exist_ok=True)
    items, tiles = [], []
    base_palette = a.base_palette.read_bytes() if a.base_palette else None
    for name in ('BIGPIC1.DAX', 'TITLE.DAX', 'PIC1.DAX'):
        path = a.orig / name
        raw = path.read_bytes()
        for bid, decoded in dax_blocks(raw):
            parser=sprite_vga if name=='PIC1.DAX' else block_vga
            w, h, frames, palette, start, pal_end = parser(decoded, base_palette)
            for frame, px in enumerate(frames):
                key = f'{path.stem.lower()}-{bid:03d}' + (f'@{frame}' if len(frames)>1 else '')
                im = Image.frombytes('P', (w,h), px)
                im.putpalette(palette)
                im.convert('RGB').save(a.out / f'{key}.png')
                (a.out / f'{key}.idx').write_bytes(px)
                (a.out / f'{key}.pal').write_bytes(palette)
                items.append(dict(key=key, file=name, archive_sha256=sha(raw), block=bid, frame=frame,
                                  width=w, height=h, decoded_sha256=sha(decoded), indexed_sha256=sha(px),
                                  palette_sha256=sha(palette), pixel_offset=start, palette_end=pal_end,
                                  evidence_level='strong inference', runtime_observed=False))
                tile=Image.new('RGB',(640,h*2+28),'#242424')
                ImageDraw.Draw(tile).text((8,7),key,fill='white')
                tile.paste(im.convert('RGB').resize((w*2,h*2),Image.Resampling.NEAREST),(8,28))
                tiles.append(tile)
    sheet=Image.new('RGB',(1280,sum(max(t.height for t in tiles[i:i+2]) for i in range(0,len(tiles),2))),'#242424')
    y=0
    for i in range(0,len(tiles),2):
        for x,t in enumerate(tiles[i:i+2]):sheet.paste(t,(x*640,y))
        y+=max(t.height for t in tiles[i:i+2])
    sheet.save(a.out/'atlas.png')
    (a.out/'inventory.json').write_text(json.dumps({'items':items,
        'base_palette_sha256':sha(base_palette) if base_palette else None},ensure_ascii=False,indent=2)+'\n')
    print(f'{len(items)} 個 VGA 影像狀態；執行期依觀測收據分別驗證')


if __name__ == '__main__':
    main()
