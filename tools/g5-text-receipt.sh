#!/usr/bin/env bash
# G5 譯文收據：以正式譯文表（text/ui-text.zh-TW.tsv、text/ecl-text.zh-TW.tsv）與開發字型跑 DEMO，開場旁白到戰鬥。
# 從 G2 收據的 credits-300m.state 送空白鍵進 DEMO；5 張截圖範圍外差異都要是 0，開場兩張各 9 筆疊字。
# 前置：tools/g2-startup-receipt.sh、tools/font-dev.sh、workplace/dosgolem 含 apps/pod。
# 戰鬥畫面的怪物名、隊員名、物品名組句仍是英文，屬 G6（#7）。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
R="$ROOT/workplace/probe/g2-receipt"
OUT="$ROOT/workplace/probe/g5-full"
for f in "$R/credits-300m.state" "$ROOT/text/ui-text.zh-TW.tsv" "$ROOT/text/ecl-text.zh-TW.tsv" "$ROOT/workplace/font/pod-rounded.golemfnt"; do
  test -f "$f" || { echo "缺 $f" >&2; exit 1; }
done
mkdir -p "$ROOT/workplace/probe/states"
cp "$R/credits-300m.state" "$ROOT/workplace/probe/states/"
rm -rf "$OUT"
mkdir -p "$OUT/scratch"
cp "$R/scratch/"* "$OUT/scratch/"
PROBE_CMD=./cmd/pod-overlay-shot "$ROOT/tools/probe.sh" "$OUT" -state /st/credits-300m.state -root /orig -scratch /out/scratch \
  -font /font/pod-rounded.golemfnt -events /text/ui-text-events.tsv -translations /text/ui-text.zh-TW.tsv \
  -ecl-events /text/ecl-text-events.tsv -ecl-translations /text/ecl-text.zh-TW.tsv \
  -bios-keys ' ' -bios-key-from 330000000 \
  -shots '350000000:welcome,440000000:peace,700000000:d700,1000000000:d1000,1190000000:d1190' -until 1200000000 -out /out \
  > "$OUT/run.json" 2> "$OUT/run.err"
test "$(grep -c '"diff_outside_stamps": 0,' "$OUT/run.json")" = 5 || { echo "有截圖在疊字矩形外有差異" >&2; exit 1; }
test "$(grep -c '"stamps_shown": 9,' "$OUT/run.json")" -ge 2 || { echo "開場旁白疊字不足" >&2; exit 1; }
grep -E '"(name|overlay_sha256)"' "$OUT/run.json" | paste - -
echo "G5 譯文收據通過"
