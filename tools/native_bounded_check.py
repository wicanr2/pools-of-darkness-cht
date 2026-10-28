#!/usr/bin/env python3
"""驗證有界 Go 跨度的原版串流、續行分類及正常玩家收據。"""
import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import re
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


def verify_resolution(spec):
    text = spec.read_text()
    initial, ready = '<a id="bounded-memory-initial"></a>', '<a id="bounded-memory-ready"></a>'
    assert text.count(initial) == text.count(ready) == 1
    before, after = text.split(initial)[1].split(ready)
    marker = 'BOUNDED-MEMORY-READY-20261028'
    assert marker in before and '(#bounded-memory-ready)' in before
    assert marker in after and '(#bounded-memory-initial)' in after
    keys = ('DOS/GAME.EXE+GAME.OVR/runtime:1D1F:046E', 'DOS/GAME.EXE+GAME.OVR/runtime:1ED4:1A5C',
            'DOS/VGA.DRV/eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f:0x68F4',
            'DOS/VGA.DRV/eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f:0x6B33')
    assert all(key in after for key in keys)
    return dict(keys=keys, correction=marker, spec_sha256=sha(spec))


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--commit', required=True)
    ap.add_argument('--stage', choices=('trace', 'full'), default='full')
    ap.add_argument('--out', type=Path, required=True)
    a = ap.parse_args()
    assert not a.out.exists(), '保留舊收據，請改輸出檔名'
    p = a.project
    module = p/'workplace/dosgolem'
    root = p/'workplace/probe/draw-intercept-v1/native-bounded-v1'
    previous = p/'workplace/probe/draw-intercept-v1/native-fade-v1/trace-original-entries'
    baseline = read(p/'docs/re/019-native-dissolve-runtime.json')
    assert subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=module, text=True).strip() == a.commit
    assert not subprocess.check_output(['git', 'status', '--porcelain'], cwd=module)
    parent = subprocess.check_output(['git', 'rev-parse', 'HEAD^'], cwd=module, text=True).strip()
    assert parent == baseline['program_commit']
    names = sorted(f.name for f in (module/'apps/pod/drawblocks').glob('native_bounded*.go'))
    assert len(names) == 8
    changed = subprocess.check_output(['git', 'diff', '--name-only', parent, a.commit], cwd=module, text=True).splitlines()
    assert sorted(changed) == sorted('apps/pod/drawblocks/'+name for name in names+['effects.go'])
    for name in names:
        if name.endswith('_test.go'):
            continue
        source = (module/'apps/pod/drawblocks'/name).read_text()
        assert all(token not in source for token in ('.Step(', 'effect{', 'b.run(', 'DecodeOpcode'))
    build = (root/'build.txt').read_text()
    assert 'go1.24.13' in build and 'vcs.revision='+a.commit in build and 'vcs.modified=false' in build
    spec = p/'docs/spec/025-direct-render-commands.md'
    resolution = verify_resolution(spec)
    old = read(previous/'calls.json')
    current = root/'trace-native'
    new = read(current/'calls.json')
    state_keys = ('Calls', 'Digest', 'VideoCount', 'AudioCount', 'FrameCount', 'DOSCount',
                  'VideoSHA256', 'AudioSHA256', 'FramesSHA256', 'DOSSHA256')
    for folder in (current, root/'trace-observed', root/'trace-entry-observed'):
        row = read(folder/'calls.json')
        assert all(row[key] == old[key] for key in state_keys), folder
        assert tree(folder/'save') == tree(previous/'save')
        assert row['Commands']['Mismatches'] == 0
        for field in ('Captures', 'DrawCaptures'):
            sources = {item['Memory']: item for item in old[field]}
            for item in row[field]:
                assert item == sources[item['Memory']]
                assert sha(folder/item['Memory']) == sha(previous/item['Memory']) == item['MemorySHA256']
    observed = read(root/'trace-entry-observed/calls.json')
    assert len(observed['DrawCaptures'])+len(observed['Captures']) == 47
    assert len(new['DrawCaptures'])+len(new['Captures']) == 46
    manifest = read(root/'entry-observed-build.json')
    assert manifest['program_commit'] == a.commit
    assert manifest['binary_sha256'] == sha(module/'workplace/bin'/manifest['binary'])
    assert manifest['builder_sha256'] == sha(p/'tools/build_observed_draw_probe.py')
    for name, row in manifest['source_files'].items():
        assert sha(module/'apps/pod/drawblocks'/name) == row['original_sha256']
        assert sha(module/'workplace/native-bounded-entry-observed-blocks'/name) == row['observed_sha256']
        if name != 'effects.go':
            assert row['original_sha256'] == row['observed_sha256']
    counts = new['AlgorithmicPreparedByExit']
    assert sum(row[0] for row in counts.values()) == new['AlgorithmicPlans']
    assert sum(row[1] for row in counts.values()) == new['AlgorithmicInstructions']
    fixed = new['HostInstructions']+new['BlockInstructions']-new['AlgorithmicInstructions']
    fixed_starts = read(root/'trace-observed/fixed-starts.json')
    assert sum(row[1] for row in fixed_starts.values()) == fixed
    assert 0 < fixed < baseline['fixed_semantic_instructions']
    assert not any(key.startswith(('1d1f:', '1ed4:')) for key in fixed_starts)
    assert not any(key.startswith('48a1:'+ip) for key in fixed_starts
                   for ip in ('68f4', '68fd', '690b', '691a'))
    log = (module/'workplace/native-bounded-product-tests.log').read_text()
    for line in ('zero synthetic cursors=180 bounded plans=2160', 'zero captured cursors=24 bounded plans=288',
                 'copy synthetic cursors=756 bounded plans=12852', 'copy captured cursors=86 bounded plans=1462',
                 'RLE synthetic cursors=624 bounded plans=3120', 'RLE captured cursors=58 bounded plans=290',
                 'memory whole SaveState cases=240 Go plans=1232 CPU/fixed pixel fallback=0',
                 'RLE whole SaveState cases=200 Go plans=2200 CPU/fixed pixel fallback=0',
                 'bounded refusal controls=18 unchanged inputs', 'full-screen spans original comparisons=4'):
        assert line in log, line
    assert '\nPASS\n' in log
    result = dict(scope='有界共享清除、複製及兩個 RLE 核心；其他繪圖仍未全部接管',
                  program_commit=a.commit, parent_commit=parent,
                  binary_sha256=sha(module/'workplace/bin/pod-play-nativebounded'),
                  original_state_and_ordered_events_equal=True, virtual_steps=new['Digest']['Steps'],
                  guest_files=len(tree(current/'save')), guest_files_equal=True,
                  direct_witnesses=46, split_observation_witnesses=47,
                  algorithmic_plans=new['AlgorithmicPlans'], algorithmic_instructions=new['AlgorithmicInstructions'],
                  fixed_semantic_instructions=fixed, previous_fixed_semantic_instructions=baseline['fixed_semantic_instructions'],
                  fixed_by_start=fixed_starts, algorithmic_by_exit=counts,
                  new_shared_or_rle_fixed_pixel_writes=0,
                  classification_scope='本條正常 demo 的計畫分類，並非全遊戲覆蓋率、全部像素回退或 FPS 聲明',
                  prototype_and_original_tests=dict(zero_plans=2448, copy_plans=14314, rle_plans=3410,
                      shared_whole_cases=240, rle_whole_cases=200, large_spans=4, refusal_controls=18),
                  source_sha256={name: sha(module/'apps/pod/drawblocks'/name) for name in names+['effects.go']},
                  spec_sha256=sha(spec), resolution_backlinks=resolution,
                  test_log_sha256=sha(module/'workplace/native-bounded-product-tests.log'),
                  build_info_sha256=sha(root/'build.txt'),
                  original_baseline_sha256=sha(previous/'calls.json'),
                  source_reports={name: sha(root/name/'calls.json') for name in ('trace-native', 'trace-observed', 'trace-entry-observed')},
                  entry_observation=manifest, previous_receipt_sha256=sha(p/'docs/re/019-native-dissolve-runtime.json'))
    if a.stage == 'full':
        for name in ('native-bounded-all-drawblocks.log', 'native-bounded-core-tests.log'):
            body = (module/'workplace'/name).read_text()
            assert 'FAIL' not in body and 'ok  ' in body, name
        assert not (module/'workplace/native-bounded-vet.log').read_text()
        gui = p/'workplace/probe/g9-hd'
        demo = gui/'nativebounded-play-v1'
        summary = read(demo/'bigpic-summary.json')
        assert summary['binary_sha256'] == result['binary_sha256']
        assert summary['prior_original_state_equal'] and summary['prior_sample_composite_equal']
        ui = load(p/'tools/direct_ui_check.py')
        frontend = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt',
                    'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
        for mode in ('hd', 'base', 'english', 'twox'):
            folder, prior = demo/mode, gui/'nativedissolve-play-v2'/mode
            assert (folder/'script.txt').read_bytes() == (prior/'script.txt').read_bytes()
            assert '1800:d' in (folder/'script.txt').read_text()
            assert ui.endpoint(folder) == ui.endpoint(prior)
            assert read(folder/'inputs.json')['binary_sha256'] == result['binary_sha256']
            assert read(folder/'draw-report.json')['Mismatches'] == 0
        rows = []
        for route in ('wall', 'movie', 'title'):
            folder, prior = gui/f'nativebounded-ui-v1/{route}-direct', gui/f'nativedissolve-ui-v2/{route}-direct'
            assert ui.endpoint(folder) == ui.endpoint(prior)
            ci, bi = read(folder/'inputs.json'), read(prior/'inputs.json')
            assert ci['binary_sha256'] == result['binary_sha256']
            assert bi['binary_sha256'] == baseline['binary_sha256']
            assert all(ci[key] == bi[key] for key in ('orig', 'font', 'text', 'theme'))
            guests = [{k: v for k, v in tree(d/'save').items() if k not in frontend} for d in (folder, prior)]
            assert guests[0] == guests[1]
            for name in ('source.idx', 'source.pal', 'original.rgba', 'frame.png'):
                assert (folder/name).read_bytes() == (prior/name).read_bytes()
            for name in ('shots', 'window-shots', 'sources'):
                assert tree(folder/name) and tree(folder/name) == tree(prior/name)
            assert ui.audit(folder, route) == ui.audit(prior, route)
            assert read(folder/'draw-report.json')['Mismatches'] == 0
            rows.append(dict(route=route, screenshot_pairs=len(tree(folder/'shots')), **ui.audit(folder, route)))
        before, after = gui/'nativedissolve-save-v2/native', gui/'nativebounded-save-v1/native'
        inputs = [read(d/'before.json') for d in (before, after)]
        assert inputs[0].pop('binary_sha256') == baseline['binary_sha256']
        assert inputs[1].pop('binary_sha256') == result['binary_sha256']
        assert inputs[0] == inputs[1] and 'SAVE/savgamB.pty' not in inputs[1]
        saved = [read(d/'after-save.json') for d in (before, after)]
        assert saved[0] == saved[1] and tree(before/'game/SAVE') == tree(after/'game/SAVE')
        slot_a = {k[5:]: v for k, v in inputs[1].items() if k.startswith('SAVE/')}
        assert all(saved[1][key] == value for key, value in slot_a.items())
        for phase, frames, shots in (('save', 4700, 5), ('load', 3250, 1)):
            endpoints = []
            for d in (before, after):
                body = (d/f'{phase}.log').read_text()
                first = re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)', body)
                second = re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)', body)
                assert first and second and int(first[1]) == frames and int(first[2]) == frames*96250
                assert '字型缺字' not in body and '截圖失敗' not in body
                endpoints.append(first.groups()+second.groups())
            assert endpoints[0] == endpoints[1]
            for suffix in ('final.rgba', 'source.idx', 'source.pal'):
                assert (before/f'{phase}-{suffix}').read_bytes() == (after/f'{phase}-{suffix}').read_bytes()
            assert len(tree(after/f'{phase}-shots')) == shots
            assert tree(before/f'{phase}-shots') == tree(after/f'{phase}-shots')
            assert read(after/f'{phase}-draw.json')['Mismatches'] == 0
        assert tree(p/'workplace/hd/theme') == read(demo/'hd/inputs.json')['theme']
        abort_root = p/'workplace/probe/bounded-demo-control-v1'
        aborts = read(abort_root/'summary.json')
        reference = read(p/'docs/re/020-demo-control-runtime.json')['normal']
        assert aborts['binary_sha256'] == result['binary_sha256']
        for case in ('menu', 'story', 'combat-repeat'):
            actual, prior = aborts['cases'][case], reference['cases'][case]
            for key in ('timeouts_cleared', 'aborts', 'returns', 'last_abort_steps', 'hashes', 'screenshots'):
                assert actual[key] == prior[key], (case, key)
            assert actual['log_sha256'] == sha(abort_root/case/'run.log')
            assert actual['screenshots'] == tree(abort_root/case/'shots')
        result.update(normal_demo=summary, normal_ui=rows,
                      normal_save_roundtrip=dict(screenshot_pairs=6, all_save_bytes_equal=True,
                          slot_a_unchanged_files=len(slot_a), save_files=len(saved[1]), slot_b_sha256=saved[1]['savgamB.pty']),
                      theme_files=len(tree(p/'workplace/hd/theme')),
                      demo_abort_regression=aborts,
                      demo_abort_summary_sha256=sha(abort_root/'summary.json'),
                      regression_logs={name: sha(module/'workplace'/name) for name in
                          ('native-bounded-all-drawblocks.log', 'native-bounded-core-tests.log', 'native-bounded-vet.log')})
        if (root/'activation.json').exists():
            activation = read(root/'activation.json')
            assert activation['candidate_sha256'] == result['binary_sha256'] == sha(module/'workplace/bin/pod-play')
            assert sha(module/'workplace/bin'/activation['previous_binary']) == activation['previous_sha256'] == baseline['binary_sha256']
            result['activation'] = activation
    result['tool_sha256'] = sha(Path(__file__))
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({key: result[key] for key in ('algorithmic_plans', 'algorithmic_instructions', 'fixed_semantic_instructions', 'guest_files')}))


if __name__ == '__main__':
    main()
