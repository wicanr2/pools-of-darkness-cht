#!/usr/bin/env python3
"""將已驗證的實際 UI 與核准封面預覽匯入 README，保存來源及裁切雜湊。"""
import argparse
import hashlib
import json
from pathlib import Path
import re

import numpy as np
from PIL import Image


def add_battle(root, commit, runtime):
    out=root/'docs/screenshots'
    sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
    report=json.loads((runtime/'summary.json').read_text())
    if not report['guest_state_equal'] or report['outside_battlefield_changes'] or report['audit']!='pod-play: 閃現 0 次':
        raise ValueError('戰鬥正常路徑尚未通過')
    manifest=json.loads((out/'manifest.json').read_text())
    for row in manifest['items']:
        if sha(out/row['file'])!=row['sha256']:raise ValueError('現有展示圖雜湊改變')
    rows=[]
    for name,case,crop in [('demo-battle-original.png','english',[240,0,960,600]),
                           ('demo-battle-cht-hd.png','hd',[240,0,960,600]),
                           ('demo-battle-window.png','hd',None)]:
        source=runtime/case/'frame.png'
        if sha(source)!=report['frames_sha256'][case]:raise ValueError('戰鬥擷取已變更')
        image=Image.open(source).convert('RGBA')
        if crop:image=image.crop((crop[0],crop[1],crop[0]+crop[2],crop[1]+crop[3]))
        image.save(out/name)
        rows.append(dict(file=name,kind='normal-cold-boot-runtime-battle',source=str(source.relative_to(root)),
            source_sha256=sha(source),sha256=sha(out/name),dimensions=list(image.size),crop_rect=crop,
            dosgolem_commit=commit,research='docs/re/017-vga-battle-hd.md',
            issue='https://github.com/wicanr2/pools-of-darkness-cht/issues/23'))
    for row in rows:manifest['items']=[r for r in manifest['items']if r['file']!=row['file']]+[row]
    (out/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
    (out/'battle-receipt.json').write_text(json.dumps(dict(runtime_receipt_sha256=sha(runtime/'summary.json'),
        dosgolem_commit=commit,screenshots=rows),ensure_ascii=False,indent=2)+'\n')
    print('README 戰鬥三圖：原版／繁中HD中央精確裁切及完整視窗，來源雜湊通過。')


def add_titles(root,commit,runtime):
    out=root/'docs/screenshots';sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
    report=json.loads((runtime/'summary.json').read_text())
    if not report['all_pass']or report['flashes']or len(report['poster_checks'])!=17:raise ValueError('正常開場驗收未完成')
    original=rgba_file(out/'game-cover-original.png')
    if not np.array_equal(original,rgba_file(runtime/'base/shots/pod-f001100.png')):raise ValueError('原版封面不符正常冷啟動')
    manifest=json.loads((out/'manifest.json').read_text());rows=[]
    for row in manifest['items']:
        if sha(out/row['file'])!=row['sha256']:raise ValueError('現有展示圖變更')
    for name,case,frame,locale,whole in [('game-cover-hd.png','hd',1100,'zh-TW',False),
            ('game-cover-hd-en.png','switch',1106,'en',False),('game-cover-window.png','hd',1100,'zh-TW',True)]:
        folder='window-shots'if whole else'shots';source=runtime/case/folder/f'pod-f{frame:06d}.png'
        candidates=report['poster_checks']if whole else report['screenshots']
        expected=next(v['sha256']for v in candidates if v['case']==case and v['frame']==frame)
        if sha(source)!=expected:raise ValueError('開場收據已變更')
        image=Image.open(source).convert('RGBA');image.save(out/name)
        if not whole:
            preview=out/('game-cover-hd-preview.png'if locale=='zh-TW'else'game-cover-hd-preview-en.png')
            if not np.array_equal(np.array(image),rgba_file(preview)):raise ValueError('偏離使用者核准的封面')
        rows.append(dict(file=name,kind='normal-cold-boot-runtime-title',locale=locale,dimensions=list(image.size),
            source=str(source.relative_to(root)),source_sha256=sha(source),sha256=sha(out/name),dosgolem_commit=commit,
            research='docs/re/014-vga-hd.md',issue='https://github.com/wicanr2/pools-of-darkness-cht/issues/26'))
    for row in rows:manifest['items']=[v for v in manifest['items']if v['file']!=row['file']]+[row]
    (out/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
    (out/'title-receipt.json').write_text(json.dumps(dict(runtime_receipt_sha256=sha(runtime/'summary.json'),
        dosgolem_commit=commit,screenshots=rows),ensure_ascii=False,indent=2)+'\n')
    print('正常冷啟動封面三圖已匯入README；核准服裝及標題完全相同。')


def rgba_file(path):return np.array(Image.open(path).convert('RGBA'))


def add_walls(root, commit, runtime):
    out = root/'docs/screenshots'
    sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
    report = json.loads((runtime/'summary.json').read_text())
    if not report['original_state_equal'] or not report['guest_state_equal'] or report['flash_count'] or report['outside_viewport_changes']:
        raise ValueError('正常街景尚未通過')
    manifest = json.loads((out/'manifest.json').read_text())
    for row in manifest['items']:
        if sha(out/row['file']) != row['sha256']: raise ValueError('現有展示圖已變更')
    rows = []
    captures = [('street-original.png', 'original.rgba', None),
                ('street-cht-hd.png', 'shots/pod-f004200.png', None),
                ('street-window.png', 'window-shots/pod-f004200.png', None),
                ('map-east-phlan.png', 'window-shots/pod-f004360.png', None),
                ('toolbar.png', 'window-shots/pod-f004720.png', (240,0,1200,36)),
                ('toolbar-window.png', 'window-shots/pod-f004720.png', None),
                ('toolbar-speed-menu.png', 'window-shots/pod-f004750.png', None),
                ('toolbar-font-menu.png', 'window-shots/pod-f004780.png', None)]
    for name, rel, crop in captures:
        source = runtime/'hd'/rel
        if source.suffix == '.rgba':
            im = Image.frombytes('RGBA', (960,600), source.read_bytes())
        else:
            im = Image.open(source).convert('RGBA')
        if crop: im = im.crop(crop)
        im.save(out/name)
        rows.append(dict(file=name, kind='normal-cold-boot-runtime-wall-gallery',
            source=str(source.relative_to(root)), source_sha256=sha(source), sha256=sha(out/name),
            dimensions=list(im.size), crop_rect=crop, dosgolem_commit=commit,
            research='docs/re/014-vga-hd.md', issue='https://github.com/wicanr2/pools-of-darkness-cht/issues/30'))
    for row in rows: manifest['items'] = [r for r in manifest['items'] if r['file'] != row['file']]+[row]
    (out/'manifest.json').write_text(json.dumps(manifest, ensure_ascii=False, indent=2)+'\n')
    (out/'wall-receipt.json').write_text(json.dumps(dict(runtime_receipt_sha256=sha(runtime/'summary.json'),
        binary_sha256=report['binary_sha256'], dosgolem_commit=commit, screenshots=rows), ensure_ascii=False, indent=2)+'\n')
    print('正常街景原版／HD、地圖及工具列八圖已匯入README。')


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root',type=Path,required=True)
    ap.add_argument('--dosgolem-commit',required=True)
    ap.add_argument('--battle',type=Path,help='已通過四路正常戰鬥驗收的收據根')
    ap.add_argument('--titles',type=Path,help='已通過正常開場及語系切換的收據根')
    ap.add_argument('--ui-runtime',type=Path,help='相同正常讀檔路線的新地圖／工具列收據根')
    ap.add_argument('--walls',type=Path,help='已通過正常街景的兩路收據根')
    a=ap.parse_args()
    if a.walls:
        add_walls(a.root,a.dosgolem_commit,a.walls)
        return
    if a.battle:
        add_battle(a.root,a.dosgolem_commit,a.battle)
        return
    if a.titles:
        add_titles(a.root,a.dosgolem_commit,a.titles)
        return
    p=a.root;out=p/'docs/screenshots';runtime=a.ui_runtime or p/'workplace/probe/readme-ui'
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
