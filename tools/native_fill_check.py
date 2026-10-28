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


def verify_dissolve_resolution(spec):
    text = spec.read_text()
    initial = '<a id="dissolve-4337-initial"></a>'
    ready = '<a id="dissolve-4337-ready"></a>'
    assert text.count(initial) == text.count(ready) == 1, '4337回鏈anchor缺失或重複'
    before, after = text.split(initial)[1].split(ready)
    key = 'DOS/VGA.DRV/eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f:0x4337'
    marker = '4337-READY-20261027'
    assert marker in before and '(#dissolve-4337-ready)' in before, '初始未知缺勘誤或前向連結'
    assert key in after and '48A1:4337' in after, '接入契約缺不可變原始定位'
    assert marker in after and '(#dissolve-4337-initial)' in after, '接入契約缺回鏈帳'
    return dict(key=key, correction=marker, spec_sha256=sha(spec))


def verify_hd_resolution(spec):
    text = spec.read_text()
    initial = '<a id="battle-hold-three-initial"></a>'
    ready = '<a id="battle-hold-four-ready"></a>'
    assert text.count(initial) == text.count(ready) == 1, 'HD回鏈anchor缺失或重複'
    before, after = text.split(initial)[1].split(ready)
    marker = 'BATTLE-HOLD-4-20261028'
    key = 'Go/a83043c02b28b7dc07db41a62a81300c0f0351f3/apps/pod/battle_hold.go:battleHold.applySource'
    assert marker in before and '(#battle-hold-four-ready)' in before
    assert key in after and marker in after and '(#battle-hold-three-initial)' in after
    return dict(key=key, correction=marker, spec_sha256=sha(spec))


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--stage', choices=('trace', 'full'), default='full')
    ap.add_argument('--commit', required=True)
    ap.add_argument('--family', choices=('fill', 'rle', 'palette', 'image', 'fade', 'color', 'setup', 'row', 'output', 'clip', 'control', 'primitive', 'axis', 'line', 'copy', 'ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve'), default='fill')
    ap.add_argument('--frontend-baseline', help='已驗收demo版本的GUI基準名稱前綴')
    ap.add_argument('--run-id', default='v1', help='另存新一輪候選收據，不覆寫先前失敗或完成的資料')
    a = ap.parse_args()
    assert re.fullmatch(r'v[1-9][0-9]*', a.run_id), '收據輪次不合法'
    if a.frontend_baseline:
        assert re.fullmatch(r'[a-z0-9_-]+', a.frontend_baseline), '基準名稱不合法'
    assert not a.out.exists(), '保留既有收據，請另定輸出檔名'
    p = a.project
    rle = a.family != 'fill'
    axis = a.family in ('axis', 'line', 'copy', 'ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    line = a.family in ('line', 'copy', 'ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    surface_copy = a.family in ('copy', 'ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    ellipse = a.family in ('ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    ellipsefill = a.family in ('ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    bitmap = a.family in ('bitmap', 'scale', 'surfaceflip', 'dissolve')
    scale = a.family in ('scale', 'surfaceflip', 'dissolve')
    surfaceflip = a.family in ('surfaceflip', 'dissolve')
    dissolve = a.family == 'dissolve'
    primitive = a.family in ('primitive', 'axis', 'line', 'copy', 'ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    control = a.family in ('control', 'primitive', 'axis', 'line', 'copy', 'ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    clip = a.family in ('clip', 'control', 'primitive', 'axis', 'line', 'copy', 'ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    output = a.family in ('output', 'clip', 'control', 'primitive', 'axis', 'line', 'copy', 'ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    row = a.family in ('row', 'output', 'clip', 'control', 'primitive', 'axis', 'line', 'copy', 'ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    setup = a.family in ('setup', 'row', 'output', 'clip', 'control', 'primitive', 'axis', 'line', 'copy', 'ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    color = a.family in ('color', 'setup', 'row', 'output', 'clip', 'control', 'primitive', 'axis', 'line', 'copy', 'ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    fade = a.family in ('fade', 'color', 'setup', 'row', 'output', 'clip', 'control', 'primitive', 'axis', 'line', 'copy', 'ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    image = a.family in ('image', 'fade', 'color', 'setup', 'row', 'output', 'clip', 'control', 'primitive', 'axis', 'line', 'copy', 'ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    palette = a.family in ('palette', 'image', 'fade', 'color', 'setup', 'row', 'output', 'clip', 'control', 'primitive', 'axis', 'line', 'copy', 'ellipse', 'ellipsefill', 'bitmap', 'scale', 'surfaceflip', 'dissolve')
    stem = 'native'+a.family
    artifact_stem = stem if a.run_id == 'v1' else stem+'-'+a.run_id
    root = p/f'workplace/probe/draw-intercept-v1/native-{a.family}-{a.run_id}'
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
                  binary_sha256=sha(p/f'workplace/dosgolem/workplace/bin/pod-play-{artifact_stem}'),
                  trace_binary_sha256=sha(p/(f'workplace/dosgolem/workplace/bin/draw-call-probe-{artifact_stem}'+('-v2' if not rle else ''))),
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
        demo = p/f'workplace/probe/g9-hd/{stem}-play-{a.run_id}'
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
            folder = p/f'workplace/probe/g9-hd/{stem}-ui-{a.run_id}/{route}-direct'
            baseline = p/(f'workplace/probe/g9-hd/{a.frontend_baseline}-ui-v1/{route}-direct' if a.frontend_baseline else f'workplace/probe/g9-hd/all-ui-v2/{route}-direct')
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
        before = p/(f'workplace/probe/g9-hd/{a.frontend_baseline}-save-v1/native' if a.frontend_baseline else 'workplace/probe/g9-hd/all-save-v2/native')
        after = p/f'workplace/probe/g9-hd/{stem}-save-{a.run_id}/native'
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
    if axis:
        for marker in ('axis original cases 144 independent plan comparisons 288',
                       'axis parent original comparisons 24', 'axis refusals 18',
                       'axis committed machine original comparisons 16',
                       'region row original comparisons 92', 'region row refusals 15',
                       'region bounded prefix original comparisons 48',
                       'bounded rectangle whole original comparisons 48 Go plans 858'):
            assert marker in test_log.read_text(), marker
        assert re.search(r'production bounded region whole comparisons 12 row plans [1-9]\d*',test_log.read_text())
        result['scope']='Go水平／垂直線與指定矩形逐列續行；新增路徑限受控原版比較，保留完整正常回歸與既有Go演算法，斜線及其他固定／未驗路徑仍保留，#49未全部完成'
        result['axis_and_rows_sources']={name:sha(module/'apps/pod/drawblocks'/name) for name in
            ('native_axis.go','native_axis_test.go','axis_commit_test.go','native_region_rows.go',
             'native_region_rows_test.go','region_selector_test.go')}
        for name in ('native_axis.go','native_region_rows.go'):
            source=(module/'apps/pod/drawblocks'/name).read_text()
            assert '.Step(' not in source and 'effect{' not in source and 'b.run(' not in source
        result['prototype_and_original_tests']['axis_and_rows']=dict(
            axis_plan_comparisons=312,axis_refusals=18,axis_committed_machine_comparisons=16,
            row_plan_comparisons=92,row_refusals=15,prefix_comparisons=48,
            bounded_whole_calls=48,bounded_go_plans=858,production_selector_whole_calls=12,
            scope='Controlled source entries and budget continuations; independent CPU/full RAM/ordered writes/model-time and complete SaveState comparisons')
        result['normal_machine_axis_and_rows_by_exit']={key:counts.get(key,[0,0]) for key in
            ('48a1:3235','48a1:3463','48a1:3469')}
        result['native_go_source_review']['axis_and_rows_scope']='Go coordinate sorting/clipping emits horizontal spans and a vertical pixel cursor; original eight positions only label ordered notices. Bounded rectangles resume a validated typed row cursor and retain ABI metadata separately; no guest instruction dispatch in geometry or pixels.'
    if line:
        for marker in ('steep original cases 108 independent plan comparisons 216',
                       'steep parent original comparisons 24', 'steep refusals 20',
                       'shallow original cases 108 independent plan comparisons 216',
                       'shallow parent original comparisons 24', 'shallow refusals 20',
                       'diagonal committed machine original comparisons 32',
                       'line cursor original comparisons 280', 'line cursor parent comparisons 144',
                       'line cursor refusals 44', 'diagonal prefix original comparisons 32',
                       'bounded diagonal whole original comparisons 128 Go plans 848',
                       'counter clipping pixel comparisons 5020 seed 11516',
                       'clipped cursor original comparisons 1114',
                       'clipped prefix comparisons 534 seed 11516',
                       'clipped complete machine comparisons 108 Go plans 1280'):
            assert marker in test_log.read_text(), marker
        assert test_log.read_text().count('line cursor refusals 44') == 2
        result['scope']='Go陡斜線、緩斜線、相交裁切與有界像素游標；新增路徑限受控原版比較，完整正常回歸保留，非相交及其他固定／未驗路徑仍保留，#49未全部完成'
        names=('native_line.go','native_diagonal.go','native_line_cursor.go','native_line_clip.go',
               'native_diagonal_test.go','native_line_cursor_test.go','diagonal_commit_test.go',
               'native_line_clip_test.go','clipped_cursor_test.go','clipped_prefix_test.go','clipped_commit_test.go')
        result['line_sources']={name:sha(module/'apps/pod/drawblocks'/name) for name in names}
        for name in names[:4]:
            source=(module/'apps/pod/drawblocks'/name).read_text()
            assert '.Step(' not in source and 'effect{' not in source and 'b.run(' not in source
        result['prototype_and_original_tests']['lines']=dict(
            interior_plans=480,interior_refusals=40,interior_committed_machine_comparisons=32,
            cursor_plans=424,cursor_refusals=44,prefix_comparisons=32,bounded_whole_calls=128,bounded_go_plans=848,
            clip_pixel_cases=5020,clip_seed=11516,clip_cursor_plans=1114,clip_cursor_refusals=44,
            clip_prefix_comparisons=534,clip_production_whole_calls=108,clip_production_selected_plans=1280,
            clip_production_go_plans=898,clip_production_fixed_plans=382,
            scope='Controlled original ABI states and cursors; full CPU/RAM/ordered writes/model-time and complete SaveState comparisons; not new normal-player-route evidence')
        result['native_go_source_review']['line_scope']='Go coordinate/error arithmetic generates pixels. Typed counters retain the original narrow-tall clip behavior. Source positions label ABI/scratch/model-time notices only; no guest opcode dispatch in geometry or pixel generation.'
    if line:
        audit_log=module/'workplace/plan-class-audit.log'
        assert 'clipped class audit comparisons 108 algorithmic 898 fixed 382' in audit_log.read_text()
        result['plan_class_audit']=dict(log_sha256=sha(audit_log),
            clipped_test_sha256=sha(module/'workplace/palette-advance-research/clipped-plan-class-audit_test.go'),
            copy_test_sha256=sha(module/'workplace/palette-advance-research/copy-plan-class-audit_test.go'),
            scope='Audits exact source-bound observed copies with empty extra-entry map; original selected-plan test labels included compatibility plans. Pixel/full-machine comparisons remain independent.')
    if surface_copy:
        for marker in ('copy row independent comparisons 1120',
                       'copy parent and unequal surface comparisons 64',
                       'copy row refusals 44', 'copy prefix independent comparisons 168',
                       'copy prefix refusals 16', 'copy whole machine comparisons 192 Go plans 1040'):
            assert marker in test_log.read_text(),marker
        assert 'copy class audit comparisons 192 algorithmic 448 fixed 592' in audit_log.read_text()
        result['scope']='Go表面矩形搬移，包含重疊、四方向與奇數尾byte；新增路徑限受控原版比較，完整正常回歸保留，其他固定／未驗路徑仍保留，#49未全部完成'
        names=('native_surface_copy.go','native_surface_copy_prefix.go','native_surface_copy_test.go',
               'surface_copy_boundary_test.go','surface_copy_commit_test.go','surface_copy_prefix_test.go','surface_copy_refusal_test.go')
        result['surface_copy_sources']={name:sha(module/'apps/pod/drawblocks'/name) for name in names}
        for name in names[:2]:
            source=(module/'apps/pod/drawblocks'/name).read_text()
            assert '.Step(' not in source and 'effect{' not in source and 'b.run(' not in source
        result['prototype_and_original_tests']['surface_copy']=dict(
            row_plans=1120,prefix_plans=168,parent_and_unequal_plans=64,row_refusals=44,prefix_refusals=16,
            controlled_whole_calls=192,algorithmic_plans=448,fixed_compatibility_plans=592,selected_plans=1040,
            scope='Original full CPU/RAM/ordered writes/model-time and SaveState; selected-plan label corrected by Algorithmic audit; not new normal-player-route evidence')
        result['normal_machine_surface_copy_by_exit']={'48a1:41ae':counts.get('48a1:41ae',[0,0])}
        result['native_go_source_review']['surface_copy_scope']='Typed source/destination rectangles and row direction generate word/byte transfers. Go shadow reads preserve overlap dependencies. Fixed IP/step labels preserve notices only; geometry does not dispatch guest instructions.'
    if ellipse:
        for marker in ('ellipse ordered pixel comparisons 321 seed 13418',
                       'ellipse first cursor original comparisons 316',
                       'ellipse second cursor original comparisons 180',
                       'ellipse transition cursor original comparisons 24',
                       'ellipse prefix original comparisons 24',
                       'ellipse whole machine comparisons 72 algorithmic plans 1632',
                       'ellipse boundary whole comparisons 32 algorithmic plans 3832',
                       'ellipse DF whole comparisons 72 algorithmic plans 1632',
                       'ellipse one-step Go pixel comparisons 912',
                       'ellipse cursor and pixel refusals 57', 'ellipse prefix refusals 18',
                       'ellipse parent prefix and machine comparisons 8',
                       'ellipse production-shape selector comparisons 96 algorithmic plans 3772 CPU metadata steps 63780 fixed metadata plans 9250'):
            assert marker in test_log.read_text(),marker
        result['scope']='Go正半徑橢圓外框、兩段中點幾何及單步像素；新增路徑限受控原版比較，完整正常回歸保留，零半徑、填滿橢圓及其他未驗路徑仍保留，#49未全部完成'
        names=sorted(f.name for f in (module/'apps/pod/drawblocks').glob('*.go')
                     if f.name.startswith('native_ellipse') or f.name.startswith('ellipse_'))
        assert len(names)==19
        result['ellipse_sources']={name:sha(module/'apps/pod/drawblocks'/name) for name in names}
        for name in names:
            if name.endswith('_test.go'):continue
            source=(module/'apps/pod/drawblocks'/name).read_text()
            assert '.Step(' not in source and 'effect{' not in source and 'b.run(' not in source
        result['prototype_and_original_tests']['ellipse']=dict(
            pixel_cases=321,first_cursor_plans=316,second_cursor_plans=180,transition_plans=24,prefix_plans=24,
            full_native_calls=72,full_native_plans=1632,boundary_calls=32,boundary_plans=3832,
            df_calls=72,df_plans=1632,single_step_pixels=912,cursor_pixel_refusals=57,prefix_refusals=18,
            parent_comparisons=8,short_selector_calls=96,short_algorithmic_plans=3772,
            short_fixed_metadata_plans=9250,short_cpu_metadata_steps=63780,
            scope='Controlled original full CPU/RAM/ordered pixels/model-time and SaveState; short budgets retain CPU/fixed metadata but no emulated ellipse pixel writes; not new normal-player-route evidence')
        result['normal_machine_ellipse_by_exit']={key:counts.get(key,[0,0]) for key in
            ('48a1:35f1','48a1:37ba','48a1:37f5','48a1:39aa','48a1:39b1',
             '48a1:3660','48a1:36b6','48a1:370c','48a1:3762','48a1:384e','48a1:38a4','48a1:38fa','48a1:3950')}
        result['native_go_source_review']['ellipse_scope']='Go midpoint error/gradient arithmetic generates four ordered quadrants. Source positions preserve ABI/scratch/model-time notices. Single-step fallback recomputes positions from the same Go geometry; fixed blocks with ellipse pixel writes are rolled back before commit, metadata may continue through CPU. No guest opcode dispatcher in geometry.'
    if ellipsefill:
        for marker in ('filled ellipse ordered pixel comparisons 321 seed 14770',
                       'filled ellipse first cursor original comparisons 316',
                       'filled ellipse second cursor original comparisons 180',
                       'filled ellipse transition cursor original comparisons 24',
                       'filled ellipse prefix original comparisons 24',
                       'filled ellipse whole machine comparisons 72 algorithmic plans 1632',
                       'filled ellipse boundary whole comparisons 32 algorithmic plans 3832',
                       'filled ellipse DF whole comparisons 72 algorithmic plans 1632',
                       'filled ellipse one-step span comparisons 1184',
                       'filled ellipse cursor and span refusals 60',
                       'filled ellipse prefix refusals 18',
                       'filled ellipse count/address/color refusals 6',
                       'filled ellipse parent prefix and machine comparisons 8',
                       'filled ellipse selector comparisons 96 algorithmic plans 4150 CPU metadata steps 54510 fixed metadata plans 6260'):
            assert marker in test_log.read_text(),marker
        result['scope']='Go正半徑填滿橢圓、對稱水平區間及單步span；新增路徑限受控原版比較，完整正常回歸保留，零半徑及其他未驗路徑仍保留，#49未全部完成'
        names=sorted(f.name for f in (module/'apps/pod/drawblocks').glob('*.go')
                     if f.name.startswith('native_filled_ellipse') or f.name.startswith('filled_ellipse_'))
        assert len(names)==19
        result['filled_ellipse_sources']={name:sha(module/'apps/pod/drawblocks'/name) for name in names}
        for name in names:
            if name.endswith('_test.go'):continue
            source=(module/'apps/pod/drawblocks'/name).read_text()
            assert '.Step(' not in source and 'effect{' not in source and 'b.run(' not in source
        result['prototype_and_original_tests']['filled_ellipse']=dict(
            pixel_cases=321,first_cursor_plans=316,second_cursor_plans=180,transition_plans=24,prefix_plans=24,
            full_native_calls=72,full_native_plans=1632,boundary_calls=32,boundary_plans=3832,
            df_calls=72,df_plans=1632,single_step_spans=1184,cursor_span_refusals=60,prefix_refusals=18,
            span_count_address_color_refusals=6,parent_comparisons=8,short_selector_calls=96,
            short_algorithmic_plans=4150,short_fixed_metadata_plans=6260,short_cpu_metadata_steps=54510,
            scope='Original CPU/full RAM/ordered pixels/model-time and SaveState; short budgets retain control metadata, all actual filled-ellipse pixel transfers generated by Go. Zero-count transfers preserve model-time; not new normal-player-route evidence.')
        result['normal_machine_filled_ellipse_by_exit']={key:counts.get(key,[0,0]) for key in
            ('48a1:3b3a','48a1:3c84','48a1:3cbf','48a1:3df5','48a1:3dfc',
             '48a1:3bd8','48a1:3bdc','48a1:3c2a','48a1:3c2e','48a1:3d47','48a1:3d4b','48a1:3d99','48a1:3d9d')}
        result['native_go_source_review']['filled_ellipse_scope']='Go midpoint arithmetic generates two clipped ordered horizontal intervals, including duplicate rows and odd tails. Single-step spans derive counts and positions from the same geometry. Fixed pixel-transfer blocks are rolled back; CPU/fixed control metadata separately counted. Source IPs label notices, not an opcode dispatcher.'
    if bitmap:
        for marker in ('bitmap51 chunky row original comparisons 80',
                       'bitmap51 planar prefix comparisons 80',
                       'bitmap51 planar group comparisons 336',
                       'bitmap51 single-step span comparisons 4576 row advances 160',
                       'bitmap51 whole original comparisons 400 algorithmic plans 32492 fixed metadata plans 154444 CPU metadata steps 499008',
                       'bitmap51 boundary full machine comparisons 64',
                       'bitmap51 refused invalid states 150'):
            assert marker in test_log.read_text(), marker
        assert hashlib.sha256(code.read_bytes()[0x51cc:0x5414]).hexdigest() == 'dd209e837a213ee07c0b9602563167e0eed60702d8a96f4c20d98f27ebb157a2'
        result['scope']='Go八平面色號組合、直接色號列及短預算像素輸出；新增路徑限受控原版比較，資料查找與準備維持原流程，其他未驗路徑仍保留，#49未全部完成'
        names=sorted(f.name for f in (module/'apps/pod/drawblocks').glob('native_bitmap51*.go'))
        assert len(names)==10
        result['bitmap_row_sources']={name:sha(module/'apps/pod/drawblocks'/name) for name in names}
        for name in names:
            if name.endswith('_test.go'): continue
            source=(module/'apps/pod/drawblocks'/name).read_text()
            assert '.Step(' not in source and 'effect{' not in source and 'b.run(' not in source
        result['prototype_and_original_tests']['bitmap_row']=dict(
            direct_rows=80,planar_prefixes=80,planar_groups=336,
            single_step_spans=4576,row_advances=160,whole_calls=400,
            whole_algorithmic_plans=32492,whole_fixed_metadata_plans=154444,
            whole_cpu_metadata_steps=499008,boundary_calls=64,refusals=150,
            scope='Controlled original full CPU/RAM/ordered output/model-time and SaveState. Five budgets; no CPU or fixed pixel output. BMHD/BODY search and prepared source rows remain original; not new normal-player-route evidence.')
        result['normal_machine_bitmap_by_exit']={key:counts.get(key,[0,0]) for key in
            ('48a1:53c3','48a1:53ed','48a1:53d8','48a1:53f6','48a1:53fa','48a1:5408','48a1:5413')}
        result['native_go_source_review']['bitmap_row_scope']='Typed source row and surface cursor; direct Go copies and eight-plane bit composition. One-step output reconstructs data from the same source. Fixed raster blocks roll back before commit. CPU/fixed source preparation and metadata are counted separately. Original IPs label ABI notices and model costs; no opcode dispatcher.'
    if scale:
        for marker in ('scale5615 ordered pixel comparisons 125',
                       'scale5615 prefix comparisons 250',
                       'scale5615 prepared row comparisons 930',
                       'scale5615 output group comparisons 930',
                       'scale5615 single-step span comparisons 10488',
                       'scale5615 full Go machine comparisons 250 algorithmic plans 2110',
                       'scale5615 short selector comparisons 1000 algorithmic plans 38610 fixed metadata plans 34268 CPU metadata steps 320562',
                       'scale5615 boundary parent whole comparisons 144 algorithmic plans 576',
                       'scale5615 invalid state refusals 88'):
            assert marker in test_log.read_text(), marker
        assert hashlib.sha256(code.read_bytes()[0x5615:0x5886]).hexdigest() == '6d9aac63d38f8b0e5ddd9bdbdfcfdda9144318a34398436b6b468c52f0271106'
        result['scope']='Go表面整數縮放、置中取樣、補色、重複列及短預算scratch與畫面輸出；新增路徑限受控原版比較，其他未驗範圍保留，#49未全部完成'
        names=sorted(f.name for f in (module/'apps/pod/drawblocks').glob('native_scale5615*.go'))
        assert len(names)==15
        result['surface_scale_sources']={name:sha(module/'apps/pod/drawblocks'/name) for name in names}
        for name in names:
            if name.endswith('_test.go'): continue
            source=(module/'apps/pod/drawblocks'/name).read_text()
            assert '.Step(' not in source and 'effect{' not in source and 'b.run(' not in source
        result['prototype_and_original_tests']['surface_scale']=dict(
            ordered_pixels=125,prefixes=250,prepared_rows=930,output_groups=930,
            single_step_spans=10488,full_go_calls=250,full_go_plans=2110,
            short_calls=1000,short_algorithmic_plans=38610,short_fixed_metadata_plans=34268,
            short_cpu_metadata_steps=320562,boundary_parent_calls=144,
            boundary_parent_plans=576,refusals=88,
            scope='Controlled full original CPU/RAM/ordered output/model-time and SaveState. Source row sampled before overlapping output; five integer repeat factors and four short budgets. No CPU/fixed scaling scratch or display writes; not new normal-player-route evidence.')
        result['normal_machine_scale_by_exit']={key:counts.get(key,[0,0]) for key in
            ('48a1:5798','48a1:5819','48a1:587f','48a1:586b','48a1:5885','48a1:5813','48a1:5879','48a1:587d','48a1:5856','48a1:585a')}
        result['native_go_source_review']['surface_scale_scope']='Typed source/destination surfaces and centered integer sampling grid. Go prepares clipped/fill samples and repeats rows, preserving overlap order and extra scratch writes. One-step plans derive the same samples/copy geometry. Fixed pixel blocks roll back before commit; CPU/fixed control metadata separately counted. Source IPs label notices/costs, no opcode dispatcher.'
    if surfaceflip:
        for marker in ('flip5886 ordered pixel comparisons 160',
                       'flip5886 prefix comparisons 320',
                       'flip5886 paired row comparisons 656 address plans 656',
                       'flip5886 one-step span comparisons 63952',
                       'flip5886 whole Go machine comparisons 320 algorithmic plans 1392',
                       'flip5886 selector comparisons 160 algorithmic plans 8022 fixed metadata plans 1722 CPU metadata steps 33254',
                       'flip5886 boundary parent whole comparisons 64 algorithmic plans 3336',
                       'flip5886 invalid state refusals 174'):
            assert marker in test_log.read_text(), marker
        assert hashlib.sha256(code.read_bytes()[0x5886:0x5a28]).hexdigest() == '2ff20b76959b7f3972f76530c6142c8337d0f935893752f9a23d417dde2eac36'
        result['scope']='Go表面原地翻轉、成對列、同列讀寫相依與短預算scratch／畫面輸出；新增路徑限受控原版比較，其他未驗範圍保留，#49未全部完成'
        names=sorted(f.name for f in (module/'apps/pod/drawblocks').glob('native_flip5886*.go'))
        assert len(names)==14
        result['surface_flip_sources']={name:sha(module/'apps/pod/drawblocks'/name) for name in names}
        for name in names:
            if name.endswith('_test.go'): continue
            source=(module/'apps/pod/drawblocks'/name).read_text()
            assert '.Step(' not in source and 'effect{' not in source and 'b.run(' not in source
        result['prototype_and_original_tests']['surface_flip']=dict(
            ordered_pixels=160,prefixes=320,paired_rows=656,address_plans=656,
            single_step_spans=63952,full_go_calls=320,full_go_plans=1392,
            short_calls=160,short_algorithmic_plans=8022,short_fixed_metadata_plans=1722,
            short_cpu_metadata_steps=33254,boundary_parent_calls=64,
            boundary_parent_plans=3336,refusals=174,
            scope='Controlled original full CPU/RAM/ordered writes/model-time and SaveState. Intermediate same-row read-after-write and original odd-height middle row retained. No CPU/fixed flip scratch or display writes; not new normal-player-route evidence.')
        result['normal_machine_surface_flip_by_exit']={key:counts.get(key,[0,0]) for key in
            ('48a1:597a','48a1:59a7','48a1:5a20','48a1:5a27','48a1:59c8','48a1:59cf','48a1:59d3','48a1:59f1','48a1:59f8','48a1:59fc','48a1:5a0c','48a1:5a10')}
        result['native_go_source_review']['surface_flip_scope']='Typed surface/row pair; Go reads and mirrors samples with ordered shadow reads for aliasing, saves a row and performs both output phases. One-step plans reconstruct sources and destinations. Fixed pixel blocks roll back before commit; CPU/fixed control metadata counted separately. Source IPs label notices and model costs; no opcode dispatcher.'
    if dissolve:
        for marker in ('dissolve4337 ordered pixel comparisons 9 seed 1',
                       'dissolve4337 prefix comparisons 72',
                       'dissolve4337 cursor comparisons 432',
                       'dissolve4337 bounded batch comparisons 680',
                       'dissolve4337 single-step pixel comparisons 1584 control plans 2242',
                       'dissolve4337 whole machine comparisons 44 algorithmic plans 3028 shared VGA wait plans 1404 CPU control steps 2808',
                       'dissolve4337 selector comparisons 40 algorithmic plans including shared wait 39682 fixed metadata plans 7148 CPU control steps 50084',
                       'dissolve4337 DF prefix and machine comparisons 16',
                       'dissolve4337 boundary parent machine comparisons 32 algorithmic plans 15480 shared wait plans 2336 CPU control steps 4672',
                       'dissolve4337 invalid state refusals 86',
                       'dissolve4337 one-pixel bounded continuation comparisons 4 cutoff 10000 seed 1'):
            assert marker in test_log.read_text(), marker
        assert hashlib.sha256(code.read_bytes()[0x4337:0x45e0]).hexdigest() == '6888a66a54dd072b40f5c95f990d81d53f75257852bce465c95dff15e1168796'
        result['scope'] = 'Go表面像素轉場、LFSR樣本、有序重疊讀回與短預算輸出；新增路徑限受控原版比較，正常未觀測入口與其他範圍不外推，#49未全部完成'
        names = sorted(f.name for f in (module/'apps/pod/drawblocks').glob('native_dissolve4337*.go'))
        assert len(names) == 17 and sum(not n.endswith('_test.go') for n in names) == 6
        result['dissolve_sources'] = {n: sha(module/'apps/pod/drawblocks'/n) for n in names}
        for name in names:
            if name.endswith('_test.go'):
                continue
            source = (module/'apps/pod/drawblocks'/name).read_text()
            assert '.Step(' not in source and 'effect{' not in source and 'b.run(' not in source
        result['resolution_backlinks'] = verify_dissolve_resolution(p/'docs/spec/025-direct-render-commands.md')
        build = (root/'build.txt').read_text()
        assert f'vcs.revision={a.commit}' in build and 'vcs.modified=false' in build
        result['build_info_sha256'] = sha(root/'build.txt')
        result['prototype_and_original_tests']['dissolve'] = dict(
            pixel_cases=9, pixel_seed=1, prefixes=72, cursors=432, batches=680,
            single_step_pixels=1584, single_step_control_plans=2242,
            full_calls=44, full_dissolve_plans=3028, full_shared_wait_plans=1404,
            full_cpu_control_steps=2808, short_calls=40,
            short_algorithmic_plans_including_wait=39682, short_fixed_control_plans=7148,
            short_cpu_control_steps=50084, df_calls=16, boundary_calls=32,
            boundary_dissolve_plans=15480, boundary_shared_wait_plans=2336,
            boundary_cpu_control_steps=4672, refusals=86,
            one_pixel_bounded_cases=4, one_pixel_step_limit=10000,
            scope='Controlled original full CPU/RAM/ordered writes/model-time/SaveState. Go samples the original LFSR with immediate ordered readback for overlap. Shared wait and fixed/CPU call control counted separately; no CPU/fixed dissolve pixels. Original one-pixel nontermination retained within the bounded interval; not normal-player-route evidence.')
        result['normal_machine_dissolve_by_exit'] = {k: v for k, v in counts.items()
            if k.startswith('48a1:') and 0x4337 <= int(k.split(':')[1], 16) < 0x45e0}
        result['native_go_source_review']['dissolve_scope'] = 'Typed surfaces and 32-bit LFSR sample selection generate pixels in Go, with ordered shadow reads and bounded batches. One-step plans derive the same sample. Source IPs label ABI/model-time notices; no guest opcode or instruction-effect dispatcher in raster generation.'
        result['tools']['tools/vga-battle-play-receipt.sh'] = sha(p/'tools/vga-battle-play-receipt.sh')
        if a.run_id != 'v1':
            result['hd_resolution_backlinks'] = verify_hd_resolution(p/'docs/spec/010-hd-theme.md')
            result['presentation_sources'] = {name: sha(module/name) for name in (
                'apps/pod/hd_layers.go', 'apps/pod/hd_frame_cache_test.go',
                'apps/pod/battle_hold.go', 'apps/pod/vga_hd_fixture_test.go',
                'apps/pod/explicit_demo_hd_test.go')}
            product_log = module/'workplace/hd-presentation-product-tests-v2.log'
            for marker in (
                'direct=false original flashes=1 fixed flashes=0 frames=11 changed frames=1 changed cells=1',
                'direct=true original flashes=1 fixed flashes=0 frames=11 changed frames=1 changed cells=1',
                'HD cached/uncached original output comparisons 22',
                '--- PASS: TestVGABattleTransitionFixture',
                '--- PASS: TestVGABattleHoldUnknownBoundAndPalette',
                '--- PASS: TestVGABattleHoldBound'):
                assert marker in product_log.read_text(), marker
            assert '\nPASS\n' in product_log.read_text()
            result['presentation_test_log_sha256'] = sha(product_log)
            result['presentation_source_proof_sha256'] = sha(p/'docs/re/021-hd-frame-presentation.json')
        if a.stage == 'full':
            assert a.frontend_baseline, '新demo必須指定已驗收版本的GUI基準'
            gui = p/'workplace/probe/g9-hd'
            old_demo = gui/f'{a.frontend_baseline}-play-v1'
            expected = read(p/'docs/re/020-demo-control-runtime.json')['binary_sha256']
            for mode in ('hd', 'base', 'english', 'twox'):
                assert read(old_demo/mode/'inputs.json')['binary_sha256'] == expected
                assert (demo/mode/'script.txt').read_bytes() == (old_demo/mode/'script.txt').read_bytes()
                assert '1800:d' in (demo/mode/'script.txt').read_text()
            for route in ('wall', 'movie', 'title'):
                assert read(gui/f'{a.frontend_baseline}-ui-v1/{route}-direct/inputs.json')['binary_sha256'] == expected
            assert read(gui/f'{a.frontend_baseline}-save-v1/native/before.json')['binary_sha256'] == expected
            result['frontend_baseline'] = dict(prefix=a.frontend_baseline,
                binary_sha256=expected, demo_observation_sha256=sha(old_demo/'baseline-observation.json'),
                scope='Current accepted demo controls with explicit D; earlier timeout-driven GUI receipts remain historical evidence.')
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({k: result[k] for k in ('algorithmic_plans', 'algorithmic_instructions', 'fixed_semantic_instructions', 'guest_files')}, ensure_ascii=False))


if __name__ == '__main__':
    main()
