#!/usr/bin/env python3
"""驗證 Go 單點、填色跨度及斜線單步；受控原版與正常前端範圍分列。"""
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
    names = ('point-whole-initial', 'region-whole-initial', 'region-rows-initial',
             'diagonal-cursor-initial', 'primitive-short-diagnostic', 'primitive-spans-initial', 'primitive-spans-ready')
    anchors = {name: f'<a id="{name}"></a>' for name in names}
    assert all(text.count(anchor) == 1 for anchor in anchors.values())
    section = lambda name: text.split(anchors[name])[1].split('<a id=')[0]
    marker = 'PRIMITIVE-PIXEL-READY-20261029'
    ready = section('primitive-spans-ready')
    keys = ['DOS/VGA.DRV/eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f:0x'+address
            for address in ('2B10', '3338', '2CFC')]
    assert marker in ready and all(key in ready for key in keys)
    for name in names[:-1]:
        assert marker in section(name) and '(#primitive-spans-ready)' in section(name)
        assert '(#'+name+')' in ready
    return dict(keys=keys, correction=marker)


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
    raw = p/'workplace/probe/draw-intercept-v1/native-primitive-spans-v1'
    baseline_path = p/'docs/re/019-native-axis-scale-runtime.json'
    baseline = read(baseline_path)
    git = lambda *args: subprocess.check_output(['git', *args], cwd=module, text=True).strip()
    assert git('rev-parse', 'HEAD') == a.commit and not git('status', '--porcelain')
    parent = git('rev-parse', 'HEAD^')
    assert parent == baseline['program_commit']
    names = ('effects.go', 'native_line_cursor.go', 'native_primitive_pixels.go',
             'native_primitive_spans_test.go', 'native_primitive_spans_guards_test.go', 'native_primitive_clip_test.go')
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
    for lo, length, digest in ((0x2b10, 244, '5ac8a4f4342f02d70c6cf19c77b91d50002271c85a437fc5830e351be6634705'),
                               (0x3338, 306, '48c54ccea05f6f319457cf4bcc454a759cb81650a14c89b26b722b3a3fce7a3e'),
                               (0x2cfc, 1338, '06d69c6ecaffd16f3d050756a9485ea703e3bb67821f3f9c6f62f45772b46fba')):
        assert hashlib.sha256(original.read_bytes()[lo:lo+length]).hexdigest() == digest
    log = module/'workplace/native-primitive-spans-product-tests.log'
    body = log.read_text()
    for marker in (
        'line cursor original comparisons 280',
        'primitive original clipping-count quirk SaveState+PortLog cases=6 original pixels=62 maxX=20 selected right=10 CPU/fixed primitive pixels=0',
        'primitive positive phases=7 source/frame/cursor/model refusals=146 unchanged',
        'primitive clipped subroot/eight-parent whole SaveState+PortLog cases=72 CPU/fixed primitive pixels=0',
        'primitive full-canvas region SaveState+PortLog cases=4 CPU/fixed primitive pixels=0',
        'primitive ordered original point=28 region=200 diagonal=4176 zero-count=60 pixel phases=7',
        'primitive whole SaveState+PortLog cases=112 Go plans=1404 fixed control=872 CPU control=14972 CPU/fixed primitive pixels=0',
    ):
        assert marker in body, marker
    assert '\nPASS\n' in body
    paths = [raw/'trace-native', p/'workplace/probe/draw-intercept-v1/native-axis-scale-v1/trace-native',
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
    pixel_ips = {0x2bfc, 0x345c, 0x3460, 0x309c, 0x30b9, 0x3206, 0x321c}
    observed = {key: value for key, value in counts.items() if key.startswith('48a1:') and int(key.split(':')[1], 16) in pixel_ips}
    delta = {key: [value[i]-previous['AlgorithmicPreparedByExit'].get(key, [0, 0])[i] for i in range(2)]
             for key, value in observed.items()}
    fixed = current['HostInstructions']+current['BlockInstructions']-current['AlgorithmicInstructions']
    assert fixed <= baseline['fixed_semantic_instructions']
    result = dict(scope='Go單點、填色與斜線單步像素；其他繪圖與未驗條件仍待盤點',
        program_commit=a.commit, parent_commit=parent,
        binary_sha256=sha(module/'workplace/bin/pod-play-nativeprimitivespans'),
        original_state_and_ordered_events_equal=True, virtual_steps=current['Digest']['Steps'],
        guest_files=len(tree(paths[0]/'save')), guest_files_equal=True,
        direct_original_witnesses=len(current['Captures'])+len(current['DrawCaptures']), original_witnesses_equal=True,
        normal_primitive_pixel_exit_counts=observed, normal_primitive_pixel_exit_delta=delta,
        algorithmic_plans=current['AlgorithmicPlans'], algorithmic_instructions=current['AlgorithmicInstructions'], fixed_semantic_instructions=fixed,
        controlled_original=dict(point_ordered=28, region_ordered=200, diagonal_ordered=4176, zero_count=60, pixel_phases=7,
            whole_calls=112, whole_go_plans=1404, whole_fixed_control=872, whole_cpu_control=14972,
            parent_calls=72, full_canvas_calls=4, original_clip_quirk_calls=6, refusals=146,
            existing_line_cursor_plans=280, cpu_or_fixed_primitive_pixels=0),
        source_sha256={name: sha(module/name) for name in expected},
        product_test_log_sha256=sha(log), resolution_backlinks=resolution(p/'docs/spec/025-direct-render-commands.md'),
        spec_sha256=sha(p/'docs/spec/025-direct-render-commands.md'), prior_full_receipt_sha256=sha(baseline_path),
        build_info_sha256=sha(raw/'build.txt'), trace_binary_sha256=sha(module/'workplace/bin/draw-call-probe-nativeprimitivespans'),
        source_reports={str(path.relative_to(p)): sha(path/'calls.json') for path in paths},
        prior_gates_note='前批完整四路demo、39對介面、6對存讀檔及demo中斷另存原收據，本批正常抽樣另列')
    if a.stage == 'full':
        logs = ['native-primitive-spans-all-drawblocks-r2.log', 'native-primitive-spans-core-tests-r2.log', 'native-primitive-spans-vet-r2.log']
        for name in logs[:2]:
            text = (module/'workplace'/name).read_text()
            assert 'FAIL' not in text and 'ok  ' in text
        assert not (module/'workplace'/logs[2]).read_text()
        ui = load(p/'tools/direct_ui_check.py')
        frontend = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt', 'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
        rows = []
        for mode, suffix in (('hd', 'play-v1/hd'), ('wall', 'ui-v1/wall-direct')):
            folder = p/('workplace/probe/g9-hd/nativeprimitivespans-'+suffix)
            prior = p/('workplace/probe/g9-hd/nativeaxisscale-'+suffix)
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
    print(json.dumps({key: result[key] for key in ('algorithmic_plans', 'fixed_semantic_instructions', 'guest_files', 'normal_primitive_pixel_exit_delta')}))


if __name__ == '__main__':
    main()
