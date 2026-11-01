#!/usr/bin/env python3
"""建立BMA片頭HD候選主題；不改現行主題。研究031與規格010。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import struct
import sys

sys.path.insert(0,str(Path(__file__).resolve().parent))
from intro_bma import decode, states


def sha(data):return hashlib.sha256(data).hexdigest()


def build(root, original, artwork, capture, out):
    if out.exists():raise ValueError('輸出必須是新目錄')
    data=(original/'INTRO.BMA').read_bytes()
    proof=json.loads((root/'docs/re/031-intro-bma-source.json').read_text())
    if sha(data)!=proof['input_sha256']['INTRO.BMA']:raise ValueError('BMA來源失配')
    if sha((original/'INTRO.ACF').read_bytes())!=proof['input_sha256']['INTRO.ACF']:raise ValueError('ACF來源失配')
    blocks=decode(data)
    art=json.loads((artwork/'manifest.json').read_text())
    items={r['source_indexed_sha256']:r for r in art['items']}
    if len(items)!=97:raise ValueError('敘事相位不齊')
    for r in items.values():
        if sha((artwork/r['png']).read_bytes())!=r['sha256']:raise ValueError('PNG已改變')
    requests=json.loads((root/'docs/re/031-intro-bma-requests.json').read_text())['items']
    # 只保存正常來源的色盤雜湊與明暗比例，不把原版色盤複製進Git。
    paletteByBlock=[{} for _ in blocks]
    fullpal={}
    for r in requests:
        if r['block'] not in fullpal:
            fullpal[r['block']]=(root/'workplace/probe/intro-hd-v1/requests-v1'/r['palette']).read_bytes()
    for e in proof['events']:
        if e['kind']=='outside-BMA' or not e['image_call']:continue
        call=proof['calls'][e['image_call']-1]
        b=next(i for i,v in enumerate(blocks) if v.sha256==call['SHA'])
        p=(capture/(e['source']+'.pal')).read_bytes()
        idx=(capture/(e['source']+'.idx')).read_bytes()[:53760]
        used=set(idx)
        maximum=max(p[c*3+ch] for c in used for ch in range(3))
        baseline=max(fullpal[b][c*3+ch] for c in used for ch in range(3))
        light=65535 if baseline==0 else min(65535,round(maximum*65535/baseline))
        # 純黑清屏仍由索引驗證；同色盤不可因當幀索引不同而改變亮度。
        h=sha(p)
        if len(used)>1:paletteByBlock[b][h]=light
        elif h not in paletteByBlock[b]:paletteByBlock[b][h]=65535
    out.mkdir(parents=True)
    recipes=[]
    def emit(name,payload,recipe):
        p=out/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(payload)
        if not any(r['path']==name for r in recipes):recipes.append(dict(path=name,sha256=sha(payload),**recipe))
        return dict(path=name,sha256=sha(payload))
    refs={};frames=[]
    def index(raw,f,side):
        h=sha(raw);name='intro/source/'+h+'.idx'
        return emit(name,raw,dict(kind='original-intro',decoder='bma-index',block=f.block,frame=f.index,side=side))
    for f,before,after in states(blocks):
        ref=index(after[:53760],f,'after');prev=index(before[:53760],f,'before')
        points=[]
        for span in f.spans:
            points.extend(p for p in range(span.y*320+span.x,span.y*320+span.x+len(span.pixels)) if p<53760)
        if len(points)!=len(set(points)):raise ValueError('本次來源含未驗重複寫入')
        ops=struct.pack('<'+'H'*len(points),*points)
        writes=emit('intro/source/b%02d-f%02d.ops'%(f.block,f.index),ops,
                    dict(kind='original-intro',decoder='bma-writes',block=f.block,frame=f.index))
        png=dict(path='',sha256='')
        if f.block:
            r=items[ref['sha256']]
            png=emit('intro/art/'+r['png'],(artwork/r['png']).read_bytes(),dict(kind='repo-file',source='review/hd/b46-intro/'+r['png'],rights=r['rights']))
        frames.append(dict(block=f.block,frame=f.index,index=ref,before=prev,writes=writes,png=png))
    meta=dict(version=1,scale=3,blocks=[dict(sha256=b.sha256,size=b.size,count=len(b.frames),
              palettes=[dict(sha256=h,light=v) for h,v in sorted(paletteByBlock[b.index].items())]) for b in blocks],frames=frames)
    encoded=(json.dumps(meta,ensure_ascii=False,indent=2)+'\n').encode()
    emit('intro.json',encoded,dict(kind='repo-file',source='review/hd/b46-intro/intro.json',rights='original source fingerprints and HD presentation metadata'))
    (out/'rebuild-recipes.json').write_text(json.dumps(recipes,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(dict(frames=len(frames),artwork=len(items),files=len(recipes),palette_counts=[len(v) for v in paletteByBlock]),indent=2))


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for name in ('root','original','artwork','capture','out'):ap.add_argument('--'+name,type=Path,required=True)
    a=ap.parse_args();build(a.root,a.original,a.artwork,a.capture,a.out)


if __name__=='__main__':main()
