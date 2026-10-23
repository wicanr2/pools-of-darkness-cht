#!/usr/bin/env python3
"""16 位元 x86 反組譯一段記憶體傾印（docs/re/008）。

用法（Docker sc-cht-disasm 內）：python3 -I tools/re/dis16.py <傾印檔> <段 hex> <起始偏移 hex> <結束偏移 hex>
傾印檔是從線性 0 起的記憶體（探針 `-dump-mem-at '<步數>:lin:0:655360:<檔名>'`）；輸出「段:偏移 bytes 指令」。
"""
import sys

import capstone

f, seg, start, end = sys.argv[1], int(sys.argv[2], 16), int(sys.argv[3], 16), int(sys.argv[4], 16)
b = open(f, 'rb').read()
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_16)
base = seg * 16
for i in md.disasm(b[base + start:base + end], start):
    print("%04X:%04X  %-20s %s %s" % (seg, i.address, i.bytes.hex(), i.mnemonic, i.op_str))
