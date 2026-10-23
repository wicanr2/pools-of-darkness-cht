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
| 2026-10-23 | 手札譯名裁決：Zhentarim 桑塔林會（字頭隨珍157 桑提爾城）、spirit of fire 保留珍157「火神」、Silk 絲綢、Gothmenes 戈斯梅尼斯、Tanetal 塔內塔爾等，見 docs/re/005 | G4 |
| 2026-10-23 | 說明書其餘章節比照銀色匕首之謎做全冊 OCR 草稿與名稱對照；正文不對照錯譯，以完成中文化為先 | 使用者回答 |
| 2026-10-23 | 法術、怪物、物品名與銀色匕首之謎一致（前作術語表、前作遊戲內譯文優先）；前作沒有的才用珍157 或依前作命名類推（`text/zh157-names.tsv`，docs/re/006） | 使用者回答 |
| 2026-10-23 | 試譯文風定案；用字「裡」不寫「裏」 | 使用者確認（`text/TRANSLATION-GUIDE.md`） |
| 2026-10-23 | 攻略面板用自寫摘要：讀雜誌掃描後以自己的話整理每章路線、地點、要點與地圖編號說明，標明期數頁碼，譯名依術語表；雜誌原文不照抄、不進 repo | 使用者回答 |
| 2026-10-24 | G6 跨畫面盤點以自動探索兩輪為止（街上起點一輪、進店家神殿一輪），不再追加路線 | 使用者指示 |

## 待定

| 項目 | 何時決定 |
|---|---|
| 片頭片尾字幕：推定是畫格點陣圖（docs/re/004），辨識與覆繪方式 | G6／G9 |
| `pool4.cfg` 第 6 行 `F` 與後續 `0` 欄位語意 | 需要時 |

## dosgolem 基線

| 項目 | 值 |
|---|---|
| commit | 基底 `a6e0254`（銀色匕首之謎 `dosgolem/sosb`）；本作分支頂端 `e67971e`（規格 430、431、432 CONFORMED，`apps/pod`、`cmd/pod-play`；本作查驗題題型），推到本 repo 的 `dosgolem/pod` |
| Go／image | go1.24.13；`dosgolem-go-ebiten:1.24-r1` `sha256:8ed6bdee2bce…` |
| 冷啟動 | 0 種缺服務；mode 13h；`START.BAT` 程式鏈；片頭送空白鍵跳過，製作群後送 `p` 進建立隊伍主選單。收據 `tools/g2-startup-receipt.sh`，見 [docs/re/003](docs/re/003-dosgolem-startup.md) |
| adapter | `apps/pod/`（由 `apps/sosb/` 複製，路徑 A／B／C 換本作常數與 13h 換算）；`cmd/pod-overlay-shot` |
| play 指令 | `cmd/pod-play`（由 `cmd/sosb-play` 複製）：照 `START.BAT` 以 `snd.com` 常駐、佇列 `CONTROL.EXE`、`RootTSRShrink`（dosgolem 規格 432）；冷開機收據 `tools/play-smoke.sh` |

## 中文化進度

