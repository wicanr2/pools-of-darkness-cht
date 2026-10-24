#!/usr/bin/env python3
"""以正常前端的配對色號／色盤核對 CPIC 前景與戰鬥地形，不修改遊戲。"""
import argparse
import hashlib
import json
from pathlib import Path
import sys

import numpy as np
sys.path.insert(0,str(Path(__file__).resolve().parent))
from ecl_text_catalog import dax_blocks


def sha(data):return hashlib.sha256(data).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for name in ('inventory','capture','orig','out'):ap.add_argument('--'+name,type=Path,required=True)
    a=ap.parse_args()
    raw=(a.capture/'battle.idx').read_bytes();rawpal=(a.capture/'battle.pal').read_bytes()
    if len(raw)!=64000 or len(rawpal)!=768:raise ValueError('原始畫格或色盤長度不合')
    screen=np.frombuffer(raw,np.uint8).reshape(200,320)
    palette=np.frombuffer(rawpal,np.uint8).reshape(256,3)
    inventory=json.loads((a.inventory/'inventory-other.json').read_text())
    families=('CPIC1.DAX','DUNGCOM.DAX','WILDCOM.DAX')
    blocks={name:dict(dax_blocks((a.orig/name).read_bytes())) for name in families}
    hits=[];banks={};archives={}
    for it in inventory['items']:
        family=it['file']
        if family not in families:continue
        archives[family]=sha((a.orig/family).read_bytes())
        if archives[family]!=it['archive_sha256']:raise ValueError('原始 archive 改變')
        key=it['key'];w,h=it['width'],it['height']
        source=(a.inventory/(key+'.idx')).read_bytes();sourcepal=(a.inventory/(key+'.pal')).read_bytes()
        if sha(source)!=it['indexed_sha256'] or sha(sourcepal)!=it['palette_sha256']:raise ValueError('盤點來源改變')
        reference=np.frombuffer(source,np.uint8).reshape(h,w)
        colors=np.frombuffer(sourcepal,np.uint8).reshape(256,3)
        header=blocks[family][it['block']]
        start,count=header[8],header[9]+1
        bankkey=key.split('@')[0] if family!='CPIC1.DAX' else 'cpic1'
        value=dict(start=start,count=count,sha256=sha(sourcepal[start*3:(start+count)*3]))
        if bankkey in banks and banks[bankkey]!=value:raise ValueError('同一圖族／區塊色盤不同')
        banks[bankkey]=value
        for flip in ([False,True]if family=='CPIC1.DAX'else[False]):
            ref=reference[:,::-1]if flip else reference
            mask=ref!=8 if family=='CPIC1.DAX'else np.ones((h,w),bool)
            if mask.sum()<16:continue
            # 單色地形不提供位置辨識證據，也不作進入戰鬥的守衛。
            if family!='CPIC1.DAX' and len(np.unique(ref))<2:continue
            for y in range(8,177-h,24):
                for x in range(8,177-w,24):
                    actual=screen[y:y+h,x:x+w]
                    if not np.array_equal(ref[mask],actual[mask]):continue
                    used=np.unique(ref[mask])
                    used_equal=bool(np.array_equal(colors[used],palette[used]))
                    hits.append(dict(key=key,rect=[x,y,w,h],hflip=flip,foreground_pixels=int(mask.sum()),
                        indexed_equal=True,used_rgb_equal=used_equal,
                        source_background_index=8 if family=='CPIC1.DAX'else None,
                        background_rgb_colors=np.unique(palette[actual[~mask]],axis=0).tolist()if (~mask).any()else [],
                        bank_equal=sourcepal[start*3:(start+count)*3]==rawpal[start*3:(start+count)*3]))
    actors=[r for r in hits if r['key'].startswith('cpic1-')]
    terrain=[r for r in hits if r['key'].startswith('dungcom-')]
    if len(actors)!=8 or len(terrain)!=5 or not all(r['used_rgb_equal']and r['bank_equal']for r in hits):
        raise AssertionError(f'已知正常樣本改變：actors={len(actors)} terrain={len(terrain)}')
    result=dict(all_pass=True,scope='正常冷啟動7400格DEMO；三個CPIC來源八個鏡像位置與兩個DUNGCOM來源五個完整圖塊；不推斷driver內部',
        inputs=dict(archives=archives,indexed_sha256=sha(raw),palette_sha256=sha(rawpal),log_sha256=sha((a.capture/'run.log').read_bytes())),
        palette_banks=banks,hits=hits)
    a.out.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    print('原始色號與用色完全吻合：CPIC 八個位置、DUNGCOM 五個位置；來源提供的色盤區段一致。')


if __name__=='__main__':main()
