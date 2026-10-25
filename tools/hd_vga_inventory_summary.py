#!/usr/bin/env python3
"""只讀盤點：依尺寸、色號與實際用色去重，不以缺少key推斷尚未重畫。"""
import argparse
from collections import defaultdict
import hashlib
import json
from pathlib import Path


def sha(data):
    return hashlib.sha256(data).hexdigest()


def summarize(root):
    inputs, groups, families, seen = {}, defaultdict(list), defaultdict(set), set()
    for name in ('inventory.json', 'inventory-other.json'):
        data = (root/name).read_bytes()
        inputs[name] = sha(data)
        for item in json.loads(data)['items']:
            key = item['key']
            assert key not in seen, f'duplicate inventory key {key}'
            seen.add(key)
            idx, pal = ((root/(key+ext)).read_bytes() for ext in ('.idx','.pal'))
            assert len(idx) == item['width']*item['height'] and len(pal) == 768
            assert sha(idx) == item['indexed_sha256'] and sha(pal) == item['palette_sha256']
            used = bytes(v for c in sorted(set(idx)) for v in pal[c*3:c*3+3])
            signature = (item['width'],item['height'],sha(idx),sha(used))
            groups[signature].append({'key':key,'file':item['file']})
            families[item['file']].add(key)
    unique_by_family = defaultdict(int)
    for rows in groups.values():
        for name in {row['file'] for row in rows}:
            unique_by_family[name] += 1
    return {'scope':'two existing image catalogs only; excludes separately inventoried wall banks and INTRO/FINAL; exact size/indexed/used-RGB; no mirror equivalence or completion inference',
            'inputs':inputs, 'source_states':len(seen), 'unique_visible_signatures':len(groups),
            'families':[{'file':name,'source_states':len(keys),'unique_visible_signatures':unique_by_family[name]}
                        for name,keys in sorted(families.items())],
            'aliases':[{'size':list(signature[:2]),'indexed_sha256':signature[2],
                        'used_palette_sha256':signature[3],'sources':rows}
                       for signature,rows in groups.items() if len(rows)>1]}


if __name__ == '__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--inventory',type=Path,required=True)
    parser.add_argument('--out',type=Path,required=True)
    args=parser.parse_args()
    report=summarize(args.inventory)
    args.out.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    for row in report['families']:
        print(row['file'],row['source_states'],row['unique_visible_signatures'])
    print('全部來源狀態',report['source_states'],'可見圖樣',report['unique_visible_signatures'])
