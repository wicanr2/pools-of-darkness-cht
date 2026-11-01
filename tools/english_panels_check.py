#!/usr/bin/env python3
"""核對一條正常英文面板路線及局部分頁，沿027固定原版終點，不擴大探索。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import sys

import numpy as np
from PIL import Image

sys.path.insert(0,str(Path(__file__).resolve().parent))
from english_panels import check as check_data
from zhcn_check import state

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())
def tree(p):return {str(f.relative_to(p)):sha(f) for f in sorted(p.rglob('*')) if f.is_file()}
def pixels(p):return np.array(Image.open(p).convert('RGBA'))
def dos_inputs(script):
    host={'journal','guide','map','pgup','pgdn','prev','next','shot','speeds','barfonts','scale','mapall','toggle','lang','cheat','bar'}
    return [(int(frame),name) for frame,name in (item.split(':',1) for item in script.strip().split(',')) if name not in host and not name.startswith(('speed=','barfont='))]

def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for n in ('root','out','receipt'):ap.add_argument('--'+n,type=Path,required=True)
    a=ap.parse_args();r=a.root;o=a.out;p=o/'gui';parent=read(r/'docs/re/027-zhcn-runtime.json');lock=read(r/'review/hd/runtime/manifest.json');log=(p/'run.log').read_text()
    assert sha(r/'workplace/dosgolem/workplace/bin/pod-play')==parent['binary_sha256']
    assert sha(r/'tools/zhcn_gui.sh')==parent['tools']['zhcn_gui.sh']
    old_script=re.search(r"script='([^']+)'",(r/'tools/zhcn_gui.sh').read_text())[1];new_script=(p/'script.txt').read_text();assert dos_inputs(old_script)==dos_inputs(new_script)
    end=state(log);assert end==parent['runtime']['zh-TW']['state']
    front={'pod-map-visited.txt','pod-journal-seen.txt','pod-walkthrough-pos.txt','pod-lang.txt','pod-speed.txt','pod-barfont.txt'}
    guest={n:h for n,h in tree(p/'game').items() if n not in front};before={n:h for n,h in read(p/'before.json').items() if n not in front}
    assert guest==before==parent['runtime']['zh-TW']['guest'] and len(guest)==132
    assert (p/'game/pod-lang.txt').read_text().strip()=='en'
    assert sha(p/'source.idx')==parent['runtime']['zh-TW']['source_sha256'] and sha(p/'source.pal')==parent['runtime']['zh-TW']['palette_sha256']
    assert '字型缺字' not in log and '截圖失敗' not in log
    draw=read(p/'draw-report.json');assert draw['HostBytes']==draw['NativeBytes'] and draw['FallbackBytes']==draw['Mismatches']==draw['LegacyRGBFrames']==0
    raw=int(re.search(r'閃現 (\d+) 次',log)[1]);events=re.findall(r'閃現 第 (\d+) 格原樣呈現、第 (\d+) 格被換 列 (\d+) 欄 ([0-9 ]+)',log)
    assert raw==121 and len(events)==11 and all(x[0]=='4749' and x[1]=='4750' for x in events)
    assert {(int(row),int(col)) for _,_,row,cols in events for col in cols.split()}=={(y,x) for y in range(3,14) for x in range(3,14)}
    for suffix in ('.idx','.pal'):assert (p/'sources'/('pod-f004740'+suffix)).read_bytes()==(p/'sources'/('pod-f004770'+suffix)).read_bytes()
    shots=sorted((p/'shots').glob('*.png'));assert len(shots)==17
    poster_checks=[]
    for f in shots:
        frame=int(f.stem.split('f')[-1]);scale=2 if frame==4740 else 3;code={4354:'zh-CN',4358:'zh-TW'}.get(frame,'en');full=pixels(p/'window'/f.name)
        assert full.shape==(200*scale,480*scale,4)
        for side,x in (('left',0),('right',400*scale)):
            folder=r/'workplace/hd/theme/posters';b=Image.open(folder/f'poster-{side}-{scale}x.png').convert('RGBA');t=Image.open(folder/f'poster-text-{side}.{code}-{scale}x.png').convert('RGBA')
            assert np.array_equal(full[:,x:x+80*scale],np.array(Image.alpha_composite(b,t)))
        poster_checks.append(dict(frame=frame,scale=scale,locale=code,sha256=sha(p/'window'/f.name)))
    theme=tree(r/'workplace/hd/theme');assert theme=={x['path']:x['sha256'] for x in lock['theme']}
    assert hashlib.sha256(json.dumps(theme,sort_keys=True).encode()).hexdigest()==parent['assets']['theme_tree_sha256']
    for n,h in parent['assets']['fonts'].items():assert sha(r/'workplace/font'/n)==h
    layout=read(o/'test/layout.json');assert layout['all_text_preserved'] and layout['all_lines_bounded'] and layout['locale_progress_preserved'] and layout['bad_header_rejected']
    assert layout['journals']==99 and layout['guide_chapters']==78
    negative=read(o/'negative/receipt.json');assert len(negative)==2 and all(x['rejected'] for x in negative)
    for name in ('layout.log','frontend.log'):assert '\nok ' in '\n'+(o/'test'/name).read_text()
    assert not(o/'test/vet.log').read_text().strip()
    data=check_data(r)
    assert read(o/'test/clean-data.json')==data
    source_tree=(o/'main-tree.txt').read_text().strip();assert re.fullmatch('[0-9a-f]{40}',source_tree)
    proof=dict(date='2026-11-01',issue=71,status='PASS within English panel sample scope',engine=lock['engine']['commit'],binary_sha256=parent['binary_sha256'],
        baseline=dict(path='docs/re/027-zhcn-runtime.json',sha256=sha(r/'docs/re/027-zhcn-runtime.json'),same_DOS_inputs=True,state=end),
        data=data,layout=layout,negative=negative,guest_files=132,guest_unchanged=True,original_state_equal=True,
        clean_git_data=dict(source_tree=source_tree,receipt_sha256=sha(o/'test/clean-data.json'),all_English_tables_and_source_match=True),
        normal=dict(script_sha256=sha(p/'script.txt'),source_sha256=sha(p/'source.idx'),palette_sha256=sha(p/'source.pal'),shots={x.name:sha(x) for x in shots},posters=poster_checks,draw=draw,
            raw_flash=raw,unexpected_flash=0,classification='explicit2x-to3x HD switch at4750; same121 cells and source indices/palette'),
        assets=dict(theme_files=len(theme),fonts=10,theme_and_fonts_unchanged_from027=True,manifest_sha256=sha(r/'review/hd/runtime/manifest.json')),
        rights='checked English journal transcript is original SSI text; private repository exception only; no public redistribution rights granted',
        local_layout_samples={f.name:sha(f) for f in sorted((o/'layout').glob('*.png'))},logs={f.name:sha(f) for f in sorted((o/'test').glob('*.log'))},
        tools={f.name:sha(f) for f in sorted((r/'tools').glob('english_panels*'))},
        limits=['one normal player route,17 screenshots; no extra exploration','fresh journal auto-trigger not replayed; journal body samples are component rendering','English guide translates existing project summaries and retains their uncertainties; no new gameplay oracle claims','font,HD assets,original rules and save format unchanged','Japanese,Korean,music approval,invincibility,full animationHD,HD patch and final packages remain pending'])
    a.receipt.write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n');print(json.dumps(dict(status=proof['status'],journals=99,guide_chapters=78,frames=end['frames'],guest_files=132,raw_flash=121,unexpected_flash=0),indent=2))

if __name__=='__main__':main()
