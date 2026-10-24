#!/usr/bin/env python3
"""HD 原圖全幅轉 3 倍不透明 RGBA；動畫只替換來源差異遮罩外擴六像素內。"""
import argparse
import hashlib
import json
from pathlib import Path
import re

from PIL import Image, ImageFilter


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--req',type=Path,required=True)
    ap.add_argument('--out',type=Path,required=True)
    ap.add_argument('--keys',nargs='+')
    a=ap.parse_args()
    request=json.loads((a.req/'manifest.json').read_text())
    items=[i for i in request['items'] if not a.keys or i['key'] in a.keys]
    if a.keys and {i['key'] for i in items}!=set(a.keys):raise ValueError('缺指定來源')
    rows=[]
    for it in sorted(items,key=lambda i:bool(i.get('animation'))):
        key=it['key']
        raws=list(a.out.glob(key+'-r*-raw.png'))
        if not raws:raise ValueError(f'{key}: 缺生成原圖')
        raw=max(raws,key=lambda p:int(re.search(r'-r(\d+)-raw\.png$',p.name)[1]))
        job=a.out/'jobs'/(raw.name.replace('-raw.png','.json'))
        if job.exists() and json.loads(job.read_text())['raw_sha256']!=sha(raw):raise ValueError(f'{key}: raw changed')
        im=Image.open(raw).convert('RGBA').resize(tuple(it['output_size']),Image.Resampling.LANCZOS)
        im.putalpha(255)
        anim=it.get('animation')
        if anim:
            path=a.req/anim['mask']
            if sha(path)!=anim['mask_sha256']:raise ValueError(f'{key}: mask changed')
            base=Image.open(a.out/(anim['base']+'.png')).convert('RGBA')
            mask=Image.open(path).convert('L').filter(ImageFilter.MaxFilter(13))
            if base.size!=im.size or mask.size!=im.size:raise ValueError('動畫尺寸不合')
            im=Image.composite(im,base,mask)
        preparation='whole-image LANCZOS; opaque RGBA; source mask expanded 6 output pixels'
        if it['source_file']=='8X8D1.DAX':
            if it['source_size']!=[8,8]or it['output_size']!=[24,24]or it.get('transparent_index')!=13:
                raise ValueError('VGA牆面來源契約不符')
            idx_path,pal_path=a.req/it['indexed'],a.req/it['palette']
            if sha(idx_path)!=it['indexed_sha256']or sha(pal_path)!=it['palette_sha256']:
                raise ValueError('牆面來源雜湊改變')
            ref=idx_path.read_bytes();pal=pal_path.read_bytes()
            if len(ref)!=64 or len(pal)!=768:raise ValueError('牆面來源長度不合')
            for y in range(24):
                for x in range(24):
                    c=ref[y//3*8+x//3]
                    if c==13:im.putpixel((x,y),(85,85,85,255))
                    elif x in(0,23)or y in(0,23):im.putpixel((x,y),tuple(pal[c*3:c*3+3])+(255,))
            preparation='whole-image LANCZOS; opaque RGBA; source13 neutral-gray mask; outer1px source RGB seam lock'
        dst=a.out/(key+'.png');im.save(dst)
        rows.append(dict(key=key,raw=raw.name,raw_sha256=sha(raw),png=dst.name,png_sha256=sha(dst),
                         preparation=preparation))
        print(key,im.size,'RGBA')
    p=a.out/'prepared.json'
    old=json.loads(p.read_text())['items'] if p.exists() else []
    merged={r['key']:r for r in old}
    merged.update({r['key']:r for r in rows})
    p.write_text(json.dumps(dict(batch=request['batch'],items=list(merged.values())),ensure_ascii=False,indent=2)+'\n')


if __name__=='__main__':main()
