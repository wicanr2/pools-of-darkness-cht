#!/usr/bin/env python3
"""核對日韓候選並準備現有重建清單；預設唯讀，--apply才保存素材與字集。

輸入：--candidate下的theme、font、font-build/fonts.json、charset*.json及text。
先把候選日韓正式文字同步到--root/text，再用--apply登錄。工具不複製文字、
GOLEMFNT或原版idx，不修改引擎與歷史正常路徑的執行期預期值。
"""
import argparse
import copy
import hashlib
import json
import os
from pathlib import Path
import re
import struct
import tempfile


LANGUAGES = ('ja', 'ko')
FONT_CELLS = {'': 16, '-3x': 24, '-bar12': 12, '-bar22': 22}
MANIFEST = 'review/hd/runtime/manifest.json'
REVIEW = 'review/hd/jako/'


def digest(data):
    return hashlib.sha256(data).hexdigest()


def sha(path):
    return digest(path.read_bytes())


def inside(root, name):
    if not isinstance(name, str) or not name or '\\' in name or Path(name).is_absolute():
        raise ValueError(f'不合法的相對路徑：{name}')
    path = root/name
    if root.resolve() not in path.resolve().parents:
        raise ValueError(f'路徑超出根目錄：{name}')
    return path


def tree(folder):
    if not folder.is_dir():
        raise ValueError(f'缺少目錄：{folder}')
    return {str(p.relative_to(folder)): sha(inside(folder, str(p.relative_to(folder))))
            for p in sorted(folder.rglob('*')) if p.is_file()}


def unique_rows(rows, label):
    result = {row['path']: row for row in rows}
    if len(result) != len(rows):
        raise ValueError(f'{label}路徑重複')
    return result


def font_sources(root, lang, cell):
    region = 'JP' if lang == 'ja' else 'TW'
    weight = 'B' if cell in (22, 24) else 'M'
    otf = f'font/source/rounded/GenSenRounded2{region}-{weight}.otf'
    fallback = 'font/source/unifont/'+('unifont_jp-17.0.05.hex' if lang == 'ja' else 'unifont_t-17.0.05.hex')
    return {name: sha(inside(root, name)) for name in (otf, fallback)}


