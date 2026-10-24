#!/usr/bin/env python3
"""核對既有正常G8路線的街景HD與舊主題，保留原版狀態及存檔。"""
import argparse
import hashlib
import json
from pathlib import Path
import re


def sha(data):
    return hashlib.sha256(data).hexdigest()


def tree(root):
    return {str(p.relative_to(root)): sha(p.read_bytes()) for p in sorted(root.rglob('*')) if p.is_file()}


def endpoint(log):
    a = re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)', log)
    b = re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)', log)
    assert a and b and a.groups()[:2] == ('4800', '462000000')
    return a.groups() + b.groups()


def check(a):
    import numpy as np
    from PIL import Image
    folders = [a.out/'hd', a.out/'base']
    logs = [(p/'run.log').read_text() for p in folders]
    state = [endpoint(log) for log in logs]
    assert state[0] == state[1] == endpoint((a.reference/'run.log').read_text())
    records = [json.loads((p/'inputs.json').read_text()) for p in folders]
    assert records[0] == records[1]
    assert tree(a.theme) == records[0]['walls']
    for name in ('source.idx', 'source.pal'):
        data = [(p/name).read_bytes() for p in (*folders, a.reference)]
        assert data[0] == data[1] == data[2]
    assert (folders[0]/'original.rgba').read_bytes() == (folders[1]/'original.rgba').read_bytes()
    frontend = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt', 'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
    guest = [{k:v for k,v in tree(p/'save').items() if k not in frontend} for p in (*folders, a.reference)]
    assert guest[0] == guest[1] == guest[2]
    frames = [np.asarray(Image.open(p/'final.png').convert('RGBA')) for p in folders]
    assert frames[0].shape == frames[1].shape == (600, 1440, 4)
    diff = np.any(frames[0] != frames[1], axis=2)
    allowed = np.zeros((600, 1440), bool)
    allowed[72:336, 312:576] = True
    assert np.count_nonzero(diff) > 0 and not np.any(diff & ~allowed), '差異超出3D視窗'
    diagnostic = np.asarray(Image.open(a.diagnostic).convert('RGBA'))
    assert np.array_equal(frames[0][72:336, 312:576], diagnostic[72:336, 72:336]), '正常街景不符逐件已驗收的組合'
    audit = re.search(r'pod-play: 閃現 (\d+) 次', logs[0])
    assert audit and audit[1] == '0', '首次繪製閃現未過'
    report = dict(scope='同一G8正常冷啟動讀槽A及4800格；18來源／121格街景；不外推其他地點',
                  frames=4800, steps=462000000, memory_sha256=state[0][2], indexed_sha256=state[0][3], cpu_sha256=state[0][4], palette_sha256=state[0][5],
                  binary_sha256=records[0]['binary_sha256'], original_state_equal=True, guest_files=len(guest[0]), guest_state_equal=True,
                  changed_pixels=int(np.count_nonzero(diff)), outside_viewport_changes=0, final_viewport_matches_diagnostic=True, flash_count=0,
                  frame_sha256={p.name:sha((p/'final.png').read_bytes()) for p in folders}, inputs_sha256=sha((folders[0]/'inputs.json').read_bytes()))
    (a.out/'summary.json').write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(report, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--record', type=Path)
    for name in ('out', 'reference', 'theme', 'diagnostic'):
        ap.add_argument('--'+name, type=Path)
    a = ap.parse_args()
    if a.record:
        data = {name:tree(Path('/'+name)) for name in ('orig','font','text','seed','base','walls')}
        data['binary_sha256'] = sha(Path('/app/pod-play-wall').read_bytes())
        a.record.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')
    else:
        assert a.out and a.reference and a.theme and a.diagnostic
        check(a)
