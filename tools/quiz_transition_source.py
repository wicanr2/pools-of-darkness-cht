#!/usr/bin/env python3
"""固定正常查驗題轉場的來源、原始呼叫與量尺正對照。"""
import argparse
import csv
import hashlib
import json
from pathlib import Path


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--source', type=Path, required=True)
    ap.add_argument('--legacy', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    a = ap.parse_args()
    sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
    rows = list(csv.DictReader((a.source/'inv.tsv').open(), delimiter='\t'))
    calls = [r for r in rows if r['path'] == 'A' and 337066144 <= int(r['step']) <= 337122418]
    expected = [' journal, page ', ' (1)         ', ' what is word # ', '1     ',
                ' after the heading ', 'CHARACTERS AND PARTIES']
    assert [r['text'] for r in calls] == expected
    log = a.legacy.read_text()
    assert 'pod-play: 閃現 9 次' in log
    files = [a.source/'inputs.json', a.source/'inv.tsv', a.source/'run.log', a.legacy]
    for f in range(3503, 3507):
        for suffix in ('idx', 'pal'):
            p = a.source/'sources'/f'pod-f{f:06d}.{suffix}'
            assert p.stat().st_size == (64000 if suffix == 'idx' else 768)
            files.append(p)
        files.append(a.source/'shots'/f'pod-f{f:06d}.png')
    report = dict(scope='既有正常讀槽A查驗題；F10檔名減一為呈現格；原版CPU照常執行',
                  evidence_level='已證實', tools='dosgolem b47a617; Go1.24.13; Python3.13',
                  address_space='dosgolem解壓後實模式CS:IP；overlay段隨載入改變',
                  original_inputs=json.loads((a.source/'inputs.json').read_text())['orig'],
                  calls=calls, legacy_flash_count=9, shown_frame=3503, replaced_frame=3504,
                  cells=[[18,2],[19,2],[20,2],[4,4],[5,4],[6,4],[7,4],[9,4],[10,4]],
                  source_files={str(p):sha(p) for p in files},
                  limitation='逐格覆繪診斷不取代正常4800格驗收，亦不外推其他題目路徑')
    a.out.write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n')
    print('查驗題六段及9格原始正對照已固定。')


if __name__ == '__main__':
    main()
