#!/usr/bin/env python3
"""由 mocagh 英文手札的 OCR 逐字座標組出各則草稿（docs/re/005）。改寫自銀色匕首之謎 tools/journal_en_ocr.py。

輸入 `pdftotext -bbox-layout` 的 XHTML，每個檔一個 PDF 頁（`p<PDF 頁>.html`，手札在 PDF 第 15–25 頁）。
每個 PDF 頁是一個跨頁：左頁印刷頁碼 2p−5、右頁 2p−4（PDF 第 15 頁＝手札 p.25、26，頁碼在頁首花邊裡，OCR 讀不到）。
每頁依行首 x 分成四欄（界線見 pages()）、欄內依 y 排序，遇到 `Journal Entry N` 開新的一則。標題取其後連續的全大寫行
（不含小寫字母、字母占多數），其餘是正文。正文含地圖圖例（`KEY` 與 `Impassable`）時 note 標地圖。
OCR 草稿未校對：錯譯對照時對頁面圖校讀（workplace/journal/en/img/）。
輸出 TSV：entry、page、title、body、note。
"""
import argparse
import html
import re
import sys
from pathlib import Path

PAGE = re.compile(r'<page width="([\d.]+)"')
LINE = re.compile(r'<line xMin="([\d.]+)" yMin="([\d.]+)"[^>]*>(.*?)</line>', re.S)
WORD = re.compile(r'<word[^>]*>(.*?)</word>')
HEAD = re.compile(r'Journal\s*Entry\s*(\d+)', re.I)


def is_title(words):
    """標題是小型大寫字，OCR 會混進少數小寫（`5CRYIN(j`、`ELMINSTER's`）；字母八成以上大寫才算。"""
    s = ' '.join(words)
    letters = [c for c in s if c.isalpha()]
    upper = sum(c.isupper() for c in letters)
    return len(letters) >= 3 and upper >= 0.8 * len(letters) and len(letters) >= 0.6 * len(s.replace(' ', ''))


def pages(path, pdf_page):
    xhtml = path.read_text(encoding='utf-8')
    w = float(PAGE.search(xhtml).group(1))
    lines = []
    for m in LINE.finditer(xhtml):
        x, y = float(m.group(1)), float(m.group(2))
        words = [html.unescape(t) for t in WORD.findall(m.group(3))]
        # 欄的行首 x 實測聚在 0–20、190–210、400–410、580–590（PDF 第 15–25 頁），四等分會把第二欄切進第一欄。
        col = sum(x >= b * w / 767.28 for b in (150, 370, 560))
        printed = 2 * pdf_page - 5 + (1 if col >= 2 else 0)
        lines.append((col, y, printed, words))
    lines.sort(key=lambda t: (t[0], t[1]))
    out = []
    for _, _, printed, ws in lines:
        # 第 38 則的標題列被拆成「JOURNAL」「ENTRY38」兩行：併回一行。
        if out and out[-1][1] == ['JOURNAL'] and ws and ws[0].upper().startswith('ENTRY'):
            out[-1] = (out[-1][0], ['JOURNAL', 'ENTRY', ws[0][5:]] + ws[1:] if len(ws[0]) > 5 else ['JOURNAL'] + ws)
            continue
        out.append((printed, ws))
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--bbox', required=True, type=Path, nargs='+', help='p<PDF 頁>.html，依頁序')
    ap.add_argument('--out', required=True, type=Path)
    ap.add_argument('--last', type=int, default=100)
    a = ap.parse_args()
    got = {}
    cur, in_title = None, False
    for path in a.bbox:
        pdf_page = int(re.search(r'p(\d+)', path.stem).group(1))
        for printed, ws in pages(path, pdf_page):
            text = ' '.join(ws)
            m = HEAD.search(text)
            if m:
                after = text[m.end():].split()
                cur, in_title = int(m.group(1)), True
                got[cur] = {'page': printed, 'title': [], 'body': list(after)}
                continue
            if cur is None:
                continue
            e = got[cur]
            if in_title and not e['body'] and is_title(ws):
                e['title'].extend(ws)
                continue
            in_title = False
            e['body'].extend(ws)
    rows, missing = [], []
    for n in range(1, a.last + 1):
        e = got.get(n)
        if e is None:
            missing.append(n)
            continue
        body = re.sub(r'(\w)- (\w)', r'\1\2', ' '.join(e['body']))  # 行尾斷字
        note = 'OCR 草稿未校對'
        if 'KEY' in e['body'] and any(w.startswith('Impass') for w in e['body']):
            note += '；含地圖圖例'
        rows.append((n, e['page'], ' '.join(e['title']), body, note))
    with open(a.out, 'w', encoding='utf-8', newline='') as f:
        f.write('entry\tpage\ttitle\tbody\tnote\n')
        for row in rows:
            f.write('\t'.join(str(c).replace('\t', ' ') for c in row) + '\n')
    print(f'{len(rows)} 則；缺 {missing}', file=sys.stderr)


if __name__ == '__main__':
    main()
