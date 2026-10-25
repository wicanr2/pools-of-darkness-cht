#!/usr/bin/env python3
"""補驗既有正常DEMO的BIGPIC244；須先通過完整戰鬥四路及舊原版終點比較。"""
import argparse
import hashlib
import json
from pathlib import Path

import numpy as np
from PIL import Image


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def rgba(p):
    return np.array(Image.open(p).convert('RGBA'))


def check(out, theme, proof):
    summary=out/'summary.json';report=json.loads(summary.read_text())
    assert report['frames']==7400 and report['registered_actor_states']==238
    assert report['prior_original_state_equal'] and report['prior_sample_composite_equal']
    assert report['guest_state_equal'] and report['guest_files']==127 and report['audit']=='pod-play: 閃現 0 次'
    sources=json.loads(proof.read_text());source=next(i for i in sources['sources']if i['key']=='bigpic1-244')
    items=json.loads((theme/'theme.json').read_text())['items'];item=next(i for i in items if i['key']=='bigpic1-244')
    assert item['screen_rect']==[8,8,304,120] and sha(theme/item['indexed'])==source['indexed_sha256']
    assert sha(theme/item['png'])==item['png_sha256']
    reference=np.frombuffer((theme/item['indexed']).read_bytes(),np.uint8).reshape(120,304)
    expected=rgba(theme/item['png']);cases={};screenshots=[]
    for case in ('hd','base','english','twox'):
        folder=out/case;stem='pod-f003650';scale=2 if case=='twox'else 3
        idxfile=folder/'sources'/(stem+'.idx');palfile=idxfile.with_suffix('.pal')
        indexed=np.frombuffer(idxfile.read_bytes(),np.uint8).reshape(200,320)
        pal=np.frombuffer(palfile.read_bytes(),np.uint8).reshape(256,3)
        assert np.array_equal(indexed[8:128,8:312],reference),case+'：來源圖不符'
        used=sorted(set(reference.flatten().tolist()));signature=bytes(v for c in used for v in [c,*pal[c].tolist()])
        assert hashlib.sha256(signature).hexdigest()==source['used_palette_sha256'],case+'：使用RGB不符'
        original=pal[indexed].repeat(scale,0).repeat(scale,1)
        original=np.concatenate([original,np.full((*original.shape[:2],1),255,np.uint8)],axis=2)
        path=folder/'shots'/(stem+'.png');image=rgba(path)
        assert image.shape==(200*scale,320*scale,4)
        crop=image[8*scale:128*scale,8*scale:312*scale]
        assert np.array_equal(crop,expected if case=='hd'else original[8*scale:128*scale,8*scale:312*scale]),case+'：成品或原版回退不符'
        window=folder/'window-shots'/(stem+'.png');wide=rgba(window)
        assert wide.shape==(200*scale,480*scale,4) and np.array_equal(wide[:,80*scale:400*scale],image)
        cases[case]=(idxfile.read_bytes(),palfile.read_bytes(),image)
        screenshots.append(dict(case=case,frame=3650,source_frame=3649,scale=scale,sha256=sha(path),window_sha256=sha(window)))
    assert all(cases[case][:2]==cases['hd'][:2]for case in cases),'四路抽樣來源不同'
    diff=np.any(cases['hd'][2]!=cases['base'][2],axis=2);allowed=np.zeros((600,960),bool);allowed[24:384,24:936]=True
    assert np.any(diff) and not np.any(diff&~allowed),'BIGPIC抽樣差異超出圖框'
    report.update(scope='完整7400格正常DEMO四路與舊原版終點；3650格BIGPIC244；其他六張只驗條件式來源',
                  producer=dict(script='tools/bigpic_play_check.py',sha256=sha(Path(__file__))),
                  battle_summary_sha256=sha(summary),bigpic_source_proof_sha256=sha(proof),
                  bigpic_theme_sha256=sha(theme/'theme.json'),bigpic_key='bigpic1-244',
                  bigpic_changed_pixels=int(np.count_nonzero(diff)),bigpic_outside_changes=0,
                  bigpic_f12_original=True,bigpic_twox_original=True,bigpic_source_equal=True,
                  bigpic_screenshots=screenshots,conditional_bigpic_keys=[i['key']for i in items if i['key'].startswith('bigpic1-')and i['key']!='bigpic1-244'])
    (out/'bigpic-summary.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    return report


if __name__=='__main__':
    ap=argparse.ArgumentParser(description=__doc__)
    for arg in ('out','theme','proof'):
        ap.add_argument('--'+arg,type=Path,required=True)
    a=ap.parse_args();r=check(a.out,a.theme,a.proof)
    print(json.dumps({k:v for k,v in r.items()if k not in ('unobserved_actor_keys','matches')},ensure_ascii=False,indent=2))
