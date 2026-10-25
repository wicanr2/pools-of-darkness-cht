#!/usr/bin/env bash
# 已有的7400格冷啟動DEMO路徑，驗VGA戰鬥HD與前端切換。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-hd}"
TAG="${2:-vga-play}"
THEME="${3:-workplace/hd/theme-vga-stage}"
BINARY="${4:-pod-play-vga}"
RECEIPT_TIMEOUT="${POD_RECEIPT_TIMEOUT:-900}"
BIGPIC_SAMPLE="${POD_BIGPIC_SAMPLE:-0}"
DIRECT_REPORT="${POD_DIRECT_REPORT:-0}"
case "$DIRECT_REPORT" in 0|1) ;; *) echo '繪圖命令收據只接受0或1' >&2; exit 2;; esac
case "$BIGPIC_SAMPLE" in 0|1) ;; *) echo 'BIGPIC抽樣只接受0或1' >&2; exit 2;; esac
case "$RECEIPT_TIMEOUT" in *[!0-9]*|'') echo '收據逾時須為60至3600秒' >&2; exit 2;; esac
test "$RECEIPT_TIMEOUT" -ge 60 && test "$RECEIPT_TIMEOUT" -le 3600 || { echo '收據逾時須為60至3600秒' >&2; exit 2; }
case "$THEME" in workplace/hd/*) ;; *) echo '主題須在 workplace/hd 內' >&2; exit 2;; esac
case "$THEME" in *..*|*[!A-Za-z0-9_/-]*) echo '主題路徑不合法' >&2; exit 2;; esac
case "$BINARY" in *[!A-Za-z0-9_-]*|'') echo '執行檔名稱不合法' >&2; exit 2;; esac
case "$TAG" in *[!A-Za-z0-9_-]*|'') echo '收據名稱只接受英數、底線與連字號' >&2; exit 2;; esac
case "$MODE" in hd|base|english|twox) ;; *) echo '用法：vga-battle-play-receipt.sh [hd|base|english|twox]' >&2; exit 2;; esac
for d in workplace/dosgolem/workplace/bin workplace/input/dos/pod workplace/font "$THEME" text tools; do test -d "$ROOT/$d"; done
test -f "$ROOT/workplace/dosgolem/workplace/bin/$BINARY"
OUT="$ROOT/workplace/probe/g9-hd/$TAG/$MODE"
mkdir -p "$OUT"
test ! -e "$OUT/save" || { echo '保留既有收據，請另定輸出位置' >&2; exit 2; }
timeout "${RECEIPT_TIMEOUT}s" docker run --rm --init --label pod.task=vga-battle-play --network none \
  --memory 2g --cpus 2 --pids-limit 256 --log-opt max-size=10m --log-opt max-file=3 \
  --user "$(id -u):$(id -g)" -e HOME=/tmp -e POD_NO_DIALOG=1 -e GOMAXPROCS=2 -e MODE="$MODE" -e BINARY="$BINARY" -e BIGPIC_SAMPLE="$BIGPIC_SAMPLE" -e DIRECT_REPORT="$DIRECT_REPORT" \
  -v "$ROOT/workplace/dosgolem/workplace/bin:/app:ro" -v "$ROOT/workplace/input/dos/pod:/orig:ro" \
  -v "$ROOT/workplace/font:/font:ro" -v "$ROOT/text:/text:ro" -v "$ROOT/$THEME:/theme:ro" \
  -v "$ROOT/tools:/tools:ro" -v "$OUT:/out" dosgolem-go-ebiten:1.24-r1 sh -eu -c '
    test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
    python3 -I /tools/vga_battle_play_check.py --record /out/inputs.json --record-binary "/app/$BINARY"
    script="900: ,1800: "
    theme="-hd /theme"
    audit=""
    capture=""
    draw_report=""
    if test "$DIRECT_REPORT" = 1; then draw_report="-draw-report /out/draw-report.json"; fi
    if test "$BIGPIC_SAMPLE" = 1; then
      capture="-shot-dir /out/shots -shot-frame-dir /out/window-shots -shot-source-dir /out/sources"
      case "$MODE" in
        english) script="$script,3640:toggle,3650:shot,3660:toggle" ;;
        twox) script="$script,3640:scale,3650:shot,3660:scale" ;;
        *) script="$script,3650:shot" ;;
      esac
    fi
    case "$MODE" in
      hd) audit=-flash-audit ;;
      base) theme="" ;;
      english) script="$script,7390:toggle" ;;
      twox) script="$script,7390:scale" ;;
    esac
    xvfb-run -a "/app/$BINARY" -original /orig -save /out/save -text-dir /text \
      -font /font/pod-rounded.golemfnt -scale 3 -audio=false -frames 7400 -script "$script" \
      $draw_report \
      $theme $audit $capture -posters /theme/posters -shot-frame /out/frame.png -shot-original /out/original.rgba \
      -shot-source /out/source > /out/run.log 2>&1
    find /out -user root -o -type d -name "*.md"
  '
