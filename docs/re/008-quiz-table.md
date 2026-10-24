# 008 手冊查驗題題目表

日期：2026-10-23。Issue：[#8](https://github.com/wicanr2/pools-of-darkness-cht/issues/8)。前置：[004](004-output-paths.md)、[007](007-g6-normal-path.md)。

後續補證：2026-10-25，同一正常讀槽A路線的查驗題替換時序，見[規格006](../spec/006-quiz-rendering.md)。
`workplace/probe/g9-hd/wall-transition-v1/source/`保留3500–3509格配對idx／pal、原版及合成圖、印字盤點與執行輸入雜湊。
3503格答案已顯示，但頁碼三格及字序問句六格仍是原版；3504格被中文替換，完整路線量尺9。
當時街景完整組合格命中0，不能由螢幕座標落在3D視窗就當成街景閃現。
固定來源、六段呼叫與雜湊見[轉場正對照收據](008-quiz-transition-source.json)，
由[`tools/quiz_transition_source.py`](../../tools/quiz_transition_source.py)重生。

逐格診斷`TestQuizSourceTransition`使用上述真實3502–3505格及正式字型。
第一格原文字形尚未完成，舊譯文以Early顯示；下一格原文完成才返回並定色。
頁碼先於採樣出現，其餘段落於採樣後完成組句。舊行為重現完全相同9個座標，繼承Early／BG／FG後0。
診斷沒有執行原版CPU，正常路徑另由`tools/wall-play-receipt.sh`驗證。

## 證據來源

| 證據 | 取得方式 |
|---|---|
| 記憶體傾印 | `tools/g6-saveload-receipt.sh` 的 `lbegin` 狀態（查驗題畫面），探針 `-dump-mem-at '1200050000:lin:0:655360:/out/mem.bin'`（`workplace/probe/g7-quiz/`） |
| 答案寫入者 | 同一段重跑加 `-watch 2874A-28755`：`5509:011F` 清旗標後，`1ED4:0AA2`／`0AA8`（字串複製）把 `05 'LIVES'` 寫到 `DS:7DEE`（本次 DS＝`2096`） |
| 反組譯 | [`tools/re/dis16.py`](../../tools/re/dis16.py)，`mem.bin` 段 `5509` 偏移 `0000`–`0318` |
| 靜態表 | [`tools/quiz_table.py`](../../tools/quiz_table.py) 讀 `workplace/text-src/game-image.bin`（GAME.EXE 解開 EXEPACK 後的映像，`tools/g3-game-image.sh`） |

查驗題程序所在的 overlay 段這次載在 `5509`（G6 建角收據那次在 `37F7`），段位址隨 overlay 載入而變，以下偏移以段內偏移記。

## 程序（`5509:00F4`–`0317`，已證實）

| 偏移 | 行為 |
|---|---|
| `0000`–`00A1` | 解碼副程式：Pascal 字串長度 L 不變，內容倒轉、每個字元減 L |
| `0124` | `Random(21)+1` → `DS:7FF0`，選第 k 筆標題 |
| `0131`–`0144` | 記錄位址 `DS:0x60 + 0x62×k` |
| `0147`–`0163` | `Random(3)+1` → `DS:7FF1`，選第 j 題；記錄 `+0x1F+j` 是 0xFF 就重抽 |
| `0165`–`018D` | 解碼記錄 `+0x0E+0x15×j` 的答案，存到 `DS:7DEE` |
| `0192`–`029B` | 路徑 A 印六段（下表） |
| `02B0`–`02C7` | 底列輸入（`1B2E:0711`，最長 20 字）存到 `DS:7EEE` |
| `02D1`–`02E9` | 輸入與答案不同且答錯未滿 3 次（`DS:7FEE`）就回 `0116` 重抽一題 |
| `02EC`–`0317` | 回傳是否答對 |

答錯會換題，所以 adapter 每看到一組六段就獨立改寫一次；收據只收預期的題目，題目變了就停（007）。

## 畫面（路徑 A 六段，007 的兩支收據）

| 段 | 列、欄 | 格數 | 內容 |
|---|---|---|---|
| 1 | 2、3 | 15 | ` journal, page `（catalog `ui.b001a1701050`） |
| 2 | 2、17 | 13 | ` (P)` 補空白到 13 格，未命中 |
| 3 | 4、3 | 16 | ` what is word # `（`ui.3cf3aa82aeb1`） |
| 4 | 4、19 | 6 | 字序，補空白到 6 格，未命中 |
| 5 | 6、3 | 19 | ` after the heading `（`ui.39ce92102d66`） |
| 6 | 8、4 | 標題長 | 標題，未命中 |

第 1 段從第 3 欄佔到第 17 欄，第 2 段也從第 17 欄起印，兩段重疊一格空白。原版畫面讀作「JOURNAL, PAGE (4) / WHAT IS WORD # 7 / AFTER THE HEADING / ALIGNMENT」。

## 題目表（已證實）

GAME.EXE 映像 0x9F32 起 21 筆，每筆 0x62 bytes。起點與程式的 k＝1 相符（強推論）：`ALIGNMENT` 的加密字串在傾印 0x20C0C、映像 0xA11C，以 DS 線性 0x20960 換算，資料段在映像 0x9E70，`0x9E70+0x60+0x62`＝0x9F32；往前一筆欄位不合。

| 位移 | 長度 | 內容 |
|---|---|---|
| `+0x00` | 31 | 標題，string[30]，加密 |
| `+0x1F` | 1 | 英文手札頁碼 |
| `+0x20`–`0x22` | 3 | 第 1–3 題的字序，0xFF 表示沒有 |
| `+0x23`、`+0x38`、`+0x4D` | 各 21 | 第 1–3 題的答案，string[20]，加密 |

`tools/quiz_table.py` 以實機答對的「第 4 頁 ALIGNMENT 第 7 字 LIVES」定位，前後以 0x62 延伸到欄位不合為止，筆數不是 21 就報錯。
解出 21 個標題、54 題，頁碼 1–10，輸出 [`text/quiz-answers.tsv`](../../text/quiz-answers.tsv)（`page`、`word`、`heading`、`answer`）。

## 核對

- 實機作答通過 2 題：第 2 頁 ABILITY SCORES 第 3 字 HAS（`tools/g6-create-receipt.sh`）、第 4 頁 ALIGNMENT 第 7 字 LIVES（`tools/g6-saveload-receipt.sh`）。
- 英文手札 PDF（`workplace/input/en-docs/mocagh/pod-journal.pdf`，`pdftotext -layout` 第 1 頁）：INTRODUCTION 正文「For ten years your glorious band」、
  CHARACTERS AND PARTIES 正文「You need a party」，表中第 2、3、6 字 TEN、YEARS、BAND 與第 1、2、4 字 YOU、NEED、PARTY 都對上。字序從標題後正文第一個字算起。
- 答案來源是原版 GAME.EXE，不取自珍157 操作手冊附錄二；收據的兩題與附錄二相同。

## 改寫與收據（規格 006）

`apps/pod/quiz.go` 依序認六段（列、欄照上表；中間夾別的呼叫就放棄這一組），第 6 段時改寫：

| 列 | 格數 | 內容（`text/host-text.zh-TW.tsv`） |
|---|---|---|
| 2 | 27（延伸到頁碼段結尾） | `quiz.page`「英文手札第 {page} 頁」 |
| 4 | 22（延伸到字序段結尾） | `quiz.word`「下列標題之後的第 {word} 個字？」 |
| 6 | 19 | `quiz.answer`「答案：{answer}」；答案關閉或表中沒有時空白 |
| 8 | 標題長 | 標題留英文；萬一命中 catalog 也清掉疊字 |

答案表鍵是（頁、字序、標題）。F9 切換答案顯示，只改覆繪。

| 收據 | 結果 |
|---|---|
| [`tools/g7-quiz-receipt.sh`](../../tools/g7-quiz-receipt.sh) | 題目表重產與 `text/quiz-answers.tsv` 相同；第 4 頁 ALIGNMENT 題答案顯示與隱藏兩張，範圍外差異 0，第 6 列分別是「答案：LIVES」與空白 |
| `tools/g6-create-receipt.sh`（begin、answer） | 第 2 頁 ABILITY SCORES 題顯示「答案：HAS」，範圍外差異 0；作答後進入遊戲 |
| `tools/g6-saveload-receipt.sh`（lbegin、lanswer） | 第 4 頁 ALIGNMENT 題顯示「答案：LIVES」，範圍外差異 0；作答後回到街上 |
