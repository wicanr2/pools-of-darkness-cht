#!/usr/bin/env bash
# G6 存讀檔收據：接 tools/g6-create-receipt.sh 進入遊戲的狀態（菲蘭城碼頭，港務長問要不要導覽）。
# 婉拒導覽（N）→ 兩次 Enter 回到街上 → 紮營（E）→ 存檔（S）→ 槽 A；再從製作群狀態冷開機（同一暫存目錄）的主選單
# 載入儲存遊戲 → 本作（P）→ 槽 A → 開始冒險 → 查驗題（LIVES）→ 回到紮營畫面。
# 分段做法見 tools/g6-stage.sh。用法：tools/g6-saveload-receipt.sh [起始段]；起始段是 tour、street、encamp、save、slot、lmenu、load、lslot、lbegin、lanswer 之一。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
C="$ROOT/workplace/probe/g6-create-receipt"
OUT="$ROOT/workplace/probe/g6-saveload-receipt"
ST="$ROOT/workplace/probe/states"
test -f "$ST/g6r-answer.state" || { echo "缺 $ST/g6r-answer.state（先跑 tools/g6-create-receipt.sh）" >&2; exit 1; }
FROM="${1:-tour}"
CR=$'\r'
if [[ "$FROM" == tour ]]; then
  rm -rf "$OUT"; mkdir -p "$OUT"; cp -r "$C/scratch" "$OUT/scratch"
fi
source "$ROOT/tools/g6-stage.sh"
go=0
for s in tour street encamp save slot lmenu load lslot lbegin lanswer; do
  [[ "$s" == "$FROM" ]] && go=1
  [[ $go == 1 ]] || continue
  case $s in
    tour)   stage tour answer 2900000000 -bios-keys 'N' -bios-key-from 2701000000 ;;
    street) stage street tour 3100000000 -bios-key-names "Return,Return" -bios-key-from 2901000000 -bios-key-every 90000000 ;;
    encamp) stage encamp street 3300000000 -bios-keys 'E' -bios-key-from 3101000000 ;;
    save)   stage save encamp 3500000000 -bios-keys 'S' -bios-key-from 3301000000 ;;
    slot)   stage slot save 3800000000 -bios-keys 'A' -bios-key-from 3501000000 ;;
    lmenu)  cp "$ST/g6r-credits.state" "$ST/g6r-lcredits.state"
            stage lmenu lcredits 400000000 -bios-keys 'p' -bios-key-from 330000000 ;;
    load)   stage load lmenu 600000000 -bios-key-names "Down,Down,Return" -bios-key-from 401000000 -bios-key-every 30000000 ;;
    lslot)  stage lslot load 900000000 -bios-keys 'PA' -bios-key-from 601000000 -bios-key-every 100000000 ;;
    lbegin) PROBE_ONLY=(-call-args "1B2E:0354:6:900000000:1200000000" -call-length-string)
            stage lbegin lslot 1200000000 -bios-key-names "Down,Down,Down,Down,Down,Return" -bios-key-from 901000000 -bios-key-every 10000000 ;;
    lanswer) # 查驗題的讀法同 tools/g6-create-receipt.sh；答案查珍157 操作手冊附錄二，題目變了就停。
             q=$(grep -o 'string="[^"]*"' "$OUT/lbegin.log" | sed 's/^string="//; s/"$//' | tr '\n' '|')
             [[ "$q" == *"| journal, page | (4) "*"| what is word # |7 "*"| after the heading |ALIGNMENT|"* ]] || { echo "查驗題不是本收據預期的題目：$q" >&2; exit 1; }
             echo "查驗題：手札第 4 頁 ALIGNMENT 後第 7 字 → LIVES"
             stage lanswer lbegin 1500000000 -bios-keys "LIVES$CR" -bios-key-from 1201000000 -bios-key-every 5000000 ;;
  esac
done
echo "G6 存讀檔收據完成（截圖在 $OUT）"
