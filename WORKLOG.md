# 工作歷程

## 2026-10-23 G0 啟動

- 目標：依使用者指示，參考銀色匕首之謎建立治理文件、private 儲存庫與初始 Issues；步驟完全比照銀色匕首之謎。
- 讀取：銀色匕首之謎 `AGENTS.md`、`CLAUDE.md`、`CONTEXT.md`、`README.md`、`.gitignore`、`WORKLOG.md`、`LICENSE`（含 G0 版 `dcd0115`）、
  `docs/spec/000-index.md`、`010-hd-theme.md`、`021-no-original-flash.md`、`docs/re/README.md` 與 Issue #1–#17；
  黑暗之后 `CONTEXT.md` 的 mode 13h 段落；幽靈騎士 `CONTEXT.md` 的 mode 與 HD 段落。
- 原始輸入：在 `sosb-tools:bookworm-r1`（`b9b6c7540eff`，沿用銀色匕首之謎；`--network none`、唯讀掛載）內計算兩個壓縮檔 SHA-256，
  以 `unzip -l` 列 zip 成員、`lsar -l -e cp950` 列 rar 成員，計算 `walkthrough/` 31 張掃描的 SHA-256（G1 記入 `docs/re/001`）。未解壓。
  成員清單暫存在本 session 的 scratchpad，不入 Git。
- 觀察：zip 是已展開的遊戲目錄，不是安裝器；附 `EGA.DRV`、`VGA.DRV`、`TGA.DRV`，前作都在 mode 0Dh 驗證，本作可能是 mode 13h。
  珍157 是 107 張 JPG，沒有 PDF。
- 使用者決定：repo `pools-of-darkness-cht`、短名 `pod`；G0 全做；授權上網取得英文手冊與手札；顯示模式 VGA 優先。
- dosgolem 候選以 `git for-each-ref` 與 `gh api` 回讀：銀色匕首之謎 `dosgolem/sosb` `5e9a961`；克萊恩英豪 `origin/dosgolem/cok` `29f7607`；
  黑暗之后 clone 的 `san5-mixer` `f79eab6`；dosgolem `main` `81c85be`。
- 變更：新增 `AGENTS.md`、`CLAUDE.md`、`.gitignore`、`LICENSE`（由銀色匕首之謎 G0 版 RRSAL-1.0 改填本作名稱與網址）、
  `CONTEXT.md`、`WORKLOG.md`、`README.md`。
- GitHub：`gh auth status` 為 `wicanr2`；同名 repo 不存在；建立 private repo 並推送，回讀 PRIVATE、Issues 啟用；
  建立 #1–#12（G0–G11）與 #13–#16（規格 018–021，對應銀色匕首之謎後來新增的同號需求），連結回填 AGENTS.md 第 2 節。
  銀色匕首之謎 #17（無敵位址驗證）併入 #11 的完成條件。
- 清理：容器皆 `--rm`，未寫入本目錄。

## 2026-10-23 G1 原始輸入盤點（#2）

- 沿用銀色匕首之謎的 `sosb-tools:bookworm-r1`（`b9b6c7540eff`）與 `sosb-img:r1`（`8b9012adb991`，Pillow 11.0.0），Dockerfile 複製到 `tools/docker/` 並註明。
- 新增 `tools/inventory.py`：zip 是 deflate 的已安裝目錄，`zipfile` 直接展開到 `workplace/input/dos/pod/`；rar 以 `unar -e cp950` 展開到
  `workplace/input/manual-zh/`；攻略掃描只計雜湊。`tools/inventory_tsv.py` 產生 `docs/re/001-input-files.tsv`（265 列）。
- 新增 `tools/exe_probe.py`：版本 1.10（`GAME.EXE` 字串）；`START.BAT` → `STARTUP.EXE` → `CONTROL.EXE` → `VACE.EXE`／`GAME.EXE`；設定檔 `pool4.cfg`；
  `GAME.OVR` 魔數 `FBOV`（與克萊恩英豪、幽靈騎士相同），`AGENTS.md` 第 4 節假說 1 依此改寫。查驗題問手札頁碼與標題後第 N 字。見 `docs/re/001`。
- `SAVE/` 讀 `.SAV` 名字欄：A–E 五組都是職業名的 6 人，B、C、D 另有 NACACIA 或 SHAL；判定為 1995 年玩家資料，沒有 `SAVGAMx.DAT`。
- 新增 `tools/contact_sheet.py`：珍157 與攻略掃描做縮圖拼版（`workplace/manual-zh/sheets/`、`workplace/walkthrough/sheets/`）逐張讀頁碼。
  珍157 是操作手冊 63 頁、冒險者手冊 110 頁，無缺頁；旅行手札第 1–100 則在冒險者手冊 p.55–106；操作手冊附錄二是查驗題答案表（不轉錄進 Git）。
  攻略是《軟體世界》第 43–45 期「完全攻略」三回、第 1–24 章。見 `docs/re/002`。
- 依使用者授權從 mocagh.org 下載 `pod-journal.pdf`、`pod-manual.pdf`、`pod-refcard.pdf`、`pod-hintbook.pdf`（容器只在下載時開網路），
  `pdftotext -layout` 轉文字；英文手札 100 則。
