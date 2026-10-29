#!/usr/bin/env python3
"""驗證繪圖收尾的實際像素來源、原版狀態、正常存讀檔與來源邊界。"""
import argparse
import hashlib
import json
import re
import subprocess
from pathlib import Path


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def read(p):
    return json.loads(p.read_text())


def tree(p):
    assert p.is_dir(), p
    return {str(f.relative_to(p)): sha(f) for f in sorted(p.rglob('*')) if f.is_file()}


def state(p):
    text = p.read_text()
    result = {key: re.search(key + r'_sha256=([0-9a-f]{64})', text)[1] for key in ['cpu', 'memory', 'indexed', 'palette']}
    match = re.search(r'frames=(\d+) steps=(\d+)', text)
    result.update(frames=int(match[1]), steps=int(match[2]))
    return result


def same_files(a, b):
    first, second = tree(a), tree(b)
    assert first == second, (a, b, sorted(k for k in set(first) | set(second) if first.get(k) != second.get(k))[:10])


def check_coverage(path, originals):
    data = read(path)
    assert data['LinkErrors'] == 0, path
    assert data['Bytes'].get('typed_go', 0) > 0, path
    assert {k: v for k, v in data['Bytes'].items() if v and k != 'typed_go'} == {}, path
    assert {k: v for k, v in data['Commands'].items() if v and k != 'typed_go'} == {}, path
    assert sum(data['Writers']['typed_go'].values()) == data['Bytes']['typed_go']
    for key, h in data['CodeSHA'].items():
        assert originals[key] == h, (path, key)
    return {'bytes': data['Bytes']['typed_go'], 'commands': data['Commands']['typed_go'],
            'fixed_pixel_bytes': 0, 'x86_string_pixel_bytes': 0, 'x86_pixel_bytes': 0,
            'plan_link_errors': 0, 'observed_entries': sorted(data['Visits']),
            'runtime_code_sha256': data['CodeSHA'], 'coverage_sha256': sha(path)}


