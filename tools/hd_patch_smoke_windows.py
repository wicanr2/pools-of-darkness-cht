#!/usr/bin/env python3
"""實際HD patch ZIP的Wine首次匯入／重啟；不冒稱Windows真機。"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import zipfile

VERSION='v.0.1.1-20261101'
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def tree(p):return {str(q.relative_to(p)):sha(q) for q in p.rglob('*') if q.is_file()}


def main():
    out=Path('/out');name=f'pod-{VERSION}-windows-hd-patch'
    os.environ.update(HOME='/out/home',WINEPREFIX='/out/wine',WINEARCH='win64',WINEDEBUG='-all',
        WINEDLLOVERRIDES='mscoree,mshtml=',LIBGL_ALWAYS_SOFTWARE='1',GOMAXPROCS='2',POD_NO_DIALOG='1')
    (out/'home').mkdir();(out/'unpack').mkdir()
    with zipfile.ZipFile(Path('/packages')/(name+'.zip')) as z:assert z.testzip() is None;z.extractall(out/'unpack')
    root=out/'unpack'/name;exe=root/'PODCHT.exe';before=tree(root)
    try:
        phases=[('version',['-version']),('play',['-original',r'Z:\orig','-audio=false','-scale','3','-lang','zh-TW','-frames','1250',
            '-script','900: ,1000:shot,1100:shot,1200:shot','-shot-dir',r'Z:\out\shots',
            '-shot-frame',r'Z:\out\frame.png','-shot-source',r'Z:\out\source'])]
        for phase,args in phases:
            with (out/(phase+'.log')).open('w') as log:subprocess.run(['xvfb-run','-a','wine',str(exe),*args],stdout=log,stderr=subprocess.STDOUT,timeout=900,check=True)
        games=list((out/'wine/drive_c/users').glob('*/AppData/Roaming/PODCHT/game'));assert len(games)==1
        game=games[0];caches=list((game.parent/'hd-cache').iterdir());assert len(caches)==1
        cache=caches[0];first=tree(cache);times={str(q.relative_to(cache)):q.stat().st_mtime_ns for q in cache.rglob('*') if q.is_file()};guest=tree(game)
        assert len(first)==1866 and (game/'pool4.cfg').is_file() and len(list((out/'shots').glob('*.png')))==3
        with (out/'restart.log').open('w') as log:subprocess.run(['xvfb-run','-a','wine',str(exe),'-audio=false','-frames','1'],stdout=log,stderr=subprocess.STDOUT,timeout=300,check=True)
        assert first==tree(cache) and guest==tree(game) and before==tree(root)
        assert times=={str(q.relative_to(cache)):q.stat().st_mtime_ns for q in cache.rglob('*') if q.is_file()}
        (out/'windows-pass.json').write_text(json.dumps(dict(status='PASS',version=VERSION,scope='Wine only; not native Windows',
             exe_sha256=sha(exe),cache_files=len(first),cache=first,game=guest,cache_reused_without_write=True,
             package_unchanged=True,normal_frames=1250,shots=3),indent=2)+'\n')
        print('Wine actual ZIP first import,1250 frames,3 shots and cache reuse PASS')
    finally:subprocess.run(['wineserver','-k'],timeout=15,check=False)


if __name__=='__main__':main()
