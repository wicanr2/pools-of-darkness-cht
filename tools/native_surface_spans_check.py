#!/usr/bin/env python3
"""驗證最少一步的 Go 表面搬移跨度及正常前端；保留原版範圍限制。"""
import argparse
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
    old, initial, ready = [f'<a id="{name}"></a>' for name in
                           ('surface-copy-initial', 'surface-copy-span-initial', 'surface-copy-span-ready')]
    assert all(text.count(anchor) == 1 for anchor in (old, initial, ready))
    before, middle = text.split(old)[1].split(initial)
    middle, after = middle.split(ready)
    marker = 'SURFACE-COPY-SPAN-READY-20261028'
    key = 'DOS/VGA.DRV/eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f:0x3F5E'
    assert all(marker in part for part in (before, middle, after))
    assert '(#surface-copy-span-ready)' in before and '(#surface-copy-span-ready)' in middle
    assert '(#surface-copy-initial)' in after and '(#surface-copy-span-initial)' in after and key in after
    return dict(key=key, correction=marker)


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--commit', required=True)
    ap.add_argument('--stage', choices=('trace', 'full'), default='full')
    ap.add_argument('--out', type=Path, required=True)
    a = ap.parse_args()
    assert not a.out.exists(), '請保留既有收據，使用新檔名'
    p = a.project
    helper = load(p/'tools/native_palette_check.py')
    sha, read, tree = helper.sha, helper.read, helper.tree
    module = p/'workplace/dosgolem'
    raw = p/'workplace/probe/draw-intercept-v1/native-surface-spans-v1'
    baseline_path = p/'docs/re/019-native-palettein-runtime.json'
    baseline = read(baseline_path)
    git = lambda *args: subprocess.check_output(['git', *args], cwd=module, text=True).strip()
    assert git('rev-parse', 'HEAD') == a.commit and not git('status', '--porcelain')
    parent = git('rev-parse', 'HEAD^')
    assert parent == baseline['program_commit']
    changed = git('diff', '--name-only', parent, a.commit).splitlines()
    expected = ['apps/pod/drawblocks/'+name for name in (
        'effects.go', 'native_surface_copy.go', 'native_surface_copy_spans.go',
        'native_surface_copy_spans_test.go', 'native_surface_copy_spans_whole_test.go', 'native_surface_copy_short_test.go')]
    expected += ['docs/spec/000-index.md', 'docs/spec/436-host-call-effects.md']
    assert sorted(changed) == sorted(expected)
    for name in ('native_surface_copy.go', 'native_surface_copy_spans.go'):
        source = (module/'apps/pod/drawblocks'/name).read_text()
        assert all(token not in source for token in ('.Step(', 'effect{', 'b.run(', 'DecodeOpcode'))
    for name, digest in baseline['source_sha256'].items():
        if name not in expected:
            assert sha(module/name) == digest
    build = (raw/'build.txt').read_text()
    assert 'go1.24.13' in build and 'vcs.modified=false' in build and 'vcs.revision='+a.commit in build
    original = p/'workplace/input/dos/pod/VGA.DRV'
    assert sha(original) == 'eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f'
    import hashlib
    assert hashlib.sha256(original.read_bytes()[0x3f5e:0x41af]).hexdigest() == '559f234ab6f1b97b5c6970fc84c40982e5525fdcac06436fd355bf5ec8c8d4fa'
    log = module/'workplace/native-surface-spans-product-tests.log'
    body = log.read_text()
    for marker in (
        'surface-copy short SaveState cases=40 Go pixel bytes=240 fixed pixel bytes=0 CPU write instructions=0',
        'surface span independently reached ordered original comparisons=320 zero-count=96',
        'surface span source/frame/flags/model refusals=184 unchanged',
        'surface span whole SaveState+PortLog cases=400 Go plans=1360 fixed control=2380 CPU control=42960 CPU/fixed surface pixels=0',
        'surface span full canvas/eight parents/zero-origin backward SaveState cases=12 CPU/fixed surface pixels=0',
    ):
        assert marker in body, marker
    assert '\nPASS\n' in body
    paths = [raw/'trace-native', p/'workplace/probe/draw-intercept-v1/native-palettein-v1/trace-native',
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
    observed = {key: value for key, value in counts.items() if key.startswith('48a1:') and 0x3f5e <= int(key.split(':')[1], 16) < 0x41af}
    fixed = current['HostInstructions']+current['BlockInstructions']-current['AlgorithmicInstructions']
    assert fixed <= baseline['fixed_semantic_instructions']
    result = dict(scope='Go表面矩形搬移的一步word／尾byte跨度；完整VGA root與合法子表面，其他繪圖仍待完成',
        program_commit=a.commit, parent_commit=parent,
        binary_sha256=sha(module/'workplace/bin/pod-play-nativesurfacespans'),
        original_state_and_ordered_events_equal=True, virtual_steps=current['Digest']['Steps'],
        guest_files=len(tree(paths[0]/'save')), guest_files_equal=True,
        direct_original_witnesses=len(current['Captures'])+len(current['DrawCaptures']), original_witnesses_equal=True,
        normal_surface_entries=observed, algorithmic_plans=current['AlgorithmicPlans'],
        algorithmic_instructions=current['AlgorithmicInstructions'], fixed_semantic_instructions=fixed,
        controlled_original=dict(ordered_boundary_comparisons=320, zero_count_comparisons=96, refusals=184,
            whole_calls=400, whole_go_plans=1360, whole_fixed_control=2380, whole_cpu_control=42960,
            boundary_calls=12, diagnostic_calls=40, cpu_or_fixed_surface_pixels=0),
        source_sha256={name: sha(module/name) for name in changed},
        product_test_log_sha256=sha(log), resolution_backlinks=resolution(p/'docs/spec/025-direct-render-commands.md'),
        spec_sha256=sha(p/'docs/spec/025-direct-render-commands.md'),
        prior_full_receipt_sha256=sha(baseline_path), build_info_sha256=sha(raw/'build.txt'),
        trace_binary_sha256=sha(module/'workplace/bin/draw-call-probe-nativesurfacespans'),
        source_reports={str(path.relative_to(p)): sha(path/'calls.json') for path in paths},
        prior_gates_note='前批完整四路demo、39對介面、6對存讀檔及demo中斷另存原收據，本批正常抽樣另列')
    if a.stage == 'full':
        logs = ['native-surface-spans-all-drawblocks.log', 'native-surface-spans-core-tests.log', 'native-surface-spans-vet.log']
        for name in logs[:2]:
            text = (module/'workplace'/name).read_text()
            assert 'FAIL' not in text and 'ok  ' in text
        assert not (module/'workplace'/logs[2]).read_text()
        ui = load(p/'tools/direct_ui_check.py')
        frontend = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt', 'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
        rows = []
        for mode, suffix in (('hd', 'play-v1/hd'), ('wall', 'ui-v1/wall-direct')):
            folder = p/('workplace/probe/g9-hd/nativesurfacespans-'+suffix)
            prior = p/('workplace/probe/g9-hd/nativepalettein-'+suffix)
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
    print(json.dumps({key: result[key] for key in ('algorithmic_plans', 'fixed_semantic_instructions', 'guest_files', 'normal_surface_entries')}))


if __name__ == '__main__':
    main()
