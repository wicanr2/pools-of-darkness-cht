"""把 -dump-mem-at 傾印的 B8000 文字頁（80×25，字元＋屬性）印成文字，去掉空白尾。

用法（Docker 內）：python3 -I tools/b8000_text.py <傾印檔> [...]
"""
import sys

for path in sys.argv[1:]:
    b = open(path, "rb").read()
    print(f"== {path}")
    for row in range(25):
        line = bytes(b[(row * 80 + c) * 2] for c in range(80) if (row * 80 + c) * 2 < len(b))
        s = line.decode("cp437").rstrip()
        if s:
            print(f"{row:2d} {s}")
