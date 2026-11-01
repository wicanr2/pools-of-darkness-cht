#!/usr/bin/env bash
# 容器內，實際HD patch AppImage；/input/original是另外唯讀掛入的合法原版。
set -euo pipefail
MODE=${1:?intro|demo|save|missing|corrupt}
VERSION=v.0.1.1-20261101
APP="/input/pod-$VERSION-linux-hd-patch.AppImage"
export HOME=/out/home XDG_DATA_HOME=/out/data XDG_CONFIG_HOME=/out/config POD_NO_DIALOG=1 APPIMAGE_EXTRACT_AND_RUN=1 GOMAXPROCS=2
mkdir -p "$HOME" "$XDG_DATA_HOME" "$XDG_CONFIG_HOME"
run(){ xvfb-run -a "$APP" "$@"; }
snapshot(){
 python3 -I - "$1" <<'PY'
from pathlib import Path
import hashlib,json,sys
r=Path('/out/data/pod-cht');cache=list((r/'hd-cache').iterdir());assert len(cache)==1
tree=lambda p:{str(q.relative_to(p)):hashlib.sha256(q.read_bytes()).hexdigest() for q in p.rglob('*') if q.is_file()}
Path('/out/'+sys.argv[1]).write_text(json.dumps(dict(game=tree(r/'game'),cache=tree(cache[0]),mtime={str(q.relative_to(cache[0])):q.stat().st_mtime_ns for q in cache[0].rglob('*') if q.is_file()}),indent=2)+'\n')
PY
}
run -version > /out/version.log 2>&1
case "$MODE" in
 missing)
  if run -audio=false -frames 1 > /out/missing.log 2>&1;then exit 1;fi
  python3 -I -c 'from pathlib import Path;import json;p=Path("/out");assert "1.10" in (p/"missing.log").read_text();assert not list((p/"data").rglob("GAME.EXE"));(p/"pass.json").write_text(json.dumps({"missing_original_rejected":True}))'
  ;;
 corrupt)
  run -audio=false -frames 1 > /out/install.log 2>&1
  snapshot before.json
  python3 -I -c 'from pathlib import Path;r=Path("/out/data/pod-cht/hd-cache");p=next(r.iterdir())/"intro/source/b01-f00.ops";Path("/out/saved.ops").write_bytes(p.read_bytes());p.write_bytes(b"wrong")'
  if run -audio=false -frames 1 > /out/corrupt-cache.log 2>&1;then exit 1;fi
  python3 -I -c 'from pathlib import Path;r=Path("/out/data/pod-cht");p=next((r/"hd-cache").iterdir())/"intro/source/b01-f00.ops";p.write_bytes(Path("/out/saved.ops").read_bytes());p=r/"game/INTRO.BMA";Path("/out/saved.bma").write_bytes(p.read_bytes());p.write_bytes(b"wrong")'
  if run -audio=false -frames 1 > /out/corrupt-original.log 2>&1;then exit 1;fi
  python3 -I -c 'from pathlib import Path;import json;p=Path("/out");(p/"data/pod-cht/game/INTRO.BMA").write_bytes((p/"saved.bma").read_bytes());assert "HD快取" in (p/"corrupt-cache.log").read_text();assert "INTRO.BMA" in (p/"corrupt-original.log").read_text();(p/"pass.json").write_text(json.dumps({"modified_cache_rejected":True,"modified_original_rejected":True}))'
  snapshot restored.json
  ;;
 intro)
  script='601:shot,700:shot,701:lang,702:shot,703:scale,704:shot,705:lang,706:shot,707:toggle,708:shot,709:scale,710:shot,711:toggle,712:shot,713:lang,1334:shot,1625:shot,1979:shot,2293:shot,2516:shot,2939:shot,3162:shot,3495:shot,3548:shot,3767:shot,3768:shot,4346:shot,4504:shot,4995:shot'
  run -audio=false -scale 3 -lang zh-TW -frames 5100 -script "$script" -flash-audit \
   -shot-dir /out/shots -shot-frame-dir /out/window -shot-source-dir /out/sources \
   -shot-frame /out/frame.png -shot-source /out/source > /out/run.log 2>&1
  snapshot first.json
  run -audio=false -scale 3 -lang zh-TW -frames 5100 -script "$script" -flash-audit \
   -shot-dir /out/restart-shots -shot-frame-dir /out/restart-window -shot-source-dir /out/restart-sources \
   -shot-frame /out/restart-frame.png -shot-source /out/restart-source > /out/restart.log 2>&1
  snapshot restart.json
  ;;
 demo)
  run -audio=false -scale 3 -lang zh-TW -frames 3800 -script '900: ,1800:d,2050:shot,2300:x,2500:shot,3000:shot,3200:p,3700:shot' \
   -flash-audit -shot-dir /out/shots -shot-frame /out/frame.png -shot-source /out/source > /out/run.log 2>&1
  snapshot first.json
  ;;
 save)
  run -audio=false -frames 1 > /out/install.log 2>&1
  test -f /seed/savgamA.pty;test ! -f /seed/savgamB.pty
  mkdir -p /out/data/pod-cht/game/SAVE
  cp -r /seed/. /out/data/pod-cht/game/SAVE/
  python3 -I -c 'from pathlib import Path;import os;r=Path("/out/data/pod-cht/game");[os.utime(p,(946684800,946684800)) for p in r.rglob("*") if p.is_file()]'
  run -audio=false -scale 2 -lang zh-TW -posters - -frames 4700 \
   -script '900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:A,3400:Down,3420:Down,3440:Down,3460:Down,3480:Down,3500:Return,3900:shot,4000:Y,4010:O,4020:U,4030:Return,4200:shot,4300:E,4400:S,4450:shot,4500:B,4550:shot,4650:shot' \
   -shot-dir /out/save-shots -shot-source /out/save-source > /out/save.log 2>&1
  test -s /out/data/pod-cht/game/SAVE/savgamB.pty
  snapshot saved.json
  run -audio=false -scale 2 -lang zh-TW -posters - -frames 3250 -script '900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:B,3200:shot' \
   -shot-dir /out/load-shots -shot-source /out/load-source > /out/load.log 2>&1
  snapshot loaded.json
  ;;
 *) exit 2;;
esac
find /out -user root -o -type d -name '*.md'
