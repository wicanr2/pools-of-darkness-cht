# 001 原始輸入基線

日期：2026-10-23。Issue：[#2](https://github.com/wicanr2/pools-of-darkness-cht/issues/2)。
工具：`tools/inventory.py` 解壓與雜湊，`tools/exe_probe.py` 讀 MZ 標頭與字串，`tools/inventory_tsv.py` 產生逐檔雜湊表。
都在 `sosb-tools:bookworm-r1`（`tools/docker/tools.Dockerfile`）內以 `--network none`、原始輸入唯讀掛載執行。
遊戲檔展開在 `workplace/input/dos/pod/`，盤點結果 `workplace/input/inventory.json`（不進 Git）。

## 壓縮檔

| 檔案 | SHA-256 | 內容 |
|---|---|---|
| `pools-of-darkness.zip` | `a67022ffd89128cc1e35b6fefa79e456fd94ee03ea5ec26c0082f4bebc770483` | 頂層目錄 `pools-of-darkness/`：遊戲檔 47 個（時間戳 1995-08-04）、`SAVE/` 78 個（1995-08-06 至 08-13）；deflate，Python `zipfile` 可直接解 |

這是已安裝的遊戲目錄，沒有安裝程式。珍157 附的磁片標籤是 DISK 1、DISK 2（1.2M），操作手冊 p.3–4 說明以安裝程式載入硬碟；
zip 與那兩片磁片的對應無法從本輸入判定。

逐檔大小與 SHA-256（遊戲檔、存檔、珍157、攻略掃描共 265 列）見 [001-input-files.tsv](001-input-files.tsv)。

## 版本與入口

| 項目 | 值 | 證據 | 等級 |
|---|---|---|---|
| 版本 | 1.10 | `GAME.EXE` 0xB890 字串 `POOLS OF DARKNESS 1.10` | 已證實 |
| 啟動流程 | `START.BAT`：`startup` → `if exist snd.com snd` → `control %1 %2 %3 %4` → `if exist snd.com snd U` | `START.BAT` 全文 78 bytes | 已證實 |
| 設定 | `STARTUP.EXE` 問音效、顯示卡、替代輸入、存檔路徑、銀色匕首之謎存檔路徑，設定檔名 `pool4.cfg` | `STARTUP.EXE` 字串 0x9FD–0xC23；`CONTROL.EXE`、`GAME.EXE` 都引用 `pool4.cfg` | 已證實（字串）／強推論（由 STARTUP 寫入） |
| 音效驅動 | `SBLASTER.COM`、`ADLIB.COM`、`TANEXTX.COM`、`TANSLTL.COM`、`IBMSND.COM`（THE Audio Solution 1991，v1.0）；選單字串讀到 [1] Sound Blaster、[2] Adlib、[3]／[4] Tandy 1000、[6] Silent，[5] 未讀到 | 各驅動字串；`STARTUP.EXE` 選單字串 | 已證實 |
| `snd.com` | zip 內沒有；`START.BAT` 存在才載入 | 推測 `STARTUP.EXE` 把選定的音效驅動複製成 `snd.com` | 假說 |
| 顯示 | 選單 `GRAPHICS ADAPTER TYPE: [1] EGA [2] Tandy [3] VGA`；驅動 `EGA.DRV`、`TGA.DRV`、`VGA.DRV` | `STARTUP.EXE` 字串 | 已證實（選項）／未知（VGA 的實際 mode） |
| 主程式 | `CONTROL.EXE` 讀 `pool4.cfg`，依序呼叫 `vace.exe … INTRO 1`、`game.exe`、`vace.exe … FINAL 1` | `CONTROL.EXE` 字串 | 強推論（呼叫順序依字串排列） |
| 動畫 | `VACE.EXE`：「AnimationControlEngine System MultiDriver Viewer」，Kerry L. Bonin，Revision 1.1（1991-08-16），Turbo C++ 1990；參數 `VideoMode AceFile VideoFlag SoundFlag`，VideoMode 2=EGA、3=TGA、4=VGA；播放 `INTRO.ACF／.BMA`、`FINAL.ACF／.BMA` | `VACE.EXE` 字串 | 已證實（字串） |
| overlay | `GAME.OVR` 開頭 `FBOV`，其後 u32 為檔長減 8（264078） | bytes | 已證實 |
| 查驗題 | 問 Adventurer's Journal 某頁某標題之後第 N 個字 | `GAME.OVR` 0x35B4 ` journal, page `、0x35D2 ` what is word # `；操作手冊 p.19 範例與附錄二（docs/re/002） | 已證實（字串）／強推論（題型） |
| 手札引用 | 劇情以「This is recorded as journal entry」指向手札編號 | `GAME.OVR` 0x261D | 已證實（字串） |
| demo | 主選單有 `Play Demo` | `GAME.EXE` 0xBA6A | 已證實（字串） |

`GAME.OVR` 的 `FBOV` 與克萊恩英豪、幽靈騎士相同，與銀色匕首之謎的 `TPOV` 不同。`GAME.OVR` 0x18F1E 另有字串 `krynn.cfg`。
幽靈騎士 `docs/re/003-output-paths.md` 的「FBOV 容器」一節記錄了同一檔頭（`FBOV` 加檔長減 8），本作 G3 從那裡開始驗。

## `SAVE/`

78 個檔，檔名分 A–E 五組。每個 `.SAV` 510 bytes，第 0 byte 是名字長度（Pascal 字串）。

| 組 | 角色（`.SAV` 名字欄） | 其他 |
|---|---|---|
| A | RANGER、THIEF、CLERIC、PRIEST、MAGE、MAGIC USER | 沒有 `.THG`；`VAULTA.DAT` 5871 bytes |
| B | 同 A，另有第 7 人 NACACIA | 各角色 `.THG`；`VAULTB.DAT` 5871 bytes |
| C | 同 A，另有第 7 人 SHAL | 各角色 `.THG`；`VAULTC.DAT` 12 bytes，全為 0 |
| D | 同 A，另有第 7 人 NACACIA | 各角色 `.THG`；`VAULTD.DAT` 12 bytes，全為 0 |
| E | 同 A | 各角色 `.THG`；`VAULTE.DAT` 12 bytes，全為 0 |

- `.THG` 內有物品名稱（例 `50 Arrows +2`、`Long Sword +3`、`Plate Mail`），推定是攜帶物品表；`VAULTx.DAT` 推定是金庫。
- 沒有 `SAVGAMx.DAT`。這批檔案能否從「Load Saved Game」載入，待 G2 實測。
- 判定：1995 年玩家留下的角色與存檔資料，不是 SSI 預建隊伍。依據是時間戳晚於遊戲檔、名字是職業名、帶第 7 人 NPC 與冒險後的物品（強推論）。
- 權利分類：第三方玩家資料，與原版素材同樣不進 Git、不進發行包；只作本機測試夾具。
