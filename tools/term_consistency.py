#!/usr/bin/env python3
"""譯名一致性掃描（batch-subagent-localization 第 6 步）：原文出現術語表的英文專名時，譯文要含對應中文。在 sosb-tools 容器內以 python3 -I 執行。

用法：term_consistency.py --glossary text/glossary.tsv --source <key、original 的 TSV> --translations <key、translation 的 TSV …> [--out TSV]
- 只看 person、place、deity、group、monster、spell、item 類；英文少於 4 個字母的不看。原文以整字、不分大小寫比對（複數 s 也算）。
- 同一段原文命中多個術語時，較長的英文先算，被它涵蓋的較短術語不再檢查（例 Zhentil Keep 涵蓋 Zhentil）。
- 輸出每個未對上的（key、英文、應有中文、譯文），並依英文彙總次數；只報告，結束碼 0。結果是待核對候選：
  代名詞、省略或語境不同的譯法可能是對的。
"""
import argparse
import csv
import re
import sys
from collections import Counter

CATS = {'person', 'place', 'deity', 'group', 'monster', 'spell', 'item'}


def read(path):
    with open(path, encoding='utf-8', newline='') as f:
        return list(csv.DictReader(f, delimiter='\t', quoting=csv.QUOTE_NONE))


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--glossary', required=True)
    ap.add_argument('--source', required=True)
    ap.add_argument('--translations', required=True, nargs='+')
    ap.add_argument('--out')
    a = ap.parse_args()
    terms = []
    for r in read(a.glossary):
        en = r['en'].strip()
        if r['category'] in CATS and len(re.sub(r'[^A-Za-z]', '', en)) >= 4 and r['zh']:
            terms.append((en, r['zh'], re.compile(r'(?<![A-Za-z])' + re.escape(en) + r"(?:S|ES|'S)?(?![A-Za-z])", re.I)))
    terms.sort(key=lambda t: -len(t[0]))
    tr = {}
    for p in a.translations:
        for r in read(p):
            if r.get('translation') and r['translation'] != '-':
                tr[r['key']] = r['translation']
    miss, per = [], Counter()
    for r in read(a.source):
        t = tr.get(r['key'])
        if t is None:
            continue
        text = r['original']
        for en, zh, rx in terms:
            m = rx.search(text)
            if not m:
                continue
            text = text[:m.start()] + ' ' * (m.end() - m.start()) + text[m.end():]
            if zh not in t:
                miss.append((r['key'], en, zh, t))
                per[en] += 1
    lines = ['key\ten\tzh\ttranslation'] + ['\t'.join(x) for x in miss]
    if a.out:
        open(a.out, 'w', encoding='utf-8').write('\n'.join(lines) + '\n')
    print(f'術語 {len(terms)}、譯文 {len(tr)}；未對上 {len(miss)} 處、{len(per)} 個術語')
    for en, n in per.most_common(40):
        print(f'  {n:4d} {en}', file=sys.stderr)


if __name__ == '__main__':
    main()
