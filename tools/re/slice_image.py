"""從「PSP＋映像」的記憶體傾印切出映像：檢查 PSP 開頭 CD 20，丟掉前 256 bytes。

用法：slice_image.py <psp-image.bin> <輸出>；印出輸出長度與 SHA-256。
"""
import hashlib
import sys

b = open(sys.argv[1], "rb").read()
if b[:2] != b"\xcd\x20":
    raise SystemExit(f"開頭不是 PSP（{b[:2].hex()}）：落點與 G2 的程式鏈不同")
img = b[0x100:]
open(sys.argv[2], "wb").write(img)
print(len(img), hashlib.sha256(img).hexdigest(), sys.argv[2])
