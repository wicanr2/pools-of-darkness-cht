#!/usr/bin/env python3
"""將b40固定來源與已驗收成品匯入既有exact.json v1，只寫獨立候選。"""
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
        ap.add_argument('--' + name, type=Path, required=True)
    a = ap.parse_args()
    request = json.loads((a.req / 'manifest.json').read_text())
    valid = json.loads((a.out / 'validation.json').read_text())
    visual = json.loads((a.out / 'visual-review.json').read_text())
    proof = json.loads(a.proof.read_text())
    assert valid['all_pass'] and valid['scope'] == 'all'
    assert valid['request_manifest_sha256'] == sha(a.req / 'manifest.json')
    numeric = {row['key']: row for row in valid['items']}
    sources = {row['key']: row for row in proof['items']}
    keys = {f'back1-{n:03}' for n in range(1, 7)} | {f'skygrnd-{n:03}' for n in range(2, 5)}
    assert set(sources) == {row['key'] for row in request['items']} == set(numeric) == keys
    assert proof['archives']['BACK1.DAX']['sha256'] == '2efc68c5220a3f2c00315accc4b13961727cf87ae941484a8eb77cb78e054c6f'
    assert proof['archives']['SKYGRND.DAX']['sha256'] == 'e592a635c3f91615f61efeaae729bb1aed57243dcb3f9df78412b80505ea17b6'
    target = a.theme / 'exact.json'
    assert not target.exists(), '本批由無exact.json的1331檔基準開始'
    manifest = dict(version=1, scale=3, kind='complete-source-patterns', items=[])
    pending, signatures = [], set()
    for item in request['items']:
        key = item['key']
        s = sources[key]
        assert key == f"{item['source_file'].removesuffix('.DAX').lower()}-{item['block']:03}"
        assert item['source_file_sha256'] == proof['archives'][item['source_file']]['sha256']
        assert item['source_size'] == [88, 88] and item['output_size'] == [264, 264]
        png, idx, pal = a.out / (key + '.png'), a.req / item['indexed'], a.req / item['palette']
        assert sha(png) == visual[key] == numeric[key]['png_sha256']
        assert sha(idx) == item['indexed_sha256'] == s['indexed_sha256']
        assert sha(pal) == item['palette_sha256']
        ref, rgb = idx.read_bytes(), pal.read_bytes()
        assert len(ref) == 88 * 88 and len(rgb) == 768 and all(144 <= v < 176 for v in ref)
        used = sorted(set(ref))
        signature = hashlib.sha256(b''.join(bytes([v]) + rgb[v*3:v*3+3] for v in used)).hexdigest()
        assert used == s['used_colors'] and signature == s['used_rgb_sha256']
        sig = (sha(idx), signature)
        assert sig not in signatures
        signatures.add(sig)
        manifest['items'].append(dict(key=key, size=[88, 88], png=png.name,
            png_sha256=sha(png), indexed=key + '.idx', indexed_sha256=sha(idx)))
        pending.extend(((png, png.name), (idx, key + '.idx')))
    provenance_path = a.theme / 'vga-layers-provenance.json'
    provenance = json.loads(provenance_path.read_text())
    provenance['items'].append(dict(kind='complete-source-backgrounds', request_sha256=sha(a.req/'manifest.json'),
        validation_sha256=sha(a.out/'validation.json'), visual_sha256=sha(a.out/'visual-review.json'),
        source_proof_sha256=sha(a.proof), keys=sorted(keys)))
    assert a.theme.is_dir() and a.theme.stat().st_uid == 1000
    assert all(not (a.theme/name).exists() for _, name in pending)
    for source, name in pending:
        shutil.copyfile(source, a.theme / name)
    target.write_text(json.dumps(manifest, indent=2) + '\n')
    provenance_path.write_text(json.dumps(provenance, indent=2) + '\n')
    print('完整VGA背景候選：', len(manifest['items']))


if __name__ == '__main__':
    main()
