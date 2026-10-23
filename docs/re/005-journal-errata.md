# 005 珍157 旅行手札錯譯校正

日期：2026-10-23。Issue：[#5](https://github.com/wicanr2/pools-of-darkness-cht/issues/5)。來源見 [002](002-reference-sources.md)。

## 範圍

珍157 冒險者手冊第九章「旅行手札」（p.55–106）對照 SSI 英文 Adventurer's Journal 的 Journal Entries（英文手札 p.25–46）。
兩邊則號一一對應，**沒有第 13 則**：英文與珍157 都是第 12 則直接接第 14 則（對頁面圖確認），手札共 99 則。
說明書其餘章節（操作手冊、冒險者手冊的規則、怪物、法術）尚未對照。

## 方法

| 步驟 | 工具與產物 | 位置 |
|---|---|---|
| 珍157 OCR 草稿 | [`tools/ocr-batch.sh`](../../tools/ocr-batch.sh)：黑暗之后的 `dqk-ocr:ppocr5-r1`（`sha256:5625e878…`）與模型快取唯讀掛載，對 26 張跨頁掃描跑 PaddleOCR（`tools/ocr_batch.py`、`ocr_sample.py` 複製自銀色匕首之謎）；[`tools/zh_ocr_journal.py`](../../tools/zh_ocr_journal.py) 以文字框分布找書溝、分左右頁後切則 | `workplace/journal/zh157/` |
| 英文 OCR 草稿 | `pdftotext -bbox-layout` 取 PDF 第 15–25 頁（PDF 第 p 頁＝印刷頁 2p−5、2p−4，頁碼在頁首花邊裡讀不到）；[`tools/journal_en_ocr.py`](../../tools/journal_en_ocr.py) 依實測欄界（行首 x 聚在 0–20、190–210、400–410、580–590）分四欄、取大寫行為標題 | `workplace/journal/en/` |
| 對圖校讀與對照 | 四批 sonnet 子代理（第 1–25、26–50、51–75、76–100 則），共用指令檔 `workplace/journal/INSTRUCTIONS.md`；第一批先做、主 session 抽驗後才放其餘三批 | `workplace/journal/fix/b01`–`b04/` |
| 合併與檢查 | [`tools/journal_merge.py`](../../tools/journal_merge.py)：則號完整不重複、欄數、錯譯表的 zh157 是珍157 逐字稿的子字串、fix 是正式譯文的子字串 | — |
| 正式譯文 | `text/journal.zh-TW.tsv`（entry、title、body、source） | repo |
| 錯譯表 | `text/errata.tsv`，[`tools/errata_check.py`](../../tools/errata_check.py) | repo |
| 術語 | `text/glossary.tsv`（194 筆）；舊譯殘留 `text/term-avoid.tsv`（10 筆），[`tools/term_check.py`](../../tools/term_check.py) | repo |

逐字稿（英文校正稿 `workplace/journal/en/journal-en.tsv`、珍157 照原書用字的 `workplace/journal/zh157/journal-zh157.tsv`）只留在 `workplace/`。

修正原則（同銀色匕首之謎 docs/re/006）：
- 意思錯、漏譯、主客顛倒、專名誤譯才改；語感上可接受的意譯保留珍157 原文。
- 珍157 留英文的人名地名換成中文，先查前作術語表，沒有的自譯並登入 `text/glossary.tsv`。
- 地圖則（第 5、23、45、98 則）正文寫「（本則為地圖，見原版英文手札第 N 頁。）」；第 36 則文字附地圖。

## 統計

99 則中 90 則有修正，9 則照用珍157。錯譯表 353 條：

| 類別 | 條數 |
|---|---|
| 誤譯 | 265 |
| 漏譯 | 39 |
| 增譯 | 22 |
| 專名 | 17 |
| 語氣 | 4 |
| 錯字 | 3 |
| 譯名不一致 | 3 |

證據等級：已證實 320、強推論 32、假說 1。珍157 的譯法比珍056 自由得多，多則幾乎整段改寫（例第 4、7、9、14、15、16、22 則）。

## 系統性錯誤

| 原文 | 珍157 | 修正 | 影響 |
|---|---|---|---|
| The Vessel of Moander | 血脈 | 摩安德的容器 | 第 4、56、79 則；第 4 則另加了原文沒有的「這些血管是活的」 |
| Night Spawn／hellspawn | 黑夜之卵 | 夜裔／地獄之裔 | 第 42、80 則等 |
| drider | 半蜘蛛人 | 半蜘精靈 | 達里斯被變成的怪物；舊譯讓人以為他不是精靈 |
| drow | 黑精靈 | 黑暗精靈 | 全手札統一 |
| Storm（人名） | 風暴 | 史托姆 | 第 35、80 則；阿爾塔的哥哥，英文提示書列為矮人 |
| Troll Toss Tavern | 巨人酒店 | 擲巨魔酒館 | 第 9 則整段被改寫成巨人搶劫旅客 |

## 影響劇情理解的誤譯（例）

| 則 | 原文 | 珍157 | 修正 |
|---|---|---|---|
| 34 | Those storms will no longer plague the Realms | 這些風暴不久就會侵襲到王國來 | 這些風暴將不會再肆虐王國 |
| 47 | these tempests have prevented our allies … from entering these lands | 暴風雪將會領導我們的軍隊…穿過那些山脈 | 先前，這些暴風一直阻擋山脈另一側的盟軍 |
| 50 | I have sent Death to the Realms, a rider in red with a sword of flame | 我已經將死亡散播到王國的每一個角落去了 | 我已經派死神降臨王國，那是一名身穿紅衣、手持火焰之劍的騎士 |
| 67 | but as a dracolich, none shall stand against you | 除了龍巫妖之外，沒有人能勝過你 | 但成為龍巫妖之後，將無人能與你為敵 |
| 72 | Even the best of thieves find themselves in prison once in a while, but to be held as fodder for the dragons is too much for me | 成為囚犯已經是最大的懲罰了，但是對我來說，至少比被捉去餵龍要好的多了 | 再高明的賊偶爾也難免進牢裡，但被關起來當龍的飼料，我可受不了 |

## 主 session 的譯名裁決

批次之間的分歧在合併時統一（逐則替換正式譯文與錯譯表的 fix 欄）：

| 原文 | 採用 | 理由 |
|---|---|---|
| Gothmenes | 戈斯梅尼斯 | 第 60、62 則原作戈斯梅內斯 |
| Tanetal | 塔內塔爾 | 第 58 則原作塔奈塔爾 |
| Silk | 絲綢 | 青色枷的詛咒的譯名；第 73 則原作絲克 |
| Fire of Night | 黑夜之火 | 珍157 譯名；第 71 則原作夜之聖火 |
| spirit of fire | 火神 | 第 57 則指附身普洛菲力的 Tyranthraxus，光芒之池稱火神泰蘭特拉克斯；保留珍157 以接前作 |
| Zhentarim | 桑塔林會 | 珍157 留英文；字頭跟珍157 的「桑提爾城」一致（前作各作漢塔林會、漢塔林、散塔林會） |
| the Ward（第 88 則） | 通行證 | 與第 44、51、55 則的 Ward of Safe Passage 一致；刪去把它改成「守護符」的假說條目 |

## 查驗題

`GAME.OVR` 的題型是「(journal, page N) what is word # M after the heading …」（docs/re/001）。題目表的標題字串在遊戲檔、
DAX 解壓與 6-bit 解碼後都找不到明文，本階段沒有靜態取得題目表。改在 G7（#8）於執行期攔題目，答案以珍157 操作手冊附錄二的密碼表對照。

## 未做

- 說明書其餘章節未對照。
- 英文校正稿只在 `workplace/`；英文手札面板（多語系，G10）前決定是否進版控。
- 第 26 則 Gelt、第 36 則 bad robe 等子代理標為拿不準的項目，G5 試譯時再看遊戲內文。
