#!/usr/bin/env python3
"""產生可撤回的左右海報與語系示意；只讀既有截圖，不接入遊戲。"""
import argparse
import base64
import csv
import hashlib
import json
from pathlib import Path


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root',type=Path,required=True)
    ap.add_argument('--out',type=Path,required=True)
    a=ap.parse_args();a.out.mkdir(parents=True,exist_ok=True)
    inputs={
        'art':a.root/'workplace/hd/source/poster/steam-background.jpg',
        'zh':a.root/'docs/screenshots/demo-welcome-cht-hd.png',
        'en':a.root/'workplace/probe/g9-hd/frontend-switch/switched/shots/pod-f002102.png'}
    data={k:'data:image/'+('jpeg'if k=='art'else'png')+';base64,'+base64.b64encode(p.read_bytes()).decode()for k,p in inputs.items()}
    texts={}
    for lang in ['zh-TW','en']:
        path=a.root/'text'/f'poster.{lang}.tsv'
        rows=list(csv.DictReader(path.open(),delimiter='\t'))
        table={r['key']:r['translation']for r in rows}
        required={'eyebrow','title1','title2','righttitle','description1','description2'}
        if len(rows)!=len(table)or set(table)!=required:raise ValueError('海報文字 key 重複或缺漏')
        if any(not v for k,v in table.items()if k!='title2'):raise ValueError('海報文字漏譯')
        texts[lang]=dict(eyebrow=table['eyebrow'],title=table['title1']+('\n'+table['title2']if table['title2']else''),
            righttitle=table['righttitle'],description=table['description1']+'\n'+table['description2'])
    template='''<!doctype html><html lang="zh-Hant"><meta charset="utf-8"><title>黑暗之池左右海報示意</title>
<style>
@font-face{font-family:Rounded;src:url('../../../workplace/fonts-src/rounded/GenSenRounded2TW-B.otf');font-weight:700}
*{box-sizing:border-box}body{margin:0;background:#151311;color:#ded8cc;font-family:Rounded,sans-serif;font-weight:700}
header{height:56px;display:flex;align-items:center;gap:18px;padding:0 20px;font-size:14px}header b{color:#e2c58e;margin-right:auto}
select{background:#24211d;border:1px solid #62543c;border-radius:4px;color:#e6d9be;padding:5px 8px;font:inherit}
.stage{width:1440px;height:600px;display:flex;margin:auto;background:#000}.side{width:240px;height:600px;position:relative;overflow:hidden;flex-shrink:0}
.side svg{position:absolute;width:100%;height:100%;inset:0}.side:after{content:'';position:absolute;inset:0;background:linear-gradient(0deg,rgba(8,6,5,.97),transparent 53%),linear-gradient(90deg,rgba(0,0,0,.15),transparent 35%,rgba(0,0,0,.15))}
.left{border-right:1px solid #74603c}.right{border-left:1px solid #74603c}.copy{position:absolute;z-index:1;left:20px;right:20px;bottom:26px;color:#dfc396;text-shadow:0 2px 6px #000}
.eyebrow{font-size:13px;margin-bottom:9px;color:#c5b596;letter-spacing:2px}.title{font-size:36px;line-height:1.25;white-space:pre-line;letter-spacing:1px}.en .title{font-size:30px}
.righttitle{font-size:23px;line-height:1.4;margin-bottom:12px}.description{font-size:14px;line-height:1.65;color:#d3c9b6;white-space:pre-line}
.game{display:block;width:960px;height:600px;flex-shrink:0}footer{text-align:center;height:28px;font-size:12px;line-height:28px;color:#999184}
.narrow.stage{width:1066px}.narrow .side{width:53px}.narrow .copy{display:none}
</style>
<header><b>黑暗之池 · 左右海報示意</b><label>版面 <select id="layout"><option value="wide">左右各 240，中央不縮放</option><option value="narrow">只使用 16:9 餘白</option></select></label><label>語系 <select id="language"><option value="zh-TW">繁體中文</option><option value="en">English</option></select></label></header>
<main class="stage" id="stage"><aside class="side left"><svg viewBox="520 0 420 810" preserveAspectRatio="xMidYMid slice"><image width="1438" height="810" href="ART"/></svg><div class="copy"><div class="eyebrow" id="eyebrow"></div><div class="title" id="title"></div></div></aside><img class="game" id="game" alt="本作實際執行畫面"><aside class="side right"><svg viewBox="900 0 420 810" preserveAspectRatio="xMidYMid slice"><image width="1438" height="810" href="ART"/></svg><div class="copy"><div class="righttitle" id="righttitle"></div><div class="description" id="description"></div></div></aside></main>
<footer>中央為本作實際執行截圖；海報為可撤回的版面示意，尚未接入遊戲。</footer>
<script>
const images=IMAGES;
const text=TRANSLATIONS;
const params=new URLSearchParams(location.search);language.value=params.get('lang')==='en'?'en':'zh-TW';layout.value=params.get('layout')==='narrow'?'narrow':'wide';
function render(){const l=language.value;stage.className='stage '+(layout.value==='narrow'?'narrow ':'')+(l==='en'?'en':'');for(const key of ['eyebrow','title','righttitle','description'])document.getElementById(key).textContent=text[l][key];game.src=images[l==='en'?'en':'zh'];document.documentElement.lang=l==='en'?'en':'zh-Hant'}
language.onchange=layout.onchange=render;render();
</script></html>'''
    html=template.replace('href="ART"','href="'+data['art']+'"').replace('IMAGES',json.dumps({'zh':data['zh'],'en':data['en']})).replace('TRANSLATIONS',json.dumps(texts,ensure_ascii=False))
    (a.out/'prototype.html').write_text(html)
    meta=dict(status='DRAFT disposable prototype',issue=21,canvas=dict(game=[960,600],wide=[1440,600],narrow=[1066,600]),
        source_page='https://store.steampowered.com/app/1904522/Pools_of_Darkness/',
        inputs={k:dict(path=str(p.relative_to(a.root)),sha256=hashlib.sha256(p.read_bytes()).hexdigest())for k,p in inputs.items()},
        text='separate DOM text from text/poster.<locale>.tsv; switches with locale',private_only=True,
        text_sha256={lang:hashlib.sha256((a.root/'text'/f'poster.{lang}.tsv').read_bytes()).hexdigest()for lang in texts})
    previous=a.out/'manifest.json'
    if previous.exists():
        old=json.loads(previous.read_text())
        if old.get('approved_screenshots'):
            if old['inputs']['art']['sha256']!=meta['inputs']['art']['sha256']:raise ValueError('已定版海報來源不得自行替換')
            for name,expected in old['approved_screenshots'].items():
                if hashlib.sha256((a.out/name).read_bytes()).hexdigest()!=expected:raise ValueError('定版示意圖雜湊變更')
            for key in ['status','approval','approved_screenshots']:meta[key]=old[key]
    (a.out/'manifest.json').write_text(json.dumps(meta,ensure_ascii=False,indent=2)+'\n')
    print('HTML 與來源雜湊已產生')


if __name__=='__main__':main()
