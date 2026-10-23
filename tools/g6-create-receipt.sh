#!/usr/bin/env bash
# G6 正常玩家路徑收據：主選單 → 建角（精靈戰士 ROLAND）→ 存檔 → 加入隊伍（A 加入、E 離開）→ 開始冒險 → 查驗題作答（HAS）。
# 前置：tools/g2-startup-receipt.sh（製作群狀態 credits-300m.state 與首次設定寫出的 pool4.cfg、snd.com）。
# 分段做法見 tools/g6-stage.sh（探針存狀態，再以 pod-overlay-shot 用該段開始前的暫存目錄快照重播、截圖、寫印字盤點）。
# 建角：種族、性別、職業、陣營等各 Enter（第一項：精靈、男、戰士、嚴守善良），要重擲屬性嗎？按 N，名字 ROLAND，圖示編輯器 E、Y，存檔 Y。
# 用法：tools/g6-create-receipt.sh [起始段]；起始段是 menu、create1、create2、create3、create4、create、add1、add、party、begin、answer 之一（從該段續跑）。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
G2="$ROOT/workplace/probe/g2-receipt"
OUT="$ROOT/workplace/probe/g6-create-receipt"
ST="$ROOT/workplace/probe/states"
test -f "$G2/credits-300m.state" || { echo "缺 $G2/credits-300m.state（先跑 tools/g2-startup-receipt.sh）" >&2; exit 1; }
FROM="${1:-menu}"
CR=$'\r'
mkdir -p "$ST"
if [[ "$FROM" == menu ]]; then
  rm -rf "$OUT"; mkdir -p "$OUT/scratch"; cp "$G2/scratch/pool4.cfg" "$G2/scratch/snd.com" "$OUT/scratch/"
  cp "$G2/credits-300m.state" "$ST/g6r-credits.state"
fi
source "$ROOT/tools/g6-stage.sh"
go=0
for s in menu create1 create2 create3 create4 create add1 add party begin answer; do
  [[ "$s" == "$FROM" ]] && go=1
  [[ $go == 1 ]] || continue
  case $s in
    menu)    stage menu credits 400000000 -bios-keys 'p' -bios-key-from 330000000 ;;
    create1) stage create1 menu 760000000 -bios-keys "$CR$CR$CR$CR$CR" -bios-key-from 401000000 -bios-key-every 60000000 ;;
    create2) stage create2 create1 900000000 -bios-keys 'N' -bios-key-from 761000000 ;;
    create3) stage create3 create2 1100000000 -bios-keys "ROLAND$CR" -bios-key-from 901000000 -bios-key-every 5000000 ;;
    create4) stage create4 create3 1300000000 -bios-keys 'EY' -bios-key-from 1101000000 -bios-key-every 100000000 ;;
    create)  stage create create4 1500000000 -bios-keys 'Y' -bios-key-from 1301000000 ;;
    add1)    stage add1 create 1700000000 -bios-key-names "Down,Return" -bios-key-from 1501000000 -bios-key-every 60000000 ;;
    add)     stage add add1 1900000000 -bios-key-names "Return" -bios-key-from 1701000000 ;;
    party)   stage party add 2100000000 -bios-keys 'AE' -bios-key-from 1901000000 -bios-key-every 80000000 ;;
    begin)   PROBE_ONLY=(-call-args "1B2E:0354:6:2100000000:2400000000" -call-length-string)
             stage begin party 2400000000 -bios-key-names "Down,Down,Down,Down,Down,Down,Return" -bios-key-from 2101000000 -bios-key-every 10000000 ;;
    answer)  # 查驗題由路徑 A 分段印出：「 journal, page 」「 (頁) 」「 what is word # 」「字序 」「 after the heading 」「標題」（docs/re/001）。
             # 答案查珍157 操作手冊附錄二（不轉錄進 Git，docs/re/002）；這裡只收本收據會遇到的題，題目變了就停。
             q=$(grep -o 'string="[^"]*"' "$OUT/begin.log" | sed 's/^string="//; s/"$//' | tr '\n' '|')
             [[ "$q" == *"| journal, page | (2) "*"| what is word # |3 "*"| after the heading |ABILITY SCORES|"* ]] || { echo "查驗題不是本收據預期的題目：$q" >&2; exit 1; }
             echo "查驗題：手札第 2 頁 ABILITY SCORES 後第 3 字 → HAS"
             stage answer begin 2700000000 -bios-keys "HAS$CR" -bios-key-from 2401000000 -bios-key-every 5000000 ;;
  esac
done
ls "$OUT/scratch/SAVE" 2>/dev/null | head
echo "G6 建角收據完成（截圖在 $OUT）"