def validate_font(data, points, cell, characters):
    if (not isinstance(points, list) or not points or
            any(type(cp) is not int or not 0 <= cp <= 0x10FFFF or 0xD800 <= cp <= 0xDFFF for cp in points) or
            points != sorted(set(points))):
        raise ValueError('字集須為唯一遞增的合法碼點')
    if len(data) < 16 or data[:8] != b'GOLEMFNT':
        raise ValueError('候選字型標頭不合')
    width, height, count = struct.unpack_from('<HHI', data, 8)
    stride = 5+((cell+7)//8)*cell
    if (width, height, count) != (cell, cell, len(points)) or count != characters or len(data) != 16+stride*count:
        raise ValueError('候選字型尺寸、字數或長度不合')
    for i, cp in enumerate(points):
        actual, tag = struct.unpack_from('<IB', data, 16+stride*i)
        if actual != cp or tag not in (7, 8):
            raise ValueError('候選字型碼點或來源標記不合')


def expected_theme_paths():
    return {f'localized/{lang}/{name}' for lang in LANGUAGES
            for name in ['theme.json']+[f'title-00{i}.{ext}' for i in (1, 2, 3) for ext in ('idx', 'png')]} | {
                f'posters/poster-text-{side}.{lang}-{scale}x.png'
                for lang in LANGUAGES for side in ('left', 'right') for scale in (2, 3)}


def plan(root, candidate):
    target = inside(root, MANIFEST)
    lock_bytes = target.read_bytes()
    lock = json.loads(lock_bytes)
    if lock['version'] != 1:
        raise ValueError('只支援現有version 1重建清單')
    old_theme = unique_rows(lock['theme'], '主題')
    old_fonts = unique_rows(lock['fonts'], '字型')
    for name, want in lock['source_files'].items():
        if sha(inside(root, name)) != want:
            raise ValueError(f'既有重建輸入不同，須先核對：{name}')
    for row in lock['theme']:
        if row['kind'] == 'repo-file' and sha(inside(root, row['source'])) != row['sha256']:
            raise ValueError(f'既有素材不同，須先核對：{row["source"]}')

    managed_theme = expected_theme_paths()
    managed_fonts = {f'pod-{lang}{suffix}.golemfnt' for lang in LANGUAGES for suffix in FONT_CELLS}
    actual_theme = tree(candidate/'theme')
    baseline_theme = {k: v['sha256'] for k, v in old_theme.items() if k not in managed_theme}
    if set(actual_theme) != set(baseline_theme) | managed_theme:
        raise ValueError('候選主題須完整保留既有素材，並只增補日韓22檔')
    if any(actual_theme[k] != v for k, v in baseline_theme.items()):
        raise ValueError('既有主題素材改變，須獨立驗收')
    actual_fonts = tree(candidate/'font')
    baseline_fonts = {k: v['sha256'] for k, v in old_fonts.items() if k not in managed_fonts}
    if len(baseline_fonts) != 10 or set(actual_fonts) != set(baseline_fonts) | managed_fonts:
        raise ValueError('候選字型須包含既有10檔及日韓8檔，不得混入暫存字型')
    if any(actual_fonts[k] != v for k, v in baseline_fonts.items()):
        raise ValueError('既有10份字型的SHA改變')

    updated = copy.deepcopy(lock)
    generator = 'tools/rounded_font.py'
    updated['source_files'][generator] = sha(inside(root, generator))
    writes = {}
    theme_rows = []
    for name in sorted(managed_theme):
        path = inside(candidate/'theme', name)
        data = path.read_bytes()
        if digest(data) != actual_theme[name]:
            raise ValueError(f'候選主題在檢查期間改變：{name}')
        lang = name.split('/')[1] if name.startswith('localized/') else re.search(r'\.(ja|ko)-', name)[1]
        reference = name.replace(f'localized/{lang}/', 'localized/zh-TW/').replace(f'.{lang}-', '.zh-TW-')
        if reference not in old_theme:
            raise ValueError(f'缺少已驗繁中來源列：{reference}')
        row = copy.deepcopy(old_theme[reference])
        row.update(path=name, sha256=digest(data))
        if name.endswith('.idx'):
            if row['kind'] != 'original-index' or row['sha256'] != old_theme[reference]['sha256']:
                raise ValueError(f'日韓TITLE索引與已驗繁中來源不同：{name}')
        else:
            if row['kind'] != 'repo-file' or path.suffix not in ('.png', '.json'):
                raise ValueError(f'未知日韓主題來源：{name}')
            row['source'] = REVIEW+name
            writes[row['source']] = data
        theme_rows.append(row)
    for lang in LANGUAGES:
        meta = json.loads(writes[REVIEW+f'localized/{lang}/theme.json'])
        reference = json.loads(inside(root, old_theme['localized/zh-TW/theme.json']['source']).read_bytes())
        expected = copy.deepcopy(reference)
        for item in expected['items']:
            for field in ('png', 'indexed'):
                name = f'localized/{lang}/'+item[field]
                if name not in managed_theme:
                    raise ValueError('語系主題引用了未登錄的檔案')
            item['png_sha256'] = actual_theme[f'localized/{lang}/'+item['png']]
            if item['indexed_sha256'] != actual_theme[f'localized/{lang}/'+item['indexed']]:
                raise ValueError('語系主題的索引SHA不合')
        if meta != expected:
            raise ValueError('日韓TITLE僅允許改PNG雜湊，不可變更既有版面或來源')
    updated['theme'] = [r for r in lock['theme'] if r['path'] not in managed_theme]+theme_rows

    metadata = json.loads((candidate/'font-build/fonts.json').read_bytes())
    if set(metadata) != set(LANGUAGES):
        raise ValueError('fonts.json須明列ja與ko')
    font_rows, text_inputs = [], {}
    for lang in LANGUAGES:
        section = metadata[lang]
        expected_names = {f'pod-{lang}{suffix}.golemfnt' for suffix in FONT_CELLS}
        if set(section['fonts']) != expected_names or not section['inputs']:
            raise ValueError(f'{lang}字型建置清單不完整')
        for name, want in section['inputs'].items():
            if not name.endswith(f'.{lang}.tsv') or sha(inside(candidate/'text', name)) != want:
                raise ValueError(f'字型使用的候選文字已改變，請先重生字型：{name}')
            if name in text_inputs and text_inputs[name] != want:
                raise ValueError('字型文字來源互相衝突')
            text_inputs[name] = want
        for suffix, cell in FONT_CELLS.items():
            name = f'pod-{lang}{suffix}.golemfnt'
            record = section['fonts'][name]
            if record['cell'] != cell or actual_fonts[name] != record['sha256']:
                raise ValueError(f'候選字型與建置收據不合：{name}')
            charset_name = record['charset']
            if Path(charset_name).name != charset_name:
                raise ValueError('字集來源只能是font-build下的檔名')
            charset = inside(candidate/'font-build', charset_name).read_bytes()
            if digest(charset) != record['charset_sha256']:
                raise ValueError(f'字集SHA不合：{name}')
            font_data = inside(candidate/'font', name).read_bytes()
            if digest(font_data) != record['sha256']:
                raise ValueError(f'候選字型在檢查期間改變：{name}')
            validate_font(font_data, json.loads(charset), cell, record['characters'])
            sources = font_sources(root, lang, cell)
            if tuple(sources.values()) != (record['otf_sha256'], record['fallback_sha256']):
                raise ValueError(f'候選字型的OTF或Unifont來源不合：{name}')
            updated['source_files'].update(sources)
            destination = 'font/charset-'+Path(name).stem+'.json'
            writes[destination] = charset
            updated['source_files'][destination] = digest(charset)
            font_rows.append(dict(path=name, kind='rounded', cell=cell, charset=destination, sha256=record['sha256']))
    updated['fonts'] = [r for r in lock['fonts'] if r['path'] not in managed_fonts]+font_rows
    text_tree = tree(root/'text')
    pending_text = [name for name, want in sorted(text_inputs.items()) if text_tree.get(name) != want]
    text_tree.update(text_inputs)
    updated['runtime_expected']['text_tree'] = dict(sorted(text_tree.items()))
    updated['scope'] = f'日韓開場與海報增補；{len(updated["theme"])}檔主題、{len(updated["fonts"])}份字型；私人素材與本機重建，歷史執行期預期值不變'
    if set(updated) != set(lock) or set(updated['runtime_expected']) != set(lock['runtime_expected']):
        raise ValueError('不得改變manifest schema')
    for name, data in writes.items():
        destination = inside(root, name)
        if destination.exists() and sha(destination) != digest(data):
            prior = lock['source_files'].get(name)
            prior = prior or next((r['sha256'] for r in lock['theme'] if r.get('source') == name), None)
            if prior != sha(destination):
                raise ValueError(f'目的地有未登錄的不同內容，不覆寫：{name}')
    report = dict(status='DRY-RUN', manifest=MANIFEST, manifest_before_sha256=digest(lock_bytes),
                  totals=dict(theme_before=len(lock['theme']), theme_after=len(updated['theme']),
                              fonts_before=len(lock['fonts']), fonts_after=len(updated['fonts']),
                              original_index_aliases=6, repo_theme_files=16, charset_files=8),
                  baseline_theme_unchanged=True, baseline_theme_files=len(baseline_theme), existing_10_fonts_unchanged=True,
                  historical_runtime_outputs_unchanged=True, pending_text_files=pending_text,
                  generator_source=dict(path=generator, sha256=updated['source_files'][generator]),
                  writes=sorted(writes), copied_original_indexes=[], copied_font_binaries=[],
                  expected_theme_sha256=digest(json.dumps(actual_theme, sort_keys=True).encode()),
                  expected_fonts=actual_fonts, expected_text_tree_sha256=digest(json.dumps(text_tree, sort_keys=True).encode()))
    return target, lock_bytes, updated, writes, report


def atomic_write(path, data):
    path.parent.mkdir(parents=True, exist_ok=True)
    for existing in (path.parent, path):
        if existing.exists() and (existing.stat().st_uid, existing.stat().st_gid) != (os.getuid(), os.getgid()):
            raise ValueError(f'目的地擁有權不同：{existing}')
    temporary = None
    try:
        with tempfile.NamedTemporaryFile(dir=path.parent, prefix='.'+path.name+'.', delete=False) as f:
            temporary = Path(f.name)
            f.write(data)
        temporary.chmod(path.stat().st_mode & 0o777 if path.exists() else 0o644)
        temporary.replace(path)
    finally:
        if temporary is not None and temporary.exists():
            temporary.unlink()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root', type=Path, required=True)
    ap.add_argument('--candidate', type=Path, required=True)
    mode = ap.add_mutually_exclusive_group()
    mode.add_argument('--apply', action='store_true', help='核對完成後保存16份主題來源、8份字集及manifest；不複製文字')
    mode.add_argument('--dry-run', action='store_true', help='預設；只核對並列出待寫入路徑')
    a = ap.parse_args()
    try:
        target, before, updated, writes, report = plan(a.root, a.candidate)
        if a.apply:
            if report['pending_text_files']:
                raise ValueError('請先把候選正式文字同步到root/text：'+', '.join(report['pending_text_files']))
            if tree(a.root/'text') != updated['runtime_expected']['text_tree']:
                raise ValueError('正式文字在檢查期間改變，請重新執行')
            if target.read_bytes() != before:
                raise ValueError('manifest在檢查期間改變，請重新執行')
            for name, data in writes.items():
                destination = inside(a.root, name)
                if not destination.exists() or sha(destination) != digest(data):
                    atomic_write(destination, data)
            atomic_write(target, (json.dumps(updated, ensure_ascii=False, indent=2)+'\n').encode())
            report.update(status='APPLIED', manifest_after_sha256=sha(target))
        print(json.dumps(report, ensure_ascii=False, indent=2))
    except (ValueError, KeyError, IndexError, TypeError, OSError, struct.error) as e:
        ap.exit(1, str(e)+'\n')


if __name__ == '__main__':
    main()
