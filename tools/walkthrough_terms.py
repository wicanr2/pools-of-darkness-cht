#!/usr/bin/env python3
"""在 Docker 對齊自寫攻略的既定專名，--check 只檢查，不寫入。

只處理完整英文詞組與 glossary 的精確對應；英文括註保留。
指令、未知詞組與一般物品詞不作部分替換，避免誤改玩家需輸入的字。
"""
import argparse
import csv
from pathlib import Path
import re

PHRASE = re.compile(r"[A-Za-z][A-Za-z0-9' .+–-]*")


def align(text, terms):
    def replace(m):
        word = m.group().rstrip()
        prefix = ''
        if re.match(r'^[A-G]\. ', word):
            prefix, word = word[:3], word[3:]
        if m.start() and text[m.start()-1] in '（(':
            return m.group()
        zh = terms.get(word.casefold())
        if not zh:
            return m.group()
        # 已存在的反向括註留待人工判讀，不產生「中文（中文）」
        if text[m.end():].startswith('（' + zh + '）'):
            return m.group()
        return prefix + zh + m.group()[len(m.group().rstrip()):]
    result = PHRASE.sub(replace, text)
    # 英文移除後，正文的中文之間不留下分詞空格；標題仍保留層級間距。
    return '\n'.join(line if line.lstrip().startswith('#') else
                     re.sub(r'(?<=[\u3400-\u9fff，。：；、！？]) +(?=[\u3400-\u9fff，。：；、！？])', '', line)
                     for line in result.split('\n'))


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--src', type=Path, default=Path('reference/walkthrough-zh'))
    ap.add_argument('--glossary', type=Path, default=Path('text/glossary.tsv'))
    ap.add_argument('--check', action='store_true')
    a = ap.parse_args()
    with a.glossary.open(encoding='utf-8', newline='') as f:
        terms = {r['en'].casefold(): r['zh'] for r in csv.DictReader(f, delimiter='\t')
                 if r['category'] in {'place', 'person', 'deity', 'group'} and r['zh']}
    changed = []
    for path in sorted(a.src.glob('softworld-*.md')):
        old = path.read_text(encoding='utf-8')
        new = align(old, terms)
        if old != new:
            changed.append(str(path))
            if not a.check:
                path.write_text(new, encoding='utf-8')
    print(f'攻略專名：{len(terms)} 個詞組，{len(changed)} 份待對齊' if a.check
          else f'攻略專名：{len(terms)} 個詞組，更新 {len(changed)} 份')
    if a.check and changed:
        raise SystemExit('\n'.join(changed))


if __name__ == '__main__':
    main()
