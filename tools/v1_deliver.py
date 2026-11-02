#!/usr/bin/env python3
"""完整驗收後交付正式版十二包與影片，保護舊版及原版輸入。"""
import argparse,hashlib,json,os,re,shutil
from pathlib import Path

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())
def write(p,d):p.write_text(json.dumps(d,ensure_ascii=False,indent=2)+'\n')
def tree(p):return {str(q.relative_to(p)):sha(q) for q in sorted(p.rglob('*')) if q.is_file()}
def state(p):
    text=p.read_text();a=re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',text)
    b=re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)',text);assert a and b;return list(a.groups())+list(b.groups())

def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for n in ('root','work'):ap.add_argument('--'+n,type=Path,required=True)
    a=ap.parse_args();assert Path('/.dockerenv').exists();r,w=a.root,a.work
    package=read(w/'package-content.json');promo=read(w/'promo-runtime.json');version=package['version']
    assert version==promo['version']=='v.1.0.0-20261102' and package['status']=='PASS archive content only'
    assert promo['status'].startswith('PASS') and len(package['files'])==12
    cases={n:read(w/'smoke'/n/'pass.json') for n in ['linux-intro','linux-base','linux-full','linux-save','linux-missing','windows']}
    assert all(c['status']=='PASS' and c['version']==version for c in cases.values())
    byname={row['name']:row for row in package['files']}
    for case in cases.values():
        for run in case.get('runs',[]):assert run['app_sha256']==byname[run['app']]['sha256']
    win=next(x for x in package['files'] if x['platform']=='windows' and x['edition']=='hd' and x['kind']=='patch')
    assert cases['windows']['zip_sha256']==win['sha256']
    lock=read(w/'main/review/hd/runtime/manifest.json');theme={x['path']:x['sha256'] for x in lock['theme']}
    assert cases['linux-intro']['first']['cache']==cases['windows']['cache']==theme
    old=read(r/'docs/re/038-input-package.json')['runtime']['linux-save'];before={x['label']:x for x in old['runs']}
    comparisons={}
    for run in cases['linux-save']['runs']:
        if run['label'] not in ('save','load'):continue
        control=before[run['label']];assert run['end']==control['end'] and run['digest']==control['digest']
        comparisons[run['label']]=dict(end=run['end'],digest=run['digest'],equal_v013=True)
    assert cases['linux-save']['save_tree']==old['save_tree']
    old_packages=[]
    for name in ['026-package-runtime.json','033-hd-patch-runtime.json','036-package-promo-runtime.json','038-input-package.json']:
        m=read(r/'docs/re'/name)
        for row in m['files']:
            found=list((r/'dist-all'/m['version']).rglob(row['name']));assert len(found)==1
            p=found[0];assert sha(p)==row['sha256'] and p.stat().st_size==row['size']
            old_packages.append(dict(path=str(p.relative_to(r)),sha256=row['sha256'],size=row['size']))
    assert len(old_packages)==36
    originals=tree(r/'workplace/input/dos/pod');assert originals==read(r/'docs/re/035-final-runtime.json')['originals_unchanged']
    old_movie=read(r/'docs/re/037-promo-runtime.json')['video']
    assert sha(r/'dist-all/v.0.1.2-20261102/promo'/old_movie['path'])==old_movie['sha256']
    capture=read(w/'capture/capture-build.json');assert capture['version']==version and capture['engine_commit']==package['source']['engine_commit']
    oracle=read(r/'docs/re/031-intro-hd-runtime.json')
    expected=['5100',str(oracle['steps']),oracle['memory_sha256'],oracle['indexed_sha256'],oracle['cpu_sha256'],oracle['palette_sha256']]
    assert state(w/'capture/intro/run.log')==expected
    assert 'quiz autofill frame=3505 answer=YOU enter=false' in (w/'capture/player/run.log').read_text()
    film=w/'promo'/promo['video']['path'];assert sha(film)==promo['video']['sha256']
    dest=r/'dist-all'/version;assert not dest.exists() and dest.parent.stat().st_uid==os.getuid();dest.mkdir()
    for n in ('patch','full-local','promo','smoke'):(dest/n).mkdir()
    sums=[]
    for row in package['files']:
        src=w/'packages'/row['name'];assert sha(src)==row['sha256'];out=dest/row['kind']/row['name'];shutil.copy2(src,out)
        sums.append(dict(path=str(out.relative_to(dest)),size=out.stat().st_size,sha256=sha(out),rights='local-only original data' if row['kind']=='full-local' else 'private patch; original required'))
    for name in [film.name,'contact.png','ffprobe.json','audio-check.log','image-check.log','storyboard.json']:
        shutil.copy2(w/'promo'/name,dest/'promo'/name)
    shutil.copytree(w/'promo/review',dest/'promo/review')
    sums.append(dict(path='promo/'+film.name,size=film.stat().st_size,sha256=sha(film),rights='local private original images/audio; no public upload'))
    for name in cases:
        src=w/'smoke'/name;out=dest/'smoke'/name;out.mkdir()
        for p in src.iterdir():
            if p.is_file() and (p.suffix in ('.log','.png') or p.name=='pass.json'):shutil.copy2(p,out/p.name)
        for folder in ('window','shots','save-shots','load-shots'):
            if (src/folder).exists():shutil.copytree(src/folder,out/folder)
    package.update(status='PASS formal v1 package scope',date='2026-11-02',issue=80,delivery_root=str(dest.relative_to(r)),
        runtime=cases,assets=read(w/'assets/assets-receipt.json'),save_state_comparisons=comparisons,
        old_36_packages_unchanged=old_packages,originals_unchanged=originals,old_video_unchanged=old_movie,
        formal_manifest=True,renamed_packages=True,
        limits=['Windows Wine only; native Windows/macOS unverified','macOS universal architecture/dependencies only; unsigned','New music listening approval and native language review remain separate','User-selected v1 milestone; no full campaign or all-path parity claim'])
    promo.update(delivery_root=str((dest/'promo').relative_to(r)),source=package['source'],capture_build=capture,original_state_equals_031=True,
        player_capture=dict(binary_sha256=sha(w/'bin/linux/pod-play'),script_sha256=sha(w/'capture/player/script.txt'),state=state(w/'capture/player/run.log')))
    toolnames=['release_prepare.py','package_build.sh','package_payload.py','package_appimage.py','release_verify.py','release_smoke_linux.py','release_smoke_windows.py',
        'promo_capture_build.py','promo_capture.sh','promo_player_capture.sh','promo_audio_metrics.py','promo_assets.py','promo_render.sh','promo_verify.py','v1_deliver.py']
    package['tools']=promo['tools']={n:sha(r/'tools'/n) for n in toolnames}
    write(r/'docs/re/039-v1-package-runtime.json',package);write(r/'docs/re/039-v1-promo-runtime.json',promo)
    write(dest/'smoke/package-runtime.json',package);write(dest/'promo/promo-runtime.json',promo)
    write(w/'package-runtime.json',package);write(w/'promo-runtime.json',promo)
    write(dest/'SHA256SUMS.json',dict(version=version,source=package['source'],files=sums,rights='private patches; full builds and original audiovisual media local only'))
    (dest/'README.txt').write_text(f'黑暗之池 正式版 {version}\n\npatch/：三平台基本與HD patch，需自備原版。\nfull-local/：本機完整版，不可散布。\npromo/：63秒1080p正式版推廣影片與驗收，只供本機私人展示。\nsmoke/：實際封包驗收。\nSHA256SUMS.json：十二包及影片的SHA。\n\nF8五語、F2倍率、F12原版；查驗題自動填入後按Enter，F9關閉輔助。\nWindows以Wine抽驗；macOS未簽章，原生Windows與macOS真機未驗。\n日韓未經母語者審校，新編配樂未經試聽。影片音源有技術檢查，未另作人耳驗收。\n')
    assert len(sums)==13 and all(sha(dest/x['path'])==x['sha256'] for x in sums)
    for p in [dest,*dest.rglob('*')]:assert (p.stat().st_uid,p.stat().st_gid)==(os.getuid(),os.getgid())
    print('PASS formal',version,'12 packages +63s1080p promo;',sum(x['size'] for x in sums),'bytes; old36 preserved')
if __name__=='__main__':main()
