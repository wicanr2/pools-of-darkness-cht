# 429 — 執行中改時脈與 Turbo Pascal Delay 快轉

狀態：CONFORMED（2026-10-22，§2.1–§2.4；§2.5 DRAFT 未實作）。前端驗收見 death-knights-of-krynn-cht 規格 016 §4.1
前置：[`242-machine-clock-scale-draft`](242-machine-clock-scale-draft.md)、[`238-session-step-observer`](238-session-step-observer-draft.md)、
[`240-opl-audio-output`](240-opl-audio-output-draft.md)。需求來源：death-knights-of-krynn-cht #58。

## 1. 為什麼要做

使用者要求前端提供 CPU 模擬速度 1x 到 32x，遊戲時鐘不變，HD 下維持 30fps。242 只允許開機前設定、範圍 10–100。

本機量測（已證實，2026-10-22，dkk-play、Docker＋Xvfb、2 核、主機另有負載約 10）：3000 畫格 demo 段，每格平均耗時：

| 設定 | `Advance` | 合成 | 上傳 |
|---|---|---|---|
| p=50、3 倍＋HD | 19.3 ms（p50 15.2） | 7.6 ms | 0.5 ms |
| p=100、3 倍＋HD | 28.2 ms（p50 25.7） | 5.0 ms | 0.3 ms |

模擬是主要成本：不掛觀測器的 probe 約 85 ns／步，dkk-play 約 150–200 ns／步。

《Death Knights of Krynn》3D 行走與訊息停留時，30 萬步軌跡有 99% 在 Turbo Pascal 7 CRT 的 `Delay`（已證實，`-trace`）：

```
Delay:     mov dx,[ss:bx+4] / or dx,dx / jz done / mov ax,40h / mov es,ax / xor di,di / mov al,es:[di]
perMs:     mov cx,[DelayCnt] / call DelayLoop / dec dx / jnz perMs
DelayLoop: mov bx,4 / dec bx / jnz $-1 / cmp al,es:[di] / loope DelayLoop / ret
```

`DelayCnt` 在開機時依 CPU 速度校準（本作 p=100 為 1051）。等待因此是計數迴圈，不是服務輪詢。
dkk-play 開機即用 p=100 與 p=50 跑到同一畫格：第 2500 格畫面相同；第 4000 格同一場景、差 112 像素（圖片動畫相位），
因為讀取等運算段落提早完成。開機校準下遊戲時間不隨 CPU 速度改變（已證實）。

## 2. 契約

### 2.1 machine：執行中改時脈

- `Machine.SetClockPercentLive(p int) error`：任何時刻可呼叫，`p` 須在 10–1600；`CycleClock` 為真時回錯誤。
- 以舊比例 `q` 換算：`IRQ0Base = ScaleSteps(DefaultIRQ0Every, p)` 並依目前分頻 `recalcIRQ0`；`VGAFrameEvery`、`KeyEvery` 依預設值縮放；
  `nextIRQ0`、`nextFrame`、`nextKey` 的剩餘距離（相對 `Steps`）乘 `p/q`，至少 1；`nextKey` 為 0 時維持 0。`p == q` 不改任何欄位。
- 已登記 TP Delay（§2.2）時，`DelayCnt = round(登記值 × p / 登記時比例)`，以登記值為基準，不累積捨入誤差；結果夾在 1–65535。
- `SetClockPercent`（開機前、10–100）不變。快照與存檔沿用 242 的欄位，不改格式。

### 2.2 machine：TP Delay 精確快轉

- `Machine.SetTPDelay(loop, count uint32)`：`loop` 為 `DelayLoop` 線性位址，`count` 為 `DelayCnt` 線性位址；兩者為 0 表示取消。
  登記時記下 `DelayCnt` 現值與時脈比例。
- `Step` 在執行指令前，若 `CS:IP` 的線性位址等於 `loop`、位址上的 12 位元組等於 `BB 04 00 4B 75 FD 26 3A 05 E1 F5 C3`、
  `ES:DI` 的位元組等於 `AL`、`CX ≥ 2`，就一次完成 `k` 圈：
  - `k = min(CX − 1, ⌊(E − Steps) / 11⌋)`，`E` 為下一個事件的步數：`nextIRQ0`、`nextFrame`（`VGAFrameEvery>0`）、`nextKey`（佇列非空）、本回合上限 `RunLimit`。
  - `k ≥ 1` 時：`CX −= k`、`BX = 0`、`Steps += 11k`，旗標設為 `cmp al,al` 的結果（ZF、PF 為 1，CF、SF、OF、AF 為 0），IP 不變。累加 `DelayAccelerated += 11k`。
  - 每次抵達 `loop` 時記下 `Steps` 與 `CPU.Cycles`；前一次抵達恰好是 11 步前時，兩次的週期差即一圈週期數 `c`。
    還沒量到 `c` 時不快轉；快轉時 `CPU.Cycles += c×k`，與逐條執行相同。
  - 有排定回呼（`cbQueue`、`periodic`、`timedKeys`）、斷點、讀取監看、`TraceSegs`／`WatchDSOn`、`Coverage` 或 `CycleClock` 時不快轉。
    寫入監看不影響（迴圈不寫記憶體）。
