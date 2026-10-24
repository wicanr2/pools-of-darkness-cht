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

## 2026-10-23 G4 手札轉錄與錯譯校正（#5）

- 珍157 手札頁（冒險者手冊 p.55–106）以黑暗之后的 `dqk-ocr:ppocr5-r1` 與模型快取（唯讀）做 OCR（`tools/ocr-batch.sh`，26 張跨頁，0 錯誤）。
  先試切單頁，書溝位置不穩、右頁帶到左頁字，改成整張跨頁辨識、組稿時依文字框分布找書溝（`tools/zh_ocr_journal.py`）。
- 英文手札 PDF 第 15–25 頁以 `pdftotext -bbox-layout` 組稿（`tools/journal_en_ocr.py`）。四等分欄界把第二欄切進第一欄，改用實測欄界；
  第 38 則標題列被拆成兩行，已併回。頁碼在頁首花邊裡讀不到，以 PDF 頁換算。
- 查驗題：題目表的標題字串在遊戲檔、DAX 解壓與 6-bit 解碼後都找不到，延到 G7 執行期攔。
- 校讀與對照分四批派 sonnet 子代理（共用 `workplace/journal/INSTRUCTIONS.md`）。第一批先做，主 session 抽驗第 4、9、12 則與錯譯條目後才放其餘三批（併發 3）。
  子代理不能寫報告檔，報告改在回報訊息。第 13 則兩本書都沒有，手札共 99 則。
- `tools/journal_merge.py` 合併檢查；批次間 6 組譯名分歧與 Zhentarim、spirit of fire、第 88 則 Ward 由主 session 裁決，逐則替換譯文與錯譯表 fix 欄
  （一次性腳本在 session 暫存區）。第 34、72 則的「風暴」是天氣，不動。結果：99 則、錯譯 353 條、術語 194 筆、舊譯 10 筆，兩支檢查都通過。見 `docs/re/005`。
- 勘誤：本輪曾在主機執行一次與工作無關的 `python3 -c`（空操作）；主機不跑 Python，之後注意。
- 清理：容器皆 `--rm`；`find . -user root` 0 筆。
- 補：抽驗後三批各一則整段重寫的（第 44、67、83 則），修正成立。子代理標為沒把握的八處字以 `tools/ocr_crop.py` 裁原解析度局部親看：
  四處成立，「名喚／召喚」（26、74）、「變像／蠻像」（57）、「盛甲／盔甲」（97）分不清，錯譯表註明，第 97 則錯字條目降為假說。
  補 `j88-ward`（異體字統一），錯譯表 354 條；三支檢查通過。AGENTS 的 G4 完成條件註明說明書其餘章節待使用者決定。

## 2026-10-23 G4 說明書其餘章節、G5 試譯（#5、#6）

- 使用者：說明書其餘章節要對照，G5 開工。回報 G4 時說「銀色匕首之謎沒做其餘章節」不成立：前作做了全冊 OCR 草稿與下冊怪物、法術名稱對照
  （其 docs/re/006、`text/zh056-names.tsv`），本作照這個範圍做，正文逐段對照另行決定。
- 珍157 冒險者手冊 p.21–54（17 張跨頁）OCR 後，`tools/zh157_names.py draft` 只抽到 115 個候選：20 條怪物名只留英文、部分條目被拆行。
  改由主 session 逐張看圖轉錄，得 `text/zh157-names.tsv` 157 列；遊戲內英文取 UI 字串與 `MON1CHA.DAX`（`tools/monster_names.py`，124 個名字）。
  依 AGENTS 第 3.3 節採珍157 69 列、誤譯 5、珍157 自相矛盾 9、錯字 1（docs/re/006）。
- 術語表：`tools/glossary_import.py` 併入銀色匕首之謎術語表 579 筆（手札合併留下的 note 重複片段 28 筆一併去重）；試譯新詞 23 筆；
  `zh157_names.py check --apply` 改 46、增 25。`Magic-User` 從前作帶來時是「魔法師法術」（spell），本作 UI 是職業名，改「魔法師」。
  第一次 `--apply` 把物品 `Shield`（盾牌）改成法術名「魔盾術」，名稱表那列的 Edit 沒套上就先跑了；術語表那列已還原，法術改用 `Shield (spell)` 鍵。
  換掉的前作譯名 47 筆登入 `text/term-avoid.tsv`；衍生名（Wand of …、Neo Otyugh）同步改。術語表 821 筆，三支檢查通過。
- 翻譯指引 `text/TRANSLATION-GUIDE.md`（改寫自前作）。`tools/ecl_batches.py` 切 ECL 61 批、介面 5 批（`workplace/tr/`）。`ECL1` 區塊 1–3 是開發用跳關與戰鬥測試選單
  （含克萊恩的 `DARGAARD` 牆型，強推論）。
- 試譯：ECL 第 04（開場旁白）、05 批（菲蘭城港務長導覽、班恩現身）127 則，sonnet 子代理一個；主 session 改 8 處（「裏」統一為「裡」、Banite 祭司、Blackards 推定為
  blackguards 作「惡棍」、DEMURELY、GOOD HUNTING、COMMON ROOM）。子代理原稿留在 `workplace/tr/ecl/agent-raw/`。`tools/translation_merge.py` 0 錯誤，舊譯 0。
