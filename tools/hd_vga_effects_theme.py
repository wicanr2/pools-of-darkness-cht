#!/usr/bin/env python3
"""將已驗收的COMSPR圖匯入既有effects.json v1；只寫獨立候選主題。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for name in ('req', 'out', 'theme'):
        ap.add_argument('--' + name, type=Path, required=True)
    a = ap.parse_args()
    request = json.loads((a.req / 'manifest.json').read_text())
    valid = json.loads((a.out / 'validation.json').read_text())
    visual = json.loads((a.out / 'visual-review.json').read_text())
    assert valid['all_pass'] and valid['scope'] == 'all'
    assert valid['request_manifest_sha256'] == sha(a.req / 'manifest.json')
    numeric = {row['key']: row for row in valid['items']}
    target = a.theme / 'effects.json'
    manifest = dict(version=1, scale=3, kind='combat-effects',
                    grid=[8, 8, 7, 7, 24], background_index=8, items=[])
    if target.exists():
        existing = json.loads(target.read_text())
        assert all(existing[k] == v for k, v in manifest.items() if k != 'items')
        manifest = existing
    keys = {item['key'] for item in manifest['items']}
    assert len(keys) == len(manifest['items'])
    signatures = {item['indexed_sha256'] for item in manifest['items']}
    pending = []
    for item in request['items']:
        key = item['key']
        assert key == f"comspr-{item['block']:03d}" and item['block'] in (*range(22), *range(128, 150))
        assert item['source_file'] == 'COMSPR.DAX'
        assert item['source_file_sha256'] == 'f8b111aba66646aee47fbd0c5f20b82456c7b47208701d4f4c8a851c2a855cee'
        assert item['source_size'] == [24, 24] and key not in keys
        png, idx, pal = a.out / (key + '.png'), a.req / item['indexed'], a.req / item['palette']
        assert sha(png) == visual[key] == numeric[key]['png_sha256']
        assert sha(idx) == item['indexed_sha256'] and sha(pal) == item['palette_sha256']
        ref = idx.read_bytes()
        assert len(ref) == 576 and all(c <= 15 or 64 <= c < 96 for c in ref)
        assert hashlib.sha256(pal.read_bytes()[192:288]).hexdigest() == '903dcf3a1bc9aaa4b5b0829df3923b73fc4b8e7548ce33a84c8f154639d510a0'
        assert sha(idx) not in signatures
        signatures.add(sha(idx)); keys.add(key)
        manifest['items'].append(dict(key=key, size=[24, 24], png=png.name,
            png_sha256=sha(png), indexed=key + '.idx', indexed_sha256=sha(idx)))
        pending.extend(((png, png.name), (idx, key + '.idx')))
    assert 0 < len(manifest['items']) <= 44
    provenance_path = a.theme / 'vga-layers-provenance.json'
    provenance = json.loads(provenance_path.read_text())
    provenance['items'].append(dict(kind='combat-effects', request_sha256=sha(a.req/'manifest.json'),
        validation_sha256=sha(a.out/'validation.json'), visual_sha256=sha(a.out/'visual-review.json'),
        keys=[item['key'] for item in request['items']]))
    assert a.theme.is_dir() and a.theme.stat().st_uid == 1000
    for source, name in pending:
        assert not (a.theme / name).exists()
        shutil.copyfile(source, a.theme / name)
    target.write_text(json.dumps(manifest, indent=2) + '\n')
    provenance_path.write_text(json.dumps(provenance, indent=2) + '\n')
    print('VGA戰鬥特效候選主題：', len(manifest['items']))


if __name__ == '__main__':
    main()
