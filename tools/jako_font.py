#!/usr/bin/env python3
"""由指定日韓資料集產生原生16／24與工具列12／22字模；不碰正式字型。"""
import argparse
import csv
import hashlib
import json
from pathlib import Path
import subprocess
import sys


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for key in ('root','text','out'):ap.add_argument('--'+key,type=Path,required=True)
    a=ap.parse_args();a.out.mkdir(parents=True,exist_ok=True);proof={}
    for lang in ('ja','ko'):
        chars=set(map(chr,range(32,127)));bar=set(chars);inputs={}
        catalogs=list(a.text.glob('*.'+lang+'.tsv'));assert catalogs
        for p in catalogs:
            inputs[p.name]=sha(p)
            for row in csv.DictReader(p.open(),delimiter='\t',quoting=csv.QUOTE_NONE):
                for field in ('translation','title','body','label','summary'):
                    if field in row:chars.update(row[field])
                if p.name.startswith('host-text.'):
                    bar.update(row['translation'])
        for points in (chars,bar):points.difference_update('\t\n\r')
        files={}
        for cell,suffix,points in ((16,'',chars),(24,'-3x',chars),(12,'-bar12',bar),(22,'-bar22',bar)):
            charset=a.out/('charset-'+lang+suffix+'.json');charset.write_text(json.dumps(sorted(map(ord,points)))+'\n')
            region='JP' if lang=='ja' else 'TW';weight='B' if cell in (22,24) else 'M'
            otf=a.root/'font/source/rounded'/f'GenSenRounded2{region}-{weight}.otf'
            fallback=a.root/'font/source/unifont'/('unifont_jp-17.0.05.hex' if lang=='ja' else 'unifont_t-17.0.05.hex')
            dest=a.out/('pod-'+lang+suffix+'.golemfnt')
            subprocess.run([sys.executable,'-I',str(a.root/'tools/rounded_font.py'),'--otf',str(otf),'--cell',str(cell),
                            '--charset-file',str(charset),'--fallback',str(fallback),'--out',str(dest)],check=True)
            files[dest.name]=dict(cell=cell,characters=len(points),sha256=sha(dest),charset=charset.name,
                                  charset_sha256=sha(charset),otf_sha256=sha(otf),fallback_sha256=sha(fallback))
        proof[lang]=dict(inputs=inputs,fonts=files)
    (a.out/'fonts.json').write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(proof,ensure_ascii=False,indent=2))


if __name__=='__main__':main()