- 字型：源泉圓體與 Unifont 由銀色匕首之謎 `workplace/fonts-src` 複製到本作（GenSenRounded2TW-M `b8286a42…`、-B `050f9ebf…`、unifont_t `16963425…`）；
  `tools/font-dev.sh` 改寫自前作，加 `rounded_font.py`、`eten_font.py`。
- 覆繪：DEMO 開場路徑 B 7 則旁白以試譯覆繪，範圍外差異 0（`workplace/probe/g5-trial/`）。第一次截圖缺「齊聚」「姓」「稍」：字型在試譯前建、字集沒收試譯表；
  `font-dev.sh` 補收 `workplace/text-src/trial/`，重建後正常。教訓：譯文改了要重建字型。補 `tools/font_coverage.py`（AGENTS 第 6 節的字型覆蓋）：新字型對手札與試譯缺 0 字；拿前作佔位字型當負對照，列出缺字並結束碼 1。
- 勘誤：本輪一個主機指令誤帶了 `python3 -c 1`（空操作，殘留在指令列）；主機不跑 Python。
- 清理：容器皆 `--rm`。
- 補：名稱表只比對自己的鍵，前作錯拼法另立的列（`Delay Blast Fireball`、`Otto's Irresitable Dance`、`Protection Evil, 10' Radius`）留著舊譯，同一法術兩種譯名；
  改成與正確拼法同譯，`Cold/Hot Fire Shield` 隨火焰之盾改。名稱表補章節標題「德魯依法術」與「德魯依」（前作作得魯依），共 159 列；術語表由名稱表改 48、增 25。
  與 apply 前的術語表逐列比對，名稱表以外的改動只有這 12 筆衍生名與拼法列。舊譯 63 筆。

## 2026-10-23 G5 名稱改回前作、介面沿用、全量翻譯（#6）

- 使用者：試譯文風可以；法術與怪物名跟銀色匕首之謎一樣；說明書正文不對照錯譯，加速完成中文化。
- 名稱表 159 列改依前作：前作術語表 → 前作遊戲內譯文（同一英文字串）→ 依前作命名類推 → 珍157。沿用前作 57、類推 7、採珍157 13（前作沒有的）。
  術語表由 `workplace/manual-zh/glossary.before-zh157.tsv`（名稱裁定前）還原 57 列再重跑 `zh157_names.py check --apply`。
  前作術語表本身同一法術有兩種寫法（Protection From Evil 10' Radius 免受／不受；Disintegrate、Delayed Blast Fireball 正確拼法與錯拼法列不同譯），取前作遊戲內譯文。
  `ID` 前作術語表作「驗証」，前作遊戲內譯文是「驗證」，從之。
- 禁用表重建 38 筆：先前為擋前作名而加的 53 筆移除（現在就是採用名），留珍157 誤譯與不一致 9 筆，併入前作 term-avoid 19 筆（排除前作題名規則與「火神」）。
- `MON1CHA.DAX` 缺譯名的 60 個名字自譯（`workplace/tr/monster-terms.tsv` 併入術語表），124 個名字全有譯名；術語表 881 筆。
- 介面：與前作 key 相同的 654 筆沿用前作譯文（「裏」改「裡」）寫入 `text/ui-text.zh-TW.tsv`，舊譯 0；其餘 703 筆切 7 批（`workplace/text-src/ui-source-rest.tsv`）。
- 全量：ECL 59 批分 12 組、介面 7 批分 2 組，sonnet 子代理每組一個，同時 4 個。
- 全量交回：ECL 59 批、介面 7 批，14 個 sonnet 子代理（同時 4 個），每組 4–9 分鐘。多數子代理用 Write 寫檔後以 awk 補回空 note 的結尾 tab；
  兩個介面組改用 awk 由對照檔產生 done 檔。合併檢查以 key 與欄數為準，0 錯誤。
- 指引勘誤：「黑暗之池只用於題名」照搬自前作（前作題名是誤譯才有這條）；本作劇情裡的 Pool of Darkness 也作黑暗之池，第 1 組譯的「黑暗水池」已改。
- 統一：批次間分歧 11 種 18 處（陸鯊→土鯊、猛獁象→長毛象、Salamander 改火蠑螈以免與 Fire Lizard 火蜥蜴撞名、赫爾伯格→胡爾堡、艾文翠→精靈樹、鷹巢→龍巢、
  法師之門→魔法師之門、摩安德製造廠→摩安德工坊、油膏→藥膏、泰多姆城堡→泰多姆的城堡、黑暗水池→黑暗之池），替換前的批次留在 `workplace/tr/ecl/done.before-unify/`。
  新詞 65 筆併入術語表（946 筆）。`tools/term_consistency.py`（新增）剩 254 處未對上，抽看專名類都是語境變化（天氣的風暴、STORM THE GATE、敘述的坑洞）。
- 結果：`text/ecl-text.zh-TW.tsv` 4,634 則、`text/ui-text.zh-TW.tsv` 980 則；舊譯 0、珍157 舊名 0；字型 2,712 字、缺 0。
  `tools/g5-text-receipt.sh`：DEMO 開場到戰鬥 5 張截圖範圍外差異 0。戰鬥畫面的怪物名、隊員名、物品名組句仍是英文。
