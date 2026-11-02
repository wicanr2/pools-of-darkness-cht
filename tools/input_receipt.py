#!/usr/bin/env python3
"""Docker內由正常冷啟動抽驗查驗題，不使用狀態注入。"""
import argparse,hashlib,json,os,re,shutil,subprocess
from pathlib import Path

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('mode',choices=['auto','manual','raw','retry'])
    ap.add_argument('--suffix',default='');ap.add_argument('--binary',default='/out/pod-play')
    a=ap.parse_args();assert Path('/.dockerenv').exists()
    root=Path('/repo');out=Path('/out')/(a.mode+a.suffix);out.mkdir()
    os.environ.update(HOME=str(out/'home'),XDG_CONFIG_HOME=str(out/'config'),POD_NO_DIALOG='1',LIBGL_ALWAYS_SOFTWARE='1',GOMAXPROCS='2')
    Path(os.environ['HOME']).mkdir()
    assets=root/'workplace/package-promo-v1/assets'
    args=[a.binary,'-original','/original','-save',str(out/'game'),'-text-dir',str(root/'text'),
          '-font',str(assets/'font/pod-rounded.golemfnt'),'-hd',str(assets/'theme'),'-posters','-',
          '-audio=false','-scale','3','-lang','zh-TW']
    def run(name,extra):
        with (out/(name+'.log')).open('w') as log:
            p=subprocess.run(['xvfb-run','-a',*args,*extra],stdout=log,stderr=subprocess.STDOUT,timeout=600)
        text=(out/(name+'.log')).read_text();assert p.returncode==0,(p.returncode,text[-1500:]);assert '字型缺字' not in text
        return text
    run('install',['-frames','1'])
    shutil.copytree(root/'workplace/package-promo-v1/seed-clean',out/'game/SAVE',dirs_exist_ok=True)
    for p in (out/'game').rglob('*'):
        if p.is_file():os.utime(p,(946684800,946684800))
    route='900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:A,3400:Down,3420:Down,3440:Down,3460:Down,3480:Down,3500:Return,3900:shot'
    if a.mode=='manual':route='100:answers,'+route+',4000:Y,4010:O,4020:U,4030:Return,4200:shot'
    elif a.mode=='retry':route='100:answers,'+route+',4000:X,4030:Return,4200:answers,4400:shot,4600:Return,4750:shot'
    else:route+=',4030:Return,4200:shot'
    if a.mode=='raw':route='100:toggle,'+route
    if a.mode=='auto':
        route='3490:toggle,'+route
        route=route.replace('3900:shot','3800:shot,3810:toggle,3820:shot,3830:lang,3840:shot,3850:lang,3860:shot,3870:lang,3880:shot,3890:lang,3900:shot,3910:lang')
    frames=4800 if a.mode=='retry' else 4300
    text=run('play',['-frames',str(frames),'-script',route,'-shot-dir',str(out/'shots'),
         '-shot-frame-dir',str(out/'frames'),'-shot-source-dir',str(out/'source'),
         '-shot-source',str(out/'final')])
    fills=re.findall(r'quiz autofill frame=(\d+) answer=([A-Z]+) enter=false',text)
    assert len(fills)==(0 if a.mode=='manual' else 1),fills
    if a.mode in ('auto','raw'):assert fills[0][1]=='YOU',fills
    end=re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',text)
    assert end
    result=dict(mode=a.mode,status='PASS',binary_sha256=sha(Path(a.binary)),script=route,
                fills=fills,end=list(end.groups()),shots={p.name:sha(p) for p in (out/'shots').glob('*.png')},
                save={p.name:sha(p) for p in (out/'game/SAVE').iterdir() if p.is_file()})
    (out/'pass.json').write_text(json.dumps(result,indent=2)+'\n');print(a.mode,'PASS',fills)
if __name__=='__main__':main()
