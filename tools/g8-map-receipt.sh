#!/usr/bin/env bash
# G8 地圖疊層的同狀態 A/B：東／西菲蘭城，2/3 倍，原版／走過／全圖。
# 用法：tools/g8-map-receipt.sh [east|west]，預設 east。
# 比較同一步數的原版色號、完整記憶體、暫存器與存檔；差異限面板內。
# 這是既有正常路徑狀態的收據，不代替 pod-play 冷開機操作驗收。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DG="$ROOT/workplace/dosgolem"
PROBES="$ROOT/workplace/probe"
OUT="$PROBES/g8-map-receipt"
ORIG="$ROOT/workplace/input/dos/pod"
case "${1:-east}" in
  east) STATE=g6-town/town.state; SCRATCH=g6-town/scratch; SHOT=1520000000; UNTIL=1521000000; MAP_KEY=1/10; X=11; Y=2; D=2 ;;
  west) STATE=g6-explore/town2/end.state; SCRATCH=g6-explore/town2/scratch; SHOT=21665000000; UNTIL=21666000000; MAP_KEY=1/11; X=9; Y=12; D=2; OUT="$OUT-west" ;;
  *) echo '用法：g8-map-receipt.sh [east|west]' >&2; exit 2 ;;
esac
for d in "$DG" "$ORIG" "$PROBES/$SCRATCH" "$ROOT/text" "$ROOT/workplace/font"; do test -d "$d"; done
for f in "$PROBES/$STATE" "$ROOT/text/map-geo.tsv" "$ROOT/workplace/font/pod-rounded.golemfnt"; do test -f "$f"; done
mkdir -p "$OUT"
timeout 180s docker run --rm --init --network none --memory 2g --cpus 2 --pids-limit 256 \
  --log-opt max-size=10m --log-opt max-file=3 --user "$(id -u):$(id -g)" \
  -v "$DG:/src" -v "$PROBES:/states:ro" -v "$OUT:/out" -v "$ORIG:/orig:ro" \
  -v "$ROOT/text:/text:ro" -v "$ROOT/workplace/font:/font:ro" \
  -e HOME=/tmp -e GOCACHE=/src/workplace/gocache -e GOMODCACHE=/src/workplace/gomodcache -e GOMAXPROCS=2 \
  -e STATE="$STATE" -e SCRATCH="$SCRATCH" -e SHOT="$SHOT" -e UNTIL="$UNTIL" \
  -w /src dosgolem-go-ebiten:1.24-r1 sh -eu -c '
    test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
    go build -p 2 -o workplace/bin/pod-overlay-shot ./cmd/pod-overlay-shot
    for scale in 2 3; do
      for mode in original visited all; do
        dir=/out/$scale-$mode
        rm -rf "$dir/scratch"
        mkdir -p "$dir/scratch"
        cp -r "/states/$SCRATCH/." "$dir/scratch/"
        case "$mode" in
          original) set -- -no-overlay ;;
          visited) set -- -map-preview 走過 ;;
          all) set -- -map-preview 全圖 ;;
        esac
        workplace/bin/pod-overlay-shot -state "/states/$STATE" -root /orig \
          -scratch "$dir/scratch" -text-dir /text -font /font/pod-rounded.golemfnt \
          -scale "$scale" -shots "$SHOT:map" -until "$UNTIL" -out "$dir" "$@" \
          > "$dir/result.json" 2> "$dir/error.log"
      done
    done
    find /out -user root -o -type d -name "*.md"
  '
timeout 30s docker run --rm --network none --memory 256m --cpus 1 --pids-limit 32 \
  --log-opt max-size=10m --log-opt max-file=3 --user "$(id -u):$(id -g)" \
  -v "$OUT:/out:ro" -v "$ROOT/tools:/tools:ro" python:3.13-slim python3 -I /tools/map_receipt_check.py /out --map-key "$MAP_KEY" --position "$X" "$Y" "$D"