- 子代理回報的組句與語境疑點（程式插入名字、方位、數字的片段，玩家輸入字，手札引用與查驗題用語）整理在 `workplace/tr/review-notes.tsv`，轉到 #7、#8。
- 清理：容器皆 `--rm`。

## 2026-10-23 G6 pod-play 開機鏈、名稱 catalog 與共用表（#7）

- `STARTUP.EXE` 在 `pool4.cfg` 已存在時跑 7,520,306 道指令就以離開碼 0 結束，不開檔、不等鍵（`workplace/probe/g6-startup-existing/run.log`）。
  play 指令因此照前作做法在安裝時寫入設定檔，開機時從 `snd.com` 起。
- dosgolem 規格 432（CONFORMED）：`session.BootInput` 加 `Format`（EXE／COM）、`Queue`、`RootTSRShrink`，零值等於規格 235。
  READY 前自審發現 `.COM` 上限寫成 0xFF00，載入器是 0xFEFE，已改。單元測試含開與不開縮回的對照、10 種拒絕。
- `cmd/pod-play` 由 `cmd/sosb-play` 複製：必要原版檔改為 CONTROL.EXE、GAME.EXE、GAME.OVR、ADLIB.COM；安裝時寫 `pool4.cfg`（G2 收據同內容，SHA-256 `6161a336…`）、
  把 ADLIB.COM 複製成 `snd.com`；開機 `snd.com` 常駐、佇列 `CONTROL.EXE`、`RootTSRShrink`。測試在 xvfb 下全過。dosgolem `7046489` 推到 `dosgolem/pod`。
- `tools/play-smoke.sh`：空存檔目錄冷開機到建立隊伍主選單，原版色號雜湊 `fefa0514…` 與 G2 收據相同，主選單覆繪 8 筆。
- 名稱 catalog：`tools/name_catalog.py`（移植）由術語表產生 `text/name-events.tsv`、`text/name.zh-TW.tsv` 1,400 筆；`tools/menu_masters.py` 產生 `text/menu-masters.tsv` 32 組。
- 由前作沿用：`text/composed-templates.tsv`（引擎組句）、`text/host-text.zh-TW.tsv`（攻略目錄標題改「黑暗之池攻略」）、音譯資料（`translit-*.tsv`、`cmudict/`、`LICENSE-translit-table.md`）；
  LICENSE 第 2 條 (d) 照前作列出這兩項第三方資料（條款本身預留「日後收錄時逐項列於此處」）。`text/quiz-answers.tsv` 先放表頭，G7 填。字型 2,808 字。
- 勘誤：docs/re/003 寫 `pool4.cfg` 其後 13 行 `0`，實際 14 行（85 bytes），已改。
- 建角收據 `tools/g6-create-receipt.sh`：主選單 → 建角（精靈戰士 ROLAND）→ 存檔 → 加入隊伍（A 加入、E 離開；只按 Enter 是標記）→ 開始冒險 → 查驗題 → 菲蘭城碼頭。
  每段探針存狀態，再以 pod-overlay-shot 用該段開始前的暫存目錄快照重播（共用暫存目錄時重播會看到探針已寫好的存檔，出現「overwrite ROLAND?」）。
  各段範圍外差異 0；未命中只剩傷害骰、名字輸入回顯。補組句模板「Save {name}? 」「overwrite {name}?」。隊伍名單的玩家名音譯成「羅倫德」。
- 查驗題由路徑 A 分段印出（`37F7:01B2`–`029B`）：「(journal, page 2) what is word # 3 after the heading ABILITY SCORES」，查珍157 操作手冊附錄二答 HAS 後進入遊戲。
  收據只收這一題，題目變了就停。
- 路徑 B 第二組：碼頭對白參數與 DEMO 相同（1、0、14、22、38、17、1），規格 002 升 READY。新增 `docs/spec/000-index.md`（本作規格狀態）。
- 存讀檔收據 `tools/g6-saveload-receipt.sh`：婉拒導覽、紮營、存檔槽 A；冷開機載入本作槽 A、開始冒險、查驗題（第 4 頁 ALIGNMENT 後第 7 字 LIVES）回到街上。10 段範圍外差異 0。
  分段函式抽到 `tools/g6-stage.sh`，建角收據改用它（PROBE_ONLY 給只有探針吃的 `-call-args`；第一次把它也傳給 overlay-shot，旗標不認得而失敗）。
- 譯文修正：讀檔選單 `Pools Secret Exit` 的 Pools 指本作，子代理誤作「光芒之池」，改「本作 銀色匕首之謎 離開」，選單母串重產。見 docs/re/007。
- 勘誤：本輪又有一個主機指令誤帶 `python3 -c 1`（空操作），是第二次；之後組指令不再附這段。
- 選單盤點 `tools/g6-menus-receipt.sh`（檢視人物、物品、紮營、魔法、設定、區域地圖，12 段範圍外差異 0）。補底列整列模板三條、術語表 `Arrows`、`Darts`，名稱 catalog 1,404 筆。
  第一次重跑時用 `| head -3` 截輸出，腳本在印第 4 行時被 SIGPIPE 結束，後半段沒跑；改成輸出寫檔再看。

