#!/usr/bin/env python3
"""核對日韓完整候選的正常玩家抽樣；保留明示切換的原始量尺。"""
import argparse
import csv
import hashlib
import json
from pathlib import Path
import re
import struct
import sys
import traceback


def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p): return json.loads(p.read_text())
def tree(p): return {str(f.relative_to(p)): sha(f) for f in sorted(p.rglob('*')) if f.is_file()}
def pixels(p): return np.array(Image.open(p).convert('RGBA'))


TABLE_COUNTS = {'ecl-text': 4634, 'ui-text': 980, 'name': 1404,
    'composed-templates': 192, 'menu-masters': 150, 'host-text': 49,
    'title': 18, 'cover': 3, 'poster': 6, 'journal': 99,
    'walkthrough': 78, 'map-titles': 32, 'map-guide': 7}


def cell_sha(value): return hashlib.sha256(value.encode('utf-8')).hexdigest()
def review_key(x): return tuple(x[k] for k in ('lang', 'stem', 'index', 'field', 'kind'))


def bound_input(data, path):
    hashes = [h for name, h in data['inputs'].items() if Path(name).name == path.name]
    assert len(hashes) == 1 and hashes[0] == sha(path), ('stale data input', path)


def check_data(r, w, c):
    data = read(w/'data-check.json')
    assert data['status'] == 'PASS' and data['errors'] == []
    assert data['tool_sha256'] == sha(r/'tools/jako_check.py'), 'data checker changed'
    for path in (w/'plan.json', w/'numeric-review.json', r/'text/jako-term-fixes.tsv'):
        bound_input(data, path)
    assert data['counts']['tables'] == 26
    assert data['counts']['per_language'] == {'ja': 7652, 'ko': 7652}
    assert data['counts']['reviewed_warnings'] == 43
    assert len(data['negative_controls']) == 11 and all(x['passed'] for x in data['negative_controls'])
    plans = {lang: {t['stem']: t for t in tables} for lang, tables in read(w/'plan.json').items()}
    recorded = {(t['lang'], t['stem']): t for t in data['tables']}
    assert len(data['tables']) == len(recorded) == 26
    assert set(recorded) == {(lang, stem) for lang in ('ja', 'ko') for stem in TABLE_COUNTS}
    candidates = {}
    for lang in ('ja', 'ko'):
        assert set(plans[lang]) == set(TABLE_COUNTS)
        for stem, count in TABLE_COUNTS.items():
            path = c/'text'/f'{stem}.{lang}.tsv'
            with path.open() as f:
                reader = csv.DictReader(f, delimiter='\t', quoting=csv.QUOTE_NONE)
                rows = list(reader); header = reader.fieldnames
            report = recorded[(lang, stem)]
            assert len(rows) == len(plans[lang][stem]['rows']) == report['rows'] == report['expected_rows'] == count
            assert header == plans[lang][stem]['header'] == report['fields']
            assert sha(path) == report['sha256'], ('stale checked table', path)
            assert sha(r/'text'/path.name) == report['sha256'], ('formal text differs', path.name)
            candidates[(lang, stem)] = rows
    original_review = read(w/'numeric-review.json')
    reviews = {review_key(x): x for x in original_review}
    checks = {review_key(x): x for x in data['warning_reviews']}
    assert len(original_review) == len(data['warning_reviews']) == len(reviews) == len(checks) == 43
    assert set(reviews) == set(checks)
    for key, check in checks.items():
        lang, stem, index, field, _ = key
        source = plans[lang][stem]['rows'][index]['changes'][field]['english']
        target = candidates[(lang, stem)][index][field]
        assert check['matched'] is True and check['verdict'] in ('equivalent-number', 'layout-title-copy')
        assert check['source_sha256'] == reviews[key]['source_sha256'] == cell_sha(source)
        assert check['target_sha256'] == reviews[key]['target_sha256'] == cell_sha(target)
        assert check['verdict'] == reviews[key]['verdict'] and check['reason'] == reviews[key]['reason']
    with (r/'text/jako-term-fixes.tsv').open() as f:
        fixes = list(csv.DictReader(f, delimiter='\t', quoting=csv.QUOTE_NONE))
    applied = {(x['lang'], x['stem'], x['index'], x['field']): x for x in data['display_fixes']}
    assert len(fixes) == len(data['display_fixes']) == len(applied) == 32
    for fix in fixes:
        key = fix['lang'], fix['stem'], int(fix['index']), fix['field']
        lang, stem, index, field = key; check = applied[key]
        target = candidates[(lang, stem)][index][field]
        assert check['matched'] is True
        assert cell_sha(fix['before']) == fix['before_sha256'] == check['before_sha256']
        assert target == fix['after'] and cell_sha(target) == check['target_sha256'] == check['after_sha256']
    return data, recorded


