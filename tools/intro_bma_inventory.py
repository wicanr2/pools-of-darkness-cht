#!/usr/bin/env python3
"""核對正常BMA呼叫／原版狀態，產生可重建來源索引與美術請求；不接入產品。"""
import argparse
import copy
import json
import re
from pathlib import Path
import sys

from PIL import Image

sys.path.insert(0, str(Path(__file__).resolve().parent))
from intro_bma import decode, digest, prefix, states

AREA = 53760


def read(path):
    return json.loads(path.read_text())


def sha(path):
    return digest(path.read_bytes())


def signature(idx, pal):
    if len(idx) != 64000 or len(pal) != 768:
        raise ValueError('索引或色盤長度不同')
    return digest(idx[:AREA] + b''.join(bytes([c])+pal[c*3:c*3+3] for c in sorted(set(idx[:AREA]))))


def verify_calls(images, blocks):
    table = {b.sha256: b for b in blocks}
    wanted = [(b.index, f.index) for b in blocks for f in b.frames]
    got = []
    for call in images:
        if call['SHA'] not in table:
            raise ValueError('未知BMA區塊')
        block = table[call['SHA']]
        if call['Size'] != block.size or not 0 <= call['Frame'] < len(block.frames):
            raise ValueError('BMA來源範圍不同')
        if (call['CS'], call['IP'], call['X'], call['Y'], call['Flags']) != (0x48a1, 0x6733, 0, 0, 0):
            raise ValueError('正常來源入口、座標或方向不同')
        got.append((block.index, call['Frame']))
    if got != wanted:
        raise ValueError('圖面呼叫次序或數量不同')
    return table


