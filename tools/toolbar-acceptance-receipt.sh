#!/usr/bin/env bash
# 規格019／020：正常冷開機的工具列、字模與1x Delay對照；不更動正式程式。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG="${1:?指定新的收據名稱}"
case "$TAG" in *[!A-Za-z0-9_-]*|'') exit 2;; esac
for d in tools text workplace/font workplace/hd/theme workplace/input/dos/pod workplace/dosgolem; do test -d "$ROOT/$d"; done
test -f "$ROOT/workplace/dosgolem/workplace/bin/pod-play"
OUT="$ROOT/workplace/probe/toolbar/$TAG"
PREF_ONLY="${POD_TOOLBAR_PREF_ONLY:-0}"
case "$PREF_ONLY" in
 0) test ! -e "$OUT" || { echo '保留既有收據，請換名稱' >&2; exit 2; }; mkdir -p "$OUT";;
 1) test -f "$OUT/pod-play-observed"; test ! -e "$OUT/restart.jsonl";;
 *) exit 2;;
esac
timeout 1200s docker run --rm --init --network none --label pod.task=toolbar-acceptance \
 --user "$(id -u):$(id -g)" --memory 2g --cpus 2 --pids-limit 256 \
 --log-opt max-size=10m --log-opt max-file=3 \
 -e HOME=/tmp -e POD_NO_DIALOG=1 -e GOMAXPROCS=2 -e PREF_ONLY="$PREF_ONLY" \
 -e GOCACHE=/module/workplace/gocache -e GOMODCACHE=/module/workplace/gomodcache \
 -v "$ROOT/workplace/dosgolem:/module:ro" -v "$ROOT/workplace/dosgolem/workplace:/module/workplace:rw" \
 -v "$ROOT/workplace/input/dos/pod:/orig:ro" -v "$ROOT/workplace/font:/font:ro" \
 -v "$ROOT/text:/text:ro" -v "$ROOT/workplace/hd/theme:/theme:ro" \
 -v "$ROOT/tools:/tools:ro" -v "$OUT:/out:rw" -w /module \
 dosgolem-go-ebiten:1.24-r1 bash -eu -c '
 test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
 python3 -I /tools/toolbar_receipt_probe.py
 if test "$PREF_ONLY" = 0; then
 xvfb-run -a go test -count=1 -v ./workplace/toolbar-receipt-cmd > /out/contracts.log 2>&1
 go build -o /out/pod-play-observed ./workplace/toolbar-receipt-cmd
 for mode in fonts baseline delay-off speeds; do
  mkdir -p "/out/$mode"
  script="900: ,1000:bar,1005:barfonts,1011:shot,1014:shot,1017:shot,1020:scale,1024:shot,1028:shot,1032:shot,1040:scale,1050:lang,1055:shot,1060:lang,1070:barfonts"
  accel=true
  case "$mode" in
   fonts) script="$script,1010:barfont=1,1013:barfont=2,1016:barfont=3,1023:barfont=1,1027:barfont=2,1031:barfont=3";;
   delay-off) accel=false;;
   speeds) script="900: ,1000:bar,1005:speeds,1010:speed=1,1012:shot,1015:speed=2,1017:shot,1020:speed=4,1022:shot,1025:speed=8,1027:shot,1030:speed=16,1032:shot,1035:speed=32,1037:shot,1040:speed=1,1070:speeds";;
  esac
  printf "%s\n" "$script" > "/out/$mode/script.txt"
  POD_TOOLBAR_OBSERVE="/out/$mode/observations.jsonl" xvfb-run -a /out/pod-play-observed \
   -original /orig -save "/out/$mode/save" -text-dir /text -font /font/pod-rounded.golemfnt \
   -hd /theme -posters /theme/posters -lang zh-TW -scale 3 -audio=false \
   -frames 1100 -script "$script" -delay-accel="$accel" -perf \
   -shot-dir "/out/$mode/shots" -shot-frame-dir "/out/$mode/window-shots" \
   -shot-source-dir "/out/$mode/sources" -shot-source "/out/$mode/source" \
   -shot-original "/out/$mode/original.rgba" -shot-frame "/out/$mode/frame.png" \
   > "/out/$mode/run.log" 2>&1
 done
 # 同一輸入再跑正式程式，確認研究用讀值沒有改變執行結果。
 mkdir /out/official
 xvfb-run -a /module/workplace/bin/pod-play \
  -original /orig -save /out/official/save -text-dir /text -font /font/pod-rounded.golemfnt \
  -hd /theme -posters /theme/posters -lang zh-TW -scale 3 -audio=false -frames 1100 \
  -script "$(cat /out/baseline/script.txt)" -perf -shot-dir /out/official/shots \
  -shot-frame-dir /out/official/window-shots -shot-source-dir /out/official/sources \
  -shot-source /out/official/source -shot-original /out/official/original.rgba \
  -shot-frame /out/official/frame.png > /out/official/run.log 2>&1
 fi
 mkdir -p /out/click-preferences
 test ! -f /out/click-preferences/CONTROL.EXE
 cp -r /out/baseline/save/. /out/click-preferences/
 POD_TOOLBAR_PREF_TEST_DIR=/out/click-preferences xvfb-run -a go test -count=1 -v \
  -run TestAcceptanceNativeToolbarAndPreferenceClicks ./workplace/toolbar-receipt-cmd > /out/contracts-preferences.log 2>&1
 # 使用互動模式真正的啟動讀回分支。八秒只驗開機讀值，不當作互動效能測試。
 set +e
 POD_TOOLBAR_OBSERVE=/out/restart.jsonl timeout --kill-after=2s 8s xvfb-run -a /out/pod-play-observed \
  -original /orig -save /out/click-preferences -text-dir /text -font /font/pod-rounded.golemfnt \
  -lang zh-TW -scale 3 -audio=false -frames 0 > /out/restart.log 2>&1
 rc=$?
 set -e
 test "$rc" = 124
 test -s /out/restart.jsonl
 find /out /module/workplace/toolbar-receipt-cmd -user root -o -type d -name "*.md"
 '
