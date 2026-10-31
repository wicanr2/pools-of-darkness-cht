#!/usr/bin/env python3
"""將b41固定SPRIT041來源與驗收成品匯入既有encounters.json v1。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for name in ('req', 'out', 'theme', 'proof'):
        ap.add_argument('--'+name, type=Path, required=True)
    a = ap.parse_args()
    request = json.loads((a.req/'manifest.json').read_text())
    valid = json.loads((a.out/'validation.json').read_text())
    visual = json.loads((a.out/'visual-review.json').read_text())
    proof = json.loads(a.proof.read_text())
    assert valid['all_pass'] and valid['scope'] == 'all'
    assert valid['request_manifest_sha256'] == sha(a.req/'manifest.json')
    numeric = {r['key']: r for r in valid['items']}
    sources = {r['key']: r for r in proof['items']}
    keys = {f'sprit1-041@{f}' for f in range(3)}
    assert set(sources) == set(numeric) == {r['key'] for r in request['items']} == keys
    original = 'df356577a5d884222138c486170caf0bb21809f58201ed52827b2cd9dc441c7f'
    assert proof['archive']['sha256'] == original and proof['position'] == [24, 24]
    assert proof['normal']['full_nonzero_source_equal'] and proof['normal']['used_rgb_equal']
    target = a.theme/'encounters.json'
    assert not target.exists(), '本批由無encounters.json的1350檔主題開始'
    manifest = dict(version=1, scale=3, kind='encounter-sprites', items=[])
    pending = []
    for item in request['items']:
        key = item['key']; s = sources[key]
        assert item['source_file'] == 'SPRIT1.DAX' and item['source_file_sha256'] == original
        assert item['block'] == 41 and item['source_size'] == [88, 88] and item['output_size'] == [264, 264]
        png, idx, pal = a.out/(key+'.png'), a.req/item['indexed'], a.req/item['palette']
        assert sha(png) == visual[key] == numeric[key]['png_sha256']
        assert sha(idx) == item['indexed_sha256'] == s['indexed_sha256'] and sha(pal) == item['palette_sha256']
        ref, rgb = idx.read_bytes(), pal.read_bytes()
        assert len(ref) == 7744 and len(rgb) == 768 and all(v == 0 or v >= 176 for v in ref)
        used = sorted(set(ref)-{0})
        signature = hashlib.sha256(b''.join(bytes([v])+rgb[v*3:v*3+3] for v in used)).hexdigest()
        assert used == s['used_colors'] and signature == s['used_rgb_sha256']
        manifest['items'].append(dict(key=key, size=[88, 88], position=[24, 24], png=png.name,
            png_sha256=sha(png), indexed=key+'.idx', indexed_sha256=sha(idx)))
        pending.extend(((png, png.name), (idx, key+'.idx')))
    provenance_path = a.theme/'vga-layers-provenance.json'
    provenance = json.loads(provenance_path.read_text())
    provenance['items'].append(dict(kind='encounter-sprites', request_sha256=sha(a.req/'manifest.json'),
        validation_sha256=sha(a.out/'validation.json'), visual_sha256=sha(a.out/'visual-review.json'),
        source_proof_sha256=sha(a.proof), keys=sorted(keys)))
    assert a.theme.is_dir() and a.theme.stat().st_uid == 1000
    assert all(not(a.theme/name).exists() for _, name in pending)
    for source, name in pending:
        shutil.copyfile(source, a.theme/name)
    target.write_text(json.dumps(manifest, indent=2)+'\n')
    provenance_path.write_text(json.dumps(provenance, indent=2)+'\n')
    print('VGA遭遇圖候選：', len(manifest['items']))


if __name__ == '__main__':
    main()
