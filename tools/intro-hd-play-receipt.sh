#!/usr/bin/env bash
# 規格010片頭HD：正常SND／CONTROL，不送DOS鍵，5100回合。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE=${1:-hd}
TAG=${2:-intro-hd-runtime}
case "$MODE" in hd|base|switch) ;; *) exit 2;; esac
case "$TAG" in *[!A-Za-z0-9_-]*|'') exit 2;; esac
THEME=${POD_INTRO_THEME:-$ROOT/workplace/hd/theme}
BIN=${POD_INTRO_BIN:-$ROOT/workplace/dosgolem/workplace/bin/pod-play}
OUT="$ROOT/workplace/probe/$TAG/$MODE"
for d in "$THEME" "$ROOT/workplace/font" "$ROOT/workplace/input/dos/pod" "$ROOT/text" "$ROOT/tools";do test -d "$d";done
test -x "$BIN"
test ! -e "$OUT"
mkdir -p "$OUT"
timeout 900s docker run --rm --init --name "pod-intro-$MODE-$TAG" --label pod.task=intro-hd --network none \
 --memory 2g --cpus 2 --pids-limit 256 --user "$(id -u):$(id -g)" \
 --log-opt max-size=10m --log-opt max-file=3 -e HOME=/tmp -e POD_NO_DIALOG=1 -e GOMAXPROCS=2 \
 --mount "type=bind,source=$BIN,target=/app/pod-play,readonly" \
 --mount "type=bind,source=$ROOT/workplace/input/dos/pod,target=/orig,readonly" \
 --mount "type=bind,source=$ROOT/workplace/font,target=/font,readonly" \
 --mount "type=bind,source=$ROOT/text,target=/text,readonly" \
 --mount "type=bind,source=$ROOT/tools,target=/tools,readonly" \
 --mount "type=bind,source=$THEME,target=/theme,readonly" \
 --mount "type=bind,source=$OUT,target=/out" dosgolem-go-ebiten:1.24-r1 bash -euo pipefail -c '
 test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
 python3 -I /tools/vga_battle_play_check.py --record /out/inputs.json --record-binary /app/pod-play
 script="229:shot,601:shot,602:shot,700:shot,702:shot,704:shot,706:shot,708:shot,710:shot,712:shot,1334:shot,1346:shot,1625:shot,1910:shot,1979:shot,2036:shot,2072:shot,2293:shot,2516:shot,2811:shot,2858:shot,2939:shot,3162:shot,3443:shot,3495:shot,3548:shot,3767:shot,3768:shot,4346:shot,4438:shot,4504:shot,4994:shot,4995:shot"
 view=()
 if [[ $1 == base ]];then view=(-original-view);fi
 if [[ $1 == switch ]];then script="$script,701:lang,703:scale,705:lang,707:toggle,709:scale,711:toggle";fi
 trap "jobs -pr | xargs -r kill" EXIT
 xvfb-run -a /app/pod-play -original /orig -save /out/save -text-dir /text \
  -font /font/pod-rounded.golemfnt -scale 3 -lang zh-TW -audio=false -frames 5100 -script "$script" \
  -hd /theme -posters /theme/posters -flash-audit "${view[@]}" -shot-dir /out/shots \
  -shot-frame-dir /out/window-shots -shot-source-dir /out/sources -shot-source /out/final \
  -shot-frame /out/frame.png > /out/run.log 2>&1
 tail -8 /out/run.log
 find /out -user root -o -type d -name "*.md"
 ' -- "$MODE"
