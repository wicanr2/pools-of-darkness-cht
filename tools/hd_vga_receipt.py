#!/usr/bin/env python3
"""容器內執行 PIC 首批同狀態 A/B、2 倍及來源失配負對照。掛載契約見研究 014。"""
import hashlib
import json
import shutil
import subprocess
from pathlib import Path

import numpy as np
from PIL import Image


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    out = Path('/out')
    bad = out/'mismatched-theme'
    shutil.copytree('/theme', bad, dirs_exist_ok=True)
    t = json.loads((bad/'theme.json').read_text())
    for it in t['items']:
        p = bad/it['indexed']
        p.write_bytes(bytes(v ^ 255 for v in p.read_bytes()))
        it['indexed_sha256'] = sha(p)
    (bad/'theme.json').write_text(json.dumps(t))
    results = {}
    for name, step in [('welcome', 350_000_000), ('fire', 700_000_000)]:
        for scale, mode in [(3, 'base'), (3, 'hd'), (3, 'mismatch'), (2, 'base'), (2, 'hd')]:
            case = f'{name}-{scale}-{mode}'
            d = out/case
            d.mkdir(exist_ok=True)
            shutil.copytree('/state/scratch', d/'scratch', dirs_exist_ok=True)
            cmd = ['/app/pod-overlay-shot', '-state', f'/state/{name}.state',
                   '-root', '/orig', '-scratch', str(d/'scratch'), '-text-dir', '/text',
                   '-font', '/font/pod-rounded.golemfnt', '-scale', str(scale),
                   '-shots', f'{step}:{name}', '-until', str(step+1000), '-out', str(d)]
            if mode != 'base':
                cmd += ['-hd', '/theme' if mode == 'hd' else str(bad)]
            p = subprocess.run(cmd, capture_output=True, timeout=90)
            (d/'stderr.log').write_bytes(p.stderr)
            (d/'receipt.json').write_bytes(p.stdout)
            if p.returncode:
                raise RuntimeError(f'{case}: {p.stderr[-1000:]}')
            receipt = json.loads(p.stdout)
            shot = receipt['shots'][0]
            # 舊 capture 的 stamps 清單不含 Layer.Art／HD Hold；在下方直接比較
            # A/B PNG 的已證實圖面矩形，不將舊工具的 art 誤報當作正式範圍判定。
            if shot.get('missing_glyphs'):
                raise AssertionError(f'{case}: 缺字')
            results[case] = shot
            print(case, '完成', flush=True)
    comparisons = []
    for name in ('welcome', 'fire'):
        for scale, mode in [(3, 'hd'), (3, 'mismatch'), (2, 'hd')]:
            base, other = f'{name}-{scale}-base', f'{name}-{scale}-{mode}'
            a, b = results[base], results[other]
            same = all(a[k] == b[k] for k in ('memory_sha256', 'indexed_sha256', 'registers', 'base_sha256'))
            scratch_a = {str(p.relative_to(out/base/'scratch')): sha(p) for p in (out/base/'scratch').rglob('*') if p.is_file()}
            scratch_b = {str(p.relative_to(out/other/'scratch')): sha(p) for p in (out/other/'scratch').rglob('*') if p.is_file()}
            same = same and scratch_a == scratch_b
            im_a, im_b = [np.asarray(Image.open(out/c/(name+'-overlay.png'))) for c in (base, other)]
            diff = np.any(im_a != im_b, axis=2)
            outside = diff.copy()
            outside[24*scale:112*scale, 24*scale:112*scale] = False
            changed = int(diff.sum())
            passed = same and not outside.any() and ((changed > 0) if scale == 3 and mode == 'hd' else changed == 0)
            row = dict(base=base, other=other, original_state_and_scratch_equal=same,
                       changed_pixels=changed, outside_portrait=int(outside.sum()), passed=bool(passed))
            comparisons.append(row)
    result = dict(scope='PIC1-131、PIC1-043@0；正常 DEMO 的加速同狀態重播，非首次繪製閃現驗收',
                  comparisons=comparisons, all_pass=all(r['passed'] for r in comparisons))
    (out/'summary.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(result, ensure_ascii=False, indent=2))
    if not result['all_pass']:
        raise SystemExit(1)


if __name__ == '__main__':
    main()
