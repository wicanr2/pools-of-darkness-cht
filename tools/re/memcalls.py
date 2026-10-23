"""在記憶體傾印（線性 0 起）裡找指向某些線性位址的遠呼叫 `9A off seg`。

用法：memcalls.py <mem.bin> <目標線性 hex> [...]
輸出每個目標的 段:位移 組合、呼叫次數與前幾個呼叫位置（線性）。
"""
import struct
import sys

mem = open(sys.argv[1], "rb").read()
targets = {int(t, 16): t for t in sys.argv[2:]}
found = {t: {} for t in targets}
i = mem.find(b"\x9a")
while i >= 0 and i + 5 <= len(mem):
    off, seg = struct.unpack_from("<HH", mem, i + 1)
    lin = seg * 16 + off
    if lin in found:
        found[lin].setdefault((seg, off), []).append(i)
    i = mem.find(b"\x9a", i + 1)
for t, hits in found.items():
    if not hits:
        print(f"{t:05X} 無")
    for (seg, off), at in sorted(hits.items()):
        print(f"{t:05X} {seg:04X}:{off:04X} ×{len(at)} 例 {', '.join(f'{a:05X}' for a in at[:6])}")
