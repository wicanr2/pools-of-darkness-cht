#!/usr/bin/env bash
# G2 冷啟動收據（docs/re/003）。照 START.BAT 的順序在 dosgolem 重現：
#   1. STARTUP.EXE 首次執行設定：音效 [2] Adlib、顯示 [3] VGA、輸入 [2] MOUSE、兩個存檔路徑按 Enter，
#      寫出 scratch 層的 pool4.cfg 與 snd.com（ADLIB.COM 的複本），到 #60,000,000 前以 AH=4Ch 結束。
#   2. snd.com 常駐（以位元組相同的 ADLIB.COM 為疊底程式），監督佇列接 CONTROL.EXE；
#      -root-tsr-shrink（dosgolem 規格 431）讓 CONTROL 緊接在常駐區之後，與真 DOS 一致。
#      CONTROL.EXE 以 EXEC 叫起 VACE.EXE 播片頭（mode 13h），#160,000,000 送空白鍵跳過，接著 GAME.EXE。
#      到 #300,000,000（製作群畫面）存狀態。
#   3. 從狀態送 p（PLAY）重播兩次到 #400,000,000，比對建立隊伍主選單的色號雜湊。
# 注意：-save-state／-shots 的步數須小於 -steps。
# 前置：workplace/input/dos/pod（G1 展開）、workplace/dosgolem（dosgolem/pod 分支，含規格 430、431）。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="$ROOT/workplace/probe/g2-receipt"
P="$ROOT/tools/probe.sh"
rm -rf "$OUT"
mkdir -p "$OUT/scratch"

# 1. 首次執行設定。
PROBE_EXE=STARTUP.EXE "$P" "$OUT" -scratch-dir /out/scratch -steps 60000000 \
  -bios-keys $'232\r\r' -bios-key-from 9500000 -bios-key-every 5000000 > "$OUT/startup.log" 2>&1
grep -q '停止原因：程式呼叫 int 21h AH=4Ch，離開碼 0' "$OUT/startup.log" || { echo "STARTUP.EXE 沒有正常結束" >&2; exit 1; }
cmp -s "$OUT/scratch/snd.com" "$ROOT/workplace/input/dos/pod/ADLIB.COM" || { echo "snd.com 與 ADLIB.COM 不同" >&2; exit 1; }

# 2. 常駐驅動＋CONTROL.EXE 到製作群畫面。
PROBE_EXE=ADLIB.COM "$P" "$OUT" -scratch-dir /out/scratch -queue CONTROL.EXE -root-tsr-shrink \
  -steps 300001000 -bios-keys ' ' -bios-key-from 160000000 -bios-key-every 100000000 \
  -save-state 300000000:/out/credits-300m.state > "$OUT/cold.log" 2>&1

# 3. 送 p 進主選單，重播兩次。
for tag in a b; do
  PROBE_EXE=ADLIB.COM "$P" "$OUT" -scratch-dir /out/scratch -root-tsr-shrink \
    -load-state /out/credits-300m.state -steps 400001000 \
    -bios-keys 'p' -bios-key-from 330000000 -bios-key-every 50000000 \
    -shots "399000000:/out/menu-$tag.raw" > "$OUT/menu-$tag.log" 2>&1
done
cd "$OUT"
sha256sum scratch/pool4.cfg scratch/snd.com credits-300m.state menu-a.raw menu-b.raw menu-a.pal
a=$(sha256sum < menu-a.raw); b=$(sha256sum < menu-b.raw)
[[ "$a" == "$b" ]] && echo "重播一致" || { echo "重播不一致" >&2; exit 1; }