- `LICENSE` 第 2 條 (c) 補上攻略掃描雜湊與引用的少量原版英文字串。
- 勘誤：本輪曾用主機 `python3` 替換 `AGENTS.md` 字串；之後主機上的文字修改一律用編輯工具。
- 清理：容器皆 `--rm`；`find . -user root` 0 筆。

## 2026-10-23 G2 dosgolem 基線（#3）

- 重新回讀候選：銀色匕首之謎 `dosgolem/sosb` 已從 `5e9a961` 前進到 `a6e0254`（已推送、工作樹乾淨），選它為基底。
  由其本機 clone `--no-hardlinks` 複製到 `workplace/dosgolem/`，開分支 `pools-of-darkness-cht`，repo-local 身分 `wicanr2@gmail.com`；
  遠端 `sosb-local` 指回前作 clone（只讀），`origin` 指本 repo。module cache 由前作 clone 複製，未開網路。
- 新增 `tools/probe.sh`（`-program-path C:\POD\<程式>`，安裝預設目錄依 `pod-refcard.pdf`）、`tools/b8000_text.py`、`tools/indexed_png.py`。
- `STARTUP.EXE`：文字模式問音效、顯示、輸入與兩個存檔路徑；送 `232` 與兩次 Enter，寫出 `pool4.cfg`，並把 `ADLIB.COM` 複製成 `snd.com`。
- `CONTROL.EXE` 第一次跑：`VACE.EXE` 收到亂碼 ACF 檔名。追到載入器沒算 MZ `e_minalloc`，子程式落在父程式堆疊上蓋掉 EXEC 命令列。
  寫 dosgolem 規格 430（READY）後改 `loadEXEAt`／`imageParags`，加兩條單元測試；修正前的負對照重現同一症狀（命令列讀到 `0E 1F BA 0E`）。
- 片頭能播（mode 13h，256 色，有英文字幕），空白鍵跳過後進 `GAME.EXE`。選 DEMO 時執行流跳進 A0000 停機；照 `START.BAT` 先常駐驅動
  （`-exe ADLIB.COM -queue CONTROL.EXE`）後 DEMO 能跑，但頂層 `.COM` 常駐留下 64 KB 間隙，遊戲報 `INSUFFICIENT MEMORY`。
  該間隙由 dosgolem 規格 197 §5.1 以收據釘住，改用預設關閉的選項：寫規格 431（READY），加 `RootTSRShrink`／`-root-tsr-shrink` 與兩條單元測試。
  開啟後 DEMO 演到戰鬥，沒有記憶體不足。
- `tools/go.sh test ./...` 全綠（中途一次失敗是規格索引沒登 430，補上）。dosgolem `fee91a6` 推到本 repo 的 `dosgolem/pod`。
- 新增 `tools/g2-startup-receipt.sh`：設定 → 常駐驅動＋CONTROL 到 #300M 存狀態 → 送 `p` 重播兩次到 #399M，主選單色號雜湊一致。見 `docs/re/003`。
- `SAVE/`：LOAD SAVED GAME 找 `savgamA–J.pty`、ADD CHARACTER 找 `*.pc`，zip 裡都沒有，遊戲內讀不到。
- 清理：容器皆 `--rm`；`find . -user root` 0 筆。

## 2026-10-23 G3 輸出路徑盤點（#4）

- 以幽靈騎士、銀色匕首之謎的簽章在 `GAME.EXE` 找到逐字、路徑 A、路徑 B（`tools/re/sigscan.py`、`fn_diff.py`），與幽靈騎士逐 byte 對齊。
  以檔案位移反推段值失敗：`GAME.EXE` 是 EXEPACK（進入點在映像尾端、`RB` 簽名、重定位 0）。改在主選單狀態以 `-find` 找線性位址，
  `-dump-mem` 傾印後用 `tools/re/memcalls.py` 找遠呼叫，得到 resident 段 `1B2E`。
- DEMO 中以 `-call-args` 攔下路徑 A（485 次，混合大小寫）、路徑 B（13 次，整句 ECL）、逐字程序、路徑 D。第一次把兩組 `-call-args`
  用逗號合併，只有第一組生效，路徑 B 被誤以為沒呼叫；改成分次跑。
- `tools/g3-game-image.sh` 取解開後的映像；`tools/ui_text_catalog.py`（改讀映像＋GAME.OVR）1,357 則，`tools/ecl_text_catalog.py` 4,725 則；
  DEMO 錨點：路徑 A 靜態字串 8/8、路徑 B 12/12 命中。
- 片頭字幕在 ACF／BMA／VACE.EXE 都沒有明文，推定是畫格點陣圖。
- dosgolem `apps/pod`（由 `apps/sosb` 複製）：換本作簽章、游標、逐字位移、底列選單呼叫端；路徑 B、C 的顯存換算改 13h；
  demo、作弊、地圖、性別欄註明「銀色匕首之謎值」不啟用。測試裡三處 planar 位移改 13h 後全過。`cmd/pod-overlay-shot` 由 sosb 版複製。
  `ce5974b` 推到 `dosgolem/pod`。
- `tools/g3-overlay-receipt.sh`：佔位譯文（`tools/trial_text.py` 取前作 key 相同 824 則）與佔位字型（複製銀色匕首之謎開發字型），
  主選單覆繪 6 筆、範圍外差異 0、原版色號同 G2 收據。見 `docs/re/004`。
- 清理：容器皆 `--rm`；`find . -user root` 0 筆。
