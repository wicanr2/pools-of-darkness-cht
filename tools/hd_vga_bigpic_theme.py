#!/usr/bin/env python3
"""將已驗BIGPIC合併至既有theme.json；位置只證實244，其餘為完整來源候選。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil

from PIL import Image


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root',type=Path,required=True)
    ap.add_argument('--theme',type=Path,required=True)
    a=ap.parse_args();r=a.root;dst=a.theme
    assert dst.is_dir() and dst.stat().st_uid==1000
    proof=r/'docs/re/014-bigpic-source-proof.json'
    evidence=json.loads(proof.read_text());assert evidence['hit']['key']=='bigpic1-244' and evidence['hit']['used_rgb_equal'] and evidence['hit']['indexed_equal']
    sources={i['key']:i for i in evidence['sources']};assert len(sources)==7
    manifest=json.loads((dst/'theme.json').read_text());assert manifest['scale']==3
    old={i['key']for i in manifest['items']};assert not old&sources.keys(),'不得覆寫已接入BIGPIC'
    selected=[];review_hashes={}
    for batch,review_path in [('b01','review/hd/b01-manifest.json'),('b03','review/hd/b03-bigpic/manifest.json')]:
        folder=r/'workplace/hd/out'/batch
        report=folder/'validation.json';numeric=json.loads(report.read_text());assert numeric['all_pass']
        review=r/review_path;reviewed={i['key']:i for i in json.loads(review.read_text())['items']};review_hashes[batch]=sha(review)
        for item in numeric['items']:
            key=item['key'];png=folder/(key+'.png');idx=r/'workplace/hd/inventory'/(key+'.idx')
            assert key in sources and key in reviewed and item['target_met']
            assert sha(png)==item['png_sha256'] and sha(idx)==sources[key]['indexed_sha256']
            if batch=='b01':
                accepted=reviewed[key]['numeric_acceptance'];assert accepted['target_met'] and accepted['png_sha256']==sha(png)
            else:
                assert reviewed[key]['numeric']['png_sha256']==sha(png) and reviewed[key]['visual']['png_sha256']==sha(png)
            with Image.open(png)as im:
                assert im.mode=='RGBA' and im.size==(912,360) and im.getextrema()[3]==(255,255)
            row=dict(key=key,screen_rect=[8,8,304,120],png=key+'.png',png_sha256=sha(png),indexed=key+'.idx',indexed_sha256=sha(idx))
            selected.append((row,png,idx))
    assert {row['key']for row,_,_ in selected}==set(sources) and len(selected)==7
    for row,png,idx in selected:
        shutil.copyfile(png,dst/row['png']);shutil.copyfile(idx,dst/row['indexed']);manifest['items'].append(row)
    (dst/'theme.json').write_text(json.dumps(manifest,indent=2)+'\n')
    (dst/'bigpic-provenance.json').write_text(json.dumps(dict(issue=44,source_proof_sha256=sha(proof),review_manifests=review_hashes,
        items=[dict(key=row['key'],observed=row['key']=='bigpic1-244',used_palette_sha256=sources[row['key']]['used_palette_sha256'])for row,_,_ in selected],
        scope='seven conditional full-source items; only244 has normal-frame evidence; runtime validation separate'),indent=2)+'\n')
    print('seven BIGPIC items merged; source244 observed, remaining six conditional')


if __name__=='__main__':
    main()
