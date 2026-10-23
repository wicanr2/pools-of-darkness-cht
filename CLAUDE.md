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
| dosgolem 探針（`PROBE_EXE` 選疊底程式，預設 `STARTUP.EXE`） | `PROBE_EXE=ADLIB.COM tools/probe.sh <輸出目錄> -scratch-dir /out/scratch -queue CONTROL.EXE -root-tsr-shrink …` | 走 dosgolem `tools/go.sh` |
| G2 冷啟動收據：設定 → 常駐驅動 → 片頭 → 主選單，重播比對 | `tools/g2-startup-receipt.sh` | 同上 |
| 文字模式畫面（`-dump-mem-at 'N:lin:B8000:4000:/out/x.bin'`）轉文字 | `python3 -I tools/b8000_text.py <x.bin …>` | `sosb-tools:bookworm-r1` |
| mode 13h 截圖（`-shots 'N:/out/x.raw,…'`；`-dump-at` 只支援 planar）轉 PNG | `python3 -I tools/indexed_png.py 320 200 <x.raw …>` | `sosb-img:r1` |
| GAME.EXE 解開後的映像（EXEPACK）→ `workplace/text-src/game-image.bin` | `tools/g3-game-image.sh` | 走 dosgolem |
| 文字 catalog | `python3 -I tools/ui_text_catalog.py --image … --ovr … --events text/ui-text-events.tsv --source workplace/text-src/ui-source.tsv`；`python3 -I tools/ecl_text_catalog.py --orig workplace/input/dos/pod --events text/ecl-text-events.tsv --source workplace/text-src/ecl-source.tsv` | `sosb-tools:bookworm-r1` |
| G3 主選單覆繪收據（佔位譯文與字型） | `tools/g3-overlay-receipt.sh` | 走 dosgolem |
| 簽章與記憶體遠呼叫 | `tools/re/sigscan.py`、`tools/re/fn_diff.py`、`tools/re/memcalls.py` | `sosb-tools:bookworm-r1` |
| 記憶體傾印（`-dump-mem-at 'N:lin:0:655360:/out/mem.bin'`）16 位元反組譯 | `python3 -I tools/re/dis16.py <mem.bin> <段> <起> <迄>` | `sc-cht-disasm:capstone5.0.3-r1` |
| 查驗題題目表（GAME.EXE 映像）→ 答案表 | `python3 -I tools/quiz_table.py workplace/text-src/game-image.bin text/quiz-answers.tsv` | `sosb-tools:bookworm-r1` |
| G7 查驗題收據：題目表重產比對、答案顯示與隱藏截圖 | `tools/g7-quiz-receipt.sh` | 走 dosgolem |
| 珍157 掃描 OCR（黑暗之后的 PaddleOCR 映像與模型快取，唯讀） | `tools/ocr-batch.sh <輸入目錄> <輸出目錄>` | `dqk-ocr:ppocr5-r1` |
| 手札草稿組稿 | `python3 -I tools/zh_ocr_journal.py --spreads <左頁碼>=<json> …`；`python3 -I tools/journal_en_ocr.py --bbox workplace/journal/en/bbox/p15.html … p25.html` | `sosb-tools:bookworm-r1` |
| 手札批次合併、錯譯與術語檢查 | `python3 -I tools/journal_merge.py workplace/journal/fix`；`python3 -I tools/errata_check.py`；`python3 -I tools/term_check.py` | `sosb-tools:bookworm-r1` |
| 說明書逐頁 OCR 草稿；名稱候選與裁定檢查（`--apply` 寫進術語表） | `python3 -I tools/zh_manual_draft.py --book <冊名> --spreads <左頁碼>=<json> … --out <TSV>`；`python3 -I tools/zh157_names.py check --names text/zh157-names.tsv --glossary text/glossary.tsv --monsters workplace/text-src/monster-names.tsv --ui workplace/text-src/ui-source.tsv` | `sosb-tools:bookworm-r1` |
| 怪物資料檔的名字 | `python3 -I tools/monster_names.py workplace/input/dos/pod/MON1CHA.DAX --out workplace/text-src/monster-names.tsv` | `sosb-tools:bookworm-r1` |
| 併前作術語表、本作新詞表 | `python3 -I tools/glossary_import.py --into text/glossary.tsv --from sosb=<前作 glossary.tsv>`（新詞表用 `--from keep=<TSV>`） | `sosb-tools:bookworm-r1` |
| 切翻譯批次；批次驗證與合併 | `python3 -I tools/ecl_batches.py --source workplace/text-src/ecl-source.tsv --out workplace/tr/ecl/batches`；`python3 -I tools/translation_merge.py --events text/ecl-text-events.tsv --batches … --done … --out …` | `sosb-tools:bookworm-r1` |
| 開發字型（譯文改了就重建，否則覆繪缺字） | `tools/font-dev.sh` | `sosb-img:r1` |
| 譯名一致性掃描（原文有術語、譯文沒用對應中文的候選） | `python3 -I tools/term_consistency.py --glossary text/glossary.tsv --source workplace/text-src/ecl-source.tsv --translations text/ecl-text.zh-TW.tsv` | `sosb-tools:bookworm-r1` |
| G5 譯文收據：正式譯文跑 DEMO 開場到戰鬥 | `tools/g5-text-receipt.sh` | 走 dosgolem |
| pod-play 冷開機冒煙（建 `workplace/dosgolem/workplace/bin/pod-play`） | `tools/play-smoke.sh [畫格] [送鍵腳本]` | `dosgolem-go-ebiten:1.24-r1`（xvfb） |
| 名稱 catalog、選單母串 | `python3 -I tools/name_catalog.py --glossary text/glossary.tsv --events text/name-events.tsv --out text/name.zh-TW.tsv`；`python3 -I tools/menu_masters.py --source workplace/text-src/ui-source.tsv --translations text/ui-text.zh-TW.tsv --out text/menu-masters.tsv` | `sosb-tools:bookworm-r1` |
| 字型覆蓋：譯文表要顯示的字是否都在字型裡 | `python3 -I tools/font_coverage.py --font workplace/font/pod-rounded.golemfnt <譯文 TSV …>` | `sosb-tools:bookworm-r1` |

