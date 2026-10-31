#!/usr/bin/env python3
"""由正式簡中顯示欄重建原生16／24遊戲字模與12／22工具列字模。"""
import argparse
import csv
import hashlib
import json
from pathlib import Path
import subprocess
import sys

sys.path.insert(0,str(Path(__file__).resolve().parent))
from zhcn_text import RECIPES

def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('--root',type=Path,required=True);ap.add_argument('--out',type=Path,required=True);a=ap.parse_args();r=a.root;a.out.mkdir(parents=True,exist_ok=True)
    chars=set(map(chr,range(32,127)));toolbar=set(chars)
    for target,_,columns in RECIPES:
        for row in csv.DictReader((r/'text'/target).open(),delimiter='\t',quoting=csv.QUOTE_NONE):
            for column in columns:chars.update(row[column])
    for code in ('zh-TW','zh-CN','en'):
        for row in csv.DictReader((r/'text'/('host-text.'+code+'.tsv')).open(),delimiter='\t',quoting=csv.QUOTE_NONE):toolbar.update(row['translation'])
    for collection in (chars,toolbar):collection.difference_update('\r\n\t');collection.add(' ')
    proof={}
    for cell,suffix,points in ((16,'',chars),(24,'-3x',chars),(12,'-bar12',toolbar),(22,'-bar22',toolbar)):
        charset=a.out/('charset'+suffix+'.json');charset.write_text(json.dumps(sorted(map(ord,points)))+'\n')
        otf=r/'font/source/rounded'/('GenSenRounded2TW-B.otf' if cell in (22,24) else 'GenSenRounded2TW-M.otf')
        p=a.out/('pod-zh-CN'+suffix+'.golemfnt')
        subprocess.run([sys.executable,'-I',str(r/'tools/rounded_font.py'),'--otf',str(otf),'--cell',str(cell),'--charset-file',str(charset),'--fallback',str(r/'font/source/unifont/unifont_t-17.0.05.hex'),'--out',str(p)],check=True)
        proof[p.name]=dict(cell=cell,characters=len(points),sha256=hashlib.sha256(p.read_bytes()).hexdigest(),charset_sha256=hashlib.sha256(charset.read_bytes()).hexdigest())
    (a.out/'fonts.json').write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n');print(json.dumps(proof,indent=2))

if __name__=='__main__':main()