def check_frontend(r, w, c, order):
    front = read(w/'frontend/receipt.json')
    assert front['status'] == 'PASS', ('frontend incomplete', front['status'])
    assert front['available_langs'] == order
    assert front['binary']['sha256'] == sha(r/'workplace/dosgolem/workplace/bin/pod-play')
    assert front['tool']['sha256'] == sha(r/'tools/jako_frontend.py')
    expected_inputs = {f'{directory}/{name}': h for directory in ('text', 'font', 'theme')
                       for name, h in tree(c/directory).items()}
    assert front['input_sha256'] == expected_inputs, 'frontend tested different candidate'
    expected = {f'explicit-{code}': order[(i+1) % len(order)] for i, code in enumerate(order)}
    expected.update({f'cycle-{i+1}': order[(i+1) % len(order)] for i in range(len(order))})
    for code in ('ja', 'ko'):
        i = order.index(code)
        expected[f'system-{code}-fresh'] = order[(i+1) % len(order)]
        expected[f'system-{code}-restart-en'] = order[(i+2) % len(order)]
    negatives = {f'{code}-{kind}' for code in ('ja', 'ko') for kind in ('missing-host', 'bar12-as-3x')}
    names = [case['name'] for case in front['cases']]
    assert len(names) == len(set(names)) == 18 and set(names) == set(expected) | negatives
    commands = {x['name']: x for x in front['commands']}
    origin = Path(front['run_directory'])
    actual_run = w/'frontend'/origin.name
    def evidence(path): return actual_run/Path(path).relative_to(origin)
    for case in front['cases']:
        name = case['name']; command = commands[name]
        log = evidence(command['log'])
        assert sha(log) == command['log_sha256'] == case['log_sha256']
        assert case['actual'] == case['expected']
        if name in negatives:
            error = '沒有這個語言或缺字型' if name.endswith('missing-host') else '24×24'
            assert case['kind'] == 'negative' and case['expected'] == 'rejected'
            assert case['rejected'] is True and case['error_present'] is True
            assert case['exit_code'] == command['exit_code'] != 0
            assert case['expected_error'] == error and error in log.read_text()
        else:
            assert case['kind'] == 'frontend-language' and case['actual'] == expected[name]
            assert case['exit_code'] == command['exit_code'] == 0
            assert sha(evidence(case['shot'])) == case['shot_sha256']
            if not name.startswith('explicit-'):
                assert case['same_frame_pixels'] is True
                assert sha(evidence(case['reference_shot'])) == case['reference_shot_sha256'] == case['shot_sha256']
    assert front['observed_f8_cycle'] == order[1:] + order[:1]
    summary = front['summary']
    assert summary['cases'] == 18 and summary['negative_cases'] == 4 and summary['positive_cases'] == 14
    assert summary['same_frame_comparisons'] == 9 and summary['input_hashes_unchanged'] is True
    return front


