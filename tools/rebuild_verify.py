#!/usr/bin/env python3
"""核對跨機重建的7400格正常路徑，不借用舊工作目錄。"""
import argparse
import hashlib
import json
from pathlib import Path
import re

def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()

def tree(p):
    return {str(f.relative_to(p)):sha(f) for f in sorted(p.rglob('*')) if f.is_file()}

def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root',type=Path,required=True)
    ap.add_argument('--out',type=Path,required=True)
    a=ap.parse_args();m=a.root/'review/hd/runtime/manifest.json';lock=json.loads(m.read_text());expect=lock['runtime_expected'];run=a.out/'smoke'
    if (a.root/'tools/rebuild-engine.commit').read_text().strip()!=lock['engine']['commit']:
        raise ValueError('引擎commit與素材清單不同')
    log=(run/'run.log').read_text()
    end=re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',log)
    digest=re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)',log)
    if not end or list(end.groups())!=expect['end'] or not digest or list(digest.groups())!=expect['digest']:
        raise ValueError('正常路徑原版終點不同')
    if 'pod-play: 閃現 0 次' not in log:
        raise ValueError('閃現量尺未通過')
    for name,key in [('source.idx','source_indexed_sha256'),('source.pal','source_palette_sha256'),('frame.png','frame_sha256')]:
        if sha(run/name)!=expect[key]:raise ValueError(f'正常路徑圖片／來源不同：{name}')
    front={'pod-map-visited.txt','pod-journal-seen.txt','pod-walkthrough-pos.txt','pod-lang.txt','pod-speed.txt','pod-barfont.txt'}
    guest={k:v for k,v in tree(run/'save').items() if k not in front}
    if guest!=expect['guest']:raise ValueError('正常路徑客體檔案不同')
    if tree(a.out/'theme')!={r['path']:r['sha256'] for r in lock['theme']}:raise ValueError('主題檔案集合不同')
    if tree(a.out/'font')!={r['path']:r['sha256'] for r in lock['fonts']}:raise ValueError('字型檔案集合不同')
    if tree(a.root/'text')!=expect['text_tree']:raise ValueError('正式文字不同，須先重驗正常基準')
    if sha(a.root/expect['baseline_receipt'])!=expect['baseline_receipt_sha256']:raise ValueError('基準收據不同')
    proof=dict(status='PASS',scope='clean Git source; 7400-frame normal HD DEMO; Linux amd64 container',
               engine_commit=lock['engine']['commit'],binary_sha256=sha(a.out/'bin/pod-play'),
               manifest_sha256=sha(m),baseline_receipt_sha256=expect['baseline_receipt_sha256'],
               cpu_memory_source_equal=True,guest_files=len(guest),guest_equal=True,
               frame_sha256=sha(run/'frame.png'),flash_count=0,theme_files=len(lock['theme']),fonts=len(lock['fonts']))
    (run/'receipt.json').write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(proof,ensure_ascii=False,indent=2))

if __name__=='__main__':main()
