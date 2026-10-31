#!/usr/bin/env python3
"""b40正常DEMO增量抽驗：新程式配舊／新主題，固定b39終點不變。"""
import argparse
import copy
import hashlib
import json
from pathlib import Path
import sys

import numpy as np

sys.path.insert(0, str(Path(__file__).resolve().parent))
from vga_effect_play_check import read_case, tree


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verify(prior, before, after, previous, theme, source, binary_sha):
    for case in (before, after):
        for name in ('state', 'guest', 'idx', 'pal', 'original'):
            assert case[name] == prior[name], name + '不同'
        assert case['state']['frames'] == 7400 and case['state']['steps'] == 712250000
        assert case['flash'] == prior['flash'] == 0
        assert len(case['guest']) == 127
        assert case['frame'].shape == prior['frame'].shape == (600, 1440, 4)
        assert np.array_equal(case['frame'], prior['frame']), '正常合成像素不同'
        assert case['png_sha256'] == prior['png_sha256']
        for name in ('orig', 'font', 'text'):
            assert case['inputs'][name] == prior['inputs'][name], name + '輸入不同'
        assert case['inputs']['binary_sha256'] == binary_sha
        for name, field in (('idx', 'indexed'), ('pal', 'palette')):
            assert hashlib.sha256(case[name]).hexdigest() == case['state'][field]
    old, new = before['inputs']['theme'], after['inputs']['theme']
    assert old == prior['inputs']['theme'] and len(old) == 1331 and len(new) == 1350
    assert tree(previous) == old and tree(theme) == new
    assert 'exact.json' not in old
    manifest = json.loads((theme / 'exact.json').read_text())
    assert {k: v for k, v in manifest.items() if k != 'items'} == dict(version=1, scale=3, kind='complete-source-patterns')
    rows = {r['key']: r for r in manifest['items']}
    sources = {r['key']: r for r in source['items']}
    wanted = {f'back1-{n:03}' for n in range(1, 7)} | {f'skygrnd-{n:03}' for n in range(2, 5)}
    assert len(rows) == len(manifest['items']) == 9 and set(rows) == set(sources) == wanted
    added = {'exact.json'}
    for key, row in rows.items():
        assert row['size'] == [88, 88] and not row.get('algorithm')
        assert row['indexed_sha256'] == sources[key]['indexed_sha256']
        assert row['png'] == key+'.png' and row['indexed'] == key+'.idx'
        for name, digest in ((row['png'], row['png_sha256']), (row['indexed'], row['indexed_sha256'])):
            assert new[name] == digest
            added.add(name)
    assert set(new) - set(old) == added
    assert all(new[k] == v for k, v in old.items() if k != 'vga-layers-provenance.json'), '既有素材改變'
    return dict(status='PASS', scope='九個背景條件式來源；正常冷啟動DEMO增量抽驗，不外推正常背景事件或遮擋',
                binary_sha256=binary_sha, prior_binary_sha256=prior['inputs']['binary_sha256'],
                frames=7400, steps=712250000, state=after['state'], guest_files=127,
                guest_state_equal=True, original_endpoint_equal=True,
                old_theme_new_program_composite_equal=True, final_composite_equal=True,
                prior_assets_preserved=True, theme_files=1350, background_candidates=9, flash_count=0,
                normal_background_events_observed=False,
                frames_sha256=dict(prior=prior['png_sha256'], before=before['png_sha256'], after=after['png_sha256']))


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for name in ('prior', 'before', 'after', 'theme', 'previous', 'source-proof', 'baseline-receipt', 'binary', 'report'):
        ap.add_argument('--'+name, type=Path, required=True)
    a = ap.parse_args()
    assert sha(a.baseline_receipt) == '2eaad8797fdb852bf0ed45a1767f1bc0b4956ea8a7440b85b561cba9c96dac21'
    baseline, source = json.loads(a.baseline_receipt.read_text()), json.loads(a.source_proof.read_text())
    prior, before, after = [read_case(p) for p in (a.prior, a.before, a.after)]
    assert prior['state'] == baseline['runtime']['state']
    assert prior['png_sha256'] == baseline['runtime']['frames_sha256']['after']
    assert sha(a.prior/'run.log') == baseline['runtime']['runtime_log_sha256']
    assert sha(a.prior/'inputs.json') == baseline['runtime']['runtime_inputs_sha256']
    binary_sha = sha(a.binary)
    report = verify(prior, before, after, a.previous, a.theme, source, binary_sha)
    controls = []
    for kind in ('cpu', 'source', 'palette', 'guest', 'output-pixel'):
        bad = copy.deepcopy(after)
        if kind == 'cpu': bad['state']['cpu'] = '0'*64
        elif kind == 'source': bad['idx'] = bytes([bad['idx'][0]^1]) + bad['idx'][1:]
        elif kind == 'palette': bad['pal'] = bytes([bad['pal'][0]^1]) + bad['pal'][1:]
        elif kind == 'guest': bad['guest'][next(iter(bad['guest']))] = '0'*64
        else: bad['frame'][0, 0, 0] ^= 1
        try:
            verify(prior, before, bad, a.previous, a.theme, source, binary_sha)
        except AssertionError:
            controls.append(dict(case=kind, rejected=True))
        else:
            raise AssertionError('負對照未拒絕：'+kind)
    report.update(negative_controls=controls, negative_controls_all_pass=True,
                  source_proof_sha256=sha(a.source_proof), baseline_receipt_sha256=sha(a.baseline_receipt),
                  runtime_inputs_sha256=sha(a.after/'inputs.json'), runtime_log_sha256=sha(a.after/'run.log'),
                  old_theme_runtime_inputs_sha256=sha(a.before/'inputs.json'), old_theme_runtime_log_sha256=sha(a.before/'run.log'))
    a.report.write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(report, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    main()
