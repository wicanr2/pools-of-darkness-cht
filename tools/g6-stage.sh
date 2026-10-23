# G6 收據共用的分段函式（由 tools/g6-create-receipt.sh 等 source）。呼叫端先設 ROOT、OUT、ST。
# stage 名稱 上一段 終點步數 其餘參數…
#   以探針（START.BAT 鏈：snd.com 常駐、-root-tsr-shrink）從 $ST/g6r-<上一段>.state 送鍵、在終點存 $ST/g6r-<名稱>.state；
#   再以 pod-overlay-shot 從同一狀態重播同樣的鍵（用該段開始前的暫存目錄快照），在終點前 100 萬步截原版與中文圖、寫印字盤點。
#   PROBE_ONLY 陣列是只給探針的參數（例 -call-args），用過即清。範圍外差異不是 0 就停。
PROBE_ONLY=()
stage() {
  local name=$1 from=$2 end=$3; shift 3
  rm -rf "$OUT/pre-$name"; cp -r "$OUT/scratch" "$OUT/pre-$name"
  PROBE_EXE=ADLIB.COM "$ROOT/tools/probe.sh" "$OUT" -scratch-dir /out/scratch -root-tsr-shrink -load-state "/st/g6r-$from.state" \
    -steps $((end + 1000)) -save-state "$end:/out/$name.state" "$@" "${PROBE_ONLY[@]}" > "$OUT/$name.log" 2>&1
  PROBE_ONLY=()
  cp "$OUT/$name.state" "$ST/g6r-$name.state"
  rm -rf "$OUT/shot-scratch"; cp -r "$OUT/pre-$name" "$OUT/shot-scratch"
  PROBE_CMD=./cmd/pod-overlay-shot "$ROOT/tools/probe.sh" "$OUT" -state "/st/g6r-$from.state" -root /orig -scratch /out/shot-scratch \
    -font /font/pod-rounded.golemfnt -text-dir /text -shots "$((end - 1000000)):$name" -until "$end" -out /out \
    -inventory "/out/$name-inv.tsv" "$@" > "$OUT/$name-shot.json" 2> "$OUT/$name-shot.err"
  grep -q '"diff_outside_stamps": 0,' "$OUT/$name-shot.json" || { echo "$name：疊字矩形外有差異" >&2; exit 1; }
  printf '%-8s 疊字 %s、未命中 %s\n' "$name" "$(grep -o '"stamps_shown": [0-9]*' "$OUT/$name-shot.json" | grep -o '[0-9]*$')" \
    "$(awk -F'\t' '$8 ~ /未命中/ && $9 !~ /^[-0-9 \/]*$/' "$OUT/$name-inv.tsv" | wc -l)"
}
