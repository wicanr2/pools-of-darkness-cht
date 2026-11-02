#!/usr/bin/env python3
"""驗證輸入修正版的實際封包收據，再交付全新版本。"""
import argparse,hashlib,json,os,shutil
from pathlib import Path
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())
def write(p,d):p.write_text(json.dumps(d,ensure_ascii=False,indent=2)+'\n')
def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for n in ('root','work','input'):ap.add_argument('--'+n,type=Path,required=True)
    a=ap.parse_args();assert Path('/.dockerenv').exists();r,w=a.root,a.work
    contents=read(w/'package-content.json');version=contents['version'];assert version=='v.0.1.3-20261102'
    assert contents['status']=='PASS archive content only' and len(contents['files'])==12
    cases={n:read(w/'smoke'/n/'pass.json') for n in ['linux-intro','linux-base','linux-full','linux-save','linux-missing','windows']}
    assert all(c['status']=='PASS' and c['version']==version for c in cases.values())
    byname={row['name']:row for row in contents['files']}
    for c in cases.values():
        for run in c.get('runs',[]):assert run['app_sha256']==byname[run['app']]['sha256']
    win=next(x for x in contents['files'] if x['platform']=='windows' and x['edition']=='hd' and x['kind']=='patch')
    assert cases['windows']['zip_sha256']==win['sha256']
    lock=read(w/'main/review/hd/runtime/manifest.json');theme={x['path']:x['sha256'] for x in lock['theme']}
    assert cases['linux-intro']['first']['cache']==cases['windows']['cache']==theme
    inputproof=read(r/'docs/re/038-input-runtime.json');assert inputproof['status']=='PASS input sampled scope'
    assert inputproof['engine_commit']==contents['source']['engine_commit']
    # 同一合法原版125檔及自建A槽的新版自動輸入，與已驗舊版手動輸入比較。
    old=read(r/'docs/re/036-package-promo-runtime.json')['runtime']['linux-save']
    before={x['label']:x for x in old['runs']};comparisons={}
    for run in cases['linux-save']['runs']:
        if run['label'] not in ('save','load'):continue
        control=before[run['label']]
        assert run['end']==control['end'] and run['digest']==control['digest'],(run['label'],run,control)
        comparisons[run['label']]=dict(end=run['end'],digest=run['digest'],equal_v012_manual=True)
    assert cases['linux-save']['save_tree']==old['save_tree']
    for row in inputproof['old_packages']:assert sha(r/row['path'])==row['sha256']
    for row in contents['files']:assert sha(w/'packages'/row['name'])==row['sha256']
    dest=r/'dist-all'/version;assert not dest.exists() and dest.parent.stat().st_uid==os.getuid();dest.mkdir()
    for n in ('patch','full-local','smoke'):(dest/n).mkdir()
    sums=[]
    for row in contents['files']:
        p=dest/row['kind']/row['name'];shutil.copy2(w/'packages'/row['name'],p)
        sums.append(dict(path=str(p.relative_to(dest)),size=p.stat().st_size,sha256=sha(p),rights='local-only original data' if row['kind']=='full-local' else 'private patch; original required'))
    for n in cases:
        src=w/'smoke'/n;out=dest/'smoke'/n;out.mkdir()
        for p in src.iterdir():
            if p.is_file() and (p.suffix in ('.log','.png') or p.name=='pass.json'):shutil.copy2(p,out/p.name)
        for folder in ('window','shots','save-shots','load-shots'):
            if (src/folder).exists():shutil.copytree(src/folder,out/folder)
    for n in ['auto-fixed','manual-fixed','retry-fixed','mouse-2','mouse-3']:
        out=dest/'smoke'/n;out.mkdir();src=a.input/n
        for p in src.iterdir():
            if p.is_file() and (p.suffix=='.log' or p.name=='pass.json'):shutil.copy2(p,out/p.name)
        if (src/'shots').exists():shutil.copytree(src/'shots',out/'shots')
    write(dest/'SHA256SUMS.json',dict(version=version,files=sums))
    result=dict(status='PASS input revision packages',date='2026-11-02',issue=79,version=version,
        delivery_root=str(dest.relative_to(r)),source=contents['source'],files=contents['files'],runtime=cases,
        input_proof_sha256=sha(r/'docs/re/038-input-runtime.json'),save_load_same_state= comparisons,
        saves_equal_v012_manual=True,old_24_packages_unchanged=True,theme_files=2034,fonts=18,
        promo='dist-all/v.0.1.2-20261102/promo; existing video unchanged, not relabelled as this version',
        limits=['Windows Wine only; native Windows/macOS unverified','macOS universal structure and linkage verified; unsigned','No full campaign required; no whole-game completion claim'],
        tools={p.name:sha(p) for p in [r/'tools'/n for n in ['input_release.py','release_prepare.py','package_build.sh','package_payload.py','package_appimage.py','release_smoke_linux.py','release_smoke_windows.py','release_verify.py']]})
    write(w/'package-runtime.json',result);write(r/'docs/re/038-input-package.json',result);print('PASS',version,'12 packages; old24 unchanged; saves/state equal v012 manual')
if __name__=='__main__':main()
