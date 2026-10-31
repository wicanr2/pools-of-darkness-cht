#!/bin/sh
# 容器內執行實際AppImage；/out獨立可寫，/packages及/seed唯讀。
set -eu
version=${1:?version};mode=${2:?base|hd|save|flash|missing}
case "$mode" in base|hd|save|flash|missing) ;; *) exit 2;; esac
export HOME=/out/home XDG_DATA_HOME=/out/data XDG_CONFIG_HOME=/out/config POD_NO_DIALOG=1 APPIMAGE_EXTRACT_AND_RUN=1 GOMAXPROCS=2
mkdir -p "$HOME" "$XDG_DATA_HOME" "$XDG_CONFIG_HOME"
edition=$mode;test "$mode" != save || edition=hd
case "$mode" in flash|missing) edition=base;; esac
app="/packages/pod-$version-linux-$edition-full-local.AppImage"
xvfb-run -a "$app" -version > /out/version.txt 2>&1
if test "$mode" = missing; then
 if xvfb-run -a "/packages/pod-$version-linux-base-patch.AppImage" -audio=false -frames 1 > /out/missing.log 2>&1; then
  echo '缺原版卻成功啟動' >&2;exit 1
 fi
 python3 -I -c 'from pathlib import Path;import json;r=Path("/out");s=(r/"missing.log").read_text();assert "1.10" in s and not list((r/"data").rglob("GAME.EXE"));(r/"missing-pass.json").write_text(json.dumps({"missing_original_rejected":True,"version_hint":"1.10"},indent=2))'
elif test "$mode" = flash; then
 xvfb-run -a "$app" -audio=false -scale 3 -lang zh-TW -frames 2820 -script '900: ,1800:d,2300:x,2803:shot,2804:shot,2805:shot,2806:shot,2810:shot' -flash-audit -shot-dir /out/shots -shot-source-dir /out/sources > /out/run.log 2>&1
elif test "$mode" = save; then
 xvfb-run -a "$app" -audio=false -frames 1 > /out/install.log 2>&1
 test -f /out/data/pod-cht/game/GAME.EXE
 test -f /seed/savgamA.pty
 test ! -f /seed/savgamB.pty
 mkdir -p /out/data/pod-cht/game/SAVE
 cp -r /seed/. /out/data/pod-cht/game/SAVE/
 python3 -I -c 'from pathlib import Path;import os,json,hashlib;r=Path("/out/data/pod-cht/game");[os.utime(p,(946684800,946684800)) for p in r.rglob("*") if p.is_file()];Path("/out/before-save.json").write_text(json.dumps({str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest() for p in r.rglob("*") if p.is_file()},indent=2))'
 xvfb-run -a "$app" -audio=false -scale 2 -lang zh-TW -frames 4700 -posters - -script '900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:A,3400:Down,3420:Down,3440:Down,3460:Down,3480:Down,3500:Return,3900:shot,4000:Y,4010:O,4020:U,4030:Return,4200:shot,4300:E,4400:S,4450:shot,4500:B,4550:shot,4650:shot' -shot-dir /out/save-shots -shot-source /out/save-source > /out/save.log 2>&1
 test -s /out/data/pod-cht/game/SAVE/savgamB.pty
 python3 -I -c 'from pathlib import Path;import json,hashlib;r=Path("/out/data/pod-cht/game/SAVE");Path("/out/after-save.json").write_text(json.dumps({p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in r.iterdir() if p.is_file()},indent=2))'
 xvfb-run -a "$app" -audio=false -scale 2 -lang zh-TW -frames 3250 -posters - -script '900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:B,3200:shot' -shot-dir /out/load-shots -shot-source /out/load-source > /out/load.log 2>&1
 python3 -I -c 'from pathlib import Path;import hashlib,json;r=Path("/out/data/pod-cht/game/SAVE");a=json.loads(Path("/out/after-save.json").read_text());assert all(hashlib.sha256((r/n).read_bytes()).hexdigest()==h for n,h in a.items());Path("/out/save-load-pass.json").write_text(json.dumps({"saved_slot":"B","cold_loaded_slot":"B","save_hashes_preserved":True,"files":a},indent=2))'
else
 xvfb-run -a "$app" -audio=false -scale 3 -lang zh-TW -frames 3800 -script '900: ,1800:d,2050:shot,2300:x,2500:shot,3000:shot,3200:p,3700:shot' -flash-audit -shot-dir /out/shots -shot-frame /out/frame.png -shot-original /out/original.rgba -shot-source /out/source -draw-report /out/draw-report.json > /out/run.log 2>&1
 test -f /out/data/pod-cht/game/GAME.EXE
 test -f /out/data/pod-cht/game/pool4.cfg
 python3 -I -c 'from pathlib import Path;import json,hashlib;r=Path("/out/data/pod-cht/game");Path("/out/first-import.json").write_text(json.dumps({str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest() for p in r.rglob("*") if p.is_file()},indent=2))'
 # 不含原版的patch沿用已匯入資料，驗證第二次啟動不依賴完整版內的original。
 xvfb-run -a "/packages/pod-$version-linux-base-patch.AppImage" -audio=false -frames 1 > /out/restart.log 2>&1
 python3 -I -c 'from pathlib import Path;import json,hashlib;r=Path("/out/data/pod-cht/game");a=json.loads(Path("/out/first-import.json").read_text());assert a=={str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest() for p in r.rglob("*") if p.is_file()};Path("/out/restart-pass.json").write_text(json.dumps({"existing_import_used":True,"game_files_unchanged":True}))'
 tail -n 7 /out/run.log
fi
find /out -user root -o -type d -name '*.md'
