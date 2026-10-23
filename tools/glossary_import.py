#!/usr/bin/env python3
"""把前作術語表併進本作 text/glossary.tsv（G5，AGENTS.md 第 3.3 節譯名優先序第 3 項）。

用法（sosb-tools 容器內）：python3 -I tools/glossary_import.py --into text/glossary.tsv --from sosb=<前作 glossary.tsv>
- 英文只留字母、不分大小寫比對（同 tools/journal_merge.py 的術語合併）；本作已有的列不動，前作只補本作沒有的英文。
- 補進的列 source 寫前作短名，note 前置「前作來源 X」（X 是前作表的 source 欄），保留前作 note。
  短名寫 keep 時（本作新詞表）source 與 note 照原表。
- 類別 god 併成本作的 deity。
- 本作既有列的 note 去掉重複片段（以「；」分段，保留第一次出現的順序），這是手札批次合併留下的。
- 依英文不分大小寫排序後寫回；--dry-run 只印統計。
"""
import argparse
import csv
import re
from pathlib import Path

HEAD = ['en', 'zh', 'category', 'source', 'note']
CATEGORY = {'god': 'deity'}


def norm(t):
    return re.sub(r'[^a-z]', '', t.lower())


def read(path):
    with open(path, encoding='utf-8', newline='') as f:
        rows = list(csv.reader(f, delimiter='\t', quoting=csv.QUOTE_NONE))
    if rows[0] != HEAD:
        raise SystemExit(f'{path}: 表頭應為 {HEAD}')
    return [(r + [''] * 5)[:5] for r in rows[1:]]


def dedupe(note):
    seen = []
    for part in note.split('；'):
        part = part.strip()
        if part and part not in seen:
            seen.append(part)
    return '；'.join(seen)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--into', required=True, type=Path)
    ap.add_argument('--from', dest='sources', required=True, action='append', help='短名=路徑；先列的優先')
    ap.add_argument('--dry-run', action='store_true')
    a = ap.parse_args()
    rows = read(a.into)
    cleaned = 0
    for r in rows:
        d = dedupe(r[4])
        cleaned += d != r[4]
        r[4] = d
    have = {norm(r[0]) for r in rows}
    added = {}
    for spec in a.sources:
        name, path = spec.split('=', 1)
        n = 0
        for en, zh, cat, src, note in read(Path(path)):
            k = norm(en)
            if not k or k in have:
                continue
            have.add(k)
            if name == 'keep':  # 本作新詞表（例：試譯批次的新詞），source 與 note 照原表
                rows.append([en, zh, CATEGORY.get(cat, cat), src, note])
                n += 1
                continue
            rows.append([en, zh, CATEGORY.get(cat, cat), name, '；'.join(x for x in (f'前作來源 {src}', note) if x)])
            n += 1
        added[name] = n
    rows.sort(key=lambda r: (r[0].lower(), r[0]))
    print(f'既有 {len(rows) - sum(added.values())}，note 去重 {cleaned}，補入', '、'.join(f'{k} {v}' for k, v in added.items()),
          f'，共 {len(rows)}')
    if not a.dry_run:
        a.into.write_text('\n'.join('\t'.join(r) for r in [HEAD] + rows) + '\n', encoding='utf-8')


if __name__ == '__main__':
    main()
