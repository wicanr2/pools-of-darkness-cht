# 規格索引

狀態：DRAFT → 證據審查 → READY → 實作 → 同狀態驗證 → CONFORMED（`AGENTS.md` 第 7 節）。
規格 001–017 的契約沿用克萊恩英豪（`/home/anr2/cht/golden_box/champions_of_krynn/docs/spec/`），018–022 沿用銀色匕首之謎
（`/home/anr2/cht/golden_box/secret-of-the-sliver-blader/docs/spec/`）。adapter 是由 `apps/sosb` 複製的 `apps/pod`，play 指令是由
`cmd/sosb-play` 複製的 `cmd/pod-play`。本表的狀態只算本作已有的證據，前作的 CONFORMED 不代表本作已驗。
本作差異寫在「差異」欄，細節在對應的 `docs/re/`。通用層的規格在 dosgolem（430、431、432，皆 CONFORMED，docs/re/003）。

| 規格 | 本作狀態 | Issue | 差異 | 證據 |
|---|---|---|---|---|
| 001 路徑 A：Pascal 字串覆繪 | READY | #4、#7 | resident 段 `1B2E:0354`；EXEPACK，位址以解開後映像為準；mode 13h 每列 320 bytes | docs/re/004；`tools/g3-overlay-receipt.sh`、`g5-text-receipt.sh`、`play-smoke.sh`、`g6-create-receipt.sh` |
| 002 路徑 B：ECL 文字視窗覆繪 | READY | #4、#7 | `1B2E:0498`，9 個 word；DEMO 與遊戲內碼頭對白參數相同（1、0、14、22、38、17、1，推定為旗標、底色、前景、下、右、上、左）；游標 `DS:9D92／9D93`（強推論） | docs/re/004；`tools/g5-text-receipt.sh`、`g6-create-receipt.sh` |
| 003 路徑 C：底列水平選單 | READY | #4、#7 | 逐字程序 `1B2E:01DF`；四種呼叫端（selected `E9 94 00`、hotkey `EB 3B`、normal `FE 46 FF`、pad `C4 7E FC…`）在主選單與建角畫面相同；熱鍵字母接在譯文後 | docs/re/004；`tools/g6-create-receipt.sh` |
| 004 組句訊息 | DRAFT | #7 | 模板沿用前作 188 條，另加本作建角的存檔、覆蓋確認；DEMO 戰鬥的物品名＋加值（釘頭錘 +3）、施法訊息已命中 | `tools/play-smoke.sh`（DEMO 10,500 畫格） |
| 005 中文句中的半形英數 | DRAFT | #7 | 數字照原版保留；未單獨驗證 | — |
| 006 手冊查驗題 | DRAFT | #8 | 題型「(journal, page N) what is word # M after the heading …」，由路徑 A 分段印出（`37F7:01B2`–`029B`）；題目表靜態找不到；答案對照珍157 操作手冊附錄二；正常玩家路徑作答一題（第 2 頁 ABILITY SCORES 後第 3 字 HAS）後進入遊戲 | docs/re/001、005；`tools/g6-create-receipt.sh` |
| 007 手札面板 | DRAFT | #8 | 99 則珍157 修正譯文（`text/journal.zh-TW.tsv`）；引用與彈出未在本作觀測 | docs/re/005 |
| 008 攻略面板 | 未開始 | #9 | 來源《軟體世界》第 43–45 期 | docs/re/002 |
| 009 地圖疊層 | 未開始 | #9 | — | — |
| 010 HD 主題 | 未開始 | #10 | — | — |
| 011 切換列 | 未驗 | #10 | 程式由前作沿用 | — |
| 012 發行包 | 未開始 | #12 | — | — |
| 013 玩家名音譯 | DRAFT | #7 | 音譯資料沿用前作；隊伍名單的玩家名已音譯（ROLAND → 羅倫德）；建角畫面的名字與性別欄位址未在本作驗證 | `tools/g6-create-receipt.sh` |
| 014 配樂 | 未開始 | #11 | — | — |
| 015 多語系 | 未開始 | #11 | — | — |
| 016 推廣影片 | 未開始 | #11 | — | — |
| 017 作弊：我方無敵 | 停用 | #11 | 位址是前作值，`CheatVerified` 為假 | — |
| 018 開場 demo 控制 | 未開始 | #13 | `GAME.EXE` 有 `Play Demo` | docs/re/001 |
| 019 CPU 模擬速度 | 未驗 | #14 | 程式由前作沿用；本作 CRT `Delay` 寫法未驗 | — |
| 020 切換列字級 | 未驗 | #15 | 程式由前作沿用 | — |
| 021 不露原版 | 未開始 | #16 | — | — |
| 022 手札地圖頁 | 未開始 | #8 | 地圖則 5、23、45、98，第 36 則附地圖 | docs/re/005 |

升到 CONFORMED 的條件：本作正常玩家路徑（建角、存讀檔）的同狀態收據覆蓋該規格的行為。
