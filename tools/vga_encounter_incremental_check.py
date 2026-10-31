#!/usr/bin/env python3
"""b42正常DEMO增量抽驗：固定b41、同程式舊新主題及十四張既有截圖。"""
import argparse
import copy
import hashlib
import json
from pathlib import Path
import sys

import numpy as np
from PIL import Image

sys.path.insert(0, str(Path(__file__).resolve().parent))
from vga_effect_play_check import read_case, tree


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verify(prior, before, after, previous, theme, source, binary_sha):
    for case in (before, after):
        for name in ('state', 'guest', 'idx', 'pal', 'original'):
            assert case[name] == prior[name], name+'不同'
        assert case['state']['frames'] == 7400 and case['state']['steps'] == 712250000
        assert case['flash'] == prior['flash'] == 0 and len(case['guest']) == 127
        assert case['frame'].shape == prior['frame'].shape == (600, 1440, 4)
        assert np.array_equal(case['frame'], prior['frame']) and case['png_sha256'] == prior['png_sha256']
        for name in ('orig', 'font', 'text'):
            assert case['inputs'][name] == prior['inputs'][name], name+'輸入不同'
        assert case['inputs']['binary_sha256'] == binary_sha
        assert hashlib.sha256(case['idx']).hexdigest() == case['state']['indexed']
        assert hashlib.sha256(case['pal']).hexdigest() == case['state']['palette']
    old, new = before['inputs']['theme'], after['inputs']['theme']
    assert old == prior['inputs']['theme'] and len(old) == 1357 and len(new) == 1405
    assert tree(previous) == old and tree(theme) == new
    manifests = [json.loads((p/'encounters.json').read_text()) for p in (previous, theme)]
    a, b = [{row['key']:row for row in m['items']} for m in manifests]
    assert all({k:v for k,v in m.items() if k!='items'} == dict(version=1, scale=3, kind='encounter-sprites') for m in manifests)
    assert len(a) == 3 and len(b) == len(manifests[1]['items']) == 27
    assert all(b.get(k) == v for k,v in a.items()), '既有041條目改變'
    sources = {row['key']:row for row in source['items']}
    assert set(b)-set(a) == set(source['new_keys']) == set(sources) and len(sources) == 24
    added = set()
    for key,s in sources.items():
        row = b[key]
        assert row['size'] == [88,88] and row['position'] == [24,24]
        assert row['indexed_sha256'] == s['indexed_sha256']
        assert row['png'] == key+'.png' and row['indexed'] == key+'.idx'
        for name,digest in ((row['png'],row['png_sha256']),(row['indexed'],row['indexed_sha256'])):
            assert new[name] == digest
            added.add(name)
    assert set(new)-set(old) == added
    assert all(new[k]==v for k,v in old.items() if k not in ('encounters.json','vga-layers-provenance.json')), '既有素材改變'
    return dict(status='PASS', scope='八組24張來源素材；六類診斷另計，正常DEMO只驗既有十四張及終點不變，新正常事件未知',
                binary_sha256=binary_sha, prior_binary_sha256=prior['inputs']['binary_sha256'],
                frames=7400, steps=712250000, state=after['state'], guest_files=127,
                guest_state_equal=True, original_endpoint_equal=True, old_theme_new_program_composite_equal=True,
                final_composite_equal=True, prior_assets_preserved=True, theme_files=1405, encounter_candidates=27, flash_count=0,
                frames_sha256=dict(prior=prior['png_sha256'],before=before['png_sha256'],after=after['png_sha256']))