def classify(frame, before, after, target):
    if target == after[:AREA]:
        return dict(kind='complete')
    bases = [('previous', before[:AREA])]
    if frame.index == 0 and before[:AREA] != bytes(AREA):
        bases.append(('clear-0', bytes(AREA)))
    for name, base in bases:
        count = prefix(frame, base, target)
        if count is not None:
            return dict(kind='draw-prefix', base=name, pixels=count)
    # 已知完整來源先被線性清為0；仍須核對所有未清除像素。
    changed = [i for i, (a, b) in enumerate(zip(before[:AREA], target)) if a != b]
    if changed:
        cut = changed[-1] + 1
        if target == bytes(cut) + before[cut:AREA]:
            return dict(kind='clear-prefix', color=0, pixels=cut)
    raise ValueError(f'未辨識來源：block {frame.block} frame {frame.index}')


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for name in ('root', 'raw', 'fast', 'probe', 'out', 'proof'):
        ap.add_argument('--'+name, type=Path, required=True)
    ap.add_argument('--engine', required=True)
    ap.add_argument('--seed', type=Path)
    args = ap.parse_args()
    if not re.fullmatch(r'[0-9a-f]{40}', args.engine):
        raise ValueError('需完整引擎commit')
    root, out = args.root, args.out
    if out.exists():
        raise ValueError('保留舊請求，輸出須為新目錄')
    baseline_path = root/'docs/re/018-b44-source-proof.json'
    baseline = read(baseline_path)
    original = root/'workplace/input/dos/pod'
    tree = lambda p: {str(f.relative_to(p)): sha(f) for f in sorted(p.rglob('*')) if f.is_file()}
    inputs = tree(original)
    if inputs != baseline['inputs']['original']:
        raise ValueError('自備原版集合或SHA不同')
    seed = args.seed or root/'workplace/probe/g2-receipt/scratch'
    if {n: sha(seed/n) for n in ('pool4.cfg', 'snd.com')} != baseline['inputs']['seed']:
        raise ValueError('啟動設定或驅動不同')
    raw, fast = read(args.raw/'timing.json'), read(args.fast/'timing.json')
    assert raw['Frames'] == fast['Frames'] == 5100 and raw['Budget'] == fast['Budget'] == 96250
    assert raw['GoDraw'] is False and fast['GoDraw'] is True
    assert raw['HostPlans'] == 0 and fast['HostPlans'] > 0 and fast['HostInstructions'] > 0
    assert raw['CommittedPlans'] == raw['CommittedInstructions'] == 0
    assert fast['CommittedPlans'] == fast['HostPlans']
    assert fast['CommittedInstructions'] == fast['HostInstructions']
    normalized = {k: bytes(v).hex() if isinstance(v, list) else v for k, v in raw['Digest'].items()}
    assert raw['Digest'] == fast['Digest'] and normalized == baseline['state']
    clean = lambda image: {k: v for k, v in image.items() if k != 'Path'}
    assert [clean(x) for x in raw['Images']] == [clean(x) for x in fast['Images']]
    identity = lambda timing: [(e['Update'], e['File'], e['SHA'], e['ImageCall']) for e in timing['Events']]
    assert identity(raw) == identity(fast)
    assert [(e['Update'], e['File'], e['SHA']) for e in raw['Events']] == [(e['update'], e['source'], e['signature']) for e in baseline['sequence']]
    guest = tree(args.raw/'save')
    assert guest == tree(args.fast/'save') and len(guest) == 127
    assert digest(json.dumps(guest, sort_keys=True).encode()) == baseline['guest_sha256']
    blocks = decode((original/'INTRO.BMA').read_bytes())
    table = verify_calls(raw['Images'], blocks)
    verify_calls(fast['Images'], blocks)
    materialized = {(f.block, f.index): (f, before, after) for f, before, after in states(blocks)}
    snapshots, events = {}, []
    for event in raw['Events']:
        stem = event['File']
        idx, pal = (args.raw/(stem+'.idx')).read_bytes(), (args.raw/(stem+'.pal')).read_bytes()
        assert signature(idx, pal) == event['SHA']
        assert sha(args.raw/(stem+'.idx')) == sha(args.fast/(stem+'.idx'))
        assert sha(args.raw/(stem+'.pal')) == sha(args.fast/(stem+'.pal'))
        snapshots[stem] = (idx, pal)
        row = dict(update=event['Update'], source=stem, identity_sha256=event['SHA'], image_call=event['ImageCall'])
        if len(set(idx[:AREA])) == 1:
            row.update(kind='uniform', color=idx[0])
        elif stem == 'source-162':
            row.update(kind='outside-BMA-scope')  # 018固定來源的後續TITLE，不能冒充BMA。
        else:
            call = raw['Images'][event['ImageCall']-1]
            block = table[call['SHA']]
            frame, before, after = materialized[block.index, call['Frame']]
            row.update(block=block.index, frame=frame.index, **classify(frame, before, after, idx[:AREA]))
        events.append(row)
    assert len(snapshots) == 163 and len(events) == 172
    marker = 'BMA來源已由031對應'
    backlinks = []
    for path in ('docs/re/018-movie-subtitles.md', 'docs/spec/010-hd-theme.md', 'CONTEXT.md'):
        content = (root/path).read_text()
        if marker not in content or '031-intro-bma.md' not in content:
            raise ValueError('舊來源未知未回填：'+path)
        backlinks.append(dict(older_spec=path, required_correction=marker))
    entry_hash = digest((original/'VGA.DRV').read_bytes()[0x6733:0x6832])
    assert entry_hash == 'd9ebc163f106be514229472090f2497771df8e36ab5aeca61681f758d32baa54'
    approved = read(root/'review/hd/b44-intro/manifest.json')
    approved_by_idx = {item['source']['indexed_sha256']: item for item in approved['items']}
    for item in approved['items']:
        h = sha(root/'review/hd/b44-intro'/item['preparation']['png'])
        assert h == item['preparation']['png_sha256'] == item['visual']['png_sha256']
        assert item['visual']['status'] == 'accepted-after-viewing'
    out.mkdir(parents=True)
    req, images, unique = [], [], {}
    for block in blocks:
        for frame in block.frames:
            _, _, after = materialized[block.index, frame.index]
            art = after[:AREA]
            h = digest(art)
            row = dict(block=block.index, frame=frame.index, file_offset=frame.file_offset,
                       size=frame.size, spans=len(frame.spans), art_indexed_sha256=h)
            images.append(row)
            if h in unique:
                row['same_as'] = unique[h]
                continue
            key = f'intro-b{block.index:02}-f{frame.index:02}'
            unique[h] = key
            matches = [(n, i, p) for n, (i, p) in snapshots.items() if digest(i[:AREA]) == h]
            assert matches, ('沒有正常完整來源', key)
            if h in approved_by_idx:
                selected = approved_by_idx[h]['source']['source_capture']
                name, idx, pal = next(v for v in matches if v[0] == selected)
            else:
                # 僅選預覽參考；全部實際色盤變體另記，不把此選擇當播放契約。
                name, idx, pal = max(matches, key=lambda value: sum(value[2][c*3+k] for c in set(art) for k in range(3)))
            (out/(key+'.idx')).write_bytes(art)
            (out/(key+'.pal')).write_bytes(pal)
            im = Image.frombytes('P', (320, 168), art)
            im.putpalette(pal)
            im = im.convert('RGB')
            im.save(out/(key+'-1x.png'))
            im.resize((960, 504), Image.Resampling.NEAREST).save(out/(key+'-3x.png'))
            item = dict(key=key, **row, source_capture=name, source_size=[320, 168], output_size=[960, 504],
                        source_file='INTRO.BMA', source_file_sha256=inputs['INTRO.BMA'],
                        reference_1x=key+'-1x.png', reference_3x=key+'-3x.png',
                        indexed=key+'.idx', palette=key+'.pal', palette_sha256=digest(pal),
                        observed_palette_variants=[dict(source=n, sha256=digest(p)) for n, _, p in matches],
                        style='B', state='needs-HD-frame-or-proven-layer-composition',
                        prompt='依原版3倍參考及已接受的同場景HD重畫。保留全部構圖、角色姿勢、動畫相位、石框、原配色與黑邊；960×504不透明RGBA。下方字幕由既有系統處理，不畫文字，不用濾鏡放大。')
            if h in approved_by_idx:
                old = approved_by_idx[h]
                item.update(state='accepted-b44-material', approved_key=old['key'],
                            approved_png='review/hd/b44-intro/'+old['preparation']['png'],
                            approved_sha256=old['preparation']['png_sha256'])
            item['reference_1x_sha256'] = sha(out/item['reference_1x'])
            item['reference_3x_sha256'] = sha(out/item['reference_3x'])
            req.append(item)
    request = dict(issue=74, status='source requests only; no production theme',
                   style='B', count=len(req), accepted=sum(x['state']=='accepted-b44-material' for x in req),
                   rights='local original inputs; private evidence; approved HD scope unchanged', items=req)
    (out/'manifest.json').write_text(json.dumps(request, ensure_ascii=False, indent=2)+'\n')
    negatives = []
    for label in ('block-sha', 'frame', 'order', 'direction'):
        bad = copy.deepcopy(raw['Images'])
        if label == 'block-sha': bad[0]['SHA'] = '0'*64
        elif label == 'frame': bad[0]['Frame'] = 1
        elif label == 'order': bad[1], bad[2] = bad[2], bad[1]
        else: bad[0]['Flags'] = 1
        try: verify_calls(bad, blocks)
        except ValueError: negatives.append(dict(case=label, rejected=True))
        else: raise ValueError('來源負對照未拒絕')
    sample = raw['Events'][20]
    idx, pal = snapshots[sample['File']]
    for label in ('indexed', 'used-rgb'):
        b, p = bytearray(idx), bytearray(pal)
        if label == 'indexed': b[0] ^= 1
        else: p[min(set(idx[:AREA]))*3] ^= 1
        assert signature(bytes(b), bytes(p)) != sample['SHA']
        negatives.append(dict(case=label, rejected=True))
    proof = dict(date='2026-11-01', issue=74, status='source mapping PASS; runtime HD remains DRAFT',
                 engine=args.engine, input_sha256={n: inputs[n] for n in ('INTRO.BMA', 'INTRO.ACF', 'VACE.EXE', 'VGA.DRV')},
                 original_collection_equal=True, baseline_sha256=sha(baseline_path), seed=baseline['inputs']['seed'],
                 probe_sha256=sha(args.probe), probe_source_sha256=sha(Path(__file__).with_name('movie_bma_probe.go')),
                 address_spaces=dict(BMA='absolute byte offsets in INTRO.BMA', framebuffer='y*320+x; top53760 bytes',
                                     runtime='CS:IP48A1:6733; source head is separate linear RAM address'),
                 image_entry=dict(file='VGA.DRV', file_offset=0x6733, length=255, sha256=entry_hash,
                                  runtime_cs=0x48a1, runtime_ip=0x6733, evidence_level='proven'),
                 resolutions=[dict(platform_module='DOS/INTRO.BMA@'+inputs['INTRO.BMA'], address=0,
                                   semantic='14 consecutive blocks and133 frame calls;98 unique complete top-art indices',
                                   evidence_level='proven', evidence_spec='docs/re/031-intro-bma.md', backlinks=backlinks)],
                 frames=5100, budget=96250, digest=normalized, guest_files=127, guest_equal=True,
                 guest_sha256=baseline['guest_sha256'], blocks=[dict(index=b.index, file_offset=b.file_offset, size=b.size, sha256=b.sha256, count=len(b.frames)) for b in blocks],
                 images=images, calls=raw['Images'], events=events, source_states=163, art_indexed_states=129,
                 raw_and_go_same_calls=True, raw_and_go_same_events=True, raw_and_go_same_sources=True,
                 go_drawing=dict(plans=fast['HostPlans'], instructions=fast['HostInstructions'], entry_visits=fast['HostEntryVisits'],
                                 committed_plans=fast['CommittedPlans'], committed_instructions=fast['CommittedInstructions']),
                 request=dict(count=len(req), accepted_b44=request['accepted'], pending=len(req)-request['accepted'],
                              manifest_sha256=sha(out/'manifest.json')),
                 negatives=negatives, tools={n: sha(Path(__file__).with_name(n)) for n in
                     ('intro_bma.py', 'intro_bma_inventory.py', 'intro_bma_test.py', 'movie_bma_probe.go', 'intro-bma-receipt.sh')},
                 limits=['research and art requests only; production binary/theme unchanged',
                         'frame-call order proven only for this normal INTRO; not a generic ACF player',
                         'palette variants and partial drawing phases need runtime rendering contracts',
                         'six accepted paintings do not cover remaining phases; FINAL not included'])
    args.proof.write_text(json.dumps(proof, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(dict(blocks=len(blocks), images=len(images), unique=len(req), accepted=request['accepted'],
                          pending=proof['request']['pending'], events=len(events), negatives=len(negatives), go_plans=fast['HostPlans'])))


if __name__ == '__main__':
    main()
