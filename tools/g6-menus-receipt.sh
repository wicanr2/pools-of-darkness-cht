#!/usr/bin/env bash
# G6 遊戲內選單盤點：接 tools/g6-saveload-receipt.sh 讀檔後回到街上的狀態（lanswer），逐一打開不用走路就能到的選單。
# 檢視人物（V）→ 物品（I）→ 離開；紮營（E）→ 魔法（M）→ 離開、休息（R）→ 離開、設定（A）→ 離開；區域地圖（A）。
# 分段做法見 tools/g6-stage.sh。用法：tools/g6-menus-receipt.sh [起始段]。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
S="$ROOT/workplace/probe/g6-saveload-receipt"
OUT="$ROOT/workplace/probe/g6-menus-receipt"
ST="$ROOT/workplace/probe/states"
test -f "$ST/g6r-lanswer.state" || { echo "缺 $ST/g6r-lanswer.state（先跑 tools/g6-saveload-receipt.sh）" >&2; exit 1; }
FROM="${1:-view}"
if [[ "$FROM" == view ]]; then
  rm -rf "$OUT"; mkdir -p "$OUT"; cp -r "$S/scratch" "$OUT/scratch"
fi
source "$ROOT/tools/g6-stage.sh"
go=0
for s in view items vexit camp magic mexit rest rexit alter aexit cexit area; do
  [[ "$s" == "$FROM" ]] && go=1
  [[ $go == 1 ]] || continue
  case $s in
    view)  stage view lanswer 1650000000 -bios-keys 'V' -bios-key-from 1501000000 ;;
    items) stage items view 1800000000 -bios-keys 'I' -bios-key-from 1651000000 ;;
    vexit) stage vexit items 2000000000 -bios-keys 'EE' -bios-key-from 1801000000 -bios-key-every 80000000 ;;
    camp)  stage camp vexit 2150000000 -bios-keys 'E' -bios-key-from 2001000000 ;;
    magic) stage magic camp 2300000000 -bios-keys 'M' -bios-key-from 2151000000 ;;
    mexit) stage mexit magic 2450000000 -bios-keys 'E' -bios-key-from 2301000000 ;;
    rest)  stage rest mexit 2600000000 -bios-keys 'R' -bios-key-from 2451000000 ;;
    rexit) stage rexit rest 2750000000 -bios-keys 'E' -bios-key-from 2601000000 ;;
    alter) stage alter rexit 2900000000 -bios-keys 'A' -bios-key-from 2751000000 ;;
    aexit) stage aexit alter 3050000000 -bios-keys 'E' -bios-key-from 2901000000 ;;
    cexit) stage cexit aexit 3200000000 -bios-keys 'E' -bios-key-from 3051000000 ;;
    area)  stage area cexit 3350000000 -bios-keys 'A' -bios-key-from 3201000000 ;;
  esac
done
echo "G6 選單盤點完成（截圖在 $OUT）"
