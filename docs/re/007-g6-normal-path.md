# 007 G6 正常玩家路徑覆蓋

日期：2026-10-23。Issue：[#7](https://github.com/wicanr2/pools-of-darkness-cht/issues/7)。前置：[003](003-dosgolem-startup.md)、[004](004-output-paths.md)。

## 收據

| 收據 | 路徑 | 結果 |
|---|---|---|
| [`tools/play-smoke.sh`](../../tools/play-smoke.sh) | `pod-play` 空存檔目錄冷開機：照 `START.BAT` 以 `snd.com` 常駐、佇列 `CONTROL.EXE`（dosgolem 規格 432）→ 片頭 → 製作群 → 建立隊伍主選單 | 原版色號雜湊 `fefa0514…` 與 G2 收據相同；主選單覆繪 8 筆 |
| 同上，`10500 '900: ,1800: '` | 製作群後送空白進 DEMO，演到戰鬥 | 隊員名（辛尼漢爵士、艾爾弗瑞德、克莉絲卡）、物品名＋加值（釘頭錘 +3）、施法訊息、座標時間列都已覆繪；未命中只剩防護與生命力數字、一格空白 |
| [`tools/g6-create-receipt.sh`](../../tools/g6-create-receipt.sh) | 主選單 → 建角（精靈戰士 ROLAND）→ 存檔 → 加入隊伍 → 開始冒險 → 查驗題 → 菲蘭城碼頭（港務長） | 11 段範圍外差異 0；未命中只剩傷害骰（1d2）與名字輸入回顯 |
| [`tools/g6-menus-receipt.sh`](../../tools/g6-menus-receipt.sh) | 讀檔後的街上：檢視人物 → 物品 → 離開；紮營 → 魔法 → 離開、設定 → 離開；區域地圖 | 12 段範圍外差異 0；未命中只剩傷害骰與檢視標題的玩家名（重播期間隊伍欄沒重印，音譯表沒學到，實際遊玩先印隊伍欄） |
| [`tools/g6-saveload-receipt.sh`](../../tools/g6-saveload-receipt.sh) | 婉拒導覽 → 回到街上 → 紮營 → 存檔槽 A；冷開機 → 載入本作槽 A → 開始冒險 → 查驗題 → 街上 | 10 段範圍外差異 0；讀檔後隊伍與底列選單覆繪；未命中只剩作答回顯 |

兩支分段收據共用 [`tools/g6-stage.sh`](../../tools/g6-stage.sh)：每段以探針（`ADLIB.COM` 疊底、`-root-tsr-shrink`）從上一段狀態送鍵並存狀態，
再以 `pod-overlay-shot` 從同一狀態重播同樣的鍵，用該段開始前的暫存目錄快照。重播共用暫存目錄時會看到探針已寫好的存檔
（第一次跑出「overwrite ROLAND?」）。從狀態重播時，狀態存檔前就印好的字（隊伍欄、圖示編輯器標籤）不會再印，截圖上是英文，
上一段的截圖裡是中文，不算缺譯。

## 本作的操作差異

- 加入隊伍：在名單上按 Enter 只是標記（`*`），要按 `A` 才加入，再按 `E` 離開。
- 載入儲存遊戲先問 `LOAD FROM WHERE? POOLS SECRET EXIT`：POOLS 是本作的存檔，SECRET 是可轉入角色的《銀色匕首之謎》存檔。
- 港務長婉拒導覽後要按兩次 Enter 才回到街上。
- 存檔寫在 `SAVE/`：`savgamA.pty`、`CHRDATA1.sav`、`.thg`、`.efx`、`vaultA.dat`（槽 A）。

## 查驗題

由路徑 A 分段印出（`37F7:01B2`–`029B`）：「 journal, page 」「 (頁) 」「 what is word # 」「字序 」「 after the heading 」「標題」，
現在逐段覆繪成「手札，頁 (2) 請問字序 # 3 於標題之後 ABILITY SCORES」，標題維持英文。兩支收據遇到的題目：

| 收據 | 頁 | 字序 | 標題 | 答案（珍157 操作手冊附錄二） |
|---|---|---|---|---|
| 建角 | 2 | 3 | ABILITY SCORES | HAS |
| 存讀檔 | 4 | 7 | ALIGNMENT | LIVES |

收據只收這兩題，題目變了就停。題目格式、答案顯示（F9）屬 G7（#8）。

## 譯文修正

- `Pools Secret Exit` 子代理譯成「光芒之池 銀色匕首之謎 離開」，依上節改成「本作 銀色匕首之謎 離開」（前作同位置的 `Secret Curse Exit` 作「本作 青色枷 離開」）。
- 組句模板補「Save {name}? 」「overwrite {name}?」。
- 底列三條整列模板：`Items Trade Drop Exit`、`Ready Use Trade Drop Halve Join Exit`、`Save View Magic Alter Exit`。單詞 `Drop ` 會命中提示片段「要永遠放棄」放不進詞寬，
  物品清單底列的母串沒有 Exit 對不上，前作母串譯文「物品數量分半」「集合物品」「儲存遊戲」佔滿詞寬、熱鍵字母放不下。
- 術語表補 `Arrows`、`Darts`：物品清單帶加值的數量列（`Yes  50 Arrows +2`）引擎的單數回退查不到；現在作「是 箭 +2 50」（數量置後，同前作）。
- 本作紮營選單沒有 Rest。
- 紮營底列第 4 格（Save 與 View 之間）曾看到綠色「…」。現行字型與 catalog 下重跑，有整列模板與拿掉整列模板（`workplace/probe/g6-ellipsis/`）
  兩種都沒有，範圍外差異都是 0；當初的截圖已被覆蓋，來源沒查到。

## 待查

- 城鎮服務（商店、神殿、訓練所、旅店）、野外移動、非 DEMO 的戰鬥、升級、多職業尚未盤點；前作的自動探索器要隊伍位置位址，本作未驗（G8）。
