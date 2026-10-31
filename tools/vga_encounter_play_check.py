#!/usr/bin/env python3
"""b41正常遭遇圖抽驗：新程式／舊新主題，同狀態、完整前景及透明區。"""
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
    assert old == prior['inputs']['theme'] and len(old) == 1350 and len(new) == 1357
    assert tree(previous) == old and tree(theme) == new and 'encounters.json' not in old
    manifest = json.loads((theme/'encounters.json').read_text())
    assert {k:v for k,v in manifest.items() if k!='items'} == dict(version=1, scale=3, kind='encounter-sprites')
    rows = {r['key']:r for r in manifest['items']};sources = {r['key']:r for r in source['items']}
    assert len(rows) == len(manifest['items']) == 3 and set(rows) == set(sources) == {f'sprit1-041@{f}' for f in range(3)}
    added = {'encounters.json'}
    for key, row in rows.items():
        assert row['size'] == [88,88] and row['position'] == [24,24]
        assert row['indexed_sha256'] == sources[key]['indexed_sha256']
        assert row['png'] == key+'.png' and row['indexed'] == key+'.idx'
        for name,digest in ((row['png'],row['png_sha256']),(row['indexed'],row['indexed_sha256'])):
            assert new[name] == digest
            added.add(name)
    assert set(new)-set(old) == added
    assert all(new[k]==v for k,v in old.items() if k!='vga-layers-provenance.json'), '既有素材改變'
    return dict(status='PASS', scope='SPRIT041三距離來源；正常DEMO增量及近遠景，中景正常事件未知',
                binary_sha256=binary_sha, prior_binary_sha256=prior['inputs']['binary_sha256'],
                frames=7400,steps=712250000,state=after['state'],guest_files=127,
                guest_state_equal=True,original_endpoint_equal=True,old_theme_new_program_composite_equal=True,
                final_composite_equal=True,prior_assets_preserved=True,theme_files=1357,encounter_candidates=3,flash_count=0,
                frames_sha256=dict(prior=prior['png_sha256'],before=before['png_sha256'],after=after['png_sha256']))


