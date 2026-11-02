#!/usr/bin/env python3
"""在Xvfb內用xdotool操作實際Ebiten視窗，驗證兩倍率與海報偏移。"""
import argparse,hashlib,json,os,subprocess,time
from pathlib import Path

def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('scale',type=int,choices=[2,3]);a=ap.parse_args()
    assert Path('/.dockerenv').exists();out=Path('/out')/f'mouse-{a.scale}';out.mkdir()
    root=Path('/repo');assets=root/'workplace/package-promo-v1/assets'
    os.environ.update(HOME=str(out/'home'),XDG_CONFIG_HOME=str(out/'config'),POD_NO_DIALOG='1',LIBGL_ALWAYS_SOFTWARE='1',GOMAXPROCS='2')
    Path(os.environ['HOME']).mkdir()
    args=['/out/pod-play-fixed','-original','/original','-save',str(out/'game'),'-text-dir',str(root/'text'),
          '-font',str(assets/'font/pod-rounded.golemfnt'),'-hd',str(assets/'theme'),'-posters',str(assets/'theme/posters'),
          '-audio=false','-scale',str(a.scale),'-lang','zh-TW','-receipt-host-input','-pointer-audit',
          '-frames','2450','-script','900: ,1800:p,1900:Down,1950:Down,2000:shot,2100:shot,2200:shot,2300:shot,2400:shot',
          '-shot-dir',str(out/'shots'),'-shot-frame-dir',str(out/'frames'),'-shot-source-dir',str(out/'source')]
    def xd(*args):return subprocess.check_output(['xdotool',*map(str,args)],text=True,timeout=10).strip()
    with (out/'play.log').open('w') as log:
        proc=subprocess.Popen(args,stdout=log,stderr=subprocess.STDOUT)
        try:
            limit=time.monotonic()+600
            def wait_frame(n):
                f=out/'shots'/f'pod-f{n:06d}.png'
                while not f.exists():
                    assert proc.poll() is None,(proc.returncode,(out/'play.log').read_text()[-1000:])
                    assert time.monotonic()<limit,'GUI timeout';time.sleep(.05)
            wait_frame(2000)
            win=xd('search','--onlyvisible','--name','黑暗之池').splitlines()[-1]
            xd('windowfocus','--sync',win)
            def click(button):
                xd('mousedown',button);time.sleep(.12);xd('mouseup',button)
            # B版海報每側80個邏輯像素。原版Load Saved Game位於第14列。
            xd('mousemove','--window',win,(80+120)*a.scale,115*a.scale);click(1)
            wait_frame(2100)
            click(3)
            wait_frame(2200)
            # 左鍵按住拖到海報、返回遊戲、放開。返回不應再有Press。
            xd('mousedown','1');time.sleep(.12)
            xd('mousemove','--window',win,20*a.scale,100*a.scale);time.sleep(.12)
            xd('mousemove','--window',win,(80+120)*a.scale,115*a.scale);time.sleep(.12);xd('mouseup','1')
            wait_frame(2300)
            # 工具列和海報右鍵由前端隔離。
            xd('mousemove','--window',win,100*a.scale,2*a.scale);click(3)
            xd('mousemove','--window',win,20*a.scale,100*a.scale);click(3)
            proc.wait(timeout=90);assert proc.returncode==0
        finally:
            if proc.poll() is None:proc.terminate();proc.wait(timeout=10)
    text=(out/'play.log').read_text();lines=[x for x in text.splitlines() if x.startswith('pointer ')]
    assert any('buttons=2 ' in x and 'target=1 ' in x and 'dos_calls=1' in x for x in lines),lines
    assert any('target=3 ' in x and 'dos_calls=0' in x for x in lines),lines
    before=(out/'source/pod-f002000.idx').read_bytes();after=(out/'source/pod-f002100.idx').read_bytes()
    assert before!=after,'原版畫面未回應滑鼠Load Saved Game'
    result=dict(status='PASS',scale=a.scale,script=args,events=lines,original_menu_changed=True,
                screenshots={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in (out/'shots').glob('*.png')})
    (out/'pass.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print('mouse',a.scale,'PASS',len(lines))
if __name__=='__main__':main()
