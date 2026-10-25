#!/usr/bin/env bash
# 既有讀槽A路線後由正常紮營選單另存空槽B，再冷啟動讀回B；不注入狀態。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:?legacy|native}"
case "$MODE" in legacy) BINARY=pod-play-direct;; native) BINARY=pod-play-whole;; *) exit 2;; esac
for d in workplace/dosgolem/workplace/bin workplace/input/dos/pod workplace/font text tools workplace/probe/g6-saveload-receipt/scratch/SAVE; do test -d "$ROOT/$d"; done
SEED="$ROOT/workplace/probe/g6-saveload-receipt/scratch/SAVE"
test -f "$SEED/savgamA.pty"
test ! -e "$SEED/savgamB.pty"
test -f "$ROOT/workplace/dosgolem/workplace/bin/$BINARY"
OUT="$ROOT/workplace/probe/g9-hd/whole-save/$MODE"
test ! -e "$OUT" || { echo '保留舊收據，請另定名稱' >&2; exit 2; }
mkdir -p "$OUT"
timeout 1200s docker run --rm --init --label pod.task=whole-save --network none --memory 2g --cpus 2 --pids-limit 256 \
 --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 \
 -e HOME=/tmp -e POD_NO_DIALOG=1 -e GOMAXPROCS=2 -e BINARY="$BINARY" \
 -v "$ROOT/workplace/dosgolem/workplace/bin:/app:ro" -v "$ROOT/workplace/input/dos/pod:/orig:ro" \
 -v "$ROOT/workplace/font:/font:ro" -v "$ROOT/text:/text:ro" -v "$SEED:/seed:ro" -v "$ROOT/tools:/tools:ro" -v "$OUT:/out" \
 dosgolem-go-ebiten:1.24-r1 sh -eu -c '
 test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
 xvfb-run -a "/app/$BINARY" -original /orig -save /out/game -text-dir /text -font /font/pod-rounded.golemfnt -audio=false -frames 1 > /out/install.log 2>&1
 rm -rf /out/game/SAVE
 mkdir -p /out/game/SAVE
 cp -r /seed/. /out/game/SAVE/
 python3 -I -c "import os,pathlib,json,hashlib;root=pathlib.Path(\"/out/game\");[(os.utime(p,(946684800,946684800)))for p in root.rglob(\"*\")if p.is_file()];files={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest()for p in root.rglob(\"*\")if p.is_file()};files[\"binary_sha256\"]=hashlib.sha256(pathlib.Path(\"/app/\"+os.environ[\"BINARY\"]).read_bytes()).hexdigest();pathlib.Path(\"/out/before.json\").write_text(json.dumps(files,indent=2))"
 xvfb-run -a "/app/$BINARY" -original /orig -save /out/game -text-dir /text -font /font/pod-rounded.golemfnt \
  -audio=false -scale 2 -lang zh-TW -frames 4700 -posters - \
  -script "900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:A,3400:Down,3420:Down,3440:Down,3460:Down,3480:Down,3500:Return,3900:shot,4000:Y,4010:O,4020:U,4030:Return,4200:shot,4300:E,4400:S,4450:shot,4500:B,4550:shot,4650:shot" \
  -shot-dir /out/save-shots -shot /out/save-final.rgba -shot-source /out/save-source -draw-report /out/save-draw.json > /out/save.log 2>&1
 test -s /out/game/SAVE/savgamB.pty
 python3 -I -c "import pathlib,hashlib,json;root=pathlib.Path(\"/out/game/SAVE\");pathlib.Path(\"/out/after-save.json\").write_text(json.dumps({p.name:hashlib.sha256(p.read_bytes()).hexdigest()for p in root.iterdir()if p.is_file()},indent=2))"
 xvfb-run -a "/app/$BINARY" -original /orig -save /out/game -text-dir /text -font /font/pod-rounded.golemfnt \
  -audio=false -scale 2 -lang zh-TW -frames 3250 -posters - \
  -script "900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:B,3200:shot" \
  -shot-dir /out/load-shots -shot /out/load-final.rgba -shot-source /out/load-source -draw-report /out/load-draw.json > /out/load.log 2>&1
 find /out -user root -o -type d -name "*.md"
 '
