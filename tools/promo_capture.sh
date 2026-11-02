#!/usr/bin/env bash
# Docker內正常CONTROL開場擷取；只做前端切換，不送改變客體流程的鍵。
set -euo pipefail
test -f /.dockerenv
test ! -e /out/intro
mkdir /out/intro /out/intro/home
export HOME=/out/intro/home GOMAXPROCS=2 POD_NO_DIALOG=1
export POD_PROMO_PCM=/out/intro/audio.s16le POD_PROMO_OPL=/out/intro/opl.jsonl
script=$(python3 -I -c 'n=set(range(550,751,4))|set(range(1250,1421,4))|{601,700,1300,1334,1362,1625,1952,1962,1972,1982,1992,2293,2516,2939,3162,3495,3548,3768,4346,4504,4995}; print(",".join(str(i)+":shot" for i in sorted(n)))')
script="$script,610:toggle,670:toggle,1300:scale,1360:scale,1950:lang,1960:lang,1970:lang,1980:lang,1990:lang"
printf '%s\n' "$script" > /out/intro/script.txt
xvfb-run -a /out/pod-promo-capture -original /original -save /out/intro/save \
 -text-dir /main/text -font /assets/font/pod-rounded.golemfnt -lang zh-TW -audio=false \
 -scale 3 -hd /assets/theme -posters /assets/theme/posters -frames 5100 -script "$script" \
 -shot-dir /out/intro/shots -shot-frame-dir /out/intro/window -shot-source-dir /out/intro/sources \
 -shot-frame /out/intro/frame.png -shot-source /out/intro/final > /out/intro/run.log 2>&1
tail -n 6 /out/intro/run.log
find /out/intro -user root -o -type d -name '*.md'
