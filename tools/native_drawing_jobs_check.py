#!/usr/bin/env python3
"""在Docker驗證Go圖像、RGB及填色工作短續行，保留原始收據。"""
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
    return {str(p.relative_to(path)):sha(p) for p in sorted(path.rglob('*')) if p.is_file()}


def job_group(key):
    cs, ip = [int(x,16) for x in key.split(':')]
    if cs != 0x48a1:
        return None
    for group, lo, hi in (('header',0x689e,0x68f2),('blend',0x4b6c,0x4b8b),
                           ('color',0x4b9a,0x4bc4),('fill',0x4268,0x429f)):
        if lo <= ip <= hi:
            return group
    return None


def verify(project, stage):
    module = project/'workplace/dosgolem'
    raw = project/'workplace/probe/draw-intercept-v1/native-drawing-jobs-v1'
    prior_path = project/'docs/re/019-native-surface-setup-runtime.json'
    prior = read(prior_path)
    git = lambda *args:subprocess.check_output(['git',*args],cwd=module,text=True).strip()
    commit = git('rev-parse','HEAD')
    assert git('rev-parse','HEAD^') == prior['program_commit'] and not git('status','--porcelain')
    names = ['apps/pod/drawblocks/'+n for n in ('effects.go','native_job_boundaries.go','native_image_header_boundary.go',
             'native_rgb_job_boundaries.go','native_fill_row_boundary.go','drawing_job_boundaries_test.go','drawing_job_guards_test.go')]
    assert sorted(git('diff','--name-only','HEAD^','HEAD').splitlines()) == sorted(names)
    for name in names[1:5]:
        text = (module/name).read_text()
        assert all(token not in text for token in ('.Step(', 'b.run(', 'effect{', 'DecodeOpcode', '.Write8(', '.Write16('))
    for name in ('build.txt','probe-build.txt'):
        text = (raw/name).read_text()
        assert 'go1.24.13' in text and 'vcs.modified=false' in text and 'vcs.revision='+commit in text
    logs = [module/'workplace'/f'drawing-jobs-{name}.log' for name in ('regression','related','frontend','vet','product')]
    for path in logs:
        assert 'FAIL' not in path.read_text()
        if path != logs[3]:
            assert 'ok  ' in path.read_text()
    assert not logs[3].read_text()
    for package in ('internal/machine','session','apps/pod','apps/pod/bgm','xlate'):
        assert 'github.com/wicanr2/dosgolem/'+package+'\t' in logs[1].read_text()
    product = logs[4].read_text()
    markers = ('drawing boundary normal original samples comparisons=160',
               'drawing header paths=16 ordered and SaveState checkpoints=984',
               'drawing RGB blend paths=14 ordered and SaveState checkpoints=348',
               'drawing RGB call paths=16 ordered and SaveState checkpoints=408',
               'drawing fill row paths=12 ordered and SaveState checkpoints=612',
               'drawing boundary unchanged refusals=128',
               'drawing boundary semantic/frame/source refusals=36 unchanged',
               'drawing boundary unused payload positive controls=3',
               'drawing boundary real resumed positive and rejection pairs=8',
               'drawing boundary production selector cases=32')
    assert all(m in product for m in markers) and 'SKIP' not in product
    spec_path = project/'docs/spec/025-direct-render-commands.md'
    spec = spec_path.read_text()
    older = ('surface-fill-whole-initial','image-header-whole-initial','fade-blend-whole-initial',
             'fade-color-whole-initial','drawing-job-short-initial')
    anchor = 'drawing-job-boundary-ready'
    correction = 'DRAWING-JOB-BOUNDARY-READY-20261029'
    for name in (*older,anchor):
        assert spec.count(f'<a id="{name}"></a>') == 1
    section = lambda a:spec.split(f'<a id="{a}"></a>')[1].split('<a id=')[0]
    ready = section(anchor)
    assert correction in ready
    for name in older:
        assert correction in section(name) and '(#'+anchor+')' in section(name) and '(#'+name+')' in ready
    for address in ('689E','4B6C','4B9A','41AF','4268'):
        assert ':0x'+address in ready
    paths = [raw/'trace-native',project/'workplace/probe/draw-intercept-v1/native-fixed-control-v1/trace-native',
             project/'workplace/probe/draw-intercept-v1/native-fade-v1/trace-original-entries']
    current, previous, oracle = [read(p/'calls.json') for p in paths]
    for field in ('Calls','Digest','VideoCount','AudioCount','FrameCount','DOSCount','VideoSHA256','AudioSHA256','FramesSHA256','DOSSHA256'):
        assert current[field] == previous[field] == oracle[field], field
    assert tree(paths[0]/'save') == tree(paths[1]/'save') == tree(paths[2]/'save')
    for field in ('Captures','DrawCaptures'):
        assert current[field] == previous[field]
        for row in current[field]:
            assert sha(paths[0]/row['Memory']) == sha(paths[1]/row['Memory']) == row['MemorySHA256']
    original_fixed = read(paths[1]/'fixed-records.json')
    remaining = read(paths[0]/'fixed-records.json')
    assert len(original_fixed) == 18 and sum(r['Plan']['Instructions'] for r in original_fixed) == 125
    assert remaining == [r for r in original_fixed if r['Before']['IP'] in (0x2a66,0x2a77)]
    assert len(remaining) == 2 and sum(r['Plan']['Instructions'] for r in remaining) == 14
    for path, rows in ((paths[0],remaining),(paths[1],original_fixed)):
        for row in rows:
            assert sha(path/row['Before']['Memory']) == row['Before']['MemorySHA256']
    assert current['Commands'] == previous['Commands']
    exits, old_exits = current['AlgorithmicPreparedByExit'], previous['AlgorithmicPreparedByExit']
    assert sum(r[0] for r in exits.values()) == current['AlgorithmicPlans']
    assert sum(r[1] for r in exits.values()) == current['AlgorithmicInstructions']
    delta = {}
    groups = {k:[0,0] for k in ('header','blend','color','fill')}
    for key in set(exits)|set(old_exits):
        now, old = exits.get(key,[0,0]), old_exits.get(key,[0,0])
        if now != old:
            group = job_group(key)
            assert group is not None, key
            d = [now[0]-old[0],now[1]-old[1]]
            delta[key] = d
            groups[group] = [groups[group][i]+d[i] for i in range(2)]
    assert all(v[0]>0 and v[1]>0 for v in groups.values())
    fixed = current['HostInstructions']+current['BlockInstructions']-current['AlgorithmicInstructions']
    assert fixed == 14
    extra_hosted = current['HostInstructions']+current['BlockInstructions']-previous['HostInstructions']-previous['BlockInstructions']
    extra_go = current['AlgorithmicInstructions']-previous['AlgorithmicInstructions']
    assert extra_go == 111+extra_hosted and extra_hosted > 0
    assert current['HostCalls']+current['HostBlocks']-current['AlgorithmicPlans'] == 2
    ida = project/'workplace/probe/draw-intercept-v1/vga-ida.json'
    assert sha(ida) == prior['ida_export_sha256']
    metadata = read(ida)
    driver = project/'workplace/input/dos/pod/VGA.DRV'
    assert sha(driver) == prior['original_vga_sha256']
    original_bytes = driver.read_bytes()
    witness = next(r for r in original_fixed if r['Before']['IP'] == 0x689e)['Before']
    memory = (paths[1]/witness['Memory']).read_bytes()
    code_hashes = {}
    for lo, size in ((0x689e,86),(0x4b6c,33),(0x4b9a,44),(0x41af,260)):
        ins = [i for f in metadata['functions'] for i in f['instructions'] if lo<=i['file_offset']<lo+size]
        data = bytes.fromhex(''.join(i['bytes'] for i in ins))
        assert data == original_bytes[lo:lo+size] == memory[0x48a10+lo:0x48a10+lo+size]
        code_hashes[hex(lo)] = hashlib.sha256(data).hexdigest()
    result = dict(scope='四類typed繪圖資料工作最低一步續行；BIOS及未驗條件保留，全部Go仍DRAFT',
        program_commit=commit, baseline_commit=prior['program_commit'], binary_sha256=sha(module/'workplace/bin/pod-play-nativedrawingjobs'),
        source_sha256={n:sha(module/n) for n in names}, original_vga_sha256=sha(driver), ida_export_sha256=sha(ida),
        source_code_sha256=code_hashes, source_address_space='IDA9.4 binary file base 0; runtime CS=48A1',
        controlled_original=dict(boundary_comparisons=2523, main_paths=58, normal_samples=16, refusals=172,
                                 unused_payload_positives=3, resumed_positive_rejection_pairs=8, selector_cases=32),
        normal=dict(steps=current['Digest']['Steps'], original_state_events_equal=True, guest_files=len(tree(paths[0]/'save')),
                    direct_witnesses=len(current['Captures'])+len(current['DrawCaptures']),
                    algorithmic_plans=current['AlgorithmicPlans'], algorithmic_instructions=current['AlgorithmicInstructions'],
                    previous_fixed_instructions=125, fixed_instructions=14, reclassified_fixed_steps=111,
                    cpu_steps_moved_to_go=extra_hosted, net_new_go_steps=extra_go,
                    host_video_bytes=current['Commands']['HostBytes'], new_pixel_bytes=0),
        algorithmic_exit_delta=dict(sorted(delta.items())), job_delta=groups,
        remaining_bios_plans=[dict(entry=f"{r['Before']['Seg'][1]:04x}:{r['Before']['IP']:04x}",steps=r['Plan']['Instructions']) for r in remaining],
        logs_sha256={str(p.relative_to(module)):sha(p) for p in logs},
        trace_sha256={str(p.relative_to(project)):sha(p/'calls.json') for p in paths},
        fixed_records_sha256={str(p.relative_to(project)):sha(p/'fixed-records.json') for p in paths[:2]},
        build_sha256={n:sha(raw/n) for n in ('build.txt','probe-build.txt')},
        probe_sha256=sha(module/'workplace/bin/draw-call-probe-drawingjobs-audit'),
        prior_receipt_sha256=sha(prior_path), spec_sha256=sha(spec_path), correction=correction)
    if stage == 'full':
        rows = []
        for name, old, flash in (('nativedrawingjobs-play-v1/hd','nativesurfacesetup-play-v1/hd',0),
                                 ('nativedrawingjobs-font-v1/wall-direct','nativesurfacesetup-font-v1/wall-direct',121)):
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
            rows.append(dict(route=name,pngs=len(images),image_sha256={str(p.relative_to(now)):sha(p) for p in images},
                             state_and_guest_files_equal=True,raw_flash=flash,unexpected_flash=0,
                             inputs_sha256=sha(now/'inputs.json'),log_sha256=sha(now/'run.log'),report_sha256=sha(now/'draw-report.json')))
        result['frontend'] = rows
        activation = raw/'activation.json'
        if activation.exists():
            a = read(activation)
            assert a['candidate_sha256'] == result['binary_sha256'] == sha(module/'workplace/bin/pod-play')
            assert a['previous_sha256'] == prior['binary_sha256'] == sha(module/'workplace/bin'/a['previous_binary'])
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
