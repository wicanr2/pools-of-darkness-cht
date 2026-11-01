#!/usr/bin/env python3
"""在 Docker 以正式 pod-play 驗日韓首次語言、偏好重啟及拒絕分支。"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import signal
import subprocess
import sys
import time


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def require(condition, message):
    if not condition:
        raise ValueError(message)


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root', type=Path, required=True)
    ap.add_argument('--candidate', type=Path)
    ap.add_argument('--binary', type=Path)
    ap.add_argument('--original', type=Path)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--timeout', type=int, default=60)
    a = ap.parse_args()
    require(Path('/.dockerenv').is_file(), '本工具只能在 Docker 內執行')
    root = a.root.resolve()
    candidate = (a.candidate or root / 'workplace/multilingual-jako/candidate').resolve()
    binary = (a.binary or root / 'workplace/dosgolem/workplace/bin/pod-play').resolve()
    original = (a.original or root / 'workplace/input/dos/pod').resolve()
    out = a.out.resolve()
    engine = root / 'workplace/dosgolem'
    require(1 <= a.timeout <= 180, '--timeout 須在 1 至 180 秒')
    for path in (candidate / 'text', candidate / 'font', candidate / 'theme', original):
        require(path.is_dir(), '缺少輸入目錄：' + str(path))
    require(binary.is_file() and os.access(binary, os.X_OK), '缺少可執行 binary')
    require(shutil.which('xvfb-run') is not None, '缺少 xvfb-run')
    out.mkdir(parents=True, exist_ok=True)
    require(out.stat().st_uid == os.getuid(), '輸出目錄 UID 不符')
    for n in range(1, 10000):
        run_dir = out / ('run-%03d' % n)
        try:
            run_dir.mkdir()
            break
        except FileExistsError:
            continue
    else:
        raise ValueError('驗證輸出目錄已滿')

    sources = ['apps/pod/lang.go', 'cmd/pod-play/lang.go',
               'cmd/pod-play/lang_sys_other.go', 'cmd/pod-play/input.go',
               'cmd/pod-play/main.go', 'apps/pod/font.go']
    input_hashes = {}
    for directory in ('text', 'font', 'theme'):
        for path in sorted((candidate / directory).rglob('*')):
            if path.is_file():
                input_hashes[str(path.relative_to(candidate))] = sha(path)
    source_hashes = {name: sha(engine / name) for name in sources}
    original_hashes = {str(p.relative_to(original)): sha(p)
                       for p in sorted(original.rglob('*')) if p.is_file()}
    lang_source = (engine / sources[0]).read_text()
    match = re.search(r'func AvailableLangs\(dir string\).*?\n}', lang_source, re.S)
    require(match is not None, '找不到 pod.AvailableLangs')
    order_contract = match.group(0)
    require('sort.Slice(langs' in order_contract and
            'langs[i] == DefaultLang' in order_contract and
            'return langs[i] < langs[j]' in order_contract,
            'AvailableLangs 排序契約已變更，需重新審查')
    default = re.search(r'const DefaultLang = "([^"]+)"', lang_source).group(1)
    codes = [p.name[len('host-text.'):-len('.tsv')]
             for p in (candidate / 'text').glob('host-text.*.tsv')]
    codes = sorted((c for c in codes if c and '.' not in c),
                   key=lambda c: (c != default, c))
    require(codes == ['zh-TW', 'en', 'ja', 'ko', 'zh-CN'],
            '候選語系與已審查五語契約不同：' + repr(codes))
    for code in ('en', 'ja', 'ko', 'zh-CN'):
        require((candidate / 'font' / ('pod-' + code + '.golemfnt')).is_file(),
                '缺少語系字型：' + code)

    receipt = {
        'status': 'RUNNING', 'cases': [], 'run_directory': str(run_dir),
        'scope': {
            'platform': 'Linux 正式 pod-play，Xvfb 真實視窗',
            'positive_frames': 3, 'positive_script': '1:lang,2:bar',
            'negative_frames': 1,
            'observations': '實際前端 F8 後的 pod-lang.txt 與同格完整視窗 PNG',
            'isolation': '所有 save、負對照、XDG 快取、暫存與紀錄只在本次 run 目錄',
            'limits': ['三格開機只驗前端選語言與拒絕分支，不代替4800格正常讀檔路線',
                       '不驗Windows／macOS原生系統語言',
                       '未修改既有玩家存檔、原版輸入、候選文字／字型／主題'],
        },
        'binary': {'path': str(binary), 'sha256': sha(binary)},
        'tool': {'path': str(Path(__file__).resolve()), 'sha256': sha(__file__),
                 'python': sys.version.split()[0]},
        'source': {'engine': str(engine), 'files_sha256': source_hashes,
                   'available_langs_contract': order_contract},
        'available_langs': codes,
        'input_sha256': input_hashes, 'original_sha256': original_hashes,
        'commands': [], 'started_unix': time.time(),
    }

    def store():
        payload = json.dumps(receipt, ensure_ascii=False, indent=2) + '\n'
        (run_dir / 'receipt.json').write_text(payload)
        (out / 'receipt.json').write_text(payload)

    def execute(name, arguments, locale='en_US.UTF-8'):
        case_dir = run_dir / name
        case_dir.mkdir(parents=True, exist_ok=True)
        for child in ('tmp', 'cache', 'config'):
            (case_dir / child).mkdir(exist_ok=True)
        env = dict(os.environ, POD_NO_DIALOG='1', GOMAXPROCS='2',
                   LANGUAGE=locale, LC_ALL='C', LC_MESSAGES='C', LANG='C',
                   XDG_CACHE_HOME=str(case_dir / 'cache'),
                   XDG_CONFIG_HOME=str(case_dir / 'config'),
                   TMPDIR=str(case_dir / 'tmp'))
        command = ['xvfb-run', '-a', '-s', '-screen 0 1600x1200x24 -nolisten tcp',
                   str(binary)] + list(arguments)
        logfile = case_dir / 'frontend.log'
        started = time.monotonic()
        with logfile.open('wb') as log:
            process = subprocess.Popen(command, cwd=case_dir, env=env,
                                       stdout=log, stderr=subprocess.STDOUT,
                                       start_new_session=True)
            try:
                code = process.wait(timeout=a.timeout)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid, signal.SIGTERM)
                try:
                    process.wait(timeout=3)
                except subprocess.TimeoutExpired:
                    os.killpg(process.pid, signal.SIGKILL)
                    process.wait()
                raise ValueError('前端逾時：' + name)
        record = {'name': name, 'command': command, 'cwd': str(case_dir),
                  'environment': {k: env[k] for k in
                                  ('LANGUAGE', 'LC_ALL', 'LC_MESSAGES', 'LANG',
                                   'POD_NO_DIALOG', 'GOMAXPROCS')},
                  'exit_code': code, 'log': str(logfile),
                  'log_sha256': sha(logfile),
                  'elapsed_seconds': round(time.monotonic() - started, 3)}
        receipt['commands'].append(record)
        print(name, 'exit', code, flush=True)
        return record, logfile.read_text(errors='replace')

    def common(save, text=None):
        return ['-original', str(original), '-save', str(save),
                '-text-dir', str(text or candidate / 'text'),
                '-font', str(candidate / 'font/pod-rounded.golemfnt'),
                '-hd', str(candidate / 'theme'),
                '-posters', str(candidate / 'theme/posters'),
                '-scale', '2', '-audio=false']

    references = {}

    def positive(name, save, locale, expected, explicit=None, reference=True):
        save.parent.mkdir(parents=True, exist_ok=True)
        pref = save / 'pod-lang.txt'
        before = pref.read_text().strip() if pref.is_file() else None
        shot = run_dir / name / 'frame-003.png'
        arguments = common(save) + ['-frames', '3', '-script', '1:lang,2:bar',
                                    '-shot-frame', str(shot)]
        if explicit:
            arguments += ['-lang', explicit]
        process, log = execute(name, arguments, locale)
        require(process['exit_code'] == 0, name + ' 啟動失敗：' + log[-3000:])
        require(pref.is_file(), name + ' 沒有寫入語言偏好')
        actual = pref.read_text().strip()
        case = {'name': name, 'kind': 'frontend-language',
                'expected': expected, 'actual': actual,
                'locale': locale, 'explicit_lang': explicit,
                'preference_before': before, 'preference_path': str(pref),
                'preference_sha256': sha(pref), 'frames': 3,
                'script': '1:lang,2:bar', 'shot': str(shot),
                'shot_sha256': sha(shot), 'log_sha256': process['log_sha256'],
                'exit_code': process['exit_code']}
        if reference:
            baseline = references[expected]
            case.update(reference_shot=baseline['shot'],
                        reference_shot_sha256=baseline['shot_sha256'],
                        same_frame_pixels=case['shot_sha256'] == baseline['shot_sha256'])
        receipt['cases'].append(case)
        require(actual == expected, name + ' 選到 ' + actual + '，預期 ' + expected)
        if reference:
            require(case['same_frame_pixels'], name + ' 與同格顯式語系畫面不同')
        store()
        return case

    try:
        version, version_text = execute('version', ['-version'])
        require(version['exit_code'] == 0, '無法讀取binary版本')
        receipt['binary']['version_output'] = version_text.strip()
        git = subprocess.run(['git', '-C', str(engine), 'rev-parse', 'HEAD'],
                             capture_output=True, text=True, timeout=10)
        require(git.returncode == 0, '無法讀取引擎HEAD')
        receipt['source']['git_head'] = git.stdout.strip()
        for code in ('ja', 'ko'):
            name = code + '-missing-host'
            case_dir = run_dir / name
            text = case_dir / 'text'
            shutil.copytree(candidate / 'text', text)
            (text / ('host-text.' + code + '.tsv')).unlink()
            arguments = common(case_dir / 'game', text) + ['-lang', code, '-frames', '1']
            process, log = execute(name, arguments)
            error = '沒有這個語言或缺字型'
            rejected = process['exit_code'] != 0 and error in log
            receipt['cases'].append({'name': name, 'kind': 'negative', 'lang': code,
                                     'actual': 'rejected' if rejected else 'accepted-or-wrong-error',
                                     'expected': 'rejected', 'rejected': rejected,
                                     'expected_error': error, 'error_present': error in log,
                                     'exit_code': process['exit_code'],
                                     'log_sha256': process['log_sha256']})
            require(rejected, name + ' 未按預期拒絕：' + log[-3000:])
            name = code + '-bar12-as-3x'
            case_dir = run_dir / name
            font = case_dir / 'font' / ('pod-' + code + '.golemfnt')
            font.parent.mkdir(parents=True)
            shutil.copyfile(candidate / 'font' / font.name, font)
            scaled = font.with_name(font.stem + '-3x.golemfnt')
            bar12 = candidate / 'font' / ('pod-' + code + '-bar12.golemfnt')
            shutil.copyfile(bar12, scaled)
            arguments = common(case_dir / 'game') + [
                '-lang', code, '-frames', '1', '-lang-fonts', code + '=' + str(font)]
            process, log = execute(name, arguments)
            error = '24×24'
            rejected = process['exit_code'] != 0 and error in log
            receipt['cases'].append({'name': name, 'kind': 'negative', 'lang': code,
                                     'actual': 'rejected' if rejected else 'accepted-or-wrong-error',
                                     'expected': 'rejected', 'rejected': rejected,
                                     'expected_error': error, 'error_present': error in log,
                                     'exit_code': process['exit_code'],
                                     'injected_font_sha256': sha(scaled),
                                     'bar12_source_sha256': sha(bar12),
                                     'log_sha256': process['log_sha256']})
            require(rejected, name + ' 未按預期拒絕：' + log[-3000:])
            store()

        for i, code in enumerate(codes):
            expected = codes[(i + 1) % len(codes)]
            name = 'explicit-' + code
            references[expected] = positive(name, run_dir / name / 'game',
                                            'en_US.UTF-8', expected,
                                            explicit=code, reference=False)

        cycle_save = run_dir / 'cycle-state/game'
        observed_cycle = []
        for i in range(len(codes)):
            expected = codes[(i + 1) % len(codes)]
            case = positive('cycle-%d' % (i + 1), cycle_save, 'en_US.UTF-8',
                            expected, explicit=codes[0] if i == 0 else None)
            observed_cycle.append(case['actual'])
        receipt['observed_f8_cycle'] = observed_cycle

        for code, locale in [('ja', 'ja_JP.UTF-8'), ('ko', 'ko_KR.UTF-8')]:
            save = run_dir / ('system-' + code + '-state/game')
            require(not save.exists(), '系統選語言測試必須從乾淨save起跑')
            expected = codes[(codes.index(code) + 1) % len(codes)]
            first = positive('system-' + code + '-fresh', save, locale, expected)
            require(first['preference_before'] is None, '首次執行已有偏好')
            expected_restart = codes[(codes.index(expected) + 1) % len(codes)]
            restarted = positive('system-' + code + '-restart-en', save,
                                 'en_US.UTF-8', expected_restart)
            require(restarted['preference_before'] == expected, '重啟前偏好未保留')

        require(sha(binary) == receipt['binary']['sha256'], '驗證中binary被改動')
        for relative, digest in input_hashes.items():
            require(sha(candidate / relative) == digest, '候選輸入被改動：' + relative)
        for relative, digest in source_hashes.items():
            require(sha(engine / relative) == digest, '引擎來源被改動：' + relative)
        for relative, digest in original_hashes.items():
            require(sha(original / relative) == digest, '原版輸入被改動：' + relative)
        require(all(c['actual'] == c['expected'] for c in receipt['cases']), '案件未全數通過')
        receipt['status'] = 'PASS'
        receipt['summary'] = {'cases': len(receipt['cases']),
                              'negative_cases': 4, 'positive_cases': 14,
                              'same_frame_comparisons': 9,
                              'input_hashes_unchanged': True}
        receipt['finished_unix'] = time.time()
        store()
        print(json.dumps(receipt['summary'], ensure_ascii=False), flush=True)
    except Exception as error:
        receipt['status'] = 'FAIL'
        receipt['error'] = str(error)
        receipt['finished_unix'] = time.time()
        store()
        raise


if __name__ == '__main__':
    main()
