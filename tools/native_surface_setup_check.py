#!/usr/bin/env python3
"""核對Go根畫布及子表面初始化；於Docker執行，原始RAM只留本機。"""
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


def verify(project, stage):
    module = project/'workplace/dosgolem'
    raw = project/'workplace/probe/draw-intercept-v1/native-fixed-control-v1'
    baseline_path = project/'docs/re/019-native-color-fill-runtime.json'
    baseline = read(baseline_path)
    git = lambda *args: subprocess.check_output(['git', *args], cwd=module, text=True).strip()
    commit = git('rev-parse', 'HEAD')
    assert git('rev-parse', 'HEAD^') == baseline['program_commit'] and not git('status', '--porcelain')
    names = ['apps/pod/drawblocks/'+n for n in ('effects.go','native_surface_setup.go','native_surface_setup_test.go')]
    assert sorted(git('diff','--name-only','HEAD^','HEAD').splitlines()) == sorted(names)
    source = (module/names[1]).read_text()
    assert all(token not in source for token in ('.Step(', 'b.run(', 'effect{', 'DecodeOpcode', '.Write8(', '.Write16('))
    for name in ('build.txt','audit-candidate-build.txt'):
        text = (raw/name).read_text()
        assert 'go1.24.13' in text and 'vcs.modified=false' in text and 'vcs.revision='+commit in text
    logs = [module/'workplace'/f'surface-setup-{name}.log' for name in ('regression','related','frontend','vet','product')]
    for path in logs:
        assert 'FAIL' not in path.read_text()
        if path != logs[3]:
            assert 'ok  ' in path.read_text()
    assert not logs[3].read_text()
    for package in ('internal/machine','session','apps/pod','apps/pod/bgm','xlate'):
        assert 'github.com/wicanr2/dosgolem/'+package+'\t' in logs[1].read_text()
    text = logs[4].read_text()
    for marker in ('surface-root ordered original and whole SaveState+PortLog=72 unused flag high bytes preserved',
                   'surface-child ordered original and whole SaveState+PortLog=48 free slots through255 parents through8',
                   'surface setup unchanged refusals=28'):
        assert marker in text
    assert 'SKIP' not in text
    spec_path = project/'docs/spec/025-direct-render-commands.md'
    spec = spec_path.read_text()
    for anchor in ('surface-setup-initial','surface-setup-ready'):
        assert spec.count(f'<a id="{anchor}"></a>') == 1
    initial, ready = [spec.split(f'<a id="{a}"></a>')[1].split('<a id=')[0] for a in ('surface-setup-initial','surface-setup-ready')]
    correction = 'SURFACE-SETUP-READY-20261029'
    assert correction in initial and correction in ready and '(#surface-setup-ready)' in initial and '(#surface-setup-initial)' in ready
    assert ':0x2786' in ready and ':0x28D5' in ready
    paths = [raw/'trace-native', raw/'trace-audit', project/'workplace/probe/draw-intercept-v1/native-color-fill-v1/trace-native-v3',
             project/'workplace/probe/draw-intercept-v1/native-fade-v1/trace-original-entries']
    current, audit, previous, oracle = [read(p/'calls.json') for p in paths]
    for field in ('Calls','Digest','VideoCount','AudioCount','FrameCount','DOSCount','VideoSHA256','AudioSHA256','FramesSHA256','DOSSHA256'):
        assert current[field] == audit[field] == previous[field] == oracle[field], field
    assert tree(paths[0]/'save') == tree(paths[1]/'save') == tree(paths[2]/'save') == tree(paths[3]/'save')
    for field in ('Captures','DrawCaptures'):
        assert current[field] == audit[field] == previous[field]
        for row in current[field]:
            assert sha(paths[0]/row['Memory']) == sha(paths[1]/row['Memory']) == sha(paths[2]/row['Memory']) == row['MemorySHA256']
    original_fixed = read(paths[1]/'fixed-records.json')
    remaining = read(paths[0]/'fixed-records.json')
    selected = [r for r in original_fixed if r['Before']['IP'] in (0x2786,0x28d5)]
    assert len(original_fixed) == 20 and sum(r['Plan']['Instructions'] for r in original_fixed) == 1018
    assert len(selected) == 2 and sum(r['Plan']['Instructions'] for r in selected) == 893
    assert remaining == [r for r in original_fixed if r not in selected]
    assert len(remaining) == 18 and sum(r['Plan']['Instructions'] for r in remaining) == 125
    for folder, rows in ((paths[1],original_fixed),(paths[0],remaining)):
        for row in rows:
            assert sha(folder/row['Before']['Memory']) == row['Before']['MemorySHA256']
    ida = project/'workplace/probe/draw-intercept-v1/vga-ida.json'
    assert sha(ida) == '810e653c23f5d9e04b1069e40408d34056b4e2d126d7958c1a5e4b5e7731fb44'
    driver = project/'workplace/input/dos/pod/VGA.DRV'
    assert sha(driver) == 'eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f'
    code = driver.read_bytes()
    metadata = read(ida)
    code_hashes = {}
    for ip, size in ((0x2786,210),(0x28d5,124),(0x2772,20)):
        function = next(f for f in metadata['functions'] if f['file_offset'] == ip)
        exported = bytes.fromhex(''.join(i['bytes'] for i in function['instructions']))
        runtime = (paths[1]/selected[0]['Before']['Memory']).read_bytes()[0x48a10+ip:0x48a10+ip+size]
        assert code[ip:ip+size] == exported == runtime
        code_hashes[hex(ip)] = hashlib.sha256(exported).hexdigest()
    builder_path = project/'tools/build_fixed_control_probe.py'
    module_spec = importlib.util.spec_from_file_location('probe_builder',builder_path)
    builder = importlib.util.module_from_spec(module_spec)
    module_spec.loader.exec_module(builder)
    assert builder.render(project) == (module/'workplace/native-fixed-control-audit/main.go').read_text()
    exits = current['AlgorithmicPreparedByExit']
    added = {'48a1:2857':[1,837],'48a1:2948':[1,56]}
    for key, value in added.items():
        assert exits[key] == value and key not in previous['AlgorithmicPreparedByExit']
    for key in set(exits) | set(previous['AlgorithmicPreparedByExit']):
        if key not in added:
            assert exits.get(key) == previous['AlgorithmicPreparedByExit'].get(key), key
    assert sum(r[0] for r in exits.values()) == current['AlgorithmicPlans'] == previous['AlgorithmicPlans']+2
    assert sum(r[1] for r in exits.values()) == current['AlgorithmicInstructions'] == previous['AlgorithmicInstructions']+893
    assert current['HostInstructions']+current['BlockInstructions']-current['AlgorithmicInstructions'] == 125
    assert current['HostCalls']+current['HostBlocks']-current['AlgorithmicPlans'] == 18
    assert current['Commands'] == previous['Commands']
    result = dict(scope='根畫布與合法VGA子表面整段Go初始化；短預算與失敗配置保留原版，全部Go仍DRAFT',
        program_commit=commit, baseline_commit=baseline['program_commit'], binary_sha256=sha(module/'workplace/bin/pod-play-nativesurfacesetup'),
        source_sha256={n:sha(module/n) for n in names}, source_code_sha256=code_hashes,
        original_vga_sha256=sha(driver), ida_export_sha256=sha(ida), original_address_space='IDA9.4 binary file base 0; runtime CS=48A1',
        controlled_original=dict(root_ordered_and_whole=72, child_ordered_and_whole=48, unchanged_refusals=28),
        normal=dict(steps=current['Digest']['Steps'], state_events_equal=True, guest_files=len(tree(paths[0]/'save')),
                    direct_witnesses=len(current['Captures'])+len(current['DrawCaptures']), new_go_plans=2, reclassified_go_steps=893,
                    algorithmic_plans=current['AlgorithmicPlans'], algorithmic_instructions=current['AlgorithmicInstructions'],
                    previous_fixed_instructions=1018, remaining_fixed_plans=18, remaining_fixed_instructions=125,
                    host_video_bytes=current['Commands']['HostBytes'], new_pixel_bytes=0),
        remaining_fixed_by_start={}, logs_sha256={str(p.relative_to(module)):sha(p) for p in logs},
        trace_sha256={str(p.relative_to(project)):sha(p/'calls.json') for p in paths},
        fixed_records_sha256={str(p.relative_to(project)):sha(p/'fixed-records.json') for p in paths[:2]},
        builder_sha256=sha(builder_path), audit_source_sha256=sha(module/'workplace/native-fixed-control-audit/main.go'),
        probe_sha256=sha(module/'workplace/bin/draw-call-probe-surfacesetup-audit'),
        build_sha256={n:sha(raw/n) for n in ('build.txt','audit-candidate-build.txt')},
        prior_receipt_sha256=sha(baseline_path), spec_sha256=sha(spec_path), correction=correction)
    for row in remaining:
        before, plan = row['Before'], row['Plan']
        key = f"{before['Seg'][1]:04x}:{before['IP']:04x}-{plan['LastIP']:04x}"
        values = result['remaining_fixed_by_start'].setdefault(key,[0,0])
        values[0] += 1
        values[1] += plan['Instructions']
    if stage == 'full':
        rows = []
        for name, old, flash in (('nativesurfacesetup-play-v1/hd','nativecolorfill-play-v1/hd',0),
                                 ('nativesurfacesetup-font-v1/wall-direct','nativecolorfill-font-v1/wall-direct',121)):
            root = project/'workplace/probe/g9-hd'
            now, before = root/name, root/old
            images = [p for p in now.rglob('*.png') if p.is_file()]
            assert images and sorted(str(p.relative_to(now)) for p in images) == sorted(str(p.relative_to(before)) for p in before.rglob('*.png'))
            for folder in ('shots','window-shots','sources','save'):
                assert tree(now/folder) == tree(before/folder), folder
            for file in ('frame.png','source.idx','source.pal','source.glyphs.json','original.rgba','script.txt'):
                assert sha(now/file) == sha(before/file), file
            for pattern in (r'frames=\d+ steps=\d+ memory_sha256=\w+ indexed_sha256=\w+',r'cpu_sha256=\w+ palette_sha256=\w+',r'閃現.*'):
                assert re.findall(pattern,(now/'run.log').read_text()) == re.findall(pattern,(before/'run.log').read_text()), pattern
            assert f'閃現 {flash} 次' in (now/'run.log').read_text()
            inputs, old_inputs = read(now/'inputs.json'), read(before/'inputs.json')
            assert inputs['binary_sha256'] == result['binary_sha256']
            for key in ('orig','font','text','theme'):
                assert inputs[key] == old_inputs[key], key
            report = read(now/'draw-report.json')
            assert report['HostBytes'] == report['NativeBytes'] and report['FallbackBytes'] == report['Mismatches'] == 0
            rows.append(dict(route=name, pngs=len(images), image_sha256={str(p.relative_to(now)):sha(p) for p in images},
                             state_and_guest_files_equal=True, raw_flash=flash, unexpected_flash=0,
                             log_sha256=sha(now/'run.log'), inputs_sha256=sha(now/'inputs.json'), report_sha256=sha(now/'draw-report.json')))
        result['frontend'] = rows
        activation = raw/'activation.json'
        if activation.exists():
            a = read(activation)
            assert a['candidate_sha256'] == result['binary_sha256'] == sha(module/'workplace/bin/pod-play')
            assert a['previous_sha256'] == baseline['binary_sha256'] == sha(module/'workplace/bin'/a['previous_binary'])
            result['activation'] = a
    result['tool_sha256'] = sha(Path(__file__))
    return result


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project',type=Path,required=True)
    ap.add_argument('--stage',choices=('trace','full'),default='full')
    ap.add_argument('--out',type=Path,required=True)
    a = ap.parse_args()
    assert not a.out.exists(), '既有收據不得覆寫'
    result = verify(a.project,a.stage)
    a.out.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(result['normal'],ensure_ascii=False))


if __name__ == '__main__':
    main()
