# 《黑暗之池》（Pools of Darkness）dosgolem 繁體中文化與 HD

本檔是 Claude 與 Codex 共用的專案入口。回覆、文件與提交訊息預設使用繁體中文。
新工作階段先讀本檔，再讀 `CONTEXT.md`、目前 GitHub Issue 與直接相關規格。
檔案尚未建立時明列缺項，不把前作狀態當成本作進度。

## 1. 使用者授權與專案範圍

使用者於 2026-10-23 指定：

- 參考 `/home/anr2/cht/golden_box/secret-of-the-sliver-blader/`（以下稱銀色匕首之謎）的經驗，用 dosgolem
  執行原版 DOS《Pools of Darkness》，以輸出攔截覆繪做繁體中文化，並把原版素材做成 HD。
  中文化與 HD 的步驟完全比照銀色匕首之謎。執行與呈現的整合介面一律用 dosgolem。
- 同日回答：repo `pools-of-darkness-cht`、遊戲短名 `pod`；G0 全做（文件、Git、private repo、初始 Issue）；
  授權上網取得英文手冊與 Adventurer's Journal；顯示模式 VGA 優先。

「步驟比照銀色匕首之謎」包含：修正當年中文譯本的錯譯，對象是 `珍157-黑暗之池.rar` 內的軟體世界珍藏版
說明書；說明書裡的手札修正後放進遊戲，玩家不必再查說明書；英文手冊與 Adventurer's Journal 作為校正原文。

儲存庫名稱 `pools-of-darkness-cht`，owner 以主機 `gh auth status` 回讀的登入帳號為準。
建立前查同名庫；存在時核對用途，不接管、不覆寫。本授權含初始化本目錄 Git、設定 repo-local
身分 `wicanr2@gmail.com`、提交與推送本範圍內的文件、原創程式與建置腳本，以及建立、更新、
留言、關閉本專案 Issue。轉公開、公開 Release、上傳原版素材、加入協作者、改動前作或其他
專案、推送 dosgolem 上游及破壞性操作，另行取得授權。

下表沿用銀色匕首之謎截至 2026-10-23 的定案（其 `AGENTS.md` 第 1 節與 `CONTEXT.md` 有效決策），
使用者以「步驟完全跟銀色匕首之謎一樣」確認，不重複詢問。

| 項目 | 定案 |
|---|---|
| 中文題名 | 「黑暗之池」（使用者指定，與珍157 題名相同） |
| 遊戲短名 | `pod`，用於 dosgolem `apps/pod/` 與程式識別 |
| 工作順序 | 先完成繁中、錯譯修正、手札與查驗題，再做 HD 與其他擴充 |
| 顯示模式 | VGA 優先（使用者 2026-10-23）。G2 確認 dosgolem 能以 VGA 跑到第一個可操作畫面；跑不起來時回報，不自動改用 EGA |
| 手札 | 修正後的說明書手札全部放進遊戲，劇情引用時自動彈出，另以前端鍵（F5）翻閱已觸發的手札。說明書有酒店傳聞等其他引用段落時一併處理 |
| 手冊查驗題 | 問句後顯示答案，預設開啟、可關閉（F9）；玩家仍要作答，判定照原版。題型是 Adventurer's Journal 頁碼、標題後第 N 字（docs/re/001） |
| HD | HD 圖以 3 倍（960×600）為準，風格 B（奇幻插畫），配色照原版。預設 2 倍顯示原版，F2 切 3 倍，HD 圖只在 3 倍顯示 |
| HD 分工 | 來源盤點、請求包、前處理、驗收對位、接入與規格 021 不露原版圖由 Claude 處理；只有依原版手繪重畫（生圖）交 Codex，介面見第 6 節。不用濾鏡放大。第一批先給使用者確認畫風再全量 |
| 切換列 | 視窗最上方，平常隱藏，滑鼠移到頂端時出現；HD 與擴充功能切換放這裡。沒有 HD 素材時停用 HD 切換並顯示原因。F12 切回英文原版 |
| 切換列字級 | 小、中、大，記住上次選擇。3 倍 16／22／24、2 倍 12／16／16；只影響切換列（銀色匕首之謎規格 020） |
| CPU 模擬速度 | 切換列「速度」下拉 1x–32x，只顯示倍率，記住上次倍率，不加快速鍵（銀色匕首之謎規格 019） |
| 開場 demo | 選單不逾時，demo 中按鍵中止回選單（銀色匕首之謎規格 018）；`GAME.EXE` 有 `Play Demo`（docs/re/001） |
| 不露原版 | 所有語言與 HD 模式都不先顯示原版字或原版圖再換成譯文或 HD（銀色匕首之謎規格 021）。必要時可在記憶體修改原版程式，不改磁碟上的原版檔、不散布修改後的 EXE |
| 字型 | 繁中源泉圓體（2 倍 Medium、3 倍 Bold），比照銀色匕首之謎現行版本 |
| 多語系 | 繁中先交付，之後比照克萊恩英豪規格 015：简中由繁中 OpenCC 轉換加詞彙表，日、韓以英文為底機器翻譯並標註，英文以新字型重繪；第一次執行依系統語言，F8 切換 |
| 攻略面板與地圖疊層 | 遊戲內可翻頁攻略面板（F6）與地圖疊層攻略點（F7）。來源是使用者提供的《軟體世界》第 43–45 期掃描（`walkthrough/`），缺的部分以 SSI 官方提示書補 |
| 配樂 | 以原版動機新編 OPL2 配樂，依地圖與戰鬥切換；使用者試聽通過才定案 |
| 作弊 | 我方無敵（F1，預設關）：HP 不減少，不進入昏迷、中毒、麻痺、睡眠等異常。位址在本作重驗後才啟用 |
| 推廣影片 | 故事背景字卡、HD 切換與 F12 切換示範 |
| 發行包 | 三平台（Linux AppImage、Windows zip、macOS universal zip）；可散布 patch 分原版、原版＋HD 兩包；本機另出完整版，不散布。只私下分享給持有原版的人，不公開上架 |
| 截圖入 repo | README 展示用截圖放 `docs/screenshots/`；其他截圖只留 `workplace/` |
| 驗證抽樣 | 收據截圖與驗證抽樣即可，不窮舉畫面與路線 |
| 與使用者溝通 | 原型並列與取捨確認可用 Artifact；規劃文件仍寫在 `docs/` |

