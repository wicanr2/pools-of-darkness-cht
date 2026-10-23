#!/usr/bin/env bash
# dosgolem 探針包裝：tools/probe.sh <輸出目錄> <cmd/probe 參數…>
# 原版唯讀掛 /orig，輸出目錄掛 /out，狀態庫 workplace/probe/states 掛 /st（唯讀）。
# 固定 -root /orig；-program-path 預設 C:\POD\<PROBE_EXE>（安裝預設目錄 C:\POD，見 docs/re/003），
# PROBE_EXE 預設 STARTUP.EXE。scratch 由呼叫端用 -scratch-dir 指定。
# 換成別的 cmd：PROBE_CMD=./cmd/pod-xxx（此時不帶 -root／-program-path／-exe，由呼叫端給）。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="$(mkdir -p "$1" && cd "$1" && pwd)"; shift
mkdir -p "$ROOT/workplace/probe/states"
test -f "$ROOT/workplace/input/dos/pod/GAME.EXE" || { echo "缺原版（先跑 G1 的 tools/inventory.py）" >&2; exit 1; }
test -x "$ROOT/workplace/dosgolem/tools/go.sh" || { echo "缺 workplace/dosgolem（dosgolem 隔離 clone）" >&2; exit 1; }
CMD="${PROBE_CMD:-./cmd/probe}"
EXE="${PROBE_EXE:-STARTUP.EXE}"
FIXED=(-root /orig -exe "/orig/$EXE" -program-path "C:\\POD\\$EXE")
[[ "$CMD" == ./cmd/probe ]] || FIXED=()
cd "$ROOT/workplace/dosgolem"
DOSGOLEM_ORIG="$ROOT/workplace/input/dos/pod" \
DOSGOLEM_EXTRA_MOUNTS="$OUT:/out;$ROOT/workplace/probe/states:/st:ro" \
DOSGOLEM_TIMEOUT="${PROBE_TIMEOUT:-20m}" DOSGOLEM_CPUS="${PROBE_CPUS:-2}" DOSGOLEM_MEM="${PROBE_MEM:-2g}" \
  tools/go.sh run "$CMD" "${FIXED[@]}" "$@"
