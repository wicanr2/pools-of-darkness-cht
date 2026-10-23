#!/usr/bin/env python3
"""珍157 冒險者手冊魔法寶物（p.21–22）、怪物（p.23–36）、法術（p.37–54）的名稱對照（#5、#6；比照銀色匕首之謎 tools/zh056_names.py）。
在 sosb-tools 容器內以 python3 -I 執行。

draft：由跨頁 OCR 抽「N. 中文名（English）」「(N)中文名（English）」候選，給逐張對圖轉錄用（OCR 會漏掉只留英文或被拆行的條目）。
  zh157_names.py draft --spreads <左頁碼>=<OCR JSON> … --out <TSV>
check：驗證 text/zh157-names.tsv（逐張對頁面圖轉錄、裁定後的名稱表）與術語表一致；--apply 時把 adopted 寫進術語表。
  zh157_names.py check --names text/zh157-names.tsv --glossary text/glossary.tsv --monsters <monster-names.tsv> --ui <ui-source.tsv> [--apply] [--report TSV]
  - 比對鍵是 game_en（遊戲內英文，'-' 表示同 en），英文只留字母、不分大小寫；查不到時去掉字尾 s 再試。
  - 同一個鍵在表中出現多次時 adopted 必須相同；adopted 空白（珍157 留英文、尚未裁定）不檢查。
  - decision 只能是 同名、採珍157、珍157誤譯、珍157不一致、珍157錯字、留英文。
  - --apply：術語表沒有的鍵新增一列，譯名不同的改成 adopted，note 記珍157 頁碼、原書用字、裁定與前作譯名。
  - --report：每列附術語表譯名與遊戲內證據（MON1CHA 區塊或 UI 字串是否有這個英文）。
不一致時結束碼 1。
"""
import argparse
import csv
import re
import sys
from pathlib import Path

DECISIONS = {'同名', '採珍157', '珍157誤譯', '珍157不一致', '珍157錯字', '留英文'}
GLOSSARY_HEAD = ['en', 'zh', 'category', 'source', 'note']
CJK = r'㐀-鿿'
LINE = re.compile(rf'^\s*(?:[(（]?\s*\d{{1,2}}\s*[)）.．]?)?\s*([{CJK}·・]{{1,14}})\s*[（(]\s*([A-Za-z][^）)]*?)\s*([)）]|$)')


def norm(t):
    return re.sub(r'[^a-z]', '', t.lower().replace('1', 'l'))


def read(path):
    with open(path, encoding='utf-8', newline='') as f:
        return list(csv.DictReader(f, delimiter='\t', quoting=csv.QUOTE_NONE))


def lines(bs):
    """文字框依中線 y 併行，行內依 x 排序。"""
    rows = []
    for t, (x0, y0, x1, y1) in sorted(bs, key=lambda b: ((b[1][1] + b[1][3]) / 2, b[1][0])):
        mid, half = (y0 + y1) / 2, (y1 - y0) / 2
        if rows and abs(rows[-1][0] - mid) < half:
            rows[-1][1].append((x0, t))
        else:
            rows.append([mid, [(x0, t)]])
    return [' '.join(t for _, t in sorted(parts)) for _, parts in rows]


def draft(a):
    sys.path.insert(0, str(Path(__file__).resolve().parent))
    from zh_ocr_journal import boxes, gutter  # noqa: E402
    out = ['page\tzh157\ten']
    for spec in a.spreads:
        left, path = spec.split('=', 1)
        bs = boxes(path)
        h = max(b[3] for _, b in bs)
        g = gutter(bs)
        bs = [x for x in bs if 0.06 * h < x[1][1] and x[1][3] < 0.93 * h]
        for side in (0, 1):
            ls = lines([x for x in bs if ((x[1][0] + x[1][2]) / 2 < g) == (side == 0)])
            for i, t in enumerate(ls):
                m = LINE.match(t)
                if not m:
                    continue
                en = m.group(2)
                if not m.group(3) and i + 1 < len(ls):
                    nxt = re.match(r'^\s*([A-Za-z0-9\' ’.\-]+)\s*[)）]', ls[i + 1])
                    if nxt:
                        en = f'{en} {nxt.group(1)}'
                out.append(f'{int(left) + side}\t{m.group(1)}\t{" ".join(en.split())}')
    a.out.write_text('\n'.join(out) + '\n', encoding='utf-8')
    print(f'{len(out) - 1} 個候選（OCR 草稿，未校對）', file=sys.stderr)


