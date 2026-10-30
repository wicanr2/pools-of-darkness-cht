#!/usr/bin/env bash
# 僅在Docker執行；/project唯讀，/out為本批可寫收據目錄。
set -euo pipefail
test -f /.dockerenv
test -d /project
test -d /out
phase="${1:?before|after}"
case "$phase" in before) text=/out/text-before;; after) text=/project/text;; *) exit 2;; esac
dest="/out/gui-$phase"
test ! -e "$dest"
test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
mkdir "$dest"
binary=/project/workplace/dosgolem/workplace/bin/pod-play
args=(-original /project/workplace/input/dos/pod -save "$dest/game" -text-dir "$text" -font /project/workplace/font/pod-rounded.golemfnt -audio=false -posters -)
xvfb-run -a "$binary" "${args[@]}" -frames 1 > "$dest/install.log" 2>&1
rm -rf "$dest/game/SAVE"
mkdir "$dest/game/SAVE"
cp -r /project/workplace/probe/g6-saveload-receipt/scratch/SAVE/. "$dest/game/SAVE/"
python3 -I - "$dest" "$binary" "$text" <<'PY'
import os,pathlib,sys,json,hashlib
dest,binary,text=map(pathlib.Path,sys.argv[1:])
for p in (dest/'game').rglob('*'):
    if p.is_file(): os.utime(p,(946684800,946684800))
h=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
files={str(p.relative_to(dest/'game')):h(p)for p in (dest/'game').rglob('*')if p.is_file()}
(dest/'before.json').write_text(json.dumps({'files':files,'binary_sha256':h(binary),'guide_sha256':h(text/'walkthrough.zh-TW.tsv')},indent=2))
PY
script='900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:A,3400:Down,3420:Down,3440:Down,3460:Down,3480:Down,3500:Return,3900:shot,4000:Y,4010:O,4020:U,4030:Return,4200:shot,4300:map,4320:shot,4340:guide,4342:shot'
for ((ch=0;ch<36;ch++)); do script+=",$((4350+ch*2)):next"; done
script+=',4440:shot,4450:pgdn,4455:shot,4460:pgdn,4465:shot,4500:map,4520:shot,4600:map,4620:shot'
printf '%s\n' "$script" > "$dest/script.txt"
xvfb-run -a "$binary" "${args[@]}" -scale 2 -lang zh-TW -frames 4800 -script "$script" \
 -shot-dir "$dest/shots" -shot "$dest/final.rgba" -shot-source "$dest/source" -inventory "$dest/inv.tsv" \
 > "$dest/run.log" 2>&1
find "$dest" -user root -o -type d -name '*.md'
