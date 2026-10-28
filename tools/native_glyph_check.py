#!/usr/bin/env python3
"""驗證Go變寬字模像素核心；正常路線抽樣與受控原版證據分列。"""
import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import subprocess


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    return json.loads(path.read_text())


def tree(path):
    return {str(p.relative_to(path)): sha(p) for p in sorted(path.rglob('*')) if p.is_file()}


def load(path):
    spec = importlib.util.spec_from_file_location(path.stem, path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def verify_resolution(path):
    text = path.read_text()
    initial, ready = '<a id="glyph-5aa4-initial"></a>', '<a id="glyph-5aa4-ready"></a>'
    assert text.count(initial) == text.count(ready) == 1
    before, after = text.split(initial)[1].split(ready)
    marker = 'GLYPH5AA4-READY-20261028'
    key = 'DOS/VGA.DRV/eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f:0x5AA4'
    assert marker in before and '(#glyph-5aa4-ready)' in before
    assert marker in after and key in after and '48A1:5AA4' in after and '(#glyph-5aa4-initial)' in after
    return dict(key=key, correction=marker, spec_sha256=sha(path))


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--commit', required=True)
    ap.add_argument('--stage', choices=('trace', 'full'), default='full')
    ap.add_argument('--out', type=Path, required=True)
    a = ap.parse_args()
    assert not a.out.exists(), '保留既有收據，請使用新檔名'
    p = a.project
    module = p/'workplace/dosgolem'
    root = p/'workplace/probe/draw-intercept-v1/native-glyph-v1'
    baseline_path = p/'docs/re/019-native-bounded-runtime.json'
    baseline = read(baseline_path)
    assert subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=module, text=True).strip() == a.commit
    assert not subprocess.check_output(['git', 'status', '--porcelain'], cwd=module)
    parent = subprocess.check_output(['git', 'rev-parse', 'HEAD^'], cwd=module, text=True).strip()
    assert parent == baseline['program_commit']
    folder = module/'apps/pod/drawblocks'
    names = sorted(f.name for f in folder.glob('native_glyph5aa4*.go'))
    assert len(names) == 7
    changed = subprocess.check_output(['git', 'diff', '--name-only', parent, a.commit], cwd=module, text=True).splitlines()
    assert sorted(changed) == sorted('apps/pod/drawblocks/'+name for name in names+['effects.go'])
    for name in names:
        if name.endswith('_test.go'):
            continue
        source = (folder/name).read_text()
        assert all(token not in source for token in ('.Step(', 'effect{', 'b.run(', 'DecodeOpcode'))
    for name, digest in baseline['source_sha256'].items():
        if name != 'effects.go':
            assert sha(folder/name) == digest
    build = (root/'build.txt').read_text()
    assert 'go1.24.13' in build and 'vcs.modified=false' in build and 'vcs.revision='+a.commit in build
    original = p/'workplace/input/dos/pod/VGA.DRV'
    assert sha(original) == 'eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f'
    assert hashlib.sha256(original.read_bytes()[0x5aa4:0x5c77]).hexdigest() == '0ac343cdf83afd9cff35de6fa8dc5d42035e3be7f885c6da18e090dd2c2906aa'
    log = module/'workplace/native-glyph-product-tests.log'
    text = log.read_text()
    for marker in ('variable glyph raw/mapped ordered pixel comparisons=30',
                   'glyph source-driven row cursor cases=30 original comparisons=2418',
                   'glyph source-driven single-step spans cases=30 original comparisons=572',
                   'glyph whole SaveState comparisons=300 Go plans=1298 fixed control=1840 CPU control=23640 CPU/fixed glyph pixels=0',
                   'glyph boundary/full canvas/8 parents/segment end/DF SaveState cases=20 Go plans=824',
                   'glyph unproved/alias states refused=54 unchanged inputs'):
        assert marker in text, marker
    assert '\nPASS\n' in text
    resolution = verify_resolution(p/'docs/spec/025-direct-render-commands.md')
    current = root/'trace-native'
    previous = p/'workplace/probe/draw-intercept-v1/native-bounded-v1/trace-native'
    oracle = p/'workplace/probe/draw-intercept-v1/native-fade-v1/trace-original-entries'
    old, new, reference = [read(d/'calls.json') for d in (previous, current, oracle)]
    fields = ('Calls', 'Digest', 'VideoCount', 'AudioCount', 'FrameCount', 'DOSCount',
              'VideoSHA256', 'AudioSHA256', 'FramesSHA256', 'DOSSHA256')
    assert all(old[key] == new[key] == reference[key] for key in fields)
    assert tree(current/'save') == tree(previous/'save') == tree(oracle/'save')
    assert new['Commands']['Mismatches'] == 0
    for field in ('Captures', 'DrawCaptures'):
        assert new[field] == old[field]
        for row in new[field]:
            assert sha(current/row['Memory']) == sha(previous/row['Memory']) == row['MemorySHA256']
    counts = new['AlgorithmicPreparedByExit']
    assert counts == old['AlgorithmicPreparedByExit']
    assert sum(row[0] for row in counts.values()) == new['AlgorithmicPlans'] == 112012
    assert sum(row[1] for row in counts.values()) == new['AlgorithmicInstructions'] == 3400397
    fixed = new['HostInstructions']+new['BlockInstructions']-new['AlgorithmicInstructions']
    assert fixed == baseline['fixed_semantic_instructions'] == 1062
    normal_glyph = {key: row for key, row in counts.items() if key.startswith('48a1:') and 0x5aa4 <= int(key.split(':')[1], 16) < 0x5c77}
    assert not normal_glyph, '新入口若已在正常路線使用，須重評抽樣範圍'
    result = dict(scope='Go變寬字模像素核心；來源準備、換列與回車控制保留原版，其他繪圖未全部完成',
                  program_commit=a.commit, parent_commit=parent,
                  binary_sha256=sha(module/'workplace/bin/pod-play-nativeglyph'),
                  original_state_and_ordered_events_equal=True, virtual_steps=new['Digest']['Steps'],
                  guest_files=len(tree(current/'save')), guest_files_equal=True,
                  direct_original_witnesses=len(new['Captures'])+len(new['DrawCaptures']),
                  original_witnesses_equal=True, normal_glyph_plans=0,
                  normal_scope='新字模入口在此正常demo未觀測；新增像素語意的完成聲明只限受控原版條件',
                  algorithmic_plans=new['AlgorithmicPlans'], algorithmic_instructions=new['AlgorithmicInstructions'], fixed_semantic_instructions=fixed,
                  controlled_original=dict(ordered_pixels=30, single_step_spans=572, row_cursors=2418,
                      whole_calls=300, whole_go_plans=1298, whole_fixed_control=1840, whole_cpu_control=23640,
                      boundary_calls=20, boundary_go_plans=824, refusals=54, cpu_or_fixed_glyph_pixels=0),
                  source_sha256={name: sha(folder/name) for name in names+['effects.go']},
                  product_test_log_sha256=sha(log), resolution_backlinks=resolution,
                  prior_full_receipt_sha256=sha(baseline_path), build_info_sha256=sha(root/'build.txt'),
                  trace_binary_sha256=sha(module/'workplace/bin/draw-call-probe-nativeglyph'),
                  source_reports={str(d.relative_to(p)): sha(d/'calls.json') for d in (current, previous, oracle)},
                  prior_gates_note='前版四路demo、39對介面、6對真存讀檔及demo中斷完整驗收保留於前版收據；本批不冒稱重跑這些全部路線')
    if a.stage == 'full':
        for name in ('native-glyph-all-drawblocks.log', 'native-glyph-core-tests.log'):
            body = (module/'workplace'/name).read_text()
            assert 'FAIL' not in body and 'ok  ' in body
        assert not (module/'workplace/native-glyph-vet.log').read_text()
        ui = load(p/'tools/direct_ui_check.py')
        frontend = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt',
                    'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
        rows = []
        for mode, folder, prior in (
            ('hd', p/'workplace/probe/g9-hd/nativeglyph-play-v1/hd', p/'workplace/probe/g9-hd/nativebounded-play-v1/hd'),
            ('wall', p/'workplace/probe/g9-hd/nativeglyph-ui-v1/wall-direct', p/'workplace/probe/g9-hd/nativebounded-ui-v1/wall-direct')):
            assert ui.endpoint(folder) == ui.endpoint(prior)
            ci, bi = read(folder/'inputs.json'), read(prior/'inputs.json')
            assert ci['binary_sha256'] == result['binary_sha256'] and bi['binary_sha256'] == baseline['binary_sha256']
            assert all(ci[key] == bi[key] for key in ('orig', 'font', 'text', 'theme'))
            assert ui.audit(folder, mode)['unexpected_flash_count'] == 0
            assert ui.audit(folder, mode) == ui.audit(prior, mode)
            guests = [{k: v for k, v in tree(d/'save').items() if k not in frontend} for d in (folder, prior)]
            assert guests[0] == guests[1]
            for name in ('source.idx', 'source.pal', 'original.rgba', 'frame.png'):
                assert (folder/name).read_bytes() == (prior/name).read_bytes()
            for name in ('shots', 'window-shots', 'sources'):
                assert tree(folder/name) and tree(folder/name) == tree(prior/name)
            if mode == 'hd':
                assert '1800:d' in (folder/'script.txt').read_text()
                assert (folder/'script.txt').read_bytes() == (prior/'script.txt').read_bytes()
            assert read(folder/'draw-report.json')['Mismatches'] == 0
            assert tree(p/'workplace/hd/theme') == ci['theme']
            rows.append(dict(route=mode, endpoint=ui.endpoint(folder), screenshot_pairs=len(tree(folder/'shots')),
                             screenshots_and_original_state_equal=True, unexpected_flash_count=0))
        result['normal_samples'] = rows
        result['regression_logs'] = {name: sha(module/'workplace'/name) for name in ('native-glyph-all-drawblocks.log', 'native-glyph-core-tests.log', 'native-glyph-vet.log')}
        if (root/'activation.json').exists():
            activation = read(root/'activation.json')
            assert sha(module/'workplace/bin/pod-play') == activation['candidate_sha256'] == result['binary_sha256']
            assert sha(module/'workplace/bin'/activation['previous_binary']) == activation['previous_sha256'] == baseline['binary_sha256']
            result['activation'] = activation
    result['tool_sha256'] = sha(Path(__file__))
    result['tools'] = {name: sha(p/name) for name in ('tools/draw_call_probe.go', 'tools/direct_ui_check.py', 'tools/vga-battle-play-receipt.sh', 'tools/direct-ui-receipt.sh')}
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({key: result[key] for key in ('normal_glyph_plans', 'algorithmic_plans', 'fixed_semantic_instructions', 'guest_files')}))


if __name__ == '__main__':
    main()
