#!/usr/bin/env bash
# Docker內正常玩家路線；只抽一條讀檔／面板路線，與027既有原版狀態比較。
set -euo pipefail
test -f /.dockerenv
dest=/out/gui
test ! -e "$dest";test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
mkdir "$dest"
binary=/project/workplace/dosgolem/workplace/bin/pod-play
args=(-original /project/workplace/input/dos/pod -save "$dest/game" -text-dir /project/text -font /project/workplace/font/pod-rounded.golemfnt -lang en -audio=false -hd /project/workplace/hd/theme -posters /project/workplace/hd/theme/posters)
export GOMAXPROCS=2 POD_NO_DIALOG=1
xvfb-run -a "$binary" "${args[@]}" -frames 1 > "$dest/install.log" 2>&1
test ! -e "$dest/game/SAVE/savgamA.pty"
mkdir -p "$dest/game/SAVE"
cp -r /project/workplace/probe/g6-saveload-receipt/scratch/SAVE/. "$dest/game/SAVE/"
python3 -I - "$dest" <<'PY'
from pathlib import Path
import os,sys,json,hashlib
d=Path(sys.argv[1]);r=d/'game'
for p in r.rglob('*'):
 if p.is_file():os.utime(p,(946684800,946684800))
(d/'before.json').write_text(json.dumps({str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest() for p in r.rglob('*') if p.is_file()},indent=2))
PY
script='900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:A,3400:Down,3420:Down,3440:Down,3460:Down,3480:Down,3500:Return,3900:shot,4000:Y,4010:O,4020:U,4030:Return,4200:shot,4300:map,4320:shot,4340:guide,4342:shot,4344:next,4346:shot,4348:pgdn,4350:shot,4352:lang,4354:shot,4356:lang,4358:shot,4360:lang,4362:shot,4364:pgup,4366:shot,4370:next,4372:shot,4380:pgdn,4382:shot,4500:map,4520:shot,4600:map,4620:shot,4660:journal,4680:shot,4700:journal,4720:scale,4740:shot,4750:scale,4760:bar,4770:shot'
printf '%s\n' "$script" > "$dest/script.txt"
xvfb-run -a "$binary" "${args[@]}" -scale 3 -frames 4800 -script "$script" -flash-audit \
 -shot-dir "$dest/shots" -shot-original "$dest/original.rgba" -shot-source "$dest/source" -shot-source-dir "$dest/sources" -shot-frame-dir "$dest/window" -draw-report "$dest/draw-report.json" > "$dest/run.log" 2>&1
find "$dest" -user root -o -type d -name '*.md'
