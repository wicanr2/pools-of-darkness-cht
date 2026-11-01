#!/usr/bin/env bash
# 容器內：與029相同DOS/F1輸入；擷取標號N對應推進前N-1。
set -euo pipefail
test -f /.dockerenv
test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
gui="${POD_OCC_GUI_DIR:-/out/gui}"
test ! -e "$gui"
mkdir "$gui"
export GOMAXPROCS=2 POD_NO_DIALOG=1 LIBGL_ALWAYS_SOFTWARE=1
script='900: ,1800:d,5800:cheat,6900:shot,7100:shot,7600:shot,8100:shot,8500:cheat,8520:shot,8600:cheat,8650:bar,8700:shot,9000:Esc,9600:shot,9800:p,10400:shot,8190:shot,8191:shot,8192:shot,8193:shot,8194:shot,8195:shot,8196:shot,8197:shot,8198:shot,8898:shot,8899:shot,8900:shot,8901:shot'
script+=',6592:shot,6593:shot,6594:shot,7955:shot,7956:shot,7957:shot,8210:shot,8211:shot,8212:shot,8750:shot,8751:shot,8752:shot'
printf '%s\n' "$script" > "$gui/script.txt"
xvfb-run -a /out/pod-play -original /project/workplace/input/dos/pod -save "$gui/game" -text-dir /project/text -font /project/workplace/font/pod-rounded.golemfnt -lang en -audio=false -scale 3 -hd /project/workplace/hd/theme -posters /project/workplace/hd/theme/posters -frames 10500 -script "$script" -flash-audit -shot-dir "$gui/shots" -shot-original "$gui/original.rgba" -shot-source "$gui/source" -shot-source-dir "$gui/sources" -shot-frame-dir "$gui/window" -draw-report "$gui/draw-report.json" > "$gui/run.log" 2>&1
find "$gui" -user root -o -type d -name '*.md'
