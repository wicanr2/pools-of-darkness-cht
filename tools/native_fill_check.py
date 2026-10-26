"""驗#49 Go矩形、圖片與色號掃描的原版串流及正常路線。"""
import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import re


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    return json.loads(path.read_text())


def tree(path):
    return {str(p.relative_to(path)): sha(p) for p in sorted(path.rglob('*')) if p.is_file()}


def load(path):
    spec = importlib.util.spec_from_file_location('native_fill_ui_check', path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--stage', choices=('trace', 'full'), default='full')
    ap.add_argument('--commit', required=True)
    ap.add_argument('--family', choices=('fill', 'rle', 'palette', 'image', 'fade', 'color', 'setup', 'row', 'output', 'clip', 'control', 'primitive'), default='fill')
    a = ap.parse_args()
    assert not a.out.exists(), '保留既有收據，請另定輸出檔名'
    p = a.project
    rle = a.family != 'fill'
    primitive = a.family == 'primitive'
    control = a.family in ('control', 'primitive')
    clip = a.family in ('clip', 'control', 'primitive')
    output = a.family in ('output', 'clip', 'control', 'primitive')
    row = a.family in ('row', 'output', 'clip', 'control', 'primitive')
    setup = a.family in ('setup', 'row', 'output', 'clip', 'control', 'primitive')
    color = a.family in ('color', 'setup', 'row', 'output', 'clip', 'control', 'primitive')
    fade = a.family in ('fade', 'color', 'setup', 'row', 'output', 'clip', 'control', 'primitive')
    image = a.family in ('image', 'fade', 'color', 'setup', 'row', 'output', 'clip', 'control', 'primitive')
    palette = a.family in ('palette', 'image', 'fade', 'color', 'setup', 'row', 'output', 'clip', 'control', 'primitive')
    stem = 'native'+a.family
    root = p/f'workplace/probe/draw-intercept-v1/native-{a.family}-v1'
    original_root = p/'workplace/probe/draw-intercept-v1/native-fade-v1' if color else root
    canonical = p/'workplace/probe/draw-intercept-v1/all-functions-v1/trace-v2-legacy'
    previous = original_root/'trace-original-entries' if fade else canonical
    current = root/('trace-native' if rle else 'trace-native-v2')
    old, new = read(previous/'calls.json'), read(current/'calls.json')
    if fade:
        baseline = read(canonical/'calls.json')
        for key in ('Calls', 'Digest', 'VideoCount', 'AudioCount', 'FrameCount', 'DOSCount',
                    'VideoSHA256', 'AudioSHA256', 'FramesSHA256', 'DOSSHA256'):
            assert old[key] == baseline[key], ('original entry extension', key)
        assert old['HostCalls'] == old['HostBlocks'] == old['AlgorithmicPlans'] == 0
        assert tree(previous/'save') == tree(canonical/'save')
        for field in ('Captures', 'DrawCaptures'):
            extended = {row['Memory']: row for row in old[field]}
            for row in baseline[field]:
                assert extended[row['Memory']] == row
                assert sha(previous/row['Memory']) == sha(canonical/row['Memory'])
    for key in ('Calls', 'Digest', 'VideoCount', 'AudioCount', 'FrameCount', 'DOSCount',
                'VideoSHA256', 'AudioSHA256', 'FramesSHA256', 'DOSSHA256'):
        assert old[key] == new[key], key
    assert new['Digest']['Steps'] == 712250000 and new['VideoCount'] == 3343240
    assert tree(previous/'save') == tree(current/'save')
    witnesses = []
    missing_entries = {(0x48a1, ip) for ip in ((0x2772,) if control else (0x2772, 0x2aeb, 0x2a90))}
    for field in ('Captures', 'DrawCaptures'):
        if field == 'Captures':
            assert len(new[field]) == len(old[field]), ('entry witness count', field)
        else:
            old_entries = {(row['Seg'][1], row['IP']) for row in old[field]}
            new_entries = {(row['Seg'][1], row['IP']) for row in new[field]}
            assert new_entries <= old_entries and old_entries-new_entries == missing_entries, 'unexpected missing fixed-plan entry'
        sources = {(c['Seg'][1], c['IP']): c for c in old[field]}
        if field == 'Captures':
            sources = {(c['Seg'][1], c['IP'], c['Memory']): c for c in old[field]}
        for row in new[field]:
            key = (row['Seg'][1], row['IP'], row['Memory']) if field == 'Captures' else (row['Seg'][1], row['IP'])
            assert row == sources[key]
            assert sha(current/row['Memory']) == sha(previous/row['Memory']) == row['MemorySHA256']
            witnesses.append(dict(cs=row['Seg'][1], ip=row['IP'], memory_sha256=row['MemorySHA256']))
    counts = new['AlgorithmicPreparedByExit']
    assert sum(row[0] for row in counts.values()) == new['AlgorithmicPlans']
    assert sum(row[1] for row in counts.values()) == new['AlgorithmicInstructions']
    assert counts['48a1:42ac'][0] > 0 and counts['48a1:42b2'][0] > 0
    if rle:
        assert any(counts.get(key, [0])[0] > 0 for key in ('48a1:6907', '48a1:6909'))
        assert any(counts.get(key, [0])[0] > 0 for key in ('48a1:6918', '48a1:691a'))
    if palette:
        assert counts['48a1:4a70'][0] > 0
    if image:
        assert counts['48a1:68f2'][0] > 0
    if fade:
        assert counts['48a1:4b38'][0] > 0 and counts['48a1:4b8b'][0] > 0
    if color:
        assert counts['48a1:4bbe'][0] > 0 and counts['48a1:4b9d'][0] > 0
    if setup:
        assert all(counts[key][0] > 0 for key in ('48a1:4a8f', '48a1:4b09', '48a1:4b53'))
    fixed = new['HostInstructions'] + new['BlockInstructions'] - new['AlgorithmicInstructions']
    if row:
        assert counts['48a1:6899'][0] > 0
    if output:
        assert counts['48a1:2ac2'][0] > 0
    assert fixed > 0, '其餘固定區塊仍存在，不能稱全部Go演算法完成'
    assert new['Commands']['Mismatches'] == 0
    code = p/'workplace/input/dos/pod/VGA.DRV'
    assert sha(code) == 'eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f'
    assert hashlib.sha256(code.read_bytes()[0x41af:0x42b3]).hexdigest() == 'c7db9b7b119f14464a7e76d35125922a3559d10fef36624f79f5120dd0d02f00'
    test_log = p/f'workplace/dosgolem/workplace/native-{a.family}-product-tests.log'
    if output:
        assert hashlib.sha256(code.read_bytes()[0x2aa8:0x2ac3]).hexdigest() == '7134f5f3fbaf2adf007afa01632ff5ec980bd4c127a095defaa954fd831b4801'
        for line in ('RGB job original comparisons 192', 'RGB job original entry comparisons 6'):
            assert line in test_log.read_text(), line
        core_log = p/f'workplace/dosgolem/workplace/native-{a.family}-core-tests.log'
        assert '\nPASS\n' in core_log.read_text()
    elif row:
        assert hashlib.sha256(code.read_bytes()[0x6850:0x689e]).hexdigest() == '1c4ab9672accc6ec8dcd86e8703b26cf877efa2ca54e8afa5268d2c5fed7a62b'
        for line in ('image row synthetic original comparisons 174', 'image row original entry-state comparisons 384'):
            assert line in test_log.read_text(), line
    elif setup:
        for lo, hi, expected in ((0x4a83, 0x4a91, '1814a41894eea9e581384ac7c8a316353039634bf747a66ceb5ec19aff3361af'),
                                 (0x4afd, 0x4b0b, '343202170a7507bd27ba58e074710568546eac5f1fc99691bcac8d8e11a98349'),
                                 (0x4b3a, 0x4b55, '592b32248b4af06920a2979be064570ce726975b43eae84d9591126394638a7d')):
            assert hashlib.sha256(code.read_bytes()[lo:hi]).hexdigest() == expected
        for line in ('palette list synthetic original comparisons 36', 'palette target synthetic original comparisons 72',
                     'palette error synthetic original comparisons 32', 'palette setup original entry-state comparisons 390'):
            assert line in test_log.read_text(), line
    elif color:
        assert hashlib.sha256(code.read_bytes()[0x4b9a:0x4bc6]).hexdigest() == '548776df6fec7df3ebb2f57ade4412c8704fee225a928411593571d6c7c3b525'
        assert 'fade color original job comparisons 32' in test_log.read_text()
    elif fade:
        for lo, hi, expected in ((0x4b19, 0x4b3a, '426b9c7bab4ee58a83495f6ee251ade3d04d01ce17e57c163cfdcd18c4b00f86'),
                                 (0x4b6c, 0x4b8d, '391005d80b86b978189cde8e392fbfbca279480fda3f7f3a175cb9c14ed2f2ae')):
            assert hashlib.sha256(code.read_bytes()[lo:hi]).hexdigest() == expected
        log = test_log.read_text()
        for line in ('fade difference synthetic original comparisons 72', 'fade blend synthetic original comparisons 40',
                     'fade normal original comparisons 384'):
            assert line in log, line
    elif image:
        assert hashlib.sha256(code.read_bytes()[0x689e:0x68f4]).hexdigest() == '56ecc68cd838cf71e5de2da105cf370ccd3c2f7656159bdb2de14800a9c888a1'
        for base in (0x6c9e, 0x6d42, 0x6df3):
            assert hashlib.sha256(code.read_bytes()[base:base+50]).hexdigest() == '71ffb9fb40b99437973306c9ef582b52f24b0014fb32e1feed59703a7ce12575'
        log = test_log.read_text()
        for line in ('image span synthetic original comparisons 18', 'image span normal original comparisons 64',
                     'mirror pair independent original comparisons 804', 'image mirror state comparisons 144 states 24'):
            assert line in log, line
        whole = re.search(r'controlled whole mirror original cases 24 committed mirror plans (\d+)', log)
        assert whole and int(whole[1]) > 0
    elif palette:
        assert hashlib.sha256(code.read_bytes()[0x4a67:0x4a72]).hexdigest() == '3c8934166cf6cd8d1ecdf79eac51b0cb33bb1bb3e545b883578dcb85971e1b5a'
        assert 'palette synthetic original plan comparisons 60' in test_log.read_text()
        assert 'palette normal original plan comparisons 192' in test_log.read_text()
    elif rle:
        assert hashlib.sha256(code.read_bytes()[0x68f4:0x691c]).hexdigest() == '1926d07bb1c1e91594b6a388138e57b204d8efb327addc7cebf32d0dd7ec5818'
        assert '22 plan comparisons 110' in test_log.read_text()
        assert 'normal original image RLE state comparisons 192' in test_log.read_text()
    else:
        assert '18 independent plan comparisons 194' in test_log.read_text()
    assert '\nPASS\n' in test_log.read_text()
    if clip:
        for lo, hi, expected in (
            (0x6733, 0x6832, 'd9ebc163f106be514229472090f2497771df8e36ab5aeca61681f758d32baa54'),
            (0x6832, 0x6850, 'a8b2dbb0dbf04e34cca2aca89962a68329389b7bdfa17bf485b5feba3fec1383'),
            (0x691c, 0x6a0c, 'e6e2c71563b3b8b991f450d0c1184b02acfd37ec8245bff32009f7d03ac1f23d'),
            (0x6a0c, 0x6e63, 'e95a2f6e11b374a9faa3b42a3936665692a05d289ce55e5718c2e58179c43048'),
        ):
            assert hashlib.sha256(code.read_bytes()[lo:hi]).hexdigest() == expected
        for marker in ('mirror full packet original comparisons 160', 'mirror clipped packet original comparisons 1440',
                       'controlled flip state comparisons 266 states 133', 'flip geometry original comparisons 388',
                       'source/surface initialization original comparisons 20',
                       'surface chain / frame address original comparisons 96'):
            assert marker in test_log.read_text(), marker
        assert re.search(r'controlled whole mirror original cases 48 committed mirror plans [1-9]\d*', test_log.read_text())
    tests = (dict(rgb_synthetic_plan_comparisons=192, isolated_original_entry_plan_comparisons=6,
                  refusal_cases=14, device_comparison='CPU/full RAM/VGA/DAC/Ports/ordered PortLog values, steps and model cycles',
                  initial_dac_write_phases=[0, 1, 2], generic_core_refusal_cases=15,
                  generic_core_test_log_sha256=sha(core_log)) if output else
             dict(row_synthetic_plan_comparisons=174, original_entry_states=64,
                  original_entry_state_plan_comparisons=384, refusal_cases=22) if row else
             dict(list_synthetic_plan_comparisons=36, target_synthetic_plan_comparisons=72,
                  error_synthetic_plan_comparisons=32, isolated_entry_states=65,
                  isolated_entry_state_plan_comparisons=390, refusal_cases=21) if setup else
             dict(original_color_job_comparisons=32, refusal_cases=13) if color else
             dict(difference_synthetic_plan_comparisons=72, blend_synthetic_plan_comparisons=40,
                  normal_entry_states=64, normal_state_plan_comparisons=384, refusal_cases=23) if fade else
             dict(span_synthetic_plan_comparisons=18, span_normal_states=32, span_normal_plan_comparisons=64,
                  span_refusal_cases=16, mirror_synthetic_plan_comparisons=804, mirror_controlled_image_states=24,
                  mirror_controlled_state_plan_comparisons=144, mirror_refusal_cases=9,
                  controlled_whole_mirror_cases=24, controlled_whole_mirror_committed_plans=int(whole[1])) if image else
             dict(synthetic_plan_comparisons=60, normal_scan_states=32,
                  normal_state_plan_comparisons=192, refusal_cases=9) if palette else
             dict(packet_cases=22, packet_plan_comparisons=110, normal_image_states=32,
                  normal_state_plan_comparisons=192, refusal_cases=9) if rle else
             dict(geometries=9, cost_models=2, cases=18, plan_comparisons=194, refusal_cases=8))
    tests['log_sha256'] = sha(test_log)
    if clip:
        tests['image_control'] = dict(forward_original_comparisons=942, forward_refusals=26,
            mirror_packet_original_comparisons=160, clipped_mirror_original_comparisons=1440,
            flip_geometry_original_comparisons=388, controlled_flip_states=133, controlled_state_original_comparisons=266,
            cpu_model_original_comparisons=12, mirror_refusals=17, exhausted_clip_original_comparisons=8,
            initialization_original_comparisons=116, initialization_refusals=16, controlled_whole_image_cases=48)
    result = dict(scope=('Go完整RGB工作交VGA服務，以有界調色盤命令保留埠時序，保留既有Go圖像演算法；其他裁切／鏡像控制及固定區塊仍保留，#49未全部完成' if output else
                        'Go圖片列標頭、上下界判斷與列位置，保留既有Go圖像演算法；其他裁切／鏡像控制及固定區塊仍保留，#49未全部完成' if row else
                        'Go使用色清單、目標RGB與誤差陣列準備，保留既有Go圖像演算法；其餘控制、硬體I/O及固定區塊仍保留，#49未全部完成' if setup else
                        'Go逐色RGB工作與完成交接，保留既有Go圖像演算法；其餘控制、硬體I/O及固定區塊仍保留，#49未全部完成' if color else
                        'Go淡出色差與插值，保留既有Go圖像演算法；其餘控制、硬體I/O及固定區塊仍保留，#49未全部完成' if fade else
                        'Go未裁切圖片區塊解析、既有RLE與鏡像像素對；鏡像以受控原圖驗證，其他圖片控制、裁切與固定區塊仍保留，#49未全部完成' if image else
                        'Go矩形、未裁切RLE像素及使用色號掃描；圖片控制、裁切及鏡像等固定區塊仍保留，#49未全部完成' if palette else
                        'Go矩形與未裁切RLE像素循環；圖片控制、裁切及鏡像等固定區塊仍保留，#49未全部完成' if rle else
                         '首批Go矩形填色與相容續做；其他VGA固定區塊仍保留，#49未全部完成'),
                  program_commit=a.commit,
                  binary_sha256=sha(p/f'workplace/dosgolem/workplace/bin/pod-play-{stem}'),
                  trace_binary_sha256=sha(p/(f'workplace/dosgolem/workplace/bin/draw-call-probe-{stem}'+('-v2' if not rle else ''))),
                  prototype_and_original_tests=tests,
                  virtual_steps=new['Digest']['Steps'], original_state_equal=True,
                  ordered_video_writes=new['VideoCount'], ordered_video_sha256=new['VideoSHA256'],
                  retraces=new['FrameCount'], retraces_sha256=new['FramesSHA256'],
                  audio_events=new['AudioCount'], audio_sha256=new['AudioSHA256'],
                  dos_events=new['DOSCount'], dos_sha256=new['DOSSHA256'],
                  guest_files=len(tree(current/'save')), guest_files_equal=True, witnesses=witnesses,
                  algorithmic_plans=new['AlgorithmicPlans'], algorithmic_instructions=new['AlgorithmicInstructions'],
                  fixed_semantic_instructions=fixed, algorithmic_by_exit=counts,
                  prepared_equals_committed=True, classification_scope='HostWork only; ordinary game CPU instructions are not drawing fallback',
                  digest=new['Digest'], source_reports={str(d.relative_to(p)): sha(d/'calls.json') for d in (previous, current)},
                  benchmark=read(root/'function-cost-v1.json'),
                  benchmark_scope='seven interleaved groups, 32 complete calls each, two-CPU Docker; memory-only entry fixtures with freshly initialized VGA; 4944 is not full-state normal fade timing; includes source guards, compatibility writes and notifications; not FPS')
    if clip:
        result['scope'] = 'Go圖片來源／表面初始化、列控制、正反向RLE、水平／垂直翻轉與裁切及ABI返回；其他VGA固定區塊仍保留，#49未全部完成'
    result['direct_witness_count'] = len(witnesses)
    result['missing_direct_entries'] = [dict(cs=cs, ip=ip) for cs, ip in sorted(missing_entries)]
    result['direct_witness_scope'] = 'BeforeStep omits the entries listed in missing_direct_entries inside fused plans; witnesses above are only the directly observed subset'
    if fade:
        manifest = read(root/'observed-build.json')
        assert manifest['program_commit'] == a.commit
        assert manifest['original_probe_sha256'] == sha(p/'tools/draw_call_probe.go')
        assert manifest['builder_sha256'] == sha(p/'tools/build_observed_draw_probe.py')
        module = p/'workplace/dosgolem'
        name = manifest['binary'].removeprefix('draw-call-probe-')
        assert sha(module/'workplace/bin'/manifest['binary']) == manifest['binary_sha256']
        assert sha(module/f'workplace/{name}-probe/main.go') == manifest['observed_probe_sha256']
        for filename, identities in manifest['source_files'].items():
            assert sha(module/'apps/pod/drawblocks'/filename) == identities['original_sha256']
            assert sha(module/f'workplace/{name}-blocks'/filename) == identities['observed_sha256']
            if filename != 'effects.go':
                assert identities['original_sha256'] == identities['observed_sha256']
        observed_root = root/'trace-observed'
        observed = read(observed_root/'calls.json')
        for key in ('Calls', 'Digest', 'VideoCount', 'AudioCount', 'FrameCount', 'DOSCount',
                    'VideoSHA256', 'AudioSHA256', 'FramesSHA256', 'DOSSHA256'):
            assert observed[key] == old[key], ('observed fixed-plan boundaries', key)
        assert tree(observed_root/'save') == tree(previous/'save')
        assert observed['Commands']['Mismatches'] == 0
        assert sum(row[0] for row in observed['AlgorithmicPreparedByExit'].values()) == observed['AlgorithmicPlans']
        assert sum(row[1] for row in observed['AlgorithmicPreparedByExit'].values()) == observed['AlgorithmicInstructions']
        supplemental = []
        for field in ('Captures', 'DrawCaptures'):
            sources = {row['Memory']: row for row in old[field]}
            rows = {row['Memory']: row for row in observed[field]}
            assert rows.keys() == sources.keys(), ('observed entry count', field)
            for filename, row in rows.items():
                assert row == sources[filename]
                assert sha(observed_root/filename) == sha(previous/filename) == row['MemorySHA256']
                supplemental.append(dict(cs=row['Seg'][1], ip=row['IP'], memory_sha256=row['MemorySHA256']))
        assert len(supplemental) == 47 and len(witnesses) == (46 if control else 44)
        result['instrumented_entry_observation'] = dict(
            scope='research-only boundaries before registered entries; not a production execution-count substitute',
            build_manifest=manifest, calls_sha256=sha(observed_root/'calls.json'),
            original_baseline_sha256=sha(canonical/'calls.json'), all_47_entry_witnesses_equal=True,
            witnesses=supplemental, algorithmic_plans=observed['AlgorithmicPlans'],
            algorithmic_instructions=observed['AlgorithmicInstructions'])
    if a.stage == 'full':
        demo = p/f'workplace/probe/g9-hd/{stem}-play-v1'
        summary = read(demo/'bigpic-summary.json')
        assert summary['binary_sha256'] == result['binary_sha256']
        assert summary['prior_original_state_equal'] and summary['prior_sample_composite_equal']
        result['normal_demo'] = {k: v for k, v in summary.items() if k != 'unobserved_actor_keys'}
        result['normal_demo_draws'] = {}
        for mode in ('hd', 'base', 'english', 'twox'):
            draw = read(demo/mode/'draw-report.json')
            assert draw['AlgorithmicPlans'] > 0 and draw['Mismatches'] == 0
            result['normal_demo_draws'][mode] = {k: draw[k] for k in ('HostCalls', 'HostInstructions', 'HostBlocks', 'BlockInstructions', 'AlgorithmicPlans', 'AlgorithmicInstructions')}
        ui = load(p/'tools/direct_ui_check.py')
        frontend = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt',
                    'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
        rows = []
        for route in ('wall', 'movie', 'title'):
            folder = p/f'workplace/probe/g9-hd/{stem}-ui-v1/{route}-direct'
            baseline = p/f'workplace/probe/g9-hd/all-ui-v2/{route}-direct'
            assert ui.endpoint(folder) == ui.endpoint(baseline)
            ci, bi = read(folder/'inputs.json'), read(baseline/'inputs.json')
            for key in ('orig', 'font', 'text'):
                assert ci[key] == bi[key], (route, key)
            assert ci['binary_sha256'] == result['binary_sha256']
            guests = [{k: v for k, v in tree(d/'save').items() if k not in frontend} for d in (folder, baseline)]
            assert guests[0] == guests[1]
            for name in ('source.idx', 'source.pal', 'original.rgba', 'frame.png'):
                assert (folder/name).read_bytes() == (baseline/name).read_bytes(), (route, name)
            for name in ('shots', 'window-shots', 'sources'):
                assert tree(folder/name) and tree(folder/name) == tree(baseline/name), (route, name)
            audit = ui.audit(folder, route)
            assert audit == ui.audit(baseline, route)
            draw = read(folder/'draw-report.json')
            assert draw['Mismatches'] == 0 and draw['AlgorithmicPlans'] > 0
            rows.append(dict(route=route, endpoint=ui.endpoint(folder), screenshot_pairs=len(tree(folder/'shots')),
                             screenshots_equal=True, guest_files=len(guests[0]), **audit,
                             algorithmic_plans=draw['AlgorithmicPlans'], algorithmic_instructions=draw['AlgorithmicInstructions'],
                             draw_report_sha256=sha(folder/'draw-report.json')))
        result['normal_ui'] = rows
        before = p/'workplace/probe/g9-hd/all-save-v2/native'
        after = p/f'workplace/probe/g9-hd/{stem}-save-v1/native'
        inputs = [read(d/'before.json') for d in (before, after)]
        for value in inputs:
            value.pop('binary_sha256')
        assert inputs[0] == inputs[1] and 'SAVE/savgamB.pty' not in inputs[1]
        saved = [read(d/'after-save.json') for d in (before, after)]
        assert saved[0] == saved[1] and tree(before/'game/SAVE') == tree(after/'game/SAVE')
        original_slot = {k[5:]: v for k, v in inputs[1].items() if k.startswith('SAVE/')}
        assert all(saved[1][key] == value for key, value in original_slot.items())
        phases = []
        for phase, frames in (('save', 4700), ('load', 3250)):
            logs = [(d/f'{phase}.log').read_text() for d in (before, after)]
            endpoints = []
            for log in logs:
                start = re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)', log)
                end = re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)', log)
                assert start and end and int(start[1]) == frames
                assert '字型缺字' not in log and '截圖失敗' not in log
                endpoints.append(start.groups()+end.groups())
            assert endpoints[0] == endpoints[1]
            for suffix in ('final.rgba', 'source.idx', 'source.pal'):
                assert (before/f'{phase}-{suffix}').read_bytes() == (after/f'{phase}-{suffix}').read_bytes()
            assert tree(before/f'{phase}-shots') == tree(after/f'{phase}-shots')
            draw = read(after/f'{phase}-draw.json')
            assert draw['Mismatches'] == 0 and draw['AlgorithmicPlans'] > 0
            phases.append(dict(phase=phase, endpoint=endpoints[1], algorithmic_plans=draw['AlgorithmicPlans'],
                               algorithmic_instructions=draw['AlgorithmicInstructions']))
        result['normal_save_roundtrip'] = dict(slot_b_created=True, slot_b_loaded=True, save_files=len(saved[1]),
                                               all_save_bytes_equal=True, original_slot_a_unchanged=len(original_slot),
                                               slot_b_sha256=saved[1]['savgamB.pty'], screenshot_pairs=6, phases=phases)
        theme = p/'workplace/hd/theme'
        expected = read(demo/'hd/inputs.json')['theme']
        assert tree(theme) == expected
        result['theme_files'] = len(expected)
        result['theme_tree_sha256'] = hashlib.sha256(json.dumps(expected, sort_keys=True).encode()).hexdigest()
        if (root/'activation.json').exists():
            activation = read(root/'activation.json')
            assert activation['candidate_sha256'] == result['binary_sha256']
            assert sha(p/'workplace/dosgolem/workplace/bin/pod-play') == result['binary_sha256']
            previous_name = activation.get('previous_binary', 'pod-play-all')
            assert sha(p/f'workplace/dosgolem/workplace/bin/{previous_name}') == activation['previous_sha256']
            result['activation'] = activation
    result['tools'] = {name: sha(p/name) for name in ('tools/native_fill_check.py', 'tools/draw_call_probe.go',
                       'tools/vga_battle_play_check.py', 'tools/direct-ui-receipt.sh', 'tools/whole-call-save-receipt.sh')}
    if fade:
        result['tools']['tools/build_observed_draw_probe.py'] = sha(p/'tools/build_observed_draw_probe.py')
    result['native_source_sha256'] = sha(p/'workplace/dosgolem/apps/pod/drawblocks/native_fill.go')
    if rle:
        result['rle_source_sha256'] = sha(p/'workplace/dosgolem/apps/pod/drawblocks/native_rle.go')
        result['dispatch_source_sha256'] = sha(p/'workplace/dosgolem/apps/pod/drawblocks/effects.go')
    if palette:
        result['palette_source_sha256'] = sha(p/'workplace/dosgolem/apps/pod/drawblocks/native_palette.go')
    if image:
        result['image_span_source_sha256'] = sha(p/'workplace/dosgolem/apps/pod/drawblocks/native_image_span.go')
        result['mirror_source_sha256'] = sha(p/'workplace/dosgolem/apps/pod/drawblocks/native_mirror.go')
        result['native_go_source_review'] = dict(
            scope='planImageSpan, planMirrorPairs and planRLERuns; source inspection with the above SHA-256 bindings',
            asset_geometry_and_pixel_loops_call_cpu_step=False,
            asset_geometry_and_pixel_loops_dispatch_guest_opcodes=False,
            pixel_generation_uses_instruction_effect_helpers=False,
            legacy_addresses_role='entry/resume boundaries and ordered compatibility metadata',
            remaining_fixed_semantic_paths=True)
        mirror_exits = {key: row for key, row in counts.items() if key.startswith('48a1:') and
                        any(base <= int(key.split(':')[1], 16) < base+50 for base in (0x6c9e, 0x6d42, 0x6df3))}
        result['normal_machine_mirror_by_exit'] = mirror_exits
        result['mirror_normal_player_scope'] = 'Only exits explicitly recorded here are observed; controlled calls are not normal player evidence'
    if clip:
        module = p/'workplace/dosgolem'
        result['image_control_sources'] = {name: sha(module/'apps/pod/drawblocks'/name) for name in
            ('native_image_clip.go', 'native_image_flip.go', 'native_image_init.go',
             'image_clip_cases_test.go', 'image_flip_cases_test.go', 'image_init_cases_test.go', 'native_mirror_test.go')}
        result['cpu_model_query_source_sha256'] = sha(module/'session/observer.go')
        result['native_go_source_review']['image_control_scope'] = 'Typed surface/frame lookup and Go span arithmetic decode, clip and reverse asset packets; CPU model query only normalizes ABI flags. No CPU.Step, opcode dispatcher or effect helper in asset geometry/pixel generation.'
    if fade:
        result['fade_source_sha256'] = sha(p/'workplace/dosgolem/apps/pod/drawblocks/native_fade.go')
        result['fade_entry_table_sha256'] = sha(original_root/'fade-entries.json')
        fixtures = [row for row in old['DrawCaptures'] if row['IP'] in (0x4b19, 0x4b6c)]
        assert len(fixtures) == 2
        assert {row['MemorySHA256'] for row in fixtures} == {
            '5e678ca10eb2f14cdc70220830009de5b2a081da8b80eccb0a1123f1b3a70581',
            'f41e7c00aa01f10eab55e982e7e5bdfedf7e57dd4936e09355751768c2979b00'}
        result['fade_original_entry_fixtures'] = fixtures
        result['native_go_source_review']['fade_scope'] = 'planFadeDifference and planFadeBlend traverse RGB arrays directly; hardware palette reads/writes and waits remain outside these algorithms'
    if color:
        result['fade_color_source_sha256'] = sha(p/'workplace/dosgolem/apps/pod/drawblocks/native_fade_color.go')
        result['native_go_source_review']['color_scope'] = 'planFadeColor reads a typed palette index and RGB triplet, builds compatibility arguments and yields at the hardware output boundary; no instruction-effect dispatcher'
        result['color_benchmark_scope'] = 'The existing first-4944 fresh-VGA fixture does not independently measure the normal color-job sequence; no color-job speed or FPS claim'
    if setup:
        result['palette_setup_source_sha256'] = sha(p/'workplace/dosgolem/apps/pod/drawblocks/native_palette_setup.go')
        result['native_go_source_review']['setup_scope'] = 'planPaletteList filters used-color data, planPaletteTarget fills RGB triplets and planPaletteError initializes the interpolation array; no opcode or instruction-effect dispatcher'
    if row:
        result['image_row_source_sha256'] = sha(p/'workplace/dosgolem/apps/pod/drawblocks/native_image_row.go')
        result['native_go_source_review']['row_scope'] = 'planImageRow parses a row index, compares signed vertical bounds and calculates the destination using Go geometry; ABI registers, timing and ordered RAM notices are separate; no opcode or instruction-effect dispatcher'
        result['normal_machine_row_by_exit'] = {key: value for key, value in counts.items() if key in ('48a1:685c', '48a1:6866', '48a1:686d', '48a1:6899')}
    if output:
        result['palette_output_source_sha256'] = sha(p/'workplace/dosgolem/apps/pod/drawblocks/native_palette_output.go')
        result['native_go_source_review']['output_scope'] = 'planPaletteOutput decodes one typed index/RGB job and emits HostPaletteWrite for the existing VGA DAC service; generic commit merges device and memory moments without guest opcode dispatch; no CPU.Step in RGB generation'
        result['device_service_sources'] = {name: sha(p/'workplace/dosgolem'/name) for name in (
            'internal/machine/host_palette.go', 'internal/machine/host_work.go', 'internal/machine/machine.go',
            'internal/machine/host_palette_test.go', 'apps/pod/drawblocks/native_palette_output_test.go',
            'apps/pod/drawblocks/calls_test.go')}
        result['normal_machine_palette_output'] = dict(plans=counts['48a1:2ac2'][0], instructions=counts['48a1:2ac2'][1],
            scope='accepted complete RGB output calls; insufficient-budget calls retain original device/ABI boundaries')
        result['hardware_contract'] = dict(source='https://www.scs.stanford.edu/10wi-cs140/pintos/specs/freevga/vga/colorreg.htm',
            scope='standard VGA DAC index and RGB ordering reused through Machine.Out8; pinned model timing, not physical-cycle parity')
    if control:
        for marker in ('source palette original entry-state comparisons 14', 'source boundary comparisons 337',
                       'source refusals 68', 'init original comparisons 24', 'init surface comparisons 18',
                       'palette stage comparisons 144', 'stage refusals 24',
                       'RGB read independent original comparisons 36',
                       'wait independent original plan comparisons 112',
                       'wait resume independent original plan comparisons 280'):
            assert marker in test_log.read_text(), marker
        result['scope'] = 'Go色盤列控制、來源與工作陣列初始化、RGB讀取、回掃兩狀態等待及階段交接；保留既有Go圖片演算法，其他固定與未驗路徑仍保留，#49未全部完成'
        result['palette_control_sources'] = {str(q.relative_to(module)): sha(q)
            for pattern in ('apps/pod/drawblocks/native_palette_*.go', 'apps/pod/drawblocks/native_wait*.go',
                            'internal/machine/host_palette_read*.go', 'internal/machine/host_video_status*.go')
            for q in sorted(module.glob(pattern))}
        result['palette_control_sources'].update({name: sha(module/name) for name in
            ('session/observer.go','session/observer_test.go','internal/machine/host_work.go',
             'internal/machine/host_palette.go','internal/machine/machine.go', 'apps/pod/drawblocks/effects.go')})
        result['prototype_and_original_tests']['palette_control'] = dict(
            source_original=14, source_boundary_comparisons=337, source_refusals=68,
            initialization_comparisons=42, initialization_refusals=27,
            stage_comparisons=144, stage_refusals=24, rgb_read_comparisons=36,
            wait_entry_comparisons=112, wait_resume_comparisons=280, wait_refusals=14,
            scope='Independent CPU/full RAM/ordered writes/model-time; RGB and wait also compare complete SaveState bytes')
        result['native_go_source_review']['palette_control_scope'] = 'Typed surface/color arrays, index cursors and RGB jobs; retrace is a two-state bounded data loop using a readonly VGA snapshot; no CPU.Step, guest-opcode dispatch or generated effect block in accepted data processing.'
        result['hardware_contract']['status_source'] = 'https://www.scs.stanford.edu/10wi-cs140/pintos/specs/freevga/vga/extreg.htm'
        result['hardware_contract']['status_scope'] = 'Existing mode13 input-status approximation reused; no physical-cycle timing claim'
    if primitive:
        for marker in ('point independent original comparisons 384', 'point surface original comparisons 72',
                       'point refusals 26', 'point committed machine original comparisons 16',
                       'region original comparisons 80', 'region parent comparisons 12',
                       'region refusals 15', 'region committed machine original comparisons 16'):
            assert marker in test_log.read_text(), marker
        result['scope'] = 'Go單點寫入／讀回與有序座標矩形區間填色；新增路徑限受控原版比較，保留正常原版事件與既有Go演算法，其他固定及未驗路徑仍保留，#49未全部完成'
        names=('native_point.go','native_point_test.go','point_commit_test.go',
               'native_region.go','native_region_test.go','region_commit_test.go')
        result['primitive_sources']={name:sha(module/'apps/pod/drawblocks'/name) for name in names}
        for name in ('native_point.go','native_region.go'):
            source=(module/'apps/pod/drawblocks'/name).read_text()
            assert '.Step(' not in source and 'effect{' not in source and 'b.run(' not in source
        result['prototype_and_original_tests']['primitives']=dict(
            point_original_comparisons=456, point_refusals=26, point_committed_machine_comparisons=16,
            region_original_comparisons=92, region_refusals=15, region_committed_machine_comparisons=16,
            scope='Controlled source-code entry states, full CPU/RAM/ordered writes/model-time; complete SaveState bytes after commit; region also rejects insufficient run budget without mutation')
        result['normal_machine_primitive_by_exit']={key:counts.get(key,[0,0]) for key in
            ('48a1:2c03','48a1:2cfb','48a1:3469')}
        result['primitive_normal_player_scope']='Controlled whole routines validate new inputs; counts above identify which exits, if any, occur in this fixed normal DEMO. No new normal-route coverage is inferred.'
        result['native_go_source_review']['primitive_scope']='Point coordinates are clipped to the selected surface and indexed through root origin/stride; rectangle pixels are clipped Go intervals and row spans. Source IP tables only preserve ABI/scratch notices, not guest opcode dispatch.'
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({k: result[k] for k in ('algorithmic_plans', 'algorithmic_instructions', 'fixed_semantic_instructions', 'guest_files')}, ensure_ascii=False))


if __name__ == '__main__':
    main()
