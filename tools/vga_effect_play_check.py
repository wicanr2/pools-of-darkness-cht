#!/usr/bin/env python3
"""COMSPR首批正常DEMO增量抽驗；允許009特效矩形內的HD變化。"""
import argparse
import copy
import hashlib
import json
from pathlib import Path
import re

import numpy as np
from PIL import Image

FRONTEND = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt',
            'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
KEYS = ['comspr-' + key for key in ('000', '005', '006', '009', '010', '013')]


def sha(data):
    return hashlib.sha256(data).hexdigest()


def tree(root):
    return {str(p.relative_to(root)): sha(p.read_bytes()) for p in sorted(root.rglob('*')) if p.is_file()}


def read_case(path):
    log = (path / 'run.log').read_text()
    state = {k: re.search(k + r'_sha256=([0-9a-f]{64})', log)[1]
             for k in ('cpu', 'memory', 'indexed', 'palette')}
    end = re.search(r'frames=(\d+) steps=(\d+)', log)
    state.update(frames=int(end[1]), steps=int(end[2]))
    audit = re.search(r'pod-play: 閃現 (\d+) 次', log)
    return dict(state=state, flash=int(audit[1]),
                inputs=json.loads((path / 'inputs.json').read_text()),
                idx=(path / 'source.idx').read_bytes(), pal=(path / 'source.pal').read_bytes(),
                original=(path / 'original.rgba').read_bytes(),
                guest={k: v for k, v in tree(path / 'save').items() if k not in FRONTEND},
                frame=np.array(Image.open(path / 'frame.png').convert('RGBA')),
                png_sha256=sha((path / 'frame.png').read_bytes()))


def verify(prior, before, after, theme):
    for item in (before, after):
        assert item['state'] == prior['state'], '原版CPU／RAM／畫面／步數不同'
        assert item['state']['frames'] == 7400 and item['state']['steps'] == 712250000
        assert item['flash'] == 0, '正常路徑閃現不為0'
        assert len(item['guest']) == 127 and item['guest'] == prior['guest'], 'DOS檔案不同'
        for field in ('idx', 'pal', 'original'):
            assert item[field] == prior[field], field + '不同'
        for name in ('orig', 'font', 'text'):
            assert item['inputs'][name] == prior['inputs'][name], name + '輸入不同'
        assert sha(item['idx']) == item['state']['indexed']
        assert sha(item['pal']) == item['state']['palette']
        assert item['frame'].shape == (600, 1440, 4)
    assert before['inputs']['binary_sha256'] == after['inputs']['binary_sha256'], '未用同一程式比較'
    assert np.array_equal(before['frame'], prior['frame']), '新程式配舊主題改變畫面'
    old, new = before['inputs']['theme'], after['inputs']['theme']
    assert old == prior['inputs']['theme'] and len(old) == 1258 and len(new) == 1271
    assert tree(theme) == new, '執行後主題變更'
    manifest = json.loads((theme / 'effects.json').read_text())
    assert manifest['version'] == 1 and manifest['kind'] == 'combat-effects'
    assert manifest['scale'] == 3 and manifest['grid'] == [8, 8, 7, 7, 24] and manifest['background_index'] == 8
    assert [row['key'] for row in manifest['items']] == KEYS
    added = {'effects.json'}
    for row in manifest['items']:
        assert row['size'] == [24, 24]
        for name, digest in ((row['png'], row['png_sha256']), (row['indexed'], row['indexed_sha256'])):
            assert new[name] == digest
            added.add(name)
    assert set(new) - set(old) == added
    assert all(new[k] == v for k, v in old.items() if k != 'vga-layers-provenance.json'), '既有素材變更'
    indexed = np.frombuffer(after['idx'], dtype=np.uint8).reshape(200, 320)
    reference = np.frombuffer((theme / 'comspr-009.idx').read_bytes(), dtype=np.uint8).reshape(24, 24)
    assert np.array_equal(indexed[32:56, 56:80], reference), '正常009來源不符'
    assert np.all(reference != 8)
    palette = np.frombuffer(after['pal'], dtype=np.uint8).reshape(256, 3)
    assert sha(after['pal'][192:288]) == '903dcf3a1bc9aaa4b5b0829df3923b73fc4b8e7548ce33a84c8f154639d510a0'
    original = np.frombuffer(after['original'], dtype=np.uint8).reshape(600, 960, 4)
    assert np.array_equal(original[:, :, :3], palette[indexed].repeat(3, 0).repeat(3, 1))
    assert np.all(original[:, :, 3] == 255)
    hd = np.array(Image.open(theme / 'comspr-009.png').convert('RGBA'))
    assert np.array_equal(after['frame'][96:168, 408:480], hd), '正常畫面的009與HD成品不同'
    diff = np.any(before['frame'] != after['frame'], axis=2)
    allowed = np.zeros((600, 1440), dtype=bool); allowed[96:168, 408:480] = True
    assert np.any(diff) and not np.any(diff & ~allowed), '變化超出009核准矩形'
    return dict(status='PASS', scope='冷啟動明確按D，7400格正常DEMO；其餘五種施法事件未觀測',
                frames=7400, steps=712250000, state=after['state'],
                binary_sha256=after['inputs']['binary_sha256'], prior_binary_sha256=prior['inputs']['binary_sha256'],
                guest_files=127, guest_state_equal=True, original_endpoint_equal=True,
                old_theme_new_program_composite_equal=True, approved_effect_rect=[56,32,24,24],
                final_composite_equal=False, approved_hd_change_only=True,
                changed_pixels=int(np.count_nonzero(diff)), outside_effect_changes=0, flash_count=0,
                prior_assets_preserved=True, theme_files=1271,
                frames_sha256=dict(prior=prior['png_sha256'], before=before['png_sha256'], after=after['png_sha256']))


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for name in ('prior', 'before', 'after', 'theme', 'report'):
        ap.add_argument('--' + name, type=Path, required=True)
    a = ap.parse_args()
    prior, before, after = [read_case(p) for p in (a.prior, a.before, a.after)]
    report = verify(prior, before, after, a.theme)
    controls = []
    for kind in ('cpu', 'source', 'palette', 'guest', 'outside-pixel'):
        bad = copy.deepcopy(after)
        if kind == 'cpu': bad['state']['cpu'] = '0' * 64
        elif kind == 'source': bad['idx'] = bytes([bad['idx'][0] ^ 1]) + bad['idx'][1:]
        elif kind == 'palette': bad['pal'] = bytes([bad['pal'][0] ^ 1]) + bad['pal'][1:]
        elif kind == 'guest': bad['guest'][next(iter(bad['guest']))] = '0' * 64
        else: bad['frame'][0, 0, 0] ^= 1
        try:
            verify(prior, before, bad, a.theme)
        except AssertionError:
            controls.append(dict(case=kind, rejected=True))
        else:
            raise AssertionError('負對照未拒絕：' + kind)
    report['negative_controls'] = controls
    report['negative_controls_all_pass'] = True
    a.report.write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps(report, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    main()
