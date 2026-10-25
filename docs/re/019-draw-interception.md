# 019 直接繪圖攔截

日期：2026-10-25。Issue：[#45](https://github.com/wicanr2/pools-of-darkness-cht/issues/45)。狀態：像素出口與單次RGB合成已驗；完整函式空跳過仍未證。

## 目前狀態

| 項目 | 本輪基準 |
|---|---|
| 程式來源 | dosgolem/pod `c0fc72b23476d695e2ca46d3431cf2fa9a9d54b6`；Go建置資訊為乾淨來源 |
| 已驗執行檔 | SHA-256 `2b5bdaa2c6c4bdd3746108ed30e46db065f3a928b1a90a911c01e7de298deb81` |
| 正式主題 | 730檔；既有BIGPIC、CPIC與正常四路收據不重開為未完成 |
| 新要求 | 使用者指定直接攔截繪圖，避免原版底圖與HD的重複繪製；參考mm3 |
| 參考限制 | mm3規格075為DRAFT，原版空返回不等價；只沿方法，不沿其位址或完成狀態 |
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
原版背板與函式其他指令仍執行，不能稱為完整函式入口已空跳過；GPU／SIMD後端亦未實作。
未走過的圖族、動畫HD與片尾不因本輪攔截通過而視為完成。
