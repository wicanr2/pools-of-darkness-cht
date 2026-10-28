"""核對正常冷開機的 demo 中斷收據；不使用狀態注入。"""
import argparse
import hashlib
import json
from pathlib import Path
import re


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verify_resolution(path):
    text = path.read_text()
    key = 'DOS/GAME.EXE/35ad635e558852925b7028a921aacafc2df8e84be24a8a1136c6267471948d69:runtime16AF:0276'
    assert 'id="demo-initial"' in text and 'id="demo-conformed"' in text, '缺少新舊規格入口'
    old, current = text.split('<a id="demo-initial"></a>', 1)[1].split('<a id="demo-conformed"></a>', 1)
    assert 'DEMO-RESOLVED-20261028' in old and '(#demo-conformed)' in old, '缺少舊斷言勘誤與回鏈'
    assert key in current and 'A7CE' in current and '020-demo-control-runtime.json' in current, '缺少原始定位或證據'
    return dict(key=key, correction='DEMO-RESOLVED-20261028', spec_sha256=sha(path))


def summary(root, case):
    path = root / case / 'run.log'
    log = path.read_text()
    assert '缺字' not in log and '簽章命中 0' not in log, (case, '顯示或綁定失敗')
    result = re.search(r'demo 控制 逾時寫入擋下 (\d+) 次、中止 (\d+) 次、中止後返回 (\d+) 次（最近 (\d+) 步）、執行中 (true|false)、錯誤 <nil>', log)
    assert result, (case, '缺少控制結果')
    cleared, aborts, returns, steps = map(int, result.groups()[:4])
    assert result[5] == 'false' and cleared >= 1, (case, result.groups())
    count = 2 if case == 'combat-repeat' else (0 if case == 'menu' else 1)
    assert aborts == returns == count, (case, aborts, returns)
    if count:
        assert 0 < steps < 10_000_000, (case, '中止返回超出有界預算', steps)
    hashes = {key: re.search(key + r'=([0-9a-f]{64})', log)[1]
              for key in ('memory_sha256', 'cpu_sha256', 'indexed_sha256', 'palette_sha256')}
    return dict(timeouts_cleared=cleared, aborts=aborts, returns=returns,
                last_abort_steps=steps, hashes=hashes, log_sha256=sha(path),
                screenshots={p.name: sha(p) for p in sorted((root/case/'shots').glob('*.png'))})


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--reference', type=Path)
    ap.add_argument('--binary', type=Path, required=True)
    ap.add_argument('--receipt', type=Path, required=True)
    ap.add_argument('--spec', type=Path, required=True)
    args = ap.parse_args()
    assert not args.receipt.exists(), '保留既有收據，請使用新檔名'
    report = dict(schema=1, scope='normal cold boot, explicit D, host keyboard routing; no state injection',
                  binary_sha256=sha(args.binary), resolution_backlinks=verify_resolution(args.spec), cases={})
    for case in ('menu', 'story', 'combat-repeat'):
        row = summary(args.out, case)
        if args.reference:
            other = summary(args.reference, case)
            assert row['hashes'] == other['hashes'], (case, '整台原版狀態不同')
            assert row['screenshots'] == other['screenshots'], (case, '正常畫面不同')
        report['cases'][case] = row
    args.receipt.write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps(report, ensure_ascii=False))


if __name__ == '__main__':
    main()
