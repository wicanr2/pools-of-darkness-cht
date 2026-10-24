#!/usr/bin/env python3
"""核對正常前端海報、語系、倍率、定版原型及原版狀態收據。"""
import argparse
import hashlib
import json
import re
from pathlib import Path

import numpy as np
from PIL import Image


def read(p):
    return np.asarray(Image.open(p).convert('RGBA'))


def state(p):
    m = re.search(r'frames=2210 steps=(\d+) memory_sha256=([0-9a-f]+) indexed_sha256=([0-9a-f]+)', p.read_text())
    if not m:
        raise ValueError(f'缺正常結束收據：{p}')
    return m.groups()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--out', type=Path, required=True, help='switch 收據根')
    ap.add_argument('--without', type=Path, required=True, help='同版本停用海報的收據目錄')
    ap.add_argument('--prototype', type=Path, required=True)
    a = ap.parse_args()
    roots = [a.out/'base', a.out/'switched', a.without]
    states = [state(p/'run.log') for p in roots]
    if any(s != states[0] for s in states[1:]):
        raise AssertionError('海報或切換改變原版狀態')
    # 前端設定檔預期會變，DOS 寫入檔必須逐檔相同。
    frontend = {'pod-map-visited.txt', 'pod-speed.txt', 'pod-barfont.txt', 'pod-language.txt'}
    saves = [{str(p.relative_to(d/'save')): hashlib.sha256(p.read_bytes()).hexdigest()
              for p in (d/'save').rglob('*') if p.is_file() and str(p.relative_to(d/'save')) not in frontend}
             for d in roots]
    if any(s != saves[0] for s in saves[1:]):
        raise AssertionError('DOS 寫入檔不同')
    checks = []
    shots = a.out/'switched'
    for n, scale, lang in ((2050,3,'zh-TW'),(2102,3,'en'),(2142,3,'zh-TW'),(2162,2,'zh-TW'),(2182,3,'zh-TW')):
        name = f'pod-f{n:06d}.png'
        game, window = read(shots/'shots'/name), read(shots/'window-shots'/name)
        if window.shape != (200*scale,480*scale,4):
            raise AssertionError('完整視窗尺寸錯誤')
        if not np.array_equal(window[:,80*scale:400*scale], game):
            raise AssertionError('中央遊戲畫布被改動')
        row = dict(frame=n, scale=scale, locale=lang, central_equal=True)
        if scale == 3:
            original = read(a.prototype/('wide-zh.png' if lang=='zh-TW' else 'wide-en.png'))[56:656]
            sides = np.concatenate((window[:,:240],window[:,1200:]),axis=1)
            expected = np.concatenate((original[:,:240],original[:,1200:]),axis=1)
            diff = np.abs(sides.astype(int)-expected.astype(int))
            row.update(prototype_max_channel_difference=int(diff.max()),
                       prototype_different_pixels=int(np.any(diff,axis=2).sum()))
            if diff.any():
                raise AssertionError(f'海報偏離使用者定版：{row}')
        checks.append(row)
    for d in roots[:2]:
        window=read(d/'final-window.png')
        game=np.frombuffer((d/'final.rgba').read_bytes(),np.uint8).reshape(600,960,4)
        if not np.array_equal(window[:,240:1200],game):
            raise AssertionError('含切換列的中央畫布不一致')
    if (roots[0]/'final.rgba').read_bytes() != (a.without/'final.rgba').read_bytes():
        raise AssertionError('停用海報前後遊戲畫布不同')
    result = dict(all_pass=True, scope='正常冷啟動 2210 格，F2/F12 對應動作，已定版兩側；未涵蓋其他語系文字',
                  original_state=states[0], dos_written_files=saves[0], checks=checks)
    (a.out/'poster-summary.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    print(f"海報驗收通過：{len(checks)} 個畫面，中央與定版兩側像素相同，{len(saves[0])} 個 DOS 檔案一致")


if __name__ == '__main__':
    main()