def check_clean_assets(r, w, c):
    manifest_path = r/'review/hd/runtime/manifest.json'
    manifest = read(manifest_path)
    expected_theme = {x['path']: x['sha256'] for x in manifest['theme']}
    expected_fonts = {x['path']: x['sha256'] for x in manifest['fonts']}
    assert len(manifest['theme']) == len(expected_theme) == 1888
    assert len(manifest['fonts']) == len(expected_fonts) == 18
    clean = w/'clean-main/workplace/verified-assets'
    receipt = read(clean/'assets-receipt.json')
    assert receipt['status'] == 'PASS' and receipt['manifest_sha256'] == sha(manifest_path)
    assert tree(clean/'theme') == tree(c/'theme') == expected_theme
    assert tree(clean/'font') == tree(c/'font') == expected_fonts == receipt['fonts']
    assert tree(r/'workplace/hd/theme') == expected_theme
    assert tree(r/'workplace/font') == expected_fonts
    assert receipt['theme_files'] == 1888
    assert receipt['theme_tree_sha256'] == hashlib.sha256(json.dumps(expected_theme, sort_keys=True).encode()).hexdigest()
    source_tree = (w/'clean-main-tree.txt').read_text().strip()
    assert re.fullmatch('[0-9a-f]{40}', source_tree)
    return dict(source_tree=source_tree, receipt_sha256=sha(clean/'assets-receipt.json'),
                manifest_sha256=sha(manifest_path), theme_files=1888, font_files=18,
                theme_tree_sha256=receipt['theme_tree_sha256'], current_candidate_exact=True)


