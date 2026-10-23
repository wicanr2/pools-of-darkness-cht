# 432 — Sealed Owner 開機：`.COM` 疊底、監督佇列、疊底常駐縮回

狀態：CONFORMED（2026-10-23；§4 第 1、3 項單元測試、第 2 項收據見 pools-of-darkness-cht `tools/play-smoke.sh`：`pod-play` 冷開機到主選單，原版色號雜湊 `fefa0514…` 與 G2 收據相同。READY 前 DRAFT 自審：`.COM` 上限改與 `machine.loadCOMAt` 一致〔原寫 0xFF00〕、確認 `dos.Install` 不動監督佇列）。需求來源：pools-of-darkness-cht #7（G6 `pod-play`）。
前置：[`235-sealed-owner-private-boot-draft`](235-sealed-owner-private-boot-draft.md)（READY）、
[`009-exec`](009-exec.md) §4（監督佇列）、[`431-root-tsr-shrink-option`](431-root-tsr-shrink-option.md)（§5 開放項目）。

## 1. 為什麼要做

`session.Owner.BootOriginal`（235）只做 `LoadEXE → Install → Running`：映像必須是 MZ，沒有監督佇列，
也沒有 431 的 `RootTSRShrink`。這三樣 `cmd/probe` 都有（`-exe` 依檔頭分派、`-queue`、`-root-tsr-shrink`），
所以探針能照批次檔重現啟動鏈，play 指令不能。

《Pools of Darkness》v1.10 的 `START.BAT` 是 `startup` → `snd` → `control` → `snd U`
（pools-of-darkness-cht `docs/re/001`）。`snd.com`（`STARTUP.EXE` 由 `ADLIB.COM` 複製）要先常駐，否則選 DEMO 後
執行流跳進 `A000:0000` 停機；常駐時沒有 431 又會報記憶體不足（431 §1 的表）。設定檔已存在時 `STARTUP.EXE`
跑 7,520,306 道指令就以 `AH=4Ch`、離開碼 0 結束，不開任何檔、不等按鍵（pools-of-darkness-cht
`workplace/probe/g6-startup-existing/run.log`，已證實），所以 play 指令可以在安裝存檔目錄時跑一次設定，
開機時從 `snd.com` 起。431 §5 記下的缺口就是這個入口。

## 2. 契約

`BootInput` 新增三個欄位，**零值全部等於 235 的現行行為**：

```go
type BootFormat uint8

const (
    BootFormatEXE BootFormat = iota // 235 的 MZ 載入（預設）
    BootFormatCOM                   // machine.LoadCOM
)

// BootQueued 是開機後排入監督佇列的一支程式（009 §4）。
type BootQueued struct{ Name, Args string }

type BootInput struct {
    EXE               []byte
    ExpectedEXESHA256 [32]byte
    SaveRoot          string
    Format            BootFormat   // 規格 432
    Queue             []BootQueued // 規格 432
    RootTSRShrink     bool         // 規格 432（431 §2）
}
```

1. 驗證拒絕（同 235 第 1–3 條：不呼叫 `fail()`、phase 不變、不裝 DOS、不計步）再加：
   - `Format` 不是 `BootFormatEXE`、`BootFormatCOM` 之一。
   - `Format == BootFormatCOM` 而 `len(EXE) > 0xFEFE`（與 `machine.loadCOMAt` 的上限相同：映像從 0100h 起、堆疊從 FFFEh 往下，見 `internal/machine/loader.go`）。驗證先擋，載入器不會在這裡失敗。
   - `len(Queue) > 8`。
   - 任一 `Queue[i].Name` 為空、長於 12 bytes、含 `/`、`\`、`:` 或控制字元（佇列名照 basename 在 `Root` 下解析，009 §4）。
   - 任一 `Queue[i].Args` 長於 126 bytes 或含 CR、NUL（命令列尾的上限與結尾，同 `oracle.SetCmdLine`）。
2. 開機步驟（235 第 4 條的「步驟 4」）改為：設 `d.Root`、`d.Scratch` → 依 `Format` 呼叫 `LoadEXE` 或 `LoadCOM`
   → `d.RootTSRShrink = input.RootTSRShrink` → 依序 `d.Enqueue(Name, Args)` → `Install` → Running。
   載入失敗的處理同 235（Failed、firstFault、無半套映像）；`RootTSRShrink` 與佇列只在載入成功後才設。
3. 雜湊、別名、save root 檢查、零步開機與單一前端 goroutine 規則全部沿用 235，`ExpectedEXESHA256` 比對的是 `EXE`
   本身（`.COM` 也一樣）。
4. 佇列與縮回的執行期語意是 009 §4、431 §2 的既有行為，本規格不改 `internal/dos`。

## 3. 不做

- 不改 235 既有欄位與錯誤訊息；不自動依檔頭判斷格式（呼叫端明示）。
- 不設 `ProgramPath`、命令列尾（疊底 `.COM` 不帶參數；需要時另立）。
- 不處理 `START.BAT` 最後的 `snd U`（卸載驅動）；play 指令在遊戲結束時整台結束。

## 4. 驗收

1. 單元測試（不需原版）：
   - 零值欄位的開機與 235 現行測試結果相同（現有 `boot_original_test.go` 全過）。
   - 第 1 條每一種拒絕：phase 仍是 Booting、`Machine.Steps` 不變、之後還能正常開機。
   - `BootFormatCOM`＋`Queue`＋`RootTSRShrink`：疊底 `.COM` 以 `int 27h` 常駐 `DX` 段數，佇列中的 `.COM`
     結束後 `ExecLog` 記錄的 PSP 等於疊底 PSP＋保留量（縮回）；同一組輸入 `RootTSRShrink=false` 時是疊底 PSP＋1000h 之後（不縮回）。
2. 收據（需要原版，pools-of-darkness-cht）：`pod-play` 以 `snd.com` 疊底、佇列 `CONTROL.EXE`、`RootTSRShrink` 開機，
   與 `tools/g2-startup-receipt.sh` 同樣的送鍵走到主選單，原版畫面色號雜湊與 G2 收據的 `menu-a.raw` 相同。
3. `go vet`、`gofmt`、`go test ./session/... ./internal/dos/...` 全過。
