#!/usr/bin/env python3
"""沿Git既定HTML版面重排簡中開場／海報；美術底圖與原版來源辨識保持原樣。"""
import argparse
import csv
import hashlib
import json
from pathlib import Path
import shutil

import numpy as np
from PIL import Image

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def text(root,stem):return {row['key']:row['translation'] for row in csv.DictReader((root/'text'/(stem+'.zh-CN.tsv')).open(),delimiter='\t')}
def once(html,old,new):
    if html.count(old)!=1:raise ValueError('既定HTML模板已變更：'+old[:40])
    return html.replace(old,new)

def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('mode',choices=('prepare','finish'));ap.add_argument('--root',type=Path,required=True);ap.add_argument('--out',type=Path,required=True);ap.add_argument('--theme',type=Path);a=ap.parse_args();r=a.root;o=a.out
    if a.mode=='prepare':
        o.mkdir(parents=True,exist_ok=True)
        h=(r/'review/hd/b07-title/prototype.html').read_text()
        h=once(h,"const code=params.get('lang')==='en'?'en':'zh-TW'","const code='zh-CN'")
        h=once(h,'const t=texts[code]',"texts['zh-CN']="+json.dumps(text(r,'title'),ensure_ascii=False)+';const t=texts[code]')
        font="file:///project/font/source/rounded/GenSenRounded2TW-B.otf"
        h=h.replace('../../../workplace/fonts-src/rounded/GenSenRounded2TW-B.otf',font)
        (o/'title.html').write_text(h)
        h=(r/'review/hd/poster/prototype.html').read_text();t=text(r,'poster')
        values=dict(eyebrow=t['eyebrow'],title=t['title1']+('\n'+t['title2'] if t['title2'] else ''),righttitle=t['righttitle'],description=t['description1']+'\n'+t['description2'])
        h=once(h,'function render(){const l=language.value;',"text['zh-CN']="+json.dumps(values,ensure_ascii=False)+";function render(){const l='zh-CN';")
        h=h.replace('../../../workplace/fonts-src/rounded/GenSenRounded2TW-B.otf',font)
        (o/'poster.html').write_text(h)
        print('簡中開場與海報沿用原版面準備完成');return
    if a.theme is None:raise ValueError('finish需指定獨立候選theme')
    folder=a.theme/'localized/zh-CN';folder.mkdir()
    base=a.theme/'localized/zh-TW';m=json.loads((base/'theme.json').read_text())
    for item in m['items']:
        name=item['key'];src=base/(name+'.png') if name=='title-002' else o/(name+'.png')
        # 黑暗／之池的簡繁用字完全相同，封面沿用同一已核准成品。
        if name=='title-002':assert text(r,'cover')=={row['key']:row['translation'] for row in csv.DictReader((r/'text/cover.zh-TW.tsv').open(),delimiter='\t')}
        image=Image.open(src).convert('RGBA');assert image.size==(960,600) and image.getextrema()[3]==(255,255)
        shutil.copyfile(src,folder/(name+'.png'));shutil.copyfile(base/(name+'.idx'),folder/(name+'.idx'));item['png_sha256']=sha(folder/(name+'.png'))
    (folder/'theme.json').write_text(json.dumps(m,indent=2)+'\n')
    layers=[]
    for side in ('left','right'):
        for scale in (2,3):
            background=a.theme/'posters'/f'poster-{side}-{scale}x.png';full=o/f'poster-full-{side}-{scale}x.png';b=np.array(Image.open(background).convert('RGBA'));f=np.array(Image.open(full).convert('RGBA'))
            assert b.shape==f.shape==(200*scale,80*scale,4)
            changed=np.any(b!=f,axis=2);delta=f.copy();delta[~changed]=0;p=a.theme/'posters'/f'poster-text-{side}.zh-CN-{scale}x.png';Image.fromarray(delta).save(p)
            assert np.array_equal(np.array(Image.alpha_composite(Image.fromarray(b),Image.fromarray(delta))),f)
            # 文案只在既定底部，頂部美術不能因語系切換而改變。
            assert not changed[:120*scale].any()
            layers.append(dict(path=str(p.relative_to(a.theme)),sha256=sha(p),changed_pixels=int(changed.sum()),composition_exact=True))
    proof=dict(issue=70,title=[dict(path=str(p.relative_to(a.theme)),sha256=sha(p)) for p in sorted(folder.iterdir())],posters=layers,
        template_sha256={n:sha(r/'review/hd'/n/'prototype.html') for n in ('b07-title','poster')},font_sha256=sha(r/'font/source/rounded/GenSenRounded2TW-B.otf'),layout=json.loads((o/'layout.json').read_text()),scope='same approved art and layout; zh-CN type only; cover same characters')
    (o/'assets.json').write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n');print('三張開場及四個海報文字層完成')

if __name__=='__main__':main()
