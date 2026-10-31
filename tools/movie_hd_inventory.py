#!/usr/bin/env python3
"""固定正常片頭上方168列的像素／色盤狀態及六個HD關鍵畫面；不接入呈現。"""
import argparse
import copy
import hashlib
import json
from pathlib import Path

from PIL import Image

AREA = 320 * 168
SELECTIONS = [('council',601),('phlan',946),('horned',1624),
              ('heroes',2292),('followers',2515),('dreadlord',3161)]


def sha(data):
    return hashlib.sha256(data).hexdigest()


def normalized(digest):
    return {k:bytes(v).hex() if isinstance(v,list) else v for k,v in digest.items()}


def signature(idx,pal):
    assert len(idx)==64000 and len(pal)==768
    used=sorted(set(idx[:AREA]))
    return sha(idx[:AREA]+b''.join(bytes([v])+pal[v*3:v*3+3] for v in used))


def verify_digest(actual,expected):
    assert normalized(actual)==expected, '原版終點不同'


def verify_source(idx,pal,want):
    assert signature(idx,pal)==want, '像素或使用RGB改變'


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for name in ('capture','baseline','baseline-capture','proof','req'):
        ap.add_argument('--'+name,type=Path,required=True)
    a=ap.parse_args();cap=a.capture
    timing=json.loads((cap/'timing.json').read_text());inputs=json.loads((cap/'inputs.json').read_text())
    baseline=json.loads(a.baseline.read_text())
    assert timing['Frames']==5100 and timing['Budget']==96250
    verify_digest(timing['Digest'],baseline['digest'])
    assert inputs['original']==baseline['inputs']['original'] and inputs['seed']==baseline['inputs']['seed']
    assert inputs['source_sha256']==sha((Path(__file__).parent/'movie_hd_probe.go').read_bytes())
    tree=lambda p:{str(f.relative_to(p)):sha(f.read_bytes()) for f in sorted(p.rglob('*')) if f.is_file()}
    guest=tree(cap/'save');assert guest==tree(a.baseline_capture/'save') and len(guest)==127
    assert sha((cap/'movie-hd-probe').read_bytes())==inputs['probe_sha256']
    events=timing['Events'];assert len(events)==172 and events[0]['Update']==1
    assert all(x['Update']<y['Update'] for x,y in zip(events,events[1:]))
    sources={}
    for e in events:
        name=e['File'];assert name.startswith('source-') and Path(name).name==name
        idx=(cap/(name+'.idx')).read_bytes();pal=(cap/(name+'.pal')).read_bytes()
        verify_source(idx,pal,e['SHA'])
        item=sources.setdefault(name,dict(file=name,indexed_sha256=sha(idx),palette_sha256=sha(pal),
            art_indexed_sha256=sha(idx[:AREA]),art_identity_sha256=e['SHA'],
            first_update=e['Update'],last_transition_update=e['Update'],updates=[]))
        assert item['art_identity_sha256']==e['SHA'];item['updates'].append(e['Update']);item['last_transition_update']=e['Update']
    assert len(sources)==163 and len({x['art_indexed_sha256'] for x in sources.values()})==129
    controls=[]
    try:
        bad=copy.deepcopy(timing['Digest']);bad['CPUSHA256'][0]^=1;verify_digest(bad,baseline['digest'])
    except AssertionError:controls.append(dict(case='cpu',rejected=True))
    else:raise AssertionError('CPU負對照未拒絕')
    row=events[20];idx=(cap/(row['File']+'.idx')).read_bytes();pal=(cap/(row['File']+'.pal')).read_bytes()
    for kind in ('source','used-rgb'):
        b=bytearray(idx);p=bytearray(pal)
        if kind=='source':b[0]^=1
        else:p[min(set(idx[:AREA]))*3]^=1
        try:verify_source(bytes(b),bytes(p),row['SHA'])
        except AssertionError:controls.append(dict(case=kind,rejected=True))
        else:raise AssertionError('來源負對照未拒絕')
    assert not a.req.exists();a.req.mkdir(parents=True)
    request=dict(batch='b44',issue='68',scale=3,style='B',spec='010',
        scope='six canonical full intro-art frames; not an animation replacement',
        rights='local original source; existing private-review exception; not public distribution',items=[])
    selections=[]
    for label,update in SELECTIONS:
        e=next(x for x in reversed(events) if x['Update']<=update);name=e['File'];key='intro-'+label
        idx=(cap/(name+'.idx')).read_bytes();pal=(cap/(name+'.pal')).read_bytes()
        ref=idx[:AREA];one=key+'-1x.png';three=key+'-3x.png'
        im=Image.frombytes('P',(320,168),ref);im.putpalette(pal);im=im.convert('RGB')
        im.save(a.req/one);im.resize((960,504),Image.Resampling.NEAREST).save(a.req/three)
        (a.req/(key+'.idx')).write_bytes(ref);(a.req/(key+'.pal')).write_bytes(pal)
        item=dict(key=key,source_file='INTRO.BMA',source_file_sha256=inputs['original']['INTRO.BMA'],
            evidence_level='proven normal rendered frame; BMA byte mapping unknown',source_capture=name,normal_update=update,
            captured_full_indexed_sha256=sha(idx),captured_palette_sha256=sha(pal),
            source_size=[320,168],output_size=[960,504],screen_rect=[0,0,320,168],
            indexed=key+'.idx',indexed_sha256=sha(ref),palette=key+'.pal',palette_sha256=sha(pal),
            reference_1x=one,reference_1x_sha256=sha((a.req/one).read_bytes()),
            reference_3x=three,reference_3x_sha256=sha((a.req/three).read_bytes()))
        request['items'].append(item);selections.append(dict(key=key,update=update,source=name,
            indexed_sha256=sha(ref),palette_sha256=sha(pal)))
    (a.req/'manifest.json').write_text(json.dumps(request,ensure_ascii=False,indent=2)+'\n')
    proof=dict(status='source-confirmed; six HD keyframes only; runtime animation remains DRAFT',
        date='2026-10-31',issue=68,canvas=[320,200],art_rect=[0,0,320,168],frames=5100,budget=96250,
        executor='dosgolem normal SND/CONTROL session; no keys or state injection',
        address_space='rendered framebuffer [0,53760); no inferred BMA record/file offsets',
        state=normalized(timing['Digest']),prior_digest_equal=True,guest_files=127,guest_sha256=sha(json.dumps(guest,sort_keys=True).encode()),
        guest_equal=True,transitions=172,unique_source_and_used_rgb=163,unique_art_indexed=129,
        count_scope='whole capture includes SSI logo, black transitions and following TITLE; not 129 independent painted scenes',
        inputs=inputs,input_record_sha256=sha((cap/'inputs.json').read_bytes()),timing_sha256=sha((cap/'timing.json').read_bytes()),
        run_log_sha256=sha((cap/'run.log').read_bytes()),baseline_sha256=sha(a.baseline.read_bytes()),
        request_sha256=sha((a.req/'manifest.json').read_bytes()),selected=selections,sources=list(sources.values()),
        sequence=[dict(update=e['Update'],source=e['File'],signature=e['SHA']) for e in events],
        negative_controls=controls,tools=dict(inventory_sha256=sha(Path(__file__).read_bytes()),probe_sha256=inputs['source_sha256']),
        limits=['HD frames are materials only; no production changes','BMA record mapping and unselected animation phases remain unknown',
                'No freeze, interpolation or invented timer; captions outside art rect','FINAL ending not covered'])
    a.proof.write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(dict(status='PASS',states=len(sources),indexed=129,selected=6,guest_files=127,negative_controls=len(controls)),indent=2))


if __name__=='__main__':main()
