#!/usr/bin/env python3
"""規格019／020的正常冷開機、呈現隔離及固定倍率收據。於sosb-img:r1執行。"""
import argparse
import hashlib
import json
import re
import struct
from pathlib import Path

from PIL import Image, ImageChops


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def files(root):
    return {str(p.relative_to(root)): sha(p) for p in root.rglob('*') if p.is_file()}


def state(root):
    log = (root / 'run.log').read_text()
    out = {}
    for name in ['memory', 'indexed', 'cpu', 'palette']:
        m = re.search(name + r'_sha256=([0-9a-f]{64})', log)
        assert m, name
        out[name] = m[1]
    m = re.search(r'frames=(\d+) steps=(\d+) memory_sha256', log)
    assert m
    out['frames'], out['steps'] = map(int, m.groups())
    return out


def observations(root):
    return [json.loads(s) for s in (root / 'observations.jsonl').read_text().splitlines()]


def compare(a, b, name):
    assert a == b, name


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('root', type=Path)
    ap.add_argument('--project', type=Path, default=Path('/project'))
    ap.add_argument('--out', type=Path, required=True)
    args = ap.parse_args()
    r = args.root
    states = {mode: state(r / mode) for mode in ['fonts', 'baseline', 'delay-off', 'official', 'speeds']}
    assert len(files(r / 'baseline/sources')) == 21
    assert sha(args.project / 'workplace/dosgolem/workplace/bin/pod-play') == 'fb5394a8cc5985ff466e4e7c7c728756cfb98d3b5edd93670a3f8986100d69d1'
    for mode in ['fonts', 'delay-off', 'official']:
        compare(states[mode], states['baseline'], mode + ' guest state')
        compare(files(r / mode / 'sources'), files(r / 'baseline/sources'), mode + ' source snapshots')
        if mode == 'fonts':
            # finish()的-original.rgba仍會疊工具列；原版底圖另以sources逐byte核對。
            a, b = ((r / m / 'original.rgba').read_bytes() for m in ['fonts', 'baseline'])
            assert len(a) == len(b) == 960 * 600 * 4
            compare(a[36 * 960 * 4:], b[36 * 960 * 4:], 'final original composition outside toolbar')
        else:
            compare(sha(r / mode / 'original.rgba'), sha(r / 'baseline/original.rgba'), mode + ' original frame')
        guest = {k: v for k, v in files(r / mode / 'save').items() if not k.startswith('pod-')}
        base_guest = {k: v for k, v in files(r / 'baseline/save').items() if not k.startswith('pod-')}
        compare(guest, base_guest, mode + ' DOS files')
        compare(files(r / mode / 'shots'), files(r / 'baseline/shots'), mode + ' game pixels without toolbar')
    for mode in ['delay-off', 'official']:
        compare(files(r / mode / 'window-shots'), files(r / 'baseline/window-shots'), mode + ' complete windows')

    samples = [x for x in observations(r / 'fonts') if x['event'] == 'shot']
    wanted = [(3, 0, 16), (3, 1, 22), (3, 2, 24), (2, 0, 12), (2, 1, 16), (2, 2, 16)]
    compare([(x['scale'], x['size'], x['glyph_px']) for x in samples[:6]], wanted, 'six font sizes')
    diffs = []
    for x in samples:
        assert x['font_width'] == x['font_height'] == x['glyph_px']
        if x['glyph_px'] in (12, 22):
            assert x['font_name'] == f"pod-rounded-bar{x['glyph_px']}.golemfnt"
        name = f"pod-f{x['frame']:06d}.png"
        a, b = (Image.open(r / mode / 'window-shots' / name).convert('RGB') for mode in ['fonts', 'baseline'])
        assert a.size == b.size
        diff = ImageChops.difference(a, b)
        box = diff.getbbox()
        scale = x['scale']
        if box:
            assert box[0] >= 80 * scale and box[2] <= 400 * scale and box[3] <= 48 * scale, (name, box)
        diffs.append({'frame': x['frame'], 'scale': scale, 'size': x['size'], 'bbox': box})
    assert sum(x['bbox'] is not None for x in diffs) >= 4
    compare((r / 'fonts/save/pod-barfont.txt').read_text(), '2\n', 'last size saved')

    speed = [x for x in observations(r / 'speeds') if x['event'] == 'shot']
    compare([x['speed'] for x in speed], [1, 2, 4, 8, 16, 32], 'speed samples')
    for x in speed:
        assert x['delay_ready'] and x['clock'] == 50 * x['speed']
        assert x['budget'] == 96250 * x['speed'], x
        assert x['actual'] == str(x['speed']) + 'x'
    transitions = [(1010, 1), (1015, 2), (1020, 4), (1025, 8), (1030, 16), (1035, 32)]
    for a, b in zip(speed, speed[1:]):
        want = sum(96250 * next(m for f, m in reversed(transitions) if f <= t) for t in range(a['frame'], b['frame']))
        compare(b['steps'] - a['steps'], want, 'actual virtual step budget')

    contract_log = (r / 'contracts.log').read_text()
    assert '--- PASS: TestAcceptanceNativeToolbarAndPreferenceClicks' in contract_log
    assert '\nFAIL' not in contract_log and '\nPASS\n' in contract_log
    pref_log = (r / 'contracts-preferences.log').read_text()
    assert '--- PASS: TestAcceptanceNativeToolbarAndPreferenceClicks' in pref_log and '\nFAIL' not in pref_log
    restart = [json.loads(s) for s in (r / 'restart.jsonl').read_text().splitlines()]
    assert len(restart) == 1 and restart[0]['event'] == 'startup'
    assert restart[0]['size'] == 2 and restart[0]['glyph_px'] == 24 and restart[0]['speed'] == 32
    assert restart[0]['actual'] == '1x' and not restart[0]['delay_ready']
    fonts = {}
    for name in ['pod-rounded.golemfnt', 'pod-rounded-3x.golemfnt', 'pod-rounded-bar12.golemfnt', 'pod-rounded-bar22.golemfnt']:
        p = args.project / 'workplace/font' / name
        raw = p.read_bytes()
        assert raw[:8] == b'GOLEMFNT'
        w, h, count = struct.unpack_from('<HHI', raw, 8)
        fonts[name] = {'sha256': sha(p), 'width': w, 'height': h, 'characters': count}
    compare(fonts['pod-rounded.golemfnt']['sha256'], '0f0430e7db828b8b74249b69f02ac7c590724215d9edc0f212830790ae4531ac', '16px unchanged')
    compare(fonts['pod-rounded-3x.golemfnt']['sha256'], '52e749e344c5bf6a60fa6501ad4ab17b72de0871897382a5c58d2acc3cc21730', '24px unchanged')
    acceleration = {}
    performance = {}
    for mode in states:
        log = (r / mode / 'run.log').read_text()
        acceleration[mode] = [s for s in log.splitlines() if 'TP Delay' in s or 'accelerated' in s or '快轉' in s]
        performance[mode] = [s for s in log.splitlines() if 'pod-play: perf ' in s]
    assert any('登記 true、快轉 67740 步' in s for s in acceleration['baseline']), acceleration
    assert any('登記 false、快轉 0 步' in s for s in acceleration['delay-off']), acceleration
    result = {'status': 'PASS', 'scope': '正常冷開機至開場標題；2/3倍三級字級、英文抽樣、六種固定倍率；點擊處理與偏好另以契約測試驗證，不宣稱OS滑鼠自動化、互動30fps或全遊戲覆蓋',
              'program': {'commit': '18927120914774783278c52db9ed50360b0d4ea7', 'sha256': sha(args.project / 'workplace/dosgolem/workplace/bin/pod-play')},
              'evidence_root': str(r).replace('/project/', ''), 'states': states, 'font_samples': samples, 'font_differences': diffs,
              'speed_samples': speed, 'fonts': fonts, 'delay_logs': acceleration, 'performance': performance,
              'preference_restart': restart[0],
              'font_inputs': {p: sha(args.project / p) for p in ['workplace/fonts-src/rounded/GenSenRounded2TW-M.otf', 'workplace/fonts-src/rounded/GenSenRounded2TW-B.otf', 'text/host-text.zh-TW.tsv', 'text/host-text.en.tsv']},
              'readme_preview_sha256': sha(args.project / 'docs/screenshots/toolbar-font-sizes.png'),
              'evidence_sha256': files(r),
              'tools': {p: sha(args.project / 'tools' / p) for p in ['toolbar-acceptance-receipt.sh', 'toolbar_receipt_probe.py', 'toolbar_acceptance_test.go', 'toolbar_acceptance_check.py', 'toolbar_font_preview.py', 'font-dev.sh', 'rounded_font.py', 'eten_font.py']}}
    args.out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps({'status': 'PASS', 'font_samples': len(samples), 'speed_samples': len(speed), 'state_comparisons': 3, 'receipt_sha256': sha(args.out)}, ensure_ascii=False))


if __name__ == '__main__':
    main()
