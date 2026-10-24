#!/usr/bin/env python3
"""核對正常冷啟動的 HD／原版、2／3 倍切換收據。"""
import argparse
import hashlib
import json
import re
from pathlib import Path

import numpy as np
from PIL import Image


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--out', type=Path, required=True)
    ap.add_argument('--req', type=Path, required=True)
    ap.add_argument('--hd', type=Path, required=True)
    a = ap.parse_args()
    states = []
    for case in ('base', 'switched'):
        log = (a.out/case/'run.log').read_text()
        m = re.search(r'frames=2210 steps=(\d+) memory_sha256=([0-9a-f]+) indexed_sha256=([0-9a-f]+)', log)
        if not m:
            raise ValueError('缺正常結束收據')
        states.append(m.groups())
    if states[0] != states[1]:
        raise AssertionError('切換改變原版狀態')
    read = lambda f: np.asarray(Image.open(f).convert('RGBA'))
    shot = lambda n: read(a.out/'switched/shots'/f'pod-f{n:06d}.png')
    idx = (a.req/'pic1-131.idx').read_bytes()
    pal = (a.req/'pic1-131.pal').read_bytes()
    orig = Image.frombytes('P', (88, 88), idx)
    orig.putpalette(pal)
    expected = read(a.hd/'pic1-131.png')
    base = shot(2050)
    if not np.array_equal(base[72:336, 72:336], expected):
        raise AssertionError('HD 沒有實際出現')
    for frame in (2142, 2182):
        if not np.array_equal(shot(frame), base):
            raise AssertionError('返回 HD 畫面不同')
    for frame, scale in ((2102, 3), (2162, 2)):
        im = shot(frame)
        want = np.asarray(orig.convert('RGBA').resize((88*scale, 88*scale), Image.Resampling.NEAREST))
        if im.shape != (200*scale, 320*scale, 4) or not np.array_equal(im[24*scale:112*scale, 24*scale:112*scale], want):
            raise AssertionError('原版／2 倍模式圖框不等於原版')
    hashes = {str(p.relative_to(a.out)): hashlib.sha256(p.read_bytes()).hexdigest()
              for p in (a.out/'switched/shots').glob('*.png')}
    result = dict(all_pass=True, scope='正常冷啟動 PIC1-131；F12/F2 對應前端動作；圖框及返回整圖',
                  steps=states[0][0], memory_sha256=states[0][1], indexed_sha256=states[0][2], screenshots=hashes)
    (a.out/'summary.json').write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print('前端切換與原版狀態：通過')


if __name__ == '__main__':
    main()
