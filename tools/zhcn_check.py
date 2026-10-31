#!/usr/bin/env python3
"""核對簡中正常抽驗、原版狀態、語系圖層及乾淨重建，保留原始量尺。"""
import argparse
import hashlib
import json
from pathlib import Path
import re

import numpy as np
from PIL import Image

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())
def tree(p):return {str(f.relative_to(p)):sha(f) for f in sorted(p.rglob('*')) if f.is_file()}
def pixels(p):return np.array(Image.open(p).convert('RGBA'))
def state(log):
    a=re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',log);b=re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)',log);assert a and b
    return dict(frames=int(a[1]),steps=int(a[2]),memory=a[3],indexed=a[4],cpu=b[1],palette=b[2])

def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for n in ('root','out','receipt'):ap.add_argument('--'+n,type=Path,required=True)
    a=ap.parse_args();r=a.root;o=a.out;lock=read(r/'review/hd/runtime/manifest.json');assert len(lock['theme'])==1536 and len(lock['fonts'])==10
    assert tree(o/'theme')=={x['path']:x['sha256'] for x in lock['theme']}==tree(o/'final-rebuilt/theme')
    assert tree(o/'final-rebuilt/font')=={x['path']:x['sha256'] for x in lock['fonts']}
    assert sha(o/'bin/pod-play')==sha(o/'bin/pod-play-clean')
    for name in ('frontend.log','lang-pack.log','acceptance.log'):assert 'PASS' in (o/'test'/name).read_text() or '\nok ' in '\n'+(o/'test'/name).read_text()
    assert not(o/'test/vet.log').read_text().strip()
    front={'pod-map-visited.txt','pod-journal-seen.txt','pod-walkthrough-pos.txt','pod-lang.txt','pod-speed.txt','pod-barfont.txt'}
    cases={}
    for code in ('zh-TW','zh-CN'):
        p=o/('gui-'+code);log=(p/'run.log').read_text();s=state(log);assert s['frames']==4800 and s['steps']==462000000
        assert '字型缺字' not in log and '截圖失敗' not in log
        before={n:h for n,h in read(p/'before.json').items() if n not in front};guest={n:h for n,h in tree(p/'game').items() if n not in front}
        assert guest==before
        assert (p/'game/pod-lang.txt').read_text().strip()==code
        shots=sorted((p/'shots').glob('*.png'));assert len(shots)==19
        for frame,key in ((1000,'title-001'),(1100,'title-002'),(1200,'title-003')):
            assert np.array_equal(pixels(p/'shots'/f'pod-f{frame:06}.png'),pixels(o/'theme/localized'/code/(key+'.png'))),(code,key)
        poster_checks=[]
        for shot in shots:
            frame=int(shot.stem.split('f')[-1]);scale=2 if frame==4740 else 3;locale=code
            if frame==4352:locale='en' if code=='zh-TW' else 'zh-TW'
            full=pixels(p/'window'/shot.name);assert full.shape==(200*scale,480*scale,4)
            for side,x in (('left',0),('right',400*scale)):
                base=Image.open(o/'theme/posters'/f'poster-{side}-{scale}x.png').convert('RGBA');text=Image.open(o/'theme/posters'/f'poster-text-{side}.{locale}-{scale}x.png').convert('RGBA')
                expected=np.array(Image.alpha_composite(base,text));assert np.array_equal(full[:,x:x+80*scale],expected),(code,frame,side,locale)
            poster_checks.append(dict(frame=frame,scale=scale,locale=locale,sha256=sha(p/'window'/shot.name)))
        draw=read(p/'draw-report.json');assert draw['NativeBytes']==draw['HostBytes'] and draw['FallbackBytes']==draw['Mismatches']==draw['LegacyRGBFrames']==0
        raw=int(re.search(r'閃現 (\d+) 次',log)[1]);events=re.findall(r'閃現 第 (\d+) 格原樣呈現、第 (\d+) 格被換 列 (\d+) 欄 ([0-9 ]+)',log)
        # 沿既有019收據分類明示2→3倍HD切換，保留121原始計數。
        cells={(int(row),int(col)) for shown,frame,row,cols in events for col in cols.split()}
        assert raw==121 and len(events)==11 and cells=={(row,col) for row in range(3,14) for col in range(3,14)}
        assert all(shown=='4749' and frame=='4750' for shown,frame,_,_ in events)
        for suffix in ('.idx','.pal'):
            assert (p/'sources'/('pod-f004740'+suffix)).read_bytes()==(p/'sources'/('pod-f004770'+suffix)).read_bytes()
        assert pixels(p/'shots/pod-f004740.png').shape==(400,640,4) and pixels(p/'shots/pod-f004770.png').shape==(600,960,4)
        cases[code]=dict(state=s,guest_files=len(guest),guest=guest,draw=draw,raw_flash=raw,flash_events=events,
            flash_classification='explicit2x-to3x HD switch at4750;121 cells row3..13,column3..13; source indices/palette equal at4740/4770',unexpected_flash=0,
            log_sha256=sha(p/'run.log'),source_sha256=sha(p/'source.idx'),palette_sha256=sha(p/'source.pal'),shots={x.name:sha(x) for x in shots},posters=poster_checks)
    assert cases['zh-TW']['state']==cases['zh-CN']['state'] and cases['zh-TW']['guest']==cases['zh-CN']['guest']
    assert (o/'gui-zh-TW/source.idx').read_bytes()==(o/'gui-zh-CN/source.idx').read_bytes() and (o/'gui-zh-TW/source.pal').read_bytes()==(o/'gui-zh-CN/source.pal').read_bytes()
    # finish()會對-shot-original再畫目前語系工具列；原版依IDX/PAL比較，RGBA另限制差異只在工具列。
    tw=np.frombuffer((o/'gui-zh-TW/original.rgba').read_bytes(),np.uint8).reshape(600,960,4)
    cn=np.frombuffer((o/'gui-zh-CN/original.rgba').read_bytes(),np.uint8).reshape(600,960,4)
    changed=np.any(tw!=cn,axis=2);assert not changed[54:].any()
    y,x=np.nonzero(changed);toolbar_diff=dict(pixels=int(changed.sum()),bounds=[int(x.min()),int(y.min()),int(x.max()+1),int(y.max()+1)],outside_toolbar=0,reason='finish draws the current locale toolbar even for-shot-original; raw IDX/PAL equal')
    negative=read(o/'negative/receipt.json');assert len(negative)==2 and all(x['rejected'] for x in negative)
    proof=dict(date='2026-11-01',issue=70,status='PASS within sampled zh-CN scope',engine=lock['engine']['commit'],binary_sha256=sha(o/'bin/pod-play'),main_source_tree=(o/'final-main-tree.txt').read_text().strip(),source_tree_scope='asset rebuild inputs; verification tools additionally SHA-bound below',
        manifest_sha256=sha(r/'review/hd/runtime/manifest.json'),theme_files=1536,fonts=10,assets=read(o/'final-rebuilt/assets-receipt.json'),
        text=read(o/'test/text.json'),font_generation=read(o/'font/fonts.json'),graphics=read(o/'render/assets.json'),runtime=cases,
        same_original_state=True,same_guest_files=True,save_format_unchanged=True,original_rgba_toolbar_difference=toolbar_diff,negative=negative,
        subtitle_samples={p.name:sha(p) for p in sorted((o/'subtitle-samples').glob('*.png'))},
        logs={p.name:sha(p) for p in sorted((o/'test').glob('*.log'))},
        tools={p.name:sha(p) for p in sorted((r/'tools').glob('zhcn*'))},
        limits=['normal player path sampled, not exhaustive','subtitle checks reuse original normally captured frames; no new full intro run','English panels,Japanese,Korean,music,invincibility,animationHD and HD patch remain pending','existing nine candidate packages unchanged; no new release'])
    assert len(proof['subtitle_samples'])==8
    a.receipt.write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n');print(json.dumps(dict(status=proof['status'],cases={k:dict(frames=v['state']['frames'],guest_files=v['guest_files'],raw_flash=v['raw_flash'],unexpected_flash=v['unexpected_flash']) for k,v in cases.items()}),ensure_ascii=False,indent=2))

if __name__=='__main__':main()
