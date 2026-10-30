#!/usr/bin/env python3
"""在 Docker 對齊自寫攻略的既定專名，--check 只檢查，不寫入。

只處理完整英文詞組與 glossary 的精確對應；英文括註保留。
已知物品的加值保留置後，怪物複數沿同一譯名；指令與未知詞組不作部分替換。
"""
import argparse
import csv
import hashlib
from pathlib import Path
import re

PHRASE = re.compile(r"[A-Za-z][A-Za-z0-9' .+–-]*")
DISPLAY_CATEGORIES = {'place', 'person', 'deity', 'group', 'item', 'monster', 'spell'}
AMBIGUOUS_WORDS = {'silver', 'steel', 'copper', 'iron', 'platinum', 'dragon'}
ITEM_ALIASES = {'helmet': 'helm', 'elfin chain': 'elven chain',
                'javelins': 'javelin', 'elixirs of youth': 'elixir of youth'}


def load_terms(path):
    with path.open(encoding='utf-8', newline='') as f:
        rows = list(csv.DictReader(f, delimiter='\t'))
    terms = {r['en'].casefold(): r['zh'] for r in rows
             if r['category'] in DISPLAY_CATEGORIES and r['zh']}
    items = {}
    for r in rows:
        if r['category'] == 'item' or r['en'].casefold() == 'shield':
            items[r['en'].casefold()] = r['zh']
        if r['category'] != 'monster' or not r['zh']:
            continue
        for alias in r['en'].split(' / '):
            key = alias.casefold()
            terms.setdefault(key, r['zh'])
            if not key.endswith('s'):
                plural = key + ('es' if key.endswith(('ch', 'sh', 'x', 'z')) else 's')
                terms.setdefault(plural, r['zh'])
    # Bolt is already an exact game-name translation. Bind the lookup to its
    # original English hash rather than introducing a second Chinese term.
    names = path.parent / 'name.zh-TW.tsv'
    if names.is_file():
        with names.open(encoding='utf-8', newline='') as f:
            catalog = {r['key']: r['translation'] for r in csv.DictReader(f, delimiter='\t')}
        key = 'name.' + hashlib.sha256(b'Bolt').hexdigest()[:12]
        if key in catalog:
            items['bolt'] = items['bolts'] = catalog[key]
    for alias, canonical in ITEM_ALIASES.items():
        if canonical in items:
            items[alias] = items[canonical]
    # Only known whole item names admit an enchantment suffix. Ordinary words,
    # unknown names and player commands never enter this generated dictionary.
    for key, zh in items.items():
        terms.setdefault(key, zh)
        for bonus in range(1, 6):
            terms.setdefault(f'{key} +{bonus}', f'{zh} +{bonus}')
    return {key: value for key, value in terms.items() if key not in AMBIGUOUS_WORDS}


def align(text, terms):
    # Preserve the entire parenthesized English annotation, including phrases
    # after punctuation inside it. A bare material can also be part of a key.
    annotations = set()
    depth = 0
    for i, ch in enumerate(text):
        if ch in '（(':
            depth += 1
        if depth:
            annotations.add(i)
        if ch in '）)' and depth:
            depth -= 1

    def replace(m):
        word = m.group().rstrip()
        prefix = ''
        if re.match(r'^[A-G]\. ', word):
            prefix, word = word[:3], word[3:]
        if m.start() in annotations:
            return m.group()
        zh = terms.get(word.casefold())
        if not zh:
            return m.group()
        if text[:m.start()].rstrip().endswith(zh):
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


def check_boundaries(terms):
    cases = [
        ('取得 Battle axe +4。', '取得戰斧 +4。'),
        ('找到 Javelins +2。', '找到標槍 +2。'),
        ('得到 Helmet +4。', '得到頭盔 +4。'),
        ('獲得 Elixirs of Youth。', '獲得青春靈藥。'),
        ('得到 30 Arrows +2。', '得到 30 箭 +2。'),
        ('A. Battle axe +4', 'A. 戰斧 +4'),
        ('選 LOOK、Search、Duck。', '選 LOOK、Search、Duck。'),
        ('鑰匙有 silver、iron、steel 和 copper key。', '鑰匙有 silver、iron、steel 和 copper key。'),
        ('魔法檢哨官 Dragon。', '魔法檢哨官 Dragon。'),
        ('未知 Rare Long Sword +4。', '未知 Rare Long Sword +4。'),
        ('長劍（Long Sword +4）。', '長劍（Long Sword +4）。'),
        ('魔法杖（Wand，內容為魔法飛彈 Magic Missile）。', '魔法杖（Wand，內容為魔法飛彈 Magic Missile）。'),
        ('魔法飛彈 Magic Missile。', '魔法飛彈 Magic Missile。'),
        ('註記（原文 (Long Sword) 與 Ward of Safe Passage）。', '註記（原文 (Long Sword) 與 Ward of Safe Passage）。'),
    ]
    for source, expected in cases:
        actual = align(source, terms)
        if actual != expected:
            raise SystemExit(f'攻略詞彙邊界失配：{source!r} → {actual!r}，預期 {expected!r}')
    print(f'攻略詞彙邊界：{len(cases)} 例通過')


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--src', type=Path, default=Path('reference/walkthrough-zh'))
    ap.add_argument('--glossary', type=Path, default=Path('text/glossary.tsv'))
    ap.add_argument('--check', action='store_true')
    ap.add_argument('--self-check', action='store_true', help='驗證指令、括註、數量及同形異義邊界，不寫入')
    a = ap.parse_args()
    terms = load_terms(a.glossary)
    if a.self_check:
        check_boundaries(terms)
        return
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
