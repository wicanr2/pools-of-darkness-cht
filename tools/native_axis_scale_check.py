#!/usr/bin/env python3
"""綁定橫直線單步與寬1縮放的原版證據、來源及正常前端抽樣。"""
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
    names = ('axis-spans-initial', 'axis-spans-ready', 'narrow-scale-ready',
             'scale5615-initial', 'scale5615-conformed-initial', 'degenerate-initial')
    anchors = {name: f'<a id="{name}"></a>' for name in names}
    assert all(text.count(anchor) == 1 for anchor in anchors.values())
    section = lambda name: text.split(anchors[name])[1].split('<a id=')[0]
    am, nm = 'AXIS-SPANS-READY-20261029', 'NARROW-SCALE-READY-20261029'
    ax, narrow = section('axis-spans-ready'), section('narrow-scale-ready')
    base = 'DOS/VGA.DRV/eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f:0x'
    assert am in ax and base+'2CFC' in ax and '(#axis-spans-initial)' in ax and '(#degenerate-initial)' in ax
    assert nm in narrow and base+'5615' in narrow
    assert '(#scale5615-initial)' in narrow and '(#scale5615-conformed-initial)' in narrow and '(#degenerate-initial)' in narrow
    assert am in section('axis-spans-initial') and '(#axis-spans-ready)' in section('axis-spans-initial')
    for name in ('scale5615-initial', 'scale5615-conformed-initial'):
        assert nm in section(name) and '(#narrow-scale-ready)' in section(name)
    old = section('degenerate-initial')
    assert all(token in old for token in (am, nm, '(#axis-spans-ready)', '(#narrow-scale-ready)'))
    return dict(axis=dict(key=base+'2CFC', correction=am), narrow=dict(key=base+'5615', correction=nm))


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--commit', required=True)
    ap.add_argument('--stage', choices=('trace', 'full'), default='full')
    ap.add_argument('--out', type=Path, required=True)
    a = ap.parse_args()
    assert not a.out.exists(), '請保留既有收據，另用新檔名'
    p = a.project
    helper = load(p/'tools/native_palette_check.py')
    sha, read, tree = helper.sha, helper.read, helper.tree
    module = p/'workplace/dosgolem'
    raw = p/'workplace/probe/draw-intercept-v1/native-axis-scale-v1'
    baseline_path = p/'docs/re/019-native-surface-spans-runtime.json'
    baseline = read(baseline_path)
    git = lambda *args: subprocess.check_output(['git', *args], cwd=module, text=True).strip()
    assert git('rev-parse', 'HEAD') == a.commit and not git('status', '--porcelain')
    parent = git('rev-parse', 'HEAD^')
    assert parent == baseline['program_commit']
    names = ('effects.go', 'native_axis_spans.go', 'native_axis_spans_test.go', 'native_axis_spans_guards_test.go',
             'native_scale5615_frame.go', 'native_scale5615_prefix.go', 'native_scale5615_prepare.go',
             'native_scale5615_refusals_test.go', 'native_scale5615_narrow_whole_test.go',
             'native_scale5615_narrow_boundaries_test.go', 'native_scale5615_narrow_edges_test.go')
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
    for lo, length, digest in ((0x2cfc, 1338, '06d69c6ecaffd16f3d050756a9485ea703e3bb67821f3f9c6f62f45772b46fba'),
                               (0x5615, 625, '6d9aac63d38f8b0e5ddd9bdbdfcfdda9144318a34398436b6b468c52f0271106')):
        assert hashlib.sha256(original.read_bytes()[lo:lo+length]).hexdigest() == digest
    log = module/'workplace/native-axis-scale-product-tests.log'
    body = log.read_text()
    for marker in (
        'axis span independently reached original comparisons=3504 zero-count=48 pixel phases=10',
        'axis span whole SaveState+PortLog cases=96 Go plans=2652 fixed control=466 CPU control=14398 CPU/fixed axis pixels=0',
        'zero-radius outline/filled whole SaveState+PortLog cases=60 Go plans=1708 fixed control=344 CPU control=11450 CPU/fixed axis pixels=0',
        'axis span positive phases=10 source/frame/cursor/model refusals=202 unchanged',
        'axis span clipped subroot/eight-parent whole SaveState+PortLog cases=60 CPU/fixed axis pixels=0',
        'width1 scale original prefix=72 prepare=180 output=180 single-span=924 zero-count=552 distinct phases=5',
        'width1 scale positive phases=7 source/frame/grid refusals=77 unchanged',
        'width1 full height/eight-parent/overlap/stale-buffer SaveState+PortLog cases=16 CPU/fixed row preparation/pixels=0',
        'width1 scale whole SaveState+PortLog cases=288 Go plans=2880 fixed control=2058 CPU control=38610 CPU/fixed row preparation/pixels=0',
        'scale5615 invalid state refusals 88',
    ):
        assert marker in body, marker
    assert '\nPASS\n' in body
    paths = [raw/'trace-native', p/'workplace/probe/draw-intercept-v1/native-surface-spans-v1/trace-native',
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
    observed = {key: value for key, value in counts.items() if key.startswith('48a1:') and
                (0x2cfc <= int(key.split(':')[1], 16) < 0x3236 or 0x5615 <= int(key.split(':')[1], 16) < 0x5886)}
    fixed = current['HostInstructions']+current['BlockInstructions']-current['AlgorithmicInstructions']
    assert fixed <= baseline['fixed_semantic_instructions']
    result = dict(scope='Go橫直線單步與目的寬1整數縮放，含零半徑轉點線；其他繪圖仍待完成',
        program_commit=a.commit, parent_commit=parent,
        binary_sha256=sha(module/'workplace/bin/pod-play-nativeaxisscale'),
        original_state_and_ordered_events_equal=True, virtual_steps=current['Digest']['Steps'],
        guest_files=len(tree(paths[0]/'save')), guest_files_equal=True,
        direct_original_witnesses=len(current['Captures'])+len(current['DrawCaptures']), original_witnesses_equal=True,
        normal_axis_scale_entries=observed, algorithmic_plans=current['AlgorithmicPlans'],
        algorithmic_instructions=current['AlgorithmicInstructions'], fixed_semantic_instructions=fixed,
        controlled_original=dict(axis_ordered_comparisons=3504, axis_zero_count=48, axis_phases=10,
            axis_whole_calls=96, axis_go_plans=2652, axis_fixed_control=466, axis_cpu_control=14398,
            zero_ellipse_calls=60, zero_ellipse_go_plans=1708, zero_ellipse_fixed_control=344, zero_ellipse_cpu_control=11450,
            axis_parent_calls=60, axis_refusals=202, cpu_or_fixed_axis_pixels=0,
            narrow_prefix=72, narrow_prepare=180, narrow_output=180, narrow_spans=924, narrow_zero_count=552,
            narrow_whole_calls=288, narrow_go_plans=2880, narrow_fixed_control=2058, narrow_cpu_control=38610,
            narrow_edge_calls=16, narrow_refusals=77, existing_scale_refusals=88, cpu_or_fixed_narrow_preparation_pixels=0),
        source_sha256={name: sha(module/name) for name in expected},
        product_test_log_sha256=sha(log), resolution_backlinks=resolution(p/'docs/spec/025-direct-render-commands.md'),
        spec_sha256=sha(p/'docs/spec/025-direct-render-commands.md'), prior_full_receipt_sha256=sha(baseline_path),
        build_info_sha256=sha(raw/'build.txt'), trace_binary_sha256=sha(module/'workplace/bin/draw-call-probe-nativeaxisscale'),
        source_reports={str(path.relative_to(p)): sha(path/'calls.json') for path in paths},
        prior_gates_note='前批完整四路demo、39對介面、6對存讀檔及demo中斷另存原收據，本批正常抽樣另列')
    if a.stage == 'full':
        logs = ['native-axis-scale-all-drawblocks.log', 'native-axis-scale-core-tests.log', 'native-axis-scale-vet.log']
        for name in logs[:2]:
            text = (module/'workplace'/name).read_text()
            assert 'FAIL' not in text and 'ok  ' in text
        assert not (module/'workplace'/logs[2]).read_text()
        ui = load(p/'tools/direct_ui_check.py')
        frontend = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt', 'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
        rows = []
        for mode, suffix in (('hd', 'play-v1/hd'), ('wall', 'ui-v1/wall-direct')):
            folder = p/('workplace/probe/g9-hd/nativeaxisscale-'+suffix)
            prior = p/('workplace/probe/g9-hd/nativesurfacespans-'+suffix)
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
    print(json.dumps({key: result[key] for key in ('algorithmic_plans', 'fixed_semantic_instructions', 'guest_files', 'normal_axis_scale_entries')}))


if __name__ == '__main__':
    main()
