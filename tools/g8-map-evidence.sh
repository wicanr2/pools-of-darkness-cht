#!/usr/bin/env bash
# G8 地圖記憶體證據。只重播 M、左轉、前進兩步、右轉三次，不增加探索路線。
# 輸入沿用正常玩家路徑的 g6-town/town.state 與 town1/end.state。
# 分析、編譯、重播、圖片轉換與擁有權檢查都在 Docker 內。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DG="$ROOT/workplace/dosgolem"
PROBES="$ROOT/workplace/probe"
ORIG="$ROOT/workplace/input/dos/pod"
OUT="$PROBES/g8-map"
for d in "$DG" "$PROBES/g6-town/scratch" "$PROBES/g6-explore/town1/scratch" "$ORIG"; do test -d "$d"; done
for f in "$PROBES/g6-town/town.state" "$PROBES/g6-explore/town1/end.state" "$ORIG/GEO1.DAX"; do test -f "$f"; done
mkdir -p "$OUT"
timeout 180s docker run --rm --init --network none --memory 2g --cpus 2 --pids-limit 128 \
  --log-opt max-size=10m --log-opt max-file=3 --user "$(id -u):$(id -g)" \
  -v "$DG:/src" -v "$PROBES:/states:ro" -v "$OUT:/out" -v "$ORIG:/orig:ro" \
  -e HOME=/tmp -e GOCACHE=/src/workplace/gocache -e GOMODCACHE=/src/workplace/gomodcache -e GOMAXPROCS=2 \
  -w /src dosgolem-go-ebiten:1.24-r1 sh -eu -c '
    test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
    go build -p 2 -o /src/workplace/bin/probe ./cmd/probe
    rm -rf /out/scratch /out/gate-scratch
    mkdir -p /out/scratch /out/gate-scratch
    cp -r /states/g6-town/scratch/. /out/scratch/
    cp -r /states/g6-explore/town1/scratch/. /out/gate-scratch/
    /src/workplace/bin/probe -exe /orig/ADLIB.COM -root /orig -scratch-dir /out/scratch \
      -root-tsr-shrink -load-state /states/g6-town/town.state -steps 1711000000 \
      -bios-keys M -bios-key-names Left,Up,Up,Right,Right,Right \
      -bios-key-from 1502000000 -bios-key-every 30000000 \
      -dump-mem-at "1501000000:lin:00000:1048576:/out/start.mem;1559000000:lin:00000:1048576:/out/left.mem;1589000000:lin:00000:1048576:/out/east1.mem;1619000000:lin:00000:1048576:/out/east2.mem;1649000000:lin:00000:1048576:/out/south.mem;1679000000:lin:00000:1048576:/out/west.mem;1709000000:lin:00000:1048576:/out/north.mem" \
      -shots "1501000000:/out/start.raw,1559000000:/out/left.raw,1589000000:/out/east1.raw,1619000000:/out/east2.raw,1649000000:/out/south.raw,1679000000:/out/west.raw,1709000000:/out/north.raw" \
      > /out/move.log 2>&1
    /src/workplace/bin/probe -exe /orig/ADLIB.COM -root /orig -scratch-dir /out/gate-scratch \
      -root-tsr-shrink -load-state /states/g6-explore/town1/end.state -steps 9400028068 \
      -dump-mem 00000-100000:/out/gate.mem -dump-vram /out/gate.bin > /out/gate.log 2>&1
    go version > /out/tool-version.txt
    git rev-parse HEAD >> /out/tool-version.txt
    find /out -user root -o -type d -name "*.md"
  '
timeout 30s docker run --rm --network none --memory 512m --cpus 1 --pids-limit 32 \
  --log-opt max-size=10m --log-opt max-file=3 --user "$(id -u):$(id -g)" \
  -v "$ROOT/tools:/tools:ro" -v "$PROBES:/states:ro" -v "$OUT:/out" -v "$ORIG:/orig:ro" \
  sosb-img:r1 python3 -I /tools/map_evidence.py --out /out --orig /orig --states /states
