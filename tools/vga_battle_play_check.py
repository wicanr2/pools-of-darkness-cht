"""正常冷啟動DEMO的首批VGA圖層同狀態驗收。原版素材限本機。"""
import argparse
import hashlib
import json
from pathlib import Path
import re


def sha(data):
    return hashlib.sha256(data).hexdigest()


def tree(root):
    return {str(p.relative_to(root)): sha(p.read_bytes())
            for p in sorted(root.rglob('*')) if p.is_file()}


def record(path):
    inputs = {name: tree(Path('/' + name)) for name in ('orig', 'font', 'text', 'theme')}
    inputs['binary_sha256'] = sha(Path('/app/pod-play-vga').read_bytes())
    Path(path).write_text(json.dumps(inputs, ensure_ascii=False, indent=2) + '\n')


def check(root, theme):
    import numpy as np
    from PIL import Image

    cases = {}
    for name in ('hd', 'base', 'english', 'twox'):
        folder = root / name
        log = (folder / 'run.log').read_text()
        end = re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)', log)
        digest = re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)', log)
        assert end and digest, f'{name}: 缺少正式終點'
        assert int(end[1]) == 7400 and int(end[2]) == 712250000, f'{name}: 終點步數不同'
        idx = (folder / 'source.idx').read_bytes()
        pal = (folder / 'source.pal').read_bytes()
        assert len(idx) == 64000 and len(pal) == 768
        raw = np.frombuffer(pal, dtype=np.uint8).reshape(256, 3)[np.frombuffer(idx, dtype=np.uint8).reshape(200, 320)]
        frame = np.array(Image.open(folder / 'frame.png').convert('RGBA'))
        scale = 2 if name == 'twox' else 3
        assert frame.shape == (200 * scale, 480 * scale, 4)
        center = frame[:, 80 * scale:400 * scale]
        original = np.frombuffer((folder / 'original.rgba').read_bytes(), dtype=np.uint8).reshape(200 * scale, 320 * scale, 4)
        original_rgb = raw.repeat(scale, 0).repeat(scale, 1)
        assert np.array_equal(original[:, :, :3], original_rgb), f'{name}: 原版合成不符配對來源'
        assert np.all(original[:, :, 3] == 255)
        saved = tree(folder / 'save')
        frontend = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt',
                    'pod-language.txt', 'pod-speed.txt', 'pod-barfont.txt'}
        guest = {k: v for k, v in saved.items() if k not in frontend}
        cases[name] = dict(end=end.groups(), digest=digest.groups(), idx=idx, pal=pal, center=center,
                           original=original, guest=guest, inputs=json.loads((folder / 'inputs.json').read_text()),
                           frame_sha256=sha((folder / 'frame.png').read_bytes()))
    first = cases['hd']
    for name, value in cases.items():
        for key in ('end', 'digest', 'idx', 'pal', 'guest', 'inputs'):
            assert value[key] == first[key], f'{name}: {key} 改變'
    assert np.array_equal(cases['english']['center'], first['original']), 'F12仍顯示中文或HD'
    diff = np.any(first['center'] != cases['base']['center'], axis=2)
    allowed = np.zeros((600, 960), dtype=bool)
    allowed[24:528, 24:528] = True
    assert np.count_nonzero(diff) > 0 and not np.any(diff & ~allowed), 'HD差異超出戰鬥區'
    indexed = np.frombuffer(first['idx'], dtype=np.uint8).reshape(200, 320)
    matches = []
    for manifest, foreground in (('combat.json', False), ('actors.json', True)):
        data = json.loads((theme / manifest).read_text())
        count = 0
        for item in data['items']:
            w, h = item['size']
            reference = np.frombuffer((theme / item['indexed']).read_bytes(), dtype=np.uint8).reshape(h, w)
            png = np.array(Image.open(theme / item['png']).convert('RGBA'))
            assert sha((theme / item['png']).read_bytes()) == item['png_sha256']
            assert sha((theme / item['indexed']).read_bytes()) == item['indexed_sha256']
            for flip in ((False, True) if foreground else (False,)):
                ref = reference[:, ::-1] if flip else reference
                hd = png[:, ::-1] if flip else png
                ink = ref != 8 if foreground else np.ones((h, w), dtype=bool)
                for y in range(8, 177-h, 24):
                    for x in range(8, 177-w, 24):
                        if not np.array_equal(indexed[y:y+h, x:x+w][ink], ref[ink]):
                            continue
                        pixels = first['center'][y*3:(y+h)*3, x*3:(x+w)*3]
                        mask = ink.repeat(3, 0).repeat(3, 1)
                        assert np.array_equal(pixels[mask], hd[mask]), f'{item["key"]}: 遊戲內HD前景不符'
                        if foreground:
                            two = cases['twox']['center'][y*2:(y+h)*2, x*2:(x+w)*2]
                            base = cases['twox']['original'][y*2:(y+h)*2, x*2:(x+w)*2]
                            mask2 = ink.repeat(2, 0).repeat(2, 1)
                            assert np.array_equal(two[mask2], base[mask2]), '2倍仍畫HD怪物'
                        matches.append(dict(key=item['key'], x=x, y=y, flip=flip))
                        count += 1
        assert count == (8 if foreground else 5), f'{manifest}: 比對件數{count}不同'
    log = (root / 'hd/run.log').read_text()
    audit = re.search(r'pod-play: 閃現 (\d+) 次', log)
    assert audit, '缺少閃現量尺'
    # 量尺的摘要格式由pod-play固定；保留全文供證據審查。
    assert int(audit[1]) == 0, '閃現量尺未通過'
    report = dict(scope='7400格既有冷啟動DEMO；六張CPIC狀態／兩張地形；未觀測狀態保持條件式候選',
                  frames=7400, steps=712250000, memory_sha256=first['end'][2], indexed_sha256=first['end'][3],
                  cpu_sha256=first['digest'][0], palette_sha256=first['digest'][1],
                  paired_palette_sha256=sha(first['pal']), binary_sha256=first['inputs']['binary_sha256'],
                  guest_files=len(first['guest']), guest_state_equal=True, source_equal=True,
                  f12_equals_original=True, twox_actor_equals_original=True,
                  changed_hd_pixels=int(np.count_nonzero(diff)), outside_battlefield_changes=0,
                  matches=matches, audit=audit[0],
                  frames_sha256={name: v['frame_sha256'] for name, v in cases.items()},
                  inputs_sha256=sha(json.dumps(first['inputs'], sort_keys=True).encode()))
    (root / 'summary.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps(report, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--record')
    parser.add_argument('--out', type=Path)
    parser.add_argument('--theme', type=Path)
    args = parser.parse_args()
    if args.record:
        record(args.record)
    else:
        assert args.out and args.theme
        check(args.out, args.theme)
