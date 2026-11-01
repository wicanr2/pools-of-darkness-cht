# 規格索引

- [010 片頭BMA動畫接入](010-hd-theme.md#bma片頭動畫接入)：#74，97種敘事相位、原版133次來源呼叫與字幕共存。正常5100回合、切換、原版狀態及127檔通過；原量尺全黑與明示切換保留，非預期可見閃現0。[來源與收據031](../re/031-intro-bma.md)，SSI標誌維持原版，片尾另計。

- [012 三平台候選封包](012-package.md)：#12、#69、#75，舊九包及新增三平台HD patch本機候選範圍CONFORMED；原案完整交付尚未完成，不建立公開Release。

- [010 戰鬥透明特效與合成來源](010-hd-theme.md#戰鬥透明特效與合成來源)：#73抽樣CONFORMED；同輸入10500格DEMO閃現19→0，原版狀態及127檔相同。[030勘誤與收據](../re/030-hd-occlusion.md)取代029錯一格的鏡頭分類，未觀測的完整戰鬥動畫及其他事件仍未知。

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
| [010 HD 主題](010-hd-theme.md) | 局部 CONFORMED | #10、#18–#20、#23、#24、#26–#44、#48、#50、#51、#52、#53、#54、#56、#57、#58、#59、#60、#61、#62、#63、#64、#65、#66、#67、#68、#74 | 菲蘭起點121格街景、VGA二百三十八張怪物候選／DUNGCOM六組150張非純色圖樣與六個純色保留格、野外三組119個來源去重為116張、七張BIGPIC及三張開場語系圖接入通過；完整244個CPIC來源狀態均有素材。BIGPIC244正常DEMO、其餘六張條件式來源通過；完整四路與後續資料增量通過；b27成對與後續增量正常DEMO通過，COMSPR全44來源對應36候選，六類抽驗與009正常特效通過，歧義回退；另有九張264×264背景完整來源與使用RGB守衛，六類抽驗及正常DEMO通過；遭遇圖29組87張素材齊全，新增60張六類抽驗與既有十四張正常畫面不變通過；現行1866檔，既有主題抽樣維持；片頭量尺分類見031。041近遠景前景／透明區已驗，其他正常事件未知。片頭97種敘事HD相位與133次來源呼叫接入抽樣CONFORMED，正常5100格與切換、原版狀態及127檔通過；原量尺全黑與明示切換保留，見031。SSI標誌維持原版，片尾另計。完整背景正常事件及遮擋未知。未觀測遇怪事件仍未知；PIC81狀態／78項主題局部READY；其他未驗來源仍DRAFT | docs/re/014、017；014-wall-runtime.json、014-title-runtime.json、017-vga-battle-runtime.json、017-b08-runtime.json、017-b09-runtime.json、017-b12-runtime.json、017-b13-runtime.json、017-b14-runtime.json、017-b15-runtime.json、017-b16-runtime.json、017-b17-runtime.json、017-b20-runtime.json、017-b21-runtime.json、[017-b24-runtime.json](../re/017-b24-runtime.json)、[017-b26-runtime.json](../re/017-b26-runtime.json)、[017-b27-runtime.json](../re/017-b27-runtime.json)、[017-b28-runtime.json](../re/017-b28-runtime.json)、[017-b29-runtime.json](../re/017-b29-runtime.json)、[017-b30-runtime.json](../re/017-b30-runtime.json)、[017-b31-runtime.json](../re/017-b31-runtime.json)、[017-b32-runtime.json](../re/017-b32-runtime.json)、[017-b33-runtime.json](../re/017-b33-runtime.json)、[017-b34-runtime.json](../re/017-b34-runtime.json)、[017-b35-runtime.json](../re/017-b35-runtime.json)、[017-b36-runtime.json](../re/017-b36-runtime.json)、[017-b37-runtime.json](../re/017-b37-runtime.json)、[017-b38-runtime.json](../re/017-b38-runtime.json)、[017-b39-runtime.json](../re/017-b39-runtime.json)、[017-b40-runtime.json](../re/017-b40-runtime.json)、[017-b41-runtime.json](../re/017-b41-runtime.json)、[017-b42-runtime.json](../re/017-b42-runtime.json)、[017-b43-runtime.json](../re/017-b43-runtime.json)、[018-b44-materials.json](../re/018-b44-materials.json)、[017-cpic-coverage.json](../re/017-cpic-coverage.json)、[014-bigpic-runtime.json](../re/014-bigpic-runtime.json) |
| 011 切換列 | 未驗 | #10 | 程式由前作沿用 | — |
| [012 發行包](012-package.md) | 候選範圍CONFORMED | #12、#69、#75 | 舊九包保持，新增v.0.1.1三平台HD patch與B版海報；Linux正常路徑與Wine抽驗，macOS只驗結構，原案完整交付未完 | docs/re/026-package-runtime.json、033-hd-patch-runtime.json |
| 013 玩家名音譯 | DRAFT | #7 | 音譯資料沿用前作；隊伍名單的玩家名已音譯（ROLAND → 羅倫德）；建角畫面的名字與性別欄位址未在本作驗證 | `tools/g6-create-receipt.sh` |
| 014 配樂 | 未開始 | #11 | — | — |
| [015 多語系](015-multilingual.md) | 五語抽樣CONFORMED | #11、#26、#70、#71、#76 | 日韓各13表7652列、99手札及78攻略，八份原生字型；三側正常4800格同原版狀態及132檔，F8五語／F12／偏好通過。日韓未經母語者審校，舊十二包不改 | [034](../re/034-jako-runtime.json)、028-english-panels-runtime.json、027-zhcn-runtime.json；014-title-runtime.json、014-english-front-runtime.json |
| 016 推廣影片 | 未開始 | #11 | — | — |
| [017 作弊：我方無敵](017-invincible.md) | CONFORMED，功能抽樣 | #11、#72 | 本作兩個HP、健康／陣營及五種有害效果；正常讀檔／DEMO、七個原版受控傷害及乾淨Git重編通過。長戰鬥19格已由030更正分類並在既有正常路線修正，其他事件不外推 | [研究029](../re/029-invincible.md)、[執行期收據](../re/029-invincible-runtime.json)、[遮擋030](../re/030-hd-occlusion.md) |
| [018 開場 demo 控制](018-demo-control.md) | CONFORMED | #13 | 選單不逾時；對話／戰鬥中斷、重播、正常進入、F12與2倍通過。已啟用a83043c | [固定收據](../re/020-demo-control-runtime.json)；`tools/demo-control-receipt.sh` |
| [019 CPU 模擬速度](019-cpu-speed.md) | CONFORMED，功能抽樣 | #14 | 正常2／3倍讀檔、六倍率、Draw回饋與正常關窗通過；1x狀態及偏好保持；使用者取消30fps要求 | [固定倍率](../re/021-toolbar-runtime.json)、[3倍HD](../re/022-speed-feedback-hd-runtime.json)、[2倍](../re/022-speed-feedback-2x-runtime.json) |
| [020 切換列字級](020-toolbar-font.md) | CONFORMED，正常前端抽樣 | #15 | 原生12／22像素字模；2／3倍三級、英文抽樣與偏好重啟通過，遊戲狀態不變 | [固定收據](../re/021-toolbar-runtime.json)；`tools/font-dev.sh` |
| 021 不露原版 | 局部已驗 | #16、#73 | 完整PIC5900格、VGA首批戰鬥7400格及開場三圖1250格正常冷啟動閃現0；既有長戰鬥10500格19→0，原版狀態及127檔相同。不外推其他圖族與未走路徑 | docs/re/014、017、[030](../re/030-hd-occlusion-runtime.json) |
| 022 手札地圖頁 | 未開始 | #8 | 地圖則 5、23、45、98，第 36 則附地圖 | docs/re/005 |

升到 CONFORMED 的條件：本作正常玩家路徑（建角、存讀檔）的同狀態收據覆蓋該規格的行為。

- [012 HD patch](012-package.md#hd-patch-conformed)：#75，B版海報、首次匯入與三平台本機候選CONFORMED。核心[032](../re/032-hd-patch-core.json)與實際封包[033](../re/033-hd-patch-runtime.json)分別保存；macOS真機未驗，原案其他功能未完。
- [015 日文與韓文](015-multilingual.md#日文與韓文收尾)：#76，正式接入與正常抽樣CONFORMED；[034收據](../re/034-jako-runtime.json)保存資料、字型、乾淨Git重建及原版同狀態證據。
- [010 片尾FINAL](010-hd-theme.md#b48-片尾final)：#77，播放器直接入口及既有正常DEMO回歸CONFORMED。58次來源／43種相位、三路2001格、F2／F8／F12及2034檔／18字型乾淨重建通過；[035](../re/035-final-hd.md)保存來源與驗收，正常通關入口仍未知。