可逆的讀取、整理、文件修正與已授權測試自主進行。改變目標、架構、原版忠實度、正式資料
格式、玩家操作、授權或公開方式時，先依第 10 節共同決策。

## 2. 初始化順序與基線

初始化依序完成：本檔 → `.gitignore` → 私人儲存庫與 Git → 初始 Issue → 證據與規格 → 實作。
建立 Issue 前不解壓原版、不跑遊戲、不寫產品程式。遠端 Issue 是可執行工作狀態的權威，本機
不建 `WORKLIST.md`。未建立的 Issue 不填編號或完成狀態。

| 項目 | 2026-10-23 初始觀察 |
|---|---|
| 原始輸入 | 根目錄兩個壓縮檔；只讀取成員清單與 SHA-256，尚未解壓（第 3 節） |
| 英文原文 | 尚未取得 |
| 雜誌攻略 | 《軟體世界》第 43–45 期掃描 31 張在 `walkthrough/`，尚未轉錄 |
| 程式與驗證 | 沒有本作程式、執行收據或中文化證據 |

規格編號沿用銀色匕首之謎 `docs/spec/000-index.md` 的 001–021（001–017 對照克萊恩英豪，018–021 是銀色匕首之謎新增），
同號規格的契約相同，本作狀態另計。初始工作已登錄為
[#1–#16](https://github.com/wicanr2/pools-of-darkness-cht/issues)，狀態只在遠端維護。之後使用者新增的需求另開 Issue，不改下表。

| 階段 | 依賴 | 完成條件 |
|---|---|---|
| [G0 協作治理／#1](https://github.com/wicanr2/pools-of-darkness-cht/issues/1) | 本檔 | `.gitignore`、`LICENSE`、`CONTEXT.md`、`WORKLOG.md`、`CLAUDE.md`、`README.md` 就位；private 與 Issues 回讀 |
| [G1 原始輸入／#2](https://github.com/wicanr2/pools-of-darkness-cht/issues/2) | G0 | 壓縮檔與成員 SHA-256；版本與發行形式、入口與啟動參數、顯示與音效驅動的選擇方式；`SAVE/` 的性質與權利分類；查驗題形式；珍157 頁序與缺頁；攻略掃描頁序；英文手冊與手札的來源網址、取得日期、SHA-256 |
| [G2 dosgolem 基線／#3](https://github.com/wicanr2/pools-of-darkness-cht/issues/3) | G1 | 選定並固定 dosgolem 基底（第 5 節）；確認實際顯示模式與驅動載入方式；VGA 正常入口到第一個可操作畫面；缺服務與重播收據 |
| [G3 輸出路徑／#4](https://github.com/wicanr2/pools-of-darkness-cht/issues/4) | G2 | 印字、清除、捲動與選單掛勾的原始定位與證據等級；驗證第 4 節的待驗假說 |
| [G4 說明書轉錄與錯譯校正／#5](https://github.com/wicanr2/pools-of-darkness-cht/issues/5) | G1 | 珍157 手札逐頁轉錄；與英文原文逐段對照；錯譯表（第 3.2 節）與檢查腳本。說明書其餘章節是否對照待使用者決定 |
| [G5 譯文與字型／#6](https://github.com/wicanr2/pools-of-darkness-cht/issues/6) | G3、G4 | 術語表、TSV catalog、字型與檢查工具；先試譯一批給使用者看文風再全量 |
| [G6 繁中接入／#7](https://github.com/wicanr2/pools-of-darkness-cht/issues/7) | G3、G5 | 每條路徑 READY 規格、覆繪、A/B 收據與正常玩家路徑；含組句、英數半形、玩家名音譯 |
| [G7 手札與查驗題／#8](https://github.com/wicanr2/pools-of-darkness-cht/issues/8) | G3、G4 | 手札引用與查驗題的輸出證據、修正後譯文、自動彈出與翻閱、答案顯示的收據 |
| [G8 攻略面板與地圖疊層／#9](https://github.com/wicanr2/pools-of-darkness-cht/issues/9) | G6 | 《軟體世界》第 43–45 期轉錄、術語對齊、面板與地圖疊層收據 |
| [G9 HD 與切換列／#10](https://github.com/wicanr2/pools-of-darkness-cht/issues/10) | G6、G7 | Claude：來源與遮擋證據、請求包、驗收、接入、F2 與切換列收據、規格 021 不露原版圖；Codex：生圖 |
| [G10 配樂、多語系、作弊、推廣片／#11](https://github.com/wicanr2/pools-of-darkness-cht/issues/11) | G6、G9 | 各自規格與收據；配樂經使用者試聽；無敵位址在本作重驗 |
| [G11 封包驗收／#12](https://github.com/wicanr2/pools-of-darkness-cht/issues/12) | G6、G9 | 乾淨目錄啟動、manifest 與 SHA-256、存檔位置；明列未驗平台 |
| [規格 018 開場 demo 控制／#13](https://github.com/wicanr2/pools-of-darkness-cht/issues/13) | G3 | 選單逾時、demo 旗標與 demo 進出點的本作位址；中止收據 |
| [規格 019 CPU 模擬速度／#14](https://github.com/wicanr2/pools-of-darkness-cht/issues/14) | G2、切換列 | 本作 CRT `Delay` 寫法；1x 快轉開關等價收據；各倍率實測 |
| [規格 020 切換列字級／#15](https://github.com/wicanr2/pools-of-darkness-cht/issues/15) | 切換列 | 小、中、大三級收據 |
| [規格 021 不露原版／#16](https://github.com/wicanr2/pools-of-darkness-cht/issues/16) | G6、G9 | 閃現量尺 0：文字路徑與 HD 各以抽樣路線量測 |

## 3. 原始輸入與資料保護

| 根目錄檔案 | SHA-256 | 成員 | G1 盤點（2026-10-23） |
|---|---|---|---|
| `pools-of-darkness.zip` | `a67022ffd89128cc1e35b6fefa79e456fd94ee03ea5ec26c0082f4bebc770483` | 已展開的遊戲目錄 `pools-of-darkness/`：遊戲檔 47 個，`SAVE/` 78 個。入口候選 `START.BAT`、`STARTUP.EXE`、`GAME.EXE`＋`GAME.OVR`；另有 `CONTROL.EXE`、`VACE.EXE`；顯示驅動 `EGA.DRV`、`VGA.DRV`、`TGA.DRV`；音效 `ADLIB.COM`、`SBLASTER.COM`、`SOUND.COM` 等；動畫 `INTRO`、`FINAL` 的 `.BMA`、`.ACF` | 已安裝目錄，版本 1.10；`SAVE/` 是 1995 年玩家資料，比照原版素材不散布（[docs/re/001](docs/re/001-input-baseline.md)） |
| `珍157-黑暗之池.rar` | `3d18ec1b0d588705484f53f16921ed90af9343f3246af8b9f00f99677adcf623` | `珍157-黑暗之池/` 下 107 張 JPG（檔名分 4 個掃描批次，各 38、32、28、9 張）、`Thumbs.db`、`軟體世界說明書補完計劃.txt`；沒有 PDF | 操作手冊 63 頁、冒險者手冊 110 頁，無缺頁；手札 99 則（無第 13 則）在冒險者手冊 p.55–106（[docs/re/002](docs/re/002-reference-sources.md)、[005](docs/re/005-journal-errata.md)） |
| `walkthrough/` | 逐檔雜湊見 [docs/re/001-input-files.tsv](docs/re/001-input-files.tsv) | 《軟體世界》第 43 期 11 張、44 期 12 張、45 期 8 張（`.jpeg`；`65 (1).jpeg`、`67 (1).jpeg` 檔名含空白與括號） | 「完全攻略」三回，第 1–24 章連續（[docs/re/002](docs/re/002-reference-sources.md)）；頁內段落對應待 G8 |

- 原始壓縮檔與 `walkthrough/` 保留原位、不覆寫，只在 Docker 內唯讀掛載。解壓、OCR、雜湊與探針產物寫入
  `workplace/`。
- `.gitignore` 在 `git init` 前排除上述檔名、`workplace/`、`dist-all/`、`walkthrough/`、原版副檔名、存檔、憑證與本機快取；
  G1 盤點後補齊。提交只列明確檔案，不用 `git add .`。
- 原版遊戲、手冊與攻略掃描、英文原文 PDF、存檔與擷取的原版素材不進 Git、GitHub 或可散布
  發行包。需要原版的測試缺檔時明確 skip。
- 例外（沿用銀色匕首之謎）：譯文 TSV（含修正後的手札）、英文手札逐字稿、翻譯批次與提示詞、
  HD 主題、攻略與提示書轉錄提交到 private repo 保存。repo 轉公開前必須移除這些內容並清理歷史。
- 例外（沿用銀色匕首之謎）：HD 審閱頁 `review/hd/`（原版 3 倍參考圖與 HD 預覽並列）放 private repo
  供 GitHub 上審閱。它含擷取的原版圖，轉公開前同樣移除並清理歷史。

### 3.1 英文原文取得

- 下載在 Docker 內進行，只對該次下載開網路。每個來源放在 `workplace/input/en-docs/<來源>/`
  的獨立空目錄，記錄網址、取得日期、檔名與 SHA-256。
- 下載物是不可信資料：不執行、不從其目錄啟動直譯器，讀取它的 Python 以 `-I` 執行。
- 優先找 SSI 原版掃描（手冊、Adventurer's Journal、提示書）。銀色匕首之謎的來源是 mocagh.org（其 `docs/re/002`）。
  來源不明或為重打字版時標明，只作輔助。

### 3.2 錯譯修正

- 珍157 的轉錄維持原書用字，校正不改轉錄本身。方法照銀色匕首之謎 `docs/re/006` 與光芒之池的說明書 corpus
  （`/home/anr2/cht/golden_box/pool_of_radiance/docs/spec/054-softworld-manual-journal-corpus.md`
  與同專案 `docs/reference/manual/glossary.md`）。
- 錯譯以資料表記錄：`text/errata.tsv`，欄位為 key、頁碼、英文原文摘錄、珍157 譯文摘錄、修正譯文、
  類別（誤譯、漏譯、異名、專名不一致等）、理由、證據等級、出處。
- 遊戲內手札與其他玩家可見譯文的正式來源是 `text/` 下的 TSV，以修正後版本為準。
- 檢查腳本驗證欄數、key 唯一、頁碼存在、證據等級合法，並確認每條修正都已出現在正式 TSV；
  不一致時結束碼非 0。「已修正錯譯」只指表中通過檢查的條目。
- 英文原文與遊戲內實際字串衝突時，以遊戲內字串為準，並在錯譯表註明。
- 題名「黑暗之池」保留，不列入修正。

### 3.3 譯名與文風

本作是 Forgotten Realms 系列 Gold Box，接在光芒之池、青色枷的詛咒、銀色匕首之謎之後，不是克萊恩世界。
克萊恩英豪、幽靈騎士與黑暗之后的術語表只沿用遊戲機制與共通物品，地名、人名、神祇不從那裡取。

譯名優先序：

1. 本作遊戲內原文語境。
2. 珍157 譯名，經第 3.2 節修正後的版本。
3. 銀色匕首之謎 `text/glossary.tsv` 與 `text/term-avoid.tsv`（同系列直接前作）。
4. 更早的前作：`/home/anr2/cht/golden_box/pool_of_radiance/docs/reference/manual/glossary.md`、
   `/home/anr2/cht/golden_box/curse_of_the_azure_bonds/docs/knowledge/coab-glossary.md`。
5. 本作自譯，在術語表註明。

- 文風比照銀色匕首之謎 `text/TRANSLATION-GUIDE.md`：忠實正式、奇幻古典語氣；敘述對隊伍用「你們」；
  玩家須輸入的字寫成「中文（ENGLISH）」；英文雙關謎語附英文原句與答案；物品名加值置後（例「長弓 +3」）。
- 自動比對只產生待核對候選。異譯保留來源與選用理由，不改動前作檔案。

## 4. 參考來源與知識路由

每次開工、兩輪同類失敗、形成完成結論、建立文件或目錄前，核對
`/home/anr2/.codex/knowledge-base/knowledge-router.md`，只讀本次命中的入口。

| 來源 | 可沿用的方法 | 必須在本作重驗 |
|---|---|---|
| 銀色匕首之謎 `AGENTS.md`、`CONTEXT.md`、`WORKLOG.md`、`docs/spec/`、`docs/re/`、`tools/`、`text/` | 第一參考：G0–G11 流程與 Issue 切分、規格 001–021 的契約、收據腳本、OCR 與錯譯管線、TSV 與字型管線、手札與查驗題、攻略面板與地圖疊層、HD 批次（規格 010）、不露原版（規格 021）、demo 控制、CPU 速度、切換列字級；WORKLOG 的勘誤已轉成第 11 節規則 | 位址、bytes、座標、字串、callsite、重播輸入、素材數量、完成狀態 |
| 克萊恩英豪 `/home/anr2/cht/golden_box/champions_of_krynn/` | 規格 001–017 的原始版本；銀色匕首之謎未改動的部分以此為準 | 同上 |
| 幽靈騎士 `/home/anr2/cht/golden_box/death_knight_of_krynn/` | 規格 015 demo 控制、016 CPU 速度的原始版本；3D 牆面與 `SKYGRND.DAX` 戶外背景的 HD 經驗（本作也有 `WALLDEF1.DAX`、`SKYGRND.DAX`） | 同上 |
| 黑暗之后 `/home/anr2/cht/golden_box/dark_queen_of_krynn/` | dosgolem 在 mode 13h 冷啟動到角色選單、繪製交易與返回配對（其研究 003、005）；PaddleOCR 映像與模型快取（銀色匕首之謎 G4 唯讀借用） | 本作是否同為 13h、顯示驅動介面 |
| 光芒之池 `docs/spec/054-*`、`docs/reference/manual/` | 軟體世界說明書轉錄、跨冊譯名定案、原書誤譯註記 | 本作頁數、譯名與錯譯 |
| `/home/anr2/cht/golden_box/curse_of_the_azure_bonds/`、`golden-box-remake-engine/` | Forgotten Realms Gold Box 的 DAX／ECL 格式、TPOV overlay 逆向、術語 | 只當格式與術語參考，不採用其 remake 引擎 |
| `/home/anr2/cht/dosgolem/`（`README.md`、`CLAUDE.md`、`docs/spec/006-layering.md`） | 能力矩陣、`cmd/probe`、Docker 入口 `tools/go.sh` | 本作 binary 是否可執行與實際缺服務 |

**本作與前作的差異**（G2、G3 的結論，docs/re/003、004）：

1. 輸出路徑與前作同家族，最接近幽靈騎士（`FBOV`、版本同為 1.10、路徑 A 參數與路徑 C 呼叫端同形）。
   `GAME.EXE` 是 EXEPACK，檔案位移不等於執行期位移；字串與位址以 dosgolem 解開後的映像為準（`tools/g3-game-image.sh`）。
   resident 段 `1B2E`：逐字 `01DF`、路徑 A `0354`、路徑 B `0498`、路徑 D `030B`。
2. mode 13h（320×200、256 色）。文字格換算是每列 320 bytes；主選單文字的色號就是路徑 A 的前景、底色參數。
   前作在 0Dh 驗證的 `dax_pics.py` 解碼、HD 色號比對與 `hd_validate.py` 閘門、規格 021 以「兩種以上色號」定義的墨跡格，
   在本作仍是假說，G6、G9 重驗。
3. 片頭片尾由另一支程式 `VACE.EXE` 播放，字幕推定是畫格裡的點陣圖（強推論），不經任何印字程序；
   以畫面辨識覆繪處理，歸 G6／G9。

| 觸發 | 入口 |
|---|---|
| 老遊戲逆向與中文化規則路由 | skill `re-retro-cht-rulebook` |
| 中文高解析畫布 | `~/.claude/rulebook/81-retro-cjk-hires-canvas.md` |
| Gold Box `.OVR`／TPOV | `~/.claude/knowledge-base/retro/borland-tpov-overlay-re.md` |
| IDA Pro 9.4 | `~/.claude/knowledge-base/retro/ida-pro-9.4.md`、`~/.codex/skills/use-ida-pro-9-4/SKILL.md` |
| 翻譯與顯示隔離 | `~/.codex/knowledge-base/local/localization-display-semantic-isolation.md` |
| 中文版面與 Gold Box 畫面比較 | `~/.codex/skills/research-pc98-golden-box-ui/SKILL.md` |
| 建立文件或新目錄 | `~/.codex/knowledge-base/local/project-document-responsibilities.md` |
| 授權 | `~/.codex/knowledge-base/local/rrsal-1.0.md`（RRSAL-1.0，已定案） |
| 正式版號與交付目錄 | `~/.codex/knowledge-base/local/retro-remake-release-versioning.md`、`retro-remake-dist-all-output.md` |

## 5. dosgolem 分層

- 不另建 remake 引擎。執行、畫面、輸入、虛擬時間、中文覆繪與 HD 呈現都經 dosgolem。
- 本作位址、畫面狀態、輸出攔截點、重播與 HD 素材對應屬遊戲專屬層，放在 dosgolem
  `apps/pod/`。通用 CPU、DOS、BIOS、VGA、輸入與音訊屬通用層，不寫入本作名稱或位址。
  判準見 dosgolem `docs/spec/006-layering.md`。VGA 缺口屬通用層，修在通用層。
- 研究與開發在 `workplace/dosgolem/` 的隔離 clone，分支 `pools-of-darkness-cht`，推到本專案
  private repo 的 `dosgolem/pod` 分支保存。本專案 `main` 只放文件、譯文、字型來源與工具。
- 不改動 `/home/anr2/cht/dosgolem` 工作樹，不推送 dosgolem 遠端，不推送前作儲存庫。
- 基底（G2 選定，docs/re/003）：銀色匕首之謎 repo 的 `dosgolem/sosb` `a6e0254`（2026-10-23）。本作分支另加
  dosgolem 規格 `430-mz-minalloc-reservation` 與 `431-root-tsr-shrink-option` 兩項通用層修正。
  `apps/sosb/` 只當參考實作，其中位址、座標與字串在本作全部視為假說。從前作分支帶來、
  實屬通用層的程式留在通用層。
- 入口照 `START.BAT`：`STARTUP.EXE` 設定、常駐 `snd.com`、監督佇列接 `CONTROL.EXE`，並帶 `-root-tsr-shrink`
  （`tools/g2-startup-receipt.sh`）。不常駐驅動時 DEMO 會跳進顯存當掉；沒有 `-root-tsr-shrink` 時遊戲報記憶體不足。
  `apps/pod` 的 play 與收據指令也要照這條鏈走並設 `DOS.RootTSRShrink`；oracle 缺對應入口，先補通用層規格（dosgolem 431 §5）。
- Go 不允許外部 module 引用 dosgolem `internal/`，adapter 必須在 dosgolem module 樹內。
- dosgolem 是原版行為、畫面與虛擬時間的權威執行器，正式收據由它產生。DOSBox／DOSBox-X
  只作診斷與交叉基準，擷取明標輔助。

## 6. 繁中與 HD 契約

- 原版 EXE、資料、規則、亂數、條件比較、檔名、腳本鍵值與存檔保持原樣。本專案不翻譯 EXE。
  規格 021 允許的記憶體內修改要附位址、原位元組與收據，F12 原版模式維持原樣。
- 英文原文、呼叫位置與畫面座標用於辨識輸出。譯文只影響顯示，不進入比較、查找、序列化、
  檔名或存檔。共用字串函式須盤點顯示與語意呼叫者。
- 原版要求查閱手冊時，辨識該題的輸出事件再覆繪中文；不繞過驗證、不改判定。
- 玩家可見譯文以 UTF-8 TSV 為唯一正式來源；catalog 檢查 key 唯一、欄數、控制碼、placeholder、
  漏譯、孤兒 key、行寬與字型覆蓋。字型子集由正式譯文重建，不為缺字改譯。
- 術語表、手札、攻略面板與地圖點的譯名一致性用檢查腳本驗證（比照銀色匕首之謎
  `tools/term_check.py`），不一致時結束碼非 0。
- 中文字依 `rulebook/81`：拉高輸出畫布，原版底圖整數倍最近鄰放大，中文用正常尺寸字型。
- 未辨識、歧義或失配的輸出保留原版並記錄缺口。

### HD：流程與 Codex 的工作介面

流程照銀色匕首之謎規格 010 §3；工具自其 `tools/hd_*.py`、`dax_pics.py` 移植，解碼與色盤依本作重驗。

| 步驟 | 負責 | 產出 |
|---|---|---|
| 來源辨識 | Claude | 本作繪圖來源、繪製時機、遮擋與動畫相位證據；DRAFT 規格 |
| 盤點 | Claude | `workplace/hd/inventory/inventory.json`：以寬、高、色號 SHA-256 去重；動畫以第 0 幀為底 |
| 請求 | Claude | `workplace/hd/requests/<批次>/`：1 倍與 3 倍參考 PNG、色號索引、動畫遮罩、`manifest.json`（來源檔、offset、SHA-256、尺寸、色盤、透明色、規格編號、提示詞） |
| 生圖 | Codex | `workplace/hd/out/<批次>/`：依原版手繪重畫的 PNG 與 `manifest.json`（來源 key、輸出尺寸、風格批次、生成方式、權利註記） |
| 前處理與驗收 | Claude | 3 倍、對位、alpha、遮罩外差異、結構與色相閘門（閾值沿用銀色匕首之謎規格 010 §3，本作重驗）；並列比較頁綁定 PNG SHA-256，未目視或雜湊改變不得接入 |
| 接入 | Claude | `workplace/hd/theme/` 主題；dosgolem 呈現層的圖面 watcher（`Sticky`、`Hold`）；`review/hd/` 審閱頁 |

- 每個批次先開 Issue，寫明請求目錄、規格、風格定案、完成條件，以及 Codex 可否 commit、push。
  Issue 沒寫明時不 commit、不 push。
- Claude 以主機 `codex exec` 委派，這是繼 `gh` 之後第二個主機工具例外。前作記錄的旗標為
  `-C`、`-s workspace-write`、`-o`、`--add-dir`；首次使用前以 `codex exec --help` 重新核對。
  prompt 明寫可讀與可寫的目錄。
- Codex 只做生圖，只改 Issue 指定的目錄與檔案，不改其他規格或 `CONTEXT.md` 的有效決策。轉檔、縮放、
  建置與批次腳本一律在 Docker 內。輸出無法符合請求 manifest 時在 Issue 說明，不自行改規格。
- Codex 產出的素材視為可放進原版＋HD patch，權利分類仍記在 manifest 與 Issue。
- 與前作逐位元組相同的來源可沿用前作已通過的 HD 圖，manifest 記來源專案與 key。
- HD 只改呈現。來源、繪製時機、遮擋、清除與動畫依本作證據辨識；失配、撤除或遮擋未知時
  回退原版。切換不得重開遊戲、改亂數或重寫存檔。

## 7. 規格與驗證閘門

正式接入依序走：`原版證據 → DRAFT → 證據審查 → READY → 實作 → 同狀態驗證 → CONFORMED`。
未達 READY 的探針與原型不進產品路徑。`docs/spec/000-index.md` 列出每個規格在本作的狀態、差異與證據，
前作的 CONFORMED 不代表本作已驗。

- 每項研究記錄輸入檔名與 SHA-256、工具版本、dosgolem 位址空間、callsite、暫存器、字串位址、
  字型、座標、顏色、顯示模式與清除時機。證據等級用「已證實、強推論、假說、未知」。反組譯名稱只供導覽。
  引用前作結論註明出處，等級最高為「假說」，直到本作重驗。
- 原文／繁中、原版／HD 比較固定資料、初始狀態、輸入與 seed，差異只能出現在批准的呈現範圍。
  另走正常玩家路徑：冷啟動、建角或載入、探索、對話、選單、戰鬥、存讀檔，驗證捲動、清除、
  返回與殘字。快照與狀態注入只供診斷。驗證抽樣即可，抽樣範圍寫進收據。
- HD 另驗原版記憶體、暫存器、畫面緩衝與存檔不受影響，並用失配來源與撤除素材建立負對照。
- 不露原版以閃現量尺（銀色匕首之謎規格 021 §3，`-flash-audit`）量測，抽樣路線閃現 0 才算達成。
- 只有被 dosgolem 同狀態收據覆蓋的路徑，才能稱為「已中文化」；其餘標示未知。

## 8. Docker、Git 與主機衛生

- 分析、解壓、搜尋、OCR、下載、建置、測試、執行遊戲與 GUI 自動化只在 Docker 內進行。使用
  `--rm`、`--memory`、`--cpus`、`--pids-limit`、外層逾時、`-u "$(id -u):$(id -g)"`、
  `--log-opt max-size=10m --log-opt max-file=3`，預設 `--network none`。
- 掛載前確認 host 來源存在且型態正確，要寫入的目錄先 `mkdir -p`。每批工作後在容器內執行
  `find <工作根> -user root -o -type d -name '*.md'`。不碰其他專案的 image、volume 與 container，
  禁止任何 prune 與 `rmi`。前作建好的 image 可唯讀沿用，名稱與 ID 記入 `WORKLOG.md`。
- 主機工具例外只有 `/home/anr2/.local/bin/gh` 與 `codex`。第一次使用 gh 先執行 `gh auth status`。
- `workplace/` 是唯一可寫研究工作區，整個目錄 gitignore。不另建同義目錄。
- 前作、dosgolem 原始工作樹與 `~/.claude/`、`~/.codex/` 只讀。
- `LICENSE` 依 RRSAL-1.0 範本在第一個 commit 前就位。正式交付根為 `dist-all/<版本>/`。
- 本機目錄名 `pool-of-darkness` 是單數，與官方名 Pools of Darkness、repo `pools-of-darkness-cht` 不一致，
  不改名；三者指同一專案。

## 9. 文件職責

以下除本檔外按需建立。新增文件時同步更新此表或子目錄索引。

| 路徑 | 職責 |
|---|---|
| `AGENTS.md` | 協作規則與專案路由，即本檔 |
| `CLAUDE.md` | 指回本檔的薄入口、常用指令與資料索引 |
| `CONTEXT.md` | 有日期的現況、有效決策、glossary、dosgolem 基線與證據入口 |
| `WORKLOG.md` | 按時間追加的操作、驗證結果、勘誤與 Docker 清理狀態 |
| `README.md` | 用途、穩定入口與現況摘要 |
| `docs/re/` | 輸入雜湊、位址、實驗與證據等級；`docs/re/README.md` 為索引 |
| `docs/spec/` | DRAFT、READY、CONFORMED 規格，回鏈 Issue；`000-index.md` 為索引 |
| `text/` | 譯文 catalog、手札、`glossary.tsv`、`errata.tsv`、翻譯指引 |
| `font/` | 字型來源與字元清單 |
| `reference/` | 攻略與提示書轉錄（private，轉公開前移除） |
| `review/hd/` | HD 審閱頁（private，轉公開前移除） |
| `tools/` | 可重現的 Docker、盤點、驗證與打包入口 |
| `workplace/` | 解壓輸入、英文原文、轉錄、探針、IDA、dosgolem clone、HD 批次；不入 Git |
| `dist-all/` | 正式交付根；含原版產物不入 Git |

## 10. 共同決策與停止線

- 第 1 節與第 3 節已定案事項不重複詢問。
- HD 美術與操作取捨先展示並列原型，每次只問目前最早會阻塞的決定。
- 能由 bytes、程式、規格或測試查出的事實自行查證。實質改變規則、存檔格式、原版忠實度、
  版面、平台範圍或散布方式時先確認。

## 11. 沿用的前作教訓

下列規則由銀色匕首之謎 `WORKLOG.md` 的勘誤與操作紀錄轉成，日期是該專案的紀錄日；出處標「本作」的來自本專案。

| 規則 | 出處 |
|---|---|
| 首次執行設定的收據從乾淨 scratch 起跑，記錄設定檔是否已存在。設定檔已存在時，原本回答設定的按鍵會被標題畫面吃掉，路徑因此不同 | G2 勘誤，2026-10-22 |
| 移植前作簽章或進出點條件前，先在本作執行期記憶體看實際位元組 | demo 控制勘誤，2026-10-22 |
| 怪物譯名先查怪物資料檔（`MON1CHA.DAX` 等），不只看 ECL 文字 | G8 勘誤，2026-10-22 |
| 禁用詞替換不得作用在題名列；替換後另查題名 | G5 勘誤，2026-10-22 |
| 子代理可能以版權為由不做說明書逐字轉錄；改走 OCR（英文 `pdftotext -bbox-layout`，中文 PaddleOCR）再逐頁對圖校讀 | G4，2026-10-22 |
| 共用層的新機制先列出哪些呼叫端繞過它的狀態機 | 規格 021 勘誤，2026-10-23 |
| 背景執行中的腳本不改；bash 邊讀邊執行，改了會讀到半行 | 規格 021 勘誤，2026-10-23 |
| 等待背景工作用通知，不用 `pgrep -f` 迴圈，它會比對到自己 | G5，2026-10-22 |
| 收據的高倍率切換放在最後幾格；全程 32x 會超過逾時 | 切換列收據勘誤，2026-10-22 |
| demo 中止後回到選單的按鍵數依本作畫面序列重測 | demo 收據，2026-10-22 |
| ffmpeg 在 `--pids-limit 32` 下建立執行緒失敗，錯誤顯示成濾鏡設定錯誤；用 `-threads 1`，上限放寬到 128 | G9，2026-10-23 |
| 本作的圖可能與前作逐位元組相同；HD 盤點先比對前作主題，相同的沿用，不再生圖 | G9，2026-10-23 |
| 啟動批次檔的每一步都要在 dosgolem 重現，常駐程式不能省；省掉的那支可能在很後面才讓遊戲當掉（本作 DEMO 跳進顯存） | 本作 G2，2026-10-23 |
| dosgolem 的通用行為若有規格與收據釘住（例 197 §5.1），改它用預設關閉的選項，不改預設 | 本作 G2，2026-10-23 |
| 壓縮過的執行檔（EXEPACK 等）不以檔案位移推位址或抽字串；用 dosgolem 跑原版解壓後的記憶體 | 本作 G3，2026-10-23 |
| `cmd/probe -call-args` 一次只接受一組 `CS:IP`，多組用逗號時只有第一組生效；分次跑 | 本作 G3，2026-10-23 |
