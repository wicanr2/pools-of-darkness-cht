#!/usr/bin/env python3
"""以已驗收圖片與執行期來源證據建立既有格式的 POD HD 主題。"""
import argparse
import hashlib
import json
import shutil
from pathlib import Path


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--req', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--validation', type=Path, required=True)
    ap.add_argument('--proof', type=Path, required=True)
    ap.add_argument('--theme', type=Path, required=True)
    ap.add_argument('--pic-candidates',action='store_true',help='規格 010：登記已驗收 PIC1 的完整色號條件式候選')
    a = ap.parse_args()
    request = json.loads((a.req / 'manifest.json').read_text())
    valid = json.loads(a.validation.read_text())
    if not valid['all_pass'] or valid['request_manifest_sha256'] != sha(a.req / 'manifest.json'):
        raise ValueError('驗收未通過或請求已變更')
    sources = {i['key']: i for i in request['items']}
    positions = {}
    for capture in json.loads(a.proof.read_text()):
        for hit in capture['hits']:
            if not hit['used_palette_diff']:
                positions[hit['key']] = capture['rect']
    a.theme.mkdir(parents=True, exist_ok=True)
    rows, evidence, unique, files = [], [], {}, []
    for result in valid['items']:
        key = result['key']
        it = sources[key]
        observed=key in positions
        if not observed:
            if a.pic_candidates and it['source_file']=='PIC1.DAX' and it['source_size']==[88,88]:
                positions[key]=[24,24,88,88]
            else:raise ValueError(f'{key}: 缺執行期位置證據')
        png, idx = a.out / (key + '.png'), a.req / it['indexed']
        if sha(png) != result['png_sha256'] or sha(idx) != it['indexed_sha256']:
            raise ValueError(f'{key}: 輸入雜湊變更')
        x, y, w, h = positions[key]
        if [w, h] != it['source_size'] or len(idx.read_bytes()) != w*h:
            raise ValueError(f'{key}: 尺寸不符')
        signature=(x,y,w,h,sha(idx))
        evidence.append(dict(key=key,rect=[x,y,w,h],evidence='observed' if observed else 'conditional-candidate',
                             palette_sha256=it['palette_sha256']))
        if signature in unique:
            previous, png_sha=unique[signature]
            if png_sha!=sha(png):raise ValueError(f'{key}: 同色號候選與 {previous} 的 HD 不同')
            evidence[-1]['alias_of']=previous
            continue
        unique[signature]=(key,sha(png))
        files.extend(((png,key+'.png'),(idx,key+'.idx')))
        rows.append(dict(key=key, screen_rect=[x, y, w, h], png=key + '.png',
                         png_sha256=sha(png), indexed=key + '.idx', indexed_sha256=sha(idx)))
    # 全部來源與重複項先核對完，才寫檔；拒絕歧義時不留下半個主題。
    for path,name in files:shutil.copyfile(path,a.theme/name)
    temporary=a.theme/'theme.json.tmp'
    temporary.write_text(json.dumps(dict(scale=3,items=rows),indent=2)+'\n')
    temporary.replace(a.theme/'theme.json')
    (a.theme / 'provenance.json').write_text(json.dumps(dict(
        request_manifest_sha256=sha(a.req/'manifest.json'), validation_sha256=sha(a.validation),
        runtime_proof_sha256=sha(a.proof), evidence=evidence, keys=[i['key'] for i in rows]), indent=2)+'\n')
    print(f'{len(rows)} 項 → {a.theme}')


if __name__ == '__main__':
    main()
