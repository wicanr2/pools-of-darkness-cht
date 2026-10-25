#!/usr/bin/env bash
# 沿用既有讀槽A／片頭／標題路線，成對驗證直接與舊合成；不增加探索。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROUTE="${1:?wall|movie|title}"
MODE="${2:?direct|legacy}"
TAG="${3:-direct-ui}"
BINARY="${4:-pod-play-direct}"
case "$BINARY" in *[!A-Za-z0-9_-]*|'') exit 2;; esac
case "$ROUTE" in wall|movie|title) ;; *) exit 2;; esac
case "$MODE" in direct) DIRECT=true;; legacy) DIRECT=false;; *) exit 2;; esac
case "$TAG" in *[!A-Za-z0-9_-]*|'') exit 2;; esac
for d in workplace/dosgolem/workplace/bin workplace/input/dos/pod workplace/font workplace/hd/theme text tools; do test -d "$ROOT/$d"; done
SEED="$ROOT/workplace/probe/g6-saveload-receipt/scratch/SAVE"
test -d "$SEED"
test -f "$ROOT/workplace/dosgolem/workplace/bin/$BINARY"
OUT="$ROOT/workplace/probe/g9-hd/$TAG/$ROUTE-$MODE"
test ! -e "$OUT" || { echo '保留既有收據，請換名稱' >&2; exit 2; }
mkdir -p "$OUT"
timeout 1800s docker run --rm --init --label pod.task=direct-ui --network none --memory 2g --cpus 2 --pids-limit 256 \
 --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 \
 -e HOME=/tmp -e POD_NO_DIALOG=1 -e GOMAXPROCS=2 -e ROUTE="$ROUTE" -e DIRECT="$DIRECT" -e BINARY="$BINARY" \
 -v "$ROOT/workplace/dosgolem/workplace/bin:/app:ro" -v "$ROOT/workplace/input/dos/pod:/orig:ro" \
 -v "$ROOT/workplace/font:/font:ro" -v "$ROOT/text:/text:ro" -v "$ROOT/workplace/hd/theme:/theme:ro" \
 -v "$ROOT/tools:/tools:ro" -v "$SEED:/seed:ro" -v "$OUT:/out" \
 dosgolem-go-ebiten:1.24-r1 sh -eu -c '
 test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
 python3 -I /tools/vga_battle_play_check.py --record /out/inputs.json --record-binary "/app/$BINARY"
 case "$ROUTE" in
 wall)
  xvfb-run -a "/app/$BINARY" -original /orig -save /out/save -text-dir /text -font /font/pod-rounded.golemfnt -audio=false -frames 1 -direct-render="$DIRECT" > /out/install.log 2>&1
  rm -rf /out/save/SAVE
  mkdir -p /out/save/SAVE
  cp -r /seed/. /out/save/SAVE/
  frames=4800
  script="900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:A,3400:Down,3420:Down,3440:Down,3460:Down,3480:Down,3500:Return,3900:shot,4000:Y,4010:O,4020:U,4030:Return,4200:shot,4300:map,4320:shot,4350:mapall,4360:shot,4400:guide,4420:shot,4500:map,4520:shot,4600:map,4620:shot,4700:bar,4720:shot,4740:speeds,4750:shot,4760:speeds,4770:barfonts,4780:shot"
  ;;
 movie)
  frames=5100
  script="601:shot,602:shot,700:shot,702:shot,704:shot,706:shot,708:shot,710:shot,712:shot,947:shot,1625:shot,2293:shot,2516:shot,3162:shot,3770:shot,3985:shot,4994:shot,4995:shot,701:lang,703:scale,705:lang,707:toggle,709:scale,711:toggle"
  ;;
 title)
  frames=1250
  script="900: ,1000:shot,1100:shot,1200:shot,1005:lang,1006:shot,1050:lang,1105:lang,1106:shot,1110:lang,1205:lang,1206:shot,1210:lang,1211:shot,1215:scale,1216:shot,1220:scale,1221:shot,1225:toggle,1226:shot,1230:toggle,1231:shot"
  ;;
 esac
 xvfb-run -a "/app/$BINARY" -original /orig -save /out/save -text-dir /text \
  -font /font/pod-rounded.golemfnt -hd /theme -posters /theme/posters -lang zh-TW -audio=false -scale 3 \
  -frames "$frames" -script "$script" -direct-render="$DIRECT" -draw-report /out/draw-report.json -perf -flash-audit \
  -shot-dir /out/shots -shot-frame-dir /out/window-shots -shot-source-dir /out/sources \
  -shot-source /out/source -shot-original /out/original.rgba -shot-frame /out/frame.png > /out/run.log 2>&1
 find /out -user root -o -type d -name "*.md"
 '
