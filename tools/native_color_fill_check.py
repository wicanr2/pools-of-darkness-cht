#!/usr/bin/env python3
"""核對共享VGA填色續行的原版、正常路線及介面收據。須在Docker執行。"""
import argparse
import hashlib
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


def verify(project, stage):
    module = project/'workplace/dosgolem'
    raw = project/'workplace/probe/draw-intercept-v1/native-color-fill-v1'
    baseline = read(project/'docs/re/019-native-resident-glyph-runtime.json')
    git = lambda *args: subprocess.check_output(['git', *args], cwd=module, text=True).strip()
    commit = git('rev-parse', 'HEAD')
    assert git('rev-parse', 'HEAD^') == baseline['program_commit']
    assert not git('status', '--porcelain')
    names = ['apps/pod/drawblocks/'+n for n in ('effects.go', 'native_color_fill_span.go', 'native_color_fill_span_test.go')]
    assert sorted(git('diff', '--name-only', 'HEAD^', 'HEAD').splitlines()) == sorted(names)
    source = (module/names[1]).read_text()
    assert all(token not in source for token in ('.Step(', 'b.run(', 'effect{', 'DecodeOpcode', '.Write8(', '.Write16('))
    build = (raw/'build.txt').read_text()
    assert 'go1.24.13' in build and 'vcs.modified=false' in build and 'vcs.revision='+commit in build
    log = module/'workplace/color-fill-product.log'
    assert 'color-fill ordered original spans=50 whole SaveState+PortLog=250 Go plans=250 CPU/fixed VGA pixels=0' in log.read_text()
    assert 'color-fill unchanged refusals=19' in log.read_text()
    logs = [module/'workplace'/f'color-fill-{name}.log' for name in ('regression', 'related', 'frontend', 'vet', 'product')]
    for path in logs:
        text = path.read_text()
        assert 'FAIL' not in text
        if path.name != 'color-fill-vet.log':
            assert 'ok  ' in text
    for package in ('internal/machine', 'session', 'apps/pod', 'apps/pod/bgm', 'xlate'):
        assert 'github.com/wicanr2/dosgolem/'+package+'\t' in logs[1].read_text()
    assert not logs[3].read_text()
    spec = (project/'docs/spec/025-direct-render-commands.md').read_text()
    anchors = ('shared-fill-whole-initial', 'color-fill-span-ready')
    for name in anchors:
        assert spec.count(f'<a id="{name}"></a>') == 1
    old, new = [spec.split(f'<a id="{name}"></a>')[1].split('<a id=')[0] for name in anchors]
    marker = 'COLOR-FILL-SPAN-READY-20261029'
    assert marker in old and marker in new and '(#color-fill-span-ready)' in old and '(#shared-fill-whole-initial)' in new
    assert 'runtime:1ED4:1A7F' in new and 'f1c5fade96970aa1dd1a9ffc2d876c7aea57bef8680d929d9f65d6c5190e17f7' in new
    paths = [raw/'trace-native-v3', project/'workplace/probe/draw-intercept-v1/native-resident-glyph-v1/trace-native',
             project/'workplace/probe/draw-intercept-v1/native-fade-v1/trace-original-entries']
    current, previous, oracle = [read(p/'calls.json') for p in paths]
    for field in ('Calls', 'Digest', 'VideoCount', 'AudioCount', 'FrameCount', 'DOSCount', 'VideoSHA256', 'AudioSHA256', 'FramesSHA256', 'DOSSHA256'):
        assert current[field] == previous[field] == oracle[field], field
    assert tree(paths[0]/'save') == tree(paths[1]/'save') == tree(paths[2]/'save')
    for field in ('Captures', 'DrawCaptures'):
        assert current[field] == previous[field]
        for row in current[field]:
            assert sha(paths[0]/row['Memory']) == sha(paths[1]/row['Memory']) == row['MemorySHA256']
    exits = current['AlgorithmicPreparedByExit']
    assert exits['1ed4:1a8e'] == [15, 15] and '1ed4:1a8e' not in previous['AlgorithmicPreparedByExit']
    for key in set(exits) | set(previous['AlgorithmicPreparedByExit']):
        if key != '1ed4:1a8e':
            assert exits.get(key) == previous['AlgorithmicPreparedByExit'].get(key), key
    assert current['AlgorithmicPlans']-previous['AlgorithmicPlans'] == 15
    assert current['AlgorithmicInstructions']-previous['AlgorithmicInstructions'] == 15
    assert sum(row[0] for row in exits.values()) == current['AlgorithmicPlans']
    assert sum(row[1] for row in exits.values()) == current['AlgorithmicInstructions']
    commands = current['Commands']
    assert commands['HostBytes'] == commands['NativeBytes'] == current['VideoCount'] == 3343240
    assert commands['HostBytes']-previous['Commands']['HostBytes'] == 1920
    assert commands['FallbackBytes'] == commands['Mismatches'] == 0
    assert all(row['HostBytes'] == row['NativeBytes'] and row['FallbackBytes'] == 0 for row in commands['Writers'].values())
    fixed = current['HostInstructions']+current['BlockInstructions']-current['AlgorithmicInstructions']
    assert fixed == baseline['fixed_semantic_instructions'] == 1018
    pixels = read(raw/'trace-audit/color-pixels.json')
    declined = read(raw/'trace-audit/color-entry-declines.json')
    visible = [x for x in pixels if x['Segment'] == 0xa000]
    assert len(visible) == 15 and sum(x['Count'] for x in visible) == 1920
    diagnosis = []
    for x in visible:
        matches = [e for e in declined if e['Step'] < x['Step'] and all(e[k] == x[k] for k in ('SS', 'SP', 'Color', 'Count', 'Offset', 'Segment'))]
        entry = max(matches, key=lambda e:e['Step'])
        assert entry['Budget'] < 7
        diagnosis.append(dict(pixel_step=x['Step'], entry_budget=entry['Budget'], step_gap=x['Step']-entry['Step'], bytes=x['Count']))
    result = dict(scope='共享填色1ED4:1A8E的已認證VGA跨度；完整Go繪圖仍DRAFT', program_commit=commit,
        binary_sha256=sha(module/'workplace/bin/pod-play-nativecolorfill'), baseline_commit=baseline['program_commit'],
        source_sha256={n:sha(module/n) for n in names}, controlled_original=dict(ordered_spans=50, whole_calls=250, refusals=19, cpu_fixed_vga_pixels=0),
        normal=dict(steps=current['Digest']['Steps'], original_state_events_equal=True, guest_files=len(tree(paths[0]/'save')),
                    direct_witnesses=len(current['Captures'])+len(current['DrawCaptures']), new_go_spans=15, new_go_bytes=1920,
                    host_video_bytes=commands['HostBytes'], native_video_bytes=commands['NativeBytes'],
                    algorithmic_plans=current['AlgorithmicPlans'], algorithmic_instructions=current['AlgorithmicInstructions'], fixed_instructions=fixed),
        diagnosis=diagnosis, logs_sha256={str(p.relative_to(module)):sha(p) for p in logs},
        trace_sha256={str(p.relative_to(project)):sha(p/'calls.json') for p in paths}, build_sha256=sha(raw/'build.txt'),
        audit_sha256={n:sha(raw/'trace-audit'/n) for n in ('calls.json','color-pixels.json','color-entry-declines.json','color-cpu-commands.json')},
        prior_receipt_sha256=sha(project/'docs/re/019-native-resident-glyph-runtime.json'),
        original_exe_sha256=sha(project/'workplace/input/dos/pod/GAME.EXE'),
        entry_table_sha256=sha(project/'workplace/probe/draw-intercept-v1/native-fade-v1/fade-entries.json'),
        probe_sha256=sha(module/'workplace/bin/draw-call-probe-nativecolorfill'),
        probe_build_sha256=sha(raw/'probe-build.txt'),
        spec_sha256=sha(project/'docs/spec/025-direct-render-commands.md'))
    if stage == 'full':
        rows = []
        for current_name, previous_name, flash in (('nativecolorfill-play-v1/hd','nativeresidentglyph-play-v2/hd',0),
                                                  ('nativecolorfill-font-v1/wall-direct','nativeresidentglyph-font-v2/wall-direct',121)):
            root = project/'workplace/probe/g9-hd'
            now, before = root/current_name, root/previous_name
            images = [p for p in now.rglob('*.png') if p.is_file()]
            assert images and sorted(str(p.relative_to(now)) for p in images) == sorted(str(p.relative_to(before)) for p in before.rglob('*.png'))
            for name in ('shots','window-shots','sources','save'):
                assert tree(now/name) == tree(before/name), name
            for name in ('frame.png','source.idx','source.pal','source.glyphs.json','original.rgba','script.txt'):
                assert sha(now/name) == sha(before/name), name
            for pattern in (r'frames=\d+ steps=\d+ memory_sha256=\w+ indexed_sha256=\w+', r'cpu_sha256=\w+ palette_sha256=\w+', r'閃現.*'):
                assert re.findall(pattern,(now/'run.log').read_text()) == re.findall(pattern,(before/'run.log').read_text()), pattern
            assert f'閃現 {flash} 次' in (now/'run.log').read_text()
            inputs, previous_inputs = read(now/'inputs.json'), read(before/'inputs.json')
            assert inputs['binary_sha256'] == result['binary_sha256']
            for key in ('orig','font','text','theme'):
                if key in inputs:
                    assert inputs[key] == previous_inputs[key], key
            report = read(now/'draw-report.json')
            assert report['HostBytes'] == report['NativeBytes'] and report['FallbackBytes'] == report['Mismatches'] == 0
            rows.append(dict(route=current_name, pngs=len(images), png_sha256={str(p.relative_to(now)):sha(p) for p in images},
                             guest_files_equal=True, original_state_equal=True, raw_flash=flash, unexpected_flash=0,
                             inputs_sha256=sha(now/'inputs.json'), log_sha256=sha(now/'run.log'), report_sha256=sha(now/'draw-report.json')))
        result['frontend'] = rows
        activation_path = raw/'activation.json'
        if activation_path.exists():
            activation = read(activation_path)
            assert activation['candidate_sha256'] == result['binary_sha256'] == sha(module/'workplace/bin/pod-play')
            assert activation['previous_sha256'] == baseline['binary_sha256'] == sha(module/'workplace/bin'/activation['previous_binary'])
            result['activation'] = activation
    result['tool_sha256'] = sha(Path(__file__))
    return result


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--stage', choices=('trace','full'), default='full')
    ap.add_argument('--out', type=Path, required=True)
    a = ap.parse_args()
    assert not a.out.exists(), '既有收據不得覆寫'
    result = verify(a.project, a.stage)
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(result['normal'], ensure_ascii=False))


if __name__ == '__main__':
    main()
