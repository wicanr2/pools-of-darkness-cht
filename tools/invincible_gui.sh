#!/usr/bin/env bash
# 容器內執行，一條正常讀檔關閉對照及一條正常DEMO無敵開關。
set -euo pipefail
test -f /.dockerenv
test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
export GOMAXPROCS=2 POD_NO_DIALOG=1 LIBGL_ALWAYS_SOFTWARE=1
mode=${1:?off、city-on、demo、audit 或 baseline}
dest=/out/gui-$mode
test ! -e "$dest"
mkdir "$dest"
args=(-original /project/workplace/input/dos/pod -save "$dest/game" -text-dir /project/text -font /project/workplace/font/pod-rounded.golemfnt -lang en -audio=false -scale 3 -hd /project/workplace/hd/theme -posters /project/workplace/hd/theme/posters)
binary=/out/pod-play
if [[ $mode == baseline ]];then binary=/project/workplace/dosgolem/workplace/bin/pod-play-before-invincible;fi
if [[ $mode == city-on ]];then args+=(-cheat);fi
case "$mode" in
 off|city-on)
  xvfb-run -a "$binary" "${args[@]}" -frames 1 > "$dest/install.log" 2>&1
  cp -r /project/workplace/probe/g6-saveload-receipt/scratch/SAVE/. "$dest/game/SAVE/"
  python3 -I - "$dest/game" <<'PY'
from pathlib import Path
import os,sys
for p in Path(sys.argv[1]).rglob('*'):
 if p.is_file():os.utime(p,(946684800,946684800))
PY
  script=$(python3 -I -c 'import re;from pathlib import Path;print(re.search("script=\x27([^\x27]+)\x27",Path("/project/tools/english_panels_gui.sh").read_text())[1])')
  frames=4800;;
 demo|audit|baseline)
  script='900: ,1800:d,5800:cheat,6900:shot,7100:shot,7600:shot,8100:shot,8500:cheat,8520:shot,8600:cheat,8650:bar,8700:shot,9000:Esc,9600:shot,9800:p,10400:shot'
  frames=10500
  if [[ $mode == audit || $mode == baseline ]];then script+=',8190:shot,8191:shot,8192:shot,8193:shot,8195:shot,8196:shot,8898:shot,8899:shot';fi;;
 *) exit 2;;
esac
printf '%s\n' "$script" > "$dest/script.txt"
xvfb-run -a "$binary" "${args[@]}" -frames "$frames" -script "$script" -flash-audit -shot-dir "$dest/shots" -shot-original "$dest/original.rgba" -shot-source "$dest/source" -shot-source-dir "$dest/sources" -shot-frame-dir "$dest/window" -draw-report "$dest/draw-report.json" > "$dest/run.log" 2>&1
find "$dest" -user root -o -type d -name '*.md'
