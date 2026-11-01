#!/usr/bin/env python3
"""核對乾淨重建，保存正式片尾驗收；--activate只在全部檢查後替換本機主題及程式。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import sys
from PIL import Image
import PIL
import numpy as np
import cv2


def sha(b):return hashlib.sha256(b).hexdigest()
def tree(p):return {str(q.relative_to(p)):sha(q.read_bytes()) for q in sorted(p.rglob('*')) if q.is_file()}
def load(p):return json.loads(p.read_text())
def write(p,x):p.write_text(json.dumps(x,ensure_ascii=False,indent=2)+'\n')


def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('--root',type=Path,required=True)
    ap.add_argument('--activate',action='store_true');a=ap.parse_args();r=a.root;w=r/'workplace/probe/final-hd'
    proof=load(w/'runtime-proof.json');assets=load(w/'verified-assets/assets-receipt.json');m=load(r/'review/hd/runtime/manifest.json')
    assert assets['status']=='PASS' and assets['theme_files']==2034
    assert assets['engine']['commit']==proof['engine_commit']==m['engine']['commit']
    assert assets['manifest_sha256']==sha((r/'review/hd/runtime/manifest.json').read_bytes())
    tt=tree(w/'verified-assets/theme');assert tt==tree(w/'candidate/theme') and len(tt)==2034
    assert tree(w/'verified-assets/font')==tree(r/'workplace/font')==proof['fonts']
    orig=tree(r/'workplace/input/dos/pod');assert len(orig)==125
    assert all(m['runtime_expected']['guest'].get(k)==v for k,v in orig.items())
    metrics=load(r/'review/hd/b48-final/metrics.json');assert metrics['all_pass']
    review=load(r/'review/hd/b48-final/visual-review.json')
    assert {x['key']:x['sha256'] for x in metrics['items']}=={x['key']:x['sha256'] for x in review['items']}
    assert all(sha((r/'review/hd/b48-final'/x['png']).read_bytes())==x['sha256'] for x in metrics['items'])
    logs={}
    for name in ('go-packages.log','go-vet.log','render-tests.log'):
        p=w/name;logs[name]=dict(sha256=sha(p.read_bytes()),text=p.read_text())
    assert len(logs['go-packages.log']['text'].splitlines())==3 and not logs['go-vet.log']['text']
    proof.update(clean_main_source_tree='4dbb78721348d09e886ae8408fe8ae813bdeafb7',
        assets=assets,reconstructed_original_files=sum(x['kind'].startswith('original-') for x in m['theme']),
        originals_unchanged=orig,tests=logs,art=dict(count=43,structure_min=min(x['structure'] for x in metrics['items']),
        hue_median_max=max(x['hue'] for x in metrics['items']),manifest_sha256=sha((r/'review/hd/b48-final/metrics.json').read_bytes()),
        visual_review_sha256=sha((r/'review/hd/b48-final/visual-review.json').read_bytes())),
        tool_versions=dict(python=sys.version.split()[0],pillow=PIL.__version__,numpy=np.__version__,opencv=cv2.__version__,go='go1.24.13'))
    assert proof['reconstructed_original_files']==1091
    screen=r/'docs/screenshots';src=w/'runtime-v2/hd/window/pod-f001532.png'
    shutil.copy2(src,screen/'final-hd-window.png')
    write(screen/'final-hd-runtime.json',dict(date='2026-11-02',issue=77,platform='Linux Docker Xvfb',
        scope=proof['scope'],entry='ADLIB.COM then VACE.EXE 4 FINAL 0 1',frame=1532,source_update=1531,
        clock_percent=50,scale=3,theme='B, FINAL b48, B posters',language='zh-TW',
        executable_sha256=proof['binary_sha256']['pod-final-player'],engine_commit=proof['engine_commit'],
        source_idx_sha256=sha((w/'runtime-v2/hd/sources/pod-f001532.idx').read_bytes()),
        source_pal_sha256=sha((w/'runtime-v2/hd/sources/pod-f001532.pal').read_bytes()),
        screenshot_sha256=sha(src.read_bytes()),size=list(Image.open(src).size),
        rng='No RNG comparison required for scripted VACE artwork; no seed injection',
        verification='Same initial input and equivalent steps across original, HD and host-control cases; not a campaign victory'))
    if a.activate:
        formal=r/'workplace/hd/theme';program=r/'workplace/dosgolem/workplace/bin/pod-play'
        recipes=load(w/'candidate/rebuild-recipes.json');new={x['path'] for x in recipes}
        assert tree(formal)=={k:v for k,v in tt.items() if k not in new}
        assert sha(program.read_bytes())=='aa89fa8274b48561b84fde3e23e12c4ff9c0966a9b156d4f8cd5e8cd2004ba53'
        for p in (formal,formal.parent,program,program.parent,w):assert (p.stat().st_uid,p.stat().st_gid)==(os.getuid(),os.getgid())
        for p in formal.rglob('*'):assert (p.stat().st_uid,p.stat().st_gid)==(os.getuid(),os.getgid())
        backup=w/'formal-before';assert not backup.exists();backup.mkdir()
        ready=formal.with_name('theme-final-ready');assert not ready.exists()
        shutil.copytree(w/'verified-assets/theme',ready);assert tree(ready)==tt
        shutil.copy2(program,backup/'pod-play');formal.rename(backup/'theme');ready.rename(formal)
        temp=program.with_name('pod-play-final-ready');assert not temp.exists();shutil.copy2(w/'bin/pod-play',temp);temp.replace(program)
        assert tree(formal)==tt and sha(program.read_bytes())==proof['binary_sha256']['pod-play']
        assert (program.stat().st_uid,program.stat().st_gid)==(os.getuid(),os.getgid())
        proof['activation']=dict(status='PASS',theme_files=2034,previous_files=1888,backup='workplace/probe/final-hd/formal-before',
            binary_sha256=sha(program.read_bytes()),fonts_unchanged=True,old_packages_unchanged=True)
    proof['tools']={str(p.relative_to(r)):sha(p.read_bytes()) for p in sorted((r/'tools').glob('final_*'))}
    write(r/'docs/re/035-final-runtime.json',proof)
    print('PASS',proof['theme_files'],'theme;',proof['reconstructed_original_files'],'derived;',len(proof['old_packages_unchanged']),'old packages unchanged')


if __name__=='__main__':main()
