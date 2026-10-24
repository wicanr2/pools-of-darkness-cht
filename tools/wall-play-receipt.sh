#!/usr/bin/env bash
# 同一G8讀槽A路線，對比既有HD與新增VGA街景；不增加探索。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-hd}"
TAG="${2:-wall-play}"
case "$MODE" in hd|base) ;; *) echo '用法：wall-play-receipt.sh [hd|base] [收據名稱]' >&2; exit 2;; esac
case "$TAG" in *[!A-Za-z0-9_-]*|'') echo '收據名稱不合法' >&2; exit 2;; esac
SEED="$ROOT/workplace/probe/g6-saveload-receipt/scratch/SAVE"
for d in workplace/dosgolem/workplace/bin workplace/input/dos/pod workplace/font text tools workplace/hd/theme-wall-stage workplace/hd/theme-b09-stage; do test -d "$ROOT/$d"; done
test -d "$SEED"
test -f "$SEED/savgamA.pty"
test -f "$ROOT/workplace/dosgolem/workplace/bin/pod-play-wall"
OUT="$ROOT/workplace/probe/g9-hd/$TAG/$MODE"
mkdir -p "$OUT"
test ! -e "$OUT/save" || { echo '保留舊收據，請另定名稱' >&2; exit 2; }
timeout 900s docker run --rm --init --label pod.task=wall-play --network none --memory 2g --cpus 2 --pids-limit 256 \
  --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 \
  -v "$ROOT/workplace/dosgolem/workplace/bin:/app:ro" -v "$ROOT/workplace/input/dos/pod:/orig:ro" \
  -v "$ROOT/workplace/font:/font:ro" -v "$ROOT/text:/text:ro" -v "$SEED:/seed:ro" \
  -v "$ROOT/workplace/hd/theme-b09-stage:/base:ro" -v "$ROOT/workplace/hd/theme-wall-stage:/walls:ro" \
  -v "$ROOT/tools:/tools:ro" -v "$OUT:/out" -e MODE="$MODE" -e HOME=/tmp -e POD_NO_DIALOG=1 -e GOMAXPROCS=2 \
  dosgolem-go-ebiten:1.24-r1 sh -eu -c '
    test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
    python3 -I /tools/vga_wall_play_check.py --record /out/inputs.json
    xvfb-run -a /app/pod-play-wall -original /orig -save /out/save -text-dir /text -font /font/pod-rounded.golemfnt -audio=false -frames 1 > /out/install.log 2>&1
    rm -rf /out/save/SAVE
    mkdir -p /out/save/SAVE
    cp -r /seed/. /out/save/SAVE/
    theme=/base; audit=""
    if test "$MODE" = hd; then theme=/walls; audit=-flash-audit; fi
    xvfb-run -a /app/pod-play-wall -original /orig -save /out/save -text-dir /text \
      -font /font/pod-rounded.golemfnt -hd "$theme" -posters "$theme/posters" -lang zh-TW -audio=false -scale 3 -frames 4800 $audit \
      -script "900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:A,3400:Down,3420:Down,3440:Down,3460:Down,3480:Down,3500:Return,3900:shot,4000:Y,4010:O,4020:U,4030:Return,4200:shot,4300:map,4320:shot,4350:mapall,4360:shot,4400:guide,4420:shot,4500:map,4520:shot,4600:map,4620:shot,4700:bar,4720:shot,4740:speeds,4750:shot,4760:speeds,4770:barfonts,4780:shot" \
      -shot-dir /out/shots -shot-frame-dir /out/window-shots -shot-source-dir /out/sources \
      -shot-source /out/source -shot-original /out/original.rgba -shot-frame /out/final.png > /out/run.log 2>&1
    find /out -user root -o -type d -name "*.md"
  '
