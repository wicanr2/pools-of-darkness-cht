#!/usr/bin/env python3
"""將固定SPRIT來源與驗收成品匯入既有encounters.json v1，續批保留已驗條目。"""
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
    ap.add_argument('--merge-existing', action='store_true')
    a = ap.parse_args()
    request = json.loads((a.req/'manifest.json').read_text())
    valid = json.loads((a.out/'validation.json').read_text())
    visual = json.loads((a.out/'visual-review.json').read_text())
    proof = json.loads(a.proof.read_text())
    assert valid['all_pass'] and valid['scope'] == 'all'
    assert valid['request_manifest_sha256'] == sha(a.req/'manifest.json')
    numeric = {r['key']: r for r in valid['items']}
    sources = {r['key']: r for r in proof['items']}
    allowed_blocks = {41} if proof['issue'] == 65 else {5,27,28,29,30,42,53,69} if proof['issue'] == 66 else set()
    assert allowed_blocks, '未知來源批次須先固定證據及規格'
    keys = {f'sprit1-{bid:03}@{f}' for bid in allowed_blocks for f in range(3)}
    assert set(sources) == set(numeric) == {r['key'] for r in request['items']} == keys
    original = 'df356577a5d884222138c486170caf0bb21809f58201ed52827b2cd9dc441c7f'
    assert proof['archive']['sha256'] == original and proof['position'] == [24, 24]
    if proof['issue'] == 65:
        assert proof['normal']['full_nonzero_source_equal'] and proof['normal']['used_rgb_equal']
    else:
        assert a.merge_existing, '續批必須保留已驗041條目'
        assert proof['prior_runtime_receipt_sha256'] == 'bc54803fe1e6d9646a009481571e41bb1f5bdc156b600b391504c7a0599bdd65'
        assert proof['new_keys'] and set(proof['new_keys']) == keys
        assert len(proof['complete_foreground_unambiguous']) == 27 and len(proof['diagnostic_cases']) == 12
    target = a.theme/'encounters.json'
    manifest = dict(version=1, scale=3, kind='encounter-sprites', items=[])
    if a.merge_existing:
        prior = json.loads(target.read_text())
        assert {k:v for k,v in prior.items() if k!='items'} == {k:v for k,v in manifest.items() if k!='items'}
        assert len(prior['items']) == 3 and {r['key'] for r in prior['items']} == {f'sprit1-041@{f}' for f in range(3)}
        manifest = prior
    else:
        assert not target.exists()
    previous = {r['key'] for r in manifest['items']}
    assert not(previous & keys)
    pending = []
    for item in request['items']:
        key = item['key']; s = sources[key]
        assert item['source_file'] == 'SPRIT1.DAX' and item['source_file_sha256'] == original
        assert item['block'] in allowed_blocks and key == f"sprit1-{item['block']:03}@{s.get('frame', key.rsplit('@',1)[1])}"
        assert item['source_size'] == [88, 88] and item['output_size'] == [264, 264]
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
