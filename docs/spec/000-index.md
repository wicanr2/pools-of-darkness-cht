# 規格索引

- [024 片頭字幕覆繪](024-intro-subtitles.md)：#10，八組台詞、雙語／倍率／原版切換及正常5100格閃現0，局部CONFORMED。
- [023 左右海報](023-side-posters.md)：#21，固定加寬與語系文字，繁中／英文及 2／3 倍 CONFORMED。

狀態：DRAFT → 證據審查 → READY → 實作 → 同狀態驗證 → CONFORMED（`AGENTS.md` 第 7 節）。
規格 001–017 的契約沿用克萊恩英豪（`/home/anr2/cht/golden_box/champions_of_krynn/docs/spec/`），018–022 沿用銀色匕首之謎
（`/home/anr2/cht/golden_box/secret-of-the-sliver-blader/docs/spec/`）。adapter 是由 `apps/sosb` 複製的 `apps/pod`，play 指令是由
`cmd/sosb-play` 複製的 `cmd/pod-play`。本表的狀態只算本作已有的證據，前作的 CONFORMED 不代表本作已驗。
本作差異寫在「差異」欄，細節在對應的 `docs/re/`。通用層的規格在 dosgolem（430、431、432，皆 CONFORMED，docs/re/003）。

| 規格 | 本作狀態 | Issue | 差異 | 證據 |
|---|---|---|---|---|
| 001 路徑 A：Pascal 字串覆繪 | CONFORMED | #4、#7 | resident 段 `1B2E:0354`；EXEPACK，位址以解開後映像為準；mode 13h 每列 320 bytes | docs/re/004；`tools/g3-overlay-receipt.sh`、`g5-text-receipt.sh`、`play-smoke.sh`、`g6-create-receipt.sh`；docs/re/010 探索兩輪 |
| 002 路徑 B：ECL 文字視窗覆繪 | CONFORMED | #4、#7 | `1B2E:0498`，9 個 word；DEMO 與遊戲內碼頭對白參數相同（1、0、14、22、38、17、1，推定為旗標、底色、前景、下、右、上、左）；游標 `DS:9D92／9D93`（強推論） | docs/re/004；`tools/g5-text-receipt.sh`、`g6-create-receipt.sh`；docs/re/010 探索兩輪 |
| 003 路徑 C：底列水平選單 | CONFORMED | #4、#7 | 逐字程序 `1B2E:01DF`；四種呼叫端（selected `E9 94 00`、hotkey `EB 3B`、normal `FE 46 FF`、pad `C4 7E FC…`）在主選單與建角畫面相同；熱鍵字母接在譯文後 | docs/re/004；`tools/g6-create-receipt.sh`；docs/re/010 探索兩輪 |
| 004 組句訊息 | DRAFT | #7 | 模板沿用前作 188 條，另加本作建角的存檔、覆蓋確認；DEMO 戰鬥的物品名＋加值（釘頭錘 +3）、施法訊息已命中 | `tools/play-smoke.sh`（DEMO 10,500 畫格） |
| 005 中文句中的半形英數 | DRAFT | #7 | 數字照原版保留；未單獨驗證 | — |
| [006 手冊查驗題](006-quiz-rendering.md) | CONFORMED（列替換抽樣通過） | #8、#16、#30 | 題型「journal, page (P) / what is word # W / after the heading / 標題」，路徑 A 六段（第 2、4、6、8 列）；題目表 GAME.EXE 映像 0x9F32 起 21 筆、54 題，答案加密存在表內；答案表鍵改為頁、字序、標題（`text/quiz-answers.tsv`：`page`、`word`、`heading`、`answer`）；改寫第 2 列「英文手札第 P 頁」（延伸到頁碼結尾 27 格）、第 4 列「下列標題之後的第 W 個字？」（延伸到字序結尾 22 格）、第 6 列「答案：…」或空白，第 8 列標題留英文；前端文字鍵 `quiz.page`、`quiz.word`、`quiz.answer`；沒有酒店傳聞分支 | docs/re/008；`tools/g7-quiz-receipt.sh`、`g6-create-receipt.sh`、`g6-saveload-receipt.sh` |
| 007 手札面板 | READY | #8 | 99 則珍157 修正譯文（`text/journal.zh-TW.tsv`），面板排版 3 頁以內；引用句由引擎組成整句「This is recorded as journal entry N.」經路徑 B 印出（`GAME.OVR` 0x2643），`JournalRefOf` 與組句模板有單元測試；實機在叛徒門觸發第 57 則，整句一次印在第 17 列第 1 欄，面板自動開出（`tools/g7-journal-receipt.sh`）；F5 翻閱以 `-journal-preview` 截圖（同一腳本的 `browse` 張） | docs/re/005、009；`tools/g7-journal-receipt.sh` |
| 008 攻略面板 | READY | #9 | 內容是讀《軟體世界》第 43–45 期後自寫的摘要（雜誌原文不照抄），78 段，頁尾「整理自軟體世界 N 期 p.X–Y」；面板程式由前作沿用；主選單開面板三張截圖無缺字；冒險畫面開關與手札互斥未驗 | docs/re/011；`tools/g8-guide-receipt.sh` |
| [009 地圖疊層](009-map-overlay.md) | READY | #9 | 本作 GEO 指標 `DS:A9EA`、座標方向 `DS:A9F3–A9F5`；32 塊 GEO、24 個 GEO 名稱、東西菲蘭 5 點共 7 格。兩城 2／3 倍 A/B 通過；已有正常冷開機讀檔、地圖／攻略互斥收據。其餘名稱、點及共用場景未驗 | docs/re/012、013；`tools/g8-map-evidence.sh`、`g8-map-receipt.sh [east\|west]`、`g8-map-play-receipt.sh` |
| [010 HD 主題](010-hd-theme.md) | 局部 CONFORMED | #10、#18–#20、#23、#24、#26–#44 | 菲蘭起點121格街景、VGA二百三十八張怪物候選／兩張地形、七張BIGPIC及三張開場語系圖接入通過；完整244個CPIC來源狀態均有素材。BIGPIC244正常DEMO、其餘六張條件式來源通過；新建置完整四路狀態一致、閃現0。未觀測遇怪事件仍未知；PIC81狀態／78項主題局部READY；其他未驗來源仍DRAFT | docs/re/014、017；014-wall-runtime.json、014-title-runtime.json、017-vga-battle-runtime.json、017-b08-runtime.json、017-b09-runtime.json、017-b12-runtime.json、017-b13-runtime.json、017-b14-runtime.json、017-b15-runtime.json、017-b16-runtime.json、017-b17-runtime.json、017-b20-runtime.json、017-b21-runtime.json、[017-b24-runtime.json](../re/017-b24-runtime.json)、[017-cpic-coverage.json](../re/017-cpic-coverage.json)、[014-bigpic-runtime.json](../re/014-bigpic-runtime.json) |
| 011 切換列 | 未驗 | #10 | 程式由前作沿用 | — |
| 012 發行包 | 未開始 | #12 | — | — |
| 013 玩家名音譯 | DRAFT | #7 | 音譯資料沿用前作；隊伍名單的玩家名已音譯（ROLAND → 羅倫德）；建角畫面的名字與性別欄位址未在本作驗證 | `tools/g6-create-receipt.sh` |
| 014 配樂 | 未開始 | #11 | — | — |
| [015 多語系](015-multilingual.md) | 英文前端局部CONFORMED | #11、#26 | 本作問句格式、英文專用字型、開場F8及2210格英文對話抽樣通過；英文手札、攻略及其他語系仍DRAFT | docs/re/014；014-title-runtime.json、014-english-front-runtime.json |
| 016 推廣影片 | 未開始 | #11 | — | — |
| 017 作弊：我方無敵 | 停用 | #11 | 位址是前作值，`CheatVerified` 為假 | — |
| 018 開場 demo 控制 | 未開始 | #13 | `GAME.EXE` 有 `Play Demo` | docs/re/001 |
| 019 CPU 模擬速度 | 未驗 | #14 | 程式由前作沿用；本作 CRT `Delay` 寫法未驗 | — |
| 020 切換列字級 | 未驗 | #15 | 程式由前作沿用 | — |
| 021 不露原版 | 局部已驗 | #16 | 完整PIC5900格、VGA首批戰鬥7400格及開場三圖1250格正常冷啟動閃現0；不外推其他圖族與未走路徑 | docs/re/014、017 |
| 022 手札地圖頁 | 未開始 | #8 | 地圖則 5、23、45、98，第 36 則附地圖 | docs/re/005 |

升到 CONFORMED 的條件：本作正常玩家路徑（建角、存讀檔）的同狀態收據覆蓋該規格的行為。
