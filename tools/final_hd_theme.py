#!/usr/bin/env python3
"""沿既有BMA動畫schema建立FINAL候選；原版索引只在本機重生。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import struct
import sys

sys.path.insert(0, str(Path(__file__).resolve().parent))
from intro_bma import decode, states


def sha(b): return hashlib.sha256(b).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for key in ('root','capture','artwork','out'): ap.add_argument('--'+key,type=Path,required=True)
    a=ap.parse_args(); r=a.root; original=r/'workplace/input/dos/pod'
    proof=json.loads((a.capture.parent/'source-proof.json').read_text())
    data=(original/'FINAL.BMA').read_bytes(); assert sha(data)==proof['input_sha256']['FINAL.BMA']
    assert sha((original/'FINAL.ACF').read_bytes())==proof['input_sha256']['FINAL.ACF']
    blocks=decode(data); art=json.loads((a.artwork/'manifest.json').read_text()); assert art['all_pass']
    items={x['key']:x for x in art['items']}; keys={(x['block'],x['frame']):x['key'] for x in proof['frames']}
    requests={x['key']:x for x in proof['requests']}; assert len(items)==43
    for k,row in items.items(): assert row['passed'] and sha((a.artwork/row['png']).read_bytes())==row['sha256']
    if a.out.exists(): raise ValueError('候選輸出已存在')
    shutil.copytree(r/'workplace/hd/theme',a.out)
    recipes=[]
    def emit(name,payload,recipe):
        p=a.out/name; p.parent.mkdir(parents=True,exist_ok=True)
        if p.exists(): assert p.read_bytes()==payload
        else: p.write_bytes(payload)
        if not any(x['path']==name for x in recipes): recipes.append(dict(path=name,sha256=sha(payload),**recipe))
        return dict(path=name,sha256=sha(payload))
    def index(raw,f,side):
        return emit('final/source/'+sha(raw)+'.idx',raw,dict(kind='original-intro',archive='FINAL.BMA',decoder='bma-index',block=f.block,frame=f.index,side=side))
    frames=[]
    for f,before,after in states(blocks):
        ref=index(after[:53760],f,'after'); prior=index(before[:53760],f,'before')
        points=[p for s in f.spans for p in range(s.y*320+s.x,s.y*320+s.x+len(s.pixels)) if p<53760]
        assert len(points)==len(set(points))
        ops=emit('final/source/b%02d-f%02d.ops'%(f.block,f.index),struct.pack('<'+'H'*len(points),*points),
            dict(kind='original-intro',archive='FINAL.BMA',decoder='bma-writes',block=f.block,frame=f.index))
        row=items[keys[(f.block,f.index)]]
        png=emit('final/art/'+row['png'],(a.artwork/row['png']).read_bytes(),
            dict(kind='repo-file',source='review/hd/b48-final/'+row['png'],rights=row['rights']))
        frames.append(dict(block=f.block,frame=f.index,index=ref,before=prior,writes=ops,png=png))
    palettes=[{} for _ in blocks]
    ref_pals={}
    for row in proof['requests']:
        if row['block'] not in ref_pals: ref_pals[row['block']]=(r/'workplace/hd/requests/b48-final'/row['palette']).read_bytes()
    for e in proof['events']:
        if e['kind']=='outside-BMA': continue
        b=e['block']; p=(a.capture/(e['source']+'.pal')).read_bytes(); idx=(a.capture/(e['source']+'.idx')).read_bytes()[:53760]
        used=set(idx); value=max(p[c*3+ch] for c in used for ch in range(3)); peak=max(ref_pals[b][c*3+ch] for c in used for ch in range(3))
        light=65535 if peak==0 else min(65535,round(value*65535/peak))
        if len(used)>1: palettes[b][sha(p)]=light
        elif sha(p) not in palettes[b]: palettes[b][sha(p)]=65535
    meta=dict(version=1,scale=3,blocks=[dict(sha256=b.sha256,size=b.size,count=len(b.frames),
        palettes=[dict(sha256=h,light=v) for h,v in sorted(palettes[b.index].items())]) for b in blocks],frames=frames)
    emit('final.json',(json.dumps(meta,ensure_ascii=False,indent=2)+'\n').encode(),dict(kind='repo-file',source='review/hd/b48-final/final.json',rights='original source fingerprints and HD presentation metadata'))
    (a.out.parent/'rebuild-recipes.json').write_text(json.dumps(recipes,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(dict(final_files=len(recipes),all_theme_files=sum(p.is_file() for p in a.out.rglob('*')),frames=len(frames),art=len(items)),ensure_ascii=False))


if __name__=='__main__': main()
