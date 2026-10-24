#!/usr/bin/env bash
# G8 正常前端收據：空目錄安裝，匯入本作已建立的槽 A，冷開機讀檔，回答實際觀測的查驗題 YOU。
# 送鍵與 fixed 診斷收據相同；F7／F6 用前端既有 map／guide 腳本動作，功能鍵路由另有單元測試。
# 不載機器狀態、不寫入 guest 記憶體、不改題目或判定。失敗時保留 /out 的原始收據。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DG="$ROOT/workplace/dosgolem"
OUT="$ROOT/workplace/probe/g8-map-play"
ORIG="$ROOT/workplace/input/dos/pod"
SEED="$ROOT/workplace/probe/g6-saveload-receipt/scratch/SAVE"
for d in "$DG" "$SEED" "$ORIG" "$ROOT/text" "$ROOT/workplace/font"; do test -d "$d"; done
test -f "$SEED/savgamA.pty"
mkdir -p "$OUT"
timeout 90s docker run --rm --init --network none --memory 2g --cpus 2 --pids-limit 256 \
  --log-opt max-size=10m --log-opt max-file=3 --user "$(id -u):$(id -g)" \
  -v "$DG:/src" -e HOME=/tmp -e GOCACHE=/src/workplace/gocache -e GOMODCACHE=/src/workplace/gomodcache \
  -e GOMAXPROCS=2 -w /src dosgolem-go-ebiten:1.24-r1 go build -p 2 -o workplace/bin/pod-play ./cmd/pod-play
timeout 300s docker run --rm --init --label pod.task=g8-map-play --network none --memory 2g --cpus 2 --pids-limit 256 \
  --log-opt max-size=10m --log-opt max-file=3 --user "$(id -u):$(id -g)" \
  -v "$DG/workplace/bin:/app:ro" -v "$ORIG:/orig:ro" -v "$SEED:/seed:ro" -v "$OUT:/out" \
  -v "$ROOT/text:/text:ro" -v "$ROOT/workplace/font:/font:ro" \
  -e HOME=/tmp -e POD_NO_DIALOG=1 -e GOMAXPROCS=2 dosgolem-go-ebiten:1.24-r1 sh -eu -c '
    test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
    rm -rf /out/save /out/shots
    xvfb-run -a /app/pod-play -original /orig -save /out/save -text-dir /text \
      -font /font/pod-rounded.golemfnt -audio=false -frames 1 > /out/install.log 2>&1
    # 清掉安裝複製的原始 SAVE，避免 1995 年大寫 CHRDATA 遮住測試檔的小寫名稱。
    rm -rf /out/save/SAVE
    mkdir -p /out/save/SAVE
    cp -r /seed/. /out/save/SAVE/
    xvfb-run -a /app/pod-play -original /orig -save /out/save -text-dir /text \
      -font /font/pod-rounded.golemfnt -audio=false -scale 2 -frames 4800 \
      -script "900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:A,3400:Down,3420:Down,3440:Down,3460:Down,3480:Down,3500:Return,3900:shot,4000:Y,4010:O,4020:U,4030:Return,4200:shot,4300:map,4320:shot,4400:guide,4420:shot,4500:map,4520:shot,4600:map,4620:shot" \
      -shot-dir /out/shots -shot /out/final.rgba -inventory /out/inv.tsv > /out/run.log 2>&1
    find /out -user root -o -type d -name "*.md"
  '
timeout 30s docker run --rm --network none --memory 512m --cpus 1 --pids-limit 32 \
  --log-opt max-size=10m --log-opt max-file=3 --user "$(id -u):$(id -g)" \
  -v "$OUT:/out" -v "$SEED:/seed:ro" -v "$ROOT/tools:/tools:ro" \
  -v "$DG/workplace/bin:/app:ro" -v "$ROOT/text:/text:ro" -v "$ROOT/workplace/font:/font:ro" \
  sosb-img:r1 python3 -I /tools/map_play_check.py --out /out --seed /seed \
    --binary /app/pod-play --text /text --font /font/pod-rounded.golemfnt
