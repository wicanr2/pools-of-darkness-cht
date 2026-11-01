#!/usr/bin/env python3
"""逐相位驗證片頭HD；門檻沿用規格010，目視另綁定成品雜湊。"""
import argparse
import hashlib
import json
from pathlib import Path
import sys
import numpy as np
sys.path.insert(0,str(Path(__file__).resolve().parent))
from hd_metrics import read_png, structural_score, hue_score


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root',type=Path,required=True)
    ap.add_argument('--artwork',type=Path,required=True)
    ap.add_argument('--report',type=Path,required=True)
    a=ap.parse_args();root=a.root;req=root/'workplace/probe/intro-hd-v1/requests-v1'
    requests=json.loads((root/'docs/re/031-intro-bma-requests.json').read_text())['items']
    manifest=json.loads((a.artwork/'manifest.json').read_text());art={r['key']:r for r in manifest['items']}
    rows=[]
    canonical={(1,13):'council',(2,0):'phlan',(4,0):'horned',(6,0):'heroes',(7,0):'followers',(9,0):'dreadlord'}
    for r in requests:
        if r['block']==0:continue
        p=a.artwork/(r['key']+'.png')
        if sha(p)!=art[r['key']]['sha256'] or sha(req/r['indexed'])!=r['art_indexed_sha256'] or sha(req/r['palette'])!=r['palette_sha256']:raise ValueError('驗收輸入已變動')
        idx=np.frombuffer((req/r['indexed']).read_bytes(),np.uint8).reshape(168,320)
        pal=np.frombuffer((req/r['palette']).read_bytes(),np.uint8).reshape(256,3)
        ct,im=read_png(p);sc=structural_score(pal[idx],im);hue=hue_score(pal[idx],im)[0]
        anchor=canonical.get((r['block'],r['frame']))
        same=not anchor or p.read_bytes()==(root/f'review/hd/b44-intro/intro-{anchor}.png').read_bytes()
        # Pillow編碼可不同，核准的六張主畫面必須RGBA逐點相同。
        if anchor:same=bool(np.array_equal(im,read_png(root/f'review/hd/b44-intro/intro-{anchor}.png')[1]))
        rows.append(dict(key=r['key'],png_sha256=sha(p),structure=sc,hue=hue,
                         canonical_unchanged=same,passed=bool(ct==6 and im.shape==(504,960,4) and np.all(im[:,:,3]==255) and sc>=.4 and hue<30 and same)))
    result=dict(issue=74,count=len(rows),all_pass=len(rows)==97 and all(r['passed'] for r in rows),
                thresholds=dict(structure_min=.4,hue_median_max=30),items=rows,
                limits='數值檢查不取代逐相位目視、原尺寸抽樣及正常執行')
    a.report.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    print('通過',sum(r['passed'] for r in rows),'/',len(rows));print('未通過',[r for r in rows if not r['passed']])
    if not result['all_pass']:raise SystemExit(1)


if __name__=='__main__':main()
