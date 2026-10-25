# 434 mode13h原生文字串繪圖命令

狀態：CONFORMED，限本節8／16位mode13h契約與已列正常抽樣；原生功能預設仍關閉，由遊戲層明確選用。

## 原有契約與範圍

`internal/cpu/string.go`把單一REP內的迭代一次完成，`Machine.Step`才是事件／指令預算邊界。
MOVS／STOS不改旗標；R[SI]／R[DI]依DF與寬度變動，REP消耗CX至0，非REP不動CX。
`cycles.go`的既有每次string成本及DOSBoxCost每次1週期均須保留；此為既有模擬器契約，不宣稱硬體逐週期精確。

新增預設關閉的`CPU.NativeStrings`與可選`StringCommandBus`，只處理8／16位MOVS、STOS；CPU保留prefix解碼、IP與原指令成本。
bus收到來源段／位移、目的段／位移、寬度、次數、方向及填色值，成功時CPU按原運算更新暫存器及迭代週期，失敗時走原`stringOnce`。
32位／地址大小擴充、段內環繞、非mode13h、planar／ModeX、來源MMIO及相交的讀取監看回既有路徑。
監看僅在實際來源讀取或目的寫入範圍相交時回退；與命令不相交的遊戲狀態守衛保留，不阻擋原生指令。
前後向均支援；非交疊普通RAM來源方能先複製命令payload，未知或可能別名的來源不批次讀取。

## 機器與命令

命令為COPY／FILL／PIXEL，含step、原CS:IP、目的位移、寬度／次數、方向及複製的色號payload。
目的僅A000的可見64000 bytes，資料用現有Write8維持完整RAM、原逐點VideoWrite與write watcher語意，禁止漏掉同值寫入。
新增命令通知只改前端呈現輸入，不回傳或暴露Machine給observer。payload不與客體RAM共用，通知者修改它不改客體。
既有非原生路徑的可見寫入仍發PIXEL；原生操作中抑制重複PIXEL命令，原逐點通知照舊。
session以預設關閉的Config選項開啟，另有可選命令observer，錯誤／panic沿既有觀測器錯誤閘門處理，生命週期結束撤除callback。
模式切換依既有ModeObserver處理，不猜清除；原版後台索引資料仍可供讀回。

## 驗收

原指令／原生指令成對比較8／16位、REP／非REP、0／1／奇偶數、DF兩向、段覆寫、三個CPU model與兩種週期模型，
逐項比CPU、flags、cycles、完整RAM、逐點讀寫通知及輸出命令。反向與非交疊來源、同值寫入、邊界／環繞／監看／MMIO拒絕分開驗。
原生與原路徑在相同指令事件邊界比較計時器、按鍵、回掃及音訊事件；不跳過額外Machine.Step。
新功能關閉時原有CPU／machine／session測試不變；不放作品名稱或位址於本規格與通用程式。

## 驗收紀錄

程式來源`c0fc72b23476d695e2ca46d3431cf2fa9a9d54b6`，Go1.24.13。CPU／machine／session／xlate及前端測試通過。
原生與原指令的model、週期、方向、前綴、零次數、監看範圍、來源隔離、回復寫入及計時事件成對核對通過。
正常7400更新、8633次回掃的十個實際出口，共3343240次byte stores全數原生，命令畫布每回掃與原版索引相同。
完整CPU／RAM／色盤／索引、音訊事件、DOS請求與返回及檔案核對；另以固定檔案時間比較十份完整RAM與暫存器／堆疊，未遮蔽差異。
具體遊戲定位、輸入條件與機器可讀收據保存在消費端研究019，不將遊戲常數寫入本通用層：
[固定收據](https://github.com/wicanr2/pools-of-darkness-cht/blob/main/docs/re/019-native-runtime.json)。
