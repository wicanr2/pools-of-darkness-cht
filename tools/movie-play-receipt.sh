#!/usr/bin/env bash
# 正常片頭、不送DOS鍵；F8/F2/F12只改前端，5100回合與時序探針相同。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-subtitles}"
TAG="${2:-intro-play}"
case "$MODE" in subtitles|base) ;; *) echo '用法：movie-play-receipt.sh [subtitles|base] [收據名稱]' >&2; exit 2;; esac
case "$TAG" in *[!A-Za-z0-9_-]*|'') echo '收據名稱不合法' >&2; exit 2;; esac
for d in workplace/dosgolem/workplace/bin workplace/input/dos/pod workplace/font workplace/hd/theme text tools; do test -d "$ROOT/$d"; done
test -x "$ROOT/workplace/dosgolem/workplace/bin/pod-play-movie"
OUT="$ROOT/workplace/probe/g9-hd/$TAG/$MODE"
test ! -e "$OUT" || { echo '保留舊收據，請另定名稱' >&2; exit 2; }
mkdir -p "$OUT"
timeout 900s docker run --rm --init --label pod.task=movie-play --network none --memory 2g --cpus 2 --pids-limit 256 \
 --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 \
 -e HOME=/tmp -e POD_NO_DIALOG=1 -e GOMAXPROCS=2 -e MODE="$MODE" \
 -v "$ROOT/workplace/dosgolem/workplace/bin:/app:ro" -v "$ROOT/workplace/input/dos/pod:/orig:ro" \
 -v "$ROOT/workplace/font:/font:ro" -v "$ROOT/text:/text:ro" -v "$ROOT/workplace/hd/theme:/theme:ro" \
 -v "$ROOT/tools:/tools:ro" -v "$OUT:/out" dosgolem-go-ebiten:1.24-r1 sh -eu -c '
 test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
 python3 -I /tools/vga_battle_play_check.py --record /out/inputs.json --record-binary /app/pod-play-movie
 script="601:shot,602:shot,700:shot,702:shot,704:shot,706:shot,708:shot,710:shot,712:shot,947:shot,1625:shot,2293:shot,2516:shot,3162:shot,3770:shot,3985:shot,4994:shot,4995:shot"
 view=""
 if test "$MODE" = base; then view=-original-view; else
   script="$script,701:lang,703:scale,705:lang,707:toggle,709:scale,711:toggle"
 fi
 xvfb-run -a /app/pod-play-movie -original /orig -save /out/save -text-dir /text \
  -font /font/pod-rounded.golemfnt -scale 3 -lang zh-TW -audio=false -frames 5100 -script "$script" \
  -hd /theme -posters /theme/posters -flash-audit $view -shot-dir /out/shots \
  -shot-frame-dir /out/window-shots -shot-source-dir /out/sources -shot-source /out/final \
  -shot-frame /out/frame.png > /out/run.log 2>&1
 find /out -user root -o -type d -name "*.md"
 '
