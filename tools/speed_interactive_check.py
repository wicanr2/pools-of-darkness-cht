#!/usr/bin/env python3
"""檢查正常互動路線、Draw計數、六種倍率與既有固定1x的狀態隔離。"""
import argparse
import bisect
import hashlib
import json
import re
import statistics
import struct
from pathlib import Path


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def tree(p):
    return {str(f.relative_to(p)): sha(f) for f in p.rglob('*') if f.is_file()}


def state(p):
    s = p.read_text()
    result = {k: re.search(k + r'_sha256=([0-9a-f]{64})', s)[1] for k in ['cpu', 'memory', 'indexed', 'palette']}
    result['steps'] = int(re.search(r'frames=\d+ steps=(\d+)', s)[1])
    return result


def percentile(values, fraction):
    values = sorted(values)
    return values[min(len(values) - 1, int(len(values) * fraction))]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('root', type=Path)
    ap.add_argument('--project', type=Path, default=Path('/project'))
    ap.add_argument('--source-commit', required=True)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--require-render-feedback', action='store_true')
    a = ap.parse_args()
    p, r = a.project, a.root
    data = json.loads((r / 'timing.json').read_text())
    rows, draws = data['updates'], data['draws']
    assert data['interactive'] and len(rows) == data['frames'] and data['frames'] >= 6200
    assert [x['frame'] for x in rows] == list(range(1, len(rows) + 1))
    assert draws and all(x['total_ns'] > 0 for x in draws)
    assert json.loads((r / 'close.json').read_text())['event'] == 'WM_DELETE_WINDOW'
    # Draw回呼計數須與修正後的正式perf計數相同，工具列隱藏的初期也要計數。
    runlog = (r / 'run.log').read_text()
    m = re.search(r'perf draw n=(\d+)', runlog)
    assert m and int(m[1]) == len(draws)
    assert any(x['frame'] < 4700 for x in draws)
    draw_starts = [x['start_ns'] for x in draws]
    draw_averages = [0]
    for i in range(1, len(draws)):
        gap = draw_starts[i] - draw_starts[i - 1]
        previous = draw_averages[-1]
        draw_averages.append(gap if previous == 0 else previous + int((gap - previous) / 8))
    slow_render_increases = []
    for row in rows:
        i = bisect.bisect_right(draw_starts, row['start_ns']) - 1
        interval = draw_averages[i] if i >= 0 else 0
        if row['clock_next'] > row['clock_used'] and interval > 1e9 / 30 + 100000:
            slow_render_increases.append({'frame': row['frame'], 'interval_ns': interval})
        if a.require_render_feedback:
            assert 'draw_interval_ns' in row
            assert abs(row['draw_interval_ns'] - interval) < 100000, 'Draw間隔讀值與獨立時間戳不符'
            if row['clock_next'] > row['clock_used']:
                assert row['draw_interval_ns'] <= int(1e9 / 35)
    if a.require_render_feedback:
        assert not slow_render_increases, '低於30fps時仍提高時脈'
    prior_path = p / 'workplace/probe/g9-hd/nativedrawingjobs-font-v1/wall-direct'
    checkpoint = json.loads((r / 'checkpoint-4800.json').read_text())
    assert checkpoint == state(prior_path / 'run.log'), '正常讀檔4800格的原版狀態改變'
    assert checkpoint['steps'] == 462000000
    assert sum(x['steps'] for x in rows[:4800]) == 462000000
    for x in rows:
        assert 50 <= x['clock_used'] <= 50 * x['set']
        assert 50 <= x['clock_next'] <= 50 * x['set']
        assert x['budget'] == x['clock_used'] * 1925
        assert x['steps'] <= x['budget']
        if x['capped']:
            assert x['set'] > 1 and x['steps'] < x['budget']
        if x['set'] == 1:
            assert x['clock_used'] == x['clock_next'] == 50
    stages = [(4800, 1), (5000, 2), (5200, 4), (5400, 8), (5600, 16), (5800, 32), (6000, 1)]
    samples = []
    for frame, speed in stages:
        whole = [x for x in rows if frame <= x['frame'] < frame + 200]
        assert len(whole) == 200 and all(x['set'] == speed and x['ready'] for x in whole)
        steady = whole[50:]
        start, end = steady[0]['start_ns'], steady[-1]['start_ns'] + steady[-1]['total_ns']
        paint = [x for x in draws if start <= x['start_ns'] < end]
        assert paint, '倍率抽樣沒有Draw'
        elapsed = (end - start) / 1e9
        intervals = [b['start_ns'] - a['start_ns'] for a, b in zip(paint, paint[1:])]
        assert intervals
        samples.append({'set': speed, 'frames': [frame + 50, frame + 199], 'updates': len(steady), 'draws': len(paint),
                        'seconds': elapsed, 'draw_callback_fps': len(paint) / elapsed, 'update_fps': len(steady) / elapsed,
                        'clock_min': min(x['clock_used'] for x in steady), 'clock_max': max(x['clock_used'] for x in steady),
                        'clock_median': statistics.median(x['clock_used'] for x in steady),
                        'virtual_steps_per_second': sum(x['steps'] for x in steady) / elapsed,
                        'advance_median_ms': statistics.median(x['advance_ns'] for x in steady) / 1e6,
                        'advance_p90_ms': percentile([x['advance_ns'] for x in steady], .9) / 1e6,
                        'draw_interval_p90_ms': percentile(intervals, .9) / 1e6,
                        'capped_updates': sum(x['capped'] for x in whole),
                        'clock_increases': sum(x['clock_next'] > x['clock_used'] for x in whole),
                        'clock_decreases': sum(x['clock_next'] < x['clock_used'] for x in whole)})
    increases = sum(x['clock_increases'] for x in samples)
    decreases = sum(x['clock_decreases'] for x in samples)
    transitions_observed = increases > 0 and decreases > 0
    # 診斷修正與舊正式程式：固定1100格、同輸入、同資產。
    old = p / 'workplace/probe/toolbar/toolbar-v2/baseline'
    fixed = r / 'fixed'
    assert state(fixed / 'run.log') == state(old / 'run.log')
    for folder in ['shots', 'window-shots', 'sources']:
        assert tree(fixed / folder) == tree(old / folder), folder
    for filename in ['original.rgba', 'frame.png', 'source.idx', 'source.pal', 'source.glyphs.json', 'script.txt']:
        assert sha(fixed / filename) == sha(old / filename), filename
    assert {k: v for k, v in tree(fixed / 'save').items() if not k.startswith('pod-')} == {k: v for k, v in tree(old / 'save').items() if not k.startswith('pod-')}
    # 正常讀檔至選單之前的原版來源及中央遊戲圖須相同，無頭／互動顯示頻率不影響來源。
    prefix = [3900, 4200, 4320, 4360, 4420, 4520, 4620, 4720, 4750, 4780]
    shape = struct.unpack('>II', (r / 'window-shots/pod-f003900.png').read_bytes()[16:24])
    scale = shape[1] // 200
    assert scale in [2, 3] and shape == (480 * scale, 200 * scale)
    for frame in prefix:
        name = f'pod-f{frame:06d}'
        if scale == 3:
            assert sha(r / 'shots' / (name + '.png')) == sha(prior_path / 'shots' / (name + '.png'))
        for suffix in ['.idx', '.pal'] + (['.glyphs.json'] if scale == 3 else []):
            assert sha(r / 'sources' / (name + suffix)) == sha(prior_path / 'sources' / (name + suffix))
    fps_met = all(s['draw_callback_fps'] >= 30 for s in samples)
    module = p / 'workplace/dosgolem'
    resource_start = json.loads((r / 'resources-start.json').read_text()) if (r / 'resources-start.json').exists() else None
    resource_end = json.loads((r / 'resources-end.json').read_text()) if (r / 'resources-end.json').exists() else None
    result = {'status': 'PASS' if fps_met and transitions_observed else 'FUNCTIONAL_PASS_WITH_LIMITS', 'source_commit': a.source_commit,
              'scope': 'Linux Docker、2GiB、私人Xvfb，音訊輸出關閉；正常互動冷開機讀檔、六種倍率與回到1x。Draw回呼頻率不等於實機螢幕或GPU呈現驗收。',
              'measurement': '每段200個Update，前50格作調節，後150格的完整牆鐘區間量測，包含F10截圖成本。既有繪圖頻率只計工具列，本批改計完整Draw。',
              'fps_30_met_in_all_samples': fps_met, 'adaptive_up_and_down_observed': transitions_observed, 'samples': samples,
              'scale': scale, 'render_feedback_required': a.require_render_feedback, 'slow_render_clock_increases': slow_render_increases,
              'resources_start': resource_start, 'resources_end': resource_end,
              'draw_count': len(draws), 'updates': len(rows), 'close': json.loads((r / 'close.json').read_text()),
              'checkpoint_4800': checkpoint, 'fixed_1100': state(fixed / 'run.log'),
              'fixed_png_comparisons': 15, 'interactive_prefix_png_comparisons': len(prefix) if scale == 3 else 0,
              'program_sha256': sha(r / 'pod-play-candidate'), 'observer_sha256': sha(r / 'pod-play-observed'),
              'program_sources': {name: sha(module / 'cmd/pod-play' / name) for name in ['main.go', 'perf_test.go', 'speed.go']},
              'prior_receipts': {name: sha(p / 'docs/re' / name) for name in ['021-toolbar-runtime.json', '019-native-drawing-jobs-runtime.json']},
              'tools': {name: sha(p / 'tools' / name) for name in ['speed-interactive-receipt.sh', 'speed_interactive_probe.py', 'speed_interactive_check.py']},
              'evidence_root': str(r).replace('/project/', ''), 'evidence_sha256': tree(r)}
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps({'status': result['status'], 'draws': len(draws), 'fps': [round(s['draw_callback_fps'], 1) for s in samples], 'receipt_sha256': sha(a.out)}, ensure_ascii=False))


if __name__ == '__main__':
    main()
