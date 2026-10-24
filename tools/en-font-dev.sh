#!/usr/bin/env bash
# 沿前作規格015建置英文方形字格；來源與授權見docs/re/014-vga-hd.md。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
for d in "$ROOT/tools" "$ROOT/workplace/fonts-src/courier" "$ROOT/workplace/font"; do test -d "$d"; done
test -f "$ROOT/workplace/fonts-src/courier/CourierPrime-Bold.ttf"
timeout 60s docker run --rm --network none --memory 512m --cpus 1 --pids-limit 64 \
  --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 \
  -v "$ROOT/tools:/tools:ro" -v "$ROOT/workplace/fonts-src/courier:/source:ro" -v "$ROOT/workplace/font:/out" \
  sosb-img:r1 sh -eu -c '
    test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
    python -I /tools/en_font.py --ttf /source/CourierPrime-Bold.ttf --cell 16 --out /out/pod-en.golemfnt
    python -I /tools/en_font.py --ttf /source/CourierPrime-Bold.ttf --cell 24 --out /out/pod-en-3x.golemfnt
    find /out -user root -o -type d -name "*.md"
  '