def verify(project, tag):
    module = project / 'workplace/dosgolem'
    raw = project / 'workplace/probe/g9-hd' / tag
    before = project / 'workplace/probe/g9-hd/draw-closure-v1'
    saves = project / 'workplace/probe/g9-hd' / (tag + '-save') / 'native'
    save_before = project / 'workplace/probe/g9-hd/draw-closure-baseline-save-v1/native'
    inputs = project / 'workplace/input/dos/pod'
    driver = next(p for p in inputs.iterdir() if p.name.upper() == 'VGA.DRV')
    assert sha(driver) == 'eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f'
    binary = driver.read_bytes()
    assert binary[0x55cd:0x55cf] == b'\xf3\xae'
    assert binary[0x5602:0x5604] == b'\xf3\xae'
    assert binary[0x2cf4:0x2cf7] == b'\x26\x8a\x07'
    assert hashlib.sha256(binary[0x5593:0x5615]).hexdigest() == '776aafc34711e77e5120c2d65532dd361555f8de46855ba58e1862896bf7a8d1'
    immutable = read(project / 'workplace/probe/g9-hd/draw-closure-final-v1/wall-direct/inputs.json')
    immutable = {k: immutable[k] for k in ['orig', 'font', 'text', 'theme']}
    for key, path in [('orig', inputs), ('font', project / 'workplace/font'),
                      ('text', project / 'text'), ('theme', project / 'workplace/hd/theme')]:
        assert tree(path) == immutable[key], ('changed input', key)
    archives = {'pools-of-darkness.zip': 'a67022ffd89128cc1e35b6fefa79e456fd94ee03ea5ec26c0082f4bebc770483',
                '珍157-黑暗之池.rar': '3d18ec1b0d588705484f53f16921ed90af9343f3246af8b9f00f99677adcf623'}
    for name, digest in archives.items():
        assert sha(project / name) == digest
    table = read(module / 'apps/pod/drawblocks/blocks.json')
    originals = {}
    witnesses = project / 'workplace/probe/draw-intercept-v1/native-fade-v1/trace-original-entries'
    for cs, lo, hi, _ in table['functions']:
        if cs == 0x48a1:
            code = binary[lo:hi]
        else:
            memory = (witnesses / f'draw-{cs:04x}-{lo:04x}.mem').read_bytes()
            code = memory[cs * 16 + lo:cs * 16 + hi]
        originals[f'{cs:04x}:{lo:04x}'] = hashlib.sha256(code).hexdigest()
    reports = {}
    for name in ['title-direct', 'movie-direct', 'wall-direct', 'hd']:
        path = raw / name
        reports[name] = check_coverage(path / 'coverage.json', originals)
        reports[name]['state'] = state(path / 'run.log')
        text = (path / 'run.log').read_text()
        assert 'FAIL' not in text and '錯誤 <nil>' in text
        captured = read(path / 'inputs.json')
        for key, hashes in immutable.items():
            assert captured[key] == hashes, (name, key)
        assert captured['binary_sha256'] == sha(module / 'workplace/bin/pod-play-drawclosure')
    for name in ['title-direct', 'movie-direct']:
        assert state(raw / name / 'run.log') == state(before / name / 'run.log')
        for folder in ['shots', 'window-shots', 'sources', 'save']:
            same_files(raw / name / folder, before / name / folder)
        for filename in ['frame.png', 'original.rgba']:
            assert sha(raw / name / filename) == sha(before / name / filename)
        reports[name]['before_state_files_and_png_equal'] = True
    plain = project / 'workplace/probe/g9-hd' / (tag + '-plain') / 'title-direct'
    assert state(plain / 'run.log') == reports['title-direct']['state']
    for folder in ['shots', 'window-shots', 'sources', 'save']:
        same_files(plain / folder, raw / 'title-direct' / folder)
    for filename in ['frame.png', 'original.rgba']:
        assert sha(plain / filename) == sha(raw / 'title-direct' / filename)
    assert read(plain / 'inputs.json')['binary_sha256'] == sha(module / 'workplace/bin/pod-play-drawfinal')
    reports['title-direct']['uninstrumented_candidate_state_files_and_png_equal'] = True
    wall_prior = project / 'workplace/probe/interactive-speed-v5'
    checkpoint = read(wall_prior / 'checkpoint-4800.json')
    now = dict(reports['wall-direct']['state'])
    now.pop('frames')
    assert now == checkpoint
    prefix = [3900, 4200, 4320, 4360, 4420, 4520, 4620, 4720, 4750, 4780]
    for frame in prefix:
        filename = f'pod-f{frame:06d}'
        for folder, suffixes in [('shots', ['.png']), ('sources', ['.idx', '.pal', '.glyphs.json'])]:
            for suffix in suffixes:
                assert sha(raw / 'wall-direct' / folder / (filename + suffix)) == sha(wall_prior / folder / (filename + suffix))
    reports['wall-direct']['same_state_and_central_png_equal'] = True
    hd_prior = project / 'workplace/probe/g9-hd/nativedrawingjobs-play-v1/hd'
    assert reports['hd']['state'] == state(hd_prior / 'run.log')
    for folder in ['shots', 'window-shots', 'sources', 'save']:
        same_files(raw / 'hd' / folder, hd_prior / folder)
    for filename in ['frame.png', 'original.rgba']:
        assert sha(raw / 'hd' / filename) == sha(hd_prior / filename)
    reports['hd']['before_state_files_and_png_equal'] = True
    save_reports = {}
    for phase in ['save', 'load']:
        save_reports[phase] = check_coverage(saves / (phase + '-draw.json.coverage.json'), originals)
        save_reports[phase]['state'] = state(saves / (phase + '.log'))
        assert save_reports[phase]['state'] == state(save_before / (phase + '.log'))
        same_files(saves / (phase + '-shots'), save_before / (phase + '-shots'))
        for suffix in ['final.rgba', 'source.idx', 'source.pal', 'source.glyphs.json']:
            assert sha(saves / (phase + '-' + suffix)) == sha(save_before / (phase + '-' + suffix))
    same_files(saves / 'game/SAVE', save_before / 'game/SAVE')
    assert (saves / 'game/SAVE/savgamB.pty').is_file()
    initial = read(saves / 'before.json')
    assert 'SAVE/savgamB.pty' not in initial
    assert sha(saves / 'game/SAVE/savgamA.pty') == initial['SAVE/savgamA.pty']
    # 保留硬體查詢與非繪圖相容流程；未觀測入口不自動變成新的逆向待辦。
    observed = sorted(set(k for r in [*reports.values(), *save_reports.values()] for k in r['observed_entries']))
    native_files = [p for p in (module / 'apps/pod/drawblocks').glob('native_*.go') if not p.name.endswith('_test.go')]
    native_files += [module / 'apps/pod/drawblocks/transfer_calls.go', module / 'apps/pod/glyph_call.go']
    for p in native_files:
        text = p.read_text()
        assert all(token not in text for token in ['.Step(', 'blocks[', 'b.run(', 'DecodeOpcode']), p
    logs = {name: module / 'workplace' / ('draw-closure-' + name + '.log') for name in ['guard-targeted', 'regression', 'related', 'frontend', 'vet']}
    source_time = int(subprocess.check_output(['git', 'log', '-1', '--format=%ct'], cwd=module, text=True).strip())
    for name, path in logs.items():
        if name != 'guard-targeted':
            assert path.stat().st_mtime >= source_time, ('stale validation log', path)
        text = path.read_text()
        assert 'FAIL' not in text and 'SKIP' not in text, path
        if name == 'vet':
            assert not text
        else:
            assert 'ok  ' in text
    assert 'TestFixedCharacterRecordFramebufferAlias' in logs['guard-targeted'].read_text()
    targeted = logs['guard-targeted'].read_text()
    for proof in ['original row scan comparisons 720', 'scan refusals 30',
                  'whole flood calls 144 Go scans 882', 'point boundaries 48 refusals 288']:
        assert proof in targeted, proof
    effects = (module / 'apps/pod/drawblocks/effects.go').read_text()
    read_guard = effects.split('func (e *effect) read(', 1)[1].split('func protected(', 1)[0]
    assert 'if a >= 0xa0000 {' in read_guard and 'e.bad = true' in read_guard
    assert 'fixedFloodScan(e.s.IP, b)' in effects
    scanner = (module / 'apps/pod/drawblocks/native_flood_scan.go').read_text()
    for token in ['0x55cd', '0x5602', '0x2cf4', 'Algorithmic: true']:
        assert token in scanner, token
    build = (module / 'workplace/draw-closure-build.txt').read_text()
    commit = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=module, text=True).strip()
    assert 'vcs.revision=' + commit in build and 'vcs.modified=false' in build and 'go1.24.13' in build
    assert not subprocess.check_output(['git', 'status', '--porcelain'], cwd=module, text=True)
    return {'status': 'PASS', 'program_commit': commit,
            'binary_sha256': sha(module / 'workplace/bin/pod-play-drawfinal'),
            'audit_binary_sha256': sha(module / 'workplace/bin/pod-play-drawclosure'),
            'activation': {'official_matches_candidate': sha(module / 'workplace/bin/pod-play') == sha(module / 'workplace/bin/pod-play-drawfinal'),
                           'official_sha256': sha(module / 'workplace/bin/pod-play'),
                           'previous_backup': 'pod-play-renderfeedback',
                           'previous_backup_sha256': sha(module / 'workplace/bin/pod-play-renderfeedback')},
            'scope': '本作VGA mode13h已支援繪圖核心；標題、片頭、讀檔介面、DEMO與正常另存B／冷讀B。未知來源及未支援條件保留原版安全回退，不宣稱整個DOS程式改寫成Go。',
            'routes': reports, 'normal_save_load': save_reports,
            'total_typed_go_pixel_bytes': sum(x['bytes'] for x in [*reports.values(), *save_reports.values()]),
            'fixed_pixel_bytes': 0, 'x86_string_pixel_bytes': 0, 'x86_pixel_bytes': 0,
            'source_boundary': {'ordinary_input_and_text_control': '保留原版內容選擇、資料準備及文字組句；像素運算交由Go核心', 'platform': 'BIOS／DOS／硬體服務保留', 'unknown': '安全回退並保留未驗範圍'},
            'raster_read_evidence': {
                'scope': '原版sub_5414完整填色與sub_5593連續區段搜尋；正常六路未遇到此入口，另做受控原版比對',
                'search_sites': ['48a1:55cd', '48a1:5602'],
                'point_read_site': '48a1:2cf4',
                'original_search_comparisons': 720,
                'original_point_read_boundaries': 48,
                'search_refusals': 30, 'point_read_refusals': 288,
                'whole_flood_calls': 144, 'whole_flood_go_searches': 882,
                'budgets': [1, 13, 1024], 'root_widths': [32, 160, 320],
                'cpu_pixel_read_fallbacks_in_controlled_calls': 0,
                'full_state_and_port_log_equal': True,
                'fixed_framebuffer_reads_and_writes_guarded': True,
                'prototype_log_sha256': sha(module / 'workplace/flood-scan-prototype.log') if (module / 'workplace/flood-scan-prototype.log').is_file() else None,
                'original_span_code_sha256': '776aafc34711e77e5120c2d65532dd361555f8de46855ba58e1862896bf7a8d1'},
            'observed_entries': observed, 'registered_functions': len(table['functions']),
            'vga_sha256': sha(driver), 'blocks_table_sha256': sha(module / 'apps/pod/drawblocks/blocks.json'),
            'unchanged_inputs_sha256': immutable, 'original_archives_sha256': archives,
            'native_source_sha256': {str(p.relative_to(module)): sha(p) for p in native_files},
            'guard_source_sha256': {str(p.relative_to(module)): sha(p) for p in [module / 'apps/pod/drawblocks/effects.go', module / 'apps/pod/drawblocks/fixed_framebuffer_guard.go', module / 'apps/pod/drawblocks/fixed_framebuffer_guard_test.go']},
            'logs_sha256': {name: sha(p) for name, p in logs.items()},
            'build_sha256': sha(module / 'workplace/draw-closure-build.txt'),
            'prior_receipts': {name: sha(project / 'docs/re' / name) for name in ['019-native-drawing-jobs-runtime.json', '022-speed-feedback-hd-runtime.json']},
            'evidence_root': str(raw).replace('/project/', ''), 'route_evidence_sha256': tree(raw),
            'save_evidence_sha256': tree(saves),
            'plain_candidate_evidence_sha256': tree(plain),
            'tools_sha256': {name: sha(project / 'tools' / name) for name in ['draw-closure-receipt.sh', 'draw_closure_probe.py', 'draw_closure_check.py']}}


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--project', type=Path, default=Path('/project'))
    parser.add_argument('--tag', required=True)
    parser.add_argument('--out', type=Path, required=True)
    parser.add_argument('--official', action='store_true', help='要求正式執行檔已與驗收候選相同；不執行切換')
    args = parser.parse_args()
    receipt = verify(args.project, args.tag)
    if args.official:
        assert receipt['activation']['official_matches_candidate']
    args.out.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps({'status': 'PASS', 'typed_go_bytes': receipt['total_typed_go_pixel_bytes'], 'x86_pixel_bytes': 0, 'sha256': sha(args.out)}))