def verify_shot(idx, rgb, before, after, refs, images, expected_rgb):
    assert len(idx)==64000 and len(rgb)==768
    screen=np.frombuffer(idx,np.uint8).reshape(200,320)
    pal=np.frombuffer(rgb,np.uint8).reshape(256,3)
    mask=np.zeros((600,1440),bool)
    assert before.shape==after.shape==(600,1440,4)
    diff=np.any(before!=after,axis=2)
    matches=[f for f,ref in enumerate(refs) if np.all(screen[24:112,24:112][ref!=0]==ref[ref!=0])]
    assert len(matches)<=1, '正常截圖來源有歧義'
    signature=None
    if matches:
        f=matches[0];ref=refs[f];foreground=ref!=0
        mask[72:336,312:576]=foreground.repeat(3,0).repeat(3,1)
        used=sorted(set(ref.tobytes())-{0})
        signature=hashlib.sha256(b''.join(bytes([v])+bytes(pal[v]) for v in used)).hexdigest()
        assert signature==expected_rgb[f], '正常遭遇圖使用RGB不同'
        assert np.array_equal(after[72:336,312:576][mask[72:336,312:576]],images[f][mask[72:336,312:576]]), '正常遭遇圖完整成品前景不同'
        assert np.any(diff), '正常遭遇圖沒有實際HD變化'
    assert not np.any(diff & ~mask), '正常截圖改動透明區或框外'
    return dict(full_source=bool(matches),distance=matches[0] if matches else None,used_rgb_sha256=signature,changed_pixels=int(np.count_nonzero(diff)),
                transparent_and_outside_changes=0)


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for name in ('prior','before','after','theme','previous','source-proof','baseline-receipt','binary','report'):
        ap.add_argument('--'+name,type=Path,required=True)
    a=ap.parse_args()
    assert sha(a.baseline_receipt)=='b2cf577b183b5edfa478942a342d4689b76b1522dd0adfaa6f026d5e89c75e51'
    baseline=json.loads(a.baseline_receipt.read_text());source=json.loads(a.source_proof.read_text())
    prior,before,after=[read_case(p) for p in (a.prior,a.before,a.after)]
    assert prior['state']==baseline['runtime']['state']
    assert prior['png_sha256']==baseline['runtime']['frames_sha256']['after']
    assert sha(a.prior/'run.log')==baseline['runtime']['runtime_log_sha256']
    assert sha(a.prior/'inputs.json')==baseline['runtime']['runtime_inputs_sha256']
    binary_sha=sha(a.binary)
    report=verify(prior,before,after,a.previous,a.theme,source,binary_sha)
    controls=[]
    for kind in ('cpu','source','palette','guest','output-pixel'):
        bad=copy.deepcopy(after)
        if kind=='cpu':bad['state']['cpu']='0'*64
        elif kind=='source':bad['idx']=bytes([bad['idx'][0]^1])+bad['idx'][1:]
        elif kind=='palette':bad['pal']=bytes([bad['pal'][0]^1])+bad['pal'][1:]
        elif kind=='guest':bad['guest'][next(iter(bad['guest']))]='0'*64
        else:bad['frame'][0,0,0]^=1
        try:verify(prior,before,bad,a.previous,a.theme,source,binary_sha)
        except AssertionError:controls.append(dict(case=kind,rejected=True))
        else:raise AssertionError('負對照未拒絕：'+kind)
    keys=[f'sprit1-041@{f}' for f in range(3)]
    refs=[np.frombuffer((a.theme/(key+'.idx')).read_bytes(),np.uint8).reshape(88,88) for key in keys]
    images=[np.array(Image.open(a.theme/(key+'.png')).convert('RGBA')) for key in keys]
    expected_rgb=[next(row['used_rgb_sha256'] for row in source['items'] if row['key']==key) for key in keys]
    old_files=sorted((a.before/'sources').glob('*.idx'));new_files=sorted((a.after/'sources').glob('*.idx'))
    assert old_files and [p.name for p in old_files]==[p.name for p in new_files]
    shots=[];full_case=None
    for p,q in zip(old_files,new_files):
        assert p.read_bytes()==q.read_bytes() and p.with_suffix('.pal').read_bytes()==q.with_suffix('.pal').read_bytes()
        old_png=a.before/'window-shots'/(p.stem+'.png');new_png=a.after/'window-shots'/(q.stem+'.png')
        old=np.array(Image.open(old_png).convert('RGBA'));new=np.array(Image.open(new_png).convert('RGBA'))
        case=(q.read_bytes(),q.with_suffix('.pal').read_bytes(),old,new,refs,images,expected_rgb)
        row=verify_shot(*case)
        row.update(capture=q.stem,indexed_sha256=sha(q),palette_sha256=sha(q.with_suffix('.pal')),
                   before_png_sha256=sha(old_png),after_png_sha256=sha(new_png),source_frame_offset=-1)
        shots.append(row)
        if row['distance']==0:full_case=case
    assert full_case is not None,'預定抽樣未捕捉完整正常近景'
    for kind in ('shot-foreground','shot-background','shot-palette'):
        bad=list(full_case);bad[3]=bad[3].copy()
        ref=refs[0]
        if kind=='shot-foreground':
            y,x=np.argwhere(ref!=0)[0];bad[3][72+y*3,312+x*3,0]^=1
        elif kind=='shot-background':bad[3][72,312,0]^=1
        else:
            data=bytearray(bad[1]);data[int(ref[ref!=0][0])*3]^=1;bad[1]=bytes(data)
        try:verify_shot(*bad)
        except AssertionError:controls.append(dict(case=kind,rejected=True))
        else:raise AssertionError('截圖負對照未拒絕：'+kind)
    observed=sorted({s['distance'] for s in shots if s['full_source']})
    report.update(negative_controls=controls,negative_controls_all_pass=True,normal_shots=shots,
                  normal_full_source_captures=[s['capture'] for s in shots if s['full_source']],
                  normal_distance_frames_observed=observed,unobserved_normal_distance_frames=sorted({0,1,2}-set(observed)),
                  source_proof_sha256=sha(a.source_proof),baseline_receipt_sha256=sha(a.baseline_receipt),
                  runtime_inputs_sha256=sha(a.after/'inputs.json'),runtime_log_sha256=sha(a.after/'run.log'),
                  old_theme_runtime_inputs_sha256=sha(a.before/'inputs.json'),old_theme_runtime_log_sha256=sha(a.before/'run.log'))
    a.report.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(report,ensure_ascii=False,indent=2))


if __name__=='__main__':main()
