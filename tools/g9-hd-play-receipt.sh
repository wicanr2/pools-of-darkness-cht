#!/usr/bin/env bash
# 正常冷啟動抽樣，不載 state、不改 guest：demo 擷取兩圖；switch 驗 F12/F2 對應動作。
# 使用現有隔離 dosgolem 的 pod-play binary。先在該專案 tools/go.sh build -o workplace/bin/pod-play ./cmd/pod-play。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-switch}"
TAG="${2:-frontend-$MODE}"
case "$TAG" in *[!A-Za-z0-9_-]*|'') echo '收據名稱只接受英數、底線與連字號' >&2; exit 2;; esac
case "$MODE" in demo|switch) ;; *) echo '用法：g9-hd-play-receipt.sh [demo|switch]' >&2; exit 2;; esac
for d in workplace/dosgolem/workplace/bin workplace/input/dos/pod workplace/font workplace/hd/theme text; do test -d "$ROOT/$d"; done
test -f "$ROOT/workplace/dosgolem/workplace/bin/pod-play"
OUT="$ROOT/workplace/probe/g9-hd/$TAG"
mkdir -p "$OUT"
timeout 900s docker run --rm --init --label pod.task=g9-hd-play --network none \
  --memory 2g --cpus 2 --pids-limit 256 --log-opt max-size=10m --log-opt max-file=3 \
  --user "$(id -u):$(id -g)" -e HOME=/tmp -e POD_NO_DIALOG=1 -e MODE="$MODE" \
  -v "$ROOT/workplace/dosgolem/workplace/bin:/app:ro" -v "$ROOT/workplace/input/dos/pod:/orig:ro" \
  -v "$ROOT/workplace/font:/font:ro" -v "$ROOT/text:/text:ro" -v "$ROOT/workplace/hd/theme:/theme:ro" \
  -v "$OUT:/out" dosgolem-go-ebiten:1.24-r1 sh -eu -c '
    test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
    if test "$MODE" = demo; then
      cases=demo
    else
      cases="base switched"
    fi
    for name in $cases; do
      mkdir -p /out/$name
      test ! -e /out/$name/save || { echo "已有收據，請保留後換名再重跑：/out/$name/save" >&2; exit 2; }
      if test "$name" = demo; then
        frames=5900
        script="900: ,1800: ,2050:shot,5600:shot"
        audit=-flash-audit
      elif test "$name" = base; then
        frames=2210
        script="900: ,1800: ,2050:shot,2200:shot"
        audit=""
      else
        frames=2210
        script="900: ,1800: ,2050:shot,2100:toggle,2102:shot,2140:toggle,2142:shot,2160:scale,2162:shot,2180:scale,2182:shot,2190:bar,2192:shot"
        audit=""
      fi
      xvfb-run -a /app/pod-play -original /orig -save /out/$name/save -text-dir /text \
        -font /font/pod-rounded.golemfnt -hd /theme -scale 3 -audio=false -frames "$frames" \
        -script "$script" $audit -shot-dir /out/$name/shots -shot /out/$name/final.rgba \
        -shot-frame-dir /out/$name/window-shots -shot-frame /out/$name/final-window.png > /out/$name/run.log 2>&1
    done
    find /out -user root -o -type d -name "*.md"
  '
if test "$MODE" = switch; then
  timeout 30s docker run --rm --network none --memory 512m --cpus 1 --pids-limit 64 \
    --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 \
    -v "$ROOT/tools:/tools:ro" -v "$OUT:/out" -v "$ROOT/workplace/hd/requests/b02:/req:ro" \
    -v "$ROOT/workplace/hd/out/b02:/hd:ro" sosb-img:r1 \
    python -I /tools/hd_vga_switch_check.py --out /out --req /req --hd /hd
fi
