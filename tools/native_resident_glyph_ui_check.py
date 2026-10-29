#!/usr/bin/env python3
"""核對原生16／24像素英數字；所有截圖差異須落在已認證字格內。"""
import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import re
import tempfile

import numpy as np
from PIL import Image


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    return json.loads(path.read_text())


def tree(path):
    return {str(p.relative_to(path)): sha(p) for p in sorted(path.rglob('*')) if p.is_file()}


def load(path):
    spec = importlib.util.spec_from_file_location(path.stem, path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def audit(folder, font_sample):
    text = (folder/'run.log').read_text()
    count = int(re.search(r'pod-play: 閃現 (\d+) 次', text)[1])
    if count == 0:
        return dict(raw_flash_count=0, intentional_cells=0, unexpected_flash_count=0)
    assert font_sample, (folder, count)
    events = re.findall(r'pod-play: 閃現 第 (\d+) 格原樣呈現、第 (\d+) 格被換 列 (\d+) 欄([^\n]+)', text)
    assert events
    # These are the actual scripted language, scale and original-view edges.
    # F12 disables the audit, so 4788 is the preceding observation at 4792.
    preceding = {4783: 4782, 4786: 4785, 4792: 4788, 4793: 4792, 4794: 4793}
    assert all(preceding.get(int(changed)) == int(shown) for shown, changed, _, _ in events), events
    measured = sum(len(cols.split()) for _, _, _, cols in events)
    assert measured == count
    return dict(raw_flash_count=count, intentional_cells=measured, unexpected_flash_count=0,
                events=[dict(shown=int(a), changed=int(b), row=int(c), columns=[int(x) for x in d.split()])
                        for a, b, c, d in events])


def compare_image(current, previous, source, report_path):
    a, b = [np.array(Image.open(path).convert('RGBA')) for path in (current, previous)]
    assert a.shape == b.shape and a.shape[0] in (400, 600), (current, a.shape, b.shape)
    scale = a.shape[0]//200
    width = 320*scale
    assert a.shape[1] >= width and (a.shape[1]-width) % 2 == 0
    left = (a.shape[1]-width)//2
    indexed = np.frombuffer(source.read_bytes(), dtype=np.uint8).reshape(200, 320)
    palette = np.frombuffer(source.with_suffix('.pal').read_bytes(), dtype=np.uint8).reshape(256, 3)
    report = read(report_path)
    assert report['ASCIISourceSHA256'] == '72fb1adcd5698a5e9770b91a499c347c7c8b05c567b93a8cca1f7249c49b4a7e'
    assert report['CellPixels'] == 8*scale
    assert report['Font'] == ('pod-en.golemfnt' if scale == 2 else 'pod-en-3x.golemfnt')
    cells = report.get('Cells') or []
    assert report['Enabled'] or not cells
    mask = np.zeros(a.shape[:2], dtype=bool)
    delta = np.any(a != b, axis=2)
    assert np.all(a[delta, 3] == 255), ('non-opaque changed glyph pixels', current)
    for cell in cells:
        x, y = cell['X'], cell['Y']
        assert x % 8 == y % 8 == 0 and 0 <= x <= 312 and 0 <= y <= 192
        assert 32 <= cell['Code'] <= 126 and not cell['Pending'], (current, cell)
        crop = indexed[y:y+8, x:x+8]
        assert hashlib.sha256(crop.tobytes()).hexdigest() == cell['SourceSHA256']
        yy, xx = slice(y*scale, (y+8)*scale), slice(left+x*scale, left+(x+8)*scale)
        mask[yy, xx] = True
        changed = delta[yy, xx]
        ink = a[yy, xx, :3]
        allowed = np.all(ink == palette[cell['Foreground']], axis=2) | np.all(ink == palette[cell['Background']], axis=2)
        assert np.all(allowed[changed]), ('changed glyph colors', current, cell)
    assert not np.any(delta & ~mask), ('change outside recognized ASCII cells', current, int(np.count_nonzero(delta & ~mask)))
    return dict(image=str(current), scale=scale, overlay=report['Enabled'], cells=len(cells),
                changed_pixels=int(np.count_nonzero(delta)), outside_cell_changes=0,
                image_sha256=sha(current), previous_sha256=sha(previous), source_sha256=sha(source),
                glyph_report_sha256=sha(report_path))


def negative_controls(folder, previous):
    for path in sorted((folder/'shots').glob('*.png')):
        source = folder/'sources'/(path.stem+'.idx')
        report_path = source.with_suffix('.glyphs.json')
        report = read(report_path)
        if report['Enabled'] and report.get('Cells'):
            break
    else:
        raise AssertionError('需要可見原生字格作負對照')
    baseline = previous/'shots'/path.name
    compare_image(path, baseline, source, report_path)
    cases = 0
    with tempfile.TemporaryDirectory(prefix='pod-glyph-negative-') as tmp:
        temp = Path(tmp)
        for kind in ('size', 'font', 'icon', 'source', 'overlay', 'pending'):
            changed = json.loads(json.dumps(report))
            if kind == 'size': changed['CellPixels'] = 8
            if kind == 'font': changed['Font'] = 'unknown.golemfnt'
            if kind == 'icon': changed['Cells'][0]['Code'] = 12
            if kind == 'source': changed['Cells'][0]['SourceSHA256'] = '0'*64
            if kind == 'overlay': changed['Enabled'] = False
            if kind == 'pending': changed['Cells'][0]['Pending'] = True
            damaged = temp/(kind+'.json')
            damaged.write_text(json.dumps(changed))
            try: compare_image(path, baseline, source, damaged)
            except AssertionError: cases += 1
            else: raise AssertionError('accepted damaged '+kind)
        pixels = np.array(Image.open(path).convert('RGBA'))
        scale = pixels.shape[0]//200
        mask = np.zeros(pixels.shape[:2], dtype=bool)
        for cell in report['Cells']:
            x, y = cell['X'], cell['Y']
            mask[y*scale:(y+8)*scale, x*scale:(x+8)*scale] = True
        y, x = np.argwhere(~mask)[0]
        pixels[y, x, 0] ^= 1
        damaged = temp/'outside.png'
        Image.fromarray(pixels, 'RGBA').save(damaged)
        try: compare_image(damaged, baseline, source, report_path)
        except AssertionError: cases += 1
        else: raise AssertionError('accepted pixel change outside glyph cells')
    assert cases == 7
    return cases


def preview(folder, previous, output):
    """Crop the most visibly changed number row; preserve native screenshot pixels."""
    name = 'pod-f004200.png'
    a, b = [np.array(Image.open(path).convert('RGBA')) for path in (folder/'shots'/name, previous/'shots'/name)]
    delta = np.any(a != b, axis=2)
    report = read(folder/'sources'/'pod-f004200.glyphs.json')
    groups = {}
    for cell in report['Cells']:
        x, y = cell['X'], cell['Y']
        count = int(np.count_nonzero(delta[y*3:(y+8)*3, x*3:(x+8)*3]))
        if count:
            groups.setdefault(y, []).append((cell, count))
    assert groups
    y = max(groups, key=lambda key: sum(count for _, count in groups[key]))
    cells = [cell for cell, _ in groups[y]]
    box = (max(0, min(cell['X'] for cell in cells)-32)*3, max(0, y-8)*3,
           min(320, max(cell['X'] for cell in cells)+24)*3, min(200, y+16)*3)
    assert not output.exists() and output.parent.stat().st_uid == 1000
    output.mkdir()
    names = ('glyph-font-hd-3x.png', 'glyph-font-original-3x.png')
    for pixels, name in zip((a, b), names):
        Image.fromarray(pixels, 'RGBA').crop(box).save(output/name)
    record = dict(frame=4200, scale=3, native_crop=list(box), codes=[cell['Code'] for cell in cells],
                  current_source_sha256=sha(folder/'shots'/'pod-f004200.png'),
                  previous_source_sha256=sha(previous/'shots'/'pod-f004200.png'),
                  files={name: sha(output/name) for name in names}, method='native same-state screenshot crop; no resampling')
    (output/'preview.json').write_text(json.dumps(record, ensure_ascii=False, indent=2)+'\n')
    return record


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    a = ap.parse_args()
    assert not a.out.exists(), '請保留既有收據'
    p = a.project
    root = p/'workplace/probe/g9-hd'
    helper = load(p/'tools/direct_ui_check.py')
    current = root/'nativeresidentglyph-font-v2/wall-direct'
    legacy = root/'nativeresidentglyph-font-v2/wall-legacy'
    previous = root/'nativeresidentglyph-fontbaseline-v1/wall-direct'
    hd = root/'nativeresidentglyph-play-v2/hd'
    hd_previous = root/'nativesurfacefillspans-play-v1/hd'
    binary = p/'workplace/dosgolem/workplace/bin/pod-play-nativeresidentglyph-v2'
    rows = []
    frontend = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt', 'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
    for label, folder, before, font_sample in [('hd', hd, hd_previous, False), ('font', current, previous, True)]:
        assert helper.endpoint(folder) == helper.endpoint(before)
        assert {k: v for k, v in tree(folder/'save').items() if k not in frontend} == {k: v for k, v in tree(before/'save').items() if k not in frontend}
        ci, bi = read(folder/'inputs.json'), read(before/'inputs.json')
        assert ci['binary_sha256'] == sha(binary)
        assert all(ci[key] == bi[key] for key in ('orig', 'font', 'text', 'theme'))
        for name in ('source.idx', 'source.pal', 'original.rgba'):
            assert (folder/name).read_bytes() == (before/name).read_bytes(), (label, name)
        sources = {name: digest for name, digest in tree(folder/'sources').items() if name.endswith(('.idx', '.pal'))}
        assert sources == tree(before/'sources')
        names = sorted(path.name for path in (folder/'shots').glob('*.png'))
        assert names and names == sorted(path.name for path in (before/'shots').glob('*.png'))
        images = []
        for name in names:
            stem = Path(name).stem
            source = folder/'sources'/(stem+'.idx')
            report = folder/'sources'/(stem+'.glyphs.json')
            for directory in ('shots', 'window-shots'):
                images.append(compare_image(folder/directory/name, before/directory/name, source, report))
        images.append(compare_image(folder/'frame.png', before/'frame.png', folder/'source.idx', folder/'source.glyphs.json'))
        rows.append(dict(route=label, original_state_equal=True, guest_files_equal=True,
                         screenshot_pairs=len(names), images=images, **audit(folder, font_sample)))
    assert helper.endpoint(current) == helper.endpoint(legacy) == helper.endpoint(root/'nativesurfacefillspans-ui-v1/wall-direct')
    assert read(current/'inputs.json') == read(legacy/'inputs.json')
    for name in ('source.idx', 'source.pal', 'source.glyphs.json', 'original.rgba', 'frame.png'):
        assert (current/name).read_bytes() == (legacy/name).read_bytes(), name
    for name in ('shots', 'window-shots', 'sources'):
        assert tree(current/name) == tree(legacy/name), name
    assert {k: v for k, v in tree(current/'save').items() if k not in frontend} == {k: v for k, v in tree(legacy/'save').items() if k not in frontend}
    measured = audit(legacy, True)
    assert measured == audit(current, True)
    sources = current/'sources'
    english, twox, original = [read(sources/('pod-f'+str(frame).zfill(6)+'.glyphs.json')) for frame in (4785, 4788, 4791)]
    assert english['Enabled'] and english['CellPixels'] == 24
    assert twox['Enabled'] and twox['CellPixels'] == 16
    assert not original['Enabled'] and not original.get('Cells')
    assert sum(image['changed_pixels'] for row in rows for image in row['images']) > 0
    result = dict(scope='既有HD demo、正常讀檔／地圖／工具列加語系、2／3倍及F12；變更限已認證ASCII格，原始狀態及直接／舊合成相同',
                  rows=rows, direct_legacy_screenshot_pairs=len(list((current/'shots').glob('*.png'))),
                  native_font_sizes=[16, 24], original_toggle_preserved=True,
                  current_binary_sha256=sha(binary), tool_sha256=sha(Path(__file__)),
                  image_metadata_negative_controls=negative_controls(current, previous),
                  direct_ui_tool_sha256=sha(p/'tools/direct-ui-receipt.sh'))
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({row['route']: dict(pairs=row['screenshot_pairs'], changed=sum(image['changed_pixels'] for image in row['images']), flash=row['raw_flash_count']) for row in rows}))


if __name__ == '__main__':
    main()
