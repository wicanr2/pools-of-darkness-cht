# 436 CALL後的有界主機效果

狀態：CONFORMED，含16位CALL後效果、可重入區塊、緊湊搬移及已列驗收。參考mm3已驗HostWorkBudget／CommitHostWork方法，依本分支Step、CALL、IRQ及RunUntilObserved契約重新審查後達READY再實作。

## 範圍與控制權

新增預設關閉的16位CALL後通知，涵蓋9A、E8及FF /2、/3。正常CALL的堆疊、暫存器、IP、週期先完成，才通知持有機器的Owner。
CALL分支不在普通指令增加資格輪詢，不處理32位CALL、硬體中斷或非CALL轉移；可選指令前區塊入口另見下文。
可選入口表先篩選CALL，表以值複製、撤除callback時一併清除；空表維持全部CALL通知，適配器仍必須驗證真正入口。
前端的可選HostCallProvider只取得限期唯讀StepView及預算，回傳完整暫存器／旗標、最後指令定位、步數／週期與有序RAM byte writes。
每次寫入附相對步數與原CS:IP，語意通知的定位不代表該opcode重新執行；寫入者不能取得Machine。

## 預算及提交

HostWorkBudget取呼叫者上限、RunLimit、下一IRQ0、畫面及periodic callback的最小距離。
CPU指令已有fault時不發CALL通知；CycleClock、halt／TF、中斷、排隊呼叫／鍵盤、定時鍵、逐步或未驗觀測模式均拒絕。
此分支沒有mm3版本的MPU-401或待報裝置Fault欄位，不沿用該欄位名稱；將來新增延遲裝置錯誤時須同步擴充`HostWorkBudget`的拒絕條件。
程式斷點、word watch、讀取監看及其他不支援模式拒絕；byte write監看與實際寫入範圍相交時拒絕。
完整計畫放不進預算就不寫入。計畫要求來源與CPU效果另由適配器按原版證據驗證，共通層不猜遊戲語意。

持有者先唯讀檢查全部位址、順序、步數、溢位及監看，再對成功接管的入口送原BeforeStep一次，最後提交完整效果。
寫入只限普通RAM及mode13h線性可見表面，不寫ROM、HMA或有副作用的平面顯存。
原版VideoWrite及PIXEL依寫入原時間／指令定位送出，相容RAM與寫入後通知順序照舊；拒絕時沒有部分提交。
CPU保留未被計畫涉及的模型、32位高位、裝置與設定，最後指令定位更新為已驗效果；週期不當作硬體逐週期精確。
計入原函式虛擬步數及週期，沿既有accelCredit讓runner消耗同等配額。恰好到期的事件在下一Step處理，不能越過事件。

## 觀測生命週期與驗收

session需明確啟用並驗provider，callback僅Advance內有效，結束後撤除。Prepare、入口通知及寫入callback的panic或錯誤沿observer fault封鎖。
寫入通知的錯誤在已接受的完整原子批次後封鎖；不宣稱與逐指令通知故障的中途RAM相同。未支援的除錯／監看資格不接管。
失配回原路徑時不得先呼叫一次BeforeStep再重複通知；只有已驗可提交的計畫才通知入口。
通用測試比原CALL／主機效果的CPU、完整RAM、步數、週期、有序通知及恰好／不足的事件與回合配額；測近遠CALL、零／溢位、未知mode、監看及callback撤除。
消費端須另有獨立原版呼叫oracle、正常玩家路徑、同狀態RGB／存檔及包含相容寫入的成對成本。本規格不含任何遊戲常數。

## 驗收

