#!/usr/bin/env bash
# 三張開場正常冷啟動與F8/F2/F12切換；沒有快照或遊戲狀態注入。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-hd}"
TAG="${2:-title-play}"
case "$MODE" in hd|base|switch|english-demo) ;; *) echo '用法：title-play-receipt.sh [hd|base|switch|english-demo] [收據名稱]' >&2; exit 2;; esac
case "$TAG" in *[!A-Za-z0-9_-]*|'') echo '收據名稱不合法' >&2; exit 2;; esac
for d in workplace/dosgolem/workplace/bin workplace/input/dos/pod workplace/font workplace/hd/theme-title-stage text tools; do test -d "$ROOT/$d"; done
test -f "$ROOT/workplace/dosgolem/workplace/bin/pod-play-title"
OUT="$ROOT/workplace/probe/g9-hd/$TAG/$MODE"
mkdir -p "$OUT"
test ! -e "$OUT/save" || { echo '既有收據保留，請換收據名稱' >&2; exit 2; }
timeout 400s docker run --rm --init --label pod.task=title-play --network none --memory 2g --cpus 2 --pids-limit 256 \
  --log-opt max-size=10m --log-opt max-file=3 --user "$(id -u):$(id -g)" \
  -e HOME=/tmp -e POD_NO_DIALOG=1 -e GOMAXPROCS=2 -e MODE="$MODE" \
  -v "$ROOT/workplace/dosgolem/workplace/bin:/app:ro" -v "$ROOT/workplace/input/dos/pod:/orig:ro" \
  -v "$ROOT/workplace/font:/font:ro" -v "$ROOT/text:/text:ro" -v "$ROOT/workplace/hd/theme-title-stage:/theme:ro" \
  -v "$ROOT/tools:/tools:ro" -v "$OUT:/out" dosgolem-go-ebiten:1.24-r1 sh -eu -c '
    test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
    python3 -I /tools/vga_battle_play_check.py --record /out/inputs.json --record-binary /app/pod-play-title
    script="900: ,1000:shot,1100:shot,1200:shot"
    theme="-hd /theme";audit=""
    frames=1250;lang=zh-TW
    case "$MODE" in
      hd) audit=-flash-audit ;;
      base) theme="" ;;
      switch) script="$script,1005:lang,1006:shot,1050:lang,1105:lang,1106:shot,1110:lang,1205:lang,1206:shot,1210:lang,1211:shot,1215:scale,1216:shot,1220:scale,1221:shot,1225:toggle,1226:shot,1230:toggle,1231:shot" ;;
      english-demo) script="900: ,1800: ,2050:shot,2200:shot";frames=2210;lang=en;audit=-flash-audit ;;
    esac
    xvfb-run -a /app/pod-play-title -original /orig -save /out/save -text-dir /text -font /font/pod-rounded.golemfnt \
      -scale 3 -lang "$lang" -audio=false -frames "$frames" -script "$script" $theme $audit -posters /theme/posters \
      -shot-dir /out/shots -shot-frame-dir /out/window-shots -shot-source-dir /out/sources \
      -shot-source /out/final -shot-frame /out/frame.png > /out/run.log 2>&1
    find /out -user root -o -type d -name "*.md"
  '
