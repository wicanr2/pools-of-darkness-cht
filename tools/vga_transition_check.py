"""辨識首次戰鬥HD閃現的逐格來源；F10發生於本次推進之前，來源格為檔名格數減一。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import numpy as np


def sha(b):
    return hashlib.sha256(b).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for name in ('theme', 'capture', 'failed', 'out'):
        ap.add_argument('--' + name, type=Path, required=True)
    a = ap.parse_args()
    templates = []
    masked_templates = []
    for manifest in ('actors.json', 'combat.json'):
        for it in json.loads((a.theme / manifest).read_text())['items']:
            w, h = it['size']
            reference = np.frombuffer((a.theme / it['indexed']).read_bytes(), np.uint8).reshape(h, w)
            for flip in ((False, True) if manifest == 'actors.json' else (False,)):
                ref = reference[:, ::-1] if flip else reference
                for oy in range(8, 177-h, 24):
                    for ox in range(8, 177-w, 24):
                        for y in range(0, h, 8):
                            for x in range(0, w, 8):
                                block = ref[y:y+8, x:x+8]
                                if np.unique(block).size > 1:
                                    templates.append(((ox+x)//8, (oy+y)//8, sha(block.tobytes()), it['key']))
                                    if manifest == 'actors.json':
                                        masked_templates.append(((ox+x)//8, (oy+y)//8, block, it['key']))
    lookup = {}
    for x, y, digest, key in templates:
        lookup.setdefault((x, y, digest), set()).add(key)
    failed = a.failed.read_text()
    rows = []
    for shown, replaced, cy, columns in re.findall(r'閃現 第 (\d+) 格原樣呈現、第 (\d+) 格被換 列 (\d+) 欄([^\n]+)', failed):
        shown, replaced, cy = int(shown), int(replaced), int(cy)
        source = a.capture / f'pod-f{shown+1:06d}.idx'
        after = a.capture / f'pod-f{replaced+1:06d}.idx'
        idx = np.frombuffer(source.read_bytes(), np.uint8).reshape(200, 320)
        later = np.frombuffer(after.read_bytes(), np.uint8).reshape(200, 320)
        pal = source.with_suffix('.pal').read_bytes()
        pal2 = after.with_suffix('.pal').read_bytes()
        for cx in map(int, columns.split()):
            cell = idx[cy*8:cy*8+8, cx*8:cx*8+8]
            matched = sorted(lookup.get((cx, cy, sha(cell.tobytes())), ()))
            masked = []
            if not matched:
                for tx, ty, template, key in masked_templates:
                    ink = template != 8
                    if tx == cx and ty == cy and ink.any() and np.array_equal(cell[ink], template[ink]):
                        masked.append(dict(key=key, foreground=int(ink.sum()), colors=int(np.unique(template[ink]).size)))
            rows.append(dict(shown=shown, replaced=replaced, cx=cx, cy=cy,
                             source_cell_same=bool(np.array_equal(cell, later[cy*8:cy*8+8, cx*8:cx*8+8])),
                             palette_same=pal == pal2, exact_source_cells=matched, masked_actor_cells=masked))
    frames = []
    for p in sorted(a.capture.glob('*.idx')):
        number = int(p.stem.split('f')[-1]) - 1
        if number < 6400:
            continue
        data = np.frombuffer(p.read_bytes(), np.uint8).reshape(200, 320)
        pal = p.with_suffix('.pal').read_bytes()
        count = 0
        for it in json.loads((a.theme / 'combat.json').read_text())['items']:
            ref = np.frombuffer((a.theme / it['indexed']).read_bytes(), np.uint8).reshape(24, 24)
            count += sum(np.array_equal(data[y:y+24, x:x+24], ref)
                         for y in range(8, 153, 24) for x in range(8, 153, 24))
        frames.append(dict(frame=number, terrain_full_matches=count,
                           indexed_sha256=sha(p.read_bytes()), palette_sha256=sha(pal),
                           cpic_bank_sha256=sha(pal[144*3:256*3]), terrain_bank_sha256=sha(pal[32*3:64*3])))
    report = dict(f10_source_offset=-1, events=len(rows),
                  exact_source_cell_events=sum(bool(r['exact_source_cells']) for r in rows),
                  frames=frames, cells=rows)
    a.out.write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps({k: v for k, v in report.items() if k not in ('frames', 'cells')}, ensure_ascii=False))


if __name__ == '__main__':
    main()
