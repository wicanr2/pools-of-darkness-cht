#!/usr/bin/env python3
"""用正式前端驗證缺語系表與錯誤3倍字模拒絕，原始text及font唯讀。"""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess

def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('--root',type=Path,required=True);ap.add_argument('--out',type=Path,required=True);a=ap.parse_args();r=a.root;o=a.out
    d=o/'negative';d.mkdir();shutil.copytree(r/'text',d/'text');(d/'text/host-text.zh-CN.tsv').unlink();(d/'font').mkdir()
    shutil.copyfile(o/'font/pod-zh-CN.golemfnt',d/'font/pod-zh-CN.golemfnt');shutil.copyfile(o/'font/pod-zh-CN-bar12.golemfnt',d/'font/pod-zh-CN-3x.golemfnt')
    base=['xvfb-run','-a',str(o/'bin/pod-play'),'-original',str(r/'workplace/input/dos/pod'),'-font',str(r/'workplace/font/pod-rounded.golemfnt'),'-lang','zh-CN','-audio=false','-frames','1']
    checks=[];env=dict(os.environ,POD_NO_DIALOG='1',GOMAXPROCS='2')
    for name,text,font,expected in [('missing-host',d/'text',o/'font/pod-zh-CN.golemfnt','沒有這個語言或缺字型'),('bad-scaled-font',r/'text',d/'font/pod-zh-CN.golemfnt','24×24')]:
        p=subprocess.run(base+['-text-dir',str(text),'-lang-fonts','zh-CN='+str(font),'-save',str(d/name)],env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=60)
        log=p.stdout.decode();(d/(name+'.log')).write_text(log)
        if p.returncode==0 or expected not in log:raise ValueError('負對照沒有按預期拒絕：'+name+'\n'+log)
        checks.append(dict(case=name,exit_code=p.returncode,rejected=True))
    (d/'receipt.json').write_text(json.dumps(checks,ensure_ascii=False,indent=2)+'\n');print(checks)

if __name__=='__main__':main()
