#!/usr/bin/env bash
# G7 查驗題收據（規格 006，docs/re/008）：
# 1. 由 GAME.EXE 映像重產題目表，與 text/quiz-answers.tsv 逐字相同；
# 2. 從存讀檔收據的 lslot 狀態重播到查驗題畫面（第 4 頁 ALIGNMENT 第 7 字），答案顯示與 -hide-answers 各截一張，
#    範圍外差異 0；顯示版第 6 列是「答案：LIVES」，隱藏版第 6 列空白。
# 顯示答案後作答進入遊戲的那一段在 tools/g6-create-receipt.sh（answer）與 tools/g6-saveload-receipt.sh（lanswer）。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
S="$ROOT/workplace/probe/g6-saveload-receipt"
OUT="$ROOT/workplace/probe/g7-quiz-receipt"
ST="$ROOT/workplace/probe/states"
test -f "$ST/g6r-lslot.state" || { echo "缺 $ST/g6r-lslot.state（先跑 tools/g6-saveload-receipt.sh）" >&2; exit 1; }
test -d "$S/pre-lbegin" || { echo "缺 $S/pre-lbegin（先跑 tools/g6-saveload-receipt.sh）" >&2; exit 1; }
test -f "$ROOT/workplace/text-src/game-image.bin" || { echo "缺 GAME.EXE 映像（先跑 tools/g3-game-image.sh）" >&2; exit 1; }
rm -rf "$OUT"; mkdir -p "$OUT"
timeout 300 docker run --rm --network none -u "$(id -u):$(id -g)" --memory 512m --cpus 1 --pids-limit 16 \
  --log-opt max-size=10m --log-opt max-file=3 -v "$ROOT":/in:ro -v "$OUT":/out -w /in sosb-tools:bookworm-r1 \
  python3 -I tools/quiz_table.py workplace/text-src/game-image.bin /out/quiz-answers.tsv
cmp "$OUT/quiz-answers.tsv" "$ROOT/text/quiz-answers.tsv" || { echo "題目表與 text/quiz-answers.tsv 不同" >&2; exit 1; }
for v in shown hidden; do
  flags=(); [[ $v == hidden ]] && flags=(-hide-answers)
  rm -rf "$OUT/shot-scratch"; cp -r "$S/pre-lbegin" "$OUT/shot-scratch"
  PROBE_CMD=./cmd/pod-overlay-shot "$ROOT/tools/probe.sh" "$OUT" -state /st/g6r-lslot.state -root /orig -scratch /out/shot-scratch \
    -font /font/pod-rounded.golemfnt -text-dir /text -shots "1199000000:$v" -until 1200000000 -out /out -inventory "/out/$v-inv.tsv" \
    "${flags[@]}" -bios-key-names "Down,Down,Down,Down,Down,Return" -bios-key-from 901000000 -bios-key-every 10000000 \
    > "$OUT/$v.json" 2> "$OUT/$v.err"
  grep -q '"diff_outside_stamps": 0,' "$OUT/$v.json" || { echo "$v：疊字矩形外有差異" >&2; exit 1; }
  row6=$(grep -o '@24,48 cells=19 [^"]*"[^"]*"' "$OUT/$v.json" | sed 's/^[^"]*"//; s/\\"$//; s/ *$//')
  case $v in
    shown)  [[ "$row6" == 答案：* ]] || { echo "顯示版第 6 列不是答案：$row6" >&2; exit 1; } ;;
    hidden) [[ -z "$row6" ]] || { echo "隱藏版第 6 列不是空白：$row6" >&2; exit 1; } ;;
  esac
  printf '%-7s 疊字 %s、第 6 列「%s」\n' "$v" "$(grep -o '"stamps_shown": [0-9]*' "$OUT/$v.json" | grep -o '[0-9]*$')" "$row6"
done
echo "G7 查驗題收據完成（截圖在 $OUT）"
