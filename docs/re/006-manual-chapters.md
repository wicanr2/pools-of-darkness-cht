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

AGENTS.md 第 3.3 節的譯名優先序把珍157 放在銀色匕首之謎術語表之前，名稱照這個順序裁定。

| decision | 條件 | 列數 |
|---|---|---|
| 同名 | 珍157 與術語表相同 | 53 |
| 採珍157 | 不同，珍157 的譯名可用；術語表改成珍157 | 71 |
| 珍157誤譯 | 珍157 的名稱讓人誤會對象或效果，或與別的名稱撞名 | 5 |
| 珍157不一致 | 珍157 在不同處給同一英文兩種譯名，或給兩個英文同一譯名；取一種 | 9 |
| 珍157錯字 | 掃描字形與語意不合 | 1 |
| 留英文 | 珍157 沒有中文；有術語表或手札譯名的照用，其餘待 G5 全量 | 20 |

術語表因此改 48 筆、增 25 筆（含 `Shield (spell)`）。名稱改了的衍生詞與前作錯拼法另立的列一併改（12 筆）：`Wand of …` 的「魔法杖」改「魔杖」、`Neo Otyugh` 改「新觸手怪」、`Cold/Hot Fire Shield` 改「冷火之盾」「熱火之盾」，`Delay Blast Fireball`、`Otto's Irresitable Dance`、`Protection Evil, 10' Radius`、`Ressurection` 改成與正確拼法同譯。改動前後的術語表逐列比對過，名稱表與這 12 筆以外沒有其他改動。
被換掉的前作譯名登入 `text/term-avoid.tsv`（53 筆），只收不是新譯名子字串的詞；例如「火球」是「火球術」的一部分，不收。

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
| Detect Magic | 偵測魔法物品（牧師）、偵測魔法（德魯伊、魔法師） | 偵測魔法 |
| Dispel Magic | 解咒術（牧師）、解除法術（魔法師；內文說兩者相同） | 解除法術 |
| Remove Curse | 除去詛咒（牧師）、除咒術（魔法師） | 除去詛咒 |
| Protection From Evil, 10' Radius | 半徑十呎內免於邪惡侵害（牧師）、半徑十呎免受邪惡侵害（魔法師） | 半徑十呎內免受邪惡侵害 |
| Hold Person、Hold Monster | 都作束縛敵人 | 束縛敵人、束縛怪物（比照 Hold Animal 束縛動物） |
| Fire Giant | 火焰巨人（怪物章）、火巨人（手札 8 處） | 火巨人 |
| Elven Court | 精靈宮殿（怪物章，英文拼作 Elver）、精靈皇宮（手札） | 精靈皇宮 |
| Periapt | 護身符（與手札的 Amulet 同名） | 護符（前作） |

### 與銀色匕首之謎的差異

「採珍157」的 71 列多是法術名，前作沿用克萊恩英豪譯名，例如 Fireball 火球→火球術、Cone of Cold 冰錐術→冷凍光束、
Hold Person 束縛術→束縛敵人、Power Word Stun 暈迷術→牽制之咒、Cockatrice 雞蛇獸→雞蛇、Otyugh 奧提尤格→觸手怪（由巨型觸手怪類推）、Druid 得魯依→德魯依。
本作的法術與怪物名因此與前作不同。前作 docs/re/006 對珍056 譯名是否採用留待使用者決定；本作依 AGENTS.md 第 3.3 節已定的順序採用。

## 未做

- 規則章節（人物以及隊伍、戰鬥、法術總論）與怪物、法術說明正文的錯譯對照；只有 OCR 草稿。
- 留英文的 20 列與 `MON1CHA.DAX` 另外 64 個術語表沒有的名字（含 `GNT`、`LG` 等縮寫），G5 全量翻譯時一併定。
- UI 的 `Shield` 若同時用於物品與法術，G6 依呼叫位置區分。
