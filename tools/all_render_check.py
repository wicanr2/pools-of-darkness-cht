"""驗證#47的完整有序事件、正常視窗及真存讀檔。只寫全新收據。"""
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


def module(path, name):
    spec = importlib.util.spec_from_file_location(name, path)
    obj = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(obj)
    return obj


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--commit', required=True)
    ap.add_argument('--stage', choices=('trace', 'full'), default='full')
    a = ap.parse_args()
    assert not a.out.exists(), '保留既有收據，請換輸出檔名'
    p = a.project
    root = p/'workplace/probe/draw-intercept-v1/all-functions-v1'
    folders = [root/f'trace-v2-{mode}' for mode in ('legacy', 'native')]
    old, new = [read(d/'calls.json') for d in folders]
    for key in ('Calls', 'Digest', 'VideoCount', 'AudioCount', 'FrameCount', 'DOSCount',
                'VideoSHA256', 'AudioSHA256', 'FramesSHA256', 'DOSSHA256'):
        assert old[key] == new[key], key
    assert new['Digest']['Steps'] == 712250000 and new['VideoCount'] == 3343240
    assert new['HostCalls'] > 0 and new['HostBlocks'] > 0 and new['BlockInstructions'] > 0
    assert old['HostCalls'] == old['HostBlocks'] == 0
    assert tree(folders[0]/'save') == tree(folders[1]/'save')
    witnesses = []
    for field in ('Captures', 'DrawCaptures'):
        left = {(c['Seg'][1], c['IP'], c['Memory']): c for c in old[field]}
        for c in new[field]:
            identity = (c['Seg'][1], c['IP'], c['Memory'])
            assert identity in left and c == left[identity], identity
            assert sha(folders[0]/c['Memory']) == sha(folders[1]/c['Memory']) == c['MemorySHA256']
            witnesses.append(dict(cs=c['Seg'][1], ip=c['IP'], source=c['Memory'],
                                  memory_sha256=c['MemorySHA256']))
    assert len(new['Captures']) == 32 and len(new['DrawCaptures']) > 0
    commands = new['Commands']
    assert commands['Mismatches'] == 0 and commands['HostBytes'] > 0
    assert all(w['HostBytes'] > 0 for w in commands['Writers'].values()), '存在未接管的實際像素寫入者'
    table_path = p/'workplace/dosgolem/apps/pod/drawblocks/blocks.json'
    table = read(table_path)
    assert table['generator_sha256'] == sha(p/'tools/draw_blocks_gen.py')
    oracle = read(root/'block-oracle-final-current.json')
    assert oracle['blocks'] == len(table['blocks']) == 1432
    assert oracle['tested_blocks'] == 1430 and oracle['cases'] == 12010
    result = dict(scope='本作VGA繪圖接管；完整區塊及三支共享常式，保留硬體／中斷與事件邊界',
                  program_commit=a.commit,
                  binary_sha256=sha(p/'workplace/dosgolem/workplace/bin/pod-play-all'),
                  trace_binary_sha256=sha(p/'workplace/dosgolem/workplace/bin/draw-call-probe-all'),
                  source_reports={d.name: sha(d/'calls.json') for d in folders},
                  file_time='2000-01-01T00:00:00Z',
                  virtual_steps=new['Digest']['Steps'], host_calls=new['HostCalls'],
                  host_instructions=new['HostInstructions'], host_blocks=new['HostBlocks'],
                  block_instructions=new['BlockInstructions'],
                  video_writes=new['VideoCount'], video_sha256=new['VideoSHA256'],
                  retraces=new['FrameCount'], retraces_sha256=new['FramesSHA256'],
                  audio_events=new['AudioCount'], audio_sha256=new['AudioSHA256'],
                  dos_requests=new['DOSCount'], dos_sha256=new['DOSSHA256'],
                  full_state_equal=True,
                  digest={k: bytes(v).hex() if isinstance(v, list) else v for k, v in new['Digest'].items()},
                  guest_files=len(tree(folders[0]/'save')), witnesses=witnesses,
                  commands=commands, function_entries=len(table['functions']),
                  driver_functions=sum(f[0] == 0x48a1 for f in table['functions']),
                  shared_memory_functions=sum(f[0] != 0x48a1 for f in table['functions']),
                  blocks=len(table['blocks']), original_instruction_boundaries=len(table['original_only']),
                  block_oracle=oracle, block_index_sha256=sha(table_path),
                  actual_whole_function_cases=12, controlled_whole_function_cases=192,
                  tools={name: sha(p/'tools'/name) for name in
                         ('all_render_check.py', 'draw_blocks_gen.py', 'draw_call_probe.go',
                          'whole-call-save-receipt.sh', 'direct-ui-receipt.sh')},
                  limits=['B000及其他顯示模式留原版；硬體、POPF及未驗控制轉移由原CPU執行',
                          '函式及區塊驗收不代表全遊戲路線或全量HD完成',
                          '有界並行重播的wall time不用作FPS結論'])
    if a.stage == 'full':
        gui = p/'workplace/probe/g9-hd/all-play-v2'
        demo = read(gui/'bigpic-summary.json')
        assert demo['binary_sha256'] == result['binary_sha256']
        assert demo['prior_original_state_equal'] and demo['prior_sample_composite_equal']
        result['normal_demo'] = demo
        result['normal_demo_draws'] = {}
        for case in ('hd', 'base', 'english', 'twox'):
            draw = read(gui/case/'draw-report.json')
            assert draw['HostBlocks'] > 0 and draw['BlockInstructions'] > 0 and draw['Mismatches'] == 0
            result['normal_demo_draws'][case] = {k: draw[k] for k in
                                               ('HostCalls', 'HostInstructions', 'HostBlocks',
                                                'BlockInstructions', 'HostBytes', 'Mismatches')}
        ui = module(p/'tools/direct_ui_check.py', 'ui_check')
        rows = []
        for route in ('wall', 'movie', 'title'):
            current = p/f'workplace/probe/g9-hd/all-ui-v2/{route}-direct'
            previous = p/f'workplace/probe/g9-hd/direct-ui/{route}-direct'
            assert ui.endpoint(current) == ui.endpoint(previous), route
            ci, pi = [read(d/'inputs.json') for d in (current, previous)]
            for key in ('orig', 'font', 'text', 'theme'):
                assert ci[key] == pi[key], (route, key)
            front = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt',
                     'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
            guests = [{k: v for k, v in tree(d/'save').items() if k not in front}
                      for d in (current, previous)]
            assert guests[0] == guests[1], route
            for name in ('source.idx', 'source.pal', 'original.rgba', 'frame.png'):
                assert (current/name).read_bytes() == (previous/name).read_bytes(), (route, name)
            for name in ('shots', 'window-shots', 'sources'):
                assert tree(current/name) and tree(current/name) == tree(previous/name), (route, name)
            measured = ui.audit(current, route)
            assert measured == ui.audit(previous, route)
            draw = read(current/'draw-report.json')
            assert draw['HostBlocks'] > 0 and draw['Mismatches'] == 0
            rows.append(dict(route=route, endpoint=ui.endpoint(current), screenshot_pairs=len(tree(current/'shots')),
                             all_screenshots_equal=True, guest_files=len(guests[0]), commands=draw, **measured))
        result['normal_ui'] = rows
        before = p/'workplace/probe/g9-hd/whole-save/legacy'
        after = p/'workplace/probe/g9-hd/all-save-v2/native'
        inputs = [read(d/'before.json') for d in (before, after)]
        for j in inputs:
            j.pop('binary_sha256')
        assert inputs[0] == inputs[1] and 'SAVE/savgamB.pty' not in inputs[1]
        saved = [read(d/'after-save.json') for d in (before, after)]
        assert saved[0] == saved[1] and tree(before/'game/SAVE') == tree(after/'game/SAVE')
        original_slot = {k[5:]: v for k, v in inputs[1].items() if k.startswith('SAVE/')}
        assert all(saved[1][key] == val for key, val in original_slot.items())
        phases = []
        for phase, frames in (('save', 4700), ('load', 3250)):
            ends = []
            for d in (before, after):
                log = (d/f'{phase}.log').read_text()
                first = re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)', log)
                last = re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)', log)
                assert first and last and int(first[1]) == frames
                assert '字型缺字' not in log and '截圖失敗' not in log
                ends.append(first.groups()+last.groups())
            assert ends[0] == ends[1]
            for suffix in ('final.rgba', 'source.idx', 'source.pal'):
                assert (before/f'{phase}-{suffix}').read_bytes() == (after/f'{phase}-{suffix}').read_bytes()
            assert tree(before/f'{phase}-shots') == tree(after/f'{phase}-shots')
            draw = read(after/f'{phase}-draw.json')
            assert draw['HostBlocks'] > 0 and draw['Mismatches'] == 0
            phases.append(dict(phase=phase, endpoint=ends[1], host_blocks=draw['HostBlocks'],
                               block_instructions=draw['BlockInstructions']))
        result['normal_save_roundtrip'] = dict(slot_b_created=True, slot_b_loaded=True,
                                               slot_a_unchanged_files=len(original_slot), all_save_bytes_equal=True,
                                               save_files=len(saved[1]), screenshot_pairs=6,
                                               slot_b_sha256=saved[1]['savgamB.pty'], phases=phases)
        result['benchmark'] = read(root/'function-cost-v5.json')
        result['benchmark_source_sha256'] = sha(p/'workplace/dosgolem/apps/pod/drawblocks/calls_test.go')
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({k: result[k] for k in ('host_blocks', 'block_instructions', 'video_writes', 'retraces')}, ensure_ascii=False))


if __name__ == '__main__':
    main()
