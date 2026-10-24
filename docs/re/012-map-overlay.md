# 012 地圖疊層的記憶體與實圖證據

日期：2026-10-24。Issue：[#9](https://github.com/wicanr2/pools-of-darkness-cht/issues/9)。
規格：[009](../spec/009-map-overlay.md)。可重生入口：[tools/g8-map-evidence.sh](../../tools/g8-map-evidence.sh)，
檢查器：[tools/map_evidence.py](../../tools/map_evidence.py)。原始輸入、快照及擷取圖片留在 `workplace/`。

## 問題與方法

`apps/pod/automap.go` 尚留前作的 `DS:89CE`、`DS:89D7`。在本作正常冒險狀態，這兩處都是零，不能沿用。
從 `workplace/probe/g6-town/town.state` 重播，解碼實際 `GEO1.DAX` 各塊，再在完整記憶體尋找相同的 0x400 bytes 本體。
只對比資料與正常按鍵結果，不更名原版函式、不修改程式或注入座標。

| 定位 | 原始值與語意 | 等級／出處 |
|---|---|---|
| dosgolem 遊戲資料段 `2096:A9EA`，線性 `2B34Ah` | 遠指標 bytes `00 00 7F 5A`，指向 `5A7F:0000`，線性 `5A7F0h` | 已證實，八個定點取樣與 GEO1 區塊 10h 比對 |
| dosgolem `2096:A9F3`，線性 `2B353h` | x、y、方向各 1 byte | 已證實，轉向與前進的同次執行記憶體及原版座標列 |
| GEO1.DAX 解碼後區塊 10h 的 `[2:402h)` | 0x400 bytes 本體 SHA-256 `4f0753c57be0e7fe9201da8eb255c7051a7888376286d290e39ecc887917bc4a` | 已證實，逐位元組相同 |

另一個可算出相同線性指標的位元組組合出現在 `DS:815D`，但鄰近資料是一串表值，沒有座標／指標結構證據。
它只是不採用的候選，不賦予語意。執行結束時 CPU 的 DS 可能是 `3327h`；以上 DS 指遊戲資料段，
不是任意停止點的 CPU DS。`HDLive` 既有預設為 `2096h`，路徑 A 入口會更新它。

## 按鍵與取樣

起始狀態在第 1,500,000,000 道指令。第 1,502,000,000 步開始，每 30,000,000 步依序送
M、Left、Up、Up、Right、Right、Right。原版方向鍵經 BIOS 緩衝區送入。

| 取樣 | x,y | 方向原值 | 畫面 |
|---|---|---|---|
| start | 11,2 | 4 | 南 |
| left | 11,2 | 2 | 東 |
| east1 | 12,2 | 2 | 東 |
| east2 | 13,2 | 2 | 東 |
| south | 13,2 | 4 | 南 |
| west | 13,2 | 6 | 西 |
| north | 13,2 | 0 | 北 |
| gate | 0,4 | 6 | 西，前次 town1 正常走到叛徒門並觸發手札 57 的狀態 |

所有已取樣的 GEO 本體相同。`gate` 沿用前次 247 鍵的正常探索終點，這次僅執行 1M 步，不送新鍵。
本次不重設或重擲亂數，也不驗證亂數規則；比較從相同已保存的原版機器狀態開始。

輸入檔 SHA-256、兩個狀態 SHA-256、每份記憶體 SHA-256、Go 版本與 dosgolem commit
由 `workplace/probe/g8-map/evidence.json` 保存。Docker 沿用 `dosgolem-go-ebiten:1.24-r1`
（image ID `8ed6bdee2bce`）與 `sosb-img:r1`（`8b9012adb991`）。

| 輸入 | SHA-256 |
|---|---|
| GAME.EXE v1.10 | `35ad635e558852925b7028a921aacafc2df8e84be24a8a1136c6267471948d69` |
| GAME.OVR | `885d6a07a3658f0688dc9d0630dfa984683a3d1d1e515a8e9909d29f62012485` |
| ADLIB.COM | `7c39949b5c740d92449110e896f08f2d11b16b8602d7d3d6d59a9f250d574286` |
| GEO1.DAX | `d628bae8349044a1c1c05b821ccfbba6464ed5bfa57cef15ca05ec52cc5afc1f` |
| g6-town/town.state | `1a7e46dabc16495ffc09b02f7b1b8caa9c0cc184be48132b2272ed9347920959` |
| g6-explore/town1/end.state | `a84e2d8faf3d6271a11764244b148f9e0dbf017560c9801ad5df8b4dc4b09aaf` |

探針基底為 dosgolem `e67971e5cad6001f43081e041e2a92b41409aedc`，Go 1.24.13；
本輪沒有修改 `cmd/probe` 或通用層。

## 地圖與攻略點

實際只有一份 `GEO1.DAX`，共 32 塊，各塊解碼長度都是 0x402。
區塊 10h 的北東／南西牆形狀與 `walkthrough/軟體世界43期/62.jpeg` 的東菲蘭圖一致，出處為雜誌 p.61。
研究圖 `workplace/probe/g8-map/geo-10.png` 與掃描裁切 `east-reference.png` 供目視回查。

| 圖例 | 使用格子 | 實測證據 | 呈現內容 |
|---|---|---|---|
| 3 | (6,3)、(7,3) | `g6-explore/town1/explore.log`：#7218890289、#7249923788 印出 TRAINING HALL | 訓練廳門牌；不宣稱已驗訓練服務 |
| 13 | (13,3)、(12,3) | #1716082897 印 CLOISTER OF TYR；#1808361389 祭司說僅供冥想、不提供法術 | 名稱依術語表「提爾修院」；說明自寫摘要 |

圖例號取自雜誌，但格子取原版輸出，不把房間中央的圖例位置當觸發格。其餘圖例、其他地圖名稱、重用 GEO 的 ECL 區分仍未知。
攻略點摘要只保存自己的說明及出處，不提交雜誌原文或掃描。

## 正常前端揭露的初始化狀態

冷開機讀檔測試的 `g8-map-gui/fixed/save/pod-map-visited.txt` 先記了一筆
`?5f70bf18a086 7 13`，之後才有正確的 `1/10 11 2`。
`5f70bf18a086` 是 0x400 個零的 SHA-256 前綴，而本作 32 塊原版 GEO 本體均非全零。
此筆屬初始化期間的無效緩衝區，不能當作玩家走過的格子。規格回 DRAFT 補上零填緩衝區略過條件，
確認實檔沒有全零 GEO 後再升 READY。這項保護不修改原版記憶體。

## 驗證紀錄與限制

- `tools/g8-map-evidence.sh` 乾淨重生通過：32 塊 GEO、八個取樣，四方向與座標均吻合。證據審查後規格升 READY；adapter 測試與面板收據另記。
- [tools/map_geo.py](../../tools/map_geo.py)：從實際 archive 清單產生 `text/map-geo.tsv`，
  `--check` 重建比對通過。共 32 塊、1 個已命名地圖、4 個攻略格。名稱來源是 `text/map-titles.tsv`，
  攻略點為 `text/map-guide.zh-TW.tsv`，其餘 31 個地圖名稱留空。
- `TestAutoMapPodMemory` 以八份真實記憶體驗證本作位址、地圖身分、方向、走過格與取樣不寫入記憶體；
  `go test ./apps/pod/...` 與 Xvfb 下 `go test ./cmd/pod-play` 通過。缺原版快照時明列 skip。
- `TestMapPanelKeyRouting` 從實際鍵盤對映走到面板動作，驗證 F7 開關、V 顯示／隱藏未走過的攻略點、Esc 關閉、
  F5／F6 互斥，這些鍵不送入原版 BIOS。`TestAutoMapIgnoresInitializingBuffer` 防止初始化資料記成走過格。
- [tools/g8-map-receipt.sh](../../tools/g8-map-receipt.sh)、[tools/map_receipt_check.py](../../tools/map_receipt_check.py)：
  同一正常路徑狀態重播至 #1,520,000,000，2 倍／3 倍的原版、走過、全圖共六張。
  原版色號、完整記憶體、暫存器與 scratch 內容相同，面板外差異 0、缺字 0。
  收據在 `workplace/probe/g8-map-receipt/`；六張記憶體 SHA-256 均為
  `f9aa0bf619d647bce38ed22d1fa090a0ec0d150d03d2de39baf96ebab8a29bb4`，
  色號 SHA-256 均為 `df3c6a4c0c2247dda3d7872e31be58af78e097c4200f418ecb6fc7dcf681dd5d`。
- 目視核對 2 倍與 3 倍全圖，牆面與東菲蘭圖形相符，隊伍在 (11,2) 朝南；圖例 3、13 與右側名稱無裁切。
  全圖 PNG SHA-256 分別為 `6db994df6bb97186297cfb3add451d341dad20dd3443bc0ec0630c26b4a9daab`、
  `9379fb7fd88504f00d3bf665f99fd4368b72041c7a56fbaef7fefb22eaccf710`。
- `tools/font-dev.sh` 已收進兩份玩家可見地圖表；兩個字型各 2,842 字，地圖表缺字 0。
- 地圖表與原版失配、攻略格越界兩個負對照均以非零結束；記錄在 `g8-map-receipt/negative-*.log`。
- 首次編譯前端時 Go 預設並行度觸及容器程序上限，`gcc` 無法 fork。
  使用原 image、`GOMAXPROCS=2`、`go build -p 2` 及 `--pids-limit 256` 乾淨重跑通過，沒有重建 image。
- 前兩次探針誤用了只支援平面顯示模式的 `-dump-screen-png`／`-dump-at`。這是腳本錯誤，與產品無關。
  mode 13h 的中途畫面用 `-shots` 保存 `.raw` 和 `.pal`；終點可用 `-dump-vram` 直接保存色號及 PNG。
- 既有 `g6r-area.state` 的畫面仍是冒險主選單，沒有呈現原版區域地圖；不拿它證明地圖線條。

## pod-play 正常讀檔的路徑修正

`g8-map-gui/boot/inv.tsv`：Load Saved Game 選 Pools 後直接回主選單，槽 A 檔案存在。
sealed owner 依規格 432 沒有設定 ProgramPath，DOS 保留預設虛擬目錄 RICH2；
設定檔的 `C:\POD\SAVE\` 無法對到 save root 下的 SAVE。
探針在 DOS 建構前設定 `C:\POD\ADLIB.COM`，所以既有 probe 收據沒有揭露這個前端問題。

通用層新增 `433-session-boot-program-path`，在所有驗證通過後、載入前同步程式環境路徑與 DOS.Dir／RootDir。
空值保留既有行為；pod-play 明示 `C:\POD\SND.COM`。不修改原版檔案、規則或存檔格式。
無原版的 COM 開檔正負對照、非法路徑無副作用、session／DOS 測試與 go vet 通過。

同一份槽 A、相同畫格鍵序重跑後，`g8-map-gui/fixed/inv.tsv` 確實出現槽位清單、ROLAND、
CHARACTERS AND PARTIES 查驗題（第 1 頁第 1 字，原版答案 YOU），截圖 `fixed/shot.png` 已目視核對。
測試輸入只保留自建的 SAVE，避免原始 1995 年大寫 CHRDATA 檔案遮住測試用小寫檔名。
`bootroot.Prepare` 要求首次安裝的目錄為空，因此先初始化遊戲樹，再放入自建存檔；不繞過這項檢查。

完整正常路徑收據入口：[tools/g8-map-play-receipt.sh](../../tools/g8-map-play-receipt.sh)，
檢查器：[tools/map_play_check.py](../../tools/map_play_check.py)。檢查原版題型、作答後回冒險、
地圖／攻略互斥及走過格，圖片均留 `workplace/probe/g8-map-play/`，不進 Git。

最後定稿重跑通過，六張圖已抽看地圖、攻略與關閉後冒險畫面；地圖先後兩張相同，
關閉後與開圖前的冒險畫面相同，走過檔只有 `1/10 11 2`。
五個自建原版存檔的內容與輸入完全相同，沒有因地圖呈現改寫。

| 產物 | SHA-256 |
|---|---|
| pod-play 驗證程式 | `55e2a926c6e2453360b3b88cd291e2bbd6854d8c56274e3f12ce8c16bdc7ced9` |
| 正常前端地圖 f4320／f4520 | `c639ed168ee479b7af3679cf44558b2c29a305e641fd5a81c2aa6c796e7a9122` |
| 正常前端攻略 f4420 | `70a7731d0842aa82bb08b83a3a360ce28216f6c9995346e344ecb34431ac38a5` |
| 冒險 f4200／關圖 f4620 | `eead4fda4a23c3cfe324eec974297c0e418298529de6e208a68c9d93d9dac851` |

程式、字型、五份地圖／攻略資料表、存檔輸入與截圖雜湊另存 `receipt.json`。
此程式於提交前建置；對應源碼保存於本 repo 的 dosgolem/pod 分支 `939ec6d`（規格 433）與 `ad17b65`（地圖）。
原先 160 秒外層時限的一輪在 f4320 後停止；沒有容器殘留，以相同程式、輸入與鍵序，
僅將外層時限改 300 秒後乾淨重跑通過，歸類為驗證環境時限問題。
