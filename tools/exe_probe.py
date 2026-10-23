"""G1 執行檔探查：MZ 標頭、載入映像後的附加資料、overlay 魔數與關鍵字串。

用法（Docker sosb-tools 內）：python3 -I tools/exe_probe.py <遊戲目錄> [關鍵字 ...]
未給關鍵字時用預設清單（版本、設定檔、顯示與音效、查驗題）。只讀檔，不執行。
"""
import re, struct, sys
from pathlib import Path

root = Path(sys.argv[1])
keys = [k.lower() for k in sys.argv[2:]] or [
    "version", "copyright", ".cfg", ".drv", "vga", "ega", "tandy", "adlib", "sound blaster",
    "journal", "word #", "rule book", "page", "line", "save", "demo", "control", "vace"]

def strings(b, n=4):
    for m in re.finditer(rb"[\x20-\x7e]{%d,}" % n, b):
        yield m.start(), m.group().decode("ascii")

for p in sorted(root.iterdir()):
    if not p.is_file() or p.suffix.upper() not in (".EXE", ".COM", ".OVR", ".DRV", ".BAT"):
        continue
    b = p.read_bytes()
    head = b[:4]
    line = f"== {p.name} size={len(b)} magic={head!r}"
    if b[:2] in (b"MZ", b"ZM"):
        last, pages, nrel, hdr = struct.unpack_from("<HHHH", b, 2)
        image = pages * 512 - (512 - last if last else 0)
        line += f" MZ image={image} header={hdr * 16} relocs={nrel} extra={len(b) - image}"
        if len(b) > image:
            tail = b[image:image + 4]
            line += f" tail_magic={tail!r}"
            if b[image:image + 2] == b"\xfb\x52":
                line += " borland_debug=0x52FB"
    print(line)
    if p.suffix.upper() == ".BAT":
        print(b.decode("ascii", "replace"))
        continue
    for off, s in strings(b):
        low = s.lower()
        if any(k in low for k in keys):
            print(f"  {off:06X} {s[:120]}")