def verify_shot(idx, pal, png, row, refs):
    assert hashlib.sha256(idx).hexdigest() == row['indexed_sha256']
    assert hashlib.sha256(pal).hexdigest() == row['palette_sha256']
    assert hashlib.sha256(png).hexdigest() == row['after_png_sha256']
    assert len(idx) == 64000 and len(pal) == 768
    screen = np.frombuffer(idx,np.uint8).reshape(200,320)[24:112,24:112]
    matches = []
    for key,ref,signature in refs:
        if not np.array_equal(screen[ref!=0],ref[ref!=0]):
            continue
        used = sorted(set(ref.tobytes())-{0})
        actual = hashlib.sha256(b''.join(bytes([v])+pal[v*3:v*3+3] for v in used)).hexdigest()
        if actual == signature:
            matches.append(key)
    assert not matches, '預定未觀測的新來源出現在抽樣；須回查證據'


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for name in ('prior','before','after','theme','previous','source-proof','baseline-receipt','binary','report'):
        ap.add_argument('--'+name,type=Path,required=True)
    a = ap.parse_args()
    assert sha(a.baseline_receipt) == 'bc54803fe1e6d9646a009481571e41bb1f5bdc156b600b391504c7a0599bdd65'
    baseline = json.loads(a.baseline_receipt.read_text())
    source = json.loads(a.source_proof.read_text())
    prior,before,after = [read_case(p) for p in (a.prior,a.before,a.after)]
    assert prior['state'] == baseline['runtime']['state']
    assert prior['png_sha256'] == baseline['runtime']['frames_sha256']['after']
    assert sha(a.prior/'run.log') == baseline['runtime']['runtime_log_sha256']
    assert sha(a.prior/'inputs.json') == baseline['runtime']['runtime_inputs_sha256']
    binary_sha = sha(a.binary)
    report = verify(prior,before,after,a.previous,a.theme,source,binary_sha)
    controls = []
    for kind in ('cpu','source','palette','guest','output-pixel'):
        bad = copy.deepcopy(after)
        if kind == 'cpu': bad['state']['cpu'] = '0'*64
        elif kind == 'source': bad['idx'] = bytes([bad['idx'][0]^1])+bad['idx'][1:]
        elif kind == 'palette': bad['pal'] = bytes([bad['pal'][0]^1])+bad['pal'][1:]
        elif kind == 'guest': bad['guest'][next(iter(bad['guest']))] = '0'*64
        else: bad['frame'][0,0,0] ^= 1
        try: verify(prior,before,bad,a.previous,a.theme,source,binary_sha)
        except AssertionError: controls.append(dict(case=kind,rejected=True))
        else: raise AssertionError('負對照未拒絕：'+kind)
    expected = baseline['runtime']['normal_shots']
    assert len(expected) == 14
    refs = [(s['key'],np.frombuffer((a.theme/(s['key']+'.idx')).read_bytes(),np.uint8).reshape(88,88),s['used_rgb_sha256']) for s in source['items']]
    shots = []
    for path in (a.prior,a.before,a.after):
        assert sorted(p.stem for p in (path/'sources').glob('*.idx')) == sorted(row['capture'] for row in expected)
    for row in expected:
        capture = row['capture']
        for path in (a.prior,a.before,a.after):
            idx = (path/'sources'/(capture+'.idx')).read_bytes()
            pal = (path/'sources'/(capture+'.pal')).read_bytes()
            png = (path/'window-shots'/(capture+'.png')).read_bytes()
            verify_shot(idx,pal,png,row,refs)
        shots.append(dict(capture=capture,indexed_sha256=row['indexed_sha256'],palette_sha256=row['palette_sha256'],
                          before_png_sha256=row['after_png_sha256'],after_png_sha256=row['after_png_sha256'],
                          equal_to_b41=True,source_frame_offset=-1,new_sources_observed=[],changed_pixels=0,
                          prior_distance=row['distance']))
    for kind in ('shot-source','shot-palette','shot-png'):
        fields = [idx,pal,png]
        n = ('shot-source','shot-palette','shot-png').index(kind)
        fields[n] = bytes([fields[n][0]^1])+fields[n][1:]
        try: verify_shot(*fields,row,refs)
        except AssertionError: controls.append(dict(case=kind,rejected=True))
        else: raise AssertionError('截圖負對照未拒絕：'+kind)
    report.update(negative_controls=controls,negative_controls_all_pass=True,normal_shots=shots,
                  normal_all_14_equal_to_b41=True,new_normal_events_observed=[],unobserved_new_normal_keys=source['new_keys'],
                  source_proof_sha256=sha(a.source_proof),baseline_receipt_sha256=sha(a.baseline_receipt),
                  runtime_inputs_sha256=sha(a.after/'inputs.json'),runtime_log_sha256=sha(a.after/'run.log'),
                  old_theme_runtime_inputs_sha256=sha(a.before/'inputs.json'),old_theme_runtime_log_sha256=sha(a.before/'run.log'))
    a.report.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(dict(status=report['status'],files=1405,shots=14,flash=0,negative_controls=len(controls)),indent=2))


if __name__ == '__main__':
    main()
