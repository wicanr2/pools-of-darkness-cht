#!/usr/bin/env python3
"""對正常開場採樣的字幕區去重；只產生來源證據，不接入顯示。"""
import argparse
import hashlib
import json
import math
from pathlib import Path

from PIL import Image, ImageDraw


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--capture', type=Path, required=True)
    ap.add_argument('--proof', type=Path, required=True)
    a = ap.parse_args()
    groups, images = {}, {}
    files = sorted((a.capture/'shots').glob('*.raw'))
    assert len(files) == 59, '此收據需要59個固定採樣'
    for p in files:
        raw, pal = p.read_bytes(), p.with_suffix('.pal').read_bytes()
        assert len(raw) == 64000 and len(pal) == 768
        im = Image.frombytes('P', (320,200), raw)
        im.putpalette(pal)
        im = im.convert('RGB')
        key = hashlib.sha256(im.crop((0,168,320,200)).tobytes()).hexdigest()
        step = int(p.stem[1:])
        g = groups.setdefault(key, dict(first_step=step, last_step=step, frames=[], source=str(p),
            idx_sha256=sha(p), pal_sha256=sha(p.with_suffix('.pal'))))
        g['last_step'] = step
        g['frames'].append(step)
        images.setdefault(key, im)
    rows = list(groups.items())
    sheet = Image.new('RGB', (1280, math.ceil(len(rows)/2)*100), (35,35,35))
    draw = ImageDraw.Draw(sheet)
    for n, (key, row) in enumerate(rows):
        x, y = n%2*640, n//2*100
        draw.text((x+3,y+3), f"{n}: {row['first_step']}..{row['last_step']}", fill='white')
        sheet.paste(images[key].crop((0,168,320,200)).resize((640,64), Image.Resampling.NEAREST), (x,y+25))
        images[key].save(a.capture/f'group-{n:02d}.png')
    sheet.save(a.capture/'subtitle-contact.png')
    (a.capture/'subtitle-groups.json').write_text(json.dumps([dict(crop_rgb_sha256=k, **r) for k,r in rows], indent=2)+'\n')
    draft = json.loads((a.capture/'subtitle-draft.json').read_text())
    cues = []
    assert [it['group'] for it in draft['items']] == list(range(2,10))
    for it in draft['items']:
        key, row = rows[it['group']]
        cues.append(dict(key=it['key'], group=it['group'], first_sample_step=row['first_step'],
            last_sample_step=row['last_step'], source=row['source'], indexed_sha256=row['idx_sha256'],
            palette_sha256=row['pal_sha256'], crop_rgb_sha256=key,
            english_lines_sha256=hashlib.sha256('\n'.join(it['english_lines']).encode()).hexdigest(),
            english_line_lengths=[len(s) for s in it['english_lines']],
            draft_translation_sha256=hashlib.sha256(it['translation'].encode()).hexdigest()))
    log = (a.capture/'run.log').read_text()
    assert 'VACE.EXE       PSP=0320 exit=0' in log and 'GAME.EXE       PSP=169F' in log
    report = dict(status='source-only; rendering remains DRAFT', executor='dosgolem probe',
        input_record=json.loads((a.capture/'inputs.json').read_text()), input_record_sha256=sha(a.capture/'inputs.json'),
        run_log_sha256=sha(a.capture/'run.log'), all_raw_samples=59, crop_rect=[0,168,320,32],
        cue_count=8, vace_normal_exit=True, game_started=True,
        seed_note='原版亂數未改；片頭內容不依賴GAME亂數，後續DEMO不列本次字幕證據',
        evidence_level='已證實採樣畫格；首次繪製與清除時序未知', cues=cues)
    a.proof.write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n')
    print('八組片頭字幕來源及原版正常結束已固定；不代表正式接入。')


if __name__ == '__main__':
    main()
