#!/usr/bin/env python3
"""VGA 逐件驗收，沿用結構≥0.40、色相中位<30度、尺寸／RGBA／遮罩外差異0。"""
import argparse
import hashlib
import json
from pathlib import Path
import sys

import numpy as np
from PIL import Image, ImageFilter
sys.path.insert(0,str(Path(__file__).resolve().parent))
from hd_metrics import read_png, structural_score, hue_score


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--req',type=Path,required=True)
    ap.add_argument('--out',type=Path,required=True)
    ap.add_argument('--keys',nargs='+')
    ap.add_argument('--report',help='另存驗收檔名，保留各輪收據')
    a=ap.parse_args()
    request=json.loads((a.req/'manifest.json').read_text())
    items=[i for i in request['items'] if not a.keys or i['key'] in a.keys]
    if a.keys and {i['key'] for i in items}!=set(a.keys):raise ValueError('缺指定來源')
    rows=[]
    for it in items:
        key=it['key'];p=a.out/(key+'.png')
        if not p.exists():
            rows.append(dict(key=key,target_met=False,error='缺輸出'));continue
        idx_path,pal_path=a.req/it['indexed'],a.req/it['palette']
        if sha(idx_path)!=it['indexed_sha256'] or sha(pal_path)!=it['palette_sha256']:
            raise ValueError(f'{key}: 原版基準雜湊改變')
        w,h=it['source_size'];W,H=it['output_size']
        idx=idx_path.read_bytes();pal=pal_path.read_bytes()
        if len(idx)!=w*h or len(pal)!=768:raise ValueError('原版色號或RGB色盤長度錯誤')
        orig=np.frombuffer(pal,np.uint8).reshape(256,3)[np.frombuffer(idx,np.uint8).reshape(h,w)]
        ct,hd=read_png(p);shape_ok=ct==6 and hd.shape==(H,W,4)
        alpha_ok=bool(np.all(hd[...,3]==255));score,hue,off=0.,360.,None
        if hd.shape[:2]==(H,W):
            score=structural_score(orig,hd);hue=hue_score(orig,hd)[0]
            anim=it.get('animation')
            if anim:
                mask_path=a.req/anim['mask']
                if sha(mask_path)!=anim['mask_sha256']:raise ValueError('遮罩雜湊改變')
                mask=np.asarray(Image.open(mask_path).convert('L').filter(ImageFilter.MaxFilter(13)))>0
                _,base=read_png(a.out/(anim['base']+'.png'))
                off=int(np.any(base!=hd,axis=2)[~mask].sum()) if base.shape==hd.shape else -1
        passed=shape_ok and alpha_ok and score>=.4 and hue<30 and off in (None,0)
        row=dict(key=key,format_ok=shape_ok,alpha_ok=alpha_ok,structural_edge_overlap=score,
                 median_hue_difference_degrees=hue,outside_mask_difference=off,target_met=passed,png_sha256=sha(p))
        rows.append(row);print(key,f'structure={score:.3f} hue={hue:.1f} off={off}',passed)
    result=dict(batch=request['batch'],request_manifest_sha256=sha(a.req/'manifest.json'),
                scope='selected' if a.keys else 'all',requested=len(items),total=len(request['items']),
                all_pass=bool(rows) and all(r['target_met'] for r in rows),items=rows)
    suffix='-selected' if a.keys else ''
    report=a.report or f'validation{suffix}.json'
    if Path(report).name!=report:raise ValueError('report 須為單一檔名')
    (a.out/report).write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    raise SystemExit(0 if result['all_pass'] else 1)


if __name__=='__main__':main()
