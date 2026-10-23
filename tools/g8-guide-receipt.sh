#!/usr/bin/env bash
# G8 攻略面板收據（克萊恩英豪規格 008；由銀色匕首之謎 tools/g8-guide-receipt.sh 移植）：pod-play 冷開機到建立隊伍主選單
# （同 tools/play-smoke.sh），以腳本動作開 F6 攻略面板截圖。
# 組合：open（目錄）、next（第 1 段）、far（連換 FAR 段，預設 55，第 20 章 Moander Minis）。
# 檢查：log 沒有缺字；三張截圖彼此不同。docker run 帶 --init（xvfb-run 當 PID 1 時會卡住）。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DG="$ROOT/workplace/dosgolem"
ORIG="$ROOT/workplace/input/dos/pod"
FONT="$ROOT/workplace/font/pod-rounded.golemfnt"
OUT="$ROOT/workplace/probe/g8-guide-receipt"
IMAGE=dosgolem-go-ebiten:1.24-r1
FAR=${FAR:-55}
for f in "$ORIG/CONTROL.EXE" "$FONT" "$ROOT/text/walkthrough.zh-TW.tsv"; do test -f "$f" || { echo "缺 $f" >&2; exit 1; }; done
mkdir -p "$DG/workplace/bin"
(cd "$DG" && DOSGOLEM_TIMEOUT=10m tools/go.sh build -o /src/workplace/bin/pod-play ./cmd/pod-play)
rm -rf "$OUT"; mkdir -p "$OUT"
run() { # 名稱 腳本
  local name=$1 script=$2
  mkdir -p "$OUT/$name"
  timeout 15m docker run --rm --init --network none --memory 2g --cpus 2 --pids-limit 256 \
    --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" -e HOME=/tmp -e POD_NO_DIALOG=1 \
    -v "$DG/workplace/bin":/app:ro -v "$ORIG":/orig:ro -v "$ROOT/text":/text:ro -v "$ROOT/workplace/font":/font:ro \
    -v "$OUT/$name":/out --entrypoint xvfb-run "$IMAGE" -a /app/pod-play -original /orig -save /out/save -text-dir /text \
    -font "/font/$(basename "$FONT")" -audio=false -scale 2 -frames 2700 -script "900: ,1800:p,$script" \
    -shot /out/shot.rgba > "$OUT/$name/log.txt" 2>&1
}
nexts=""; for ((i = 0; i < FAR; i++)); do nexts+=",$((2410 + i * 4)):next"; done
run open 2400:guide &
run next 2400:guide,2450:next &
run far "2400:guide$nexts" &
wait
for n in open next far; do
  test -s "$OUT/$n/shot.rgba" || { echo "$n 沒有截圖（$OUT/$n/log.txt）" >&2; exit 1; }
  if grep -q '缺字' "$OUT/$n/log.txt"; then echo "$n 有缺字" >&2; exit 1; fi
done
timeout 120 docker run --rm --network none --memory 1g --cpus 1 --pids-limit 64 --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" -v "$OUT":/d sosb-img:r1 python3 -I -c "
from PIL import Image
for n in ('open', 'next', 'far'):
    Image.frombytes('RGBA', (640, 400), open(f'/d/{n}/shot.rgba', 'rb').read()).save(f'/d/{n}.png')"
[[ $(sha256sum "$OUT"/{open,next,far}/shot.rgba | cut -d' ' -f1 | sort -u | wc -l) == 3 ]] || { echo "三張截圖有重複" >&2; exit 1; }
sha256sum "$OUT"/{open,next,far}/shot.rgba | sed "s#$OUT/##"
