#!/usr/bin/env python3
"""核對攻略詞彙整理前後的正常玩家路線，不改寫原版或補猜攻略資料。"""
import argparse
import csv
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

from PIL import Image, ImageChops


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def tree(p):
    return {str(q.relative_to(p)): sha(q) for q in sorted(p.rglob('*')) if q.is_file()}


def rows(p):
    return list(csv.DictReader(p.open(encoding='utf-8'), delimiter='\t'))


def state(p):
    text = p.read_text()
    assert 'frames=4800' in text and '缺字' not in text and '截圖失敗' not in text
    return {key: re.search(key + r'_sha256=([0-9a-f]{64})', text)[1]
            for key in ['cpu', 'memory', 'indexed', 'palette']}


def verify(project, root):
    before, after = root / 'gui-before', root / 'gui-after'
    prior = json.loads((project / 'docs/re/023-draw-closure-runtime.json').read_text())
    binary = project / 'workplace/dosgolem/workplace/bin/pod-play'
    assert sha(binary) == prior['binary_sha256']
    old, new = rows(root / 'text-before/walkthrough.zh-TW.tsv'), rows(project / 'text/walkthrough.zh-TW.tsv')
    assert len(old) == len(new) == 78
    changed = []
    for a, b in zip(old, new):
        assert all(a[k] == b[k] for k in ['chapter', 'title', 'source'])
        if a['body'] != b['body']:
            changed.append(int(b['chapter']))
    assert changed == [36, 38, 43, 45, 46, 52, 53, 55]
    old_files, new_files = tree(root / 'text-before'), tree(project / 'text')
    assert old_files.keys() == new_files.keys()
    assert [k for k in old_files if old_files[k] != new_files[k]] == ['walkthrough.zh-TW.tsv']
    for args in [['--self-check'], ['--check']]:
        subprocess.run([sys.executable, '-I', str(project / 'tools/walkthrough_terms.py'), *args], cwd=project, check=True)
    assert state(before / 'run.log') == state(after / 'run.log')
    for phase in [before, after]:
        assert json.loads((phase / 'before.json').read_text())['binary_sha256'] == sha(binary)
        receipt = json.loads((phase / 'receipt.json').read_text())
        assert receipt['guide_frame'] == 4440 and receipt['visited'] == ['1/10 11 2']
        assert receipt['catalogs']['walkthrough.zh-TW.tsv'] == sha(
            project / 'text/walkthrough.zh-TW.tsv' if phase == after else root / 'text-before/walkthrough.zh-TW.tsv')
    for suffix in ['idx', 'pal', 'glyphs.json']:
        assert sha(before / ('source.' + suffix)) == sha(after / ('source.' + suffix))
    assert sha(before / 'final.rgba') == sha(after / 'final.rgba')
    seed = project / 'workplace/probe/g6-saveload-receipt/scratch/SAVE'
    assert tree(before / 'game/SAVE') == tree(after / 'game/SAVE') == tree(seed)
    assert tree(before / 'game') == tree(after / 'game')
    shots_before, shots_after = tree(before / 'shots'), tree(after / 'shots')
    assert shots_before.keys() == shots_after.keys()
    changed_shots = []
    for name, digest in shots_before.items():
        if digest == shots_after[name]:
            continue
        assert name in {'pod-f004440.png', 'pod-f004455.png', 'pod-f004465.png'}
        a = Image.open(before / 'shots' / name).convert('RGB')
        b = Image.open(after / 'shots' / name).convert('RGB')
        assert a.size == b.size == (640, 400)
        assert ImageChops.difference(a.crop((0, 384, 640, 400)), b.crop((0, 384, 640, 400))).getbbox() is None
        changed_shots.append(name)
    assert changed_shots
    tests = project / 'workplace/dosgolem/workplace'
    for name in ['guide-closure-pod-tests.log', 'guide-closure-input-tests.log']:
        text = (tests / name).read_text()
        assert 'ok  ' in text and 'FAIL' not in text
    assert 'TestMapPanelKeyRouting' in (tests / 'guide-closure-input-tests.log').read_text()
    return dict(status='PASS', source='normal cold boot; load A; answer YOU; map; browse guide through chapter 36; map; close',
                program_commit=prior['program_commit'], binary_sha256=sha(binary),
                rows=78, changed_sections=changed, source_and_titles_preserved=True,
                guest_state=state(after / 'run.log'), original_and_client_files_equal=True,
                original_save_files_unchanged=True, original_sources_equal=True,
                changed_shots=changed_shots, guide_outside_panel_difference=0,
                unknown='Eight GEO names, unobserved marker coordinates and ambiguous source names remain unknown; no new exploration or guessed labels.',
                text_before_sha256=old_files, text_after_sha256=new_files,
                fonts={p.name: sha(p) for p in sorted((project / 'workplace/font').glob('*.golemfnt'))},
                before_evidence_sha256=tree(before), after_evidence_sha256=tree(after),
                tools_sha256={name: sha(project / 'tools' / name) for name in
                              ['walkthrough_terms.py', 'walkthrough_build.py', 'map_play_check.py', 'guide_closure_check.py', 'guide_closure_gui.sh']},
                tests_sha256={name: sha(tests / name) for name in ['guide-closure-pod-tests.log', 'guide-closure-input-tests.log']})


if __name__ == '__main__':
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, default=Path('/project'))
    ap.add_argument('--evidence', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    args = ap.parse_args()
    report = verify(args.project, args.evidence)
    assert args.out.parent.stat().st_uid == 1000
    args.out.write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps({'status': report['status'], 'changed_sections': report['changed_sections'],
                      'changed_shots': report['changed_shots'], 'sha256': sha(args.out)}))
