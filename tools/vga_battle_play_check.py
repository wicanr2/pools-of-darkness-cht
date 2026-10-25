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


def record(path, binary='/app/pod-play-vga'):
    inputs = {name: tree(Path('/' + name)) for name in ('orig', 'font', 'text', 'theme')}
    inputs['binary_sha256'] = sha(Path(binary).read_bytes())
    Path(path).write_text(json.dumps(inputs, ensure_ascii=False, indent=2) + '\n')


def check(root, theme, original_baseline=None):
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
                    'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
        guest = {k: v for k, v in saved.items() if k not in frontend}
        cases[name] = dict(end=end.groups(), digest=digest.groups(), idx=idx, pal=pal, center=center,
                           original=original, guest=guest, inputs=json.loads((folder / 'inputs.json').read_text()),
                           frame_sha256=sha((folder / 'frame.png').read_bytes()))
    first = cases['hd']
    for name, value in cases.items():
        for key in ('end', 'digest', 'idx', 'pal', 'guest', 'inputs'):
            assert value[key] == first[key], f'{name}: {key} 改變'
    assert tree(theme) == first['inputs']['theme'], '主題已在執行後變更'
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
    actors = json.loads((theme / 'actors.json').read_text())['items']
    observed = {r['key'] for r in matches if r['key'].startswith('cpic1-')}
    report = dict(scope='7400格既有冷啟動DEMO；八個CPIC鏡像位置／兩張地形；未觀測狀態保持條件式候選',
                  registered_actor_states=len(actors),
                  actors_manifest_sha256=sha((theme / 'actors.json').read_bytes()), theme_files=len(first['inputs']['theme']),
                  unobserved_actor_keys=[r['key'] for r in actors if r['key'] not in observed],
                  frames=7400, steps=712250000, memory_sha256=first['end'][2], indexed_sha256=first['end'][3],
                  cpu_sha256=first['digest'][0], palette_sha256=first['digest'][1],
                  paired_palette_sha256=sha(first['pal']), binary_sha256=first['inputs']['binary_sha256'],
                  guest_files=len(first['guest']), guest_state_equal=True, source_equal=True,
                  f12_equals_original=True, twox_actor_equals_original=True,
                  changed_hd_pixels=int(np.count_nonzero(diff)), outside_battlefield_changes=0,
                  matches=matches, audit=audit[0],
                  frames_sha256={name: v['frame_sha256'] for name, v in cases.items()},
                  inputs_sha256=sha(json.dumps(first['inputs'], sort_keys=True).encode()))
    if original_baseline:
        previous = original_baseline / 'hd'
        old_log = (previous / 'run.log').read_text()
        for pattern in (r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',
                        r'cpu_sha256=(\w+) palette_sha256=(\w+)'):
            actual, expected = (re.search(pattern, value) for value in (log, old_log))
            assert actual and expected and actual.groups() == expected.groups(), '新執行檔的原版終點不同'
        for name in ('source.idx', 'source.pal', 'original.rgba'):
            assert (root / 'hd' / name).read_bytes() == (previous / name).read_bytes(), f'新執行檔的{name}不同'
        old_guest = {k: v for k, v in tree(previous / 'save').items() if k not in frontend}
        assert first['guest'] == old_guest, '新執行檔的DOS檔案不同'
        old_inputs = json.loads((previous / 'inputs.json').read_text())
        for key in ('orig', 'font', 'text'):
            assert first['inputs'][key] == old_inputs[key], f'{key}與原版終點基準不同'
        report.update(prior_original_state_equal=True,
                      prior_baseline_binary_sha256=old_inputs['binary_sha256'],
                      prior_baseline_summary_sha256=sha((original_baseline / 'summary.json').read_bytes()),
                      prior_sample_composite_equal=(root / 'hd/frame.png').read_bytes() == (previous / 'frame.png').read_bytes())
    (root / 'summary.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps(report, ensure_ascii=False, indent=2))


def check_incremental(root, theme, baseline, baseline_theme, added):
    """只增未觀測素材時，比較新的正常HD路徑與已通過的四路基準。"""
    assert added and len(set(added)) == len(added)
    current, previous = root / 'hd', baseline / 'hd'
    inputs = json.loads((current / 'inputs.json').read_text())
    old_inputs = json.loads((previous / 'inputs.json').read_text())
    for key in ('orig', 'font', 'text', 'binary_sha256'):
        assert inputs[key] == old_inputs[key], f'{key}與固定基準不同'
    manifests = []
    for directory, recorded in ((baseline_theme, old_inputs), (theme, inputs)):
        path = directory / 'actors.json'
        assert sha(path.read_bytes()) == recorded['theme']['actors.json']
        manifests.append(json.loads(path.read_text()))
        assert tree(directory) == recorded['theme'], '主題已在執行後變更'
    old, new = manifests
    assert {k: v for k, v in old.items() if k != 'items'} == {k: v for k, v in new.items() if k != 'items'}
    old_rows, new_rows = ({r['key']: r for r in m['items']} for m in manifests)
    assert len(old_rows) == len(old['items']) and len(new_rows) == len(new['items'])
    assert set(new_rows) - set(old_rows) == set(added)
    assert all(new_rows.get(k) == row for k, row in old_rows.items()), '舊條目被覆寫'
    allowed = {'actors.json', 'vga-layers-provenance.json'}
    for key in added:
        row = new_rows[key]
        allowed.update((row['png'], row['indexed']))
        for path, digest in ((row['png'], row['png_sha256']), (row['indexed'], row['indexed_sha256'])):
            assert inputs['theme'][path] == digest
    for path, digest in old_inputs['theme'].items():
        assert path in inputs['theme'] and (path in allowed or inputs['theme'][path] == digest), f'{path}被刪除或變更'
    assert set(inputs['theme']) - set(old_inputs['theme']) == allowed - {'actors.json', 'vga-layers-provenance.json'}
    log, old_log = ((d / 'run.log').read_text() for d in (current, previous))
    patterns = (r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',
                r'cpu_sha256=(\w+) palette_sha256=(\w+)')
    for pattern in patterns:
        actual, expected = (re.search(pattern, s) for s in (log, old_log))
        assert actual and expected and actual.groups() == expected.groups(), '原版終點不符基準'
    for name in ('source.idx', 'source.pal', 'original.rgba', 'frame.png'):
        assert (current / name).read_bytes() == (previous / name).read_bytes(), f'{name}不符基準'
    frontend = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt',
                'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
    guest = [{k: v for k, v in tree(d / 'save').items() if k not in frontend} for d in (current, previous)]
    assert guest[0] == guest[1]
    assert re.search(r'pod-play: 閃現 (\d+) 次', log)[1] == '0'
    report = dict(scope='既有四路基準加未觀測資料；重跑7400格HD正常路徑；逐件2倍/F12與失配另由來源診斷驗證',
                  added_actor_keys=added, registered_actor_states=len(new_rows),
                  binary_sha256=inputs['binary_sha256'], guest_files=len(guest[0]),
                  guest_state_equal=True, original_endpoint_equal=True, final_composite_equal=True,
                  prior_assets_preserved=True, flash_count=0,
                  frame_sha256=sha((current / 'frame.png').read_bytes()),
                  baseline_summary_sha256=sha((baseline / 'summary.json').read_bytes()),
                  inputs_sha256=sha((current / 'inputs.json').read_bytes()))
    (root / 'summary.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps(report, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--record')
    parser.add_argument('--record-binary', default='/app/pod-play-vga')
    parser.add_argument('--out', type=Path)
    parser.add_argument('--theme', type=Path)
    parser.add_argument('--baseline', type=Path)
    parser.add_argument('--baseline-theme', type=Path)
    parser.add_argument('--original-baseline', type=Path, help='四路完整驗收後，另比先前執行檔的原版終點；不取代四路')
    parser.add_argument('--added', nargs='+')
    args = parser.parse_args()
    if args.record:
        record(args.record, args.record_binary)
    else:
        assert args.out and args.theme
        assert not (args.baseline and args.original_baseline), '純資料增量與新執行檔四路驗收不能混用'
        if args.baseline:
            assert args.baseline_theme
            check_incremental(args.out, args.theme, args.baseline, args.baseline_theme, args.added)
        else:
            check(args.out, args.theme, args.original_baseline)
