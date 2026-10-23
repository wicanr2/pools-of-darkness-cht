# 011 遊戲內攻略

日期：2026-10-24。Issue：[#9](https://github.com/wicanr2/pools-of-darkness-cht/issues/9)。契約沿用克萊恩英豪規格 008（攻略面板）。

## 來源與做法

來源是《軟體世界》第 43–45 期〈黑暗之池 完全攻略〉（作者 Steeler，雜誌 p.60–70、64–75、66–73，共 31 頁；頁序見 [002](002-reference-sources.md)）。
使用者決定面板放自寫摘要，雜誌原文不照抄（CONTEXT 決策 2026-10-23）。讀掃描後依章整理路線、要點與地圖編號說明，
寫在 `reference/walkthrough-zh/softworld-43.md`–`45.md`（private，轉公開前移除），每章標出處頁碼。譯名在寫摘要時查術語表，
術語表沒有的地名、物品名保留英文。

| 期 | 章 | 章段／地圖段 | 摘要 |
|---|---|---|---|
| 43 | 第 1–11 章（重回菲蘭城到 Manshoon's Tower） | 11／17 | `softworld-43.md` |
| 44 | 第 11 章（續）到第 19 章 Tower of Marcus | 9／17 | `softworld-44.md` |
| 45 | 第 20–24 章（Moander Minis 到 Palace of Gothmenes）、結語 | 6／18 | `softworld-45.md` |

三期由子代理讀圖起草，主迴圈抽驗：

| 期 | 抽驗範圍 | 結果 |
|---|---|---|
| 43 | 第 7 章 | 只看結構（段落、要點、地圖 18 條），沒有對圖 |
| 44 | 第 12 章地圖（p.65） | 1–6、10、12–20 號相符；子代理標「編號模糊」的兩條夾在 6 與 10 之間，可能是 7、8 或 8、9，掃描上看不出 |
| 45 | 第 21 章地圖（p.70） | 7、8、10、11、14、15 號相符 |

## 照抄檢查

`tools/walkthrough_overlap.py` 拿 OCR 草稿（`tools/ocr-batch.sh walkthrough workplace/walkthrough/ocr`）每列切 10 字窗口，比對摘要：

| 期 | OCR 列 | 整列照抄 | 10 字窗口 | 出現 |
|---|---:|---:|---:|---:|
| 43 | 447 | 0 | 2,164 | 35（1.6%） |
| 44 | 448 | 0 | 1,818 | 39（2.1%） |
| 45 | 239 | 0 | 958 | 17（1.8%） |

出現的窗口多是專有名詞（例 Talisman of Bane、Crystal Ring）與短的地圖標籤。OCR 有錯字，數字是下限估計。

## 產生

`tools/walkthrough_build.py` 把摘要轉成 `text/walkthrough.zh-TW.tsv`：78 段（章 26 段，含第 11 章續與結語；地圖 52 段）。
章與該章地圖各成一段，要點每條一段，出處寫成「整理自軟體世界 N 期 p.X–Y」顯示在面板頁尾。
標題列共 40 格，標題超過時把括號裡的說明移到正文第一段，移完仍超過就報錯。

## 驗證

- `TestGuideTableFits`（`DOSGOLEM_DKK_TEXT_DIR` 指到 `text/`）：78 段都載得進、每段至少一頁。
- 字型：`tools/font-dev.sh` 收進攻略表後 2,842 字，`tools/font_coverage.py` 缺 0。
- [`tools/g8-guide-receipt.sh`](../../tools/g8-guide-receipt.sh)：pod-play 冷開機到建立隊伍主選單後以 F6 開面板，三張截圖都沒有缺字。

| 組合 | 畫面 | 截圖 SHA-256 |
|---|---|---|
| open | 目錄第 1／4 頁 | `46e1ddc6…` |
| next | 第 1 段「第1章 重回菲蘭城」第 1／1 頁 | `2aeef4a5…` |
| far | 第 55 段「第20章 Moander Minis」第 1／4 頁 | `046e48df…` |

## 待查

- 摘要保留英文的地名與物品名（例 Temple of Tyr、Hill Giant Steading、Silver Shield）要不要補進術語表，與遊戲內譯文對齊。
- 子代理標出的疑點（雜誌拼字不一致、編號模糊、地圖有編號但沒有說明）照實寫在摘要裡，沒有補。
- 遊戲中（冒險畫面）開關面板、與手札面板互斥的收據；地圖疊層（規格 009）未開始。