## 2026-10-23 G7 手冊查驗題（#8）

- 紮營底列第 4 格的「…」：有無整列模板各重跑一次（`workplace/probe/g6-ellipsis/`）都沒有，範圍外差異 0；當初截圖已被覆蓋，來源沒查到。docs/re/007 改記在「譯文修正」。
- 術語表筆數更正為 949（CONTEXT 先前寫 948）。#7 貼 G6 進度留言。
- 查驗題畫面的記憶體傾印只有當下那題的標題與答案（`DS:7DEE` 的 `05 'LIVES'`）；`-watch` 找到寫入者後反組譯 overlay 段 `5509`：
  題目表在 `DS:0x60+0x62k`、k＝Random(21)+1，每筆 3 題，答案加密（倒轉、每字加長度），答錯換題、三次為限。細節 docs/re/008。
- 反組譯沿用既有 image `sc-cht-disasm:capstone5.0.3-r1`（`sha256:3c56874558dc4f0133d8dbbf650d53f5272ce9596f1c2240189dc7cf0e413cfd`），只讀掛載傾印；腳本 `tools/re/dis16.py`。
- `tools/quiz_table.py`：以「第 4 頁 ALIGNMENT 第 7 字 LIVES」定位，GAME.EXE 映像 0x9F32 起 21 筆、54 題 → `text/quiz-answers.tsv`（鍵改為 page、word、heading）。
  英文手札 PDF 第 1 頁兩段正文對上 6 題，字序從標題後正文第一個字算起。答案取自原版執行檔，不轉錄珍157 附錄二。
- adapter `apps/pod/quiz.go` 改認六段題型，改寫第 2、4、6 列，標題留英文；`text/host-text.zh-TW.tsv` 鍵改 `quiz.page`、`quiz.word`、`quiz.answer`。單元測試含標題命中 catalog 仍留英文、前作表頭拒絕。
  `tools/go.sh test ./apps/pod/...`、xvfb 下 `cmd/pod-play` 測試通過。字型重建 2,808 字、缺 0。
- 收據：`tools/g7-quiz-receipt.sh`（題目表重產相同；答案顯示「答案：LIVES」與隱藏空白兩張，範圍外差異 0），建角與存讀檔收據的查驗題段重跑，範圍外差異 0、作答後進入遊戲。規格 006 升 CONFORMED。
- 收據第一次失敗：從 JSON 擷取第 6 列時沒處理字串結尾的跳脫引號，空白列被判成非空；改 sed 後通過。
- 手札引用（docs/re/009）：本作 ECL 文字沒有帶編號的引用句；`GAME.OVR` 0x2643 的 ECL 指令處理把參數 N 接成「This is recorded as journal entry N.」整句經路徑 B 印出。
  前作的 `JournalRefOf` 整句寫法與組句模板已涵蓋，補單元測試；`TestJournalTableFits` 掛 `text/` 跑過（99 則 3 頁內）。規格 007 升 READY，CONFORMED 要等實機觸發。
- 找觸發點：碼頭導覽答 Y 走完沒有記手札。導覽結束停在 (3,6) 面西、底列 `EXIT`；遊戲沒掛 int 09h，`-press` 的 IRQ1 鍵送不進去，方向鍵要走 `-bios-key-names`。
  先送 Enter 再送方向鍵時座標不變，原因未查。手動找路屬 G8 探索範圍，收據留到 G8。
- 勘誤：規格 007 改回 DRAFT。證據只有靜態反組譯，本作沒量到引用句（索引的狀態只算本作已有的證據）；docs/re/008 的資料段位置、009 的路徑 B 段與整句收到都標成強推論。
  方向鍵四組對照（有無先送 Enter）寫進 docs/re/009，給 G8 當起點。

## 2026-10-24 G8 攻略面板（#9）

- 攻略掃描 31 頁先跑 OCR 草稿（`tools/ocr-batch.sh`，dqk-ocr:ppocr5-r1，`workplace/walkthrough/ocr/`）。多欄排版，行序是亂的，只當輔助。
- 起初派子代理逐字轉錄，第 44 期子代理以「整篇重製雜誌文章」拒絕，第 43 期子代理改寫摘述；主迴圈停掉其餘兩個並問使用者。
  使用者決定面板放自寫摘要、雜誌原文不照抄（CONTEXT 決策、AGENTS.md 第 1 節與提交規則已改）。
- 三期摘要由 sonnet 子代理讀圖起草（依章整理、要點、地圖編號逐條、譯名查術語表），寫在 `reference/walkthrough-zh/`。
  主迴圈抽驗第 45 期第 21 章地圖 6 條與雜誌 p.70 相符；第 45 期結語補出處註解。
- `tools/walkthrough_overlap.py` 照抄檢查：三期 10 字窗口命中 1.6%–2.1%，整列照抄 0，命中多為專有名詞。
- `tools/walkthrough_build.py`（改寫自前作）產生 `text/walkthrough.zh-TW.tsv` 78 段；標題超過 40 格時括號說明移進正文。
  `tools/font-dev.sh` 收進攻略表，字型 2,842 字、缺 0；`TestGuideTableFits` 掛 `text/` 通過。
