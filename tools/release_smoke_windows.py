#!/usr/bin/env python3
"""實際新版HD patch ZIP的Wine日韓啟動與重啟，保留原生平台限制。"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import zipfile

VERSION='v.0.1.2-20261102'
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def tree(p):return {str(q.relative_to(p)):sha(q) for q in sorted(p.rglob('*')) if q.is_file()}


def main():
    out=Path('/out');name=f'pod-{VERSION}-windows-hd-patch'
    os.environ.update(HOME='/out/home',WINEPREFIX='/out/wine',WINEARCH='win64',WINEDEBUG='-all',
        WINEDLLOVERRIDES='mscoree,mshtml=',LIBGL_ALWAYS_SOFTWARE='1',GOMAXPROCS='2',POD_NO_DIALOG='1')
    (out/'home').mkdir();(out/'unpack').mkdir()
    archive=Path('/packages')/(name+'.zip')
    with zipfile.ZipFile(archive) as z:assert z.testzip() is None;z.extractall(out/'unpack')
    root=out/'unpack'/name;exe=root/'PODCHT.exe';before=tree(root)
    try:
        phases=[('version',['-version']),('play',['-original',r'Z:\orig','-audio=false','-scale','3','-lang','ja','-frames','1250',
            '-script','900: ,1000:shot,1100:shot,1200:shot','-shot-dir',r'Z:\out\shots',
            '-shot-frame',r'Z:\out\frame.png','-shot-source',r'Z:\out\source'])]
        for label,args in phases:
            with (out/(label+'.log')).open('w') as f:subprocess.run(['xvfb-run','-a','wine',str(exe),*args],stdout=f,stderr=subprocess.STDOUT,timeout=900,check=True)
        assert VERSION in (out/'version.log').read_text()
        assert '字型缺字' not in (out/'play.log').read_text()
        games=list((out/'wine/drive_c/users').glob('*/AppData/Roaming/PODCHT/game'));assert len(games)==1
        game=games[0];caches=list((game.parent/'hd-cache').iterdir());assert len(caches)==1
        cache=caches[0];first=tree(cache);times={str(q.relative_to(cache)):q.stat().st_mtime_ns for q in cache.rglob('*') if q.is_file()};guest=tree(game)
        assert len(first)==2034 and len(list((out/'shots').glob('*.png')))==3
        with (out/'restart.log').open('w') as f:subprocess.run(['xvfb-run','-a','wine',str(exe),'-audio=false','-lang','ko','-frames','1'],stdout=f,stderr=subprocess.STDOUT,timeout=300,check=True)
        assert first==tree(cache) and before==tree(root)
        assert times=={str(q.relative_to(cache)):q.stat().st_mtime_ns for q in cache.rglob('*') if q.is_file()}
        after=tree(game);front={'pod-lang.txt','pod-map-visited.txt','pod-journal-seen.txt','pod-walkthrough-pos.txt','pod-speed.txt','pod-barfont.txt'}
        assert {k:v for k,v in guest.items() if k not in front}=={k:v for k,v in after.items() if k not in front}
        result=dict(status='PASS',version=VERSION,scope='Wine; native Windows unverified',zip_sha256=sha(archive),exe_sha256=sha(exe),
            cache_files=2034,cache=first,game=guest,cache_reused_without_write=True,package_unchanged=True,normal_frames=1250,initial_language='ja',restart_language='ko',shots=3)
        (out/'pass.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print('PASS Wine actual ZIP, Japanese1250 and Korean restart')
    finally:subprocess.run(['wineserver','-k'],timeout=15,check=False)


if __name__=='__main__':main()
