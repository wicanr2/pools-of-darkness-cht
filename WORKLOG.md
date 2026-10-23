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