探針註記：`-call-args` 一次只接受一組 `CS:IP`；`-call-length-string` 把前兩個參數當字串遠指標。

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
- dosgolem 的建置與測試走 clone 內的 `tools/go.sh`，例：`tools/go.sh test ./apps/pod/...`；
  需要 X 的套件加 `DOSGOLEM_GO_CMD=xvfb-run tools/go.sh -a go test …`；
  自動模式跑 play 指令時 `docker run` 要帶 `--init`（銀色匕首之謎的做法；本作 `pod-play` 尚未建立）。

## 資料索引

| 項目 | 位置 |
|---|---|
| 原版遊戲（已展開目錄形式） | `pools-of-darkness.zip`（唯讀，不進 Git） |
| 軟體世界珍藏版 157 說明書（107 張 JPG 掃描） | `珍157-黑暗之池.rar`（唯讀，不進 Git） |
| 《軟體世界》第 43–45 期攻略掃描 | `walkthrough/`（唯讀，不進 Git） |
| 英文手冊與手札 | `workplace/input/en-docs/`（G1 下載，不進 Git）；校正後的英文手札逐字稿 `workplace/journal/en/journal-en.tsv`、珍157 照原書用字的逐字稿 `workplace/journal/zh157/journal-zh157.tsv` |
| 研究工作區 | `workplace/`（gitignore） |
| 翻譯指引、術語、說明書名稱裁定 | `text/TRANSLATION-GUIDE.md`、`text/glossary.tsv`、`text/term-avoid.tsv`、`text/zh157-names.tsv` |
| 翻譯批次 | `workplace/tr/{ecl,ui}/batches/`（含原文）、`done/`（譯文）；子代理原稿 `workplace/tr/ecl/agent-raw/` |
| 字型來源 | `workplace/fonts-src/`（由銀色匕首之謎複製，雜湊見 WORKLOG） |
| dosgolem 隔離 clone | `workplace/dosgolem/`，adapter 在 `apps/pod/`，推到本 repo 的 `dosgolem/pod` |
| 主要參考 | 銀色匕首之謎、克萊恩英豪、幽靈騎士、黑暗之后、光芒之池說明書 corpus（皆只讀，見 `AGENTS.md` 第 4 節） |
