#!/usr/bin/env python3
"""Docker內匯出明示的乾淨Git來源，建立候選封包研究目錄。"""
import argparse
import json
import os
from pathlib import Path
import re
import subprocess


def git(root,*args):return subprocess.check_output(['git','-C',str(root),*args],text=True).strip()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for key in ('root','engine','out'):ap.add_argument('--'+key,type=Path,required=True)
    ap.add_argument('--main',required=True);ap.add_argument('--version',required=True);a=ap.parse_args()
    assert Path('/.dockerenv').exists() and re.fullmatch(r'v\.\d+\.\d+\.\d+-\d{8}',a.version)
    assert re.fullmatch('[0-9a-f]{40}',a.main)
    assert not a.out.exists() and a.out.parent.stat().st_uid==os.getuid()
    commit=json.loads((a.root/'review/hd/runtime/manifest.json').read_text())['engine']['commit']
    assert git(a.root,'rev-parse',a.main)==a.main and git(a.engine,'rev-parse',commit)==commit
    a.out.mkdir()
    for label,root,revision in [('main',a.root,a.main),('engine',a.engine,commit)]:
        archive=a.out/(label+'-source.tar')
        with archive.open('wb') as f:subprocess.run(['git','-C',str(root),'archive',revision],stdout=f,check=True,timeout=60)
        dest=a.out/label;dest.mkdir()
        subprocess.run(['tar','-xf',str(archive),'-C',str(dest)],check=True,timeout=60)
    for name in ('bin','test','capture','promo'): (a.out/name).mkdir()
    info=dict(version=a.version,main_commit=a.main,main_tree=git(a.root,'rev-parse',a.main+'^{tree}'),
              engine_commit=commit,rights=('private candidate' if a.version.startswith('v.0.') else 'private release')+'; originals only in local full builds; no public release')
    (a.out/'release-source.json').write_text(json.dumps(info,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(info,ensure_ascii=False))


if __name__=='__main__':main()
