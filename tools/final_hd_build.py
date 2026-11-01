#!/usr/bin/env python3
"""Docker內建置正式程式及唯一開機佇列差異的FINAL測試程式。"""
import argparse
import difflib
import hashlib
import json
from pathlib import Path
import subprocess


def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--engine',type=Path,required=True);ap.add_argument('--out',type=Path,required=True)
    a=ap.parse_args(); assert Path('/.dockerenv').exists();a.out.mkdir(exist_ok=True)
    source=a.engine/'cmd/pod-play/main.go';before=source.read_text()
    old='Queue: []session.BootQueued{{Name: startName}}'
    new='Queue: []session.BootQueued{{Name: "VACE.EXE", Args: "4 FINAL 0 1"}}'
    assert before.count(old)==1
    after=before.replace(old,new); overlay_source=a.out/'main-final-player.go';overlay_source.write_text(after)
    diff=''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='cmd/pod-play/main.go',tofile='FINAL test overlay'))
    (a.out/'final-player-overlay.diff').write_text(diff)
    overlay=a.out/'overlay-final-player.json';overlay.write_text(json.dumps({'Replace':{str(source):str(overlay_source)}})+'\n')
    (a.out/'bin').mkdir(exist_ok=True);versions={};hashes={}
    for name,flags in (('pod-play',[]),('pod-final-player',['-overlay',str(overlay)])):
        target=a.out/'bin'/name
        cmd=['go','build','-p','2','-trimpath','-buildvcs=false',*flags,'-o',str(target),'./cmd/pod-play']
        subprocess.run(cmd,cwd=a.engine,check=True)
        info=subprocess.check_output(['go','version','-m',str(target)],text=True);(a.out/(name+'-build-info.txt')).write_text(info)
        versions[name]=info;hashes[name]=sha(target)
    proof=dict(scope='production CONTROL entry plus test-only original VACE FINAL player; not a normal victory',
        binary_sha256=hashes,entry_overlay_only=True,source_sha256=sha(source),overlay_source_sha256=sha(overlay_source),
        diff=diff,go_version=subprocess.check_output(['go','env','GOVERSION'],text=True).strip(),build_info=versions)
    (a.out/'build.json').write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n');print(hashes)


if __name__=='__main__':main()
