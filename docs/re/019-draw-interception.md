# 019 直接繪圖攔截

日期：2026-10-25。Issue：[#45](https://github.com/wicanr2/pools-of-darkness-cht/issues/45)。狀態：研究中，正式呈現尚未變更。

## 目前狀態

| 項目 | 本輪基準 |
|---|---|
| 正式來源 | dosgolem/pod `f77eac3fdea0398ca9913028c1fba174832b1f52` |
| 已驗執行檔 | SHA-256 `1c676d2bb983c1bf4b957f3cc691ea0635e77acd881d083c8218b24133c98bd7` |
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
