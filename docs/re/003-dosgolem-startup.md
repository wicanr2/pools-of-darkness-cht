# 003 dosgolem 冷啟動基線

日期：2026-10-23。Issue：[#3](https://github.com/wicanr2/pools-of-darkness-cht/issues/3)。輸入見 [001](001-input-baseline.md)。

## 固定版本

| 項目 | 值 |
|---|---|
| 基底 | 銀色匕首之謎 repo `dosgolem/sosb` 頂端 `a6e0254ff0ff72bd37c29cc1c6d3f5863ec58d49`（2026-10-23，`gh api` 回讀） |
| 本作分支 | `workplace/dosgolem/` 的 `pools-of-darkness-cht`，推到本 repo 的 `dosgolem/pod`；目前頂端 `2e47f92` |
| Go | go1.24.13 linux/amd64 |
| image | `dosgolem-go-ebiten:1.24-r1`（`sha256:8ed6bdee2bce1f32f804e888f942b25bbf62f2165a43344bb4587f3371de0581`） |
| 執行入口 | dosgolem `tools/go.sh`（`--rm`、`--network none`）；本 repo 包裝 `tools/probe.sh` |

選 `dosgolem/sosb` 的理由：步驟比照銀色匕首之謎，該分支是 Gold Box 系列最新的通用層，含克萊恩英豪 `dosgolem/cok` 的 HD 主題、切換列、
字型倍率、多語系、配樂與發行模式，另有規格 021 的閃現量尺、殘影疊字、固定圖面、保留呈現與 HD 鎖存，以及 TP5 版 CRT `Delay`（dosgolem 規格 429）。
其他候選（克萊恩英豪 `dosgolem/cok` `29f7607`、dosgolem `main` `81c85be`、黑暗之后所在的 `san5-mixer` `f79eab6`）沒有規格 021 的機制。
mode 13h 與 DOS EXEC 在本分支已有實作，下節的冷啟動證實可用。
module cache 由銀色匕首之謎 clone 的 `workplace/gomodcache` 複製，未開網路。歷史含前作的 `codex@openai.com` commit，未改寫。

## 通用層修正（dosgolem `fee91a6`）

| 規格 | 問題 | 證據 | 修正 |
|---|---|---|---|
| `430-mz-minalloc-reservation` | 載入器的 `EndSeg` 只算到映像尾，沒算 MZ `e_minalloc`。`CONTROL.EXE`（映像 0xEB 段、`e_minalloc` 0xB9、堆疊頂＝映像＋0xB9）以 `AH=4Bh` 叫 `VACE.EXE` 時，參數區在 `0234:04E2`（自己的堆疊），子程式落在 PSP 01FD 把它蓋掉；`VACE.EXE` 報 `File not found "<亂碼>.ACF"`，exit 1 | 已證實（`-log-calls`、子程式 PSP+80h 傾印、負對照單元測試） | `EndSeg`、`FreeSeg` 與 EXEC 需求量含 `e_minalloc` |
| `431-root-tsr-shrink-option` | 頂層 `.COM` 常駐後，bump 模型以 `max` 留下整段 64 KB 間隙（規格 197 §5.1 的既有限制）。常駐音效驅動後 `GAME.EXE` 的 PSP 從 1635 移到 2636，DEMO 底列出現 `WARNING: INSUFFICIENT MEMORY` 並停在同一句 | 已證實（兩種入口的 `EXEC 紀錄` 與截圖） | 新增預設關閉的 `-root-tsr-shrink`：疊底程式沒配過記憶體就常駐時縮回保留量。197 的預設行為與收據不變 |

兩項都在 `docs/spec/000-index.md` 登錄並改為 CONFORMED（`2e47f92`），`tools/go.sh test ./...` 全綠。只在 `dosgolem/pod` 分支，未推送 dosgolem 上游。
431 只接在 `cmd/probe`：`oracle.LoadWith` 只載 MZ，能載 `.COM` 的 `oracle.LoadProgram` 不吃 `Options`。G3 的 `pod-play` 要以常駐驅動為疊底程式，
得先補 oracle 入口（431 §5，另立規格）。

## 啟動流程（已證實，dosgolem 收據）

`START.BAT` 的四行在 dosgolem 依序重現：

| 步驟 | dosgolem 作法 | 觀察 |
|---|---|---|
| `startup` | `-exe STARTUP.EXE`，送 `232` 與兩次 Enter | 文字模式 03h。依序問 NUMLOCK 提示、`Sound Type`（[1] Sound Blaster、[2] Adlib、[3] Tandy 1000 EX/TX/SX/HX、[4] Tandy 1000 SL/TL、[5] PC、[6] Silent）、`GRAPHICS ADAPTER TYPE`（[1] EGA、[2] Tandy、[3] VGA）、`ALTERNATE INPUT TYPE`（[1] JOYSTICK、[2] MOUSE、[3] NEITHER）、`PATH TO SAVE DATA (DEFAULT - C:\POD\SAVE)`、`PATH TO SAVE DATA FOR SECRET OF THE SILVER BLADES (DEFAULT - C:\SECRET\SAVE)`。寫出 `pool4.cfg` 並把 `ADLIB.COM` 複製成 `snd.com`，#38,822,676 以 `AH=4Ch` 結束 |
| `if exist snd.com snd` | 疊底程式 `-exe ADLIB.COM`（與 `snd.com` 位元組相同），`-root-tsr-shrink` | 常駐 |
| `control %1 %2 %3 %4` | `-queue CONTROL.EXE` | `CONTROL.EXE` 讀 `pool4.cfg`，EXEC `vace.exe`（片頭，`INTRO.ACF`、`INTRO.BMA`、`P1.SMS`），之後 EXEC `game.exe` |
| `if exist snd.com snd U` | 未重現 | 遊戲結束後才執行 |

`pool4.cfg`（85 bytes，CRLF 分行）：`2`（音效 Adlib）、`V`（VGA）、`M`（滑鼠）、`C:\POD\SAVE\`、`C:\SECRET\SAVE\`、`F`、其後 13 行 `0`。
`-program-path` 用安裝預設目錄 `C:\POD`（`pod-refcard.pdf`「CD\POD」），預設存檔路徑才是 `C:\POD\SAVE\`。

## 冷啟動觀察

#1,014 與 #160M 兩列取自收據的 `cold.log`；片頭與製作群的步數取自不常駐驅動的探針（`workplace/probe/g2-control-3`、`-5`），
兩者的程式鏈不同，步數只作定位。

| 步數 | 事件 |
|---|---|
| #1,014 | `VACE.EXE` 切到 mode 13h |
| #10,000,000 | SSI／AD&D 標誌（256 色） |
| #60,000,000–#499,000,000 | 片頭動畫：水晶洞窟、群魔、菲蘭城、骷髏，畫面下方有英文字幕（第一句以「I have summoned you here」開頭） |
| #160,000,000 | 送空白鍵，`VACE.EXE` 結束（exit 0） |
| #160,051,918、#160,338,178 | `GAME.EXE` 切到 03h，再切到 13h |
| #200,000,000–#300,000,000 | 製作群畫面；#300M 底列出現 `POOLS OF DARKNESS 1.10 PLAY DEMO` |
| 送 `p` 後 #399,000,000 | 建立隊伍主選單：CREATE NEW CHARACTER、ADD CHARACTER TO PARTY、LOAD SAVED GAME、INITIALIZE MOUSE/JOYSTICK、EXIT TO DOS，底列 `SELECT` |

- 顯示：`VACE.EXE` 與 `GAME.EXE` 都是 mode 13h（320×200、256 色），寫過 DAC 埠 3C7–3C9。克萊恩英豪、銀色匕首之謎是 0Dh。
- 沒實作的服務：0 種。寫過的 I/O 埠：040、043、061、388、389、3C7、3C8、3C9。
- 片頭跳過後若送空白鍵選 DEMO：不常駐音效驅動時，執行流跳進 `961C:9E40`（線性 A0000）停機；常駐後 DEMO 正常演到戰鬥（#1,200,000,000 內）。
  本作要照 `START.BAT` 先常駐驅動。

## `SAVE/` 能否讀取（已證實）

主選單選 LOAD SAVED GAME 後出現 `LOAD FROM WHERE? POOLS SECRET EXIT`，選 POOLS，遊戲依序開 `C:\POD\SAVE\savgamA.pty` 到 `savgamJ.pty`，全部找不到，回主選單。
ADD CHARACTER TO PARTY 找 `C:\POD\SAVE\*.pc`，同樣找不到。zip 的 `SAVE/` 沒有這兩種檔，無法從遊戲內讀取；測試用存檔要自己建立。

## 收據

[`tools/g2-startup-receipt.sh`](../../tools/g2-startup-receipt.sh)：首次執行設定 → 常駐驅動＋`CONTROL.EXE` 到 #300,000,000 存狀態 → 從狀態送 `p` 重播兩次到 #399,000,000。

| 產物 | SHA-256 |
|---|---|
| `scratch/pool4.cfg` | `6161a336828e889c6377bd4fa327376f691a57fc3cd4f2eb6f74db06908d03bd` |
| `scratch/snd.com`（＝`ADLIB.COM`） | `7c39949b5c740d92449110e896f08f2d11b16b8602d7d3d6d59a9f250d574286` |
| `credits-300m.state` | `c24d1d0a5d733455813a47823c9888b897828ae6e4c97ef33dde64624b363e35` |
| 主選單色號陣列 `menu-*.raw` | `fefa05142448782c528128c506ffb6ce7fb7c3f39c72dd2e420016e0c1fc11f4`（重播 a、b 相同） |
| 主選單色盤 `menu-a.pal` | `62ed0c9528b22f85a7ca8a4d313bc3eb6a80ddfde959b927d71980ab2f68a486` |

## 與前作的對照

| 項目 | 本作 | 銀色匕首之謎 | 克萊恩英豪 |
|---|---|---|---|
| 視訊模式 | 13h | 0Dh | 0Dh |
| overlay 魔數 | `FBOV` | `TPOV` | `FBOV` |
| 程式鏈 | `STARTUP.EXE` → 常駐 `snd.com` → `CONTROL.EXE` → `VACE.EXE`／`GAME.EXE` | `START.EXE` 單一程式 | `START.EXE` 單一程式 |
| 設定檔 | 首次執行建立 `pool4.cfg` | 首次執行建立 `blades.cfg` | 原版附 `KRYNN.CFG`，需改 AdLib |
| 進主選單 | 空白鍵跳過片頭，製作群後送 `p` | 空白鍵跳過標題，再送 `p` | 送 `p` |

## 未知

- `pool4.cfg` 第 6 行 `F` 與後續 `0` 欄位的語意。
- 片頭片尾字幕由 `VACE.EXE` 畫出，與 `GAME.EXE` 的印字路徑不同；輸出攔截點待 G3。
- 銀色匕首之謎角色匯入（LOAD FROM WHERE? 的 SECRET）沒有測。