def check(a):
    names = read(a.names)
    with open(a.glossary, encoding='utf-8', newline='') as f:
        grows = [(r + [''] * 5)[:5] for r in csv.reader(f, delimiter='\t', quoting=csv.QUOTE_NONE)]
    if grows[0] != GLOSSARY_HEAD:
        raise SystemExit(f'{a.glossary}: 表頭應為 {GLOSSARY_HEAD}')
    grows = grows[1:]
    gl = {}
    for r in grows:
        gl.setdefault(norm(r[0]), r)
    mon = {norm(r['name']): r['block'] for r in read(a.monsters) if r['name']}
    ui = {norm(r['original']) for r in read(a.ui)} if a.ui else set()

    def find(table, k):
        return table.get(k) or (table.get(k[:-1]) if k.endswith('s') else None)

    bad, want, report = [], {}, ['page\tkind\ten\tzh157\tadopted\tglossary\tglossary_source\tdecision\tin_game']
    for i, r in enumerate(names, 2):
        if r['decision'] not in DECISIONS:
            bad.append(f'第 {i} 列 decision {r["decision"]!r} 不在 {sorted(DECISIONS)}')
        key = r['game_en'] if r['game_en'] not in ('', '-') else r['en']
        k = norm(key)
        if r['adopted']:
            prev = want.get(k)
            if prev and prev[1] != r['adopted']:
                bad.append(f'{key}：p.{prev[2]} 採「{prev[1]}」、p.{r["page"]} 採「{r["adopted"]}」')
            want.setdefault(k, (key, r['adopted'], r['page'], r))
        g = find(gl, k)
        blk = find(mon, k)
        ev = f'MON1CHA {blk}' if blk else ('UI' if k in ui or (k.endswith('s') and k[:-1] in ui) else '-')
        report.append('\t'.join([r['page'], r['kind'], key, r['zh157'], r['adopted'], g[1] if g else '', g[3] if g else '',
                                 r['decision'], ev]))
    changed, added = 0, 0
    for k, (key, zh, page, r) in want.items():
        g = find(gl, k)
        if g and g[1] == zh:
            continue
        if not a.apply:
            bad.append(f'{key}：術語表{"作「" + g[1] + "」" if g else "沒有"}，名稱表採「{zh}」')
            continue
        zh157 = '' if r['zh157'] == '-' else f'作「{r["zh157"]}」'
        note = '；'.join(x for x in (f'珍157 冒險者手冊 p.{page}{zh157}', r['decision'], r['reason']) if x)
        src = 'zh157' if r['decision'] in ('同名', '採珍157') else '自譯'
        if g:
            note += f'；原「{g[1]}」（{g[3]}）'
            g[1], g[3], g[4] = zh, src, note
            changed += 1
        else:
            row = [key, zh, r['kind'], src, note]
            grows.append(row)
            gl[k] = row
            added += 1
    if a.report:
        a.report.write_text('\n'.join(report) + '\n', encoding='utf-8')
    if a.apply and (changed or added):
        grows.sort(key=lambda r: (r[0].lower(), r[0]))
        a.glossary.write_text('\n'.join('\t'.join(r) for r in [GLOSSARY_HEAD] + grows) + '\n', encoding='utf-8')
    print(f'名稱 {len(names)} 列、採用 {len(want)} 個鍵；術語表改 {changed}、增 {added}', file=sys.stderr)
    if bad:
        print('\n'.join(bad), file=sys.stderr)
        sys.exit(1)


def main():
    ap = argparse.ArgumentParser()
    sub = ap.add_subparsers(dest='cmd', required=True)
    d = sub.add_parser('draft')
    d.add_argument('--spreads', required=True, nargs='+')
    d.add_argument('--out', required=True, type=Path)
    c = sub.add_parser('check')
    c.add_argument('--names', required=True, type=Path)
    c.add_argument('--glossary', required=True, type=Path)
    c.add_argument('--monsters', required=True, type=Path)
    c.add_argument('--ui', type=Path)
    c.add_argument('--apply', action='store_true')
    c.add_argument('--report', type=Path)
    a = ap.parse_args()
    draft(a) if a.cmd == 'draft' else check(a)


if __name__ == '__main__':
    main()
