#!/usr/bin/env python3
"""Docker內抽驗實際AppImage，所有匯入與存檔只寫/out。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess


VERSION=os.environ.get('POD_SMOKE_VERSION','v.0.1.2-20261102')
THEME_FILES=int(os.environ.get('POD_SMOKE_THEME_FILES','2034'))
AUTOFILL=os.environ.get('POD_SMOKE_AUTOFILL','0')=='1'
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def tree(p):return {str(q.relative_to(p)):sha(q) for q in sorted(p.rglob('*')) if q.is_file()}


def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('mode',choices=['intro','base','full','demo','save','missing']);a=ap.parse_args()
    out=Path('/out');assert Path('/.dockerenv').exists()
    for name in ('home','data','config'):(out/name).mkdir()
    os.environ.update(HOME='/out/home',XDG_DATA_HOME='/out/data',XDG_CONFIG_HOME='/out/config',POD_NO_DIALOG='1',
        APPIMAGE_EXTRACT_AND_RUN='1',GOMAXPROCS='2',LIBGL_ALWAYS_SOFTWARE='1')
    hd=Path('/input')/f'pod-{VERSION}-linux-hd-patch.AppImage';runs=[]
    def run(label,args,app=hd,ok=True):
        with (out/(label+'.log')).open('w') as f:
            p=subprocess.run(['xvfb-run','-a',str(app),*args],stdout=f,stderr=subprocess.STDOUT,timeout=900)
        text=(out/(label+'.log')).read_text();assert (p.returncode==0)==ok,(label,p.returncode,text[-1000:])
        assert '字型缺字' not in text and '截圖失敗' not in text
        end=re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',text)
        digest=re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)',text)
        runs.append(dict(label=label,app=app.name,app_sha256=sha(app),args=args,code=p.returncode,
            log_sha256=sha(out/(label+'.log')),end=list(end.groups()) if end else None,digest=list(digest.groups()) if digest else None))
        return text
    def snapshot(name):
        p=out/'data/pod-cht';caches=list((p/'hd-cache').iterdir());assert len(caches)==1
        cache=caches[0];assert len(tree(cache))==THEME_FILES
        record=dict(game=tree(p/'game'),cache=tree(cache),mtime={str(q.relative_to(cache)):q.stat().st_mtime_ns for q in cache.rglob('*') if q.is_file()})
        (out/(name+'.json')).write_text(json.dumps(record,indent=2)+'\n');return record,cache
    run('version',['-version']);assert VERSION in (out/'version.log').read_text()
    common=['-audio=false','-lang','zh-TW','-scale','3']
    extra={}
    if a.mode=='intro':
        script='900: ,1000:shot,1001:lang,1010:shot,1011:lang,1020:shot,1021:lang,1030:shot,1031:lang,1040:shot,1041:lang,1050:shot,1060:scale,1070:shot,1080:toggle,1090:shot,1100:scale,1110:shot,1120:toggle,1130:shot'
        for label,prefix in [('first',''),('restart','restart-')]:
            run(label,common+['-original','/original','-frames','1250','-script',script,
                '-shot-dir',f'/out/{prefix}shots','-shot-frame-dir',f'/out/{prefix}window',
                '-shot-source-dir',f'/out/{prefix}sources','-shot-source',f'/out/{prefix}source','-shot-frame',f'/out/{prefix}frame.png'])
            snap,cache=snapshot(label)
            if label=='first':first=snap
            else:assert first==snap and runs[-1]['end']==runs[-2]['end'] and runs[-1]['digest']==runs[-2]['digest']
        for folder in ('shots','window','sources'):assert tree(out/folder)==tree(out/('restart-'+folder))
        assert len(list((out/'window').glob('*.png')))==10
        target=cache/'final/source/b00-f00.ops';saved=target.read_bytes();target.write_bytes(b'wrong')
        run('corrupt-cache',['-audio=false','-frames','1'],ok=False);target.write_bytes(saved)
        target=out/'data/pod-cht/game/FINAL.BMA';saved=target.read_bytes();target.write_bytes(b'wrong')
        run('corrupt-final',['-audio=false','-frames','1'],ok=False);target.write_bytes(saved)
        assert tree(cache)==first['cache'] and tree(out/'data/pod-cht/game')==first['game']
        extra=dict(cache_files=2034,cache_reused_without_write=True,first=first,negative_cache=True,negative_final=True)
    elif a.mode=='base':
        app=Path('/input')/f'pod-{VERSION}-linux-base-patch.AppImage'
        run('base',['-original','/original','-audio=false','-lang','ja','-frames','1250','-script','900: ,1200:shot',
                    '-shot-dir','/out/shots','-shot-frame','/out/frame.png'],app)
        assert not (out/'data/pod-cht/hd-cache').exists()
    elif a.mode=='full':
        for edition in ('base','hd'):
            app=Path('/input')/f'pod-{VERSION}-linux-{edition}-full-local.AppImage'
            run('full-'+edition,['-audio=false','-lang','ko','-frames','1'],app)
        assert not (out/'data/pod-cht/hd-cache').exists()
    elif a.mode=='demo':
        text=run('demo',common+['-original','/original','-frames','3800',
            '-script','900: ,1800:d,2050:shot,2300:x,2500:shot,3000:shot,3200:p,3700:shot',
            '-flash-audit','-shot-dir','/out/shots','-shot-frame','/out/frame.png','-shot-source','/out/source'])
        assert '閃現 0 次' in text
        assert re.search(r'中止 1 次、中止後返回 1 次',text) and '執行中 false' in text
        snapshot('demo')
    elif a.mode=='save':
        run('install',['-original','/original','-audio=false','-frames','1'])
        seed=Path('/seed');assert (seed/'savgamA.pty').is_file() and not (seed/'savgamB.pty').exists()
        assert set(tree(seed))=={'savgamA.pty','CHRDATA1.efx','CHRDATA1.sav','CHRDATA1.thg','vaultA.dat'}
        game=out/'data/pod-cht/game';shutil.copytree(seed,game/'SAVE',dirs_exist_ok=True)
        for p in game.rglob('*'):
            if p.is_file():os.utime(p,(946684800,946684800))
        save_text=run('save',['-audio=false','-scale','2','-lang','zh-TW','-posters','-','-frames','4700',
            '-script','900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:A,3400:Down,3420:Down,3440:Down,3460:Down,3480:Down,3500:Return,3900:shot,'+('' if AUTOFILL else '4000:Y,4010:O,4020:U,')+'4030:Return,4200:shot,4300:E,4400:S,4450:shot,4500:B,4550:shot,4650:shot',
            '-shot-dir','/out/save-shots','-shot-source','/out/save-source'])
        if AUTOFILL:assert len(re.findall(r'quiz autofill frame=\d+ answer=YOU enter=false',save_text))==1
        assert (game/'SAVE/savgamB.pty').stat().st_size>0;saved=tree(game/'SAVE');snapshot('saved')
        run('load',['-audio=false','-scale','2','-lang','zh-TW','-posters','-','-frames','3250',
            '-script','900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:B,3200:shot',
            '-shot-dir','/out/load-shots','-shot-source','/out/load-source'])
        assert tree(game/'SAVE')==saved;snapshot('loaded');extra=dict(save_b_sha256=sha(game/'SAVE/savgamB.pty'),save_tree=saved)
    else:
        (out/'empty').mkdir();text=run('missing',['-original','/out/empty','-audio=false','-frames','1'],ok=False)
        assert '1.10' in text and not list((out/'data').rglob('GAME.EXE'))
    result=dict(status='PASS',version=VERSION,mode=a.mode,runs=runs,**extra)
    (out/'pass.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print('PASS Linux',a.mode,len(runs),'runs')


if __name__=='__main__':main()
