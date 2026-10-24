#!/usr/bin/env python3
"""只沿用尺寸、原版色號與所有使用色 RGB 完全相同的實際生圖。"""
import argparse
import collections
import hashlib
import json
import re
import shutil
from pathlib import Path


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--req',type=Path,required=True)
    ap.add_argument('--out',type=Path,required=True)
    a=ap.parse_args()
    request=json.loads((a.req/'manifest.json').read_text())
    prepared={i['key']:i for i in json.loads((a.out/'prepared.json').read_text())['items']}
    groups=collections.defaultdict(list)
    for it in request['items']:
        idx_path,pal_path=a.req/it['indexed'],a.req/it['palette']
        if sha(idx_path)!=it['indexed_sha256'] or sha(pal_path)!=it['palette_sha256']:
            raise ValueError('來源雜湊變更')
        idx,pal=idx_path.read_bytes(),pal_path.read_bytes()
        used=bytes(v for c in sorted(set(idx)) for v in pal[c*3:c*3+3])
        groups[(tuple(it['source_size']),idx,used)].append(it)
    aliases,changed=[],[]
    for items in groups.values():
        canonical=items[0]['key']
        for it in items[1:]:
            key=it['key'];row=dict(key=key,canonical=canonical,indexed_equal=True,used_palette_equal=True)
            aliases.append(row)
            if sha(a.out/(key+'.png'))==sha(a.out/(canonical+'.png')):
                row['action']='already-identical';continue
            source=a.out/prepared[canonical]['raw']
            source_job=json.loads((a.out/'jobs'/(source.name.removesuffix('-raw.png')+'.json')).read_text())
            revisions=[int(m.group(1)) for p in a.out.glob(key+'-r*-raw.png')
                       if (m:=re.search(r'-r(\d+)-raw\.png$',p.name))]
            rev=max(revisions,default=-1)+1
            raw=a.out/f'{key}-r{rev}-raw.png';shutil.copyfile(source,raw)
            job=dict(key=key,version=rev,status='reused-exact-indexed-and-used-RGB',
                     generation='reused built-in image_gen redraw',source_key=canonical,
                     raw_sha256=sha(raw),source_indexed_sha256=it['indexed_sha256'],
                     source_palette_sha256=it['palette_sha256'],source_generation=source_job)
            (a.out/'jobs'/f'{key}-r{rev}.json').write_text(json.dumps(job,ensure_ascii=False,indent=2)+'\n')
            row['action']='reuse-canonical-raw';changed.append(key)
    (a.out/'deduplication.json').write_text(json.dumps(dict(unique_visible=len(groups),aliases=aliases,
        changed=changed),ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(dict(unique_visible=len(groups),aliases=aliases,changed=changed),ensure_ascii=False))


if __name__=='__main__':main()
