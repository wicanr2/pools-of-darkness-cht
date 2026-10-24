#!/usr/bin/env python3
"""Verify normally booted movie subtitles and host-only switches."""
import argparse
import hashlib
import json
import re
from pathlib import Path


def sha(data):
    return hashlib.sha256(data).hexdigest()


def tree(root):
    return {str(p.relative_to(root)): sha(p.read_bytes()) for p in sorted(root.rglob('*')) if p.is_file()}


def check(root, source_proof):
    import numpy as np
    from PIL import Image
    proof = json.loads(source_proof.read_text())
    cues = {item['crop_indexed_sha256']: item['key'] for item in proof['items']}
    cases = {}
    frontend = {'pod-map-visited.txt','pod-journal-seen.txt','pod-walkthrough-pos.txt',
                'pod-lang.txt','pod-speed.txt','pod-barfont.txt'}
    for mode in ('base','subtitles'):
        d = root / mode
        log = (d/'run.log').read_text()
        end = re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',log)
        digest = re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)',log)
        audit = re.search(r'pod-play: 閃現 (\d+) 次',log)
        assert end and digest and audit
        assert end.groups()[:2] == ('5100','490875000')
        assert int(audit[1]) == 0
        assert '字型缺字' not in log and '截圖失敗' not in log
        guest = {k:v for k,v in tree(d/'save').items() if k not in frontend}
        cases[mode] = {'end':end.groups(),'digest':digest.groups(),'guest':guest,
                       'inputs':json.loads((d/'inputs.json').read_text()),'flash':int(audit[1])}
    assert cases['base'] == cases['subtitles'], '前端切換改變原版執行狀態或輸入'
    first = cases['base']
    assert first['end'][2] == proof['digest']['MemorySHA256']
    assert first['end'][3] == proof['digest']['IndexedSHA256']
    assert first['digest'] == (proof['digest']['CPUSHA256'],proof['digest']['PaletteSHA256'])
    shots=[]
    firsts={602:'intro.001',947:'intro.002',1625:'intro.003',2293:'intro.004',
            2516:'intro.005',3162:'intro.006',3770:'intro.007',3985:'intro.008'}
    for p in sorted((root/'subtitles/shots').glob('pod-f*.png')):
        frame=int(p.stem.split('f')[-1]);scale=2 if frame in (704,706,708) else 3
        stem=p.stem
        files={mode:root/mode/'sources'/stem for mode in cases}
        idx=files['subtitles'].with_suffix('.idx').read_bytes()
        pal=files['subtitles'].with_suffix('.pal').read_bytes()
        assert idx==files['base'].with_suffix('.idx').read_bytes()
        assert pal==files['base'].with_suffix('.pal').read_bytes()
        assert len(idx)==64000 and len(pal)==768
        raw=np.frombuffer(pal,dtype=np.uint8).reshape(256,3)[np.frombuffer(idx,dtype=np.uint8).reshape(200,320)]
        original=raw.repeat(scale,0).repeat(scale,1)
        got=np.array(Image.open(p).convert('RGBA'))
        assert got.shape==(200*scale,320*scale,4) and np.all(got[:,:,3]==255)
        assert np.array_equal(got[:168*scale,:,:3],original[:168*scale]), f'{frame}: 動畫改變'
        base=np.array(Image.open(root/'base/shots'/p.name).convert('RGBA'))
        assert np.array_equal(base[:,:,:3],raw.repeat(3,0).repeat(3,1)), 'F12原版畫面不符來源'
        cue=cues.get(sha(idx[168*320:]))
        if frame in firsts:
            assert cue==firsts[frame],f'{frame}: 首次出現來源不符'
        diff=int(np.count_nonzero(np.any(got[:,:,:3]!=original,axis=2)))
        if frame in (708,710) or cue is None:
            assert diff==0, f'{frame}: 原版或空白畫格仍覆繪'
        else:
            assert diff>0,f'{frame}: 未覆繪字幕'
        window=Image.open(root/'subtitles/window-shots'/p.name)
        assert window.size==(480*scale,200*scale)
        center=np.array(window.convert('RGBA'))[:,80*scale:400*scale]
        assert np.array_equal(center,got),'海報視窗中央與實際合成不同'
        shots.append({'frame':frame,'source_host_update':frame-1,'key':cue,'scale':scale,
                      'mode':'original' if frame in (708,710) else ('en' if frame in (702,704) else 'zh-TW'),
                      'changed_pixels':diff,'png_sha256':sha(p.read_bytes()),
                      'indexed_sha256':sha(idx),'palette_sha256':sha(pal)})
    assert len(shots)==18 and len(firsts)==8
    return {'date':'2026-10-25','status':'conformed-sampled-intro-only',
            'route':'normal SND.COM/CONTROL boot, 5100 host updates, no DOS key; F8/F2/F12 host switches',
            'frames':5100,'steps':490875000,'memory_sha256':first['end'][2],
            'indexed_sha256':first['end'][3],'cpu_sha256':first['digest'][0],
            'palette_sha256':first['digest'][1],'guest_files':len(first['guest']),
            'guest_tree_sha256':sha(json.dumps(first['guest'],sort_keys=True).encode()),
            'inputs':first['inputs'],'flash':0,'shots':shots,
            'limits':['Eight intro cues only; FINAL and movie HD are not implemented',
                      'Source-frame tests cover both languages at both scales; window switches sampled in cue 1',
                      'No claim for other movie versions or unobserved execution timing']}


if __name__=='__main__':
    p=argparse.ArgumentParser()
    p.add_argument('--capture',type=Path,required=True)
    p.add_argument('--source-proof',type=Path,required=True)
    p.add_argument('--proof',type=Path,required=True)
    a=p.parse_args()
    result=check(a.capture,a.source_proof)
    a.proof.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    print('八組片頭、雙語／倍率／原版切換、原版狀態及閃現0通過。')
