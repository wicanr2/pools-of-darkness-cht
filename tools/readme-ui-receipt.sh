#!/usr/bin/env bash
# README：沿已驗證的冷啟動讀檔路徑擷取地圖、攻略及工具列，不做新探索。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DG="$ROOT/workplace/dosgolem"
SEED="$ROOT/workplace/probe/g6-saveload-receipt/scratch/SAVE"
for d in "$DG" "$SEED" "$ROOT/workplace/input/dos/pod" "$ROOT/workplace/font" "$ROOT/workplace/hd/theme"; do test -d "$d"; done
test -f "$SEED/savgamA.pty"
TAG="${1:-readme-ui}"
case "$TAG" in *[!A-Za-z0-9_-]*|'') echo '收據名稱不合法' >&2; exit 2;; esac
OUT="$ROOT/workplace/probe/$TAG"
mkdir -p "$OUT"
test ! -e "$OUT/save" || { echo '既有收據保留，請明確更換輸出後再跑' >&2; exit 2; }
timeout 120s docker run --rm --init --network none --memory 2g --cpus 2 --pids-limit 256 \
  --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 \
  -v "$DG:/src" -e HOME=/tmp -e GOCACHE=/src/workplace/gocache -e GOMODCACHE=/src/workplace/gomodcache \
  -e GOMAXPROCS=2 -w /src dosgolem-go-ebiten:1.24-r1 \
  sh -eu -c 'gofmt -w cmd/pod-play/input.go cmd/pod-play/script_test.go; xvfb-run -a go test ./cmd/pod-play; go build -p 2 -o workplace/bin/pod-play-readme ./cmd/pod-play'
timeout 600s docker run --rm --init --label pod.task=readme-ui --network none --memory 2g --cpus 2 --pids-limit 256 \
  --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 \
  -v "$DG/workplace/bin:/app:ro" -v "$ROOT/workplace/input/dos/pod:/orig:ro" -v "$SEED:/seed:ro" \
  -v "$ROOT/workplace/font:/font:ro" -v "$ROOT/text:/text:ro" -v "$ROOT/workplace/hd/theme:/theme:ro" -v "$OUT:/out" \
  -e HOME=/tmp -e POD_NO_DIALOG=1 -e GOMAXPROCS=2 dosgolem-go-ebiten:1.24-r1 sh -eu -c '
  test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
  test ! -e /out/save
  xvfb-run -a /app/pod-play-readme -original /orig -save /out/save -text-dir /text -font /font/pod-rounded.golemfnt -audio=false -frames 1 > /out/install.log 2>&1
  # 只清除本輪全新安裝複製出的 SAVE，再匯入本作測試槽 A；原版與 seed 唯讀。
  rm -rf /out/save/SAVE
  mkdir -p /out/save/SAVE
  cp -r /seed/. /out/save/SAVE/
  xvfb-run -a /app/pod-play-readme -original /orig -save /out/save -text-dir /text \
    -font /font/pod-rounded.golemfnt -hd /theme -lang zh-TW -audio=false -scale 3 -frames 4800 \
    -script "900: ,1800:p,2400:Down,2450:Down,2500:Return,2800:P,3000:A,3400:Down,3420:Down,3440:Down,3460:Down,3480:Down,3500:Return,3900:shot,4000:Y,4010:O,4020:U,4030:Return,4200:shot,4300:map,4320:shot,4350:mapall,4360:shot,4400:guide,4420:shot,4500:map,4520:shot,4600:map,4620:shot,4700:bar,4720:shot,4740:speeds,4750:shot,4760:speeds,4770:barfonts,4780:shot" \
    -shot-dir /out/shots -shot-frame-dir /out/window-shots -shot-source-dir /out/sources \
    -shot-source /out/source -shot /out/final.rgba -inventory /out/inv.tsv > /out/run.log 2>&1
  find /out -user root -o -type d -name "*.md"
  '
