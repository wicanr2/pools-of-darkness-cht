#!/usr/bin/env python3
"""驗證整段字模呼叫的真實機器效果、通知串流與正常視窗。"""
import argparse
import hashlib
import json
from pathlib import Path
import re


def sha(data):
    return hashlib.sha256(data).hexdigest()


def tree(root):
    return {str(p.relative_to(root)): sha(p.read_bytes())
            for p in sorted(root.rglob('*')) if p.is_file()}


def endpoint(folder):
    log = (folder/'run.log').read_text()
    a = re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)', log)
    b = re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)', log)
    assert a and b
    return a.groups() + b.groups()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--project', type=Path, required=True)
    parser.add_argument('--out', type=Path, required=True)
    parser.add_argument('--gui', action='store_true')
    parser.add_argument('--save', action='store_true')
    a = parser.parse_args()
    root = a.project/'workplace/probe/draw-intercept-v1'
    folders = [root/f'whole-calls-v4-{name}' for name in ('legacy', 'native')]
    legacy, native = [json.loads((p/'calls.json').read_text()) for p in folders]
    assert not legacy['WholeCall'] and native['WholeCall']
    for key in ('Digest', 'Calls', 'Captures', 'VideoCount', 'AudioCount', 'FrameCount', 'DOSCount',
                'VideoSHA256', 'AudioSHA256', 'FramesSHA256', 'DOSSHA256'):
        assert legacy[key] == native[key], key
    assert tree(folders[0]/'save') == tree(folders[1]/'save')
    previous = json.loads((root/'native-audit-demo/report.json').read_text())
    assert native['Digest'] == previous['Digest']
    assert native['AudioSHA256'] == previous['AudioSHA256']
    assert native['DOSSHA256'] == previous['DOSSHA256']
    assert native['HostCalls'] > 0 and native['HostInstructions'] > 0
    assert legacy['HostCalls'] == legacy['HostInstructions'] == 0
    assert native['VideoCount'] == 3343240 and native['FrameCount'] == 8633
    for report in (legacy, native):
        assert report['Commands']['Mismatches'] == 0
    witnesses = []
    for c in native['Captures']:
        assert sha((folders[0]/c['Memory']).read_bytes()) == sha((folders[1]/c['Memory']).read_bytes()) == c['MemorySHA256']
        witnesses.append(dict(source=c['Memory'], memory_sha256=c['MemorySHA256']))
    result = dict(scope='既有7400更新正常DEMO；CALL後整段接管；不同時刻的部分狀態不以終點代替，另比全部回掃與有序事件',
                  program_commit='64e734c92ecaa096be9bb91d6e6f792fd5fad7fd',
                  binary_sha256='d3f97dfb97ce3951da6dae880d39cbc5cb22e28532d38dc157a5ed59d93947f7',
                  file_time='2000-01-01T00:00:00Z', observed_calls=native['Calls'], host_calls=native['HostCalls'],
                  fallback_calls=native['Calls']-native['HostCalls'], host_instructions=native['HostInstructions'],
                  virtual_steps=native['Digest']['Steps'], all_architectural_state_equal=True,
                  digest={k: bytes(v).hex() if isinstance(v, list) else v for k, v in native['Digest'].items()},
                  video_writes=native['VideoCount'], video_sha256=native['VideoSHA256'],
                  retraces=native['FrameCount'], retraces_sha256=native['FramesSHA256'],
                  audio_events=native['AudioCount'], audio_sha256=native['AudioSHA256'],
                  dos_requests=native['DOSCount'], dos_sha256=native['DOSSHA256'],
                  guest_files=len(tree(folders[0]/'save')), witnesses=witnesses,
                  source_reports={p.name:sha((p/'calls.json').read_bytes()) for p in folders},
                  trace_tool_sha256=sha((a.project/'tools/draw_call_probe.go').read_bytes()),
                  trace_binary_sha256=sha((a.project/'workplace/dosgolem/workplace/bin/draw-call-probe-whole-clean').read_bytes()),
                  check_tool_sha256=sha(Path(__file__).read_bytes()),
                  benchmark=json.loads((root/'whole-call-cost-v2.json').read_text()),
                  benchmark_tool_sha256=sha((a.project/'tools/whole_call_bench.go').read_bytes()),
                  benchmark_binary_sha256=sha((a.project/'workplace/dosgolem/workplace/bin/whole-call-bench-clean').read_bytes()),
                  trace_timing_note='兩個追蹤同時執行，含CPU profiling及完整串流核對，不用其wall time作效能結論。')
    if a.gui:
        demo = a.project/'workplace/probe/g9-hd/whole-play'
        summary = json.loads((demo/'bigpic-summary.json').read_text())
        assert summary['prior_original_state_equal'] and summary['prior_sample_composite_equal']
        assert summary['binary_sha256'] == result['binary_sha256']
        result['normal_demo'] = summary
        result['normal_demo_draws'] = {}
        for mode in ('hd', 'base', 'english', 'twox'):
            d = json.loads((demo/mode/'draw-report.json').read_text())
            assert d['HostCalls'] > 0 and d['HostInstructions'] > 0 and d['Mismatches'] == 0
            result['normal_demo_draws'][mode] = {k:d[k] for k in ('HostCalls', 'HostInstructions', 'Mismatches')}
        current = a.project/'workplace/probe/g9-hd/whole-ui/wall-direct'
        old = a.project/'workplace/probe/g9-hd/direct-ui/wall-direct'
        assert endpoint(current) == endpoint(old)
        cur_input, old_input = [json.loads((p/'inputs.json').read_text()) for p in (current, old)]
        for key in ('orig', 'font', 'text', 'theme'):
            assert cur_input[key] == old_input[key]
        front = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt',
                 'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
        guests = [{k:v for k,v in tree(p/'save').items() if k not in front} for p in (current, old)]
        assert guests[0] == guests[1] and len(guests[0]) == 54
        for name in ('source.idx', 'source.pal', 'original.rgba', 'frame.png'):
            assert (current/name).read_bytes() == (old/name).read_bytes(), name
        for name in ('shots', 'window-shots', 'sources'):
            assert tree(current/name) == tree(old/name) and len(tree(current/name)) > 0, name
        assert len(tree(current/'shots')) == 10
        assert re.search(r'pod-play: 閃現 0 次', (current/'run.log').read_text())
        draw = json.loads((current/'draw-report.json').read_text())
        assert draw['HostCalls'] > 0 and draw['Mismatches'] == 0
        result['normal_load_ui'] = dict(frames=4800, screenshot_pairs=10, all_screenshots_equal=True,
                                       guest_files=54, guest_files_equal=True, flash_count=0,
                                       host_calls=draw['HostCalls'], host_instructions=draw['HostInstructions'],
                                       frame_sha256=sha((current/'frame.png').read_bytes()))
    if a.save:
        save_root = a.project/'workplace/probe/g9-hd/whole-save'
        old, new = [save_root/name for name in ('legacy', 'native')]
        inputs = [json.loads((p/'before.json').read_text()) for p in (old, new)]
        old_input, new_input = [dict(v) for v in inputs]
        old_input.pop('binary_sha256')
        new_input.pop('binary_sha256')
        assert old_input == new_input and 'SAVE/savgamB.pty' not in old_input
        saved = [json.loads((p/'after-save.json').read_text()) for p in (old, new)]
        assert saved[0] == saved[1] and saved[1]['savgamB.pty']
        assert saved[1]['savgamA.pty'] == old_input['SAVE/savgamA.pty']
        original_slot = {k[5:]:v for k,v in old_input.items() if k.startswith('SAVE/')}
        assert all(saved[1].get(name) == value for name, value in original_slot.items())
        final_save = [tree(p/'game/SAVE') for p in (old, new)]
        assert final_save[0] == final_save[1]
        assert final_save[1]['savgamB.pty'] == saved[1]['savgamB.pty']
        roundtrips = []
        for phase, frames in (('save', 4700), ('load', 3250)):
            logs = [(p/f'{phase}.log').read_text() for p in (old, new)]
            ends = []
            for log in logs:
                first = re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)', log)
                second = re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)', log)
                assert first and second and int(first[1]) == frames and int(first[2]) == frames*96250
                assert '截圖失敗' not in log and '字型缺字' not in log
                ends.append(first.groups()+second.groups())
            assert ends[0] == ends[1], phase
            for name in (f'{phase}-final.rgba', f'{phase}-source.idx', f'{phase}-source.pal'):
                assert (old/name).read_bytes() == (new/name).read_bytes(), name
            assert tree(old/f'{phase}-shots') == tree(new/f'{phase}-shots')
            assert len(tree(new/f'{phase}-shots')) == (5 if phase == 'save' else 1)
            draw = json.loads((new/f'{phase}-draw.json').read_text())
            assert draw['HostCalls'] > 0 and draw['Mismatches'] == 0
            roundtrips.append(dict(phase=phase, frames=frames, endpoint=ends[1],
                                   host_calls=draw['HostCalls'], host_instructions=draw['HostInstructions']))
        result['normal_save_roundtrip'] = dict(scope='正常冷啟動讀A、紮營另存原先不存在的B，再冷啟動讀B至隊伍選單；不注入機器狀態',
                                              slot_b_created=True, slot_b_loaded=True, slot_a_unchanged=True,
                                              slot_a_unchanged_files=len(original_slot),
                                              all_save_bytes_equal=True, screenshot_pairs=6,
                                              slot_b_sha256=saved[1]['savgamB.pty'],
                                              save_files_after_write=len(saved[1]), save_files_after_load=len(final_save[1]),
                                              phases=roundtrips)
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps({k: result[k] for k in ('host_calls', 'host_instructions', 'video_writes', 'retraces', 'dos_requests')}))


if __name__ == '__main__':
    main()
