# 433 Sealed Owner 開機的 DOS 程式路徑

狀態：CONFORMED（2026-10-24；§3 的單元測試、go vet 與正常冷開機讀檔收據通過）。需求來源：pools-of-darkness-cht #9 的正常讀檔驗證。
前置：[432-session-boot-com-queue](432-session-boot-com-queue.md)、
[198-dos-relative-directories](198-dos-relative-directories.md)。

## 原因與證據

`cmd/probe` 在建立 DOS 前設定 `machine.ProgramPath`；`dos.New` 依該路徑設定 `Dir`、`RootDir`。
`session.Owner` 先建立 DOS，`BootOriginal` 又沒有路徑輸入，因此保留預設 `RICH2`。
`C:\POD\SAVE\savgamA.pty` 不會去掉 POD 前綴，無法對到 save root 內的 `SAVE/savgamA.pty`。
探針正常讀檔，pod-play 從 Load Saved Game 選 Pools 後回主選單，原版資料相同。
原版輸入與實測收據見 pools-of-darkness-cht `docs/re/012-map-overlay.md`。

這是 198 的已知路徑契約，無須反組譯遊戲。不得改動通用 DOS 的預設目錄或 basename fallback。

## 契約

- `BootInput` 新增 `ProgramPath string`。空字串完全保留 235／432 的既有行為。
- 非空時為絕對 DOS 程式路徑，格式 `磁碟機字母:\目錄\檔名`，也可在根目錄。
  最多 127 bytes，僅可列印 ASCII；路徑元件不得為空、`.`、`..`，不得含 `/ : * ? " < > |`。
  這是本 API 的輸入限制，不宣稱 DOS 系統的最大路徑長度。
- 所有驗證在狀態變更前完成。非法值回錯、phase 保持 Booting、零計步，之後可正常開機。
- 通過所有驗證後、載入映像前，同步設定 machine.ProgramPath、DOS.Dir、DOS.RootDir。
  目錄依既有 programDir 正規化為大寫且去掉磁碟機、首尾反斜線；明示根目錄時為空。
- `internal/dos.SetProgramPath` 重用 programDir；它不執行 chdir、不碰主機路徑、不建立目錄。
  session 只在 Booting 設定；Running 仍不暴露可變路由。
- pod-play 傳 `C:\POD\SND.COM`，與設定檔、原版安裝目錄及探針一致。
  不修改磁碟上的原版 EXE 或既有設定檔。

## 驗證

1. 不需原版的 COM：用 AH=3Dh 開 `C:\POD\SAVE\PROOF.DAT`。
   ProgramPath 空值時無法開啟；明示 `C:\POD\SND.COM` 時成功開啟 save root 內的檔案。
2. 根目錄路徑、環境區塊內程式路徑、非法輸入無副作用、既有 session／DOS 測試。
3. pod-play 冷開機，以同一份槽 A、相同按鍵時序，修正後能開槽位清單、載入並進入原版查驗題。
   完成作答後再驗 F7 地圖、F6 攻略互斥。原版亂數不修改、不重擲。

不改存檔格式、遊戲規則、顯示資料或預設行為。只在本作隔離 clone 保存，不推 dosgolem 上游。

## 驗證結果

- COM 絕對路徑開檔正負對照、11 種非法路徑無副作用、根目錄路徑通過。
- `go test ./session/... ./internal/dos/...`、對應 go vet 及 Xvfb 下 `go test ./cmd/pod-play` 通過。
- 本作 `tools/g8-map-play-receipt.sh`：同一份槽 A 與既定鍵序，冷開機讀檔後出原版查驗題，答 YOU 回冒險，
  再開關地圖及攻略；無機器狀態注入。收據與輸入 SHA-256 由本作研究 012、`workplace/probe/g8-map-play/receipt.json` 回查。
