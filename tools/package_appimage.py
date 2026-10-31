#!/usr/bin/env python3
"""用固定type2 runtime與可重現SquashFS組合AppImage，彙整九個候選封包。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for name in ('stage','runtime','out','logs'):ap.add_argument('--'+name,type=Path,required=True)
    a=ap.parse_args();assert not a.out.exists();a.out.mkdir()
    assert sha(a.runtime)=='2fca8b443c92510f1483a883f60061ad09b46b978b2631c807cd873a47ec260d'
    recipes=json.loads((a.stage/'recipes.json').read_text());rows=[]
    for row in recipes['items']:
        if row['platform']=='linux':
            tree=Path(row['tree']);squash=a.out/(row['name']+'.squashfs')
            args=['mksquashfs',str(tree),str(squash),'-noappend','-all-root','-mkfs-time','0','-all-time','0','-no-xattrs','-processors','2','-comp','gzip']
            p=subprocess.run(args,capture_output=True,text=True,timeout=180)
            (a.logs/('mksquashfs-'+row['name']+'.log')).write_text(p.stdout+p.stderr)
            if p.returncode:raise RuntimeError(p.stderr)
            target=a.out/(row['name']+'.AppImage')
            with target.open('wb') as dst:
                for source in (a.runtime,squash):
                    with source.open('rb') as src:shutil.copyfileobj(src,dst)
            target.chmod(0o755)
            offset=int(subprocess.check_output([str(target),'--appimage-offset'],text=True,timeout=15).strip())
            assert offset==a.runtime.stat().st_size
            with target.open('rb') as f:f.seek(offset);assert f.read(4)==b'hsqs'
            squash.unlink()
        else:
            target=a.out/Path(row['archive']).name;shutil.copyfile(row['archive'],target)
            assert sha(target)==row['sha256']
        rows.append(dict(name=target.name,kind=row['kind'],platform=row['platform'],edition=row['edition'],version=recipes['version'],engine_commit=recipes['engine_commit'],size=target.stat().st_size,sha256=sha(target),source_tree=row['tree'],payload_manifest_sha256=row['manifest_sha256']))
    assert len(rows)==9
    (a.out/'packages.json').write_text(json.dumps(dict(version=recipes['version'],engine_commit=recipes['engine_commit'],candidate=True,files=rows),ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(dict(packages=len(rows),bytes=sum(r['size'] for r in rows)),indent=2))


if __name__=='__main__':main()
