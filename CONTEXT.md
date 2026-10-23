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

## 待定

| 項目 | 何時決定 |
|---|---|
| dosgolem 基底（候選見 AGENTS.md 第 5 節） | G2 |
| 實際顯示模式與驅動選擇方式 | G2 |
| 版本與發行形式、查驗題形式、`SAVE/` 的性質與權利分類 | G1 |

## dosgolem 基線

尚未選定。

## 中文化進度

尚未開工。

## 原始輸入

| 檔案 | 狀態 |
|---|---|
| `pools-of-darkness.zip` | 已展開的遊戲目錄，遊戲檔 47 個、`SAVE/` 78 個；只讀成員清單與 SHA-256，未解壓（AGENTS.md 第 3 節） |
| `珍157-黑暗之池.rar` | 107 張 JPG 掃描，無 PDF；只讀成員清單與 SHA-256，未解壓 |
| `walkthrough/` | 《軟體世界》第 43–45 期 31 張 `.jpeg`，未轉錄 |
| 英文手冊與手札 | 尚未取得 |

## Glossary

- 黑暗之池：本作中文題名，與珍157 相同，使用者指定。
- 珍157：軟體世界珍藏版第 157 號，本作的官方繁中說明書，錯譯修正的對象。
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
