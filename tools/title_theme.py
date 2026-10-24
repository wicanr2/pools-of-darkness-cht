"""將逐件目視且文字邊界通過的TITLE圖匯出為既有語系主題清單。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
from PIL import Image


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root',type=Path,required=True)
    ap.add_argument('--theme',type=Path,required=True)
    a=ap.parse_args();r=a.root
    review=r/'review/hd/b07-title';cover=r/'review/hd/b06-cover'
    visual=json.loads((review/'visual-review.json').read_text())
    metadata=json.loads((review/'manifest.json').read_text())
    if sha(r/'workplace/hd/requests/b07/manifest.json')!=metadata['request_manifest_sha256']:raise ValueError('開場請求已變更')
    for code,expected in metadata['text_sha256'].items():
        if sha(r/'text'/f'title.{code}.tsv')!=expected:raise ValueError('正式開場文案已變更')
    if sha(r/'workplace/fonts-src/rounded/GenSenRounded2TW-B.otf')!=metadata['font_sha256']:raise ValueError('字型已變更')
    for row in metadata['items']:
        plate=r/'workplace/hd/out/b07'/(row['key']+'-plate.png')
        raw=r/'workplace/hd/out/b07'/(row['key']+'-r0-raw.png')
        if sha(plate)!=row['plate_sha256']or sha(raw)!=row['generation']['raw_sha256']:raise ValueError('已目視底圖或原始生成圖變更')
    approval=json.loads((cover/'manifest.json').read_text())
    if approval['generation'].get('costume_approval')!='2026-10-24: 採用禮服版':raise ValueError('缺封面服裝核准')
    for code,expected in approval['text_sha256'].items():
        if sha(r/'text'/f'cover.{code}.tsv')!=expected:raise ValueError('正式封面標題已變更')
    if sha(r/'workplace/hd/out/b06/title-002-plate-prototype.png')!=approval['plate_sha256']:raise ValueError('已核准禮服底圖變更')
    cover_visual=json.loads((cover/'capture.json').read_text())
    if len(cover_visual)!=2:raise ValueError('封面擷取不完整')
    sources={}
    for batch in ('b06','b07'):
        req=r/'workplace/hd/requests'/batch
        for row in json.loads((req/'manifest.json').read_text())['items']:
            if batch=='b06'and row!=approval['source']:raise ValueError('封面請求與核准來源不符')
            p=req/row['indexed']
            if row['source_file']!='TITLE.DAX'or row['source_size']!=[320,200]or sha(p)!=row['indexed_sha256']:
                raise ValueError('TITLE來源不符')
            sources[row['key']]=p
    selected={}
    for code in ('base','zh-TW','en'):
        items=[]
        for key in ('title-001','title-002','title-003'):
            if code=='base':
                image=r/('workplace/hd/out/b06/title-002-plate-prototype.png'if key=='title-002'else f'workplace/hd/out/b07/{key}-plate.png')
            elif key=='title-002':
                image=r/'workplace/hd/out/b06'/f'title-002.{code}.png'
                expected=approval['preview_outputs']['game-cover-hd-preview.png'if code=='zh-TW'else'game-cover-hd-preview-en.png']
                if sha(image)!=expected:raise ValueError('已核准封面變更')
            else:
                image=review/f'{key}.{code}.png'
                if sha(image)!=visual.get(image.name):raise ValueError('缺目前語系圖的目視雜湊')
                boxes=json.loads(image.with_suffix('.layout.json').read_text())
                if any(b['scrollWidth'] > b['width'] or b['x']+b['width'] > 960 or b['y']+b['height'] > 600 for b in boxes):
                    raise ValueError('文字邊界不合')
            im=Image.open(image).convert('RGBA')
            if im.size!=(960,600)or im.getextrema()[3]!=(255,255):raise ValueError('成品尺寸或透明度不合')
            selected[code,key]=(image,im)
            items.append(dict(key=key,screen_rect=[0,0,320,200],png=key+'.png',indexed=key+'.idx',indexed_sha256=sha(sources[key])))
        # 全部語系驗完後才寫入。
        selected[code,'manifest']=dict(scale=3,items=items)
    for code in ('base','zh-TW','en'):
        folder=a.theme/'localized'/code;folder.mkdir(parents=True,exist_ok=True)
        manifest=selected[code,'manifest']
        for item in manifest['items']:
            key=item['key'];image,im=selected[code,key]
            im.save(folder/item['png']);shutil.copyfile(sources[key],folder/item['indexed'])
            item['png_sha256']=sha(folder/item['png'])
        (folder/'theme.json').write_text(json.dumps(manifest,indent=2)+'\n')
    (a.theme/'localized/provenance.json').write_text(json.dumps(dict(issue=26,source='TITLE.DAX',
        preview_manifest_sha256=sha(review/'manifest.json'),visual_sha256=sha(review/'visual-review.json'),
        cover_manifest_sha256=sha(cover/'manifest.json'),text_exception=metadata['text_policy'],
        scope='asset and localized text accepted; normal runtime gate separate'),ensure_ascii=False,indent=2)+'\n')
    print('三張TITLE、三組語系主題已建置；正式接入仍需正常前端驗收。')


if __name__=='__main__':main()
