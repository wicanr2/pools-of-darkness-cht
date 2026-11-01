#!/usr/bin/env bash
# Docker內原版FINAL直接播放器及正式CONTROL回歸；測試佇列不冒稱正常通關。
set -euo pipefail
test -f /.dockerenv
mode=${1:?base|hd|switch|demo}
case "$mode" in base|hd|switch|demo) ;; *) exit 2;; esac
tag=${2:-runtime-v1}
case "$tag" in *[!A-Za-z0-9_-]*|'') exit 2;; esac
dest="/out/$tag/$mode"
test ! -e "$dest"
test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
mkdir -p "$dest"
export GOMAXPROCS=2 POD_NO_DIALOG=1 HOME="$dest/home"
mkdir "$HOME"
binary=/out/bin/pod-final-player
frames=2001
script=$(python3 -I -c 'import json;from pathlib import Path;r=json.loads(Path("/out/source-proof.json").read_text());print(",".join(str(n)+":shot" for n in sorted({e["update"]+1 for e in r["events"] if e["update"]<2001}|{10,200,300,400,411,413,415,417,419,431,441,451,461,700,1000,1300,1500,1800,1990})))')
view=()
if [[ "$mode" = base ]]; then view=(-original-view); fi
if [[ "$mode" = switch ]]; then
  script="$script,410:lang,412:lang,414:lang,416:lang,418:lang,430:toggle,440:toggle,450:scale,460:scale"
fi
if [[ "$mode" = demo ]]; then
  binary=/out/bin/pod-play
  frames=7400
  script='900: ,1800:d'
fi
printf '%s\n' "$script" > "$dest/script.txt"
trap 'jobs -pr | xargs -r kill' EXIT
xvfb-run -a "$binary" -original /project/workplace/input/dos/pod -save "$dest/save" \
 -text-dir /project/text -font /project/workplace/font/pod-rounded.golemfnt -lang zh-TW -audio=false \
 -scale 3 -frames "$frames" -script "$script" -hd /out/candidate/theme -posters /out/candidate/theme/posters \
 -flash-audit "${view[@]}" -shot-dir "$dest/shots" -shot-frame-dir "$dest/window" \
 -shot-source-dir "$dest/sources" -shot-source "$dest/final" -shot-frame "$dest/frame.png" \
 -draw-report "$dest/draw-report.json" > "$dest/run.log" 2>&1
tail -n 8 "$dest/run.log"
find "$dest" -user root -o -type d -name '*.md'
