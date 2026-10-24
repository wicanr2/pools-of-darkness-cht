# 004 輸出路徑盤點

日期：2026-10-23。Issue：[#4](https://github.com/wicanr2/pools-of-darkness-cht/issues/4)。基線見 [003](003-dosgolem-startup.md)。
位址為 dosgolem 執行期 `段:位移`，程式鏈照 `START.BAT`（常駐驅動＋`-root-tsr-shrink`）；鏈不同時段值會變，adapter 一律以簽章搜尋。
對照欄引自幽靈騎士 `/home/anr2/cht/golden_box/death_knight_of_krynn/docs/re/003-output-paths.md` 與銀色匕首之謎 `docs/re/004-output-paths.md`。
adapter 是 dosgolem `apps/pod/`（由 `apps/sosb` 複製，`dosgolem/pod` `ce5974b`），覆繪收據用 `cmd/pod-overlay-shot`。

## GAME.EXE 是 EXEPACK

`GAME.EXE` 進入點在映像尾端（`0F3C:0012`），該處是 EXEPACK 標頭（`RB` 簽名，real CS:IP `0000:0002`、dest_len `11C5h` 段），
檔內有「Packed file is corrupt」，MZ 重定位數 0（已證實，bytes）。檔案位移因此不等於執行期位移，連續相同 byte 也被壓掉。
解開後的映像由 dosgolem 執行原版 stub 取得（[`tools/g3-game-image.sh`](../../tools/g3-game-image.sh)：主選單時傾印 PSP `169F` 起，
映像 72,784 bytes，SHA-256 `817b4f5c…`，段值已重定位 +16AFh）；檔頭後第一段 bytes 與檔案相同，只差重定位的段值（`0825`→`1ED4`）。

## resident 印字程序

以幽靈騎士與銀色匕首之謎的簽章在 `GAME.EXE` 檔案中找到同一組三個程序（`tools/re/sigscan.py`），再以 dosgolem `-find` 在主選單時的記憶體
取得線性位址，並以記憶體中的遠呼叫 `9A <位移> <段>` 反推段值 `1B2E`（[`tools/re/memcalls.py`](../../tools/re/memcalls.py)；路徑 A 71 處、路徑 B 21 處、逐字 4 處）。

| 程序 | `GAME.EXE` 檔案 | 執行期 | 幽靈騎士 | 等級 |
|---|---|---|---|---|
| 逐字程序 | 0x4A20 | `1B2E:01DF`（線性 `1B4BF`） | `05F4:01F5` | 已證實：DEMO 中 298 次（`-call-args`） |
| 路徑 D（逐字之後的 near 函式） | — | `1B2E:030B` | `05F4:0361` | 已證實呼叫：DEMO 27 次，由 overlay `x:1691` 呼叫，6 個 word；參數低位元組 40、0、24、0 像整個文字格矩形，推定為清除（假說） |
| 路徑 A Pascal 字串 | 0x4B95 | `1B2E:0354`（線性 `1B634`） | `05F4:036A` | 已證實：DEMO 中 485 次 |
| 路徑 B 文字視窗 | 0x4CD9 | `1B2E:0498`（線性 `1B778`） | `05F4:04AB` | 已證實：DEMO 開場 13 次 |

- `tools/re/fn_diff.py` 與幽靈騎士 `START.EXE` 逐 byte 對齊：差異只在遠呼叫的位移與段、資料段位移與少數近呼叫。
  逐字程序的 `DS:B42E／B430／B432` 在本作是 `9D86／9D88／9D8A`；路徑 B 游標 `DS:B43A／B43B` 在本作每一處都是 `DS:9D92／9D93`（靜態對齊，強推論）。
  路徑 A 有一處幽靈騎士讀 `[43E9]`、本作傳常數 1（`b0 01`）。路徑 B 自 +0x222 起的尾段與幽靈騎士不同。
- 路徑 A 的前 23 bytes 在記憶體另有兩處：線性 `1C7AC`（第 28 byte 為 `8A`）與 overlay 內 `50224`（第 28 byte 為 `FF`）。
  以第 28–30 byte `9A 51 00` 區分後唯一；adapter 簽章含遠呼叫位移 `89 0A`、`51 00`（幽靈騎士 `DE 0A`、`58 00`）。
- 資料段：主選單時 `DS=2096`。

### 路徑 A（已證實）

參數 6 個 word：字串位移、段、底色、前景、列、欄（例 `8 30699 0 15 4 17`），後四欄取低位元組，與幽靈騎士相同。
DEMO 收到的字串是混合大小寫：`Loading...Please Wait`、`Name`、`AC  HP`、角色名、AC 與 HP 數值、`14,5 W 10:00` 座標時間列，
戰鬥中有 `Move/Attack, Move Left = 12 `、`Spell:Delayed Blast Fireball`、怪物名等組句。

### 路徑 B（已證實）

參數 9 個 word：字串遠指標（DS 緩衝 `2096:B16E`），其後依序 1、0、14、22、38、17、1（含旗標與視窗矩形，各欄語意待 G6 對照）。DEMO 開場收到整句大寫 ECL 字串
（第一則以 `'WELCOME, MY FRIENDS` 開頭，共 12 則非空字串），由程序內斷行，與銀色匕首之謎、克萊恩英豪的介面相同。
內部經 near 函式 `1B2E:03E3`、`045E` 呼叫逐字程序（逐字呼叫端 `1B2E:042F`）。
`apps/pod` 的路徑 B 已在 adapter 綁定（簽章唯一、字型檢查通過）：給空的 ECL 譯文表跑 DEMO，第一句在 #336,364,657 被攔下，
解出列 17、欄 1、前景 14、底色 0，因為沒有譯文而標「未命中」（`workplace/probe/g3-pathb/run.json`）。
G6 在遊戲內第一個對白（菲蘭城碼頭的港務長，`tools/g6-create-receipt.sh` 的 answer 段，#2,417,437,720）攔到同樣的 7 個參數 1、0、14、22、38、17、1
（`workplace/probe/g6-pathb/run.log`），覆繪正確。兩個場景都是底部文字視窗；依 adapter 解出的位置與顏色，推定依序是旗標、底色、前景、下、右、上、左（強推論）。

## 路徑 C：底列水平選單（已證實 bytes）

主選單時記憶體中呼叫 `9A DF 01 2E 1B` 的四處（overlay 段 `44B7`），返回位址後 bytes 分辨四種呼叫：

| 種類 | 呼叫位置（線性） | 返回位址後 | 幽靈騎士 | 銀色匕首之謎 |
|---|---|---|---|---|
| selected | `44EB9` | `E9 94 00` | `E9 96 00` | `E9 95 00` |
| hotkey | `44F13` | `EB 3B` | `EB 3C` | `EB 3C` |
| normal | `44F50` | `FE 46 FF` | `FE 46 FF` | `8A 46 FF 3A 46 FE` |
| pad | `463DC` | `C4 7E FC 26 8B 45 2A` | 同左 | `9A 32 00` |

DEMO 中 `44B7:03A8`、`03E5`、`034E` 逐字送出 `Play Demo` 等選單字；主選單底列 `SELECT` 經此覆繪（下節）。

## mode 13h

- 顯存是 A000 線性、一列 320 bytes；adapter 把顯存寫入位移換成文字格 `row=(off/320)/8`、`col=(off%320)/8`
  （與拯救地球 `apps/buckrogers/ecl_text.go` 相同；前作 planar 是一列 40 bytes）。
- 主選單文字像素的色號就是路徑 A 的前景、底色參數：綠字 10、黑底 0、反白列底 15；色盤 0–15 是 EGA 預設色（已證實，`menu-a.raw`／`.pal`）。
  路徑 A「最多的顏色不是 bg 就反色」的判斷因此照舊成立。
- 通用層 `Indexed`、`FramePalette`、顯存寫入觀察都已支援 13h，不需改。

## 片頭片尾字幕

`VACE.EXE` 播 `INTRO`／`FINAL` 的 `.ACF`／`.BMA`。`INTRO.ACF`、`FINAL.ACF` 各只有一段 8 字元可列印字串，`INTRO.BMA`、`FINAL.BMA`、`VACE.EXE`
都找不到字幕的明文（搜尋 `summoned`、`Phlan`、`heroes` 等）。推定字幕是動畫畫格裡的點陣圖（強推論），不經任何印字程序；
處理方式是畫面辨識後覆繪，歸 G6／G9，不在本階段。

後續來源盤點入口：[movie-source-receipt.sh](../../tools/movie-source-receipt.sh)。沿G2常駐驅動與CONTROL原始入口，
不送跳過鍵，每2000萬指令匯出色號及色盤至12億指令；屬既有開場的來源採樣，不增加G6探索。
這項收據本身不代表字幕辨識、翻譯、HD或不露原版已完成，時序仍以dosgolem實際畫格為準。

## 文字 catalog

| catalog | 筆數 | 來源 | 備註 |
|---|---|---|---|
| `text/ui-text-events.tsv` | 1,357 | `GAME.EXE` 解開後的映像＋`GAME.OVR`（[`tools/ui_text_catalog.py`](../../tools/ui_text_catalog.py)） | key 與前作同制；與銀色匕首之謎譯文 654 個 key 相同、與幽靈騎士 625 個。DEMO 路徑 A 的 61 則字串中，8 則靜態字串全部命中，其餘是數字、座標、名字與組句 |
| `text/ecl-text-events.tsv` | 4,725 | `ECL1.DAX`（[`tools/ecl_text_catalog.py`](../../tools/ecl_text_catalog.py)，移植自銀色匕首之謎） | 215,509 字元；DEMO 路徑 B 的 12 則非空字串全部命中 |

兩份 catalog 只有 key、長度、SHA-256 與來源位置，原文在 `workplace/text-src/`（不進 Git）。

## 覆繪收據（主選單）

[`tools/g3-overlay-receipt.sh`](../../tools/g3-overlay-receipt.sh)：從 G2 收據的 `credits-300m.state` 送 `p`，#399,000,000 輸出原版與覆繪。
譯文與字型是佔位（[`tools/trial_text.py`](../../tools/trial_text.py) 取前作 key 相同的 824 則；字型複製銀色匕首之謎的開發字型），不是本作成品。

| 項目 | 結果 |
|---|---|
| 原版色號 | `fefa0514…`，與 G2 收據相同 |
| 疊字 | 6 筆：路徑 A 主選單 5 行（含反白行）、路徑 C 底列 `SELECT` |
| 範圍外差異 | 0 |
| 覆繪圖 | `overlay_sha256` `ce440573568588a52fc473ba4c7ae1ab8f8647ae53d4d3d8aac738a2ce55badd` |

## 第 4 節待驗假說的結論

| 假說 | 結論 | 等級 |
|---|---|---|
| 1. 與前作同屬一個輸出路徑家族 | 同家族；與幽靈騎士最接近（路徑 A 參數、路徑 C 的 normal／pad 呼叫端、`FBOV`、版本同為 1.10）。adapter 以 `apps/sosb` 為底、常數取本作 | 已證實（bytes 與呼叫） |
| 2. VGA 為 mode 13h，前作 0Dh 的假設要重驗 | 13h 已證實（003）；文字格換算改每列 320 bytes；色號空間與路徑 A 參數一致（主選單） | 已證實（主選單）；其他畫面待 G6 |
| 3. 片頭片尾動畫 | 字幕不是文字輸出 | 強推論 |

## 未驗證的沿用常數

下列常數仍是銀色匕首之謎值，`apps/pod` 原始碼註明「銀色匕首之謎值」，對應階段前不啟用：

| 常數 | 用途 | 重驗時機 |
|---|---|---|
| `demo.go` 主程式簽章與 `8A63`、`59B4`、`87CC`、`67E9` | demo 控制 | #13 |
| `cheat.go partyHead` 與紀錄位移 | 作弊 | #11 |
| `automap.go geoPtrOff`、`partyXOff` | 地圖疊層 | #9 |
| `players.go recordSexOff 0x11F` | 玩家名音譯 | #7 |

## 未知

- 路徑 D 的用途；清除與捲動的掛勾（#4 的完成條件含這兩項，本階段只到路徑 D 的假說，交給 G6）。
- 印字是否經 `VGA.DRV` 寫像素：已知逐字程序在 `GAME.EXE` resident 段、顯存寫入觀察得到，像素寫入在哪一支沒追。G6 做清除偵測時要知道。
- 路徑 B 九個參數的各欄語意。
- 手冊查驗題、手札引用（`This is recorded as journal entry`）的實際輸出事件：要建角進遊戲才看得到（G6、G7）。
- 戰鬥、商店、神殿、訓練等畫面的呼叫端盤點。
- `VACE.EXE` 字幕在畫格中的位置與辨識方式。
