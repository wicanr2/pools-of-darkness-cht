# 019 CPU 模擬速度

狀態：DRAFT；CRT 辨識增補為 READY。Issue：[#14](https://github.com/wicanr2/pools-of-darkness-cht/issues/14)。

## 玩家契約

工具列選擇 1x、2x、4x、8x、16x、32x，保存上次設定。不增加快速鍵。
沿 dosgolem 規格429的虛擬時脈與 Delay 加速契約；沒有已確認的 Delay 時不啟用加速。
正式速度與字級、語系等呈現選項不進 DOS 存檔。

## 本作 CRT 辨識增補

已載入的 `runtime/tpcrt.FindDelay` 只認得兩種外層形式。本作是第三種：
ES 先設為0040h，次數先讀入 BX，每次以 CX=BX 呼叫同一個三指令內圈。
這項差異只描述原始 bytes，不據此推定精確編譯器版本。

| 欄位 | 證據 |
|---|---|
| 輸入 | `GAME.EXE` v1.10，SHA-256 `35ad635e558852925b7028a921aacafc2df8e84be24a8a1136c6267471948d69` 的正常解壓記憶體 |
| 工具與位址 | IDA Pro 9.4、16-bit；dosgolem 執行期 `1DF0:029C–02C6`，線性 `0x1E19C–0x1E1C6` |
| 辨識範圍 | 完整43 bytes；次數欄位之外逐 byte 符合，常規記憶體中與所有既有形式合計唯一 |
| 次數欄位 | 本體 `+0x16` 的 little-endian word，實際為 `FB82`；`DS=2096`，線性 `0x304E2` |
| 內圈 | 本體 `+0x23`，線性 `0x1E1BF`；`26 3A 05 75 02 E2 F9 C3` |
| 執行語意 | 同既有 TP5 形式的三指令內圈；只增辨識表，不修改 machine 的加速演算法 |
| 失配 | 無命中、任一固定 byte 不符、不同形式多重命中，均回 false |

外層 bytes：

```text
8B DC 36 8B 57 04 0B D2 74 16 B8 40 00 8E C0 33
FF 26 8A 05 8B 1E ?? ?? 8B CB E8 06 00 4A 75 F8
CA 02 00 26 3A 05 75 02 E2 F9 C3
```

原始匯出在 `workplace/probe/demo-control-v1/keywait-1df0-ida.json`，
記錄輸入切片雜湊、IDA 位址、原始 bytes 與 operand。這裡只需確認本作的呼叫形式與欄位，
不重新研究 PIT、校準迴圈或逐週期硬體時序。

## 驗收

- 辨識測試使用本作完整 byte fixture，驗證內圈與次數位址，另驗失配及跨形式歧義拒絕。
- 既有 machine 的兩種內圈精確加速測試及 session、runtime 測試通過。
- 正常冷開機以相同輸入、同一倍率比較加速開／關的 CPU、完整記憶體與畫面。
- 各倍率工具列、設定保存與讀回另做正常前端抽樣。這些通過前，整份規格不升 CONFORMED。
