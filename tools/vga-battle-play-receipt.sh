#!/usr/bin/env bash
# 已有的7400格冷啟動DEMO路徑，驗首批VGA戰鬥HD與前端切換。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-hd}"
TAG="${2:-vga-play}"
case "$TAG" in *[!A-Za-z0-9_-]*|'') echo '收據名稱只接受英數、底線與連字號' >&2; exit 2;; esac
case "$MODE" in hd|base|english|twox) ;; *) echo '用法：vga-battle-play-receipt.sh [hd|base|english|twox]' >&2; exit 2;; esac
for d in workplace/dosgolem/workplace/bin workplace/input/dos/pod workplace/font workplace/hd/theme-vga-stage text tools; do test -d "$ROOT/$d"; done
test -f "$ROOT/workplace/dosgolem/workplace/bin/pod-play-vga"
OUT="$ROOT/workplace/probe/g9-hd/$TAG/$MODE"
mkdir -p "$OUT"
test ! -e "$OUT/save" || { echo '保留既有收據，請另定輸出位置' >&2; exit 2; }
timeout 900s docker run --rm --init --label pod.task=vga-battle-play --network none \
  --memory 2g --cpus 2 --pids-limit 256 --log-opt max-size=10m --log-opt max-file=3 \
  --user "$(id -u):$(id -g)" -e HOME=/tmp -e POD_NO_DIALOG=1 -e GOMAXPROCS=2 -e MODE="$MODE" \
  -v "$ROOT/workplace/dosgolem/workplace/bin:/app:ro" -v "$ROOT/workplace/input/dos/pod:/orig:ro" \
  -v "$ROOT/workplace/font:/font:ro" -v "$ROOT/text:/text:ro" -v "$ROOT/workplace/hd/theme-vga-stage:/theme:ro" \
  -v "$ROOT/tools:/tools:ro" -v "$OUT:/out" dosgolem-go-ebiten:1.24-r1 sh -eu -c '
    test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
    python3 -I /tools/vga_battle_play_check.py --record /out/inputs.json
    script="900: ,1800: "
    theme="-hd /theme"
    audit=""
    case "$MODE" in
      hd) audit=-flash-audit ;;
      base) theme="" ;;
      english) script="$script,7390:toggle" ;;
      twox) script="$script,7390:scale" ;;
    esac
    xvfb-run -a /app/pod-play-vga -original /orig -save /out/save -text-dir /text \
      -font /font/pod-rounded.golemfnt -scale 3 -audio=false -frames 7400 -script "$script" \
      $theme $audit -posters /theme/posters -shot-frame /out/frame.png -shot-original /out/original.rgba \
      -shot-source /out/source > /out/run.log 2>&1
    find /out -user root -o -type d -name "*.md"
  '
