#!/usr/bin/env python3
"""驗證 Go 整個表面填色的一步跨度；受控原版與正常前端範圍分列。"""
import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import subprocess


def load(path):
    spec = importlib.util.spec_from_file_location(path.stem, path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def resolution(path):
    text = path.read_text()
    names = ('surface-fill-whole-initial', 'surface-fill-short-initial', 'surface-fill-spans-ready')
    anchors = {name: f'<a id="{name}"></a>' for name in names}
    assert all(text.count(anchor) == 1 for anchor in anchors.values())
    section = lambda name: text.split(anchors[name])[1].split('<a id=')[0]
    marker = 'SURFACE-FILL-SPAN-READY-20261029'
    ready = section('surface-fill-spans-ready')
    key = 'DOS/VGA.DRV/eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f:0x41AF'
    assert marker in ready and key in ready and '42A1' in ready and '42A5' in ready
    for name in names[:-1]:
        assert marker in section(name) and '(#surface-fill-spans-ready)' in section(name)
        assert '(#'+name+')' in ready
    return dict(key=key, correction=marker)


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--commit', required=True)
    ap.add_argument('--stage', choices=('trace', 'full'), default='full')
    ap.add_argument('--out', type=Path, required=True)
    a = ap.parse_args()
    assert not a.out.exists(), '保留既有收據，請使用新檔名'
    p = a.project
    helper = load(p/'tools/native_palette_check.py')
    sha, read, tree = helper.sha, helper.read, helper.tree
    module = p/'workplace/dosgolem'
    raw = p/'workplace/probe/draw-intercept-v1/native-surface-fill-spans-v1'
    baseline_path = p/'docs/re/019-native-primitive-spans-runtime.json'
    baseline = read(baseline_path)
    git = lambda *args: subprocess.check_output(['git', *args], cwd=module, text=True).strip()
    assert git('rev-parse', 'HEAD') == a.commit and not git('status', '--porcelain')
    parent = git('rev-parse', 'HEAD^')
    assert parent == baseline['program_commit']
    names = ('effects.go', 'native_surface_fill_spans.go',
             'native_surface_fill_spans_test.go', 'native_surface_fill_spans_guards_test.go')
    expected = ['apps/pod/drawblocks/'+name for name in names]
    assert sorted(git('diff', '--name-only', parent, a.commit).splitlines()) == sorted(expected)
    for name in expected:
        if not name.endswith(('_test.go', '/effects.go')):
            source = (module/name).read_text()
            assert all(token not in source for token in ('.Step(', 'effect{', 'b.run(', 'DecodeOpcode'))
    for name, digest in baseline['source_sha256'].items():
        if name not in expected:
            assert sha(module/name) == digest
    build = (raw/'build.txt').read_text()
    assert 'go1.24.13' in build and 'vcs.modified=false' in build and 'vcs.revision='+a.commit in build
    original = p/'workplace/input/dos/pod/VGA.DRV'
    assert sha(original) == 'eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f'
    assert hashlib.sha256(original.read_bytes()[0x41af:0x42b3]).hexdigest() == 'c7db9b7b119f14464a7e76d35125922a3559d10fef36624f79f5120dd0d02f00'
    log = module/'workplace/native-surface-fill-spans-product-tests.log'
    body = log.read_text()
    for marker in (
        'surface-fill span ordered original comparisons=1056 zero-count=500',
        'surface-fill span whole SaveState+PortLog cases=70 Go plans=3734 fixed control=1168 CPU control=31382 CPU/fixed surface-fill pixels=0',
        'surface-fill span positive phases=2 source/frame/cursor/model refusals=46 unchanged',
        'surface-fill nonzero root/eight-parent SaveState+PortLog cases=18 CPU/fixed surface-fill pixels=0',
    ):
        assert marker in body, marker
    assert '\nPASS\n' in body and 'FAIL' not in body
    old_log = module/'workplace/native-surface-fill-spans-old-fill-tests.log'
    old_body = old_log.read_text()
    assert 'complete original geometries/cost cases 18 independent plan comparisons 194' in old_body
    assert '\nPASS\n' in old_body and 'FAIL' not in old_body
    paths = [raw/'trace-native', p/'workplace/probe/draw-intercept-v1/native-primitive-spans-v1/trace-native',
             p/'workplace/probe/draw-intercept-v1/native-fade-v1/trace-original-entries']
    current, previous, oracle = [read(path/'calls.json') for path in paths]
    for field in ('Calls', 'Digest', 'VideoCount', 'AudioCount', 'FrameCount', 'DOSCount',
                  'VideoSHA256', 'AudioSHA256', 'FramesSHA256', 'DOSSHA256'):
        assert current[field] == previous[field] == oracle[field], field
    assert tree(paths[0]/'save') == tree(paths[1]/'save') == tree(paths[2]/'save')
    for field in ('Captures', 'DrawCaptures'):
        assert current[field] == previous[field]
        for row in current[field]:
            assert sha(paths[0]/row['Memory']) == sha(paths[1]/row['Memory']) == row['MemorySHA256']
    assert current['Commands']['Mismatches'] == 0
    counts = current['AlgorithmicPreparedByExit']
    assert sum(row[0] for row in counts.values()) == current['AlgorithmicPlans']
    assert sum(row[1] for row in counts.values()) == current['AlgorithmicInstructions']
    pixel_ips = {0x42a1, 0x42a5}
    observed = {key: value for key, value in counts.items() if key.startswith('48a1:') and int(key.split(':')[1], 16) in pixel_ips}
    delta = {key: [value[i]-previous['AlgorithmicPreparedByExit'].get(key, [0, 0])[i] for i in range(2)]
             for key, value in observed.items()}
    assert delta == {'48a1:42a1': [1, 1], '48a1:42a5': [1, 1]}
    for key in set(counts) | set(previous['AlgorithmicPreparedByExit']):
        if key not in delta:
            assert counts.get(key) == previous['AlgorithmicPreparedByExit'].get(key), key
    normal_delta = dict(algorithmic_plans=current['AlgorithmicPlans']-previous['AlgorithmicPlans'],
                        algorithmic_instructions=current['AlgorithmicInstructions']-previous['AlgorithmicInstructions'],
                        native_host_bytes=current['Commands']['HostBytes']-previous['Commands']['HostBytes'])
    assert normal_delta == dict(algorithmic_plans=2, algorithmic_instructions=2, native_host_bytes=320)
    fixed = current['HostInstructions']+current['BlockInstructions']-current['AlgorithmicInstructions']
    assert fixed <= baseline['fixed_semantic_instructions']
    result = dict(scope='Go整個表面填色的一步跨度；其他繪圖與未驗條件仍待盤點',
        program_commit=a.commit, parent_commit=parent,
        binary_sha256=sha(module/'workplace/bin/pod-play-nativesurfacefillspans'),
        original_state_and_ordered_events_equal=True, virtual_steps=current['Digest']['Steps'],
        guest_files=len(tree(paths[0]/'save')), guest_files_equal=True,
        direct_original_witnesses=len(current['Captures'])+len(current['DrawCaptures']), original_witnesses_equal=True,
        normal_surface_fill_pixel_exit_counts=observed, normal_surface_fill_pixel_exit_delta=delta,
        normal_delta=normal_delta,
        algorithmic_plans=current['AlgorithmicPlans'], algorithmic_instructions=current['AlgorithmicInstructions'], fixed_semantic_instructions=fixed,
        controlled_original=dict(ordered_comparisons=1056, zero_count=500, pixel_phases=2,
            whole_calls=70, whole_go_plans=3734, whole_fixed_control=1168, whole_cpu_control=31382,
            parent_calls=18, refusals=46, existing_rectangle_cases=18, existing_rectangle_plans=194,
            cpu_or_fixed_surface_fill_pixels=0),
        source_sha256={name: sha(module/name) for name in expected},
        product_test_log_sha256=sha(log), old_rectangle_log_sha256=sha(old_log), resolution_backlinks=resolution(p/'docs/spec/025-direct-render-commands.md'),
        spec_sha256=sha(p/'docs/spec/025-direct-render-commands.md'), prior_full_receipt_sha256=sha(baseline_path),
        build_info_sha256=sha(raw/'build.txt'), trace_binary_sha256=sha(module/'workplace/bin/draw-call-probe-nativesurfacefillspans'),
        source_reports={str(path.relative_to(p)): sha(path/'calls.json') for path in paths},
        prior_gates_note='前批完整四路demo、39對介面、6對存讀檔及demo中斷另存原收據，本批正常抽樣另列')
    if a.stage == 'full':
        logs = ['native-surface-fill-spans-all-drawblocks.log', 'native-surface-fill-spans-core-tests.log', 'native-surface-fill-spans-vet.log']
        for name in logs[:2]:
            text = (module/'workplace'/name).read_text()
            assert 'FAIL' not in text and 'ok  ' in text
        assert 'github.com/wicanr2/dosgolem/apps/pod/drawblocks\t' in (module/'workplace'/logs[0]).read_text()
        core = (module/'workplace'/logs[1]).read_text()
        for package in ('internal/machine', 'session', 'apps/pod', 'cmd/pod-play'):
            assert 'github.com/wicanr2/dosgolem/'+package+'\t' in core, package
        assert not (module/'workplace'/logs[2]).read_text()
        ui = load(p/'tools/direct_ui_check.py')
        frontend = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt', 'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
        rows = []
        for mode, suffix in (('hd', 'play-v1/hd'), ('wall', 'ui-v1/wall-direct')):
            folder = p/('workplace/probe/g9-hd/nativesurfacefillspans-'+suffix)
            prior = p/('workplace/probe/g9-hd/nativeprimitivespans-'+suffix)
            ci, bi = read(folder/'inputs.json'), read(prior/'inputs.json')
            assert ci['binary_sha256'] == result['binary_sha256'] and bi['binary_sha256'] == baseline['binary_sha256']
            assert all(ci[key] == bi[key] for key in ('orig', 'font', 'text', 'theme'))
            assert ui.endpoint(folder) == ui.endpoint(prior)
            assert ui.audit(folder, mode) == ui.audit(prior, mode) and ui.audit(folder, mode)['unexpected_flash_count'] == 0
            guests = [{k: v for k, v in tree(d/'save').items() if k not in frontend} for d in (folder, prior)]
            assert guests[0] == guests[1]
            for name in ('source.idx', 'source.pal', 'original.rgba', 'frame.png'):
                assert (folder/name).read_bytes() == (prior/name).read_bytes()
            for name in ('shots', 'window-shots', 'sources'):
                assert tree(folder/name) and tree(folder/name) == tree(prior/name)
            assert read(folder/'draw-report.json')['Mismatches'] == 0
            assert tree(p/'workplace/hd/theme') == ci['theme']
            rows.append(dict(route=mode, endpoint=ui.endpoint(folder), screenshot_pairs=len(tree(folder/'shots')),
                             screenshots_and_original_state_equal=True, unexpected_flash_count=0))
        result['normal_samples'] = rows
        result['regression_logs'] = {name: sha(module/'workplace'/name) for name in logs}
        if (raw/'activation.json').exists():
            activation = read(raw/'activation.json')
            assert sha(module/'workplace/bin/pod-play') == activation['candidate_sha256'] == result['binary_sha256']
            assert sha(module/'workplace/bin'/activation['previous_binary']) == activation['previous_sha256'] == baseline['binary_sha256']
            result['activation'] = activation
    result['tool_sha256'] = sha(Path(__file__))
    result['helper_sha256'] = sha(p/'tools/native_palette_check.py')
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({key: result[key] for key in ('algorithmic_plans', 'fixed_semantic_instructions', 'guest_files', 'normal_surface_fill_pixel_exit_delta')}))


if __name__ == '__main__':
    main()
