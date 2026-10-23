#!/usr/bin/env bash
# pod-play 無頭冒煙（G6 冷開機路徑）：空存檔目錄冷開機，照 START.BAT 以 snd.com 常駐、接 CONTROL.EXE（dosgolem 規格 432）。
# 片頭送空白、製作群後送 p 進建立隊伍主選單，結束時輸出中文與原版合成截圖並寫印字盤點。
# 檢查 pool4.cfg 與原版首次設定相同、snd.com 與 ADLIB.COM 相同（docs/re/003）。
# 用法：tools/play-smoke.sh [總畫格] [送鍵腳本]；預設 2600 畫格、'900: ,1800:p'。
# docker run 帶 --init：xvfb-run 當 PID 1 時等不到 Xvfb 的就緒訊號會一直卡住（銀色匕首之謎的做法）。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DG="$ROOT/workplace/dosgolem"
ORIG="$ROOT/workplace/input/dos/pod"
FONT="$ROOT/workplace/font/pod-rounded.golemfnt"
OUT="$ROOT/workplace/play-smoke"
IMAGE=dosgolem-go-ebiten:1.24-r1
CFG_SHA=6161a336828e889c6377bd4fa327376f691a57fc3cd4f2eb6f74db06908d03bd
FRAMES="${1:-2600}"
SCRIPT="${2:-900: ,1800:p}"
for f in "$ORIG/CONTROL.EXE" "$ORIG/ADLIB.COM" "$FONT"; do test -f "$f" || { echo "缺 $f" >&2; exit 1; }; done
mkdir -p "$DG/workplace/bin"
(cd "$DG" && DOSGOLEM_TIMEOUT=10m tools/go.sh build -o /src/workplace/bin/pod-play ./cmd/pod-play)
rm -rf "$OUT"; mkdir -p "$OUT"
timeout 15m docker run --rm --init --network none --memory 2g --cpus 2 --pids-limit 256 \
  --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" -e HOME=/tmp -e POD_NO_DIALOG=1 \
  -v "$DG/workplace/bin":/app:ro -v "$ORIG":/orig:ro -v "$ROOT/text":/text:ro -v "$ROOT/workplace/font":/font:ro \
  -v "$OUT":/data --entrypoint xvfb-run "$IMAGE" -a /app/pod-play -original /orig -save /data/save -text-dir /text \
  -font "/font/$(basename "$FONT")" -audio=false -scale 2 -frames "$FRAMES" -script "$SCRIPT" \
  -shot /data/menu.rgba -shot-original /data/menu-orig.rgba -inventory /data/inventory.tsv > "$OUT/run.log" 2>&1
echo "$CFG_SHA  $OUT/save/pool4.cfg" | sha256sum -c -
cmp "$ORIG/ADLIB.COM" "$OUT/save/snd.com"
# 原版色號與 G2 收據的主選單相同（tools/g2-startup-receipt.sh menu-a.raw；dosgolem 規格 432 §4 第 2 項）
grep -q "indexed_sha256=fefa05142448782c528128c506ffb6ce7fb7c3f39c72dd2e420016e0c1fc11f4" "$OUT/run.log" || { echo "原版畫面與 G2 收據不同" >&2; exit 1; }
grep -q "A map\[覆繪:" "$OUT/run.log" || { echo "主選單沒有覆繪" >&2; exit 1; }
timeout 120 docker run --rm --network none --memory 1g --cpus 1 --pids-limit 64 --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" -v "$OUT":/d sosb-img:r1 python3 -I -c "
from PIL import Image
for n in ('menu', 'menu-orig'):
    Image.frombytes('RGBA', (640, 400), open(f'/d/{n}.rgba', 'rb').read()).save(f'/d/{n}.png')"
if grep -q '缺字' "$OUT/run.log"; then echo "有缺字" >&2; exit 1; fi
tail -1 "$OUT/run.log"
