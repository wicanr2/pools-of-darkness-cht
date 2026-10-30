#!/usr/bin/env python3
"""必要輸入失配須在寫出任何重建結果前拒絕。"""
import argparse
import json
from pathlib import Path
import shutil
import sys
import tempfile

sys.path.insert(0,str(Path(__file__).resolve().parent))
from rebuild_assets import rebuild

def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root',type=Path,required=True)
    ap.add_argument('--original',type=Path,required=True)
    ap.add_argument('--report',type=Path,required=True)
    a=ap.parse_args();lock=json.loads((a.root/'review/hd/runtime/manifest.json').read_text());results=[]
    with tempfile.TemporaryDirectory(prefix='pod-rebuild-negative-') as temp:
        tmp=Path(temp)
        # 只複製已列明的輸入檔，保留正式路徑守衛，不連結舊workplace。
        sources=set(lock['source_files'])|{r['source'] for r in lock['theme'] if r['kind']=='repo-file'}|{'review/hd/runtime/manifest.json'}
        fake=tmp/'root';fake.mkdir()
        for name in sources:
            p=fake/name;p.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(a.root/name,p)
        original=tmp/'original';original.mkdir()
        for name in lock['original_archives']:shutil.copyfile(a.original/name,original/name)
        for kind in ['missing-original','changed-original','changed-png']:
            output=tmp/('output-'+kind)
            if kind=='missing-original':
                victim=original/next(iter(lock['original_archives']));victim.unlink()
            elif kind=='changed-original':
                victim.write_bytes(b'not the accepted original')
            else:
                victim.unlink();shutil.copyfile(a.original/victim.name,victim)
                name=next(r['source'] for r in lock['theme'] if r['path'].endswith('.png'))
                image=fake/name;data=bytearray(image.read_bytes());image.unlink();data[-1]^=1;image.write_bytes(data)
            try:
                rebuild(fake,original,output)
            except (OSError,ValueError) as e:
                if output.exists():raise AssertionError('拒絕後仍留下輸出')
                reason=str(e)
                if kind=='changed-original' and '原版雜湊不符' not in reason:raise
                if kind=='changed-png' and '素材雜湊不符' not in reason:raise
                results.append(dict(case=kind,rejected=True,no_output=True,reason=reason.split(': ')[0]))
            else:raise AssertionError(kind+'未拒絕')
    a.report.write_text(json.dumps(dict(all_pass=True,cases=results),ensure_ascii=False,indent=2)+'\n')
    print('3個輸入失配負對照通過')

if __name__=='__main__':main()
