#!/usr/bin/env bash
# Docker內用正式版程式重錄正常讀檔、查驗題、攻略與地圖。
set -euo pipefail
test -f /.dockerenv
test ! -e /out/player
mkdir /out/player /out/player/home
export HOME=/out/player/home GOMAXPROCS=2 POD_NO_DIALOG=1
binary=/bin-input/pod-play
args=(-original /original -save /out/player/game -text-dir /main/text
 -font /assets/font/pod-rounded.golemfnt -hd /assets/theme -posters /assets/theme/posters
 -audio=false -scale 3 -lang zh-TW)
xvfb-run -a "$binary" "${args[@]}" -frames 1 > /out/player/install.log 2>&1
mkdir -p /out/player/game/SAVE
cp -r /seed/. /out/player/game/SAVE/
python3 -I - <<'PY'
from pathlib import Path
import os
for p in Path('/out/player/game').rglob('*'):
    if p.is_file():os.utime(p,(946684800,946684800))
PY
script='900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:A,3400:Down,3420:Down,3440:Down,3460:Down,3480:Down,3500:Return,3800:shot,3900:shot,4030:Return,4200:shot,4300:map,4320:shot,4340:guide,4342:shot,4344:next,4346:shot,4348:pgdn,4350:shot,4352:lang,4354:shot,4356:lang,4358:shot,4360:lang,4362:shot,4364:lang,4366:shot,4368:lang,4370:shot,4500:map,4520:shot,4600:map,4620:shot'
printf '%s\n' "$script" > /out/player/script.txt
xvfb-run -a "$binary" "${args[@]}" -frames 4800 -script "$script" \
 -shot-dir /out/player/shots -shot-frame-dir /out/player/window -shot-source-dir /out/player/sources \
 -shot-source /out/player/final > /out/player/run.log 2>&1
grep -q 'quiz autofill frame=3505 answer=YOU enter=false' /out/player/run.log
tail -n 7 /out/player/run.log
find /out/player -user root -o -type d -name '*.md'
