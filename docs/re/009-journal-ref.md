# 009 手札引用

日期：2026-10-23。Issue：[#8](https://github.com/wicanr2/pools-of-darkness-cht/issues/8)。前置：[004](004-output-paths.md)、[005](005-journal-errata.md)。

## 引用句（靜態反組譯）

本作的 ECL 文字沒有前作那種「… JOURNAL ENTRY 51.」句子：`workplace/text-src/ecl-source.tsv` 含 journal 的只有 3 則
（`JOURNAL`、`JOURNAL NUMBER?`、紅皮書一句），都不帶編號。引用句由引擎組成。

`GAME.OVR` 檔案位置 0x261C 是 Pascal 字串 `This is recorded as journal entry `（catalog `ui.5349687899c0`），0x263F 是 `.`；
緊接的程序從 0x2643 起（反組譯用 `tools/re/dis16.py` 的同一套 capstone，起點換成檔案位置）：

| 檔案位置 | 行為 |
|---|---|
| 0x264A–0x265A | 取 ECL 指令的 1 個參數（遠呼叫 overlay stub `0055:002A`、`0055:0025`），即手札編號 N |
| 0x265E–0x2663 | 路徑 B 游標 `DS:9D93`＝0x11、`DS:9D92`＝1（第 17 列第 1 欄；游標位址見規格 002） |
| 0x2668–0x26A1 | 字串＝常數 + `IntToStr(N)` + `.` |
| 0x26A6–0x26C1 | 路徑 B（遠呼叫偏移 `0498` 與路徑 B 進入點相同，段由載入時重定位，強推論是 `1B2E:0498`）：參數依序推入 1、0x11、0x26、0x16、0x0A、0、1 與字串遠指標，讀作旗標 1、底色 0、前景 10、下 22、右 38、上 17、左 1 |

依程式，路徑 B 應一次收到整句「This is recorded as journal entry N.」（強推論，實機未量；換行與游標起點也未量）。adapter 的 `JournalRefOf` 以 `JOURNAL ENTRY ([0-9]{1,3})\.`（不分大小寫）認出編號，
開第 N 則手札面板；譯文走組句模板 `This is recorded as journal entry {n}.` →「此事記為旅行手札第 {n} 則。」（`text/composed-templates.tsv`，前作沿用）。

## 驗證

- 單元測試：`apps/pod/journal_test.go` 的 `TestJournalRefOf`（本作整句）、`TestJournalSentenceTemplate`（組句譯文）；
  `TestJournalTableFits` 以 `text/journal.zh-TW.tsv` 跑，99 則都在 3 頁以內、標題不超過 38 格。
- 實機尚未觀測引用句。碼頭港務長的導覽（答 Y）走完 6 段對白沒有記手札（`workplace/probe/g7-journal/tour.log`）。
  導覽結束停在 (3,6) 面西、底列只有 `EXIT`；此時方向鍵可轉向，往東前進會一直觸發「A SIGN OVER THIS DOOR READS 'VALHEGEN PARK'」。
  先送 Enter 再送方向鍵，座標與方向都沒變，原因未查（見「方向鍵對照」）。
- 觸發點要靠 G8 的自動探索（隊伍位置位址）或找出 ECL 指令編號後反查事件；收據到時補。

## 方向鍵對照（待查，G8 開工的起點）

同一狀態 `workplace/probe/states/g7j-tour.state`（導覽結束，第 5,000,000,000 道指令；可寫層 `workplace/probe/g7-journal/scratch` 複製一份再用），
共同參數 `PROBE_EXE=ADLIB.COM tools/probe.sh <輸出> -scratch-dir /out/<複本> -root-tsr-shrink -load-state /st/g7j-tour.state`，鍵間隔 `-bios-key-every 100000000`（look 為 150000000）：

| 鍵序（`-bios-key-names`） | `-bios-key-from` | 結果 | log |
|---|---|---|---|
| `Right` | 5101000000 | (3,6) W → N | `t.log` |
| `Right,Right,Up×8` | 5101000000 | 轉成 E；前進被公園矮牆擋住，每次 Up 觸發 VALHEGEN PARK 告示 | `walk2.log` |
| `Return,Up×9` | 5001000000 | 一直是 (3,6) W | `walk1.log` |
| `Return,Left,Left` | 5001000000 | 一直是 (3,6) W | `look.log` |

四組的鍵都被 `int 16h AH=00` 取走（log 的「按鍵去向」）。遊戲沒掛 int 09h，`-press`（IRQ1）送不進去（`press.log`）。
