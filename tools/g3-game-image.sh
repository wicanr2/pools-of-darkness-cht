#!/usr/bin/env bash
# 取 GAME.EXE 解開後的映像（docs/re/004）。GAME.EXE 以 EXEPACK 壓縮，檔案裡的字串會被 RLE 改寫；
# 這裡用 dosgolem 跑原版解壓 stub，從 G2 收據的 credits-300m.state 送 p 到主選單 #400,000,000，
# 傾印 PSP 169Fh 與其後 11C5h 段（EXEPACK 標頭的 dest_len）的映像，切掉 PSP 存到 workplace/text-src/game-image.bin。
# 映像裡的段值已重定位（+16AFh），字串不受影響；落點依 G2 的程式鏈（常駐驅動＋-root-tsr-shrink）而定。
# 前置：tools/g2-startup-receipt.sh 跑過（workplace/probe/g2-receipt/）。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
R="$ROOT/workplace/probe/g2-receipt"
OUT="$ROOT/workplace/probe/g3-image"
test -f "$R/credits-300m.state" || { echo "缺 $R/credits-300m.state（先跑 tools/g2-startup-receipt.sh）" >&2; exit 1; }
rm -rf "$OUT"
mkdir -p "$OUT/scratch" "$ROOT/workplace/text-src"
cp "$R/scratch/"* "$OUT/scratch/"
cp "$R/credits-300m.state" "$OUT/"
PROBE_EXE=ADLIB.COM "$ROOT/tools/probe.sh" "$OUT" -scratch-dir /out/scratch -root-tsr-shrink \
  -load-state /out/credits-300m.state -steps 400001000 -bios-keys 'p' -bios-key-from 330000000 -bios-key-every 50000000 \
  -dump-mem '169F0-28740:/out/psp-image.bin' > "$OUT/run.log" 2>&1
timeout 120 docker run --rm --network none -u "$(id -u):$(id -g)" --memory 256m --cpus 1 --pids-limit 16 \
  --log-opt max-size=10m --log-opt max-file=3 -v "$ROOT":/in -w /in sosb-tools:bookworm-r1 \
  python3 -I tools/re/slice_image.py workplace/probe/g3-image/psp-image.bin workplace/text-src/game-image.bin
