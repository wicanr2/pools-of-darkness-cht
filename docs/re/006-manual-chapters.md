# 006 珍157 說明書其餘章節

日期：2026-10-23。Issue：[#5](https://github.com/wicanr2/pools-of-darkness-cht/issues/5)、[#6](https://github.com/wicanr2/pools-of-darkness-cht/issues/6)。
範圍比照銀色匕首之謎 docs/re/006「說明書其餘章節」：全冊 OCR 草稿，加上遊戲內會出現的名稱逐張對圖查證。旅行手札（冒險者手冊第九章）見 [005](005-journal-errata.md)。

## 轉錄草稿

操作手冊 p.1–63（檔號 6–37，32 張跨頁）與冒險者手冊 p.1–54（檔號 3–29，27 張）以 `tools/ocr-batch.sh` 辨識（同 005 的映像與模型快取，兩批 0 錯誤），
[`tools/zh_manual_draft.py`](../../tools/zh_manual_draft.py) 依書溝分左右頁、略過頁首花邊與頁碼，組成逐頁草稿 `workplace/manual-zh/ocr-draft.tsv`（117 頁，不進 Git）。
操作手冊檔號 37 的右頁空白，不列。頁碼依 002 的檔號換算：操作手冊 p.2n−11、冒險者手冊 p.2n−5。

**整份草稿未校讀**：OCR 有簡體字形（项、谜、録）與錯字。只有下節的名稱逐張對過頁面圖。
冒險者手冊第二至八章與英文 Adventurer's Journal 前半的目錄同序（Characters and Parties、Building a Successful Party、Combat、Magic、
Magic Treasures、Creatures of the Forgotten Realms、Spell Descriptions）。操作手冊 63 頁對應 9 頁的英文 Rulebook，多出的段落有沒有英文對應未逐段確認。

## 名稱對照（冒險者手冊 p.21–54）

第六章魔法寶物（p.21–22）、第七章被遺忘王國中的生物（p.23–36）、第八章各項法術的說明（p.37–54）的條目名稱，
與術語表和遊戲內字串比對，結果在 [`text/zh157-names.tsv`](../../text/zh157-names.tsv)（159 列，只有名稱；含章節標題「德魯依法術」與內文的「德魯依」）。

| 步驟 | 工具與產物 |
|---|---|
| OCR | 17 張跨頁（檔號 13–29）以 `tools/ocr-batch.sh` 辨識，輸出 `workplace/manual-zh/ocr-names/` |
| 候選 | `tools/zh157_names.py draft` 抽「N. 中文名（English）」與「(N)中文名（English）」，115 個候選。珍157 有 20 條怪物名只留英文、部分條目被拆行，候選不完整 |
| 轉錄 | 主 session 逐張看 17 張跨頁，抄錄每個條目標題與內文提到的英文對照；原書用字照抄（「沈靜術」「名喚怪物」） |
| 遊戲內英文 | 法術名取 `GAME.EXE` 解開後映像的 UI 字串（`workplace/text-src/ui-source.tsv`）；怪物名取 `MON1CHA.DAX` 各區塊開頭的名字（[`tools/monster_names.py`](../../tools/monster_names.py)，124 個） |
| 檢查與寫入 | `tools/zh157_names.py check`：同一英文的採用名必須一致、與術語表一致；`--apply` 把採用名寫進 `text/glossary.tsv` |

`game_en` 欄是遊戲內的英文。珍157 的英文與遊戲不同時以遊戲為準（例 `Hold Monster` 遊戲作 `Hold Monsters`、`Otto's Irresistible Dance` 作 `Otto's Irr. Dance`）。
159 列中 35 列在 `MON1CHA.DAX` 找得到、114 列在 UI 字串找得到，其餘 10 列（內文提到的怪物、地名與章節標題）遊戲字串裡沒有同形的英文。

### 裁定規則

使用者 2026-10-23 決定：法術、怪物、物品名與銀色匕首之謎一致。依序取前作術語表、前作遊戲內譯文（`text/ui-text.zh-TW.tsv` 同一英文字串）；
前作都沒有的，能依前作命名方式類推的先類推，其餘用珍157，珍157 有誤時另裁。AGENTS.md 第 3.3 節記這條例外。

| decision | 條件 | 列數 |
|---|---|---|
| 同名 | 珍157 與前作相同 | 52 |
| 沿用前作 | 珍157 與前作不同，用前作（例 火球術→火球、束縛敵人→束縛術、雞蛇→雞蛇獸、德魯依→得魯依） | 57 |
| 前作類推 | 前作沒有，依前作同類名稱組成 | 7 |
| 採珍157 | 前作沒有，用珍157 | 13 |
| 珍157誤譯 | 珍157 的名稱讓人誤會對象或效果，或與別的名稱撞名；採用名與前作相同 | 5 |
| 珍157不一致 | 珍157 在不同處給同一英文兩種譯名；取與前作或手札相同的一種 | 4 |
| 珍157錯字 | 掃描字形與語意不合 | 1 |
| 留英文 | 珍157 沒有中文；用前作、手札譯名或自譯 | 20 |

前作術語表本身有幾處同一法術兩種寫法，取前作遊戲內譯文：`Protection From Evil, 10' Radius` 作「半徑10呎內免受邪惡傷害」（前作表另有「不受」），
`Disintegrate` 作「崩解術」、`Delayed Blast Fireball` 作「延遲爆炸火球」（前作表的錯拼法列作分解術、延時爆發火球）；錯拼法列改成同譯。
前作術語表的 `ID` 作「驗証」，前作遊戲內譯文與 term-avoid 都是「驗證」，從後者。

術語表相對名稱裁定前增 85 筆（說明書名稱 25、`MON1CHA.DAX` 怪物名 60）、改 4 筆（上述三個錯拼法列與 `ID`）。
`MON1CHA.DAX` 的 124 個名字（60 個不重複的新名字含 `GNT`、`LG`、`MAH`、`F K` 等縮寫）現在都有譯名，組法沿用前作基底詞（食人魔、骷髏、羅剎、火刀、紅羽衛等），Shaman 照珍157 作巫師。
禁用表 `text/term-avoid.tsv` 收本作原有 10 筆、珍157 誤譯與不一致 9 筆、前作 term-avoid 19 筆（排除前作題名規則與「火神」，本作的 spirit of fire 作火神）。

### 珍157 誤譯與錯字

| 英文 | 珍157 | 採用 | 理由 |
|---|---|---|---|
| Cloud Giant | 煙霧巨人 | 雲巨人 | Cloud 是雲（銀色匕首之謎的珍056 同錯） |
| Umber Hulk | 地底熊 | 土巨怪 | 不是熊（前作同判） |
| Silence 15' Radius | 半徑15格的沈靜術 | 半徑15呎的沉靜術 | 15' 是 15 呎，內文也寫半徑 15 呎 |
| Invisibility to Animals | 動物隱形術 | 對動物隱形 | 讀成動物隱形；法術是讓目標對動物隱形 |
| Barkskin | 鐵皮術 | 硬皮術 | bark 是樹皮；術語表的 Iron Skin 作鐵皮術，會撞名 |
| Monster Summoning | 名喚怪物 | 召喚怪物 | 掃描上像「名」；同書 p.41「名喚火柱」、手札第 26、74 則同，名／召分不清（005） |

### 珍157 不一致

| 英文 | 珍157 | 採用 |
|---|---|---|
| Dispel Magic | 解咒術（牧師）、解除法術（魔法師；內文說兩者相同） | 解除法術（同前作） |
| Fire Giant | 火焰巨人（怪物章）、火巨人（手札 8 處） | 火巨人 |
| Elven Court | 精靈宮殿（怪物章，英文拼作 Elver）、精靈皇宮（手札） | 精靈皇宮 |
| Periapt | 護身符（與手札的 Amulet 同名） | 護符（前作） |

另有 Detect Magic、Remove Curse、Protection From Evil 10' Radius、Hold Person／Hold Monster 在珍157 內部也不一致，採用名直接沿用前作，記為「沿用前作」。

## 未做

- 說明書正文（規則、怪物與法術說明）不對照錯譯（使用者 2026-10-23：以完成中文化為先）；只有 OCR 草稿。
- UI 的 `Shield` 若同時用於物品與法術，G6 依呼叫位置區分。