- `tools/g8-guide-receipt.sh`：pod-play 主選單開 F6，目錄、第 1 段、第 55 段三張截圖無缺字。規格 008 升 READY。

## 2026-10-24 G6 跨畫面盤點收尾、G7 手札收據（#7、#8）

- dosgolem `cmd/pod-explore`（由 `cmd/sosb-explore` 複製，戰鬥選單有 Quick 時按 Q；`e67971e`），腳本 `tools/g6-explore.sh`、`g6-misses.sh`、`g6-service.sh` 由前作移植。
  起點 `workplace/probe/g6-town/town.state`（存讀檔收據 `lanswer`）。
- town1：247 鍵、73 格，在叛徒門觸發「This is recorded as journal entry 57.」後停下；town2：410 鍵，進店家門口一帶，在酒館觸發第 9 則，
  最後卡在酒館門口的 Exit 事件選單，到外層時限 5,400 秒被終止。兩輪未命中只剩數字與引號。使用者指示兩輪為止，G6 收尾（docs/re/010）。
- 手札引用實機證實：路徑 B 在第 17 列第 1 欄一次收到整句，以組句模板覆繪；`tools/g7-journal-receipt.sh` 重播 247 鍵，觸發後面板自動開出，
  另以 `-journal-preview 57` 截翻閱模式。規格 007 升 READY；001–003 依建角、存讀檔與探索收據升 CONFORMED。
- 更正：上一輪把自動探索器寫成 G8、要隊伍位置位址；實際是 G6 工具、讀座標列。方向鍵失效的原因是冒險主選單要先按 M（docs/re/009 已改寫）。
- 收據腳本第一次插入翻閱步驟時 perl 的比對式沒對上，只改到註解；改用 Edit 加入後重跑。
- 補記：town2 被時限終止時送了 410 鍵，盤點快照停在第 400 鍵，010 與 CONTEXT 的印字數計到快照為止。
  攻略摘要的抽驗範圍照實列表（第 43 期只看結構；第 44 期補對第 12 章地圖；第 45 期對第 21 章地圖）。

## 2026-10-24 交接

- CONTEXT.md 加「交接」段：階段總表、接手入口（狀態檔、收據）、下一步（G8 地圖疊層先找隊伍位置位址）、已知缺口與注意事項。
- #5（G4）結案：手札錯譯校正與其餘章節名稱對照完成；正文逐段錯譯對照依使用者 2026-10-23 決定不做。

## 2026-10-24 G8 地圖疊層與正常前端讀檔（#9）

- 載入 Codex 路由的逆向／中文化技能、文件職責、規格閘門、dosgolem 與正常試玩入口，並以主機 gh 回讀 #9。
  沿用目前狀態表接 G8，沒有重開 G6 自動探索或改動前作、dosgolem 上游。
- 本作 GEO 指標 `DS:A9EA`、位置方向 `DS:A9F3–A9F5` 已由八個定點取樣證實。
  `GEO1.DAX` 32 塊各 0x402 bytes，區塊 10h 與東菲蘭原掃描形狀一致；原位址、bytes、輸入與快照雜湊見 docs/re/012。
  加規格 009，按 DRAFT、證據審查、READY 接入。訓練廳圖例 3、提爾修院圖例 13 的四個格子取既有 town1 印字紀錄。
- `tools/map_geo.py` 依實際 archive 清單產生 `text/map-geo.tsv`，只存雜湊；名稱來源 `map-titles.tsv`，攻略點 `map-guide.zh-TW.tsv`。
  其餘 31 個名稱留空，沒有猜測座標。地圖表失配、攻略格越界的負對照均拒絕。字型入口收進地圖表，2,842 字、缺字 0。
- 新收據：`g8-map-evidence.sh`（八個取樣）、`g8-map-receipt.sh`（2／3 倍原版、走過、全圖）、`g8-map-play-receipt.sh`（正常冷開機讀檔與面板切換）。
  各自的 Python 檢查器與產物位置登在研究 012；新研究與規格都掛回索引，研究 011 的舊待查補回鏈。
