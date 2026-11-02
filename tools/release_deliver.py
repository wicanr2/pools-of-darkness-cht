#!/usr/bin/env python3
"""僅在封包與影片驗收通過後建立全新dist-all版本；不覆寫舊交付。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())
def write(p,data):p.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for key in ('root','work'):ap.add_argument('--'+key,type=Path,required=True)
    a=ap.parse_args();r,w=a.root,a.work;package=read(w/'package-runtime.json');promo=read(w/'promo-runtime.json')
    assert Path('/.dockerenv').exists() and package['status'].startswith('PASS') and promo['status'].startswith('PASS')
    version=package['version'];assert version==promo['version']=='v.0.1.2-20261102'
    dest=r/'dist-all'/version;assert not dest.exists() and dest.parent.stat().st_uid==os.getuid()
    tools=('release_prepare.py','release_smoke_linux.py','release_smoke_windows.py','release_verify.py','release_deliver.py',
        'promo_capture_build.py','promo_capture.sh','promo_assets.py','promo_render.sh','promo_verify.py',
        'package_payload.py','package_appimage.py','package_build.sh')
    toolhash={n:sha(r/'tools'/n) for n in tools};package['tools']=toolhash;promo['tools']=toolhash
    package['seed_source']=read(w/'seed-source.json')
    for row in package['files']:
        p=w/'packages'/row['name'];assert sha(p)==row['sha256'] and p.stat().st_size==row['size']
    film=w/'promo'/promo['video']['path'];assert sha(film)==promo['video']['sha256']
    dest.mkdir()
    for name in ('patch','full-local','promo','smoke'):(dest/name).mkdir()
    rows=[]
    for row in package['files']:
        target=dest/row['kind']/row['name'];shutil.copy2(w/'packages'/row['name'],target)
        assert sha(target)==row['sha256']
        rows.append(dict(path=str(target.relative_to(dest)),size=target.stat().st_size,sha256=row['sha256'],rights='local-only original data' if row['kind']=='full-local' else 'private patch, original game required'))
    for name in (promo['video']['path'],'contact.png','ffprobe.json','audio-check.log','image-check.log','storyboard.json'):
        shutil.copy2(w/'promo'/name,dest/'promo'/name)
    shutil.copytree(w/'promo/review',dest/'promo/review')
    write(dest/'promo/rights.json',dict(version=version,scope='local private promotional video, no public upload',
        assets='actual game captures and original AdLib audio through existing Nuked OPL',
        video_sha256=sha(dest/'promo'/promo['video']['path']),original_audio_sha256=promo['storyboard']['capture']['audio_sha256'],
        public_distribution_cleared=False,human_audio_listening=False,source_metadata='storyboard.json'))
    target=dest/'promo'/promo['video']['path'];rows.append(dict(path=str(target.relative_to(dest)),size=target.stat().st_size,sha256=sha(target),rights='local private video; original images/audio'))
    for case in ('linux-intro','linux-base','linux-full','linux-demo','linux-save','linux-missing','windows'):
        src=w/'smoke'/case;out=dest/'smoke'/case;out.mkdir()
        for q in src.iterdir():
            if q.is_file() and (q.suffix in ('.log','.png') or q.name=='pass.json'):shutil.copy2(q,out/q.name)
        for sub in ('window','shots','save-shots','load-shots'):
            if (src/sub).is_dir():shutil.copytree(src/sub,out/sub)
    baseline=dest/'smoke/baseline-old-now';baseline.mkdir()
    for q in (w/'smoke-baseline-old-now').iterdir():
        if q.is_file() and (q.suffix=='.log' or q.name=='pass.json'):shutil.copy2(q,baseline/q.name)
    package['delivery_root']='dist-all/'+version;promo['delivery_root']='dist-all/'+version+'/promo'
    write(dest/'smoke/package-runtime.json',package);write(dest/'promo/promo-runtime.json',promo)
    write(r/'docs/re/036-package-promo-runtime.json',package);write(r/'docs/re/037-promo-runtime.json',promo)
    write(dest/'SHA256SUMS.json',dict(version=version,source=package['source'],files=rows,
        rights='six private patches, six local-only full builds, one local-only promo; no public release'))
    (dest/'README.txt').write_text(f'''黑暗之池 {version}

patch/：基本與HD patch，各三平台，需自備DOS原版1.10的47個遊戲檔。
full-local/：含原版資料的本機基本與HD完整版，不可上傳或散布。
promo/：63秒1080p推廣影片、抽樣圖與音訊／影像驗收，只供本機私人展示。
smoke/：實際封包啟動、五語切換、存讀檔與平台驗收紀錄。
SHA256SUMS.json：十二包及推廣影片的大小、雜湊與權利分類。

HD包含五語、97種片頭及43種片尾HD相位。F2切倍率，F8切語言，F12切英文原版。
Windows以Wine抽驗；原生Windows及macOS真機未驗，macOS未簽章。
日韓為機器翻譯並經語境校讀，未經母語者審校。新編配樂試聽仍未完成。
影片採原版實際音源，完成技術音訊檢查，未另作人耳試聽。沒有公開上傳、tag或Release。
''')
    assert len(rows)==13 and all(sha(dest/x['path'])==x['sha256'] for x in rows)
    for p in [dest,*dest.rglob('*')]:assert (p.stat().st_uid,p.stat().st_gid)==(os.getuid(),os.getgid())
    print('PASS delivery',version,'12 packages +63s promo',sum(x['size'] for x in rows),'bytes')


if __name__=='__main__':main()
