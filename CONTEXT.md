# 目前真相

更新：2026-10-23（Asia/Taipei）。規則見 [AGENTS.md](AGENTS.md)，歷程見 [WORKLOG.md](WORKLOG.md)。

## 儲存庫

| 項目 | 值 | 回讀方式 |
|---|---|---|
| 網址 | <https://github.com/wicanr2/pools-of-darkness-cht> | `gh repo view --json url` |
| owner | `wicanr2` | `gh auth status`、`gh repo view --json owner` |
| 可見性 | PRIVATE | `gh repo view --json visibility` |
| Issues | 啟用；初始 G0–G11 為 #1–#12，規格 018–021 為 #13–#16 | `gh issue list` |
| git 身分 | repo-local `wicanr2` <wicanr2@gmail.com> | `git config user.email` |
| 授權 | RRSAL-1.0（[LICENSE](LICENSE)） | |

## 有效決策

| 日期 | 決策 | 來源 |
|---|---|---|
| 2026-10-23 | 參考銀色匕首之謎，用 dosgolem 執行原版並覆繪繁中、做 HD；步驟完全比照銀色匕首之謎，定案表見 AGENTS.md 第 1 節 | 使用者指示 |
| 2026-10-23 | 中文題名「黑暗之池」 | 使用者指示 |
| 2026-10-23 | 錯譯修正對象是珍157 軟體世界說明書；英文手冊與手札授權上網取得 | 使用者回答 |
| 2026-10-23 | repo `pools-of-darkness-cht`；遊戲短名 `pod`，adapter 在 dosgolem `apps/pod/`，分支推到本 repo 的 `dosgolem/pod` | 使用者回答 |
| 2026-10-23 | 本輪做完 G0：文件、Git、private repo、初始 Issue | 使用者回答 |
| 2026-10-23 | 顯示模式 VGA 優先；dosgolem 跑不起來時回報，不自動改用 EGA | 使用者回答 |
| 2026-10-23 | 攻略來源：使用者提供《軟體世界》第 43–45 期掃描 31 張，放 `walkthrough/`（不進 Git） | 使用者提供 |
| 2026-10-23 | dosgolem 基底選銀色匕首之謎 `dosgolem/sosb`（`a6e0254`），理由見 docs/re/003 | G2 |
| 2026-10-23 | 入口照 `START.BAT`：`STARTUP.EXE` 設定 → 常駐 `snd.com`（`-root-tsr-shrink`）→ 監督佇列接 `CONTROL.EXE`；不常駐時 DEMO 會當掉 | G2（docs/re/003） |
| 2026-10-23 | 以 VGA（mode 13h，256 色）為準；dosgolem 可跑，不需退回 EGA | G2（docs/re/003） |
| 2026-10-23 | zip 的 `SAVE/` 遊戲內讀不到（缺 `savgam?.pty`、`*.pc`），測試存檔自建 | G2（docs/re/003） |

## 待定

| 項目 | 何時決定 |
|---|---|
| 片頭片尾字幕（`VACE.EXE` 畫的）的輸出攔截點 | G3 |
| `apps/pod` 的 play 指令要照 `START.BAT` 常駐驅動並開 `RootTSRShrink`；oracle 目前沒有能載 `.COM` 又吃 `Options` 的入口（dosgolem 規格 431 §5） | G3，先補通用層規格 |
| `pool4.cfg` 第 6 行 `F` 與後續 `0` 欄位語意 | 需要時 |

## dosgolem 基線

| 項目 | 值 |
|---|---|
| commit | 基底 `a6e0254`（銀色匕首之謎 `dosgolem/sosb`）；本作分支頂端 `2e47f92`（規格 430、431 CONFORMED），推到本 repo 的 `dosgolem/pod` |
| Go／image | go1.24.13；`dosgolem-go-ebiten:1.24-r1` `sha256:8ed6bdee2bce…` |
| 冷啟動 | 0 種缺服務；mode 13h；`START.BAT` 程式鏈；片頭送空白鍵跳過，製作群後送 `p` 進建立隊伍主選單。收據 `tools/g2-startup-receipt.sh`，見 [docs/re/003](docs/re/003-dosgolem-startup.md) |
| adapter | `apps/pod/` 尚未建立；G3 掛輸出攔截時由 `apps/sosb/` 複製 |

## 中文化進度

尚未開工。

## 原始輸入

| 檔案 | 狀態 |
|---|---|
| `pools-of-darkness.zip` | 已安裝目錄，v1.10；`START.BAT` → `STARTUP.EXE`（設定 `pool4.cfg`）→ `CONTROL.EXE`（`VACE.EXE` 片頭、`GAME.EXE`、片尾）；`GAME.OVR` 魔數 `FBOV`；查驗題問手札頁碼、標題後第 N 字；`SAVE/` 是 1995 年玩家資料。見 [docs/re/001](docs/re/001-input-baseline.md) |
| `珍157-黑暗之池.rar` | 107 張 JPG：操作手冊 63 頁、冒險者手冊 110 頁，無缺頁；旅行手札第 1–100 則在冒險者手冊 p.55–106；操作手冊附錄二是查驗題答案表。見 [docs/re/002](docs/re/002-reference-sources.md) |
| `walkthrough/` | 《軟體世界》第 43–45 期「完全攻略」三回，第 1–24 章，作者 Steeler；未轉錄。見 [docs/re/002](docs/re/002-reference-sources.md) |
| 英文手冊與手札 | mocagh.org 的 journal（規則＋手札 100 則）、manual（Rulebook）、refcard、hintbook，OCR 品質中等。見 [docs/re/002](docs/re/002-reference-sources.md) |

## Glossary

- 黑暗之池：本作中文題名，與珍157 相同，使用者指定。
- 珍157：軟體世界珍藏版第 157 號，本作的官方繁中說明書，錯譯修正的對象。
- 操作手冊、冒險者手冊：珍157 的兩冊。操作手冊對應英文 Rulebook 加查驗題答案表；冒險者手冊對應英文 Adventurer's Journal，含旅行手札。
- 密碼表：珍157 操作手冊附錄二，查驗題的頁數、字數、英文標題與答案對照。_Avoid_：答案表
- 錯譯表：`text/errata.tsv`，記錄珍157 譯文與修正譯文的對照與證據。
- 手札：說明書中 Adventurer's Journal 的段落，遊戲內以編號引用。_Avoid_：日誌、日記
- pod：本作遊戲短名，用於 dosgolem `apps/pod/` 與程式識別。
- 銀色匕首之謎：第一參考專案，本機 `/home/anr2/cht/golden_box/secret-of-the-sliver-blader/`，短名 `sosb`。
- 克萊恩英豪：銀色匕首之謎的參考專案，規格 001–017 的原始版本，本機 `/home/anr2/cht/golden_box/champions_of_krynn/`，短名 `cok`。
- 覆繪：在原版已輸出的畫面上清除原文墨跡、重畫中文。_Avoid_：替換、打字串補丁
- 收據：dosgolem 以固定輸入重播產生、可重現的執行證據。_Avoid_：截圖證明
- 閃現：原版字或原版圖先原樣呈現、8 格內才被譯文或 HD 換掉（銀色匕首之謎規格 021 §2）。_Avoid_：閃爍（只用於使用者描述的現象）
- HD 批次：一組交給 Codex 重繪的原版圖塊，含請求與輸出兩個目錄。

### 待定譯名

尚無。

### Flagged ambiguities

- 本機目錄名 `pool-of-darkness` 是單數，官方名與 repo `pools-of-darkness-cht` 是複數，指同一專案，不改名。
