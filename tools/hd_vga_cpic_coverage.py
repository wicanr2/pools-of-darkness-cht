#!/usr/bin/env python3
"""核對CPIC來源與已驗收候選的唯一圖樣覆蓋；不代表正常遭遇驗收。"""
import argparse
from collections import defaultdict
import hashlib
import json
from pathlib import Path
import struct
import sys

from PIL import Image


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def coverage(inventory, theme, original=None):
    catalog = inventory / 'inventory-other.json'
    manifest = theme / 'actors.json'
    groups, sources = defaultdict(list), {}
    for row in json.loads(catalog.read_text())['items']:
        if row['file'] != 'CPIC1.DAX':
            continue
        key = row['key']
        assert key not in sources, f'{key}: 盤點key重複'
        idxpath, palpath = (inventory / (key + ext) for ext in ('.idx', '.pal'))
        assert sha(idxpath) == row['indexed_sha256'], f'{key}: 盤點色號變更'
        assert sha(palpath) == row['palette_sha256'], f'{key}: 盤點色盤變更'
        idx, pal = idxpath.read_bytes(), palpath.read_bytes()
        assert len(idx) == row['width'] * row['height'] and len(pal) == 768
        used = bytes(v for c in sorted(set(idx)) for v in pal[c * 3:c * 3 + 3])
        signature = (row['width'], row['height'], row['indexed_sha256'],
                     hashlib.sha256(used).hexdigest())
        sources[key] = (row, signature)
        groups[signature].append(key)

    data = json.loads(manifest.read_text())
    assert data['kind'] == 'combat-actors' and data['scale'] == 3
    accepted, seen = defaultdict(list), set()
    for actor in data['items']:
        key = actor['key']
        assert key not in seen and key in sources, f'{key}: 主題key重複或來源未知'
        seen.add(key)
        row, signature = sources[key]
        assert actor['size'] == [row['width'], row['height']], f'{key}: 尺寸不符'
        assert actor['indexed_sha256'] == row['indexed_sha256'], f'{key}: 色號來源不符'
        assert sha(theme / actor['indexed']) == actor['indexed_sha256'], f'{key}: 主題色號變更'
        assert sha(theme / actor['png']) == actor['png_sha256'], f'{key}: 主題PNG變更'
        with Image.open(theme / actor['png']) as image:
            assert image.size == (row['width'] * 3, row['height'] * 3) and image.mode == 'RGBA'
        accepted[signature].append(dict(key=key, png_sha256=actor['png_sha256']))

    directory = None
    if original is not None:
        archive = original / 'CPIC1.DAX'
        expected = {row['archive_sha256'] for row, _ in sources.values()}
        assert len(expected) == 1 and sha(archive) in expected, '原版CPIC壓縮檔變更'
        raw = archive.read_bytes()
        end = struct.unpack_from('<H', raw)[0] + 2
        assert 2 <= end <= len(raw) and (end - 2) % 9 == 0
        entries = []
        for offset in range(2, end, 9):
            block, relative, word1, word2 = struct.unpack_from('<BIHH', raw, offset)
            entries.append(dict(block=block, directory_entry_file_offset=offset,
                                payload_file_offset=end + relative, raw_words=[word1, word2]))
        ids = [entry['block'] for entry in entries]
        assert len(set(ids)) == len(ids) and set(ids) == {row['block'] for row, _ in sources.values()}
        directory = dict(file=archive.name, sha256=sha(archive), address_space='original archive file byte offsets',
                         directory_end_file_offset=end, entry_size=9, blocks=sorted(ids), entries=entries)
    rows = [dict(size=list(sig[:2]), indexed_sha256=sig[2], used_palette_sha256=sig[3],
                 sources=sorted(keys), candidates=accepted.get(sig, []))
            for sig, keys in groups.items()]
    missing = [row for row in rows if not row['candidates']]
    return dict(scope='exact size/indexed/used-RGB candidate source coverage only; no mirror deduplication or normal-encounter claim',
                producer=dict(script='tools/hd_vga_cpic_coverage.py', sha256=sha(Path(__file__)), python=sys.version.split()[0]),
                inputs=dict(catalog_sha256=sha(catalog), actors_manifest_sha256=sha(manifest)),
                source_states=len(sources), unique_signatures=len(groups),
                registered_actor_states=len(seen), covered_signatures=len(groups) - len(missing),
                covered_source_states=sum(len(row['sources']) for row in rows if row['candidates']),
                missing_unique=len(missing), missing_sources=[key for row in missing for key in row['sources']],
                aliases=[row['sources'] for row in rows if len(row['sources']) > 1],
                original_directory=directory, groups=rows)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    for arg in ('inventory', 'theme', 'out'):
        parser.add_argument('--' + arg, type=Path, required=True)
    parser.add_argument('--require-complete', action='store_true')
    parser.add_argument('--original', type=Path, help='選填：唯讀原版目錄，核對CPIC目錄全部實際ID')
    args = parser.parse_args()
    report = coverage(args.inventory, args.theme, args.original)
    args.out.write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps({key: value for key, value in report.items() if key not in ('groups', 'original_directory')}, ensure_ascii=False, indent=2))
    raise SystemExit(1 if args.require_complete and report['missing_unique'] else 0)
