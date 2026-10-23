#!/usr/bin/env bash
# G7 手札引用收據（規格 007，docs/re/009）：從 g6-town/town.state（菲蘭城讀檔後的街上）以 pod-overlay-shot 重播
# tools/g7-journal-keys.txt 的送鍵（tools/g6-explore.sh town1 的探索紀錄，247 鍵），走到叛徒門的守衛，
# 路徑 B 印「This is recorded as journal entry 57.」。觸發前、觸發後各截一張：
# 觸發前沒有手札面板；觸發後面板開著、引用句以組句模板覆繪；兩張的範圍外差異都是 0（面板範圍不算）。
# 另以 -journal-preview 57 開 F5 翻閱模式截一張。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
T="$ROOT/workplace/probe/g6-town"
OUT="$ROOT/workplace/probe/g7-journal-receipt"
test -f "$T/town.state" && test -d "$T/scratch" || { echo "缺 $T/town.state 或 scratch/（見 docs/re/010）" >&2; exit 1; }
rm -rf "$OUT"; mkdir -p "$OUT"; cp -r "$T/scratch" "$OUT/scratch"; test -d "$OUT/scratch"
keys=$(paste -sd, "$ROOT/tools/g7-journal-keys.txt")
cd "$ROOT/workplace/dosgolem"
DOSGOLEM_ORIG="$ROOT/workplace/input/dos/pod" \
DOSGOLEM_EXTRA_MOUNTS="$ROOT/workplace/probe:/st:ro;$OUT:/out;$ROOT/workplace/font:/font:ro;$ROOT/text:/text:ro" \
DOSGOLEM_TIMEOUT=3600s DOSGOLEM_CPUS=2 DOSGOLEM_MEM=2g \
  tools/go.sh run ./cmd/pod-overlay-shot -state /st/g6-town/town.state -root /orig -scratch /out/scratch \
    -font /font/pod-rounded.golemfnt -text-dir /text -keys-at "$keys" \
    -shots "9098000000:before,9120000000:popup" -until 9121000000 -out /out -inventory /out/inv.tsv \
    > "$OUT/shot.json" 2> "$OUT/shot.err" || { echo "pod-overlay-shot 失敗（$OUT/shot.err）" >&2; exit 1; }
n=$(grep -c '"diff_outside_stamps": 0,' "$OUT/shot.json" || true)
[[ $n == 2 ]] || { echo "範圍外差異不是 0（$OUT/shot.json）" >&2; exit 1; }
panels=$(grep -o '"journal_panel": true' "$OUT/shot.json" | wc -l)
[[ $panels == 1 ]] || { echo "手札面板應只在觸發後開著，得到 $panels 張" >&2; exit 1; }
grep -q $'\tB\t.*\ttemplate\t\tThis is recorded as journal entry 57\.' "$OUT/inv.tsv" || { echo "引用句沒有以組句模板覆繪" >&2; exit 1; }
# F5 翻閱模式（-journal-preview）：同一街上狀態直接開第 57 則，面板開著、範圍外差異 0
rm -rf "$OUT/scratch"; cp -r "$T/scratch" "$OUT/scratch"
DOSGOLEM_ORIG="$ROOT/workplace/input/dos/pod" \
DOSGOLEM_EXTRA_MOUNTS="$ROOT/workplace/probe:/st:ro;$OUT:/out;$ROOT/workplace/font:/font:ro;$ROOT/text:/text:ro" \
DOSGOLEM_TIMEOUT=900s DOSGOLEM_CPUS=2 DOSGOLEM_MEM=2g \
  tools/go.sh run ./cmd/pod-overlay-shot -state /st/g6-town/town.state -root /orig -scratch /out/scratch \
    -font /font/pod-rounded.golemfnt -text-dir /text -journal-preview 57 -shots "1550000000:browse" -until 1551000000 -out /out \
    > "$OUT/browse.json" 2> "$OUT/browse.err" || { echo "翻閱預覽失敗（$OUT/browse.err）" >&2; exit 1; }
grep -q '"diff_outside_stamps": 0,' "$OUT/browse.json" && grep -q '"journal_panel": true' "$OUT/browse.json" \
  || { echo "翻閱模式沒有開出面板" >&2; exit 1; }
echo "G7 手札收據完成：第 57 則，觸發後面板開著、翻閱模式可開（截圖在 $OUT）"