def check(a):
    global np, Image
    import numpy as np
    from PIL import Image
    sys.path.insert(0, str(Path(__file__).resolve().parent))
    from zhcn_check import state
    from english_panels_check import dos_inputs
    r = a.root; w = a.work; c = w/'candidate'
    baseline = read(r/'docs/re/027-zhcn-runtime.json')['runtime']['zh-TW']
    front = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt', 'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
    old_script = re.search(r"script='([^']+)'", (r/'tools/zhcn_gui.sh').read_text())[1]
    order = ['zh-TW', 'en', 'ja', 'ko', 'zh-CN']; cases = {}; f12 = []
    for code in ('zh-TW', 'ja', 'ko'):
        p = w/('gui-'+code); log = (p/'run.log').read_text(); s = state(log)
        assert s == baseline['state'], (code, 'original state')
        assert dos_inputs((p/'script.txt').read_text()) == dos_inputs(old_script)
        guest = {n: h for n, h in tree(p/'game').items() if n not in front}
        before = {n: h for n, h in read(p/'before.json').items() if n not in front}
        assert guest == before == baseline['guest'] and len(guest) == 132
        assert (p/'game/pod-lang.txt').read_text().strip() == code
        assert sha(p/'source.idx') == baseline['source_sha256'] and sha(p/'source.pal') == baseline['palette_sha256']
        assert '字型缺字' not in log and '截圖失敗' not in log
        draw = read(p/'draw-report.json')
        assert draw['HostBytes'] == draw['NativeBytes'] and draw['FallbackBytes'] == draw['Mismatches'] == draw['LegacyRGBFrames'] == 0
        shots = sorted((p/'shots').glob('*.png'))
        scripted = {int(frame) for frame, cmd in (x.split(':', 1) for x in (p/'script.txt').read_text().strip().split(',')) if cmd == 'shot'}
        assert {int(x.stem.split('f')[-1]) for x in shots} == scripted
        for frame, key in ((1000, 'title-001'), (1100, 'title-002'), (1200, 'title-003')):
            assert np.array_equal(pixels(p/'shots'/f'pod-f{frame:06}.png'), pixels(c/'theme/localized'/code/(key+'.png')))
        poster_checks = []
        for shot in shots:
            frame = int(shot.stem.split('f')[-1]); scale = 2 if 4720 <= frame < 4750 else 3
            index = order.index(code)
            for at in (4352, 4356, 4360, 4364, 4368):
                if frame >= at: index = (index+1) % len(order)
            locale = 'en' if 4630 <= frame < 4640 else order[index]
            full = pixels(p/'window'/shot.name)
            assert full.shape == (200*scale, 480*scale, 4)
            for side, x in (('left', 0), ('right', 400*scale)):
                folder = c/'theme/posters'; bg = Image.open(folder/f'poster-{side}-{scale}x.png').convert('RGBA')
                text = Image.open(folder/f'poster-text-{side}.{locale}-{scale}x.png').convert('RGBA')
                assert np.array_equal(full[:, x:x+80*scale], np.array(Image.alpha_composite(bg, text))), (code, frame, side, locale)
            poster_checks.append(dict(frame=frame, locale=locale, scale=scale, sha256=sha(p/'window'/shot.name)))
        raw = int(re.search(r'閃現 (\d+) 次', log)[1])
        events = re.findall(r'閃現 第 (\d+) 格原樣呈現、第 (\d+) 格被換 列 (\d+) 欄 ([0-9 ]+)', log)
        accounted = 0; f12_cells = set(); scale_cells = set()
        for shown, frame, row, cols in events:
            shown, frame, row = int(shown), int(frame), int(row)
            cells = {(row, int(x)) for x in cols.split()}; accounted += len(cells)
            if shown == 4749 and frame == 4750: scale_cells.update(cells)
            else:
                assert 4630 <= shown < 4640 and frame == 4640, (code, 'unexpected flash', shown, frame, row, cols)
                f12_cells.update(cells)
        assert accounted == raw and scale_cells == {(y, x) for y in range(3, 14) for x in range(3, 14)}
        for suffix in ('.idx', '.pal'):
            paths = [p/'sources'/('pod-f'+str(n).zfill(6)+suffix) for n in (4620, 4632, 4642)]
            assert len({x.read_bytes() for x in paths}) == 1, (code, 'F12 source changed')
            assert (p/'sources'/('pod-f004740'+suffix)).read_bytes() == (p/'sources'/('pod-f004770'+suffix)).read_bytes()
        f12.append(sha(p/'shots/pod-f004632.png'))
        cases[code] = dict(state=s, original_equal=True, guest_files=132, guest_unchanged=True,
            raw_flash=raw, unexpected_flash=0, f12_cells=sorted(f12_cells), scale_cells=sorted(scale_cells),
            classification='explicit F12 original display4630..4640 and scale switch4749..4750; source IDX/PAL unchanged',
            source_sha256=sha(p/'source.idx'), palette_sha256=sha(p/'source.pal'), draw=draw,
            script_sha256=sha(p/'script.txt'), log_sha256=sha(p/'run.log'),
            shots={x.name: sha(x) for x in shots}, posters=poster_checks)
    assert len(set(f12)) == 1, 'F12 original differs across locales'
    data, checked_tables = check_data(r, w, c)
    fonts = read(c/'font-build/fonts.json'); glyphs = {}
    assert set(fonts) == {'ja', 'ko'}
    for code, group in fonts.items():
        assert set(group['inputs']) == {f'{stem}.{code}.tsv' for stem in TABLE_COUNTS}
        for stem in TABLE_COUNTS:
            name = f'{stem}.{code}.tsv'
            assert group['inputs'][name] == sha(c/'text'/name) == checked_tables[(code, stem)]['sha256'], ('stale font corpus', name)
        required_fonts = {f'pod-{code}.golemfnt': 16, f'pod-{code}-3x.golemfnt': 24,
                          f'pod-{code}-bar12.golemfnt': 12, f'pod-{code}-bar22.golemfnt': 22}
        assert set(group['fonts']) == set(required_fonts)
        glyphs[code] = {}
        for name, info in group['fonts'].items():
            blob = (c/'font'/name).read_bytes(); assert sha(c/'font'/name) == info['sha256']
            assert blob[:8] == b'GOLEMFNT'
            width, height, count = struct.unpack_from('<HHI', blob, 8); step = 5+((width+7)//8)*height
            assert width == height == info['cell'] == required_fonts[name]
            assert count == info['characters'] and len(blob) == 16 + count*step
            entries = [struct.unpack_from('<IB', blob, 16+i*step) for i in range(count)]
            assert sha(c/'font-build'/info['charset']) == info['charset_sha256']
            assert [x[0] for x in entries] == read(c/'font-build'/info['charset'])
            glyphs[code][name] = dict(characters=count, fallback=[f'U+{cp:04X}' for cp, tag in entries if tag == 8], sha256=info['sha256'])
    packages = []
    for receipt in ('026-package-runtime.json', '033-hd-patch-runtime.json'):
        lock = read(r/'docs/re'/receipt)
        for item in lock['files']:
            p = r/'dist-all'/lock['version']/item.get('kind', 'patch')/item['name']
            assert p.stat().st_size == item['size'] and sha(p) == item['sha256'], ('old package changed', p)
            packages.append(dict(path=str(p.relative_to(r)), sha256=item['sha256']))
    assert len(packages) == 12
    frontend = check_frontend(r, w, c, order)
    clean_assets = check_clean_assets(r, w, c)
    # 此切片只改資料；使用033已驗程式，另外重播完整繁中DEMO。
    assert sha(r/'workplace/dosgolem/workplace/bin/pod-play') == 'aa89fa8274b48561b84fde3e23e12c4ff9c0966a9b156d4f8cd5e8cd2004ba53'
    expect = read(r/'review/hd/runtime/manifest.json')['runtime_expected']
    demo = w/'demo'; demo_log = (demo/'run.log').read_text()
    end = re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)', demo_log)
    digest = re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)', demo_log)
    assert end and list(end.groups()) == expect['end']
    assert digest and list(digest.groups()) == expect['digest']
    assert 'pod-play: 閃現 0 次' in demo_log
    for name, key in (('source.idx', 'source_indexed_sha256'), ('source.pal', 'source_palette_sha256'), ('frame.png', 'frame_sha256')):
        assert sha(demo/name) == expect[key], ('DEMO source/frame', name)
    demo_guest = {n: h for n, h in tree(demo/'save').items() if n not in front}
    assert demo_guest == expect['guest'] and len(demo_guest) == 127
    assert tree(r/'text') == expect['text_tree']
    demo_proof = dict(frames=7400, steps=712250000, state=state(demo_log), flash_count=0,
        guest_files=127, guest_unchanged=True, frame_sha256=sha(demo/'frame.png'),
        log_sha256=sha(demo/'run.log'), original_and_full_B_frame_equal=True,
        binary_scope='unchanged033 binary; this data-only slice does not rebuild Go')
    subtitles = read(w/'test/subtitles.json')
    assert subtitles['passed'] and subtitles['go_version'] == 'go1.24.13'
    assert set(subtitles['languages']) == {'ja', 'ko'}
    for name, want in subtitles['inputs_sha256'].items(): assert sha(c/name) == want
    sub_log = (w/'test/subtitles.log').read_text()
    assert re.search(r'(?m)^ok\s+github\.com/wicanr2/dosgolem/apps/pod(?:\s|$)', sub_log)
    assert not re.search(r'(?m)^(?:FAIL(?:\s|$)|--- FAIL:)', sub_log)
    expected_cues = {(code, f'intro.{n:03}', scale) for code in ('ja', 'ko') for n in range(1, 9) for scale in (2, 3)}
    assert len(subtitles['samples']) == len(expected_cues) == 32
    assert {(x['language'], x['cue'], x['scale']) for x in subtitles['samples']} == expected_cues
    sub_files = set()
    for sample in subtitles['samples']:
        for flag in ('all_text_preserved', 'animation_168_rows_unchanged', 'fits_complete_cue',
                     'missing_glyph_fallback', 'missing_text_fallback', 'subtitle_present'):
            assert sample[flag] is True
        assert sample['missing_glyphs'] == 0 and 0 < len(sample['line_cells']) <= 2
        assert max(sample['line_cells']) <= 38 and sum(sample['line_cells']) == sample['packed_cells']
        host = {x.split('\t')[0]: x.split('\t')[1] for x in (c/'text'/('host-text.'+sample['language']+'.tsv')).read_text().splitlines()[1:]}
        assert hashlib.sha256(host[sample['cue']].encode()).hexdigest() == sample['text_sha256']
        source = r/sample['source']
        assert sha(source) == sample['source_indexed_sha256'] and sha(source.with_suffix('.pal')) == sample['source_palette_sha256']
        if sample['scale'] == 3:
            file = w/sample['file']; sub_files.add(file.name)
            assert sha(file) == sample['png_sha256']
            assert hashlib.sha256(pixels(file).tobytes()).hexdigest() == sample['rgba_sha256']
    assert sub_files == {p.name for p in (w/'subtitle-samples').glob('*.png')} and len(sub_files) == 16
    activation = read(w/'activation.json')
    assert activation['status'] == 'PASS' and activation['theme_files'] == 1888 and activation['fonts'] == 18
    assert activation['binary_sha256'] == sha(r/'workplace/dosgolem/workplace/bin/pod-play')
    documentation = {}
    for code in ('ja', 'ko'):
        name = f'docs/screenshots/guide-runtime-{code}.png'
        source = w/f'gui-{code}/shots/pod-f004346.png'
        assert sha(r/name) == sha(source)
        documentation[name] = dict(sha256=sha(source), source=str(source.relative_to(w)), kind='normal loaded-save guide')
    panels = read(w/'test/panels.json')
    assert set(panels) == {'ja', 'ko'}
    expected_samples = set()
    for code, panel in panels.items():
        assert panel['journal_entries'] == 99 and panel['guide_chapters'] == 78
        assert panel['journal_pages'] >= 99 and panel['guide_pages'] >= 78
        assert set(panel['sample_pages']) == {'1', '11', '57', '74'}
        for entry, pages in panel['sample_pages'].items():
            assert isinstance(pages, int) and 1 <= pages <= panel['journal_max_pages']
            for page in {1, pages}:
                for scale in (2, 3):
                    expected_samples.add(f'{code}-journal-{int(entry):03}-p{page}-{scale}x.png')
    layout_log = (w/'test/layout.log').read_text()
    assert re.search(r'(?m)^ok\s+github\.com/wicanr2/dosgolem/apps/pod(?:\s|$)', layout_log)
    assert not re.search(r'(?m)^(?:FAIL(?:\s|$)|--- FAIL:)', layout_log)
    component_samples = {p.name: sha(p) for p in sorted((w/'test/layout').glob('*.png'))}
    assert set(component_samples) == expected_samples, ('component sample set', sorted(expected_samples), sorted(component_samples))
    proof = dict(date='2026-11-02', issue=76, status='PASS within Japanese/Korean sample scope',
        engine_commit=(r/'tools/rebuild-engine.commit').read_text().strip(),
        binary_sha256=sha(r/'workplace/dosgolem/workplace/bin/pod-play'),
        baseline=dict(path='docs/re/027-zhcn-runtime.json', sha256=sha(r/'docs/re/027-zhcn-runtime.json')),
        data=data, panels=panels, frontend=frontend, fonts=glyphs,
        assets=read(w/'assets-preview/assets.json'), theme_files=len(tree(c/'theme')), font_files=len(tree(c/'font')),
        clean_git_source_tree=clean_assets['source_tree'], clean_assets=clean_assets,
        normal_demo=demo_proof, subtitle_components=subtitles,
        formal_activation=activation, documentation_samples=documentation,
        normal=cases, five_language_order=order, f12_same_original_frame=True, old_packages=packages,
        translation_provenance=dict(merge=read(c/'merge.json'), numeric_review_sha256=sha(w/'numeric-review.json'),
            term_review_sha256=sha(w/'term-review.json'), ui_name_review_sha256=sha(w/'ui-name-review.json')),
        component_samples=component_samples,
        validation_inputs={str(p.relative_to(w)): sha(p) for p in (w/'data-check.json',
            w/'frontend/receipt.json', w/'test/panels.json', w/'test/layout.log', c/'font-build/fonts.json')},
        tools={p.name: sha(p) for p in sorted((r/'tools').glob('jako_*')) if p.is_file()},
        limits=['machine translation with contextual review; no native-language review',
            'one normal loaded-save route per locale; not all dialogue or endings',
            'journal body screenshots are component samples; fresh journal auto-trigger not replayed',
            'intro subtitles use existing output keys; full intro and FINAL not replayed in this slice',
            'no new release packages; existing twelve packages retain their previous language scope'])
    return proof


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for n in ('root', 'work', 'receipt'): ap.add_argument('--'+n, type=Path, required=True)
    a = ap.parse_args()
    # 本次失敗須取代舊的PASS，不能留下看似仍有效的完成收據。
    a.receipt.write_text(json.dumps(dict(status='RUNNING', issue=76), ensure_ascii=False)+'\n')
    try:
        proof = check(a)
    except Exception as error:
        a.receipt.write_text(json.dumps(dict(status='FAIL', issue=76,
            error=str(error), exception=type(error).__name__, traceback=traceback.format_exc()),
            ensure_ascii=False, indent=2)+'\n')
        raise
    a.receipt.write_text(json.dumps(proof, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(dict(status=proof['status'], theme=proof['theme_files'], fonts=proof['font_files'], packages=len(proof['old_packages'])), ensure_ascii=False))


if __name__ == '__main__': main()
