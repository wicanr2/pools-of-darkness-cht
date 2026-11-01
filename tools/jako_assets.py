#!/usr/bin/env python3
"""沿核准版面重排日韓開場與B版海報，素材寫入獨立候選。"""
import argparse
import csv
import hashlib
import json
from pathlib import Path
import shutil

import numpy as np
from PIL import Image


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def table(folder,stem,lang):return {x['key']:x['translation'] for x in csv.DictReader((folder/(stem+'.'+lang+'.tsv')).open(),delimiter='\t')}
def once(s,old,new):
    assert s.count(old)==1,old[:60]
    return s.replace(old,new)


def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('mode',choices=('prepare','finish'))
    for key in ('root','text','out'):ap.add_argument('--'+key,type=Path,required=True)
    ap.add_argument('--theme',type=Path);a=ap.parse_args();r=a.root;o=a.out
    if a.mode=='prepare':
        o.mkdir(parents=True,exist_ok=True)
        for lang in ('ja','ko'):
            dest=o/lang;dest.mkdir(exist_ok=True)
            region='JP' if lang=='ja' else 'TW';font=f'file:///project/font/source/rounded/GenSenRounded2{region}-B.otf'
            title=(r/'review/hd/b07-title/prototype.html').read_text()
            title=once(title,"const code=params.get('lang')==='en'?'en':'zh-TW'",f"const code='{lang}'")
            title=once(title,'const t=texts[code]',f'texts[{json.dumps(lang)}]='+json.dumps(table(a.text,'title',lang),ensure_ascii=False)+';const t=texts[code]')
            title=title.replace('../../../workplace/fonts-src/rounded/GenSenRounded2TW-B.otf',font)
            (dest/'title.html').write_text(title)
            cover=(r/'review/hd/b06-cover/prototype.html').read_text();t=table(a.text,'cover',lang)
            cover=once(cover,"const code=new URLSearchParams(location.search).get('lang')==='en'?'en':'zh-TW'",f"const code='{lang}'")
            cover=once(cover,"document.documentElement.lang=code",f'texts[{json.dumps(lang)}]='+json.dumps([t['cover.title'+str(i)] for i in (1,2,3)],ensure_ascii=False)+';document.documentElement.lang=code')
            cover=once(cover,"cover.classList.toggle('en',code==='en')","cover.classList.add('en')")
            cover=cover.replace('../../../workplace/fonts-src/rounded/GenSenRounded2TW-B.otf',font)
            (dest/'cover.html').write_text(cover)
            poster=(r/'review/hd/poster/prototype.html').read_text();t=table(a.text,'poster',lang)
            values=dict(eyebrow=t['eyebrow'],title=t['title1']+'\n'+t['title2'],righttitle=t['righttitle'],description=t['description1']+'\n'+t['description2'])
            poster=once(poster,'function render(){const l=language.value;',f'text[{json.dumps(lang)}]='+json.dumps(values,ensure_ascii=False)+f";function render(){{const l='{lang}';")
            poster=poster.replace('../../../workplace/fonts-src/rounded/GenSenRounded2TW-B.otf',font)
            # 只擷取透明文字層，兩倍背景不經瀏覽器重取樣，避免污染B版底圖。
            css='<style>body,.stage,.side{background:transparent!important}header,footer,.game,.side svg{visibility:hidden}.side:after{display:none}.left,.right{border:0}</style>'
            poster=once(poster,'</style>', '</style>'+css)
            (dest/'poster.html').write_text(poster)
        print('日韓三張開場與透明海報文字已準備');return
    assert a.theme is not None and a.theme.is_dir()
    proof={}
    for lang in ('ja','ko'):
        folder=a.theme/'localized'/lang;folder.mkdir(exist_ok=True)
        base=a.theme/'localized/zh-TW';meta=json.loads((base/'theme.json').read_text())
        for item in meta['items']:
            name=item['key'];src=o/lang/(name+'.png');im=Image.open(src).convert('RGBA')
            assert im.size==(960,600) and im.getextrema()[3]==(255,255)
            shutil.copyfile(src,folder/(name+'.png'));shutil.copyfile(base/(name+'.idx'),folder/(name+'.idx'))
            item['png_sha256']=sha(folder/(name+'.png'))
        (folder/'theme.json').write_text(json.dumps(meta,indent=2)+'\n')
        layers=[]
        for side in ('left','right'):
            for scale in (2,3):
                bg=Image.open(a.theme/'posters'/f'poster-{side}-{scale}x.png').convert('RGBA')
                text=Image.open(o/lang/f'poster-text-{side}-{scale}x.png').convert('RGBA')
                assert bg.size==text.size==(80*scale,200*scale)
                full=Image.alpha_composite(bg,text);b=np.array(bg);f=np.array(full);changed=np.any(b!=f,axis=2)
                assert changed.any() and not changed[:120*scale].any()
                delta=f.copy();delta[~changed]=0
                name=f'poster-text-{side}.{lang}-{scale}x.png';Image.fromarray(delta).save(a.theme/'posters'/name)
                full.save(o/lang/f'poster-full-{side}-{scale}x.png')
                assert np.array_equal(np.array(Image.alpha_composite(bg,Image.fromarray(delta))),f)
                layers.append(dict(path='posters/'+name,sha256=sha(a.theme/'posters'/name),changed_pixels=int(changed.sum()),composition_exact=True))
        proof[lang]=dict(title=[dict(path=str(p.relative_to(a.theme)),sha256=sha(p)) for p in sorted(folder.iterdir())],posters=layers,
                        layout=json.loads((o/lang/'layout.json').read_text()),machine_translation=True)
    (o/'assets.json').write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n');print('日韓開場與B版海報完成')


if __name__=='__main__':main()
