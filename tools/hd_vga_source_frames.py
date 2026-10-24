"""以正常前端逐張配對的idx/pal核對TITLE與BIGPIC完整來源。"""
import argparse
import hashlib
import json
from pathlib import Path
import numpy as np


def sha(b):
    return hashlib.sha256(b).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for name in ('inventory', 'capture', 'out'):
        ap.add_argument('--' + name, type=Path, required=True)
    a = ap.parse_args()
    catalog = json.loads((a.inventory / 'inventory.json').read_text())
    references = []
    for it in catalog['items']:
        if it['file'] not in ('TITLE.DAX', 'BIGPIC1.DAX'):
            continue
        idx = (a.inventory / (it['key'] + '.idx')).read_bytes()
        pal = (a.inventory / (it['key'] + '.pal')).read_bytes()
        assert sha(idx) == it['indexed_sha256'] and sha(pal) == it['palette_sha256']
        references.append((it, np.frombuffer(idx, np.uint8).reshape(it['height'], it['width']),
                           np.frombuffer(pal, np.uint8).reshape(256, 3)))
    rows = []
    for p in sorted(a.capture.glob('*.idx')):
        raw, rawpal = p.read_bytes(), p.with_suffix('.pal').read_bytes()
        assert len(raw) == 64000 and len(rawpal) == 768
        screen = np.frombuffer(raw, np.uint8).reshape(200, 320)
        pal = np.frombuffer(rawpal, np.uint8).reshape(256, 3)
        for it, ref, colors in references:
            h, w = ref.shape
            used = np.unique(ref)
            for y in range(0, 201-h, 8):
                for x in range(0, 321-w, 8):
                    if np.array_equal(screen[y:y+h, x:x+w], ref):
                        rows.append(dict(capture=p.stem, key=it['key'], rect=[x, y, w, h],
                            indexed_equal=True, used_rgb_equal=bool(np.array_equal(colors[used], pal[used])),
                            indexed_sha256=sha(raw), palette_sha256=sha(rawpal),
                            source_indexed_sha256=it['indexed_sha256'],
                            archive_sha256=it['archive_sha256']))
    result = dict(scope='正常冷啟動DEMO；F10在本次推進前，檔名格數減一為配對來源格', hits=rows)
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    main()
