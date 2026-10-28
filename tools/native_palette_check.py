#!/usr/bin/env python3
"""驗證 Go 淡入、指定色盤續行及分段 DAC；原版與正常前端範圍分列。"""
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


def resolution(path):
    text = path.read_text()
    rows = []
    for initial, ready, marker, entry in (
        ('rgb-output-full-initial', 'rgb-output-bounded-ready', 'RGB-OUTPUT-BOUNDED-READY-20261028', '2AA8'),
        ('palette-in-initial', 'palette-in-ready', 'PALETTE-IN-READY-20261028', '45E0'),
    ):
        a, b = f'<a id="{initial}"></a>', f'<a id="{ready}"></a>'
        assert text.count(a) == text.count(b) == 1
        before, after = text.split(a)[1].split(b)
        key = 'DOS/VGA.DRV/eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f:0x'+entry
        assert marker in before and '(#'+ready+')' in before
        assert marker in after and key in after and '(#'+initial+')' in after
        rows.append(dict(key=key, correction=marker))
    return dict(rows=rows, spec_sha256=sha(path))


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--commit', required=True)
    ap.add_argument('--stage', choices=('trace', 'full'), default='full')
    ap.add_argument('--out', type=Path, required=True)
    a = ap.parse_args()
    assert not a.out.exists(), '保留既有收據，請另定新檔名'
    p = a.project
    module = p/'workplace/dosgolem'
    raw = p/'workplace/probe/draw-intercept-v1/native-palettein-v1'
    baseline_path = p/'docs/re/019-native-glyph-runtime.json'
    baseline = read(baseline_path)
    git = lambda *args: subprocess.check_output(['git', *args], cwd=module, text=True).strip()
    assert git('rev-parse', 'HEAD') == a.commit and not git('status', '--porcelain')
    parent = git('rev-parse', 'HEAD^')
    assert parent == baseline['program_commit']
    changed = git('diff', '--name-only', parent, a.commit).splitlines()
    assert len(changed) == 21
    code = [name for name in changed if name.endswith('.go')]
    assert len(code) == 19 and all(name.startswith(('apps/pod/drawblocks/', 'internal/machine/')) for name in code)
    for name in code:
        if not name.endswith('_test.go') and not name.endswith('/effects.go'):
            source = (module/name).read_text()
            assert all(token not in source for token in ('.Step(', 'effect{', 'b.run(', 'DecodeOpcode'))
    for name, digest in baseline['source_sha256'].items():
        if name != 'effects.go':
            assert sha(module/'apps/pod/drawblocks'/name) == digest
    build = (raw/'build.txt').read_text()
    assert 'go1.24.13' in build and 'vcs.modified=false' in build and 'vcs.revision='+a.commit in build
    original = p/'workplace/input/dos/pod/VGA.DRV'
    assert sha(original) == 'eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f'
    for lo, hi, digest in (
        (0x45e0, 0x4944, 'd6173666ef045fadd9565b60f25e0bafb8dbe721494143c513f91529387ecbda'),
        (0x2aa8, 0x2ac3, '7134f5f3fbaf2adf007afa01632ff5ec980bd4c127a095defaa954fd831b4801'),
    ):
        assert hashlib.sha256(original.read_bytes()[lo:hi]).hexdigest() == digest
    log = module/'workplace/native-palettein-product-tests.log'
    body = log.read_text()
    for marker in (
        'fade-in RGB difference cases=24 original plans=96',
        'fade-in RGB blend cases=32 original plans=176',
        'fade-in shared scan=32 list=24 target=24 error=32 original plans',
        'one-step original comparisons=116 distinct phases=16',
        'SaveState+PortLog cases=20 Go plans=101682 fixed control=10050 CPU control=168042 CPU/fixed selected palette writes=0; CPU DAC OUT=0',
        'boundaries=22 source/mode/DF/budget refusals=110 unchanged',
        'zero-distance whole SaveState+PortLog cases=4 minimum budget=1',
        'whole colored SaveState+PortLog cases=6 Go plans=14434 fixed control=5048 CPU control=10542',
        'bounded DAC original prefix/device-phase comparisons=1152',
        'bounded DAC whole original SaveState+PortLog cases=72 Go plans=522 CPU fallback=0',
        'bounded DAC CF/IF/DF whole SaveState+PortLog cases=36 Go plans=252',
        'bounded DAC source/frame/model/resume refusals=23 unchanged',
    ):
        assert marker in body, marker
    assert '\nPASS\n' in body
    paths = [raw/'trace-native', p/'workplace/probe/draw-intercept-v1/native-glyph-v1/trace-native',
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
    fixed = current['HostInstructions']+current['BlockInstructions']-current['AlgorithmicInstructions']
    assert fixed <= baseline['fixed_semantic_instructions']
    observed = {key: value for key, value in counts.items() if key.startswith('48a1:') and
                0x45e0 <= int(key.split(':')[1], 16) < 0x4944}
    rgb_counts = {key: value for key, value in counts.items() if key.startswith('48a1:') and
                  0x2aa8 <= int(key.split(':')[1], 16) < 0x2ac3}
    rgb_delta = {key: [value[i]-previous['AlgorithmicPreparedByExit'].get(key, [0, 0])[i] for i in range(2)]
                 for key, value in rgb_counts.items()}
    result = dict(scope='Go淡入共用色盤算法、指定16個單步寫入及分段DAC；其他繪圖仍待完成',
        program_commit=a.commit, parent_commit=parent,
        binary_sha256=sha(module/'workplace/bin/pod-play-nativepalettein'),
        original_state_and_ordered_events_equal=True, virtual_steps=current['Digest']['Steps'],
        guest_files=len(tree(paths[0]/'save')), guest_files_equal=True,
        direct_original_witnesses=len(current['Captures'])+len(current['DrawCaptures']), original_witnesses_equal=True,
        normal_fade_in_entries=observed, normal_rgb_output_exit_counts=rgb_counts,
        normal_rgb_output_exit_delta=rgb_delta,
        normal_scope='新增入口只在下列實際觀測數內成立；受控比較不外推未走的玩家場景',
        algorithmic_plans=current['AlgorithmicPlans'], algorithmic_instructions=current['AlgorithmicInstructions'],
        fixed_semantic_instructions=fixed,
        controlled_original=dict(difference_plans=96, blend_plans=176, setup_plans=112, single_phase_plans=116,
            colorful_whole_calls=6, short_budget_whole_calls=20, zero_distance_whole_calls=4, palette_refusals=110,
            selected_palette_write_phases=16, cpu_or_fixed_selected_palette_writes=0, cpu_dac_output=0,
            dac_prefix_comparisons=1152, dac_whole_calls=108, dac_whole_go_plans=774, dac_refusals=23),
        source_sha256={name: sha(module/name) for name in changed},
        product_test_log_sha256=sha(log), resolution_backlinks=resolution(p/'docs/spec/025-direct-render-commands.md'),
        prior_full_receipt_sha256=sha(baseline_path), build_info_sha256=sha(raw/'build.txt'),
        trace_binary_sha256=sha(module/'workplace/bin/draw-call-probe-nativepalettein'),
        source_reports={str(path.relative_to(p)): sha(path/'calls.json') for path in paths},
        prior_gates_note='前版完整四路demo、39對介面、6對存讀檔及demo中斷另存原收據，本批正常抽樣另列')
    if a.stage == 'full':
        logs = ['native-palettein-all-drawblocks.log', 'native-palettein-core-tests-display.log', 'native-palettein-vet-display.log']
        for name in logs[:2]:
            text = (module/'workplace'/name).read_text()
            assert 'FAIL' not in text and 'ok  ' in text
        assert not (module/'workplace'/logs[2]).read_text()
        spec = importlib.util.spec_from_file_location('ui', p/'tools/direct_ui_check.py')
        ui = importlib.util.module_from_spec(spec); spec.loader.exec_module(ui)
        rows = []
        frontend = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt', 'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
        for mode, suffix in (('hd', 'play-v1/hd'), ('wall', 'ui-v1/wall-direct')):
            folder = p/('workplace/probe/g9-hd/nativepalettein-'+suffix)
            prior = p/('workplace/probe/g9-hd/nativeglyph-'+suffix)
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
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({key: result[key] for key in ('algorithmic_plans', 'fixed_semantic_instructions', 'guest_files', 'normal_fade_in_entries', 'normal_rgb_output_exit_delta')}))


if __name__ == '__main__':
    main()
