"""以無字HD底圖、正式語系表與原生字型渲染開場標誌及製作名單。"""
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
    texts={};text_hashes={}
    for code in ('zh-TW','en'):
        p=a.root/'text'/f'title.{code}.tsv'
        rows=list(csv.DictReader(p.open(),delimiter='\t'))
        values={r['key']:r['translation']for r in rows}
        if len(rows)!=len(values)or len(rows)!=18 or any(not v for k,v in values.items()if k!='title.article'):
            raise ValueError('開場語系表缺列、重複或漏譯')
        texts[code]=values;text_hashes[code]=sha(p)
    for key in texts['en']:
        if key.endswith('.names')and texts['en'][key]!=texts['zh-TW'][key]:raise ValueError('原版製作人名不可因語系改變')
    plates={};provenance=[]
    request_path=a.root/'workplace/hd/requests/b07/manifest.json'
    request={row['key']:row for row in json.loads(request_path.read_text())['items']}
    for key in ('title-001','title-003'):
        p=a.root/'workplace/hd/out/b07'/(key+'-plate.png')
        job=json.loads((a.root/'workplace/hd/out/b07/jobs'/(key+'-r0.json')).read_text())
        raw=a.root/'workplace/hd/out/b07'/(key+'-r0-raw.png')
        if sha(raw)!=job['raw_sha256']:raise ValueError('原始生成圖已變更')
        plates[key]='data:image/png;base64,'+base64.b64encode(p.read_bytes()).decode()
        provenance.append(dict(key=key,source=request[key],output_size=[960,600],plate_sha256=sha(p),generation=job,
            rights='private original-reference review; AI redraw eligible for existing HD patch scope; localized type uses OFL font'))
        shutil.copyfile(a.root/'workplace/hd/requests/b07'/(key+'-3x.png'),a.out/(key+'-original.png'))
    html='''<!doctype html><html><meta charset="utf-8"><title>開場HD語系審閱</title><style>
@font-face{font-family:Rounded;src:url('../../../workplace/fonts-src/rounded/GenSenRounded2TW-B.otf');font-weight:700}
*{box-sizing:border-box}body{margin:0;background:#000}.screen{position:relative;width:960px;height:600px;overflow:hidden;font-family:Rounded,sans-serif;font-weight:700}#game{display:block;width:960px;height:600px}.text{position:absolute;white-space:nowrap;line-height:1.1;text-shadow:2px 2px #111;z-index:1}.label{color:#e4d743}.names{color:#e4e6cb;font-size:22px;line-height:24px}.row{display:block}.brand{left:82px;top:252px;width:796px;text-align:center;font-size:76px;letter-spacing:2px;line-height:1.1;background:linear-gradient(#fff3ad,#d59112 50%,#ffe7a0 53%,#ad5d05);background-clip:text;-webkit-background-clip:text;color:transparent;filter:drop-shadow(2px 3px 0 #120c04)}.article{left:0;top:128px;width:960px;text-align:center;color:#d7e7de;font-size:24px}.subtitle,.volume{left:0;width:960px;text-align:center;color:#d5d9d7;font-size:32px;top:390px}.volume{top:432px;font-size:28px}.en .brand{font-size:63px;letter-spacing:1px;top:260px}.creation{left:87px;top:16px;width:774px;text-align:center;font-size:25px}.creation .value{color:#e4e6cb;margin-left:12px}.heading{font-size:24px}.en .heading{font-size:23px}
.brand{text-shadow:none}
</style><main class="screen" id="screen"><img id="game" width="960" height="600" alt="HD無字底圖"></main><script>
const texts=TEXTS,plates=PLATES;const stage=document.getElementById('screen'),game=document.getElementById('game');const params=new URLSearchParams(location.search);const code=params.get('lang')==='en'?'en':'zh-TW';const key=params.get('layout')==='title-003'?'title-003':'title-001';const t=texts[code];document.documentElement.lang=code;stage.classList.toggle('en',code==='en');game.src=plates[key];
function add(k,cls,x,y,w){let e=document.createElement('div');e.className='text '+cls;if(x!==undefined)e.style.left=x+'px';if(y!==undefined)e.style.top=y+'px';if(w)e.style.width=w+'px';e.dataset.key=k;e.textContent=t[k];stage.append(e);return e}
if(key==='title-001'){add('title.article','article');add('title.brand','brand');add('title.subtitle','subtitle');add('title.volume','volume')}
else{let top=add('title.creation.label','creation label');let v=document.createElement('span');v.className='value';v.textContent=t['title.creation.value'];top.append(v);for(const [part,x,y,ny]of[['developer',111,66,90],['programmers',111,162,186],['music',111,282,306],['testers',111,378,402],['artists',543,66,90],['encounters',543,378,402]]){add('title.'+part+'.label','heading label',x,y,306);let e=add('title.'+part+'.names','names',x+24,ny,282);e.textContent='';for(const line of t['title.'+part+'.names'].split('|')){let s=document.createElement('span');s.className='row';s.textContent=line;e.append(s)}}}
window.titleMetrics=()=>[...document.querySelectorAll('.text')].map(e=>({key:e.dataset.key,x:e.offsetLeft,y:e.offsetTop,width:e.offsetWidth,height:e.offsetHeight,scrollWidth:e.scrollWidth}));
</script></html>'''
    (a.out/'prototype.html').write_text(html.replace('TEXTS',json.dumps(texts,ensure_ascii=False)).replace('PLATES',json.dumps(plates)))
    visual_path=a.out/'visual-review.json'
    visual=json.loads(visual_path.read_text())if visual_path.exists()else{}
    for name,digest in visual.items():
        if sha(a.out/name)!=digest:raise ValueError('目前預覽與已目視版本不符')
    (a.out/'manifest.json').write_text(json.dumps(dict(issue=26,status='art and localized type preview; runtime separate',items=provenance,
        request_manifest_sha256=sha(request_path),visual_outputs=visual,
        text_sha256=text_hashes,font_sha256=sha(a.root/'workplace/fonts-src/rounded/GenSenRounded2TW-B.otf'),
        text_policy='原版圖上的文字另由正式TSV渲染，數值驗收不將不同語系字形視為同一圖案',private_only=True),ensure_ascii=False,indent=2)+'\n')
    print('開場標誌及製作名單兩語系原型已產生。')


if __name__=='__main__':main()
