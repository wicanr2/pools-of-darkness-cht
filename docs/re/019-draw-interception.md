# 019 直接繪圖攔截

日期：2026-10-25／26。Issue：[#45](https://github.com/wicanr2/pools-of-darkness-cht/issues/45)、[#46](https://github.com/wicanr2/pools-of-darkness-cht/issues/46)。狀態：像素出口、單次RGB及限定8×8字模整段CONFORMED。

## 目前狀態

| 項目 | 本輪基準 |
|---|---|
| 程式來源 | dosgolem/pod `64e734c92ecaa096be9bb91d6e6f792fd5fad7fd`；Go建置資訊為乾淨來源 |
| 已驗執行檔 | SHA-256 `d3f97dfb97ce3951da6dae880d39cbc5cb22e28532d38dc157a5ed59d93947f7` |
| 正式主題 | 730檔；既有BIGPIC、CPIC與正常四路收據不重開為未完成 |
| 新要求 | 使用者指定直接攔截繪圖，避免原版底圖與HD的重複繪製；參考mm3 |
| 參考限制 | 前輪mm3規格075為DRAFT；2026-10-26新版方法已有限定常式接管與事件預算，只沿方法，不沿其位址或完成範圍 |
| 研究工作區 | `workplace/probe/draw-intercept-v1/`；原始資料唯讀 |

## VGA驅動初步定位

`VGA.DRV`共30079 bytes，SHA-256 `eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f`。
檔頭由三bytes的`E9 rel16`組成跳轉表，第一項目標原檔offset`0x2786`；此為已證實bytes，尚未賦予繪圖語意。
IDA以binary segment匯入16位元程式，入口、函式原名、指令bytes與原始operands由[ida_pod_vga.py](../../tools/ida_pod_vga.py)匯出。
IDA位址與原檔file offset逐項標示，執行期`CS:IP`另以dosgolem定位，兩者未確認前不混用。

已安裝映像為`ida-pro-9.4-idapython:py312-v1`，ID `4ac62de83339c215bab10e455cee3a22d9c6efed9fd0d8ed0f068327b83a06ab`。
技能優先的locked-v1不在主機，沿用既有py312映像，先驗實際非root輸出與輸入SHA，不重建工具鏈。
原始驅動與正式資料唯讀，資料庫與匯出只寫本研究工作區；不輸出或搬運授權內容。

正式攔截依[規格025](../spec/025-direct-render-commands.md)，未達READY前只執行可丟棄的來源探針。

## 實際顯存寫入者

300M正常狀態中沒有VGA.DRV檔頭或多個函式指令錨點；原始檔中只有VACE.EXE含VGA.DRV字串，GAME.EXE／OVR未見此字串。
這表示不能把VGA.DRV跳轉表當成GAME的繪圖入口；檔案用途仍依實際呼叫確認，不由顯示模式13h推定。
改由[draw_trace_probe.go](../../tools/draw_trace_probe.go)沿既有7400格DEMO，統計實際VideoWrite的CS:IP與範圍。
首次寫入後的下一道指令前保存CPU、stack及唯讀記憶體副本，最多16份；它們不是函式入口或正式攔截點。

實際GAME段由[ida_pod_draw_slice.py](../../tools/ida_pod_draw_slice.py)匯出，記錄原快照SHA、slice SHA及執行期段值。
16位元slice的file offset等於該已記錄CS的IP，IDA EA另列，不能把VGA.DRV offset與GAME段offset互換。

7400格正常DEMO的十個寫入位置與數量固定於[019-draw-writers.json](019-draw-writers.json)，原版CPU、完整RAM、色號與色盤同既有基準。
GAME段`1ED4:1A5C`至`1A7E`為來源／目的指標及長度搬移，`1A7F`至`1A92`為填色；`1D1F:046E`至`0486`為清零，`1B2E:11EA／11F1`為字模像素寫入。
Move依SI／DI比較改變方向，不能只實作向前拷貝。實際像素出口均為MOVS／STOS，VACE亦有同類出口；原函式其他指令仍有旗標、返回及計時語意。
目前dosgolem把單一REP視為一個指令邊界，在指令內完成字串操作，計時與事件在Machine.Step邊界處理。
可在這個已存在的原子邊界換成型別化繪圖命令，保留外層原指令步數、週期及原函式其餘計算，無須一次空返回整個函式。

## 已驗接入

通用規格434的原生MOVS／STOS把像素搬移、填色與單點輸出轉為COPY／FILL／PIXEL命令。
每個命令自行持有來源色號資料，CPU與呈現後端不共用可變指標。原版索引背板保留全部寫入及讀回，
前端命令畫布在每次retrace另與背板逐位元組比對。其他指令、旗標、堆疊、返回及虛擬步數仍依原版執行。
未驗指令或會觸發監看回呼的範圍走原路徑，仍以PIXEL通知前端；中央主題含透明像素時回舊RGB合成。

通用規格435提供不可變的索引／色盤來源。最終RGB先畫HD、文字及保留區域，最後僅填alpha0的原版像素。
Hold／Latch記住當時的來源色盤，歷史畫面的原版像素不會誤用目前色盤。現行352張中央圖面全部不透明，左右海報另在視窗層合成。
只有量測閃現時另外建立原版RGB對照；一般呈現不先鋪完整原版RGB。

| 驗證 | 結果與收據 |
|---|---|
| 原生指令 | 三個CPU model、兩種週期模式、REP／非REP、方向、邊界、監看、payload隔離、回復寫入順序及計時事件通過；CPU／machine／session／xlate及前端測試通過 |
| 正常DEMO命令 | 7400更新、712250000步、8633次retrace；十個出口的3343240次寫入全部原生，回退0、畫布失配0；[完整狀態及事件](019-native-runtime.json) |
| 正常DEMO視窗 | HD、無HD、F12、2倍四路終點與原版相同，127個DOS檔案一致；BIGPIC244及戰鬥截圖同舊版本，HD量尺0；[四路收據](019-demo-runtime.json) |
| 讀檔、片頭與封面 | 三條既有正常路線的39對中央與完整視窗截圖一致；54／127／127個DOS檔案一致；[正常前端收據](019-ui-runtime.json) |
| 單次填色 | 全幅HD封面576000點，原版填色0點；BIGPIC選擇328320點、只填外圍247680點；[像素與成本](019-rgb-cost.json) |

39對畫面包含查驗題、街景、地圖、攻略、工具列、下拉選單、八組片頭字幕、兩種語言、F2及F12。
讀檔與片頭的原始閃現計數皆0。封面切換測試的原始計數998，全部是指定的第1219格2倍原版轉第1220格3倍HD；
新舊合成同樣記錄，逐筆確認25列998格均屬此操作，非預期事件0。量尺未分隔顯示倍率，不能把這份原始計數寫成0。

局部CPU合成量測採七組交錯、每組12次，封面／BIGPIC／戰鬥／街景的中位耗時分別由28.46／16.23／18.01／18.02ms
降為9.90／13.89／12.86／11.56ms。這次量測在其他有界驗收並行時進行，只描述這四個來源的合成成本，
不含遊戲執行、字型、面板、量尺、GPU及視窗，不能換算整體FPS。

## 檔案時間與證據邊界

補查十份完整RAM時，早期探針因重新寫入pool4.cfg，讓本機暫存檔時間不同。
五份GAME見證在`0x345C3`／`0x345C4`出現時間word差異；相鄰日期及Pascal字串POOL4.CFG對應DOS搜尋紀錄。
`internal/dos/find.go:dosDateTime`依檔案ModTime打包，實際檔案時間與word解碼完全吻合，屬探針輸入條件。
沒有對這兩個位址作遮蔽，也沒有改產品程式。

兩側暫存檔在啟動前統一為2000-01-01 UTC後重跑，十份完整RAM、暫存器及堆疊逐位元組一致。
固定日期的兩側完整終點RAM相同；日期不同的歷史RAM不宣稱相同。歷史輸入另以未固定日期的一組成對重播，
完整終點仍同原基準。原始記憶體、堆疊、存檔與提取圖皆留workplace，固定收據只保留計數、雜湊及定位。

## 重生入口與限制

- [draw_trace_probe.go](../../tools/draw_trace_probe.go)：在隔離clone的workplace/drawtrace建置；傳入原版目錄、G2設定、全新輸出目錄，分別用`-native=false/true -fixed-file-time -dos-audit`跑7400格。
- [direct_render_check.py](../../tools/direct_render_check.py)：核對原指令／命令的完整見證、音訊、DOS事件、檔案及歷史終點。資料契約以本檔與固定JSON為準。
- [direct-ui-receipt.sh](../../tools/direct-ui-receipt.sh)：`wall|movie|title direct|legacy <全新收據名稱>`，由Docker執行既有玩家路線，輸出至workplace/probe/g9-hd。
- [direct_ui_check.py](../../tools/direct_ui_check.py)：逐張核對來源、原版終點、遊戲檔案、畫面及主動切換事件。
- [direct_rgb_bench.go](../../tools/direct_rgb_bench.go)：在隔離clone的workplace/directbench建置，只量測既有真實圖面的CPU合成。

通用程式與規格保存於同repo的[dosgolem/pod分支](https://github.com/wicanr2/pools-of-darkness-cht/tree/dosgolem/pod)。
前輪像素分支保留全部函式指令；本輪成功接管的限定字模不再逐條模擬，其索引背板與所有已驗效果仍保留。
其他完整函式與GPU／SIMD後端未實作。
未走過的圖族、動畫HD與片尾不因本輪攔截通過而視為完成。

## 2026-10-26 整段字模接管證據

使用者指定mm3新版`docs/re/direct-render-methodology.md`。該作已完成限定圖片常式、相容背板、事件預算與原子效果提交，
與前輪讀到的DRAFT狀態不同。本作不沿其位址或完成率；現有像素命令與單次RGB驗收不重新開啟。

目前唯一新原始證據在`workplace/probe/draw-intercept-v1/whole-calls-v2/`。
沿既有7400更新DEMO、900／1800空白鍵，觀測字模`1B2E:11A6`共2930次，保存32個字碼的真正入口CPU及完整RAM。
完整函式93 bytes的SHA-256為`d1c978b914f28d1a978a2cae1b7a1c52f4b5d233176e8bd6c3029b05c9b4958b`；IDA位址空間仍為已記錄CS的IP，非EXE offset。
獨立原版CPU跑到真正返回，三CPU模型及兩種週期共192組，完整RAM、暫存器、旗標、返回、步數、週期及84次有序寫入全部一致。
原始輸入、RAM與存檔不加入Git。探針與模型為[draw_call_probe.go](../../tools/draw_call_probe.go)、[draw_call_oracle.go](../../tools/draw_call_oracle.go)。

原呼叫者是PUSH CS加近CALL，far return仍正確返回，故加速接點須涵蓋16位近／遠CALL。
不在每道指令增加資格查詢；主機依CALL後的完整入口與事件預算判定，失配回原版。
字模的指令及成本公式、來源／堆疊別名停止線，先依規格025達READY，再由通用規格436接入，已完成下列驗收。

### 整段接管與正常驗收

正式程式`64e734c`，Go1.24.13，原版資料、730檔HD主題與玩法未改。
CPU只在16位近遠CALL後通知，先依入口表篩選；適配器唯讀產生完整計畫，Owner先核對事件／回合／地圖取樣預算，再提交。
拒絕未知、簽章失配、別名、未支援監看及不足預算。印字盤點模式回原版，callback僅在Advance內存在。

| 驗證 | 結果 |
|---|---|
| 原版函式 | 32個真實入口×三CPU模型×兩週期192組；另16組畫面四角、全亮／全暗及字型段尾條件；完整CPU／RAM、返回、堆疊與有序寫入一致；空返回負對照確實失配 |
| 通用機制 | 近遠、直接／間接CALL，事件恰好／不足、回合配額、入口表值複製、模式／監看／邊界拒絕、prepare／入口panic及callback撤除通過 |
| 乾淨來源機器重播 | 7400更新、712250000虛擬步；2930次字模入口中接管2401次、529次回原路徑，以主機效果計入1528405道指令 |
| 全部有序事件 | 3343240次像素通知、8633次完整索引＋色盤回掃、79012筆音訊、1827984次DOS請求／返回的有序雜湊皆同原版；32份入口RAM與全部CPU／RAM及127個DOS檔案相同 |
| 正式DEMO | HD／無HD／F12／2倍四路、BIGPIC244與戰鬥獨立像素、舊版完整終點通過；每路實際接管2400次，已納入地圖取樣門檻；HD量尺0 |
| 正常讀檔UI | 既有槽A、查驗題、街景、地圖、攻略、工具列及下拉選單10對中央／完整視窗圖全同舊版；54個DOS檔案相同，量尺0；實際接管51472次、計入33740810道指令 |
| 真存檔往返 | 冷開機讀A、正常紮營選單另存原先不存在的B，再冷開機讀B至隊伍選單；6對畫面及兩側所有存檔bytes一致、原槽A未改，B讀回羅倫德、防護−4、生命力40 |

固定收據為[019-whole-call-runtime.json](019-whole-call-runtime.json)，含完整程式／工具與事件雜湊。
機器重播的輸入檔案日期固定2000-01-01 UTC，完整RAM同該組原版輸入；正式DEMO另對照既有日期輸入的原版完整終點，不混稱兩組RAM。

### 成對成本及重生

乾淨正式來源、同一Docker兩CPU配額及GOMAXPROCS=2，三個實際字模來源各七組交錯、每組512次完整CALL。
包含客體參數準備、入口篩選、簽章、計畫、背板及逐點／命令通知。三組中位數由40.36／33.88／42.47μs降為8.22／8.00／9.70μs，約4.2–4.9倍。
此量測關閉機器事件，不含Owner.BeforeStep、RGB、GPU、視窗與遊戲其他工作，不能換算全遊戲FPS。

- [draw_call_probe.go](../../tools/draw_call_probe.go)：隔離clone的workplace/callprobe建置；`-whole-call=false/true`沿既有7400更新正常DEMO產生原版及接管串流與入口，正式資料在whole-calls-v4-legacy／native。
- [draw_call_oracle.go](../../tools/draw_call_oracle.go)：可丟棄初始函式模型及獨立原版oracle；正式適配器另由apps/pod/glyph_call_test.go逐項核對，未以模型自測取代原版。
- [whole_call_bench.go](../../tools/whole_call_bench.go)：隔離clone的workplace/callbench建置，完整CALL成本，無固定FPS聲明。
- [whole-call-save-receipt.sh](../../tools/whole-call-save-receipt.sh)：`legacy|native`正常新存B／冷讀B；只寫全新輸出，不覆寫既有存檔。
- [whole_call_check.py](../../tools/whole_call_check.py)：`--project . --gui --save --out <固定收據>`核對全部串流、四路視窗、讀檔與真存檔往返。

本輪正式入口已切換，舊2b5bdaa執行檔以pod-play-direct保留。其他繪圖常式、GPU／SIMD及未列玩家路徑不外推完成。
