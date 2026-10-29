#!/usr/bin/env bash
# 規格019：正常互動模式、固定倍率回歸及完整Draw量測。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG="${1:?指定新收據名稱}"
case "$TAG" in *[!A-Za-z0-9_-]*|'') exit 2;; esac
CPUS="${POD_SPEED_CPUS:-2}"
case "$CPUS" in 2|4) ;; *) exit 2;; esac
SCALE="${POD_SPEED_SCALE:-3}"
case "$SCALE" in 2|3) ;; *) exit 2;; esac
for d in tools text workplace/font workplace/hd/theme workplace/input/dos/pod workplace/dosgolem workplace/probe/g6-saveload-receipt/scratch/SAVE; do test -d "$ROOT/$d"; done
OUT="$ROOT/workplace/probe/$TAG"
test ! -e "$OUT" || { echo '保留既有收據，請換名稱' >&2; exit 2; }
mkdir -p "$OUT"
timeout 1200s docker run --rm --init --network none --label pod.task=interactive-speed \
 --user "$(id -u):$(id -g)" --memory 2g --cpus "$CPUS" --pids-limit 256 \
 --log-opt max-size=10m --log-opt max-file=3 -e HOME=/tmp -e POD_NO_DIALOG=1 -e GOMAXPROCS=2 -e SCALE="$SCALE" \
 -e GOCACHE=/module/workplace/gocache -e GOMODCACHE=/module/workplace/gomodcache \
 -v "$ROOT/workplace/dosgolem:/module:ro" -v "$ROOT/workplace/dosgolem/workplace:/module/workplace:rw" \
 -v "$ROOT/workplace/input/dos/pod:/orig:ro" -v "$ROOT/workplace/font:/font:ro" -v "$ROOT/text:/text:ro" \
 -v "$ROOT/workplace/hd/theme:/theme:ro" -v "$ROOT/workplace/probe/g6-saveload-receipt/scratch/SAVE:/seed:ro" \
 -v "$ROOT/tools:/tools:ro" -v "$OUT:/out:rw" -w /module \
 dosgolem-go-ebiten:1.24-r1 bash -eu -c '
 test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
 python3 -I /tools/speed_interactive_probe.py prepare /module
 go build -o /out/pod-play-candidate ./cmd/pod-play
 go build -o /out/pod-play-observed ./workplace/speed-interactive-cmd
 xvfb-run -a /out/pod-play-candidate -original /orig -save /out/save -text-dir /text \
  -font /font/pod-rounded.golemfnt -audio=false -frames 1 > /out/install.log 2>&1
 rm -rf /out/save/SAVE
 mkdir /out/save/SAVE
 cp -r /seed/. /out/save/SAVE/
 script="900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:A,3400:Down,3420:Down,3440:Down,3460:Down,3480:Down,3500:Return,3900:shot,4000:Y,4010:O,4020:U,4030:Return,4200:shot,4300:map,4320:shot,4350:mapall,4360:shot,4400:guide,4420:shot,4500:map,4520:shot,4600:map,4620:shot,4700:bar,4720:shot,4740:speeds,4750:shot,4760:speeds,4770:barfonts,4780:shot,4781:barfonts,4800:speeds,4800:speed=1,4900:shot,5000:speed=2,5100:shot,5200:speed=4,5300:shot,5400:speed=8,5500:shot,5600:speed=16,5700:shot,5800:speed=32,5900:shot,6000:speed=1,6100:shot"
 printf "%s\n" "$script" > /out/script.txt
 xvfb-run -a python3 -I /tools/speed_interactive_probe.py run /out /out/pod-play-observed \
  -original /orig -save /out/save -text-dir /text -font /font/pod-rounded.golemfnt \
  -hd /theme -posters /theme/posters -lang zh-TW -audio=false -scale "$SCALE" -frames 0 \
  -script "$script" -perf -shot-dir /out/shots -shot-frame-dir /out/window-shots -shot-source-dir /out/sources
 # 與既有021收據相同的固定1x輸入，驗證診斷修正沒有改變遊戲或PNG。
 mkdir /out/fixed
 fixed="900: ,1000:bar,1005:barfonts,1011:shot,1014:shot,1017:shot,1020:scale,1024:shot,1028:shot,1032:shot,1040:scale,1050:lang,1055:shot,1060:lang,1070:barfonts"
 printf "%s\n" "$fixed" > /out/fixed/script.txt
 xvfb-run -a /out/pod-play-candidate -original /orig -save /out/fixed/save -text-dir /text \
  -font /font/pod-rounded.golemfnt -hd /theme -posters /theme/posters -lang zh-TW -scale 3 \
  -audio=false -frames 1100 -script "$fixed" -perf -shot-dir /out/fixed/shots \
  -shot-frame-dir /out/fixed/window-shots -shot-source-dir /out/fixed/sources \
  -shot-source /out/fixed/source -shot-original /out/fixed/original.rgba -shot-frame /out/fixed/frame.png \
  > /out/fixed/run.log 2>&1
 find /out /module/workplace/speed-interactive-cmd -user root -o -type d -name "*.md"
 '
