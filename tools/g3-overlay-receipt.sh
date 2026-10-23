#!/usr/bin/env bash
# G3 覆繪收據（docs/re/004）：mode 13h 主選單以路徑 A、C 覆繪成中文。
# 譯文與字型是佔位：從銀色匕首之謎、幽靈騎士的 ui-text.zh-TW.tsv 取 key 相同者（tools/trial_text.py），
# 字型複製銀色匕首之謎的開發字型 sosb-rounded.golemfnt。兩者都只放 workplace/，不是本作成品。
# 從 G2 收據的 credits-300m.state 送 p，#399,000,000 輸出原版與覆繪；要求 6 筆疊字、範圍外差異 0、原版色號與 G2 收據相同。
# 前置：tools/g2-startup-receipt.sh、text/ui-text-events.tsv（tools/ui_text_catalog.py）、workplace/dosgolem 含 apps/pod。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GB="$(cd "$ROOT/.." && pwd)"
R="$ROOT/workplace/probe/g2-receipt"
OUT="$ROOT/workplace/probe/g3-overlay"
for f in "$R/credits-300m.state" "$ROOT/text/ui-text-events.tsv" \
  "$GB/secret-of-the-sliver-blader/text/ui-text.zh-TW.tsv" "$GB/death_knight_of_krynn/text/ui-text.zh-TW.tsv" \
  "$GB/secret-of-the-sliver-blader/workplace/font/sosb-rounded.golemfnt"; do
  test -f "$f" || { echo "缺 $f" >&2; exit 1; }
done
mkdir -p "$ROOT/workplace/font" "$ROOT/workplace/text-src/trial" "$ROOT/workplace/probe/states"
cp "$GB/secret-of-the-sliver-blader/workplace/font/sosb-rounded.golemfnt" "$ROOT/workplace/font/placeholder-sosb-rounded.golemfnt"
cp "$R/credits-300m.state" "$ROOT/workplace/probe/states/"
timeout 120 docker run --rm --network none -u "$(id -u):$(id -g)" --memory 256m --cpus 1 --pids-limit 16 \
  --log-opt max-size=10m --log-opt max-file=3 -v "$ROOT":/in -w /in \
  -v "$GB/secret-of-the-sliver-blader/text":/sosb:ro -v "$GB/death_knight_of_krynn/text":/dkk:ro sosb-tools:bookworm-r1 \
  python3 -I tools/trial_text.py text/ui-text-events.tsv workplace/text-src/trial/ui-text.zh-TW.tsv \
    sosb=/sosb/ui-text.zh-TW.tsv dkk=/dkk/ui-text.zh-TW.tsv
rm -rf "$OUT"
mkdir -p "$OUT/scratch"
cp "$R/scratch/"* "$OUT/scratch/"
PROBE_CMD=./cmd/pod-overlay-shot "$ROOT/tools/probe.sh" "$OUT" -state /st/credits-300m.state -root /orig -scratch /out/scratch \
  -font /font/placeholder-sosb-rounded.golemfnt -events /text/ui-text-events.tsv -translations /textsrc/trial/ui-text.zh-TW.tsv \
  -bios-keys 'p' -bios-key-from 330000000 -bios-key-every 50000000 -shots 399000000:menu -until 400000000 -out /out \
  > "$OUT/run.json" 2> "$OUT/run.err"
grep -q '"indexed_sha256": "fefa05142448782c528128c506ffb6ce7fb7c3f39c72dd2e420016e0c1fc11f4"' "$OUT/run.json" || { echo "原版色號與 G2 收據不同" >&2; exit 1; }
grep -q '"diff_outside_stamps": 0,' "$OUT/run.json" || { echo "疊字矩形外有差異" >&2; exit 1; }
grep -q '"stamps_shown": 6,' "$OUT/run.json" || { echo "疊字不是 6 筆" >&2; exit 1; }
grep -E '"(base|overlay)_sha256"' "$OUT/run.json"
echo "覆繪收據通過"
