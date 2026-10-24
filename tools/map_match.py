#!/usr/bin/env python3
"""研究用：比較提示書格線與 GEO 牆面，僅輸出候選，不寫正式資料。

參考矩形、格數與圖名由讀圖指定。score 只用於排序，人工核對後才可採用。
"""
import argparse
import json
from pathlib import Path
import sys

from PIL import Image

sys.path.insert(0, str(Path(__file__).resolve().parent))
from ecl_text_catalog import dax_blocks


def observed(im, rect, cols, rows):
    pix = im.convert('L').load()
    x0, y0, x1, y1 = rect
    sx, sy = (x1-x0)/cols, (y1-y0)/rows
    result = []
    # 略去邊線端點，避免牆角及門框支線；容忍掃描的輕微歪斜。
    for horizontal, across, along in ((True, rows+1, cols), (False, cols+1, rows)):
        for a in range(across):
            for b in range(along):
                values = []
                for j in range(3, 18):
                    t = j/20
                    x = x0 + (b+t)*sx if horizontal else x0+a*sx
                    y = y0 + a*sy if horizontal else y0+(b+t)*sy
                    darkest = 255
                    for delta in range(-5,6):
                        xx, yy = round(x+(0 if horizontal else delta)), round(y+(delta if horizontal else 0))
                        if 0<=xx<im.width and 0<=yy<im.height:
                            darkest = min(darkest,pix[xx,yy])
                    values.append(max(0,(210-darkest)/210))
                result.append(sum(values)/len(values))
    return result


def edges(body, x0, y0, cols, rows):
    def wall(x,y,k):
        i=(y%16)*16+x%16
        return (body[i]>>4,body[i]&15,body[256+i]>>4,body[256+i]&15)[k] != 0
    out=[]
    for y in range(rows+1):
        for x in range(cols):
            out.append((y<rows and wall(x0+x,y0+y,0)) or (y>0 and wall(x0+x,y0+y-1,2)))
    for x in range(cols+1):
        for y in range(rows):
            out.append((x<cols and wall(x0+x,y0+y,3)) or (x>0 and wall(x0+x-1,y0+y,1)))
    return out


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--orig',type=Path,required=True)
    ap.add_argument('--refs',type=Path,required=True)
    ap.add_argument('--images',type=Path,required=True)
    ap.add_argument('--out',type=Path,required=True)
    a=ap.parse_args()
    blocks=[(f'1/{bid:02x}',block[2:]) for bid,block in dax_blocks((a.orig/'GEO1.DAX').read_bytes())]
    results=[]
    for ref in json.loads(a.refs.read_text()):
        im=Image.open(a.images/f'hint-{ref["page"]:02d}.png')
        cols,rows=ref['cols'],ref['rows']
        obs=observed(im,ref['rect'],cols,rows)
        candidates=[]
        for key,body in blocks:
            origins=ref.get('origins') or [[x,y] for y in range(17-rows) for x in range(17-cols)]
            for x,y in origins:
                shape=edges(body,x,y,cols,rows)
                yes=[v for v,w in zip(obs,shape) if w]
                no=[v for v,w in zip(obs,shape) if not w]
                if not yes or not no: continue
                hit=sum(yes)/len(yes); blank=sum(no)/len(no)
                candidates.append(dict(map=key,x=x,y=y,score=round(hit-blank,4),hit=round(hit,4),blank=round(blank,4)))
        candidates.sort(key=lambda c:c['score'],reverse=True)
        result=dict(reference=ref,candidates=candidates[:5]);results.append(result)
        print(ref['name'],candidates[:3])
    a.out.write_text(json.dumps(results,ensure_ascii=False,indent=2)+'\n')


if __name__=='__main__':main()
