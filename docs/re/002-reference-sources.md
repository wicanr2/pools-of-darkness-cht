# 002 參考文件來源

日期：2026-10-23。Issue：[#2](https://github.com/wicanr2/pools-of-darkness-cht/issues/2)。
原檔與轉出物都在 `workplace/`，不進 Git。逐檔 SHA-256 見 [001-input-files.tsv](001-input-files.tsv)。

## 珍157 中文說明書

`珍157-黑暗之池.rar`（SHA-256 `3d18ec1b0d588705484f53f16921ed90af9343f3246af8b9f00f99677adcf623`）以 `unar -e cp950`
展開到 `workplace/input/manual-zh/珍157-黑暗之池/`。成員是 107 張 JPG、`Thumbs.db` 與 `軟體世界說明書補完計劃.txt`
（骨灰集散地「軟體世界說明書補完計劃」的說明，CP950）；沒有 PDF，沒有文字層。

JPG 檔名分 4 個掃描批次。內頁多為 1280×910–1053 的跨頁掃描，一張兩頁。頁序以縮圖拼版（`tools/contact_sheet.py`，
輸出 `workplace/manual-zh/sheets/`）逐張讀頁碼確認，兩冊頁碼連續，沒有缺頁。

| 批次（檔名前綴） | 張數 | 內容 | 檔號與頁碼 |
|---|---|---|---|
| `doc06143620140126130212` | 38 | 001 盒正面、002 盒背、003 盒側（「珍藏版 157」）、004 操作手冊封面、005 目錄、006–037 **操作手冊** p.1–63、038 操作手冊外封 | 檔號 n（6–37）＝ p.2n−11、2n−10；037 右頁空白 |
| `doc06143720140126133853` | 32 | 001 冒險者手冊封面、002 目錄、003–032 **冒險者手冊** p.1–60 | 檔號 n（3–32）＝ p.2n−5、2n−4 |
| `doc06144020140126140532` | 28 | 001–025 冒險者手冊 p.61–110、026 冒險者手冊外封、027–028 DISK 1、DISK 2 磁片套（1.2M） | 檔號 n（1–25）＝ p.2n+59、2n+60 |
| `doc06144120140126141051` | 9 | 《軟體總動員》（SOFTWARE NEWS）產品型錄插頁，不屬於說明書 | — |

| 冊 | 目錄（章節與起始頁） | 編輯 |
|---|---|---|
| 操作手冊 | 一、介紹 1；二、所需配備 2；三、備份磁片／載入硬碟 3；四、啟動遊戲 5；五、控制方法 7；六、其他的事項 10；七、進入遊戲 12；八、開始冒險 20；九、其他選單說明 33；附錄一、各種表格 44；附錄二、密碼表 62 | 文字編輯蔡佳機，美工編輯郭寶寶 |
| 冒險者手冊 | 一、前言 1；二、人物以及隊伍 2；三、如何建立一隻優秀的隊伍 12；四、戰鬥 13；五、法術 19；六、魔法寶物 21；七、被遺忘王國中的生物 23；八、各項法術的說明 37；九、旅行手札 55 | 文字編輯蔡佳機、胡任栓，美工編輯郭寶寶 |

- **旅行手札**：冒險者手冊 p.55–106，第 1–100 則，部分則附地圖（例第 5、23、45、98 則）。p.107 製作群，p.108–110 廣告。
- **密碼表**：操作手冊附錄二 p.62–63，欄位為頁數、字數、標題、答案，對應英文 Adventurer's Journal 的頁碼與英文標題。
  這是珍157 直接給出的查驗題答案表；表內容不轉錄進 Git。
- 操作手冊 p.19「密碼輸入」說明查驗題畫面：給頁碼與標題，問標題之後第幾個字（`GAME.OVR` 字串見 docs/re/001）。
- 操作手冊 p.10「六、其他的事項」有「傳送『銀色匕首之謎』中的人物」一節；英文 Rulebook p.1 對應「Transferring Characters from Secret of the Silver Blades」。

## 《軟體世界》攻略掃描

使用者提供，放在 `walkthrough/`（不進 Git）。標題「黑暗之池 完全攻略」，作者署名 Steeler。檔名數字比雜誌頁碼大 1。

| 期 | 檔案 | 雜誌頁碼 | 章節 |
|---|---|---|---|
| 第 43 期 | `61.jpeg`–`71.jpeg`（11 張） | 60–70 | 完全攻略（一）：第 1–11 章（重回菲蘭城到 Manshoon's Tower），文末「待續」 |
| 第 44 期 | `65 (1).jpeg`、`66.jpeg`–`76.jpeg`（12 張） | 64–75 | 完全攻略（二）：第 11 章（續）到第 19 章 Tower of Marcus，文末「下期待續」 |
| 第 45 期 | `67 (1).jpeg`、`68.jpeg`–`74.jpeg`（8 張） | 66–73 | 完全攻略（下）：第 20 章（Moander）到第 24 章 Palace of Gothmenes，文末「The End」 |

三期章節連續，第 1–24 章都在。各期頁內段落與地圖的對應在 G8 轉錄時逐頁確認。

## SSI 英文文件（mocagh.org）

2026-10-23（UTC 02:14）以 `curl` 在 Docker 內下載到 `workplace/input/en-docs/mocagh/`，網址 `https://www.mocagh.org/ssi/<檔名>`，
遊戲頁 `https://www.mocagh.org/loadpage.php?getgame=pod`（頁面 HTML 存於同目錄 `html/`）。全部是掃描加 Acrobat Paper Capture OCR；
`pdftotext -layout` 輸出在 `workplace/reference-src/txt/`。

| 檔案 | SHA-256 | 頁數 | 內容與用途 |
|---|---|---|---|
| `pod-journal.pdf` | `dc059314dee0244733361238f6cdd165826a3fc78c2d62e827b4d7fcb4189711` | 33 | Adventurer's Journal：前半是規則（Characters and Parties、Combat、Magic、Creatures、Spell Descriptions），p.25 起 Journal Entries 第 1–100 則，另有 Tavern Tale 與地圖。錯譯對照、查驗題答案 |
| `pod-manual.pdf` | `5e156eb8349b8eb1e7eebe2495520e36881c91f3b03695c8a866896e67be6504` | 9 | Rulebook：安裝、選單、移轉銀色匕首之謎角色 |
| `pod-refcard.pdf` | `88e6bcf6c170c8003b8920b1a02812325ad7b6482ee7c393d0e9a3a57fde265f` | 4 | IBM 資料卡 |
| `pod-hintbook.pdf` | `1e63d7fa7bf7ad10460135a15642f621a5879bc452d2a428d7f7d20eabdf14f2` | 41 | 官方提示書：攻略面板缺來源時的補充 |

- 遊戲頁另列 `pod-contest.pdf`（比賽活動），與手札和攻略無關，未下載。
- OCR 以 `JOURNAL ENTRY n` 檢索，第 13、38 則沒有命中，推定是辨識錯誤；英文手札版面多欄，轉錄時要對照頁面圖。
- 珍157 的手札與冒險者手冊合併成一冊，英文版則把規則與手札合在 Adventurer's Journal、安裝與選單放在 Rulebook；兩邊的頁碼不對應。
