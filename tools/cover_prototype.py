#!/usr/bin/env python3
"""產生封面原版／HD 審閱原型；文字由語系 TSV 渲染，尚非正式 TITLE 接入。"""
import argparse
import base64
import csv
import hashlib
import json
from pathlib import Path
import shutil


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root',type=Path,required=True)
    ap.add_argument('--out',type=Path,required=True)
    a=ap.parse_args();a.out.mkdir(parents=True,exist_ok=True)
    sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
    source=a.root/'workplace/hd/requests/b06'
    art=a.root/'workplace/hd/out/b06/title-002-plate-prototype.png'
    job=json.loads((a.root/'workplace/hd/out/b06/jobs/title-002-r1.json').read_text())
    if job.get('costume_approval')!='2026-10-24: 採用禮服版':raise ValueError('缺少使用者禮服核准')
    texts={};hashes={}
    for code in ('zh-TW','en'):
        p=a.root/'text'/f'cover.{code}.tsv'
        rows=list(csv.DictReader(p.open(),delimiter='\t'))
        values={r['key']:r['translation']for r in rows}
        if len(rows)!=3 or set(values)!={'cover.title1','cover.title2','cover.title3'} or not values['cover.title1'] or not values['cover.title2']:
            raise ValueError('封面標題 key 重複、缺漏或漏譯')
        texts[code]=[values[k]for k in ('cover.title1','cover.title2','cover.title3')if values[k]];hashes[code]=sha(p)
    data='data:image/png;base64,'+base64.b64encode(art.read_bytes()).decode()
    html='''<!doctype html><html><meta charset="utf-8"><title>黑暗之池封面 HD 審閱</title>
<style>
@font-face{font-family:Rounded;src:url('../../../workplace/fonts-src/rounded/GenSenRounded2TW-B.otf');font-weight:700}
*{box-sizing:border-box}body{margin:0;background:#000}.cover{position:relative;width:960px;height:600px;overflow:hidden}
img{display:block;width:960px;height:600px}.title{position:absolute;left:22px;top:108px;width:492px;font-family:Rounded,sans-serif;font-weight:700;line-height:1.16;text-align:center;transform:skewX(-7deg);font-size:174px;letter-spacing:6px}
.line{display:block;background:linear-gradient(#e7edef 0%,#859099 40%,#c5d0d3 51%,#7b8992 100%);background-clip:text;-webkit-background-clip:text;color:transparent;filter:drop-shadow(-2px 2px 0 #263039) drop-shadow(-3px 4px 0 #1c2127)}
.en .title{top:52px;left:20px;width:566px;font-size:106px;letter-spacing:-2px;line-height:1.38}.en .line:nth-child(2){font-size:72px;line-height:1.25}
</style><main class="cover" id="cover"><img id="game" width="960" height="600" src="ART" alt="HD 重畫美術預覽"><div class="title" id="title"></div></main>
<script>const texts=TEXTS;const code=new URLSearchParams(location.search).get('lang')==='en'?'en':'zh-TW';document.documentElement.lang=code;cover.classList.toggle('en',code==='en');for(const t of texts[code]){let e=document.createElement('span');e.className='line';e.textContent=t;title.append(e)}</script></html>'''
    (a.out/'prototype.html').write_text(html.replace('ART',data).replace('TEXTS',json.dumps(texts,ensure_ascii=False)))
    shutil.copyfile(source/'title-002-3x.png',a.out/'title-002-original.png')
    meta=dict(status='user-approved costume; HD art preview; not runtime',issue=25,source=json.loads((source/'manifest.json').read_text())['items'][0],
        generation=job,plate_sha256=sha(art),text_sha256=hashes,font_sha256=sha(a.root/'workplace/fonts-src/rounded/GenSenRounded2TW-B.otf'),
        costume_exception='使用者核准完整藍紫色禮服；原版標題文字改由目前語系另畫，不以全圖色相／結構判定冒稱原版一致。',private_only=True)
    (a.out/'manifest.json').write_text(json.dumps(meta,ensure_ascii=False,indent=2)+'\n')
    print('封面原版、兩語系標題原型與來源紀錄已產生')


if __name__=='__main__':main()
