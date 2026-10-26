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
    ap.add_argument('--family', choices=('fill', 'rle', 'palette', 'image'), default='fill')
    a = ap.parse_args()
    assert not a.out.exists(), '保留既有收據，請另定輸出檔名'
    p = a.project
    rle = a.family != 'fill'
    image = a.family == 'image'
    palette = a.family in ('palette', 'image')
    stem = 'native'+a.family
    root = p/f'workplace/probe/draw-intercept-v1/native-{a.family}-v1'
    previous = p/'workplace/probe/draw-intercept-v1/all-functions-v1/trace-v2-legacy'
    current = root/('trace-native' if rle else 'trace-native-v2')
    old, new = read(previous/'calls.json'), read(current/'calls.json')
    for key in ('Calls', 'Digest', 'VideoCount', 'AudioCount', 'FrameCount', 'DOSCount',
                'VideoSHA256', 'AudioSHA256', 'FramesSHA256', 'DOSSHA256'):
        assert old[key] == new[key], key
    assert new['Digest']['Steps'] == 712250000 and new['VideoCount'] == 3343240
    assert tree(previous/'save') == tree(current/'save')
    witnesses = []
    for field in ('Captures', 'DrawCaptures'):
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
    fixed = new['HostInstructions'] + new['BlockInstructions'] - new['AlgorithmicInstructions']
    assert fixed > 0, '其餘固定區塊仍存在，不能稱全部Go演算法完成'
    assert new['Commands']['Mismatches'] == 0
    code = p/'workplace/input/dos/pod/VGA.DRV'
    assert sha(code) == 'eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f'
    assert hashlib.sha256(code.read_bytes()[0x41af:0x42b3]).hexdigest() == 'c7db9b7b119f14464a7e76d35125922a3559d10fef36624f79f5120dd0d02f00'
    test_log = p/f'workplace/dosgolem/workplace/native-{a.family}-product-tests.log'
    if image:
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
    tests = (dict(span_synthetic_plan_comparisons=18, span_normal_states=32, span_normal_plan_comparisons=64,
                  span_refusal_cases=16, mirror_synthetic_plan_comparisons=804, mirror_controlled_image_states=24,
                  mirror_controlled_state_plan_comparisons=144, mirror_refusal_cases=9,
                  controlled_whole_mirror_cases=24, controlled_whole_mirror_committed_plans=int(whole[1])) if image else
             dict(synthetic_plan_comparisons=60, normal_scan_states=32,
                  normal_state_plan_comparisons=192, refusal_cases=9) if palette else
             dict(packet_cases=22, packet_plan_comparisons=110, normal_image_states=32,
                  normal_state_plan_comparisons=192, refusal_cases=9) if rle else
             dict(geometries=9, cost_models=2, cases=18, plan_comparisons=194, refusal_cases=8))
    tests['log_sha256'] = sha(test_log)
    result = dict(scope=('Go未裁切圖片區塊解析、既有RLE與鏡像像素對；鏡像以受控原圖驗證，其他圖片控制、裁切與固定區塊仍保留，#49未全部完成' if image else
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
                  benchmark_scope='seven interleaved groups, 32 complete calls each, two-CPU Docker; includes source guards, compatibility writes and notifications; not FPS')
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
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({k: result[k] for k in ('algorithmic_plans', 'algorithmic_instructions', 'fixed_semantic_instructions', 'guest_files')}, ensure_ascii=False))


if __name__ == '__main__':
    main()
