#!/usr/bin/env bash
# Docker內完整候選的正常讀檔／面板路線，原版與正式安裝唯讀。
set -euo pipefail
test -f /.dockerenv
code=${1:?zh-TW|ja|ko}
case "$code" in zh-TW|ja|ko) ;; *) exit 2;; esac
dest="/out/gui-$code"
test ! -e "$dest"
test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
mkdir "$dest"
binary=/project/workplace/dosgolem/workplace/bin/pod-play
args=(-original /project/workplace/input/dos/pod -save "$dest/game" -text-dir /out/candidate/text
 -font /out/candidate/font/pod-rounded.golemfnt -lang "$code" -audio=false
 -hd /out/candidate/theme -posters /out/candidate/theme/posters)
export GOMAXPROCS=2 POD_NO_DIALOG=1 HOME=/out/home
mkdir -p "$HOME"
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
(d/'before.json').write_text(json.dumps({str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest() for p in r.rglob('*') if p.is_file()},indent=2)+'\n')
PY
# DOS按鍵與028相同；前端增加五語循環、兩倍率各三級字型與F12。
script='900: ,1000:shot,1002:bar,1004:barfont=1,1006:shot,1008:barfont=2,1010:shot,1012:barfont=3,1014:shot,1016:bar,1100:shot,1200:shot,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:A,3400:Down,3420:Down,3440:Down,3460:Down,3480:Down,3500:Return,3900:shot,4000:Y,4010:O,4020:U,4030:Return,4200:shot,4300:map,4320:shot,4340:guide,4342:shot,4344:next,4346:shot,4348:pgdn,4350:shot,4352:lang,4354:shot,4356:lang,4358:shot,4360:lang,4362:shot,4364:lang,4366:shot,4368:lang,4370:shot,4372:pgup,4374:shot,4376:next,4378:shot,4380:pgdn,4382:shot,4500:map,4520:shot,4600:map,4620:shot,4630:toggle,4632:shot,4640:toggle,4642:shot,4660:journal,4680:shot,4700:journal,4720:scale,4722:bar,4724:barfont=1,4726:shot,4728:barfont=2,4730:shot,4732:barfont=3,4734:shot,4736:bar,4740:shot,4750:scale,4760:bar,4770:shot'
printf '%s\n' "$script" > "$dest/script.txt"
xvfb-run -a "$binary" "${args[@]}" -scale 3 -frames 4800 -script "$script" -flash-audit \
 -shot-dir "$dest/shots" -shot-frame "$dest/final.png" -shot-original "$dest/original.rgba" \
 -shot-source "$dest/source" -shot-source-dir "$dest/sources" -shot-frame-dir "$dest/window" \
 -draw-report "$dest/draw-report.json" > "$dest/run.log" 2>&1
find "$dest" -user root -o -type d -name '*.md'
