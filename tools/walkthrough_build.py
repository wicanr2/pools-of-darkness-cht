#!/usr/bin/env python3
"""由自寫攻略摘要（reference/walkthrough-zh/softworld-43.md–45.md）產生 text/walkthrough.zh-TW.tsv（克萊恩英豪規格 008）。

改寫自銀色匕首之謎 tools/walkthrough_build.py。本作的來源是讀《軟體世界》第 43–45 期掃描後自寫的摘要，不含雜誌原文
（CONTEXT 決策 2026-10-23），譯名在寫摘要時已用術語表，所以不做雜誌譯名替換。
- `## 第 N 章 標題`、`## 結語` 開新章；下一行 `<!-- source: 43 期 p.60–61 -->` 是出處。
- `### 小節` 成為一段「【小節】」；正文段落以空行分隔；`- ` 開頭的要點每條一段，前面加「・」。
- `> 地圖：名稱` 另成一段「地圖：名稱」（出處沿用所在章），其後的 `> N.內容` 每條一段。
- 其他 HTML 註解刪去。
輸出欄位 chapter（1 起連號）、title、body（段落以字面 \\n 分隔）、source（「整理自軟體世界 N 期 p.X–Y」）。
用法（Docker sosb-tools 內）：python3 -I tools/walkthrough_build.py --src reference/walkthrough-zh --out text/walkthrough.zh-TW.tsv
"""
import argparse
import re
import sys
from pathlib import Path

CHAPTER = re.compile(r'^## (第 ?(\d+) ?章|結語)\s*(.*)$')
SOURCE = re.compile(r'^<!-- source: (\d+) 期 (p\.[0-9]+(?:[–-][0-9]+)?) -->$')
MAP = re.compile(r'^> ?地圖：(.+)$')
ISSUES = (43, 44, 45)
TITLE_CELLS = 40 - 3  # 面板第 1 列「攻略：{title}」共 40 格（規格 008；apps/pod/guide.go guideCols）


def cells(s):
    """面板格數：漢字一格，連續半形字兩個併一格（apps/pod PackHalfWidth）。"""
    n, run = 0, 0
    for ch in s:
        if ord(ch) < 0x80:
            run += 1
            continue
        n += (run + 1) // 2 + 1
        run = 0
    return n + (run + 1) // 2


def fit_title(title, paras):
    """標題超過格數時，把第一個全形括號裡的說明移到正文第一段。"""
    if cells(title) <= TITLE_CELLS or '（' not in title:
        return title, paras
    head, _, note = title.partition('（')
    note = note[:-1] if note.endswith('）') else note
    return head.strip(), [note] + paras


class Section:
    def __init__(self, title, source=''):
        self.title, self.source, self.paras, self.cur = title, source, [], []

    def line(self, text):
        self.cur.append(text)

    def end_para(self):
        if self.cur:
            self.paras.append(''.join(self.cur))
            self.cur = []


def build(path):
    order, chap, mp = [], None, None
    for raw in path.read_text(encoding='utf-8').splitlines():
        line = raw.strip()
        if m := SOURCE.match(line):
            if not chap:
                sys.exit(f'{path}：出處註解前沒有章標題')
            chap.source = f'整理自軟體世界 {m.group(1)} 期 {m.group(2)}'
            continue
        line = re.sub(r'<!--.*?-->', '', line).strip()
        if m := CHAPTER.match(line):
            for s in (chap, mp):
                if s:
                    s.end_para()
            head = f'第{m.group(2)}章' if m.group(2) else '結語'
            chap, mp = Section(f'{head} {m.group(3).strip()}'.strip()), None
            order.append(chap)
            continue
        if not chap:
            continue
        if not line:
            (mp or chap).end_para()
            continue
        if m := MAP.match(line):
            chap.end_para()
            mp = Section(f'地圖：{m.group(1).strip()}', chap.source)
            order.append(mp)
            continue
        if line.startswith('>'):
            if mp:
                mp.end_para()
                mp.line(line.lstrip('> ').strip())
                mp.end_para()
            continue
        if mp:  # 地圖區塊之後回到正文
            mp.end_para()
            mp = None
        if line.startswith('### '):
            chap.end_para()
            chap.line(f'【{line[4:].strip()}】')
            chap.end_para()
        elif line.startswith('- '):
            chap.end_para()
            chap.line('・' + line[2:].strip().removeprefix('要點：'))
            chap.end_para()
        else:
            chap.line(line)
    for s in order:
        s.end_para()
    for s in order:
        if s.title.startswith('地圖：') and not s.source:
            s.source = next((c.source for c in order if c.source), '')
    return order


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--src', required=True, type=Path, help='摘要目錄')
    ap.add_argument('--out', required=True, type=Path)
    a = ap.parse_args()
    rows = ['chapter\ttitle\tbody\tsource']
    n = 0
    for issue in ISSUES:
        for s in build(a.src / f'softworld-{issue}.md'):
            title, paras = fit_title(s.title, s.paras)
            if cells(title) > TITLE_CELLS:
                sys.exit(f'第 {issue} 期「{title}」標題 {cells(title)} 格，超過 {TITLE_CELLS} 格')
            body = '\\n'.join(p.replace('\t', ' ') for p in paras if p)
            if not body:
                continue
            if not s.source:
                sys.exit(f'第 {issue} 期「{s.title}」沒有出處註解')
            n += 1
            rows.append(f'{n}\t{title}\t{body}\t{s.source}')
    a.out.write_text('\n'.join(rows) + '\n', encoding='utf-8')
    print(f'{n} 段寫到 {a.out}')


if __name__ == '__main__':
    main()
