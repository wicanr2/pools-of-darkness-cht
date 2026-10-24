#!/usr/bin/env python3
"""從相同原型的不透明成品與無字底圖匯出文字差分層，保留定版 LCD 字緣。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--dir',type=Path,required=True)
    a=ap.parse_args()
    rows=[]
    for full in sorted(a.dir.glob('poster-full-*.png')):
        side,code=full.stem.removeprefix('poster-full-').split('.',1)
        lang,scale=code.rsplit('-',1)
        base=a.dir/f'poster-{side}-{scale}.png'
        original=np.array(Image.open(full).convert('RGBA'))
        background=np.array(Image.open(base).convert('RGBA'))
        if original.shape!=background.shape:raise ValueError('海報底圖與語系層尺寸不同')
        changed=np.any(original!=background,axis=2)
        layer=original.copy()
        layer[~changed]=0
        target=a.dir/f'poster-text-{side}.{lang}-{scale}.png'
        Image.fromarray(layer).save(target)
        result=Image.alpha_composite(Image.fromarray(background),Image.fromarray(layer))
        if not np.array_equal(np.asarray(result),original):raise AssertionError('文字層合成偏離原定版')
        rows.append(dict(file=target.name,sha256=hashlib.sha256(target.read_bytes()).hexdigest(),
                         base=base.name,base_sha256=hashlib.sha256(base.read_bytes()).hexdigest(),
                         full=full.name,full_sha256=hashlib.sha256(full.read_bytes()).hexdigest(),
                         changed_pixels=int(changed.sum()),exact_composition=True))
    if len(rows)!=8:raise ValueError('預期兩語系、兩側、兩倍率共八層')
    (a.dir/'layers.json').write_text(json.dumps(dict(method='opaque DOM render delta; unchanged pixels transparent; preserve approved LCD edges',items=rows),indent=2)+'\n')
    print('八個文字差分層合成與不透明原型完全一致')


if __name__=='__main__':main()
