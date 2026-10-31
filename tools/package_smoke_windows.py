#!/usr/bin/env python3
"""在隔離Wine內從實際ZIP驗證Windows候選的首次匯入與開場。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import zipfile


def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('--version',required=True);a=ap.parse_args()
    out=Path('/out');packages=Path('/packages');name=f'pod-{a.version}-windows-hd-full-local'
    os.environ.update(HOME='/out/home',WINEPREFIX='/out/wine',WINEARCH='win64',WINEDEBUG='-all',
        WINEDLLOVERRIDES='mscoree,mshtml=',LIBGL_ALWAYS_SOFTWARE='1',GOMAXPROCS='2')
    (out/'home').mkdir();(out/'unpack').mkdir()
    with zipfile.ZipFile(packages/(name+'.zip')) as z:
        assert z.testzip() is None;z.extractall(out/'unpack')
    exe=out/'unpack'/name/'PODCHT.exe'
    try:
        for phase,args in [('version',['-version']),('play',['-audio=false','-scale','3','-lang','zh-TW','-frames','1250',
            '-script','900: ,1000:shot,1100:shot,1200:shot','-shot-dir',r'Z:\out\shots',
            '-shot-frame',r'Z:\out\frame.png','-shot-source',r'Z:\out\source','-draw-report',r'Z:\out\draw-report.json'])]:
            with (out/(phase+'.log')).open('w') as log:
                subprocess.run(['xvfb-run','-a','wine',str(exe),*args],stdout=log,stderr=subprocess.STDOUT,timeout=900,check=True)
        roots=list((out/'wine/drive_c/users').glob('*/AppData/Roaming/PODCHT/game'))
        assert len(roots)==1 and (roots[0]/'GAME.EXE').is_file() and (roots[0]/'pool4.cfg').is_file()
        assert len(list((out/'shots').glob('*.png')))==3 and (out/'frame.png').is_file()
        (out/'windows-pass.json').write_text(json.dumps(dict(status='PASS',scope='Wine first import and1250-frame normal title; not Windows physical machine',
            version=a.version,exe_sha256=hashlib.sha256(exe.read_bytes()).hexdigest(),
            user_game=str(roots[0].relative_to(out)),game_files={str(p.relative_to(roots[0])):hashlib.sha256(p.read_bytes()).hexdigest() for p in roots[0].rglob('*') if p.is_file()}),indent=2)+'\n')
        print('Windows Wine首次匯入與三張正常開場擷取通過')
    finally:
        subprocess.run(['wineserver','-k'],timeout=15,check=False)


if __name__=='__main__':main()
