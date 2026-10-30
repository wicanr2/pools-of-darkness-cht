#!/usr/bin/env python3
"""將數值與目視 SHA-256 都通過的 HD 圖匯出至私人審閱頁。"""
import argparse
import hashlib
import json
import re
import shutil
from pathlib import Path


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for name in ('req', 'out', 'validation', 'visual', 'review'):
        ap.add_argument('--'+name, type=Path, required=True)
    a = ap.parse_args()
    req = json.loads((a.req/'manifest.json').read_text())
    valid = json.loads(a.validation.read_text())
    visual = json.loads(a.visual.read_text())
    prepared = {i['key']: i for i in json.loads((a.out/'prepared.json').read_text())['items']}
    source = {i['key']: i for i in req['items']}
    if not valid['all_pass'] or valid['request_manifest_sha256'] != sha(a.req/'manifest.json'):
        raise ValueError('數值驗收未過或請求已變更')
    a.review.mkdir(parents=True, exist_ok=True)
    rows, table = [], ['# HD 圖片審閱', '', '風格 B，沿用使用者確認的奇幻手繪與原版配色。',
                      '本頁為素材對照；執行期覆蓋範圍見[主審閱入口](../README.md)。', '',
                      '| 來源 key／原版 3 倍 | HD 3 倍 |', '|---|---|']
    for result in valid['items']:
        key = result['key']
        if not re.fullmatch(r'[a-z0-9@-]+', key):
            raise ValueError('key 含路徑字元')
        s, pr = source[key], prepared[key]
        png, ref, raw = a.out/(key+'.png'), a.req/s['reference_3x'], a.out/pr['raw']
        if sha(png) != result['png_sha256'] or visual.get(key) != sha(png):
            raise ValueError(f'{key}: 缺目前 PNG 的目視驗收')
        if sha(ref) != s['reference_3x_sha256'] or sha(raw) != pr['raw_sha256']:
            raise ValueError(f'{key}: 原版參考或生成圖雜湊變更')
        raw_name = pr['raw'].removesuffix('-raw.png')
        job = json.loads((a.out/'jobs'/(raw_name+'.json')).read_text())
        if job.get('extra_reference'):
            extra = job['extra_reference']
            if Path(extra).name != extra or sha(a.req/extra) != job['extra_reference_sha256']:
                raise ValueError(f'{key}: 額外輪廓參考失配')
            shutil.copyfile(a.req/extra, a.review/extra)
        shutil.copyfile(png, a.review/(key+'.png'))
        shutil.copyfile(ref, a.review/(key+'-original.png'))
        rows.append(dict(key=key, source=s, generation=job, preparation=pr, numeric=result,
                         visual=dict(png_sha256=visual[key], status='accepted-after-viewing'),
                         rights='private original-reference review; AI redraw eligible for existing HD patch scope'))
        table.append(f'| {key}<br>![原版]({key}-original.png) | ![HD]({key}.png) |')
    (a.review/'README.md').write_text('\n'.join(table)+'\n\n完整來源、提示詞與 SHA-256 見 [manifest.json](manifest.json)。\n')
    (a.review/'manifest.json').write_text(json.dumps(dict(batch=req['batch'], count=len(rows),
        private_only=True, request_manifest_sha256=sha(a.req/'manifest.json'), items=rows), ensure_ascii=False, indent=2)+'\n')
    print(f'{len(rows)} 張數值／目視已驗收圖 → {a.review}')


if __name__ == '__main__':
    main()
