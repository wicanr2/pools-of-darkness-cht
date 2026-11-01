#!/usr/bin/env python3
"""FINAL原版播放器來源與請求；直接入口不代表正常通關。"""
import argparse
import hashlib
import json
from pathlib import Path
import struct
import sys

import numpy as np
from PIL import Image, ImageDraw

sys.path.insert(0, str(Path(__file__).resolve().parent))
from intro_bma import decode, states, prefix


def sha(b): return hashlib.sha256(b).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for key in ('root', 'capture', 'requests', 'proof'): ap.add_argument('--'+key, type=Path, required=True)
    a = ap.parse_args(); r = a.root; original = r/'workplace/input/dos/pod'
    data = (original/'FINAL.BMA').read_bytes(); blocks = decode(data)
    assert len(blocks) == 6 and sum(len(b.frames) for b in blocks) == 58
    by_sha = {b.sha256: b.index for b in blocks}
    sequence = list(states(blocks)); source = {(f.block, f.index): (f, before, after) for f, before, after in sequence}
    timing = json.loads((a.capture/'timing.json').read_text())
    assert timing['Digest']['DOSExited'] and timing['Frames'] == 2002
    calls = timing['Images']; assert len(calls) == 58
    expected = [(f.block, f.index) for f, _, _ in sequence]
    actual = []
    for call in calls:
        assert call['SHA'] in by_sha and call['Slot'] == 2 and call['X'] == call['Y'] == call['Flags'] == 0
        block = by_sha[call['SHA']]
        assert call['Size'] == blocks[block].size and call['IP'] == 0x6733
        actual.append((block, call['Frame']))
    assert actual == expected
    records = []; palettes = {b.index: {} for b in blocks}; source_files = {}
    for event in timing['Events']:
        stem = event['File']; idx = (a.capture/(stem+'.idx')).read_bytes(); pal = (a.capture/(stem+'.pal')).read_bytes()
        assert len(idx) == 64000 and len(pal) == 768
        source_files[stem] = dict(indexed_sha256=sha(idx), palette_sha256=sha(pal))
        n = event['ImageCall']; kind = 'outside-BMA'; block = frame = None
        if 0 < n <= len(calls):
            block, frame = actual[n-1]; f, before, after = source[(block, frame)]
            top = idx[:53760]
            if top == after[:53760]: kind = 'full'
            elif prefix(f, before[:53760], top) is not None: kind = 'write-prefix'
            elif prefix(f, bytes(53760), top) is not None: kind = 'zero-write-prefix'
            else:
                cut = next((i for i, v in enumerate(top) if v), len(top))
                if cut > 0 and (top[cut:] == after[cut:53760] or top[cut:] == before[cut:53760]): kind = 'clear-prefix'
            if kind != 'outside-BMA':
                palettes[block][sha(pal)] = pal
        records.append(dict(update=event['Update'], image_call=n, block=block, frame=frame, kind=kind, source=stem,
                            art_identity_sha256=event['SHA']))
    assert all(palettes.values()), 'missing observed palettes'
    if a.requests.exists(): raise ValueError('保留舊請求，請用新目錄')
    a.requests.mkdir(parents=True); items = []; seen = {}; samples = {}; frame_keys = []
    for f, before, after in sequence:
        raw = after[:53760]; h = sha(raw)
        # 同索引依所在區塊色盤獨立辨識；不跨不同使用RGB任意合併。
        candidates = list(palettes[f.block].items())
        used = set(raw)
        ph, pal = max(candidates, key=lambda p: sum(p[1][c*3+ch] for c in used for ch in range(3)))
        identity = h + ':' + sha(bytes(pal[c*3+ch] for c in sorted(used) for ch in range(3)))
        if identity in seen:
            frame_keys.append(dict(block=f.block, frame=f.index, key=seen[identity])); continue
        key = 'final-b%02d-f%02d' % (f.block, f.index); seen[identity] = key
        frame_keys.append(dict(block=f.block, frame=f.index, key=key))
        p = Image.frombytes('P', (320, 168), raw); p.putpalette(pal); p = p.convert('RGB')
        p.save(a.requests/(key+'-1x.png')); p.resize((960, 504), Image.Resampling.NEAREST).save(a.requests/(key+'-3x.png'))
        (a.requests/(key+'.idx')).write_bytes(raw); (a.requests/(key+'.pal')).write_bytes(pal)
        items.append(dict(key=key, block=f.block, frame=f.index, indexed=key+'.idx', palette=key+'.pal',
            source_indexed_sha256=h, palette_sha256=ph, source_file_offset=f.file_offset,
            reference=key+'-3x.png', width=960, height=504, style='B', spec='010/b48',
            evidence='confirmed decoded source; palette observed in direct VACE FINAL playback'))
        samples[(f.block, f.index)] = p
    contact = Image.new('RGB', (5*320, ((len(items)+4)//5)*188)); draw = ImageDraw.Draw(contact)
    for n, item in enumerate(items):
        x = n%5*320; y = n//5*188; contact.paste(samples[(item['block'], item['frame'])], (x, y)); draw.text((x+2, y+169), item['key'], fill='white')
    contact.save(a.requests/'contact.png')
    inputs = {n: sha((original/n).read_bytes()) for n in ('FINAL.BMA', 'FINAL.ACF', 'VACE.EXE', 'CONTROL.EXE', 'VGA.DRV', 'P3.SMS', 'ADLIB.COM')}
    proof = dict(issue=77, status='SOURCE PASS; HD not implemented', scope='original VACE direct player, not a completed campaign',
        entry=dict(root='ADLIB.COM', virtual_path=r'C:\POD\SND.COM', queue='VACE.EXE', args='4 FINAL 0 1',
                   rationale='VACE built-in command usage; CONTROL contains FINAL 1 command; direct player selects video mode itself'),
        input_sha256=inputs, blocks=[dict(index=b.index, offset=b.file_offset, size=b.size, count=len(b.frames), sha256=b.sha256) for b in blocks],
        calls=calls, events=records, source_files=source_files, palettes={str(b): list(p) for b, p in palettes.items()},
        timing=dict(frames=timing['Frames'], steps=timing['Digest']['Steps'], digest=timing['Digest'],
            sha256=sha((a.capture/'timing.json').read_bytes())), probe_sha256=sha((a.capture.parent/'final-bma-probe').read_bytes()),
        requests=items, frames=frame_keys, codec='tools/intro_bma.py, verified FINAL6 blocks58 images',
        limits=['direct original player only; normal post-victory CONTROL entry not replayed',
            'captured frames are end-of-update observations, not every video write', 'artwork and runtime presentation not yet accepted'])
    a.proof.write_text(json.dumps(proof, ensure_ascii=False, indent=2)+'\n')
    (a.requests/'manifest.json').write_text(json.dumps(dict(issue=77, input_sha256=inputs, items=items, frames=frame_keys), ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(dict(blocks=6, calls=len(calls), unique_requests=len(items), events=len(records),
        classified={k: sum(x['kind'] == k for x in records) for k in {x['kind'] for x in records}}), ensure_ascii=False))


if __name__ == '__main__': main()