| 路徑 | 狀態 |
|---|---|
| 印字程序 | resident 段 `1B2E`：逐字 `01DF`、路徑 A `0354`、路徑 B `0498`、路徑 D `030B`，與幽靈騎士同函式（docs/re/004） |
| 路徑 A／B／C | 規格 001–003 CONFORMED：冷開機、DEMO、建角、存讀檔、選單收據範圍外差異 0（docs/re/007），自動探索兩輪未命中只剩數字與引號（docs/re/010） |
| catalog | ui 1,357 筆（654 筆與銀色匕首之謎、625 筆與幽靈騎士 key 相同）、ECL 4,725 筆 |
| 手札譯文 | 99 則修正譯文 `text/journal.zh-TW.tsv`（90 則有修正）；錯譯表 354 條通過 `tools/errata_check.py`（docs/re/005） |
| 術語 | `text/glossary.tsv` 949 筆：手札 194、併入銀色匕首之謎術語表 579、試譯新詞 23、說明書名稱 25（前作沒有的）、`MON1CHA.DAX` 怪物名 60、全量新詞 64、別名與複數 4；舊譯 `text/term-avoid.tsv` 38 筆；`tools/term_check.py`、`tools/zh157_names.py check` 通過，`tools/term_consistency.py` 只剩語境變化 |
| 說明書其餘章節 | 名稱 159 列對照與裁定（`text/zh157-names.tsv`）；全冊 OCR 草稿未校讀（docs/re/006） |
| 介面與劇情譯文 | `text/ecl-text.zh-TW.tsv` 4,634 則（其餘 91 則是雜訊）；`text/ui-text.zh-TW.tsv` 980 則（654 沿用前作、326 新譯；377 則雜訊）；`tools/translation_merge.py` 0 錯誤。翻譯指引 `text/TRANSLATION-GUIDE.md` |
| 字型 | 源泉圓體開發字型 `workplace/font/pod-rounded.golemfnt`（16）、`-3x`（24），2,842 字，`tools/font_coverage.py` 缺 0。DEMO 以正式譯文覆繪，開場到戰鬥 5 張截圖範圍外差異 0（`tools/g5-text-receipt.sh`）；DEMO 戰鬥畫面的隊員名、物品名與加值、施法訊息已覆繪（docs/re/007） |
| 查驗題 | 規格 006 CONFORMED：GAME.EXE 題目表 21 筆 54 題（`tools/quiz_table.py` → `text/quiz-answers.tsv`），改寫第 2、4、6 列並顯示答案（F9 切換），收據 `tools/g7-quiz-receipt.sh`（docs/re/008） |
| 手札面板 | 規格 007 READY：叛徒門實機觸發第 57 則，面板自動開出、翻閱模式可開（`tools/g7-journal-receipt.sh`，docs/re/009）；酒館另觸發第 9 則 |
| 攻略面板 | 規格 008 READY：《軟體世界》第 43–45 期自寫摘要 78 段（`text/walkthrough.zh-TW.tsv`，docs/re/011）；主選單開 F6 三張截圖無缺字 |
| 跨畫面盤點 | G6 完成：自動探索兩輪（街上、店家門口）印字 1,857 次（第二輪計到第 400 鍵的快照），未命中只剩數字與引號；商店清單、神殿、訓練、戰鬥、野外未覆蓋（docs/re/010） |
| 片頭片尾字幕 | 推定是畫格點陣圖，待 G9 |

## 原始輸入

| 檔案 | 狀態 |
|---|---|
| `pools-of-darkness.zip` | 已安裝目錄，v1.10；`START.BAT` → `STARTUP.EXE`（設定 `pool4.cfg`）→ `CONTROL.EXE`（`VACE.EXE` 片頭、`GAME.EXE`、片尾）；`GAME.OVR` 魔數 `FBOV`；查驗題問手札頁碼、標題後第 N 字；`SAVE/` 是 1995 年玩家資料。見 [docs/re/001](docs/re/001-input-baseline.md) |
| `珍157-黑暗之池.rar` | 107 張 JPG：操作手冊 63 頁、冒險者手冊 110 頁，無缺頁；旅行手札 99 則（第 1–100 則，無第 13 則）在冒險者手冊 p.55–106；操作手冊附錄二是查驗題答案表。見 [docs/re/002](docs/re/002-reference-sources.md) |
| `walkthrough/` | 《軟體世界》第 43–45 期「完全攻略」三回，第 1–24 章，作者 Steeler；未轉錄。見 [docs/re/002](docs/re/002-reference-sources.md) |
| 英文手冊與手札 | mocagh.org 的 journal（規則＋手札 99 則，無第 13 則）、manual（Rulebook）、refcard、hintbook，OCR 品質中等。見 [docs/re/002](docs/re/002-reference-sources.md) |

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
