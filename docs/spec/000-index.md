# 規格索引

- [012 三平台候選封包](012-package.md)：#12、#69，本機九包候選範圍CONFORMED；正式原案功能與完整交付尚未完成，不建立公開Release。

- [025 直接攔截繪圖與前端命令](025-direct-render-commands.md#draw-closure-conformed)：#45–#47及#49已支援VGA核心收尾CONFORMED。[最終收據](../re/023-draw-closure-runtime.json)記錄六條正常路線15,073,225 bytes全由Go演算法計畫寫入，固定／原x86像素寫入0，正常狀態／畫面／存檔相同；補上填色搜尋與讀點短續行，完整回歸通過。原版內容、文字、呼叫編排、BIOS／DOS及未知條件依契約保留。各批歷史DRAFT與計數由最新收尾回鏈統一解釋。

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
| [010 HD 主題](010-hd-theme.md) | 局部 CONFORMED | #10、#18–#20、#23、#24、#26–#44、#48、#50、#51、#52、#53、#54、#56、#57、#58、#59、#60、#61、#62、#63、#64、#65、#66、#67、#68 | 菲蘭起點121格街景、VGA二百三十八張怪物候選／DUNGCOM六組150張非純色圖樣與六個純色保留格、野外三組119個來源去重為116張、七張BIGPIC及三張開場語系圖接入通過；完整244個CPIC來源狀態均有素材。BIGPIC244正常DEMO、其餘六張條件式來源通過；完整四路與後續資料增量通過；b27成對與後續增量正常DEMO通過，COMSPR全44來源對應36候選，六類抽驗與009正常特效通過，歧義回退；另有九張264×264背景完整來源與使用RGB守衛，六類抽驗及正常DEMO通過；遭遇圖29組87張素材齊全，新增60張六類抽驗與既有十四張正常畫面不變通過；1525檔已啟用，閃現0。041近遠景前景／透明區已驗，其他正常事件未知。六張片頭關鍵畫面素材通過，完整動畫接入DRAFT。完整背景正常事件及遮擋未知。未觀測遇怪事件仍未知；PIC81狀態／78項主題局部READY；其他未驗來源仍DRAFT | docs/re/014、017；014-wall-runtime.json、014-title-runtime.json、017-vga-battle-runtime.json、017-b08-runtime.json、017-b09-runtime.json、017-b12-runtime.json、017-b13-runtime.json、017-b14-runtime.json、017-b15-runtime.json、017-b16-runtime.json、017-b17-runtime.json、017-b20-runtime.json、017-b21-runtime.json、[017-b24-runtime.json](../re/017-b24-runtime.json)、[017-b26-runtime.json](../re/017-b26-runtime.json)、[017-b27-runtime.json](../re/017-b27-runtime.json)、[017-b28-runtime.json](../re/017-b28-runtime.json)、[017-b29-runtime.json](../re/017-b29-runtime.json)、[017-b30-runtime.json](../re/017-b30-runtime.json)、[017-b31-runtime.json](../re/017-b31-runtime.json)、[017-b32-runtime.json](../re/017-b32-runtime.json)、[017-b33-runtime.json](../re/017-b33-runtime.json)、[017-b34-runtime.json](../re/017-b34-runtime.json)、[017-b35-runtime.json](../re/017-b35-runtime.json)、[017-b36-runtime.json](../re/017-b36-runtime.json)、[017-b37-runtime.json](../re/017-b37-runtime.json)、[017-b38-runtime.json](../re/017-b38-runtime.json)、[017-b39-runtime.json](../re/017-b39-runtime.json)、[017-b40-runtime.json](../re/017-b40-runtime.json)、[017-b41-runtime.json](../re/017-b41-runtime.json)、[017-b42-runtime.json](../re/017-b42-runtime.json)、[017-b43-runtime.json](../re/017-b43-runtime.json)、[018-b44-materials.json](../re/018-b44-materials.json)、[017-cpic-coverage.json](../re/017-cpic-coverage.json)、[014-bigpic-runtime.json](../re/014-bigpic-runtime.json) |
| 011 切換列 | 未驗 | #10 | 程式由前作沿用 | — |
| [012 發行包](012-package.md) | 候選範圍CONFORMED | #12、#69 | 基本patch三包、本機完整版六包；Linux與Wine抽驗，macOS只驗結構，HD patch及原案完整交付未完 | docs/re/026-package-runtime.json |
| 013 玩家名音譯 | DRAFT | #7 | 音譯資料沿用前作；隊伍名單的玩家名已音譯（ROLAND → 羅倫德）；建角畫面的名字與性別欄位址未在本作驗證 | `tools/g6-create-receipt.sh` |
| 014 配樂 | 未開始 | #11 | — | — |
| [015 多語系](015-multilingual.md) | 英文前端局部CONFORMED | #11、#26 | 本作問句格式、英文專用字型、開場F8及2210格英文對話抽樣通過；英文手札、攻略及其他語系仍DRAFT | docs/re/014；014-title-runtime.json、014-english-front-runtime.json |
| 016 推廣影片 | 未開始 | #11 | — | — |
| 017 作弊：我方無敵 | 停用 | #11 | 位址是前作值，`CheatVerified` 為假 | — |
| [018 開場 demo 控制](018-demo-control.md) | CONFORMED | #13 | 選單不逾時；對話／戰鬥中斷、重播、正常進入、F12與2倍通過。已啟用a83043c | [固定收據](../re/020-demo-control-runtime.json)；`tools/demo-control-receipt.sh` |
| [019 CPU 模擬速度](019-cpu-speed.md) | CONFORMED，功能抽樣 | #14 | 正常2／3倍讀檔、六倍率、Draw回饋與正常關窗通過；1x狀態及偏好保持；使用者取消30fps要求 | [固定倍率](../re/021-toolbar-runtime.json)、[3倍HD](../re/022-speed-feedback-hd-runtime.json)、[2倍](../re/022-speed-feedback-2x-runtime.json) |
| [020 切換列字級](020-toolbar-font.md) | CONFORMED，正常前端抽樣 | #15 | 原生12／22像素字模；2／3倍三級、英文抽樣與偏好重啟通過，遊戲狀態不變 | [固定收據](../re/021-toolbar-runtime.json)；`tools/font-dev.sh` |
| 021 不露原版 | 局部已驗 | #16 | 完整PIC5900格、VGA首批戰鬥7400格及開場三圖1250格正常冷啟動閃現0；不外推其他圖族與未走路徑 | docs/re/014、017 |
| 022 手札地圖頁 | 未開始 | #8 | 地圖則 5、23、45、98，第 36 則附地圖 | docs/re/005 |

升到 CONFORMED 的條件：本作正常玩家路徑（建角、存讀檔）的同狀態收據覆蓋該規格的行為。
