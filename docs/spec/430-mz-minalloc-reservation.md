# 430 — MZ 的 e_minalloc 算進程式佔用的記憶體

狀態：CONFORMED（2026-10-23；§5 第 1、3 項單元測試與全套測試，第 2 項收據見 pools-of-darkness-cht `docs/re/003`、`tools/g2-startup-receipt.sh`：
`VACE.EXE` 讀到 `INTRO.ACF` 並播完片頭）。需求來源：pools-of-darkness-cht #3（G2）。
取代 [`007-exec-ems`](007-exec-ems.md) §6「MZ 的 MINALLOC／MAXALLOC 配置語意：忽略」與「父程式記憶體保護：無」兩列中，
與 MINALLOC 有關的部分。MAXALLOC 仍不做（§4）。

## 1. 為什麼要做

載入器把 MZ 程式的 `EndSeg` 算到映像結尾為止（`loadEXEAt`），可配置區 `FreeSeg` 與 EXEC 子程式的落點都從那裡開始。
MZ 檔頭 `e_minalloc`（offset 0Ah）是映像之後程式至少要的段數，Turbo Pascal 把未初始化資料與堆疊放在這一段。
不算進去，EXEC 會把子程式疊在父程式的堆疊上。

觀測（已證實，dosgolem 收據，pools-of-darkness-cht `workplace/probe/g2-control-*`）：《Pools of Darkness》v1.10 的 `CONTROL.EXE`

| 欄位 | 值 |
|---|---|
| 映像 | 3760 bytes＝0xEB 段 |
| `e_minalloc`／`e_maxalloc` | 0xB9／0xB9 |
| SS:SP（相對載入段） | 0124:0800，堆疊頂＝載入段＋0x1A4＝映像＋`e_minalloc` |
| 載入段 | 0110（`PSPSeg` 0100＋10h） |

- `EndSeg`＝0110＋0xEB＋1＝01FC，子程式 MCB 落在 01FC、PSP 01FD。
- `#1498` `AH=4Bh AL=00`，參數區 `ES:BX`＝0234:04E2，在父程式堆疊段（SS＝0110＋0124＝0234）裡。
- 子程式 `VACE.EXE` 載入時寫進 01FD 起的位址，蓋掉了參數區與它指向的命令列尾巴；`copyCmdTail` 讀到的是子程式的 bytes，
  `VACE.EXE` 報 `File not found "<亂碼>.ACF"` 後以 exit 1 結束。

真 DOS 配給程式的塊至少是 PSP＋映像＋`e_minalloc`（DOS 載入規則；`e_minalloc` 不足時 EXEC 回 8「記憶體不足」）。

## 2. 契約

1. `parseMZ` 讀出 `e_minalloc`（u16，offset 0Ah）。
2. `loadEXEAt` 的 `EndSeg` ＝ 載入段 ＋ 映像段數 ＋ `e_minalloc` ＋ 1。頂層 `LoadEXE` 的 `FreeSeg` 與 EXEC 後父程式的可配置游標
   都因此跳過程式的未初始化資料與堆疊。
3. 新增 `MZLoadParags(data)`＝映像段數＋`e_minalloc`。EXEC 的需求量 `imageParags` 改用它，配置器找洞、`carveBlock` 與
   `WriteMCB` 的長度一致。`MZImageParags` 的語意（只算映像）不變。
4. `.COM` 不受影響（仍擁有整個段）。
5. 子程式的 PSP:0002（塊頂）規則不變（`205-spawn-arena-placement`）。

## 3. 相容性

- 改變所有 MZ 程式載入後的 `FreeSeg`：`AH=48h` 第一次配到的段會往後移 `e_minalloc` 段。依賴舊落點的既有收據
  （記憶體雜湊、狀態檔）在本分支重跑會不同；畫面不依賴 `AH=48h` 落點的收據不受影響。
- 本變更只在 pools-of-darkness-cht 的 `dosgolem/pod` 分支；其他專案分支不受影響。併回上游前要先盤點各 app 的收據。

## 4. 不做

- `e_maxalloc`（程式要的上限、0xFFFF 表示全部）：真 DOS 會把整塊給 maxalloc 為 0xFFFF 的程式，再由程式用 `AH=4Ah` 縮。
  dosgolem 現行「映像之後即可配置」的模型對 BASIC runtime 等既有案例成立，改它的影響面另立規格。

## 5. 驗收

1. 單元測試（不需原版）：
   - 造一個 `e_minalloc`＝0x40 的 MZ，`LoadEXE` 後 `FreeSeg`＝載入段＋映像段數＋0x40＋1。
   - 父程式（`e_minalloc` 涵蓋堆疊）把 EXEC 參數區與命令列放在堆疊上，EXEC 一個子程式；子程式 PSP 的命令列尾巴與父程式給的相同，
     子程式 PSP ≥ 父程式 `EndSeg`。
2. 整合（缺檔 skip）：pools-of-darkness-cht 的 `CONTROL.EXE` 經 EXEC 叫起 `VACE.EXE` 時，`VACE.EXE` 讀到正確的 ACF 檔名。
3. `tools/go.sh build ./...` 與 `tools/go.sh test ./...` 全綠。