- 較舊版 CRT（Secret of the Silver Blades 1990 實測，ES:DI＝0000:0000）的 DelayLoop 是 `26 3A 05 75 02 E2 F9 C3`
  （`cmp al,es:[di] / jne done / loop DelayLoop / done: ret`），每圈 3 步、不動 `BX`。`SetTPDelay` 登記時依 `loop` 上的位元組
  判斷寫法；兩種都不符時不快轉。快轉規則同上，`11` 換成每圈步數，TP5 寫法不清 `BX`。
- 等價性：迴圈內只讀 `ES:DI`，沒有寫入；`k` 圈在下一個事件前結束，所以每個中斷與回合邊界上的暫存器、記憶體、`Steps` 都與逐條執行相同。
  跳過的指令不呼叫觀測器。
- `RunUntil`、`RunUntilObserved` 把快轉的步數計入迴圈計數；`Machine.RunLimit` 由呼叫端設為本回合結束步數，0 表示不快轉。

### 2.3 session

- `Owner.SetClockPercent(p int) error`：回合之間呼叫 `SetClockPercentLive`。
- `Owner.SetTPDelay(loop, count uint32)`：回合之間登記。
- `Advance` 開頭設 `RunLimit = Steps + budget`，結束清 0。`TickReceipt` 新增 `Accelerated`（本回合快轉的步數）。

### 2.4 runtime/tpcrt 與 audio/mixer

- `tpcrt.FindDelay(read func(uint32) uint8, ds uint16) (loop, count uint32, ok bool)`：在 0–A0000 找 `Delay` 簽章
  `8B DC 36 8B 57 04 0B D2 74 14 B8 40 00 8E C0 33 FF 26 8A 05 8B 0E ?? ?? E8 06 00 4A 75 F6 CA 02 00` 恰好一處；
  `loop` 為簽章 `+0x21`，`count` 為 `ds×16 +` 簽章 `+0x16` 的字組。也找較舊版的簽章
  `8B DC 36 8B 57 04 0B D2 74 13 33 FF 8E C7 26 8A 05 8B 1E ?? ?? 8B CB E8 06 00 4A 75 F8 CA 02 00`（`loop` 為 `+0x20`、`count` 位移欄在 `+0x13`）。
  兩種寫法合計不唯一或 `loop` 位元組不符就回 false。
- `Mixer.SetStepsPerSecond(rate, atStep uint64)`：以 `atStep` 為接點換速率，前後取樣位置連續。

### 2.5 服務輪詢跳過（DRAFT，未實作）

選單等待時遊戲以 `int 16h AH=01`、`int 21h AH=2Ch` 輪詢。是否需要另做非精確的空等跳過，待 #58 量測選單畫格的主機成本後決定。

## 3. 不做什麼

- 不改 CPU 核心與觀測器逐步回呼的成本。
- 沒有登記 TP Delay、比例不變時，行為與前一版完全相同；probe、收據與 oracle 不受影響。

## 4. 驗收

1. 單元測試：`SetClockPercentLive` 換算、剩餘距離、`p==q` 不變、範圍與 `CycleClock` 錯誤、`DelayCnt` 依基準重算；
   TP Delay 快轉：合成程式逐條執行與快轉在每個 `RunUntil` 邊界的暫存器、記憶體、`Steps` 相同，事件前停下、
   `CX` 剩 1 時不快轉、`ES:DI` 不等於 `AL` 時不快轉；`RunUntil` 差分等於預算；`tpcrt.FindDelay` 的命中與拒絕；mixer 換速率前後連續。
2. 既有 machine、session、mixer 單元測試全過。
3. 實機等價性：dkk-play 1x 同一腳本，快轉開與關跑到同一畫格，記憶體與畫面雜湊相同（death-knights-of-krynn-cht 規格 016）。
