# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

所有工作規則以 [`AGENTS.md`](AGENTS.md) 為單一權威。本檔只保留使用者最初需求、常用指令與資料索引。
目前狀態見 `CONTEXT.md`。

## 原始目標（2026-10-23）

1. 參考 `/home/anr2/cht/golden_box/secret-of-the-sliver-blader/`（銀色匕首之謎）的經驗，用 dosgolem 執行
   原版 DOS《Pools of Darkness》，以輸出攔截覆繪繁體中文化並做 HD；步驟完全比照銀色匕首之謎。
2. 建立 private GitHub 儲存庫 `pools-of-darkness-cht`，遊戲短名 `pod`，Issue 先行。
3. 修正軟體世界珍藏版 157 說明書的錯譯，英文原文上網取得（`AGENTS.md` 第 3.1、3.2 節）；
   修正後的手札放進遊戲，不必再查說明書。
4. 顯示模式 VGA 優先，G2 確認（`AGENTS.md` 第 1 節）。

## 常用指令

全部在 Docker 內執行；原始輸入（本目錄）唯讀掛 `/in`，`workplace/` 掛 `/work`。腳本開頭的註解說明參數。
尚未移植的前作工具，對照銀色匕首之謎 `CLAUDE.md` 的指令表。

| 用途 | 指令 | image |
|---|---|---|
| 解壓原版 zip 與珍157、逐檔雜湊（目標目錄已存在就停） | `python3 -I tools/inventory.py /in /work/input` | `sosb-tools:bookworm-r1` |
| 逐檔雜湊表 `docs/re/001-input-files.tsv` | `python3 -I tools/inventory_tsv.py workplace/input/inventory.json <輸出>` | `sosb-tools:bookworm-r1` |
| 執行檔 MZ 標頭、overlay 魔數、關鍵字串 | `python3 -I tools/exe_probe.py workplace/input/dos/pod [關鍵字 …]` | `sosb-tools:bookworm-r1` |
| 掃描縮圖拼版（讀頁序） | `python3 -I tools/contact_sheet.py <輸出目錄> <欄數> <縮圖寬> <檔案 …>` | `sosb-img:r1` |

一次性 Docker 工作的固定旗標（`AGENTS.md` 第 8 節）：

```sh
mkdir -p workplace   # 掛載來源不存在時 dockerd 會以 root 建空目錄
timeout 600 docker run --rm --network none -u "$(id -u):$(id -g)" \
  --memory 2g --cpus 2 --pids-limit 128 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -v "$PWD":/in:ro -v "$PWD/workplace":/work -w /work <image> <command>
```

- 前作已建的 image 可唯讀沿用，例如 `sosb-tools:bookworm-r1`（python3、unzip、unar、poppler-utils、curl）與
  `dosgolem-go-ebiten:1.24-r1`（Go 建置）；沿用時把名稱與 ID 記入 `WORKLOG.md`。
- dosgolem 的建置與測試走 clone 內的 `tools/go.sh`。`apps/pod/` 在 G2 建立後才能跑
  `tools/go.sh test ./apps/pod/`；需要 X 的套件加 `DOSGOLEM_GO_CMD=xvfb-run tools/go.sh -a go test …`；
  自動模式跑 play 指令時 `docker run` 要帶 `--init`（銀色匕首之謎的做法）。

## 資料索引

| 項目 | 位置 |
|---|---|
| 原版遊戲（已展開目錄形式） | `pools-of-darkness.zip`（唯讀，不進 Git） |
| 軟體世界珍藏版 157 說明書（107 張 JPG 掃描） | `珍157-黑暗之池.rar`（唯讀，不進 Git） |
| 《軟體世界》第 43–45 期攻略掃描 | `walkthrough/`（唯讀，不進 Git） |
| 英文手冊與手札 | `workplace/input/en-docs/`（G1 下載，不進 Git） |
| 研究工作區 | `workplace/`（gitignore） |
| dosgolem 隔離 clone | `workplace/dosgolem/`，adapter 在 `apps/pod/`，推到本 repo 的 `dosgolem/pod` |
| 主要參考 | 銀色匕首之謎、克萊恩英豪、幽靈騎士、黑暗之后、光芒之池說明書 corpus（皆只讀，見 `AGENTS.md` 第 4 節） |