- 正常前端揭露的讀檔問題：sealed owner 沒有 ProgramPath，DOS 保留 RICH2，設定檔 `C:\POD\SAVE\` 對不到存檔根。
  規格 433 先 DRAFT、審查後 READY，新增選用 ProgramPath，空值保留舊行為；pod-play 明示 `C:\POD\SND.COM`。
  COM 絕對路徑開檔正負對照、非法輸入無副作用、session／DOS 測試與 go vet 通過。
- 原版 cold boot、同一槽 A 與同一鍵序，修正後確實載入羅倫德並出原版查驗題「第 1 頁 CHARACTERS AND PARTIES 後第 1 字」，答 YOU 後回冒險。
  地圖、攻略、再回地圖與關閉已有完整前端收據；實際 F7、V、Esc、F5／F6 路由由 `TestMapPanelKeyRouting` 驗證，不混稱實體鍵盤試玩。
- 初始化時曾把全零 GEO 緩衝區及垃圾座標記成走過格。確認 32 塊原版 GEO 都非全零後，規格 009 回 DRAFT 補條件、再審至 READY。
  無效緩衝區清除目前地圖快取，不記走過格；既有未知地圖測試、八份真實記憶體測試與前端鍵盤路由測試通過。
- 2／3 倍 A/B 收據乾淨重生：原版色號、完整記憶體、暫存器、scratch 內容相同，面板外差異 0、缺字 0；兩個倍率全圖已目視核對。
  `go test ./apps/pod/...`、Xvfb 下 `go test ./cmd/pod-play`、受影響套件 go vet 及收據腳本語法檢查通過。
- 環境與驗證修正：mode 13h 改用支援的色號擷取；Go 編譯限制並行度並沿用原 image；修正測試版面、型別及 basename 斷言。
  冷開機收據先以空目錄安裝，再放自建 SAVE，排除 1995 年大寫 CHRDATA 的干擾。最後一輪 GUI 曾在地圖畫面碰到 160 秒外層時限，容器已退出，保持同一鍵序改 300 秒重跑。
- 映像沿用 `dosgolem-go-ebiten:1.24-r1`（`8ed6bdee2bce`）、`sosb-img:r1`（`8b9012adb991`）、`python:3.13-slim`（`bf44cdfcb76c`）與 `alpine:3`（`71e9b7bae160`）；沒有重建或刪除 image。
- #9 仍未結案：其餘地圖名稱、攻略圖例與摘要術語對齊待做；規格 009 維持 READY，沒有宣稱全地圖驗收。
- 最後重跑通過：正常冷開機、原版查驗題 YOU、地圖／攻略／地圖／關閉的六張截圖與走過格檢查。
  五個原版存檔與輸入逐檔相同；程式、字型、資料表及圖片 SHA-256 綁在 `g8-map-play/receipt.json`。
  正常前端地圖、攻略、關閉後冒險畫面已目視核對。規格 433 升 CONFORMED。
- 隔離 dosgolem 已提交 `939ec6d`（開機路徑）與 `ad17b65`（地圖位址、防護與測試）。
- 兩個提交已推到本專案 private repo 的 `dosgolem/pod`；主儲存庫提交包含規格、研究、資料表與工具，不含原版、存檔或擷取圖片。
- Docker 收尾：本輪一次性容器均已退出並自動刪除，標記 `pod.task` 的容器為 0；專案工作根沒有 root-owned 檔案或誤建的 `.md` 目錄，抽查輸出均為 `1000:1000`。
- main `518e960` 已推送；[G8 進度留言](https://github.com/wicanr2/pools-of-darkness-cht/issues/9#issuecomment-6377882247)與[G6 讀檔補驗](https://github.com/wicanr2/pools-of-darkness-cht/issues/7#issuecomment-6377882575)已回讀，#9 保持 OPEN、#7 保持 CLOSED。

## 2026-10-24 G8 地圖命名與攻略專名（#9）

- 使用者「ok go」後續做地圖名稱、已觀測攻略點及摘要專名。開工回讀主機 gh 授權與 #9，核對 Codex 逆向／中文化路由；收尾重讀文件職責入口。
- 新增研究 013、提示書裁切 metadata 與 GEO 牆面參考／候選比對工具，均掛回研究索引。原版圖、PDF、記憶體與候選結果只留 `workplace/geo-reference/`。
- SSI 提示書對照全部 32 塊 GEO，24 個 GEO 已命名。馬爾瑪斯特與桑提爾分占上下半圖，提爾神殿四層對上四象限，沿用既有格式。
  馬可斯多層只採整座塔名；蛛網據點只採共用通稱。東菲蘭／黑暗菲蘭、兩個龍巢候選牆面相同或近似，沒有以排名猜命名，餘下八塊留空。
- 第二輪舊快照步數 21664000000，零新增指令傾印，完整 GEO 對上西菲蘭 `1/11`，位置 (9,12) 向南。
  既有 town2 招牌紀錄與雜誌圖例對上叛徒門、擲巨魔酒館、歐洛的神秘裝置，攻略點由兩點四格增為五點七格。
  沒有追加 G6 自動探索，也沒有宣稱已驗店內買賣或酒館事件。
- 新增 32 個攻略地名詞組／拼法別名，術語表 981 筆；完整英文詞組依既定專名對齊，保留英文括註、未知詞組與操作指令。
  校讀實際替換片段，修正「雷姆斯大地」為既定的「王國」，三份摘要重建為原有 78 段。
- `map_geo.py --check`、術語唯一、舊譯檢查、`walkthrough_terms.py --check` 與標題格數通過。
  資料檢查新增錯誤象限數、同圖例多格不同摘要的拒絕條件，兩個負對照均拒絕。
- 兩份圓體字型重建為 2,843 字，Unifont 補字 0，12 份正式顯示表缺字 0。
  `go test ./apps/pod/...` 通過，八份真實記憶體測試另以 `-count=1` 重跑通過。
- `g8-map-receipt.sh` 增加 `east|west`。兩城 2／3 倍、原版／走過／全圖共 12 張收據通過：色號、完整記憶體、暫存器與 scratch 相同，面板外差異 0、缺字 0。
  東菲蘭 3 倍、西菲蘭 2／3 倍已目視核對；馬可斯之塔攻略第 52 段另驗 3 倍，原版狀態相同且無缺字，已目視。
  收據 `workplace/geo-reference/receipt.json` 綁定最終資料、字型、執行檔與 13 張顯示圖雜湊。
- 研究環境沿用 Python 3.13.16、Pillow 11.0.0、Poppler 22.12.0 與既有 image，未重建。
  提示書批次轉圖曾碰到外層時限，只採已完整產生且實際讀過的頁面。額外字集斷言曾把字元與碼點整數直接比較，屬驗證腳本錯誤；正式字型檢查與實際畫面均通過。
- dosgolem 程式未改，分支維持 `ad17b65`；規格 009 維持 READY，#9 不關閉。
  未完成項為八個名稱、未觀測攻略點、同形共用場景辨識及攻略物品詞／來源歧義。
- Docker 收尾：本輪一次性容器均已退出，自動刪除；標記 `pod.task` 的容器為 0。專案工作根無 root-owned 檔案或誤建 `.md` 目錄，抽查輸出均為 `1000:1000`。沒有清理其他專案資源。
- 主分支提交 `67d4f3e` 已推送並以 `git ls-remote` 核對。[本輪進度留言](https://github.com/wicanr2/pools-of-darkness-cht/issues/9#issuecomment-6380601693)已回讀，#9 仍為 OPEN。主儲存庫與隔離 dosgolem 工作樹均乾淨。
- 映像 ID 回讀：`cok-pdf:bookworm-r1` `05e99381d00a`、`sosb-img:r1` `8b9012adb991`、`python:3.13-slim` `bf44cdfcb76c`、`dosgolem-go-ebiten:1.24-r1` `8ed6bdee2bce`；本輪未重建或刪除映像。

## 2026-10-24 本機工作區空間清理

- 使用者要求先清理空間。路由命中文件職責入口；本輪只處理本專案，不清其他專案、Docker image 或 volume。
- 清空已確認 gitignore、無執行中工作使用的 `workplace/dosgolem/workplace/gocache/`，保留目錄與擁有權，釋放 634,284 KiB。
  Go module 快取、可執行檔與隔離開發工作樹保留，後續首次建置會重建 Go 編譯快取。
- 九份早期診斷 log 無損壓縮成同目錄的 `.log.gz`：`g2-demo-shrink`、`g2-demo-tsr`、`g2-control-4`、`g2-demo-notsr`、`g2-control-5` 的 `run.log`，
  以及 `g3-demo-calls` 的 `run.log`、`pathd.log`、`pathb.log`、`glyph.log`。
  原本 253,434,207 bytes，壓縮後 21,161,871 bytes；每檔均確認解壓 SHA-256 與原檔相同，才移除未壓縮檔。
- 清單、原始 SHA-256 與大小保存在 `workplace/probe/cleanup-20261024.json`。
  需要舊 trace 時，在 Docker 內以 `gzip -dk <原路徑>.gz` 還原，保留壓縮備份；例如 `/probe/g3-demo-calls/run.log.gz`，
  將本專案 `workplace/probe/` 掛為 `/probe`，使用既有 `alpine:3`，依 AGENTS.md 加資源限制、非 root 身分與 `--network none`。
- 清空快取加壓縮 log 共釋放約 841 MiB；`workplace/` 由 1,786,564 KiB 降至 925,420 KiB。
  全磁碟同時有其他工作釋放空間，最後盤點可用 133.29 GiB；不把全磁碟的可用量變化算成本專案成果。
- 原始壓縮檔、原版與英文 PDF、掃描、字型來源、存檔、快照、現行收據及 OCR 中間圖保留。
  `geo-reference/receipt.json` 綁定的所有輸入及輸出 SHA-256 全數相同；主儲存庫與 dosgolem 工作樹的產品檔案未改。
- 所有清理與驗證在一次性 Docker 執行。沒有留下本專案容器，沒有 root-owned 檔案或誤建 `.md` 目錄；清理清單、壓縮檔與快取目錄均為 `1000:1000`。

## 2026-10-24 HD 首批接入與 README（#10、#17–#19）

- 使用者確認 HD 改由 Codex 全程負責。同步 AGENTS 與 CONTEXT，未再沿用舊分工停止線。
  核對逆向與 HD、README、文件職責路由，使用 imagegen 與逆向技能；沒有追加 G6 自動探索。
- 本作 VGA 圖不能套用前作 EGA 解碼器。以固定版本 Gold Box Explorer 原碼作線索，獨立實作 Python 解碼；
  BIGPIC 7、TITLE 3、PIC 49 塊共 91 個影像狀態，原始素材只在 workplace。
- 已知 DEMO 的 welcome／fire 圖與 PIC1-131／PIC1-043 第 0 幀逐色號、用色色盤完全一致。
  TITLE 製作群上方 192 列一致，底部 8 列被原版改寫，不採完整 320×200 圖作 watcher。
- 建立 #18 首批畫風與 #19 PIC 工作批次，保存完整提示詞、來源與輸出 SHA-256。
  使用者確認港灣城市、魔法戰鬥的奇幻手繪畫風並授權全量。原始生成圖與並列圖保存在 review/hd。
- 前處理與驗收使用每張圖自己的 256 色 RGB 色盤；結構至少 0.40、色相差小於 30 度，未放寬。
  首批兩張 BIGPIC 與兩張 PIC 通過，來源最近鄰正對照與錯圖負對照均符合預期。
- 規格 010 經來源與用色證據審查後升局部 READY，只允許已證實兩張 PIC 的條件式接入。
  主題維持既有 JSON 格式，未改 dosgolem 程式與遊戲資料。
- 10 次同狀態執行、6 組比較通過：3 倍只改圖框、2 倍無 HD、失配來源回退；
  原版記憶體、暫存器、色號、原版 PNG、scratch 內容均一致，圖框外差異 0。
  舊 pod-overlay-shot 的範圍統計沒包含 Layer.Art，另以 A/B PNG 的已證實矩形獨立核對，保留誤報原始收據。
- 正常冷開機既有 300M 狀態重播到 700M，保留中途印字掛勾，已目視兩張繁中＋HD 實際畫面。
  README 加入有原文來源的故事背景、原版 BIGPIC、重畫預覽與實際截圖，附來源與 SHA-256。
- 冷開機前端已跑到 DEMO 戰鬥，7400 格量測閃現 0；正在另抓首圖與切換收據。
  前作的 demo 控制簽章在本作命中 0，控制未啟用，列為既有規格 018 待驗，未掩蓋此限制。
- 追加六張中兩張首輪通過，四張構圖差異退件；重畫後其中三張通過，一張仍待修。
  全量生成與驗收持續進行，沒有將已生成或少數通過寫成全量完成。
- 沿用 sosb-img:r1、dosgolem-go-ebiten:1.24-r1，未重建或刪除 image。
  本次檢查工作根無 root-owned 檔案與誤建 .md 目錄，展示圖均 1000:1000；前端與生圖工作尚在執行。
- 補驗：正常冷啟動 5900 格，在 2050／5600 格確實呈現兩張 HD；整張 PNG 與 README 重播版逐像素一致，閃現 0。
  F12／F2 對應前端動作另以 2210 格雙跑比較，最終原版記憶體與色號相同，英文原圖／2 倍／返回 HD 均符合預期。
  英文原圖在不同步數會有列印進度差異，改以同狀態終點及圖框原版像素驗證，不拿不同時間點的文字作相等斷言。
- dosgolem 新增腳本 `scale` 動作，共用 F2 路由；腳本 F2 仍送 DOS，避免改壞既有收據。
  對應解析、toolbar 與面板路由測試通過，提交 `e5c94ff`。
- 生圖流程修正：第二張畫風參考會使模型把該圖場景帶入空白地板；該件退回，之後只用原版作圖片參考、用提示詞指定已定案畫風。
  原版唯一參考的下一批 13 張全數通過數值；仍依目視退回天候與石雕材質偏離者。重畫後八件均通過數值，已逐圖核對。
- 其他圖族候選解碼 817 個狀態，尺寸／色號去重為 790，尚未接入。動態角色色盤、透明度與牆面仍需本作證據。
- PIC 的 81 個狀態完成數值與逐件目視，32 個動畫遮罩外差異均 0。
  蹲伏變身畫格兩次被生圖服務拒絕，另一次構圖不合；使用者明確選擇完整紫色衣物後重畫，結構 0.535、色相 3.4°，已目視。
- 主題建置拒絕同色號對應不同生圖，據此核對出三組完全相同的原版可見狀態：044@5／7、047@1／3、143@0／151。
  各組確認用色 RGB 相同後沿用同一 raw，重跑 81 張驗收通過，主題 78 項。原版未觀測項仍標 conditional-candidate。
- 主分支 7170143、dosgolem e5c94ff 已推送。README #17 與首批畫風 #18 已關閉；完整 PIC 的正常路徑仍待重跑，#19 保持開啟。

## 2026-10-24 左右海報與堆疊數量提醒（#21、#22）

- 使用者追加左右海報及語系文字。上網找到原版盒裝正反面、1991 廣告掃描，以及官方 Steam 商店無字背景 1438×810。
  來源、日期、尺寸及雜湊保存於研究 015 與本機 source/poster；沒有把廣告稱為隨盒海報。
- 以原版無字插畫分左右裁切，標題與文案另依正式 TSV 顯示。展示固定加寬與 16:9 餘白兩種原型。
  使用者確認固定加寬：3 倍中央 960×600、兩側各 240×600；採用圖及雜湊保存在 review/hd/poster。
  規格 022 審至 READY，尚未改正式前端。首次 CLI 抓圖有視窗外框高度裁切，改用 Chromium CDP 明示 viewport 後重抓。
  三份原型中央均與既有截圖逐像素一致；繁中／英文文案已目視。沿用 Playwright v1.56.1-noble image，未新建 image。
- 使用者提供前作箭矢數量反白列漏譯圖。本作既有正常選單收據已有一般與反白的 `50 Arrows +2`，均命中 ready，已目視「箭 +2 50」。
  追加正式詞表六項回歸：30 支 +1、50 支 +2、前導零、一般／反白／獨立品名，輸入 bytes 與列寬檢查通過。
  沒有修改產品邏輯或前作；初版測試誤寫「箭矢」，已按正式 glossary 的「箭」修正。
