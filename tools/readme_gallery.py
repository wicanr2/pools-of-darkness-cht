#!/usr/bin/env python3
"""將已驗證的實際 UI 與核准封面預覽匯入 README，保存來源及裁切雜湊。"""
import argparse
import hashlib
import json
from pathlib import Path
import re

import numpy as np
from PIL import Image


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root',type=Path,required=True)
    ap.add_argument('--dosgolem-commit',required=True)
    a=ap.parse_args();p=a.root;out=p/'docs/screenshots';runtime=p/'workplace/probe/readme-ui'
    sha=lambda f:hashlib.sha256(f.read_bytes()).hexdigest()
    result=json.loads((out/'manifest.json').read_text())
    for row in result['items']:
        if sha(out/row['file'])!=row['sha256']:raise ValueError('現有展示圖雜湊改變')
    pattern=r'frames=4800 steps=(\d+) memory_sha256=([0-9a-f]+) indexed_sha256=([0-9a-f]+)'
    state=lambda f:re.search(pattern,f.read_text()).groups()
    current=state(runtime/'run.log');baseline=state(p/'workplace/probe/g8-map-play/run.log')
    if current!=baseline:raise ValueError('補截圖改變正常讀檔路線的原版狀態')
    rows=[]
    def save(name,source,kind,crop=None,**extra):
        im=Image.open(source).convert('RGBA')
        if crop:im=im.crop((crop[0],crop[1],crop[0]+crop[2],crop[1]+crop[3]))
        im.save(out/name)
        row=dict(file=name,kind=kind,source=str(source.relative_to(p)),source_sha256=sha(source),
                 sha256=sha(out/name),dimensions=list(im.size),issue='https://github.com/wicanr2/pools-of-darkness-cht/issues/25',**extra)
        if crop:row['crop_rect']=crop
        rows.append(row)
        return im
    def capture(frame):return runtime/'window-shots'/f'pod-f{frame:06d}.png'
    mapim=save('map-east-phlan.png',capture(4360),'normal-cold-boot-runtime-map',dosgolem_commit=a.dosgolem_commit,
               view='F7 對應動作，V 全圖；預設只顯示走過區域',research='docs/re/013-map-identification.md')
    game=Image.open(runtime/'shots/pod-f004360.png').convert('RGBA')
    if not np.array_equal(np.asarray(mapim)[:,240:1200],np.asarray(game)):raise ValueError('地圖中央畫布不同')
    save('bigpic1-240-original.png',p/'review/hd/b03-bigpic/bigpic1-240-original.png','decoded-original-bigpic',source_file='BIGPIC1.DAX',source_block=240)
    save('bigpic1-240-hd-preview.png',p/'review/hd/b03-bigpic/bigpic1-240.png','HD-art-preview-not-runtime',source_file='BIGPIC1.DAX',source_block=240)
    toolbar=save('toolbar.png',capture(4720),'normal-runtime-toolbar-crop',[240,0,960,36],dosgolem_commit=a.dosgolem_commit)
    if not np.array_equal(np.asarray(toolbar),np.asarray(Image.open(capture(4720)).convert('RGBA'))[:36,240:1200]):raise ValueError('工具列裁切改變像素')
    save('toolbar-window.png',capture(4720),'normal-cold-boot-runtime-toolbar',dosgolem_commit=a.dosgolem_commit)
    save('toolbar-speed-menu.png',capture(4750),'normal-cold-boot-runtime-toolbar-menu',dosgolem_commit=a.dosgolem_commit)
    save('toolbar-font-menu.png',capture(4780),'normal-cold-boot-runtime-toolbar-menu',dosgolem_commit=a.dosgolem_commit)
    review=p/'review/hd/b06-cover'
    meta=json.loads((review/'manifest.json').read_text())
    if meta['generation'].get('costume_approval')!='2026-10-24: 採用禮服版':raise ValueError('封面服裝未核准')
    save('game-cover-original.png',review/'title-002-original.png','decoded-original-title',source_file='TITLE.DAX',source_block=2)
    for code,browser,name in [('zh-TW','cover-zh.png','game-cover-hd-preview.png'),('en','cover-en.png','game-cover-hd-preview-en.png')]:
        im=save(name,review/browser,'localized-HD-cover-preview-not-runtime',locale=code,
                costume_exception='使用者核准完整藍紫色禮服',research='docs/re/014-vga-hd.md')
        if im.size!=(960,600) or im.getextrema()[3]!=(255,255):raise ValueError('封面成品尺寸／alpha 不符')
        im.save(p/'workplace/hd/out/b06'/f'title-002.{code}.png')
    for row in rows:result['items']=[r for r in result['items']if r['file']!=row['file']]+[row]
    (out/'manifest.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    receipt=dict(all_pass=True,scope='正常冷啟動4800格，沿已驗證路線讀槽A；新增F7/V全圖與工具列展開；封面為核准服裝的素材預覽',
        original_state=current,original_state_equal_to_previous=True,dosgolem_commit=a.dosgolem_commit,
        binary_sha256=sha(p/'workplace/dosgolem/workplace/bin/pod-play-readme'),script_sha256=sha(p/'tools/readme-ui-receipt.sh'),
        run_log_sha256=sha(runtime/'run.log'),screenshots=rows)
    (out/'gallery-receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
    meta['preview_outputs']={r['file']:r['sha256']for r in rows if r['kind'].endswith('not-runtime')}
    (review/'manifest.json').write_text(json.dumps(meta,ensure_ascii=False,indent=2)+'\n')
    print(f'README 新增 {len(rows)} 張地圖、工具列與封面圖；來源、精確裁切及原版狀態通過。')


if __name__=='__main__':main()
