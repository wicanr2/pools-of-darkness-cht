#!/usr/bin/env python3
"""獨立核對片尾播放器來源、HD像素、前端切換及正常CONTROL回歸。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import numpy as np
from PIL import Image


def sha(b):return hashlib.sha256(b).hexdigest()
def load(p):return json.loads(p.read_text())
def tree(p):return {str(q.relative_to(p)):sha(q.read_bytes()) for q in sorted(p.rglob('*')) if q.is_file()}
def image(p):return np.array(Image.open(p).convert('RGBA'))
FRONT={'pod-map-visited.txt','pod-journal-seen.txt','pod-walkthrough-pos.txt','pod-lang.txt','pod-speed.txt','pod-barfont.txt'}


def state(p):
    log=(p/'run.log').read_text()
    end=re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',log)
    digest=re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)',log)
    flash=re.search(r'pod-play: 閃現 (\d+) 次',log)
    assert end and digest and flash and '截圖失敗' not in log and '字型缺字' not in log
    return dict(end=list(end.groups()),digest=list(digest.groups()),flash=int(flash[1]),
                guest={k:v for k,v in tree(p/'save').items() if k not in FRONT})


def check(r,w):
    theme=w/'candidate/theme';run=w/'runtime-v2';meta=load(theme/'final.json');source=load(w/'source-proof.json')
    cases={m:state(run/m) for m in ('base','hd','switch')};base=cases['base']
    raw=load(w/'prefix-raw/timing.json');digest=raw['Digest']
    assert raw['Frames']==2001 and len(raw['Images'])==58
    assert base['end']==['2001',str(digest['Steps']),bytes(digest['MemorySHA256']).hex(),bytes(digest['IndexedSHA256']).hex()]
    assert base['digest']==[bytes(digest['CPUSHA256']).hex(),bytes(digest['PaletteSHA256']).hex()]
    for m in ('hd','switch'):
        for k in ('end','digest','guest'):assert cases[m][k]==base[k],(m,k)
    assert cases['hd']['flash']==cases['base']['flash']==0
    stats={m:json.loads(re.search(r'final_hd=(\{.*\})',(run/m/'run.log').read_text())[1]) for m in cases}
    assert all(s['Calls']==s['KnownCalls']==58 for s in stats.values())
    expected_events={e['update']:e for e in source['events']}
    shots=[];full=0;prefix=0;blocks=set();png_cache={}
    names=sorted(q.name for q in (run/'hd/shots').glob('*.png'));assert len(names)==83
    for name in names:
        stem=Path(name).stem;n=int(stem.split('f')[-1]);idxs={m:(run/m/'sources'/(stem+'.idx')).read_bytes() for m in cases}
        pals={m:(run/m/'sources'/(stem+'.pal')).read_bytes() for m in cases}
        assert len(set(idxs.values()))==len(set(pals.values()))==1,(n,'DOS source differs')
        idx,pal=idxs['base'],pals['base'];assert len(idx)==64000 and len(pal)==768
        event=expected_events[max(u for u in expected_events if u<=n-1)]
        src=source['source_files'][event['source']]
        # 原探針按上方索引及使用RGB去重；未使用色號可在同一來源再次出現時改變。
        # 全部768-byte色盤仍在三條GUI路線及終點digest嚴格比較。
        identity=bytearray(idx[:53760])
        for color in sorted(set(idx[:53760])):identity.extend(bytes([color])+pal[color*3:color*3+3])
        assert sha(idx)==src['indexed_sha256'] and sha(identity)==event['art_identity_sha256'],(n,'raw x86 oracle differs')
        rgb=np.frombuffer(pal,np.uint8).reshape(256,3)[np.frombuffer(idx,np.uint8).reshape(200,320)]
        raw3=rgb.repeat(3,0).repeat(3,1);orig=image(run/'base/shots'/name)
        assert np.array_equal(orig[:,:,:3],raw3),(n,'original view differs')
        got=image(run/'hd/shots'/name);assert got.shape==(600,960,4) and np.all(got[:,:,3]==255)
        assert np.array_equal(got[504:],orig[504:]),(n,'bottom text region changed')
        frames=[f for f in meta['frames'] if f['index']['sha256']==sha(idx[:53760])]
        compared=False
        for f in frames:
            lights=[p['light'] for p in meta['blocks'][f['block']]['palettes'] if p['sha256']==sha(pal)]
            if not lights:continue
            path=f['png']['path']
            if path not in png_cache:png_cache[path]=image(theme/path)
            want=png_cache[path].copy();want[:,:,:3]=(want[:,:,:3].astype(np.uint32)*lights[0]//65535).astype(np.uint8)
            if np.array_equal(got[:504],want):
                blocks.add(f['block']);full+=1;compared=True;break
        if not compared:
            assert event['kind'] in ('write-prefix','zero-write-prefix','outside-BMA'),(n,'HD full pixels differ')
            if event['kind']=='outside-BMA':assert not rgb.any() and not got[:,:,:3].any()
            else:prefix+=1
        sw=image(run/'switch/shots'/name);scale=2 if 451<=n<=460 else 3
        assert sw.shape==(200*scale,320*scale,4) and np.all(sw[:,:,3]==255)
        raw_scaled=rgb.repeat(scale,0).repeat(scale,1)
        if 431<=n<=440 or scale==2:assert np.array_equal(sw[:,:,:3],raw_scaled),(n,'F12/F2 bypass differs')
        else:assert np.array_equal(sw,got),(n,'language/control restore differs')
        assert np.array_equal(sw[168*scale:,:,:3],raw_scaled[168*scale:]),(n,'switch bottom differs')
        window=image(run/'switch/window'/name)
        assert window.shape==(200*scale,480*scale,4)
        assert np.array_equal(window[:,80*scale:400*scale],sw),(n,'central window differs')
        lang='zh-TW'
        for start,value in ((411,'en'),(413,'ja'),(415,'ko'),(417,'zh-CN'),(419,'zh-TW')):
            if n>=start:lang=value
        if 431<=n<=440:lang='en'
        for side,sl in [('left',slice(0,80*scale)),('right',slice(400*scale,480*scale))]:
            bg=Image.open(theme/f'posters/poster-{side}-{scale}x.png').convert('RGBA')
            text=Image.open(theme/f'posters/poster-text-{side}.{lang}-{scale}x.png').convert('RGBA')
            poster=np.array(Image.alpha_composite(bg,text))
            assert np.array_equal(window[:,sl],poster),(n,side,lang,'poster differs')
        shots.append(dict(frame=n,source_update=n-1,kind=event['kind'],source_sha256=sha(idx),palette_sha256=sha(pal),
                          hd_sha256=sha((run/'hd/shots'/name).read_bytes()),full_art_checked=compared,poster_language=lang,scale=scale))
    assert blocks==set(range(6)) and full>=60 and prefix==12
    # 保留原量尺的每個格子，只接受腳本預先指定的二倍結束時點。
    records=[];log=(run/'switch/run.log').read_text()
    for a,b,row,cols in re.findall(r'閃現 第 (\d+) 格原樣呈現、第 (\d+) 格被換 列 (\d+) 欄 ([0-9 ]+)',log):
        assert (int(a),int(b))==(459,460),'unexpected flash interval'
        for col in map(int,cols.split()):records.append(dict(first=int(a),replaced=int(b),row=int(row),column=col,classification='explicit F2 2x to 3x interval'))
    assert len(records)==cases['switch']['flash']==546
    demo=state(run/'demo');expected=load(r/'review/hd/runtime/manifest.json')['runtime_expected']
    assert demo['end']==expected['end'] and demo['digest']==expected['digest']
    assert demo['guest']==expected['guest'] and demo['flash']==0
    assert sha((run/'demo/frame.png').read_bytes())==expected['frame_sha256']
    ds=load(run/'demo/draw-report.json');assert ds['HostBytes']==ds['NativeBytes'] and ds['FallbackBytes']==ds['Mismatches']==0
    old_packages=[]
    for fname in ('026-package-runtime.json','033-hd-patch-runtime.json'):
        receipt=load(r/'docs/re'/fname)
        for row in receipt['files']:
            found=list((r/'dist-all'/receipt['version']).rglob(row['name']));assert len(found)==1
            p=found[0];assert p.stat().st_size==row['size'] and sha(p.read_bytes())==row['sha256']
            old_packages.append(dict(path=str(p.relative_to(r)),sha256=row['sha256'],size=row['size']))
    assert len(old_packages)==12
    manifest=load(r/'review/hd/runtime/manifest.json');tt=tree(theme);fonts=tree(r/'workplace/font')
    assert tt=={x['path']:x['sha256'] for x in manifest['theme']} and len(tt)==2034
    assert fonts=={x['path']:x['sha256'] for x in manifest['fonts']} and len(fonts)==18
    for name,h in source['input_sha256'].items():assert sha((r/'workplace/input/dos/pod'/name).read_bytes())==h
    build=load(w/'build.json');assert build['entry_overlay_only'] and build['binary_sha256']==load(w/'clean-build/build.json')['binary_sha256']
    return dict(date='2026-11-02',issue=77,status='PASS FINAL direct player and normal CONTROL regression',
        scope='Original VACE FINAL direct entry, not a replayed campaign victory',engine_commit=manifest['engine']['commit'],
        binary_sha256=build['binary_sha256'],clean_engine_build_equal=True,theme_files=len(tt),fonts=fonts,
        theme_tree_sha256=sha(json.dumps(tt,sort_keys=True).encode()),frames=2001,steps=int(base['end'][1]),
        memory_sha256=base['end'][2],indexed_sha256=base['end'][3],cpu_sha256=base['digest'][0],palette_sha256=base['digest'][1],
        guest_files=len(base['guest']),guest_tree_sha256=sha(json.dumps(base['guest'],sort_keys=True).encode()),
        original_source_exit=source['timing'],stats=stats,full_art_samples=full,prefix_source_samples=prefix,shots=shots,
        audit=dict(hd_raw=0,switch_raw=len(records),unexpected=0,explicit_switch_records=records),
        demo=dict(frames=7400,steps=int(demo['end'][1]),end=demo['end'],digest=demo['digest'],guest_files=len(demo['guest']),
                  frame_sha256=expected['frame_sha256'],flash=0,draw=ds),old_packages_unchanged=old_packages,
        tools={str(p.relative_to(r)):sha(p.read_bytes()) for p in sorted((r/'tools').glob('final_*'))},
        limits=['Final entry queue is test-only; production CONTROL entry is unchanged',
                'GUI ends at update 2001 to save a receipt; unmodified x86 probe exits normally at 2002',
                'HD texture motion, light and reveal are artistic approximations',
                'No Windows/macOS native validation or new packages in this batch'])


if __name__=='__main__':
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root',type=Path,required=True);ap.add_argument('--work',type=Path,required=True);ap.add_argument('--proof',type=Path,required=True)
    a=ap.parse_args();result=check(a.root,a.work);a.proof.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    print('PASS',result['full_art_samples'],'full HD,',result['prefix_source_samples'],'prefix,',result['guest_files'],'guest files; normal DEMO flash 0')
