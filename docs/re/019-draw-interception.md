# 019 直接繪圖攔截

日期：2026-10-25／26。Issue：[#45](https://github.com/wicanr2/pools-of-darkness-cht/issues/45)、[#46](https://github.com/wicanr2/pools-of-darkness-cht/issues/46)、[#47](https://github.com/wicanr2/pools-of-darkness-cht/issues/47)。狀態：本作VGA繪圖、字模、共享搬移／填色／清零及單次RGB CONFORMED。

## 初期基準

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

## 2026-10-26 入口見證數量與觀測勘誤

追加核對發現，前四批Go演算法的固定JSON各保存42份直接入口見證，原版基準45份；前文與WORKLOG的「45份真正入口見證相同」應讀為42份已觀測見證相同。
缺少的`48A1:2772／2AEB／2A90`由固定主機區塊內部進入，沒有逐原指令呼叫BeforeStep，因此原探針未擷取。這項缺口在#47的固定區塊接入後即已存在。
JSON原始明細保留，不改寫成45份；原有完整CPU／RAM終點、全部有序像素／回掃／音訊／DOS事件、正常介面與存檔比較仍有效，未將觀測缺口寫成產品狀態錯誤。

新的[build_observed_draw_probe.py](../../tools/build_observed_draw_probe.py)只在Docker內複製當前繪圖來源至隔離clone的workplace。
研究版固定計畫在已登錄函式入口前交回控制，再由原有探針擷取；不更動正式Plan、原始資料、玩家輸入或客體狀態，也不補猜中途CPU。
產生器綁定原／研究版各Go檔、探針、工具與二進位SHA-256；此為分割觀測版，須與正式未分割版分開記錄。
重生參數為`--project <專案根> --module <隔離clone> --name <全新名稱> --program-commit <來源提交> --out <全新manifest>`，需既有Go image、快取與唯讀專案輸入。
正式版的已提交Go分類以未分割trace為準；補充入口須另對原版完整RAM、CPU、順序事件與檔案核對，不以補充版取代正式分類。

## 2026-10-26 Go淡出色差與插值接入

`48A1:4B19`色差及`48A1:4B6C`插值各33 bytes，IDA9.4與Capstone5.0.3逐bytes一致；原檔offset同IP，仍以VGA.DRV固定輸入為準。
色差SHA-256為`426b9c7bab4ee58a83495f6ee251ade3d04d01ce17e57c163cfdcd18c4b00f86`，插值為`391005d80b86b978189cde8e392fbfbca279480fda3f7f3a175cb9c14ed2f2ae`。
Go直接遍歷五組RGB陣列，算方向、絕對差、最大距離與累加誤差調色；不派送opcode或指令效果器。硬體讀寫與回掃等待仍留原邊界。

72個色差、40個插值合成計畫及兩個真正入口各32個狀態的384個計畫由獨立原版CPU比較通過，另有23種拒絕。
原版正常7400更新增取兩份入口，完整狀態／事件同舊原版。第一版孤立4944夾具缺少當時VGA色盤狀態，未走插值，該失敗保留；新測試直接從真正RGB陣列入口開始。
這也限制舊4944成對成本：它使用新建VGA狀態，不能當正常完整淡出耗時。未重寫舊量測數字，最新收據明列限制。

先READY才接入，核心、session、pod、drawblocks、Xvfb前端與go vet通過。正式未分割正常路線有94880個Go計畫／3116445步，固定語意區塊234247步。
其中色差一個計畫3287步、插值70個計畫174419步；只是主機繪圖效果分類，不是全遊戲覆蓋率或FPS。
正式版的44份直接見證與原版相同；分割觀測版另補齊三個內部helper，全部47份CPU／完整RAM見證相同。正式分類與補充觀測數字分開保存，見前節勘誤。

正式與分割版的完整CPU／RAM、有序像素／回掃／音訊／DOS事件、127檔同原版；四路DEMO、39對介面及6對正常新存B／冷讀B畫面通過。
所有SAVE bytes與槽A五檔相同，非預期閃現0，封面998筆F2主動切換仍另列。未增加G6探索，776檔主題、字型與譯文未改。
產品來源`df24a242074b835079f0bb1a6a4c095bad444bec`，Go1.24.13乾淨建置`014392d55bfb2d5280559aedc240adcb1c8d198ba1307873b7704c02588cecc6`。
全門檻通過後原子啟用，前版faf0899保存在pod-play-nativeimage，較早備份保留。

固定[019-native-fade-runtime.json](019-native-fade-runtime.json)綁定正式／分割探針、全部來源、工具與原版新增入口；`tools/native_fill_check.py --family fade`重生。
原始收據在`workplace/probe/draw-intercept-v1/native-fade-v1`與`workplace/probe/g9-hd/nativefade-{play,ui,save}-v1`。
其他色盤工作準備、圖片控制、裁切及固定區塊仍待後續Go演算法，#49保持開啟。

## 2026-10-26 Go逐色RGB工作接入

`48A1:4B9A／4BC1`依使用色表與RGB游標準備下一筆色盤工作，或交回本輪完成處理。
原檔offset4B9A–4BC5共44 bytes／17道指令，IDA9.4與Capstone5.0.3逐bytes一致，SHA-256為`548776df6fec7df3ebb2f57ade4412c8704fee225a928411593571d6c7c3b525`。
Go直接取得色盤索引與RGB三byte，準備原版工作欄位、參數與呼叫框架，在2AA8交給既有硬體輸出；不逐指令派送色盤工作。
相容返回位置、旗標、堆疊殘值與虛擬時間另算。來源指紋、數值、索引、游標、模式、方向、別名或預算未知時保持原路徑。

32個獨立原版工作比較與13種拒絕通過，先READY才接入。第一版漏算六個word堆疊寫入的12個成本，查明e.push契約後修正，失敗保留，未改測試期望。
核心、session、pod、drawblocks、Xvfb前端及go vet通過。正常路線真正接手6865筆RGB工作與63次完成交接，共116831步；有條件拒絕及固定路徑仍存在，不宣稱每次工作都已替換。
全部Go計畫101808個／3233276步，固定語意區塊117416步；只是主機繪圖效果分類，不是全遊戲覆蓋率或FPS。

正式未分割版完整CPU／RAM、有序像素／回掃／音訊／DOS事件、127檔及44份直接見證同原版。
研究分割版另補三個內部helper，全部47份CPU／完整RAM見證同原版；分割版分類不取代正式數字，來源與二進位SHA另列。
四路DEMO、39對介面及6對正常新存B／冷讀B畫面通過，全部SAVE bytes與槽A五檔相同。非預期閃現0，封面998筆F2主動切換仍另列。
776檔主題、字型與譯文未改，沒有新增G6探索。

產品來源`8c2c383b747b5a0e9ec8bae2fe035c6f48fe91e3`，Go1.24.13乾淨建置`340660d02c2e8c49c49fd3d5fc602245500b8d3ce94f590b97a466eeafa75423`。
全門檻通過後原子啟用，014392d保存在pod-play-nativefade，較早備份保留。成對成本沿既有孤立入口夾具，未據此宣稱逐色RGB工作的正常路線加速。
固定[019-native-color-runtime.json](019-native-color-runtime.json)，重生使用`tools/native_fill_check.py --family color`；原始收據在`workplace/probe/draw-intercept-v1/native-color-v1`及`workplace/probe/g9-hd/nativecolor-{play,ui,save}-v1`。
剩餘圖片控制、裁切／鏡像外層、初始化與其他固定區塊仍待逐項補證；全部Go演算法閘門仍DRAFT，#49保持開啟。

## 2026-10-26 Go色盤陣列準備接入

續接4944內的三段資料工作：`48A1:4A83`整理使用色號、`48A1:4AFD`填入目標RGB、`48A1:4B3A`初始化插值誤差。
VGA.DRV輸入沿本研究固定SHA；原檔offset同IP，兩段14 bytes／七道指令及一段27 bytes／九道指令，由IDA9.4與Capstone5.0.3逐bytes交叉核對。
三段SHA-256依序為`1814a41894eea9e581384ac7c8a316353039634bf747a66ceb5ec19aff3361af`、`343202170a7507bd27ba58e074710568546eac5f1fc99691bcac8d8e11a98349`、`592b32248b4af06920a2979be064570ce726975b43eae84d9591126394638a7d`。

Go直接過濾256項使用標記，按降序建立word色號表；按RGB三byte元素填入目標陣列；再把三倍使用色數的誤差陣列填為最大距離的一半。
資料循環沒有CPU.Step、opcode或指令效果器派送；旗標、暫存器、原IP與虛擬時間另作相容效果。普通RAM區間、對齊、模式、方向、計數與指紋不符時回原路徑，硬體讀寫不變。
36個清單、72個目標、32個誤差合成計畫，加孤立原版4944呼叫65個中途狀態的390個計畫，共530次獨立原版比較及21種拒絕通過。
孤立呼叫使用新建VGA狀態，只證明其條件下的函式效果；正常狀態由以下完整機器重播另驗。先達READY才加入產品，核心、session、pod、drawblocks、Xvfb前端與go vet通過。

正常7400更新實際接手三個Go計畫，清單1498步、目標763步、誤差9步，共2270步。正式Go計畫101811個／3235546步，固定語意區塊115146步。
完整CPU／RAM、有序像素／回掃／音訊／DOS事件、127檔及44份直接入口見證同原版；三個區塊內部helper以分割觀測補證，全部47份CPU／完整RAM相同。
四路DEMO、39對介面及6對正常另存B／冷讀B畫面相同，全部SAVE bytes及槽A五檔未改。非預期閃現0，既有998筆F2主動切換仍另列。
776檔主題、字型與譯文未動，未新增G6探索，未把這份分類稱為全遊戲覆蓋率或FPS。

產品來源`eb5e49b26f8113a1f678601800668b00999b5fa2`，Go1.24.13乾淨建置`5dc0b2977b13dda1c2b955b77f7a374cd0b91884decf1ce52269ab5f4d0cc26b`。
全部門檻通過後原子啟用，前版340660d保存在pod-play-nativecolor，較早備份保留。
固定[019-native-setup-runtime.json](019-native-setup-runtime.json)，重生選`tools/native_fill_check.py --family setup`；原始收據在`workplace/probe/draw-intercept-v1/native-setup-v1`與`workplace/probe/g9-hd/nativesetup-{play,ui,save}-v1`。
觀測建置器新增module HEAD相等檢查，來源標籤填錯時在建立目錄前拒絕；不符標籤的早期manifest保留在工作區，驗收只使用正確重建的v2。
其餘圖片控制、裁切／鏡像外層與未驗入口仍待Go演算法，#49保持開啟。

## 2026-10-26 Go圖片列控制接入

沿正常6733圖片呼叫追到48A1:6850，原檔VGA.DRV的6850–689D共78 bytes／26道指令。
輸入SHA-256為`eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f`；IDA9.4二進位匯出使用base0檔案位移，執行期CS48A1的實體基底48A10。
Capstone5.0.3逐bytes相同，區間SHA-256為`1c4ab9672accc6ec8dcd86e8703b26cf877efa2ca54e8afa5268d2c5fed7a62b`；原始函式與運算元保留於vga-ida.json。

已證實：DS:SI的列號byte以255結束。加上SS:BP+10的垂直座標後，與`cs:word_EC3／word_EC7`有號比較；太低結束、太高略過該列封包。
有效列依EC1左界、ED3／ED5表面原點、EDD列寬，計算`(左界−原點X)+(列號+呼叫Y−原點Y)×列寬`，保存120C後交689E的區塊解析。
原版320寬分支以位元組交換及位移算乘積；Go直接使用幾何公式，另保存該分支AX=320、DX=64倍列差與最後旗標。
1250列末標記在讀列號之前清零，原版寫入順序與上下文獨立保留。新幾何不呼叫CPU.Step、opcode派送或指令效果器。

174個合成原版比較涵蓋320／其他列寬、端點、結束、上下裁切及有號座標，另由既有正常DEMO的6733入口重跑取得64份真實中途狀態，兩成本／三預算共384個計畫比較。
完整CPU／RAM、有序寫入、返回、最後指令與兩成本皆同原版，22種拒絕不改記憶體；558個比較通過後，規格025先READY再接產品。
研究初版一般列寬少算乘法14成本，獨立原版拒絕92，查明效果器mul的既有契約後修正為106。失敗native-row-draft-tests.log保留，未曾接產品。
這些孤立函式狀態只證明其輸入下的效果；完整事件與正常玩家路線仍須另驗。

程式來源`7a4f16852fa36eb5dcf19033a46ebfe5093866fc`，Go1.24.13乾淨候選`9c3e65214b5e45e07dbdf5ce3129786fc29058352ede46930e297490911d1ae6`。
核心、session、pod、drawblocks、Xvfb前端與go vet通過；完整正常驗證及正式啟用狀態見[本批固定收據](019-native-row-runtime.json)。
重生入口為Docker內的`tools/native_fill_check.py --family row`，使用`workplace/probe/draw-intercept-v1/native-row-v1`與`workplace/probe/g9-hd/nativerow-{play,ui,save}-v1`。

正式未分割探針完成7400更新：有效列635個／14605步、結束24個／120步，共659個Go列計畫／14725步。
全部Go計畫102470個／3250271步，固定語意區塊另100421步。分類限HostWork，不是全遊戲覆蓋率。
完整CPU／RAM、有序像素3343240筆、回掃8633筆、音訊79012筆、DOS1827984筆與127個檔案同獨立原版。
44份直接入口見證加三份固定區塊內部入口的分割補證，全部47份CPU／完整RAM相同；正式與研究分割版計數分開保存。
既有6733完整函式夾具的七組交錯成對量測約1.75倍原版速度，包含其他已接入演算法，不是本次列控制的獨立增益。
孤立4944使用新建VGA，不能代表正常完整淡出時間；三支小型RAM呼叫本輪均較慢，沒有把量測外推整體FPS。

四路DEMO、39對介面與六對真存讀檔通過；正常空槽B已建立並在冷啟動後讀回，全部10個SAVE檔同基準，槽A五檔未改。
非預期閃現0，封面既有998筆F2主動切換另列。776檔主題、字型與譯文未改，README的34張圖片保留，沒有新增G6探索。
全部門檻通過後原子啟用9c3e652，前版5dc0b29保存在pod-play-nativesetup，較早備份保留。
剩餘固定前綴研究另保存在native-row-v1/fixed-hot，完整狀態／事件與正式探針相同，100421步加總一致。
其中2AA8／2AAC／2AB3／2AB8／2ABC／2AC1入口合計75996步，主要是硬體色盤輸出的ABI與控制邊界。
這些是前綴入口的計數，不當完整函式成本；下一步先核對2AA8硬體輸出契約，不能僅把暫存器／IP轉寫換個名稱就當成Go演算法。
圖片其他裁切、鏡像外層、初始化與未驗入口仍保留；#49及全部Go演算法閘門尚未完成。

## 2026-10-26 Go完整RGB輸出接入

原檔VGA.DRV的2AA8–2AC2共27 bytes／16道指令，執行期CS48A1、實體基底48A10；IDA9.4二進位匯出base0，以檔案位移核對，Capstone5.0.3逐bytes相同。
輸入SHA沿前節，程式區間SHA-256為`7134f5f3fbaf2adf007afa01632ff5ec980bd4c127a095defaa954fd831b4801`。
已證實：遠返回之後的四個word依序是色號、紅、綠、藍，2AA8負責關中斷、選色號、寫RGB、開中斷及返回。
Go直接解析成RGB工作，交通用HostPaletteWrite；通用層只支援這筆VGA工作，沿現有Machine.Out8更新DAC、VGA、Ports與PortLog，不提供任意I/O或逐opcode效果派送。
標準索引及RGB順序引用[FreeVGA Color Registers](https://www.scs.stanford.edu/10wi-cs140/pintos/specs/freevga/vga/colorreg.htm)，沒有重做DAC硬體或週期考古。

通用命令先驗mode13h、RGB、所有步數／成本及RAM／搬移順序，再按值複製、合併提交。四筆I/O保留各自時間，不能全記在批次起點。
原生工作限定已驗指紋、普通RAM堆疊／參數、色號0–255、RGB0–63及足夠事件預算；保留原BP、AX藍色、DX3C9、旗標、遠返回與第一步的BP保存寫入。
16步／130模型成本，四次I/O在第6／9／11／13步，成本32／56／78／100；DOSBoxCost為16及對應步數。這是固定dosgolem模型，不是實體硬體逐週期對拍。
初版誤把RETF再加一次近返回成本，獨立原版拒絕138；沿CPU的RETF分支核對後修正130。失敗log與原型來源快照留在clone的workplace。

未接產品原型及正式計畫各通過192個合成、六個真實入口隔離比較及14種拒絕；含DAC初始0／1／2分量相位、兩成本、色號／RGB端點與旗標。
完整CPU／RAM、VGA／DAC、Ports及有序PortLog的值／步數／成本相同。真實入口來自原版正常DEMO，但隔離比較使用新建VGA，完整正常狀態仍另驗。
通用層15種拒絕、擁有權／影像寫入合併／事件恰好及不足通過；六支完整原版函式兩成本比較再加入VGA、DAC、Ports與逐筆PortLog，全部相同。
規格025及通用436先READY才接產品；核心、session、pod、drawblocks、Xvfb前端與go vet通過。

產品來源`c0519513309b0ac2ddc3260cee714ee5b0b4cf9d`，Go1.24.13乾淨候選`874d5d8b990073436554e14011acd62eb2986aa827d6b25aa86e1626e6055bce`。
完整正常範圍與啟用狀態見[019-native-output-runtime.json](019-native-output-runtime.json)，重生入口為Docker內的`tools/native_fill_check.py --family output`。
原始收據在`workplace/probe/draw-intercept-v1/native-output-v1`與`workplace/probe/g9-hd/nativeoutput-{play,ui,save}-v1`。

正常7400更新實際提交7597個完整RGB工作／121552步。全部Go計畫110067個／3371823步，固定語意區塊24451步。
此前普通CPU執行的四筆I/O及CLI／STI也交主機，因此HostWork總數由3350692增至3396274；全部遊戲虛擬步數仍為712250000，兩者不可混算。
完整CPU／RAM、有序像素3343240筆、回掃8633筆、音訊79012筆、DOS1827984筆及127檔同獨立原版。
44份直接入口見證與三份固定區塊內部入口分割補證，全部47份CPU／完整RAM相同；正式與研究分割版分類分開保存。

四路DEMO、39對正常介面及6對真存讀檔相同；空槽B正常建立及冷啟動讀回，全部10個SAVE檔同基準，槽A五檔未改。
非預期閃現0，既有998筆F2主動切換另列；776檔主題、字型與譯文未改，README保留34張圖片，沒有新增G6探索。
全部門檻通過後原子啟用874d5d8，前版9c3e652保存在pod-play-nativerow，較早備份保留。
通用[436規格收尾](https://github.com/wicanr2/pools-of-darkness-cht/blob/700872d8428152cc8dd11d82d451744399491c49/docs/spec/436-host-call-effects.md)只改文件，Go來源與程式提交c051951完全相同，不為文件狀態改寫已驗二進位。
本次整支6733孤立函式約1.81倍原版速度，包含既有Go演算法；不能歸因於單一RGB工作，三支小型RAM呼叫仍較慢，不外推整體FPS。

下一個窄任務是689E圖片區塊標頭的水平裁切分支與691C銜接，先補來源、幾何與ABI證據；未READY不接產品。
2A9E／2AA3回掃等待沿既有VGA平台契約，不開逐週期硬體RE。其他圖片裁切、鏡像外層、初始化與未驗入口仍保留，#49保持開啟。

## 2026-10-26 Go圖片來源、裁切與翻轉接入

本批完成前節列出的圖片窄任務，來源為固定SHA的VGA.DRV及既有IDA9.4匯出，位址仍以原檔offset和執行期CS:IP分列。
6733先用Go解析表面父鏈及畫格索引，6832／689E處理列標記與左右裁切，691C及翻轉分支直接解碼RLE並計算可見像素。
6A0C以圖頭寬高及旗標決定末列／末欄，水平鏡像由Go反排每對像素；ABI堆疊、旗標、返回及有序寫入另行保留。
不以CPU.Step、opcode派送或固定指令效果器生成圖片。來源與幾何未知、別名或不足事件預算時，仍保留原路徑。

前向控制與裁切942次、翻轉與鏡像2254次、三種CPU機型12次、零可見數八次及初始化116次獨立原版計畫比較通過。
另有48個受控完整原圖呼叫，前向26種、鏡像17種及初始化16種拒絕；未知機型與短預算另驗。受控呼叫改變位置／翻轉測試參數，不當正常玩家路徑證據；兩種成本、CPU、完整RAM、旗標、逐筆像素及模型時間各自比較。
這些只是已驗資料與模型條件，未觀測的來源、畫格及模式不外推。

原型查明三項差異，失敗資料保留在clone的workplace。數個條件跳轉的成本先由原版逐段核對後訂正；鏡像literal的零可見數由6DEF直接跳到6E29，仍輸出一個byte；圖片返回的40成本不能套用41AF已扣除分支成本後的35增量。
StepView新增唯讀CPUModel供鏡像PUSHF／POPF恢復保留位元，沿436的期限與token；未知機型拒絕。三種機型及過期視圖、獨立原版旗標恢復均通過。

程式`ed004f9c49fb405290e3919fdc0fac29882c787c`已推本專案dosgolem/pod。Go1.24.13候選為乾淨來源，SHA-256 `994d498c735e232819819ade8c138cebb7799ccf80d2353519d8e08d2a80434b`。
正常新增24個來源初始化、24個返回及635個列標記計畫，共3622步。全部Go計畫110750個／3375445步，固定語意區塊20829步；全部遊戲虛擬步數仍712250000。
完整CPU／RAM、有序像素3343240筆、回掃8633筆、音訊79012筆、DOS1827984筆及127檔同獨立原版；44份直接見證與三份分割補證同47份原版，分類不混算。

四路DEMO、39對介面及六對真存讀檔、全部十個SAVE檔相同，槽A五檔未改。非預期閃現0，既有998筆F2切換另列；776檔主題、字型與譯文未改。
全部驗收通過後原子啟用994d498，前版874d5d8以pod-play-nativeoutput保留。固定[019-native-clip-runtime.json](019-native-clip-runtime.json)保存來源、比較範圍及啟用綁定。
重生入口是Docker內的`tools/native_fill_check.py --family clip`，原始收據在`workplace/probe/draw-intercept-v1/native-clip-v1`及`workplace/probe/g9-hd/nativeclip-{play,ui,save}-v1`。

本批另在獨立研究副本解出4A72／4A31色盤列推進及4B55循環控制，分別212與72次比較、18與14種拒絕通過，規格025已READY，尚未接產品。
剩餘固定區塊主要為色盤其他外層、表面初始化、回掃／ABI及短預算續行。色盤標準沿[FreeVGA色彩介面](https://www.scs.stanford.edu/10wi-cs140/pintos/specs/freevga/vga/colorreg.htm)，回掃沿既有VGA服務，不增加逐週期硬體考古；#49仍開啟。

## 2026-10-26 Go色盤資料流程與回掃等待接入

沿固定VGA.DRV及IDA9.4原始位址，Go直接解析表面父鏈、掃描游標、來源色號清單與RGB陣列，建立有界RGB讀取工作，並以兩個等待狀態沿既有VGA模型等待回掃。幾何與資料迴圈不派送opcode或固定指令效果器，ABI、旗標、有序scratch與模型時間另存。
通用436新增受期限約束的PaletteRGB／VGAStatusSnapshot快照，以及有界RGB／3DA讀取，仍使用既有In8／Out8服務。只開放列出的mode13h線性表面與模型；未知快照、來源、模式、別名或事件預算拒絕。硬體模型沿FreeVGA契約及既有近似，不聲稱實機逐週期一致。

研究與產品測試涵蓋色盤列212次、循環216次、RGB36次、等待392次、來源清單351次、初始化42次、階段交接144次獨立比較；快照另240次服務比較。核心、session、pod、drawblocks、前端及go vet通過。
原版正常循環會由2A90令ES=0040，4B55不讀ES，故移除不必要的ES=CS限制前先以三種ES重驗。等待初始化31成本含段載入及前綴；RGB讀取只令三次In8增加portTicks，選色Out8不加。研究失敗及修正log保留，沒有修改原硬體服務來迎合測試。

程式來源15fd65ad9510c0cea4dd852fdcc288c05d6c3f9e，乾淨Go1.24.13候選b369deaad3a8dcd253a0083ee02519dd6401d84bcea371a14633f5422fe2c7bd。正式正常重播為111710個Go計畫／3398106步，固定區塊2576步；完整CPU、RAM、3343240筆像素、8633次回掃、79012筆音訊、1827984筆DOS與127檔同原版。
正式46份直接入口與一份分割補證同47份原版。四路DEMO、39對介面及六對新存B／冷讀B、十個SAVE檔同基準，槽A五檔未改；非預期閃現0，998筆F2主動切換另列，776檔主題與README的34張直接圖片保持。
全部通過後原子啟用b369dea，前版994d498以pod-play-nativeclip保留。[固定收據](019-native-control-runtime.json)綁定來源、程式、工具、事件與啟用；重生入口是Docker內的tools/native_fill_check.py --family control，原始資料在workplace/probe/draw-intercept-v1/native-control-v1及workplace/probe/g9-hd/nativecontrol-{play,ui,save}-v1。
七組交錯完整呼叫夾具中6733約1.77倍、4944約1.65倍原版，小型RAM呼叫仍可能較慢。4944夾具是新建VGA，不能代表完整正常淡出耗時，亦不外推FPS。

後續單點2B10／2C04在未接產品的研究區完成456次原版比較及26種拒絕，規格READY；3338範圍填色原型80次比較通過，仍DRAFT。未觀測入口、其他固定與短預算續行仍待接管，#49保持開啟。

## 2026-10-26 Go單點與指定範圍矩形接入

原檔VGA.DRV offset2B10／2C04為單點寫入／讀回，3338為指定範圍矩形。IDA9.4輸入及位址基準同前節；三份函式指紋與範圍見規格025，不改原名或運算元。
Go先解表面及父鏈，以選取表面裁切，再按根表面的原點與列寬計算像素。矩形先取有序座標的區間交集，再逐列發FILL；範圍外點讀回FFFF，寫入不產生像素。ABI、scratch、有序寫入及模型成本另存，像素生成不派送opcode或固定指令效果器。
未知程式、表面、模式、堆疊／來源別名或不足預算拒絕。指定矩形尚以完整計畫為原子單位，反向座標拒絕，保留原路徑。

點384個邊界／旗標／成本計畫及72個父鏈／偏移計畫、矩形80個邊界及12個父鏈計畫同原版，另有26／15種拒絕。提交層各16次整台SaveState比較一致，矩形另驗RunLimit不足不改狀態；核心、session、pod、drawblocks、前端與go vet通過。
程式ac2ec6fbf57ca5424471151d7fdab48a8c6a1c6d已推dosgolem/pod，乾淨Go1.24.13候選6e6202864d80223c09d8c8ad2540fc948c2d9c362de734cb3d40eb7df8c70b1f。正常機器、四路DEMO、39對介面、六對新存B／冷讀B及十個SAVE檔同基準，槽A五檔未改；非預期閃現0，998筆F2切換另列，776檔主題與34張README直接圖片維持。
新三個出口在正常DEMO未觀測，111710個Go計畫／3398106步與固定2576步維持。正式46份直接加一份分割補證同原版47份；新增路徑的正確性限受控原版入口，不寫成正常玩家覆蓋。
全部驗收通過後原子啟用6e62028，前版b369dea以pod-play-nativecontrol保留。[固定收據](019-native-primitive-runtime.json)綁定來源、工具與啟用；重生入口是Docker內tools/native_fill_check.py --family primitive，原始收據在workplace/probe/draw-intercept-v1/native-primitive-v1及workplace/probe/g9-hd/nativeprimitive-{play,ui,save}-v1。

並行研究的矩形逐列續行已通過92個計畫、48個前綴與48個完整受控原版呼叫、15種拒絕，共858個Go計畫。完整CPU／RAM與SaveState bytes一致，規格READY，尚未接產品。其他固定與未驗入口仍待接管，#49保持開啟。

## 2026-10-27 Go軸向直線與矩形續行接入

沿同SHA的VGA.DRV與IDA9.4位址，2CFC的已驗水平／垂直線用Go排序與裁切、span及列游標產生像素。原版八個垂直位置只保存通知與模型成本；2F29 LOOP沿既有CPU模型兩成本，不加條件跳轉五成本。斜線、非相交及有號差值溢位仍回原路徑。
3338矩形可在3418依SS／SP、scratch、原始參數與Go列游標分批接續。Go完整計畫提供已驗ABI前綴，前綴不重複提交像素；每回合重新驗證狀態與預算，完整CPU／RAM、旗標、模型時間及SaveState同原版。

軸線144組資料共288個預算／成本／旗標計畫及24個父表面／偏移計畫、18種拒絕通過；提交層16個整台機器比較相同。矩形92個列計畫、48個前綴、48個完整受控呼叫及15種續行拒絕通過，共858個Go計畫。正式selector另12個完整呼叫與674個列計畫通過；小矩形可在足夠預算一次完成，不要求每次都分批。
首版selector測試誤把小矩形一次完成判為未續行；修正測試覆蓋條件，產品行為不改。全套核心、session、pod、drawblocks、前端及go vet通過，失敗與修正log保留。

程式7c6cf01ae37a33a488ae1a367410714780d2ea3a，乾淨Go1.24.13候選63cb5b395571b97093d0e4ecd91e69e6cf28d352ef093c31af19e2470214e51e。完整正常CPU／RAM、有序像素／回掃／音訊／DOS與127檔同原版；46份直接加一份分割補證同47份原版。
四路DEMO、39對介面、六對新存B／冷讀B與十個SAVE檔同基準，槽A五檔未改；非預期閃現0，998筆F2主動切換另列。新出口在目前DEMO未觀測，111710個Go計畫／3398106步與固定2576步維持，不能以受控輸入冒稱新正常路線。
全部驗收完成後原子啟用63cb5b3，前版6e62028以pod-play-nativeprimitive保留，776檔主題與34張README直接圖片未改。[固定收據](019-native-axis-runtime.json)綁定來源、工具及啟用；重生入口為Docker內tools/native_fill_check.py --family axis，原始收據在workplace/probe/draw-intercept-v1/native-axis-v1及workplace/probe/g9-hd/nativeaxis-{play,ui,save}-v1。

並行研究的表面內陡／緩斜線各240次獨立原版計畫、20種拒絕通過，規格READY，尚未接產品。兩者用Go整數誤差累加，緩斜線保留31EB對EC7的RAM暫存效果；其XCHG六成本只計一次ModRM記憶體成本，原型誤算已依原服務修正，沒有改CPU模型。裁切及其他未驗繪圖仍待接管，#49保持開啟。

## 2026-10-27 Go斜線與裁切游標接入

沿VGA.DRV原檔SHA及IDA9.4的2CFC–3235，以Go座標、主次距離與整數誤差產生陡／緩斜線。309C／30B9及3206／321C只選typed游標；原位置保留相容寫入與模型成本，不重播opcode。短預算可先交出第一點前的初始化，再分批接續。
裁切保留原版窄高視窗特例：子表面0,0–10,100，線段−10,0至20,90輸出62點，最大X為20。側邊重算起點沒有相應扣除1139；Go以原版主次計數保留結果，不改為一般矩形交集。最初交集假說及反例均保留於規格025。

表面內完整斜線480個原版計畫、40種拒絕與32個整台機器提交通過；游標424次、前綴32次、128次完整受控呼叫與44種拒絕通過，共848個Go計畫。裁切模型以固定種子0x2CFC通過5020個像素案例，游標1114次、前綴534次與44種拒絕通過；正式selector108次全呼叫共1280個Go計畫，完整SaveState bytes同原版。

產品來源602b79e6d44100a79dfdab2dab77fa4c07ee3843，乾淨Go1.24.13候選9e45287cec07ed9256b768d55778ea76de53f5f028dd13708e611b2895153858。核心、session、pod、drawblocks、前端及go vet通過；完整正常CPU／RAM、有序事件及127檔同原版，46份直接加一份分割補證同47份原版。
四路DEMO、39對介面與六對真存讀檔、十個SAVE檔相同，槽A五檔未改。非預期閃現0，998筆F2主動切換另列；新斜線出口在目前DEMO未觀測，正常Go計數與固定2576步維持。受控比較不外推正常玩家路線或FPS。
全部通過後原子啟用9e45287，63cb5b3以pod-play-nativeaxis保留。固定[019-native-line-runtime.json](019-native-line-runtime.json)綁定來源及啟用；重生入口為Docker內`tools/native_fill_check.py --family line`，原始收據在`workplace/probe/draw-intercept-v1/native-line-v1`及`workplace/probe/g9-hd/nativeline-{play,ui,save}-v1`。
矩形搬移仍在獨立研究副本，其他非相交與未驗繪圖保留，全部Go閘門仍DRAFT，#49開啟。

### 斜線計畫分類勘誤

前段及019-native-line-runtime.json沿正式selector測試的訊息，將108次裁切呼叫選取的1280個計畫全標為Go演算法計畫。此標籤不精確。以同一602b79e來源綁定的觀測副本、空額外入口表，逐筆檢查HostWorkPlan.Algorithmic後，結果是898個Go演算法計畫與382個固定相容前綴。原始收據保留，不改寫歷史。
獨立CPU／完整RAM、有序像素、模型時間及SaveState比較仍通過。勘誤不改正式執行檔、像素演算法或正常路線的111710個Go計畫；後者原本即以Algorithmic計數。分類原型與log在workplace/dosgolem/workplace/palette-advance-research/clipped-plan-class-audit_test.go及同clone的workplace/plan-class-audit.log，新版驗收工具綁定其SHA-256。

## 2026-10-27 Go表面矩形搬移接入

VGA.DRV原檔SHA及IDA9.4定位沿用前節。3F5E–41AE共593 bytes的SHA為559f234ab6f1b97b5c6970fc84c40982e5525fdcac06436fd355bf5ec8c8d4fa。Go解析兩個typed表面，取較小寬高，依相對位置選上下及左右方向，透過區間與列游標搬移word及奇數尾byte；影子讀寫保留重疊相依。415E只選已驗的列交接，不派送x86指令。
1120次列計畫、168次前綴、64次父表面／不同尺寸、192次完整機器及44種列／16種初始化拒絕通過。正式selector選取1040個計畫；同一來源綁定副本依Algorithmic分類後，448個為Go演算法，592個為固定相容前綴。測試原有「Go plans」訊息涵蓋所有選取計畫，固定收據已分開列出，不據此聲稱全部接管。
來源943064ee8dbb24a051b1c279da0dc1a54733a827，乾淨Go1.24.13候選a33036d07896dbeb4542730df138e16a4ecf0afa939247fb23ae66bfd4d63cba。核心、session、pod、drawblocks、前端與go vet通過；完整正常CPU／RAM、有序像素／回掃／音訊／DOS及127檔同原版，46份直接加一份分割補證同47份原版。
四路DEMO、39對介面、六對真存讀檔與十個SAVE檔同基準，槽A五檔未改。非預期閃現0，998筆F2主動切換另列；搬移出口在目前DEMO未觀測，正常111710個Go計畫／3398106步與固定2576步維持。受控比較不外推正常路線或FPS。
全部通過後原子啟用a33036d，9e45287以pod-play-nativeline保留。[固定收據](019-native-copy-runtime.json)綁定來源、測試分類及啟用；重生入口為Docker內`tools/native_fill_check.py --family copy`，原始收據在`workplace/probe/draw-intercept-v1/native-copy-v1`及`workplace/probe/g9-hd/nativecopy-{play,ui,save}-v1`。其他root、複製函式、橢圓與未驗路徑仍保留，#49開啟，全部Go閘門DRAFT。

## 2026-10-27 Go正半徑橢圓外框接入

VGA.DRV原檔SHA及IDA9.4位址沿用前節。346A–39B1共1352 bytes，SHA為4b118914e4b4a0f6286f6b8be0eb508710457be02e61eb9887f6dc2c5a9548d5。Go中點誤差與梯度生成四象限有序像素，保留重複點；35F4、37BD與37F6只選typed流程，原位置保存相容通知及模型成本。單步預算時仍由同一Go幾何重建位置，固定區塊若包含外框像素即回退至區塊前，不把x86語意繪製混入計畫。
321例有序像素、316／180次游標、24次轉換、24次初始化、912次單步像素與75種拒絕通過。72次完整Go、32次中心／半徑邊界、72次DF已設定與8次父表面前綴／機器比較同原版。96次短預算selector含3772個Go演算法、9250個固定控制計畫及63780個CPU控制步，整台SaveState同原版，沒有固定或CPU像素寫入。
來源eddd4df044d0a7426842d3301f18be0dc9aa0cb0，乾淨Go1.24.13候選d5bb3c0797ed11491b9649d0b1902eb0d11bcd0c97fa98e8ce83afa2159e917a。完整正常CPU／RAM、有序像素／回掃／音訊／DOS與127檔、47份入口證據相同；四路DEMO、39對介面、六對真存讀檔與十個SAVE檔同基準，槽A五檔未改。
非預期閃現0，998筆F2主動切換另列。新增外框出口在目前DEMO未觀測，正常111710個Go計畫／3398106步與固定2576步維持，受控比較不外推正常路線或FPS。全部通過後原子啟用d5bb3c0，a33036d以pod-play-nativecopy保留。
[固定收據](019-native-ellipse-runtime.json)綁定來源、分類及啟用；重生入口為Docker內`tools/native_fill_check.py --family ellipse`，原始收據在`workplace/probe/draw-intercept-v1/native-ellipse-v1`與`workplace/probe/g9-hd/nativeellipse-{play,ui,save}-v1`。零半徑、填滿橢圓與其他未驗路徑仍保留，#49開啟，全部Go閘門DRAFT。

## 2026-10-27 Go正半徑填滿橢圓接入

沿VGA.DRV原檔與IDA9.4位址，39B2–3DFC共1099 bytes，SHA為ee2f7dba367fe53d107b8dc0a5fc8ac1e426f7946a63699a9b2f06059d592e8c。Go中點幾何生成兩條對稱水平區間，保留重複列、左右裁切、word及奇數尾byte。3DFD以後另屬helper，沒有併入覆蓋聲明。
首段3B3D、轉換3C87、次段3CC0及八個搬移位置只選typed流程。單步搬移仍由Go重建區間與數量；固定區塊若含實際像素搬移，回退至區塊前再交Go，不派送原版繪圖opcode。
321例有序像素、316／180次游標、24次轉換與初始化、1184次單步搬移及84種拒絕通過。72次完整Go、32次邊界、72次初始DF已設定與8次父表面前綴／機器比較同原版。96次短預算selector含4150個Go演算法、6260個固定控制計畫及54510個CPU控制步，整台SaveState相同，沒有固定或CPU實際像素搬移。
來源e16b4d4796444a93cc1f11d573082f81d6545d3e，乾淨Go1.24.13候選5034a1cc1220ce1e90094d0069c55edd56b7fac7d3d559cac2bf426776eec87a。完整正常CPU／RAM、有序事件、127檔與47份入口證據相同；四路DEMO、39對介面、六對真存讀檔與十個SAVE檔同基準，槽A五檔未改。非預期閃現0，998筆F2主動切換另列。
新增填滿橢圓出口在目前DEMO未觀測，正常111710個Go計畫／3398106步與固定2576步維持，不外推正常路線或FPS。全部通過後原子啟用5034a1c，d5bb3c0以pod-play-nativeellipse保留。[固定收據](019-native-ellipsefill-runtime.json)綁定來源及啟用，重生入口為Docker內`tools/native_fill_check.py --family ellipsefill`，原始收據在`workplace/probe/draw-intercept-v1/native-ellipsefill-v1`及`workplace/probe/g9-hd/nativeellipsefill-{play,ui,save}-v1`。
51CC的資料標記為BMHD／CMAP／BODY；後續研究將列輸出與來源準備分開記錄，直接色號列已有80次原版比較，尚未接產品。其他圖像與未驗路徑仍保留，#49開啟，全部Go閘門DRAFT。

## 2026-10-27 Go圖像列輸出接入

原檔VGA.DRV之51CC–5413共584 bytes，SHA-256為dd209e837a213ee07c0b9602563167e0eed60702d8a96f4c20d98f27ebb157a2，沿既有IDA9.4匯出與位址基準。Go從typed來源列直接搬移色號，或組合八個位元平面的目前bit；5373初始化、53C5分組及53FC推進保留ABI與模型成本。一步預算仍依同一來源重建資料，固定像素區塊回退後交Go，不派送原版繪圖opcode。
80次直接列、80次平面初始化、336次平面分組、4576次單步輸出與160次列推進，同原版CPU／完整RAM／有序寫入及兩成本。400次完整呼叫含兩種列形狀、壓縮／直接來源及五種預算，32492個Go演算法、154444個固定控制計畫與499008個CPU控制步分列，整台SaveState相同，沒有固定或CPU實際像素輸出。另有64次邊界／裁切／DF及150種錯誤狀態拒絕通過。
來源519ac6389d5264f628f52745efc3140e3769efd5，乾淨Go1.24.13執行檔2e45f31d19263699fe7df5ed0053c9365923108f72e191c9949dd87ae54b689e。全部繪圖、核心、session、pod、前端與go vet通過。完整正常狀態、有序事件、127檔與47份入口同原版；四路DEMO、39對介面、六對真存讀檔及十個SAVE檔同基準，槽A未改。非預期閃現0，998筆主動F2另列。
新圖像列出口在目前DEMO未觀測，正常111710個Go計畫／3398106步與固定2576步維持；受控比較不當新玩家路線或FPS證據。全部通過後原子啟用2e45f31，前版5034a1c保留。[固定收據](019-native-bitmap-runtime.json)綁定來源與啟用；重生入口為Docker內`tools/native_fill_check.py --family bitmap`，原始收據在`workplace/probe/draw-intercept-v1/native-bitmap-v1`與`workplace/probe/g9-hd/nativebitmap-{play,ui,save}-v1`。
BMHD／BODY查找、壓縮來源列準備與5196的CMAP資料轉換保留原版，不宣稱Go檔案解析器。5615–5885整數縮放原型的125組像素與250次初始化比較通過；列準備尚有DOSBox成本與有序寫入差異，保留失敗log並繼續研究，未接產品。全部Go閘門仍DRAFT，#49開啟。

### 縮放研究成本勘誤

前段「有序寫入差異」是失敗訊息將原版展開寫入與計畫未展開清單並列造成的誤讀。實際完整有序寫入相同，失敗是列準備漏算每次LODSB的DOSBox額外成本。輸出另外少計RETF的遠返回成本；依既有CPU的base2、cycFar18及兩次stack讀取4訂正。沒有調整CPU，也沒有以常數或放寬比較補過。
修正後930次列準備、930次列輸出、10488次單步、250次完整Go及144次邊界／八層父表面同原版；1000次短預算完整交接含38610個Go演算法、34268個固定控制計畫及320562個CPU控制步，沒有固定或CPU的縮放scratch／畫面像素寫入。88種錯誤狀態拒絕通過，規格升READY後接開發分支；正式產品仍519ac63／2e45f31，未完成整批驗收前不替換。

## 2026-10-27 Go表面整數縮放接入

VGA.DRV的5615–5885共625 bytes，SHA-256為6d9aac63d38f8b0e5ddd9bdbdfcfdda9144318a34398436b6b468c52f0271106，沿既有IDA9.4位址基準。Go以置中整數取樣格處理補色、重複樣本與重複列，保留來源重疊及額外scratch寫入。短預算仍由同一取樣格重建資料，固定像素區塊回退，沒有opcode派送。
正式來源f4f2838797bf1f967ceb7d98f58f8d029ac997bd，乾淨Go1.24.13執行檔0abeee0e50bd4f407150f4e671979f828f5a3a0a42944e68993ccd97cc1186e9。250次完整Go、1000次短預算、144次邊界／父表面及88種拒絕同原版，計畫分類及初版成本勘誤見規格025與前段，不把固定控制前綴當Go演算法。
全套drawblocks 113.541秒、核心、session、pod、前端與go vet通過。完整正常CPU／RAM、有序事件、127檔與47份入口同原版；四路DEMO、39對介面、六對真存讀檔與十個SAVE檔同基準，槽A未改。非預期閃現0，998筆主動F2另列。新增縮放出口在目前DEMO未觀測，正常111710個Go計畫／3398106步與固定2576步維持，不外推正常路線或FPS。
全部通過後原子啟用0abeee0，前版2e45f31保留。[固定收據](019-native-scale-runtime.json)綁定來源及啟用；重生入口為Docker內`tools/native_fill_check.py --family scale`，原始收據在`workplace/probe/draw-intercept-v1/native-scale-v1`及`workplace/probe/g9-hd/nativescale-{play,ui,save}-v1`。
5886原地翻轉已在獨立研究副本通過完整原版與短預算比較並達READY，尚未接產品；同列倒序的中途讀寫相依與奇數高度中間列已由原版反例確認，不能用一般圖形庫取代。其他轉場、字形與未驗入口仍保留，全部Go閘門DRAFT，#49開啟。

## 2026-10-27 Go表面原地翻轉接入

VGA.DRV的5886–5A27共418 bytes，SHA-256為2ff20b76959b7f3972f76530c6142c8337d0f935893752f9a23d417dde2eac36，沿既有IDA9.4位址基準。Go以typed表面與成對列處理倒序取樣、暫存與兩階段回寫。同列倒序的中途讀寫相依與奇數高度中間列皆保留原版；首版整列預取假說由寬2高1反例否定，沒有改原版來配合模型。
正式來源5a87f57d87e5f5afbf9bc03444a6fc1d3b6d7af8，乾淨Go1.24.13執行檔dc8b977ceb420e128241bda019f0ede1d75f52d322eb90913d5bb8a42e0e1fc5。160組像素、320次初始化、656次成對列與656次位址、63952次單步、320次完整Go、160次短預算、64次邊界／父表面及174種拒絕通過，固定與CPU沒有翻轉scratch／畫面寫入。
全套drawblocks 178.167秒、核心、session、pod、前端與go vet通過。完整正常狀態、有序事件、127檔與47份入口同原版；四路DEMO、39對介面、六對真存讀檔與十個SAVE檔同基準，槽A未改。非預期閃現0，998筆主動F2另列。新翻轉出口在目前DEMO未觀測，正常Go計數與固定2576步維持，不外推新正常路線或FPS。
全部通過後原子啟用dc8b977，前版0abeee0保留。[固定收據](019-native-surfaceflip-runtime.json)綁定來源及啟用；重生入口為Docker內`tools/native_fill_check.py --family surfaceflip`，原始收據在`workplace/probe/draw-intercept-v1/native-surfaceflip-v1`及`workplace/probe/g9-hd/nativesurfaceflip-{play,ui,save}-v1`。
4337像素轉場已在獨立研究副本通過44次完整機器比較，尚DRAFT。等待採既有Go服務契約；研究副本先因函式名稱缺項，再因舊原型helper同名而建置失敗。只替副本服務加獨立前綴後乾淨重跑通過，沒有調整產品或CPU。其他轉場、字形與未驗入口仍保留，#49開啟。

## 2026-10-28 Go像素轉場與HD逐格呈現接入

來源為VGA.DRV，SHA-256 `eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f`。
IDA9.4以檔案位址基準0匯出4337–45DF共681 bytes，切片SHA-256為
`6888a66a54dd072b40f5c95f990d81d53f75257852bce465c95dff15e1168796`。
原始sub_4337名稱、入口表slot29及執行期48A1:4337保留，受控語意已證實。
Go從typed表面與32位LFSR計算樣本，逐次保留重疊讀回及原有寫入順序；沒有在像素循環派送guest opcode。
共享等待與CPU呼叫控制另列，1×1不返回的原版行為只驗固定一萬步區間，不修改終止條件。

正式來源b11290cb7d8754decf0924e1513b48aa17a783ea，乾淨Go1.24.13執行檔
`7d51351c24d07ab6f1ec97d4a3dde7a663670dbec118a47843a246610b59f0af`。
6份演算法與11份測試通過原版比較；44次完整比較分列3028個Go轉場、1404個共享等待與2808個CPU控制步。
正式40次短預算為39682個Go計畫含等待、7148個固定控制與50084個CPU控制步。
原型39378的計數保留，來源對照、DF、完整畫布／父表面、拒絕與有界1×1結果見[規格025](../spec/025-direct-render-commands.md#dissolve-4337-conformed)。

完整機器CPU／RAM、有序像素、回掃、音訊、DOS事件、127檔及47份入口同原版。
四路正常demo、39對介面、六對真存讀檔與十個SAVE檔同前版，槽A五檔未改。
正常demo未觀測4337，新入口只宣稱受控條件；正常111710個Go計畫／3398106步及固定2576步不當全遊戲覆蓋率。

同次接入HD辨識快取，只重用相同來源索引與完整色盤的辨識結果，呈現保留仍逐格提交。
明確按D的新路線證實局部CPIC102需要四格，三格上限提早退回一次。
完整逐格負對照重現1次，四格修正後直接／舊合成各11格均為0，差異只在一格一幀；
未知來源、色盤改變、非戰場、2倍、F12與第五格回退通過。正常HD為0，主動F2的998筆另列。
同一6640格快取前後的CPU／RAM、索引、色盤與共有截圖相同，耗時392.7與188.3秒僅為本次共享主機觀察。

全部通過後原子啟用7d51351，前版b484851以pod-play-demo-control保留，776檔主題不變。
demo對話、戰鬥中斷、再次播放及正常進入的終點、截圖與中斷計數同原正式版。
[正式固定收據](019-native-dissolve-runtime.json)包含啟用驗證；[021](021-hd-frame-presentation.json)保留READY研究，未覆寫歷史證據。
重生入口為Docker內`tools/native_fill_check.py --family dissolve --run-id v2 --frontend-baseline demo-control-baseline`。
原始資料在`workplace/probe/draw-intercept-v1/native-dissolve-v2/`、`workplace/probe/g9-hd/nativedissolve-{play,ui,save}-v2/`
與`workplace/probe/dissolve-demo-control-v2/`；完整重播命令及環境見WORKLOG與各入口工具。
其他色盤控制、字模、root與未驗條件仍保留，全部Go閘門DRAFT，#49保持開啟。
## 2026-10-28 有界共享跨度與RLE續行

來源 b11290c 的唯讀盤點確認，正常探針的 2576 個固定語意步中，
共享清除 1D1F:046E、共享複製 1ED4:1A5C 及 RLE 68F4／6B33
會在短預算後以固定區塊寫出像素。完整工作原本已有 Go 跨度，缺口在中途接續。
原始位址、完整輸入與區段 SHA-256、已證實範圍及拒絕條件見
[規格025](../spec/025-direct-render-commands.md#bounded-memory-ready)。

以穩定遠呼叫框架重建共享工作，以壓縮來源與游標重建 RLE 封包。
Go 產生自有資料的複製／填色跨度，ABI 檢查點保留排程中斷及有序通知，
不解碼 guest opcode，也不呼叫 CPU.Step 或固定 effect.run。
清除、複製與 RLE 共 20172 次有界原版比較，另有 440 次完整 SaveState、
4 次整張畫面跨度與 18 種拒絕通過。早期 RLE 測試缺 fixture 與 CS 的 log 保留，
補上明確輸入及非零比較護欄後重跑；沒有改產品行為來通過測試。

依 READY 契約接入候選 b97bc21，執行檔 SHA-256
`913a0291f307d284ea16e6b2e9b1c59530602e912f3d669a51b08b30a0db2b75`。
正式 selector 的原版比較通過；完整 drawblocks 354.368 秒、machine／session／pod／前端及 go vet 通過。
正常機器探針的完整 CPU／RAM、有序視訊、音訊、回掃及 DOS 事件同原版。
Go 計畫由 111710 增至 112012，演算法步數為 3400397，固定語意由 2576 降至 1062。
這是已走路線的分類，不能當作全遊戲覆蓋率或 FPS。

研究原型與 log 在 clone 的 `workplace/native-bounds-research/`、
`workplace/native-bounds-*-tests*.log`，正式測試在 `apps/pod/drawblocks/native_bounded*_test.go`。
完整機器與 GUI 收據根為 `workplace/probe/draw-intercept-v1/native-bounded-v1/`
及 `workplace/probe/g9-hd/nativebounded-{play,ui,save}-v1/`。
四路明確按D的正常demo、39對介面與六對真存讀檔全部通過；十個SAVE檔同基準，槽A五檔不變。
HD閃現0；demo選單不逾時、對話與戰鬥中斷、重播與返回後正常進入同固定020。
總檢查通過後原子啟用913a029，前版7d51351以pod-play-nativedissolve-v2保留。
固定[019-native-bounded-runtime.json](019-native-bounded-runtime.json)綁定來源、驗收與啟用。
21種剩餘固定形狀另列，整體Go接管仍DRAFT。

下一批變寬字模5AA4的30組原色／映射有序像素研究已通過；詳[規格025的DRAFT](../spec/025-direct-render-commands.md#變寬字模5aa4)。
這是原版規則研究，尚未接入Go實作，不當成正常玩家路線完成。
