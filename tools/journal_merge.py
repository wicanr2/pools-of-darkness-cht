#!/usr/bin/env python3
"""合併手札校讀批次（workplace/journal/fix/bNN/，docs/re/005）。

用法（Docker 內）：python3 -I tools/journal_merge.py <fix 目錄> [--expect 1-100] [--skip 13]
檢查：每個 TSV 欄名正確、無 tab／換行殘留；各則在每種檔裡恰好一次；錯譯表的 zh157 是珍157 逐字稿該則的子字串、
fix 是正式譯文該則的子字串、key 不重複。任一不成立就列出並以結束碼 1 結束，不寫任何輸出。
通過才寫：
  workplace/journal/en/journal-en.tsv、workplace/journal/zh157/journal-zh157.tsv（逐字稿，不入版控）
  text/journal.zh-TW.tsv、text/errata.tsv（正式）
  workplace/journal/terms-merged.tsv（術語候選，含批次欄）
"""
import argparse
import csv
import sys
from pathlib import Path

SPEC = {
    'en.tsv': ['entry', 'page', 'title', 'body'],
    'zh157.tsv': ['entry', 'page', 'title', 'body'],
    'journal.tsv': ['entry', 'title', 'body', 'source'],
    'errata.tsv': ['key', 'entry', 'page', 'en', 'zh157', 'fix', 'category', 'reason', 'evidence', 'source'],
    'terms.tsv': ['en', 'zh', 'category', 'source', 'note'],
}


def read(path, cols, bad):
    with open(path, encoding='utf-8', newline='') as f:
        head = f.readline().rstrip('\n').split('\t')
        if head != cols:
            bad.append(f'{path}：欄名 {head}，要 {cols}')
            return []
        rows = []
        for i, line in enumerate(f, 2):
            line = line.rstrip('\n')
            if not line:
                continue
            parts = line.split('\t')
            if len(parts) != len(cols):
                bad.append(f'{path}:{i} 欄數 {len(parts)}，要 {len(cols)}')
                continue
            rows.append(dict(zip(cols, parts)))
        return rows


def write(path, cols, rows):
    path.parent.mkdir(parents=True, exist_ok=True)
    with open(path, 'w', encoding='utf-8', newline='') as f:
        f.write('\t'.join(cols) + '\n')
        for r in rows:
            f.write('\t'.join(r[c] for c in cols) + '\n')


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('fix', type=Path)
    ap.add_argument('--expect', default='1-100')
    ap.add_argument('--skip', default='13', help='不存在的則號，逗號分隔（第 13 則英文與珍157 都沒有，docs/re/005）')
    a = ap.parse_args()
    lo, hi = map(int, a.expect.split('-'))
    skip = {s for s in a.skip.split(',') if s}
    bad = []
    data = {k: [] for k in SPEC}
    for b in sorted(p for p in a.fix.iterdir() if p.is_dir()):
        for name, cols in SPEC.items():
            p = b / name
            if not p.exists():
                bad.append(f'{p} 不存在')
                continue
            for r in read(p, cols, bad):
                r['_batch'] = b.name
                data[name].append(r)
    by = {}
    for name in ('en.tsv', 'zh157.tsv', 'journal.tsv'):
        seen = {}
        for r in data[name]:
            n = r['entry']
            if n in seen:
                bad.append(f'{name} 第 {n} 則重複（{seen[n]}、{r["_batch"]}）')
            seen[n] = r['_batch']
        want = {str(n) for n in range(lo, hi + 1)} - skip
        if set(seen) != want:
            bad.append(f'{name} 缺 {sorted(want - set(seen), key=int)}，多 {sorted(set(seen) - want, key=int)}')
        by[name] = {r['entry']: r for r in data[name]}
    keys = set()
    for r in data['errata.tsv']:
        k, n = r['key'], r['entry']
        if k in keys:
            bad.append(f'錯譯 key {k} 重複')
        keys.add(k)
        z, j = by['zh157.tsv'].get(n), by['journal.tsv'].get(n)
        if z is None or j is None:
            bad.append(f'{k}：第 {n} 則不在逐字稿或譯文')
            continue
        if r['zh157'] and r['zh157'] not in z['title'] + z['body']:
            bad.append(f'{k}：zh157「{r["zh157"]}」不在珍157 逐字稿第 {n} 則')
        if r['fix'] and r['fix'] not in j['title'] + j['body']:
            bad.append(f'{k}：fix「{r["fix"]}」不在譯文第 {n} 則')
    if bad:
        print('\n'.join(bad), file=sys.stderr)
        sys.exit(1)
    order = lambda rows: sorted(rows, key=lambda r: int(r['entry']))
    write(Path('workplace/journal/en/journal-en.tsv'), SPEC['en.tsv'], order(data['en.tsv']))
    write(Path('workplace/journal/zh157/journal-zh157.tsv'), SPEC['zh157.tsv'], order(data['zh157.tsv']))
    write(Path('text/journal.zh-TW.tsv'), SPEC['journal.tsv'], order(data['journal.tsv']))
    write(Path('text/errata.tsv'), SPEC['errata.tsv'], sorted(data['errata.tsv'], key=lambda r: (int(r['entry']), r['key'])))
    terms = data['terms.tsv']
    for r in terms:
        r['batch'] = r.pop('_batch')
    write(Path('workplace/journal/terms-merged.tsv'), SPEC['terms.tsv'] + ['batch'], terms)
    print(f"手札 {len(data['journal.tsv'])} 則、錯譯 {len(data['errata.tsv'])} 條、術語候選 {len(terms)} 列")


if __name__ == '__main__':
    main()
