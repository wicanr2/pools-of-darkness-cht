#!/usr/bin/env python3
"""由 GAME.EXE 解開後的映像與 GAME.OVR 靜態抽取介面字串（Pascal 長度前綴），產生不含原文的 catalog。

移植自銀色匕首之謎 tools/ui_text_catalog.py。GAME.EXE 是 EXEPACK，檔案裡的字串會被 RLE 改寫，
所以改讀 tools/g3-game-image.sh 傾印的映像（docs/re/004）。判準照舊：長度 1–80、全部可列印 ASCII、
含兩個以上連續英文字母。這是啟發式，會收到程式碼中的偶然位元組；執行期以整串雜湊比對，雜訊永遠不會命中。
key 與前作相同（ui.<SHA-256 前 12 碼>），前作譯文可直接比對。

輸出：
  --events  text/ui-text-events.tsv（key、長度、SHA-256、來源；不含原文）
  --source  workplace 內的 key→原文對照（只供翻譯，不入版控）
  --anchors 選用：執行期抓到的字串（每行一則），檢查哪些不在 catalog
"""
import argparse
import hashlib
import re
from pathlib import Path


def extract(sources):
    found = {}
    for name, d in sources:
        for i in range(len(d) - 1):
            n = d[i]
            if 1 <= n <= 80 and i + 1 + n <= len(d):
                s = d[i + 1:i + 1 + n]
                if all(32 <= c < 127 for c in s) and re.search(rb'[A-Za-z]{2}', s):
                    found.setdefault(s.decode('ascii'), []).append(f'{name}:{i}')
    return found


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--image', required=True, type=Path, help='GAME.EXE 解開後的映像')
    ap.add_argument('--ovr', required=True, type=Path, help='GAME.OVR')
    ap.add_argument('--events', required=True, type=Path)
    ap.add_argument('--source', required=True, type=Path)
    ap.add_argument('--anchors', type=Path)
    a = ap.parse_args()
    found = extract([('GAME.EXE-image', a.image.read_bytes()), ('GAME.OVR', a.ovr.read_bytes())])
    ev = ['event_key\toriginal_length\toriginal_sha256\tsources']
    src = ['key\toriginal\tsources']
    for t, locs in found.items():
        raw = t.encode('ascii')
        key = f'ui.{hashlib.sha256(raw).hexdigest()[:12]}'
        ev.append(f'{key}\t{len(raw)}\t{hashlib.sha256(raw).hexdigest()}\t{";".join(locs[:4])}')
        src.append(f'{key}\t{t}\t{";".join(locs[:4])}')
    keys = [l.split('\t')[0] for l in ev[1:]]
    if len(keys) != len(set(keys)):
        raise SystemExit('duplicate key')
    a.events.write_text('\n'.join(ev) + '\n', encoding='utf-8')
    a.source.write_text('\n'.join(src) + '\n', encoding='utf-8')
    print(f'{len(ev) - 1} strings')
    if a.anchors:
        lines = [l for l in a.anchors.read_text(encoding='utf-8').split('\n') if l]
        miss = [l for l in lines if l not in found]
        print(f'anchors: {len(lines) - len(miss)} hit, {len(miss)} miss')
        for m in miss:
            print('  miss', repr(m))


if __name__ == '__main__':
    main()
