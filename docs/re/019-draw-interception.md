# 019 直接繪圖攔截

日期：2026-10-25／26。Issue：[#45](https://github.com/wicanr2/pools-of-darkness-cht/issues/45)、[#46](https://github.com/wicanr2/pools-of-darkness-cht/issues/46)、[#47](https://github.com/wicanr2/pools-of-darkness-cht/issues/47)。狀態：本作VGA繪圖、字模、共享搬移／填色／清零及單次RGB CONFORMED。

## 目前狀態

| 項目 | 本輪基準 |
|---|---|
| 程式來源 | dosgolem/pod `191479d5492fdce1e5d925ea44033d2112f75016`；Go建置資訊為乾淨來源 |
| 已驗執行檔 | SHA-256 `2c58d1d8e10f83eabb825e83e87bf781f8f136b50c3c33a8e4d93817049165fd` |
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
#46完成時尚未實作其他完整函式；#47已補VGA繪圖與共享常式，見下文。GPU／SIMD後端未實作。
未走過的圖族、動畫HD與片尾不因本輪攔截通過而視為完成。

## 2026-10-26 整段字模接管證據

使用者指定mm3新版`docs/re/direct-render-methodology.md`。該作已完成限定圖片常式、相容背板、事件預算與原子效果提交，
與前輪讀到的DRAFT狀態不同。本作不沿其位址或完成率；現有像素命令與單次RGB驗收不重新開啟。

字模入口盤點在`workplace/probe/draw-intercept-v1/whole-calls-v2/`，乾淨正式來源重播在whole-calls-v4-legacy／native。
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

#46完成時正式入口已切換，舊2b5bdaa執行檔以pod-play-direct保留。其後#47接續補VGA其餘繪圖；GPU／SIMD及未列玩家路徑不外推完成。

## #47 VGA繪圖接管完成

使用者要求完成其餘繪圖接管。唯一新增證據工作區為`workplace/probe/draw-intercept-v1/all-functions-v1/`，研究原型在隔離clone的`workplace/drawall/`。
沿用本檔IDA9.4匯出，核對VGA.DRV的檔案SHA及每條指令bytes，再以Capstone5.0.3做型別化交叉轉換。
VACE快照48A1段的2786–757F與原VGA.DRV完全相同，已證實本作動畫實際載入此驅動。
先前在GAME快照查無驅動的觀察仍成立，但不能用來推論VACE的入口。

- [draw_blocks_gen.py](../../tools/draw_blocks_gen.py)：只將已由IDA定位的純繪圖運算轉為固定Go區塊；硬體、DOS與未知指令留原版。
- [all_render_check.py](../../tools/all_render_check.py)：核對#47的完整事件、全部實際像素寫入者、四路DEMO、讀檔／片頭／封面與真存讀檔。固定收據為[019-all-render-runtime.json](019-all-render-runtime.json)。
- 繪圖區塊的程式與來源索引在私人[dosgolem/pod分支](https://github.com/wicanr2/pools-of-darkness-cht/tree/dosgolem/pod/apps/pod/drawblocks)，不含原版EXE或原始RAM。
- 規格見[025](../spec/025-direct-render-commands.md)，本作VGA範圍已達CONFORMED並接入正式入口。

### 範圍與效果

共90項函式／helper，包括VGA跳轉表與內部繪圖運算，以及GAME的Move、FillChar及清零。產生1432個固定Go區塊、7594道運算，無須在客體CPU重新解碼。
只接管符合程式指紋、表面及事件預算的呼叫；I/O、POPF、未驗寬指令與控制轉移保持原CPU邊界。
其中兩個B000區塊屬其他顯示模式，保守拒絕。未走過的玩家路線及其他HD素材仍不由本項外推完成。

通用436新增預設關閉的指令前區塊掛勾，原tick與鍵盤先執行，再按精確CS與位址篩選。前端僅提供純計畫，不取得客體寫入權。
區塊可跨函式續接，但不跨IRQ、畫面、回合或地圖取樣邊界；失配時只保留完整已驗前綴，回到失配區塊原入口。
Move／Fill／清零另有完整呼叫公式，返回、參數清除、旗標、方向與字節／字組時序由原版驗證。

REP以HostTransfer保存擁有的來源資料或填色值，不先擴成逐byte清單；提交前複製提案，逐byte相容寫入與VideoWrite的原順序仍保留，前端收到一次COPY／FILL。
重疊、跨段或其他未知條件回到已驗區塊／原版。指紋只在一次純計畫內快取，已編譯程式區不允許被計畫寫入。
CopyRAM僅複製普通低RAM，不暴露可變客體slice。原版索引讀回、規則、亂數、存檔、美術與切換操作均未改。

### 驗收結果

| 項目 | 結果 |
|---|---|
| 獨立原版區塊 | 12010組、1430個區塊；完整CPU／RAM、旗標、步數／週期及有序寫入相同 |
| 完整呼叫 | 六個實際函式×兩週期12組；另192組零／奇偶長度、方向與重疊等邊界，均相同 |
| 通用機制 | 預算、事件、拒絕原子性、callback錯誤與撤除、精確CS、predicate、向後word及來源壽命通過 |
| 正常機器重播 | 7400更新、712250000步，125406批／1822287道區塊指令，另2401次字模接管 |
| 全部像素出口 | 十個出口皆有完整效果，3306104／3343240次像素寫入由其產生，其餘保留事件邊界；畫布失配0 |
| 狀態及事件 | 全部8633次回掃、79012筆音訊、1827984次DOS請求／返回、CPU／完整RAM及127個DOS檔案同原版 |
| 正式DEMO | HD／無HD／F12／2倍完整四路、BIGPIC244及戰鬥獨立像素通過，HD原始閃現0 |
| 正常介面 | 讀檔／地圖／攻略／工具列10對、完整片頭18對、封面11對中央及完整視窗相同，原版終點及54／127／127個DOS檔案相同 |
| 真存讀檔 | 正常另存空槽B、冷啟動讀回，6對畫面及全部10個SAVE檔案同舊版；槽A五個檔案未改 |

封面閃現原始998筆均為第1219格2倍原版轉1220格3倍HD的既有主動F2操作，逐筆分類後非預期0；讀檔與片頭原始計數0。
正式候選由乾淨191479d建置，SHA見上表，驗收通過後原子替換pod-play；舊d3f97df版以pod-play-whole保留，730檔主題未改。

### 成本與重生

七組交錯、每組32次完整呼叫，重設機器狀態放在計時外。包含來源守衛、相容寫入與畫布命令，不含Owner、最終RGB及視窗。
三個VGA常式48A1:6733、4944及41AF的原版中位數9.37／9.64／0.882ms，主機效果為7.35／7.76／0.834ms，約1.27／1.24／1.06倍。
三支初次小型RAM呼叫約1.03／0.99／0.84倍，不宣稱所有呼叫更快，也不換算遊戲FPS。
早期逐byte原型與受夾具配置干擾的v1至v4量測保留在工作區，未改寫成成功。

產生器在`sc-cht-disasm:capstone5.0.3-r1`執行，使用`--project . --out workplace/dosgolem/apps/pod/drawblocks`；Go1.24.13再gofmt與建置。
正常探針以`-whole-call=false/true -draw-blocks=false/true -draw-table <blocks.json>`在全新輸出重生trace-v2-legacy／native。
`all_render_check.py --project . --commit <程式提交> --out <全新收據>`檢查全部門檻；正式GUI與存讀檔路線由既有收據腳本執行，不注入遊戲狀態。

## 2026-10-26 原生Go繪圖要求的查核與勘誤

使用者要求確認繪圖接手由Go處理，不以x86繪圖語意重播當成完成。本輪核實正式執行檔及實際來源，結論為混合實作，全部Go繪圖演算法尚未完成。
前述#47「完成」指既有固定區塊替換與同狀態驗收，不能解讀成所有VGA常式都已改為函式層級Go演算法。
補正由[#49](https://github.com/wicanr2/pools-of-darkness-cht/issues/49)追蹤，規格025新增閘門為DRAFT。本輪沒有實作剩餘演算法。

| 分類 | 實作與可回查位置 | 結論 |
|---|---|---|
| Go字模演算法 | `apps/pod/glyph_call.go:PlanGlyphCall`第91行起，8列×8欄直接依字模位元算前景／背景像素 | 已有函式層級替換；ABI與時間另算 |
| Go共享函式效果 | `apps/pod/drawblocks/transfer_calls.go:planSmallCall`，讀參數後產生擁有資料的HostTransfer搬移／填色效果 | 限定Move／Fill／清零已有完整效果；拒絕時仍落回固定區塊 |
| 固定x86語意Go區塊 | `apps/pod/drawblocks/effects.go:Plan`第54行、`blocks_generated.go`的1432個函式與7594次mark | 成功時不呼叫通用CPU.Step，也不重新解碼；仍以CS:IP派送並逐條重現暫存器、旗標、分支與記憶體語意，未達全部Go演算法要求 |
| Go畫面合成 | `apps/pod/live.go:Compose`第318行起 | 已由Go選HD／文字與最終RGB；不能證明上游VGA常式都已重寫 |
| 原x86回退 | `apps/pod/draw_blocks.go:PrepareHostBlock`及通用HostWork預算拒絕 | 來源、模式、事件預算等失配仍執行原版 |

例：`blocks_generated.go`第41937行的`block_48a1_6733`仍依次執行`e.mark`、`e.push`、暫存器賦值與`e.s.IP`推進。
`effects.go`第66行沿當前CS:IP找區塊、第87行呼叫`b.run(e)`，第200行的`alu`仍計算x86算術旗標。
這是固定程式的指令語意轉寫，不能只因檔案是Go或繞過通用解碼器，就稱為圖像／裁切／鏡像的Go演算法。
相容返回暫存器、旗標、堆疊及時間本身仍有必要；新增閘門要移除的是繪圖熱迴圈對逐條原指令語意的依賴。

正式`pod-play` SHA-256仍為`2c58d1d8e10f83eabb825e83e87bf781f8f136b50c3c33a8e4d93817049165fd`。
Go1.24.13的`go version -m`回讀`191479d5492fdce1e5d925ea44033d2112f75016`與`vcs.modified=false`；
`go tool nm`確認字模、固定區塊Plan、6733區塊及共享完整效果符號確實連結。現行相關來源與191479d差異0，排除讀到未啟用原型。
固定[分類查核](019-native-go-audit.json)記錄上述來源SHA、正式符號及判定。本輪是靜態來源／二進位核實，沒有新增執行期或效能聲明。

原[完整狀態與畫面收據](019-all-render-runtime.json)及[b26純資料增量](017-b26-runtime.json)仍有效。
HostBytes與接管指令數混合了函式效果和固定區塊，不是原生Go演算法覆蓋率。後續需按實際入口分開統計Go演算法、固定區塊及原x86回退。

## 2026-10-26 Go矩形填色首批接入

使用者授權繼續#49，先處理正常觀測的`48A1:41AF`。原版`VGA.DRV` SHA-256為`eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f`。
IDA9.4定位的原檔offset`0x41AF–0x42B2`共260 bytes、95道指令，Capstone5.0.3逐bytes核對相同；此原檔位移不等於執行期實體位址。
完整函式SHA-256為`c7db9b7b119f14464a7e76d35125922a3559d10fef36624f79f5120dd0d02f00`。
正常入口完整RAM見證SHA-256為`cc917bce17a4307f55b40d67e5f30a4ae1694008176dba2264ff9cca7adec98c`，原始RAM留本機。

已證實：far返回後兩個word為表面索引與填色值；七組表面欄位起於CS:00BD，父連結由旗標0x40辨識。
矩形含右下界，目的位移由根表面寬度與子矩形座標直接計算。工作欄位、保存區、返回、旗標與原虛擬時間的契約見規格025同名小節。
先以測試檔原型驗證9種幾何×兩種成本共18組、194個獨立原版計畫比較及八種拒絕條件，審查至READY才移到正式`apps/pod/drawblocks/native_fill.go`。

Go逐列建立HostTransfer填色，像素熱迴圈沒有opcode或CS:IP派送，也不呼叫原算術指令效果器。
相容寫入與時間標記獨立計算；每次只做預算內完整列，未完在`48A1:4268`依實際狀態續做。未知模式、表面、指紋、別名、越界、循環及不足預算維持原路徑。
本次只接受A000表面及已驗幾何子集，沒有宣稱所有表面或全部VGA函式均已完成。

通用436新增預設false的Algorithmic診斷標記，提交成功後才累加Go計畫與虛擬指令；固定區塊維持false。
Owner傳遞每回合差值，前端draw-report分開呈現。預設、拒絕不計數及跨回合差值測試通過，核心、session、pod、繪圖與前端套件及靜態檢查通過。

| 正常驗收 | 結果 |
|---|---|
| 完整機器7400更新 | CPU／完整RAM、3343240次像素、8633次回掃、79012筆音訊、1827984筆DOS事件的完整有序雜湊同原版，127檔相同 |
| Go分類 | 84408個已提交演算法計畫／2322651步，與逐出口提案總數相等；固定語意區塊另1028041步 |
| 本次矩形 | 正常路線兩個Go矩形計畫共4840步，含一次列邊界續做；不足預算仍退回既有語意路徑 |
| 正常DEMO | HD／無HD／F12／2倍四路，BIGPIC244與戰鬥獨立像素、原版終點及127檔相同，閃現0 |
| 正常介面 | 街景／地圖／攻略／工具列10對、完整片頭18對、封面11對中央及完整視窗圖同舊版 |
| 真存讀檔 | 正常另存空槽B後冷開機讀B，6對畫面及全部10個SAVE檔相同，槽A五個原檔未改 |

封面原始998筆量尺仍全部是既有F2主動切換，非預期0；讀檔與片頭原始量尺0。沒有把主動切換的原始記錄抹成0。
正式產品來源`7cd6f7a2d437f27074f46fb7337337b7f3b04637`，乾淨Go1.24.13建置SHA-256為`491e3363f57753c28796d9a1782595fcdb6dcd662282d278d93caf510b8f0fe8`。
全部gate通過後才原子替換pod-play，舊2c58d1d以pod-play-all保留；776檔主題與受測版本相同且未改。
七組交錯、每組32次完整呼叫的矩形中位成本約為原版1.24倍速度，含相容寫入與命令，不含Owner／最終RGB／視窗。
其他小型RAM樣本未全部較快，量測在其他有界工作並行時執行，不外推全遊戲FPS。

固定[019-native-fill-runtime.json](019-native-fill-runtime.json)。重生入口：[native_fill_check.py](../../tools/native_fill_check.py)接受`--project`、`--commit`、`--stage trace|full`與全新`--out`。
來源探針為`draw_call_probe.go`，新增Algorithmic逐出口提案統計，須與真正提交的計畫／指令總數相等才接受。
正常路線沿既有DEMO、direct-ui與whole-call-save腳本，均在既有Docker image與全新輸出目錄執行，不注入狀態、不增加G6探索。
其他VGA固定區塊仍待後續Go演算法，#49保持開啟。

## 2026-10-26 Go圖片RLE像素循環接入

沿用相同VGA.DRV與IDA9.4匯出，Capstone5.0.3逐bytes核對原檔offset`68F4–691B`共40 bytes、21道指令。
SHA-256為`1926d07bb1c1e91594b6a388138e57b204d8efb327addc7cebf32d0dd7ec5818`，執行期入口`48A1:68F4`。
已證實控制byte右移一位加一為長度，最低位選連續像素或重複色段；Go按資料封包解碼，產生COPY／FILL。
像素循環不經opcode或CS:IP派送。圖片外層、裁切、鏡像及其他VGA固定區塊仍保留，不稱整個6733完成。

22組封包條件的110個計畫與32個真正圖片中途狀態的192個計畫由獨立原版CPU比較，完整RAM、暫存器、旗標、時間與寫入順序相同，九種拒絕條件通過。
先審查規格025為READY，再接產品。固定語意前綴在Go入口結束提案，讓下一次計畫從已提交狀態接手；拒絕仍可循原路徑前進。
來源、模式、表面、方向、資料邊界或預算未知時保守回退，不改原版檔案、規則、亂數或存檔。

正常7400更新的全部CPU／RAM、3343240次像素、8633次回掃、79012筆音訊及1827984筆DOS事件同獨立原版，127檔與45份入口見證相同。
89509個已提交Go計畫共2550405步，固定語意區塊另800287步；其中新RLE為5101個計畫、227754步，提案與真正提交總數相等。
四路DEMO的原版終點、BIGPIC及戰鬥獨立像素、39對正常介面與6對新存B／冷讀B畫面通過。全部SAVE bytes相同，槽A五檔未改，非預期閃現0。
封面原始998筆仍為既有F2主動切換。主題776檔、字型與譯文未改，未新增G6探索。

產品來源`dabb188bdb551a072cc95c799392da78a3bc6634`，Go1.24.13乾淨建置`0a0fa8b47ab5ca903b2116bb2893a45577b7279264fc16c78f351a3ae603707c`。
所有門檻通過後原子啟用pod-play，前版491e336保存於pod-play-nativefill，較早2c58d1d仍在pod-play-all。
七組交錯、每組32次完整呼叫的6733成本約原版1.29倍速度，含檢查、相容寫入與通知，未包含Owner／視窗，未外推FPS；小型RAM函式未全部更快。

固定[019-native-rle-runtime.json](019-native-rle-runtime.json)綁定上述來源、工具與正常驗收。
重生使用`tools/native_fill_check.py --project . --family rle --commit <來源提交> --out <全新JSON>`，在Docker內執行；`--stage trace`只查機器串流，不代替完整正常路線。
原始收據在`workplace/probe/draw-intercept-v1/native-rle-v1`與`workplace/probe/g9-hd/nativerle-{play,ui,save}-v1`，均不入Git。

## 2026-10-26 Go使用色號掃描接入

`48A1:4A67`屬4944的使用色號掃描，不能把4944誤稱為文字圖片常式。原檔offset4A67–4A71共11 bytes／四道指令，IDA9.4與Capstone5.0.3逐bytes一致。
程式SHA-256為`3c8934166cf6cd8d1ecdf79eac51b0cb33bb1bb3e545b883578dcb85971e1b5a`；VGA.DRV輸入沿本研究前節。
Go直接掃描線性畫面，依色號標記CS:1772表。每次重複寫入、完整暫存器、旗標及本模型時間保留；色盤I/O與淡出控制仍由原有邊界處理。

60個合成計畫、32個真正中途狀態的192個計畫與九種拒絕通過獨立原版比較，先READY才接入。
正常7400更新的完整CPU／RAM、全部像素／回掃／音訊／DOS事件、127檔與45份真正入口見證同原版。
本次掃描新增209個已提交Go計畫／255968步，全部Go計畫89718個／2806373步，固定語意區塊另544319步。
預算不足時保留原路徑，不能宣稱這次原版掃描的每一步均被替換。分類只涵蓋主機繪圖效果，不把普通遊戲CPU算作繪圖回退。

四路DEMO、39對正常介面、6對正常另存B／冷讀B畫面通過，全部SAVE bytes相同，槽A五檔未改。非預期閃現0，封面998筆F2主動切換仍另列。
產品來源`75113aca7f41999eb9e2fc8f2a886fddae8a3842`，Go1.24.13乾淨建置`d21dec5deb0aff6c556937febddae99aa3632a702cb332832df87070bbbbae98`。
全門檻通過才原子切換pod-play，前版0a0fa8b保存在pod-play-nativerle，較早備份未動；776檔主題、譯文與字型未改。
4944七組交錯、每組32次完整呼叫成本約原版1.79倍速度，含來源檢查、相容寫入與通知，未含Owner／視窗，未外推FPS；小型RAM函式未全部更快。

固定[019-native-palette-runtime.json](019-native-palette-runtime.json)。重生沿同一`tools/native_fill_check.py`，選`--family palette`，仍須全新輸出與完整正常路線。
原始收據在`workplace/probe/draw-intercept-v1/native-palette-v1`及`workplace/probe/g9-hd/nativepalette-{play,ui,save}-v1`。
鏡像像素對原型另以804個合成及144個受控原圖計畫比較通過，三段50 bytes已逐bytes核對。規格025已READY，未在本次d21dec5接入，不外推正式鏡像完成。

## 2026-10-26 Go圖片區塊與鏡像像素對接入

本次續接`48A1:689E`的圖片區塊解析，以及`48A1:6C9E／6D42／6DF3`的鏡像像素對複製。VGA.DRV輸入沿本研究前節。
區塊解析的原檔offset689E–68F3共86 bytes／28道指令，未裁切分支走26道；SHA-256為`56ecc68cd838cf71e5de2da105cf370ccd3c2f7656159bdb2de14800a9c888a1`。
三段鏡像原檔offset同各IP，各50 bytes／25道指令，bytes完全相同，SHA-256皆為`71ffb9fb40b99437973306c9ef582b52f24b0014fb32e1feed59703a7ce12575`。
以上位移來自IDA9.4二進位段匯出，Capstone5.0.3逐bytes核對；執行期實體基底為48A10，不把原檔位移當實體位址。

Go區塊解析直接讀四byte標頭，計算x、展開長度、壓縮長度及目的範圍，再交既有RLE。Go鏡像按來源像素對反序，從右到左建立COPY。
像素與幾何循環不呼叫CPU.Step、opcode派送或原指令效果器。舊位址只作入口／續做及有序相容通知；保存狀態、旗標與時間另算。
區塊的新列與裁切、鏡像外層與PUSHF／POPF、其他固定區塊及原版回退仍保留，沒有宣稱整支6733或全部VGA均已改為Go演算法。

區塊18個合成及64個真正中途狀態計畫比較、16種拒絕通過；鏡像804個合成及144個受控原圖中途計畫比較、九種拒絕通過。
另以原圖資料、x=−7／0／7、鏡像旗標1／3、兩種成本及6／1024步預算跑24個完整原版呼叫，169780個鏡像Go計畫真正提交，返回、完整RAM與時間全部相同。
改旗標／座標的案例明列為受控呼叫，不當正常玩家路徑。上述證據先達READY才接入，核心、session、pod、drawblocks、Xvfb前端及go vet通過。

正常7400更新的完整CPU／RAM、有序像素／回掃／音訊／DOS事件、127檔與45份入口見證同獨立原版。
本次區塊解析新增5091個Go計畫／132366步；全部Go計畫94809個／2938739步，固定語意區塊411953步。這條DEMO未觀測鏡像新入口，不能把受控鏡像計畫加進正常分類。
四路DEMO、39對介面及6對正常新存B／冷讀B畫面相同，全部SAVE bytes及槽A五檔相同。非預期閃現0，封面998筆F2主動切換另列。

產品來源`a8184cc2c48790b59a71af65d9d02dcca7ca7da3`，Go1.24.13乾淨建置`faf0899063cb527dd29b5deef9f3fe106822910d23799c6f6a37a351b7e2ebb7`。
全門檻通過後原子啟用，d21dec5保存在pod-play-nativepalette，較早備份保留；776檔主題、字型與譯文未改。
6733七組交錯、每組32次完整呼叫成本約原版1.68倍速度，含資格、相容寫入與通知，未含Owner／視窗，不外推FPS；小型RAM函式未全部更快。
固定[019-native-image-runtime.json](019-native-image-runtime.json)綁定來源、工具、完整驗收及Go演算法來源審查。重生沿`tools/native_fill_check.py --family image`，其他參數與前節相同。
原始收據在`workplace/probe/draw-intercept-v1/native-image-v1`及`workplace/probe/g9-hd/nativeimage-{play,ui,save}-v1`，#49保持開啟。
