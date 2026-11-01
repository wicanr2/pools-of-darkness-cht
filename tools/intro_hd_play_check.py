#!/usr/bin/env python3
"""核對正常片頭HD、原版狀態與切換；保留原量尺並分類可證明的全黑／明示切換。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import numpy as np
from PIL import Image


def sha(data):return hashlib.sha256(data).hexdigest()
def tree(p):return {str(q.relative_to(p)):sha(q.read_bytes()) for q in sorted(p.rglob('*')) if q.is_file()}
def image(p):return np.array(Image.open(p).convert('RGBA'))


def check(base,hd,switch,theme):
    roots=dict(base=base,hd=hd,switch=switch);cases={}
    frontend={'pod-map-visited.txt','pod-journal-seen.txt','pod-walkthrough-pos.txt','pod-lang.txt','pod-speed.txt','pod-barfont.txt'}
    for mode,d in roots.items():
        log=(d/'run.log').read_text()
        end=re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',log)
        digest=re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)',log)
        audit=re.search(r'pod-play: 閃現 (\d+) 次',log)
        assert end and digest and audit and end.groups()[:2]==('5100','490875000')
        assert '截圖失敗' not in log and '字型缺字' not in log
        guest={k:v for k,v in tree(d/'save').items() if k not in frontend}
        inputs=json.loads((d/'inputs.json').read_text())
        cases[mode]=dict(end=end.groups(),digest=digest.groups(),guest=guest,inputs=inputs,flash_raw=int(audit[1]))
    for mode in ('hd','switch'):
        for key in ('end','digest','guest'):assert cases[mode][key]==cases['base'][key],(mode,key)
        for key in ('orig','font','text'):assert cases[mode]['inputs'][key]==cases['base']['inputs'][key],(mode,key)
        assert all(cases[mode]['inputs']['theme'].get(k)==v for k,v in cases['base']['inputs']['theme'].items()),'既有主題變動'
    assert cases['hd']['inputs']==cases['switch']['inputs']
    actualTheme=tree(theme)
    assert actualTheme==cases['hd']['inputs']['theme'],'受測主題已改變'
    meta=json.loads((theme/'intro.json').read_text())
    sources={f['index']['sha256']:f for f in meta['frames']}
    shots=[];full=0;switch_checks=[]
    shot_names=sorted(q.name for q in (hd/'shots').glob('pod-f*.png'))
    assert len(shot_names)==33
    for name in shot_names:
        n=int(Path(name).stem.split('f')[-1]);stem=Path(name).stem
        indices={mode:(p/'sources'/(stem+'.idx')).read_bytes() for mode,p in roots.items()}
        palettes={mode:(p/'sources'/(stem+'.pal')).read_bytes() for mode,p in roots.items()}
        assert len(set(indices.values()))==len(set(palettes.values()))==1,(n,'原版來源不同')
        idx,pal=indices['base'],palettes['base'];assert len(idx)==64000 and len(pal)==768
        rgb=np.frombuffer(pal,np.uint8).reshape(256,3)[np.frombuffer(idx,np.uint8).reshape(200,320)]
        raw3=rgb.repeat(3,0).repeat(3,1)
        original=image(base/'shots'/name);assert np.array_equal(original[:,:,:3],raw3)
        frame=sources.get(sha(idx[:53760]));got=image(hd/'shots'/name)
        assert got.shape==(600,960,4) and np.all(got[:,:,3]==255)
        compared=False
        if frame and frame['png']['path']:
            ph=sha(pal)
            light=next((p['light'] for p in meta['blocks'][frame['block']]['palettes'] if p['sha256']==ph),None)
            # 相同來源跨相鄰區塊重用時，僅接受該來源實際登記的色盤。
            if light is None:
                matches=[(f,p['light']) for f in meta['frames'] if f['index']['sha256']==sha(idx[:53760]) for p in meta['blocks'][f['block']]['palettes'] if p['sha256']==ph]
                assert matches,(n,'未知色盤');frame,light=matches[0]
            expected=image(theme/frame['png']['path'])
            expected[:,:,:3]=(expected[:,:,:3].astype(np.uint32)*light//65535).astype(np.uint8)
            assert np.array_equal(got[:504],expected),(n,'HD像素不合')
            full+=1;compared=True
        elif n==3768:
            assert not rgb[:168].any() and not got[:504,:,:3].any(),'全黑轉場不合'
            compared=True
        assert compared or n in (4994,4995),(n,'未核對的圖面')
        sw=image(switch/'shots'/name);scale=2 if n in (704,706,708) else 3
        assert sw.shape==(200*scale,320*scale,4) and np.all(sw[:,:,3]==255)
        raw=rgb.repeat(scale,0).repeat(scale,1)
        if n in (708,710):assert np.array_equal(sw[:,:,:3],raw),(n,'F12仍覆繪')
        elif scale==2:assert np.array_equal(sw[:168*scale,:,:3],raw[:168*scale]),(n,'二倍仍畫HD')
        else:assert np.array_equal(sw[:504],got[:504]),(n,'語系影響美術')
        if n in (702,704,706,708,710,712):switch_checks.append(n)
        shots.append(dict(frame=n,source_update=n-1,hd_png_sha256=sha((hd/'shots'/name).read_bytes()),
                          source_sha256=sha(idx),palette_sha256=sha(pal),art_checked=compared))
    # 原量尺的每個格子均保留；不能略過其他時點或靠改RGB通過。
    auditProof={}
    for mode in ('hd','switch'):
        records=[];log=(roots[mode]/'run.log').read_text()
        for a,b,row,cols in re.findall(r'閃現 第 (\d+) 格原樣呈現、第 (\d+) 格被換 列 (\d+) 欄 ([0-9 ]+)',log):
            a,b,row=int(a),int(b),int(row)
            for col in map(int,cols.split()):
                if (a,b)==(3767,3769):
                    assert 0<=row<21
                    q=roots[mode]/'sources/pod-f003768';idx=q.with_suffix('.idx').read_bytes();pal=q.with_suffix('.pal').read_bytes()
                    rgb=np.frombuffer(pal,np.uint8).reshape(256,3)[np.frombuffer(idx,np.uint8).reshape(200,320)]
                    assert not rgb[row*8:row*8+8,col*8:col*8+8].any()
                    im=image(roots[mode]/'shots/pod-f003768.png')
                    assert not im[row*24:row*24+24,col*24:col*24+24,:3].any()
                    reason='verified fully black transition'
                elif mode=='switch' and (a,b)==(706,711):
                    assert {704,706,708,710,712}.issubset(switch_checks)
                    reason='explicit F2/F12 control interval'
                else:raise AssertionError(('非預期量尺記錄',mode,a,b,row,col))
                records.append(dict(first=a,replaced=b,row=row,column=col,classification=reason))
        assert len(records)==cases[mode]['flash_raw']
        auditProof[mode]=dict(raw=len(records),black=sum(r['classification'].startswith('verified') for r in records),
                              explicit_switch=sum(r['classification'].startswith('explicit') for r in records),unexpected=0,records=records)
    stats={}
    for mode in ('hd','switch'):
        log=(roots[mode]/'run.log').read_text();s=json.loads(re.search(r'intro_hd=(\{.*\})',log)[1]);assert s['Calls']==s['KnownCalls']==133;stats[mode]=s
    first=cases['base']
    return dict(date='2026-11-01',issue=74,status='PASS sampled normal intro HD',frames=5100,steps=490875000,
                memory_sha256=first['end'][2],indexed_sha256=first['end'][3],cpu_sha256=first['digest'][0],palette_sha256=first['digest'][1],
                guest_files=len(first['guest']),guest_tree_sha256=sha(json.dumps(first['guest'],sort_keys=True).encode()),
                binary_sha256={m:v['inputs']['binary_sha256'] for m,v in cases.items()},
                theme_files=len(actualTheme),theme_tree_sha256=sha(json.dumps(actualTheme,sort_keys=True).encode()),
                original_inputs=first['inputs']['orig'],fonts=first['inputs']['font'],full_art_samples=full,shots=shots,
                audit=auditProof,stats=stats,
                limits=['SSI logo stays original; FINAL is separate','HD texture motion and brightness are artistic approximations',
                        'raw flash counter retained; black/switch records individually proven','Other speeds and unobserved palettes fall back; not whole-game parity'])


if __name__=='__main__':
    ap=argparse.ArgumentParser(description=__doc__)
    for key in ('base','hd','switch','theme','proof'):ap.add_argument('--'+key,type=Path,required=True)
    a=ap.parse_args();r=check(a.base,a.hd,a.switch,a.theme);a.proof.write_text(json.dumps(r,ensure_ascii=False,indent=2)+'\n');print('PASS',r['guest_files'],'DOS檔案；',r['full_art_samples'],'完整HD抽樣；原始量尺', {m:v['raw'] for m,v in r['audit'].items()},'非預期0')
