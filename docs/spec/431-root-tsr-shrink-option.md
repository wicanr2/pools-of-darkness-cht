# 431 — 疊底程式常駐後縮回保留量（選用）

狀態：READY（2026-10-23）。需求來源：pools-of-darkness-cht #3（G2）。
前置：[`197-int27-terminate-and-stay-resident`](197-int27-terminate-and-stay-resident.md) §5.1、§8；
[`008-tsr-resident`](008-tsr-resident.md) §2.1；[`009-exec`](009-exec.md) §4（監督佇列）。

## 1. 為什麼要做

`terminate()` 的 TSR 分支是 `freeSeg = max(freeSeg, curPSP + keep)`。頂層 `.COM` 載入時 `FreeSeg = PSPSeg + 1000h`
（名義上擁有整段），所以保留量小於 64 KB 時不縮回，`[PSP + keep, PSP + 1000h)` 成為用不到的間隙。
197 §5.1 把它記成 bump 模型的既有限制，§8 記「沒有證據顯示遊戲依賴」，並以 CONFORMED 收據釘住預設行為。

這次有了遊戲受影響的證據（已證實，dosgolem 收據，pools-of-darkness-cht `workplace/probe/g2-demo-*`）：
《Pools of Darkness》v1.10 的 `START.BAT` 先執行音效驅動 `snd.com`（`STARTUP.EXE` 由 `ADLIB.COM` 複製而來，常駐），再執行 `CONTROL.EXE`。

| 情況 | `CONTROL.EXE` PSP | `GAME.EXE` PSP | 結果 |
|---|---|---|---|
| 不先常駐驅動（`-exe CONTROL.EXE`） | 0100（載入器） | 1635 | 主選單正常；選 DEMO 後執行流跳進 `A000:0000`（`961C:9E40`）停機 |
| 先常駐驅動（`-exe ADLIB.COM -queue CONTROL.EXE`），現行模型 | 1101 | 2636 | DEMO 能跑，但底列出現 `WARNING: INSUFFICIENT MEMORY` |

兩列的 `GAME.EXE` PSP 差 1001h 段，正好是頂層 `.COM` 名義上擁有的那一段。

## 2. 契約

1. `DOS` 新增欄位 `RootTSRShrink bool`，預設 false。false 時所有行為與現行相同（197 的收據不受影響）。
2. `Install()` 記下 `rootFree`＝當時的 `freeSeg`（疊底程式載入後的可配置游標）。
3. 疊底行程（行程疊空）以 TSR 結束（`AH=31h` 或 `int 27h`），而且 `RootTSRShrink` 為真、`d.arena == nil`、`d.freeSeg == rootFree`
   （疊底程式沒有配過、縮過記憶體）時：`freeSeg = curPSP + keep`。其餘情況照現行的 `max`。
4. `cmd/probe` 加旗標 `-root-tsr-shrink`，設定這個欄位。

條件 3 的理由：bump 配置器沒有洞的觀念。疊底程式若在常駐前用 `AH=48h` 拿過塊，那些塊在 `freeSeg` 之下，縮回會讓下一支程式蓋掉它們；
只有「沒配過」時，縮回與真 DOS「常駐只留 `PSP + keep`」（`008-tsr-resident` §2.1 的字面）一致。

## 3. 不做

- 行程疊不空（子行程常駐）的分支不變。
- 不改 `d.arena`、不寫 MCB（同 197 §4）。

## 4. 驗收

1. 單元測試（不需原版）：
   - `RootTSRShrink` 為真：頂層 `.COM`（`freeSeg`＝1100h、`rootFree`＝1100h）以 `int 27h DX=8000h` 常駐，佇列程式 PSP＝0901h。
   - 同樣設定但 `freeSeg`＝1200h（常駐前配過記憶體）：佇列程式 PSP＝1201h（照 `max`）。
   - `RootTSRShrink` 為假：197 的 `TestInt27RootThenQueuedProgramLandsAboveResident` 不改、照樣通過。
2. 整合（缺檔 skip）：pools-of-darkness-cht 以 `-exe ADLIB.COM -queue CONTROL.EXE -root-tsr-shrink` 跑 DEMO，
   不出現 `WARNING: INSUFFICIENT MEMORY`。
3. `tools/go.sh build ./...` 與 `tools/go.sh test ./...` 全綠。
