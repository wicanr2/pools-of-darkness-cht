#!/usr/bin/env python3
"""COMSPR完整素材的正常DEMO增量抽驗；正式程式與既有終點須相同。"""
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


def verify(before, after, theme, previous, source):
    for name in ('state', 'guest', 'idx', 'pal', 'original'):
        assert after[name] == before[name], name + '不同'
    assert after['state']['frames'] == 7400 and after['state']['steps'] == 712250000
    assert before['flash'] == after['flash'] == 0
    assert len(after['guest']) == 127
    assert before['frame'].shape == after['frame'].shape == (600, 1440, 4)
    assert np.array_equal(before['frame'], after['frame']), '正常合成像素不同'
    assert before['png_sha256'] == after['png_sha256']
    for name in ('orig', 'font', 'text', 'binary_sha256'):
        assert before['inputs'][name] == after['inputs'][name], name + '輸入不同'
    assert after['inputs']['binary_sha256'] == 'ff54dfe2fa2fc0e4462c14718f7e85dbddc3043b5259b1559f1cf97ed9cea042'
    for name, field in (('idx', 'indexed'), ('pal', 'palette')):
        assert hashlib.sha256(after[name]).hexdigest() == after['state'][field]
    old, new = before['inputs']['theme'], after['inputs']['theme']
    assert len(old) == 1271 and len(new) == 1331
    assert tree(previous) == old and tree(theme) == new
    manifests = [json.loads((p/'effects.json').read_text()) for p in (previous, theme)]
    prior, current = manifests
    assert {k: v for k, v in prior.items() if k != 'items'} == {k: v for k, v in current.items() if k != 'items'}
    a, b = [{r['key']: r for r in m['items']} for m in manifests]
    assert len(a) == 6 and len(b) == 36
    assert all(b.get(key) == value for key, value in a.items()), '既有特效條目改變'
    assert set(b) - set(a) == set(source['new_keys']) and len(source['new_keys']) == 30
    assert list(b) == source['canonical_keys']
    added = set()
    for key in source['new_keys']:
        row = b[key]
        assert row['size'] == [24, 24]
        for name, digest in ((row['png'], row['png_sha256']), (row['indexed'], row['indexed_sha256'])):
            assert new[name] == digest
            added.add(name)
    assert set(new) - set(old) == added
    assert all(new[key] == digest for key, digest in old.items() if key not in ('effects.json', 'vga-layers-provenance.json')), '既有素材改變'
    ids = set(range(22)) | set(range(128, 150))
    assert len(source['coverage']) == 44 and {int(r['key'][-3:]) for r in source['coverage']} == ids
    assert len({r['key'] for r in source['coverage']}) == 44
    aliases = []
    for row in source['coverage']:
        ref = (theme / b[row['canonical']]['indexed']).read_bytes()
        assert len(ref) == 576
        if row['flip']:
            ref = b''.join(ref[y*24:(y+1)*24][::-1] for y in range(24))
        assert hashlib.sha256(ref).hexdigest() == row['indexed_sha256'], '來源別名不同'
        assert row['source_bank_sha256'] == '903dcf3a1bc9aaa4b5b0829df3923b73fc4b8e7548ce33a84c8f154639d510a0'
        if row['key'] != row['canonical']:
            aliases.append(row)
    assert len(aliases) == 8 and sum(r['flip'] for r in aliases) == 2
    return dict(status='PASS', scope='完整COMSPR素材；正常冷啟動DEMO增量抽一條，不外推未觀測施法事件',
                binary_sha256=after['inputs']['binary_sha256'], unchanged_program=True,
                frames=7400, steps=712250000, state=after['state'], guest_files=127,
                guest_state_equal=True, original_endpoint_equal=True, final_composite_equal=True,
                prior_assets_preserved=True, theme_files=1331, flash_count=0,
                effect_candidates=36, source_states=44, direct_aliases=6, mirror_aliases=2,
                sampled_keys=source['sampled_keys'],
                unexercised_new_keys=sorted(set(source['new_keys'])-set(source['sampled_keys'])),
                frames_sha256=dict(before=before['png_sha256'], after=after['png_sha256']))


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for name in ('before', 'after', 'theme', 'previous', 'source-proof', 'baseline-receipt', 'report'):
        ap.add_argument('--'+name, type=Path, required=True)
    ap.add_argument('--root', type=Path, default=Path('.'))
    a = ap.parse_args()
    source = json.loads(a.source_proof.read_text())
    baseline = json.loads(a.baseline_receipt.read_text())
    before, after = read_case(a.before), read_case(a.after)
    assert before['state'] == baseline['runtime']['state']
    assert before['png_sha256'] == baseline['runtime']['frames_sha256']['after']
    assert sha(a.before/'run.log') == baseline['runtime_log_sha256']
    assert sha(a.before/'inputs.json') == baseline['runtime_inputs_sha256']
    lock = json.loads((a.root/'review/hd/runtime/manifest.json').read_text())
    assert before['guest'] == lock['runtime_expected']['guest']
    report = verify(before, after, a.theme, a.previous, source)
    controls = []
    for kind in ('cpu', 'source', 'palette', 'guest', 'output-pixel'):
        bad = copy.deepcopy(after)
        if kind == 'cpu': bad['state']['cpu'] = '0'*64
        elif kind == 'source': bad['idx'] = bytes([bad['idx'][0]^1]) + bad['idx'][1:]
        elif kind == 'palette': bad['pal'] = bytes([bad['pal'][0]^1]) + bad['pal'][1:]
        elif kind == 'guest': bad['guest'][next(iter(bad['guest']))] = '0'*64
        else: bad['frame'][0, 0, 0] ^= 1
        try:
            verify(before, bad, a.theme, a.previous, source)
        except AssertionError:
            controls.append(dict(case=kind, rejected=True))
        else:
            raise AssertionError('負對照未拒絕：'+kind)
    report.update(negative_controls=controls, negative_controls_all_pass=True,
                  source_proof_sha256=sha(a.source_proof), baseline_receipt_sha256=sha(a.baseline_receipt),
                  runtime_inputs_sha256=sha(a.after/'inputs.json'), runtime_log_sha256=sha(a.after/'run.log'))
    a.report.write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(report, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    main()