程式`64e734c92ecaa096be9bb91d6e6f792fd5fad7fd`，Go1.24.13。近遠／直接間接CALL、事件恰好及不足、回合配額、入口表值複製、模式／監看／溢位拒絕及observer錯誤／撤除測試通過。
消費端32個真實入口在三模型／兩週期的192組完整狀態與有序寫入相同，另驗16組邊界、空返回及失配對照。
正常機器重播的全部像素、回掃、音訊、DOS操作與CPU／RAM一致；正式視窗四路、讀檔10對畫面與正常新存B／冷讀B六對畫面及存檔位元組相同。
成對完整CALL成本包含資格檢查、計畫、相容寫入及通知，約4.2–4.9倍，不換算全遊戲FPS。
具體遊戲位址、輸入、雜湊與正常驗收入口保存在消費端[研究019](https://github.com/wicanr2/pools-of-darkness-cht/blob/main/docs/re/019-draw-interception.md)與[固定收據](https://github.com/wicanr2/pools-of-darkness-cht/blob/main/docs/re/019-whole-call-runtime.json)。

## 可重入繪圖區塊擴充（CONFORMED）

新增預設關閉的NativeHostBlocks。持有者於Machine.Step先依原版順序處理tick與鍵盤，再於真正指令執行前按值複製的程式範圍篩選。
provider仍只取得限期StepView與HostWorkBudget，回傳相同的純效果計畫。拒絕時原指令照常，接受時本Step以該計畫的原虛擬步數計入配額。
下一步先處理恰好到期的事件，區塊續接只看事件後的實際CPU位置；無需保留可變的適配器續傳狀態。
通用層不含遊戲位址、原碼或解碼器。HostBlocks與BlockInstructions獨立統計，不重算HostCalls。
回呼僅在Advance存在，provider錯誤或panic於執行原指令前封鎖；範圍／view不得外洩；byte監看、部分拒絕、取消與時間預算沿上文契約。
不補送被接管內部指令的BeforeStep；適配器必須在來源規格明列被跳過的觀測呼叫者，盤點模式改走原版。
同一原指令、相同步數與CS:IP的連續byte writes可合成擁有資料的COPY／FILL命令，逐byte的VideoWrite與相容RAM順序保持原樣。
不跨指令合併；向後word維持每個word低byte在前。未連續、單點或未知序列沿PIXEL。命令Host標記與HostBytes只記已提交效果，不記資格提案。
驗收須含預設關閉、錯誤／拒絕原子性、事件先於計畫、回合邊界、回呼撤除、範圍值複製及正常消費端獨立原版串流。

緊湊HostTransfer（CONFORMED）與Writes按原步數合併，兩種類型不能佔同一指令。每個transfer是段內不環繞的普通RAM／mode13h線性表面搬移或填色，寬1／2，次數非零。
驗證來源長度、目的範圍、模式、觀測器、時間及順序後才寫入；提交先複製提案資料，回呼不能修改後續來源。
逐byte VideoWrite維持原順序與時間，前端只收到一次擁有資料的COPY／FILL。區塊計畫可以含完整已驗前綴，未知後續區塊由適配器退回其原入口。
任意RunUntil predicate會停用主機效果，不能跨過未宣告的停止條件。範圍可加精確CS篩選，避免overlay重用實體位址時支付提案成本。

StepView.CopyRAM只複製0–A0000以下的普通RAM到呼叫者緩衝區，不回傳客體slice，期限與Read8相同。超出範圍返回false；畫面、ROM與HMA仍由既有介面處理。

擴充程式`191479d5492fdce1e5d925ea44033d2112f75016`。核心、生命週期、事件／回合／predicate、範圍值複製、拒絕、資料壽命與前後向word測試通過。
消費端12010組獨立原版區塊、12組真實完整呼叫與192組邊界通過；正常機器全部有序事件及完整CPU／RAM、四路DEMO、39對介面與6對新存讀檔相同。
具體資料與限制見消費端[全部繪圖收據](https://github.com/wicanr2/pools-of-darkness-cht/blob/main/docs/re/019-all-render-runtime.json)。源程式、工具及固定收據不含原版EXE或原始RAM。

## 函式演算法診斷分類（READY）

HostWorkPlan新增預設false的Algorithmic標記，僅供計數。消費端須先以獨立原版證據及來源審查確認函式層級演算法，才標true；固定指令語意區塊維持false。
提交成功後Machine分別累加Algorithmic計畫及虛擬指令數，Owner每回合收據傳遞差值；既有HostCalls／HostBlocks與原步數仍原樣累加。
拒絕或只有提案時不計數。標記不改資格、預算、事件、CPU、RAM或提交順序，也不自動證明消費端演算法正確。
此計數是既有HostWork總數的子集，不能把非繪圖的普通CPU指令算成繪圖回退；實際回退仍以消費端已觀測入口說明。
驗收須含預設false、true／false分流、拒絕不累加及回合差值，另由消費端正常路徑證明分類有使用。
消費端的幾何、來源與Go演算法契約見主專案規格025的48A1:41AF小節，通用層不保存其位址。
