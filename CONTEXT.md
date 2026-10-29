# 目前真相

更新：2026-10-29（Asia/Taipei）。規則見 [AGENTS.md](AGENTS.md)，歷程見 [WORKLOG.md](WORKLOG.md)。

## 儲存庫

| 項目 | 值 | 回讀方式 |
|---|---|---|
| 網址 | <https://github.com/wicanr2/pools-of-darkness-cht> | `gh repo view --json url` |
| owner | `wicanr2` | `gh auth status`、`gh repo view --json owner` |
| 可見性 | PRIVATE | `gh repo view --json visibility` |
| Issues | 啟用；初始 G0–G11 為 #1–#12，規格 018–021 為 #13–#16 | `gh issue list` |
| git 身分 | repo-local `wicanr2` <wicanr2@gmail.com> | `git config user.email` |
| 授權 | RRSAL-1.0（[LICENSE](LICENSE)） | |

## 有效決策

| 日期 | 決策 | 來源 |
|---|---|---|
| 2026-10-29 | 接續8×8字模Go接管，並提高字型解析度。沿既有字型資產補殘留英數：2倍16×16、3倍24×24；原版索引背板與F12對照保留 | 使用者「8x8 字模 建議也用更好的解析度字體」；既有英文字型接入屬實作選擇 |
| 2026-10-28 | 收尾須完成原案全部功能與交付，包含 Go 繪圖接管；demo 中按鍵可中斷並回選單是明確必驗項。保留已驗收程式，依規格逐項接入，不以現有繁中＋HD 範圍縮減原案 | 使用者選擇「完成原案全部功能與交付」，並補充「golang 繪圖接管也包含」 |
| 2026-10-23 | 參考銀色匕首之謎，用 dosgolem 執行原版並覆繪繁中、做 HD；步驟完全比照銀色匕首之謎，定案表見 AGENTS.md 第 1 節 | 使用者指示 |
| 2026-10-23 | 中文題名「黑暗之池」 | 使用者指示 |
| 2026-10-23 | 錯譯修正對象是珍157 軟體世界說明書；英文手冊與手札授權上網取得 | 使用者回答 |
| 2026-10-23 | repo `pools-of-darkness-cht`；遊戲短名 `pod`，adapter 在 dosgolem `apps/pod/`，分支推到本 repo 的 `dosgolem/pod` | 使用者回答 |
| 2026-10-23 | 本輪做完 G0：文件、Git、private repo、初始 Issue | 使用者回答 |
| 2026-10-23 | 顯示模式 VGA 優先；dosgolem 跑不起來時回報，不自動改用 EGA | 使用者回答 |
| 2026-10-23 | 攻略來源：使用者提供《軟體世界》第 43–45 期掃描 31 張，放 `walkthrough/`（不進 Git） | 使用者提供 |
| 2026-10-23 | dosgolem 基底選銀色匕首之謎 `dosgolem/sosb`（`a6e0254`），理由見 docs/re/003 | G2 |
| 2026-10-23 | 入口照 `START.BAT`：`STARTUP.EXE` 設定 → 常駐 `snd.com`（`-root-tsr-shrink`）→ 監督佇列接 `CONTROL.EXE`；不常駐時 DEMO 會當掉 | G2（docs/re/003） |
| 2026-10-23 | 以 VGA（mode 13h，256 色）為準；dosgolem 可跑，不需退回 EGA | G2（docs/re/003） |
| 2026-10-23 | zip 的 `SAVE/` 遊戲內讀不到（缺 `savgam?.pty`、`*.pc`），測試存檔自建 | G2（docs/re/003） |
| 2026-10-23 | 手札譯名裁決：Zhentarim 桑塔林會（字頭隨珍157 桑提爾城）、spirit of fire 保留珍157「火神」、Silk 絲綢、Gothmenes 戈斯梅尼斯、Tanetal 塔內塔爾等，見 docs/re/005 | G4 |
| 2026-10-23 | 說明書其餘章節比照銀色匕首之謎做全冊 OCR 草稿與名稱對照；正文不對照錯譯，以完成中文化為先 | 使用者回答 |
| 2026-10-23 | 法術、怪物、物品名與銀色匕首之謎一致（前作術語表、前作遊戲內譯文優先）；前作沒有的才用珍157 或依前作命名類推（`text/zh157-names.tsv`，docs/re/006） | 使用者回答 |
| 2026-10-23 | 試譯文風定案；用字「裡」不寫「裏」 | 使用者確認（`text/TRANSLATION-GUIDE.md`） |
| 2026-10-23 | 攻略面板用自寫摘要：讀雜誌掃描後以自己的話整理每章路線、地點、要點與地圖編號說明，標明期數頁碼，譯名依術語表；雜誌原文不照抄、不進 repo | 使用者回答 |
| 2026-10-24 | G6 跨畫面盤點以自動探索兩輪為止（街上起點一輪、進店家神殿一輪），不再追加路線 | 使用者指示 |
| 2026-10-24 | HD 改由 Codex 全程負責盤點、請求包、生圖、驗收與接入；維持 3 倍、風格 B、原版配色與第一批畫風確認閘門 | 使用者明確選擇「由你全程完成 HD」 |
| 2026-10-24 | README 加入遊戲故事背景、原版 BIGPIC 圖與 HD 截圖；展示用圖片放 `docs/screenshots/`，原版與原型、執行期截圖清楚標示 | 使用者指示 |
| 2026-10-24 | HD b01 畫風確認：BIGPIC1 區塊 243 港灣城市、246 魔法戰鬥的奇幻手繪風格，繼續全量；原始生成圖、912×360 預覽與 SHA-256 保存於 `review/hd/` | 使用者明確選擇「採用這批畫風，繼續全量」；#18 |
| 2026-10-24 | PIC1-137 第 4 幀的蹲伏變身畫格，允許加完整紫色衣物，維持原姿勢與輪廓後重畫 | 生圖服務兩次拒絕後，使用者明確選擇「加紫色衣物重畫」 |
| 2026-10-24 | 新增遊戲視窗左右兩側海報，先上網找本作海報或封面；海報標題與文案須隨遊戲目前語系切換 | 使用者追加指示；版面與裁切先作示意 |
| 2026-10-24 | 海報採固定加寬，中央尺寸不變；3 倍中央 960×600、左右各 240×600，沿用已展示的原版盒繪裁切與獨立語系文字 | 使用者選擇「固定加寬，完整海報」；#21；`review/hd/poster/wide-zh.png` 為定版示意 |
| 2026-10-24 | README 補遊戲內地圖、上方工具列，以及遊戲封面的原版／HD 對照，直接顯示圖片 | 使用者追加指示；#25 |
| 2026-10-24 | 封面女王採完整藍紫色禮服，保留原姿勢、王冠、寶座與構圖；標題與美術分離以便切換語系 | 原服裝生圖被服務拒絕後，使用者明確選擇「採用禮服版」；#25 |
| 2026-10-26 | 依mm3直接渲染方法接管其他繪圖函式；接管的圖像運算要用Go演算法，固定x86語意轉寫不能當作全部Go繪圖完成 | 使用者指示；#45–#49、規格025 |

## 待定

### 收尾目前狀態（2026-10-29）

本表是目前工作入口。下面的 2026-10-26 交接表保留當時驗證範圍，不單獨作為重新開工依據。

| 項目 | 目前狀態 | 下一個驗收點 |
|---|---|---|
| 正式程式與 HD | `3a9e95e`、`d33254a`已啟用；776檔主題未改，前版`5b7a0d0`以pod-play-nativesurfacefillspans保留 | 後續繪圖接管另建候選，保留已驗版 |
| demo 中斷 #13 | CONFORMED：選單不逾時；對話、戰鬥中斷與重播、正常進入通過；F12／2倍終點同原流程，非預期閃現0 | 固定證據[020](docs/re/020-demo-control-runtime.json)，入口[規格018](docs/spec/018-demo-control.md#demo-conformed) |
| Go 繪圖 #49 | 常駐8×8字模單點局部CONFORMED；4096次逐點、192次完整原版、48次邊界與64種拒絕，正常33728點實際改Go | 正常另有1ED4:1A8E的1920 bytes命令未標HostWork，原因須診斷；全部Go仍DRAFT |
| 高解析字型 | 已認證英數格原生16／24像素呈現通過；圖示不重畫，譯文優先，F12原版保留，README有並列原始像素裁切 | 未識別／搬移後無文字來源的格子保留原版，不猜OCR或當作全遊戲字型覆蓋 |
| 不露原版 #16 | 正常HD閃現0；字型樣本有121格明示2→3倍HD切換，逐筆分類後非預期0，舊程式同腳本也121 | 已驗範圍見[字模與字型收據](docs/re/019-native-resident-glyph-runtime.json)及[前批完整收據](docs/re/019-native-bounded-runtime.json)；其他HD來源及玩家路線不外推 |
| 攻略、HD、工具列 #9、#10、#14、#15 | 既有收據範圍保留，未驗部分仍待驗 | 以目前程式核對缺項，不追加已停止的 G6 自動探索 |
| 配樂、多語系、作弊、推廣片 #11 | 原案仍須完成 | 按各自規格與試聽閘門完成 |
| 三平台封包 #12 | 尚未交付 | 乾淨重建、啟動、存檔與 manifest 驗證；平台限制明列 |

目前工作：#49常駐8×8字模及高解析英數已驗收啟用。
來源為`3a9e95e923b1985b78b897245cc9e298df9f4a87`，執行檔SHA-256為
`d33254acd6cb0c36bddad25dddbc179d55643daf762a511b8a06c519b25f7f5f`。
前版5b7a0d0以pod-play-nativesurfacefillspans保留，17195e4及較早回復檔仍留；原始資料、字型資產與776檔主題未改。

固定[019-native-resident-glyph-runtime.json](docs/re/019-native-resident-glyph-runtime.json)記錄4096次原版有序逐點、192次完整呼叫及48次四角／全前景／全底色。
完整SaveState及PortLog相同，8192個Go單點、64個完整字模及72832道CPU控制，兩個像素位置CPU回退0；64種來源／框架／游標／旗標／模型拒絕不改輸入。
保留原完整字模CALL後快速工作；小步數用既有HostBlock的兩個單byte範圍，不改通用CPU或CALL契約。
高解析層只認證穩定ASCII760 bytes；同表的控制圖示會變，保持原樣。20次16／24字格呈現、17次原生細節、九種來源條件與重複入口冪等、譯文優先及F12測試通過。
e30160f的全部drawblocks411.316秒通過；後續僅修顯示追蹤，213份drawblocks來源SHA相同。最新字模定向15.521秒、pod／bgm／xlate／session／machine／Xvfb前端與vet通過，兩批來源及log分列。
正常712250000步的CPU／RAM、有序事件、127檔及46份直接入口同原版。背景22614點與前景11114點實際改Go，共33728點；其他Go出口不變，總計145767計畫／3434188步、固定1018，不外推FPS。
正常HD與讀檔／地圖／攻略／工具列加英文、倍率、F12共14組基準截圖，差異只在認證ASCII格；另13組直接／舊合成逐張相同。
HD閃現0；字型路線原始121格全部是4792→4793的2倍切回3倍HD，舊版同腳本也121，非預期0。七種影像／metadata負對照拒絕，README兩張原始像素裁切顯示數值及位置保留。
前批[整個表面填色](docs/re/019-native-surface-fill-spans-runtime.json)、[基本圖元單步](docs/re/019-native-primitive-spans-runtime.json)、[橫直線與窄縮放](docs/re/019-native-axis-scale-runtime.json)、[表面搬移跨度](docs/re/019-native-surface-spans-runtime.json)、淡入／DAC、字模、RLE及HD逐格呈現均維持已驗版。
前版B97的四路demo、39對介面、六對真存讀檔與demo中斷完整收據留在[019-native-bounded-runtime.json](docs/re/019-native-bounded-runtime.json)，不冒稱本批全部重跑。

全部Go閘門仍DRAFT。此正常路線3343240 bytes繪圖命令中3341320標為HostWork，剩1920來自1ED4:1A8E，原因待診斷，不直接當作新實作或完成數。
小步數常駐字模不再列待辦；原始位址、較早診斷與正式驗收見規格025，靜態ES／字串寫入候選只供盤點，不當作像素覆蓋證據。
其他色盤位置、root與未驗條件保留；來源、BMHD／BODY查找及字模context準備、換列、回車與載入不冒稱Go解析器或文字排版器完成。

HD最新完成批次：[#48](https://github.com/wicanr2/pools-of-darkness-cht/issues/48)補齊DUNGCOM第5組地形。
23張新手繪通過數值、目視與來源回退診斷；已驗@3／@13沿用，第22格是純色，保留原色，共25張非純色地形。
正常7400格同目前四路基準的原版終點、127檔與最終合成，閃現0。776檔主題已原子啟用，全部舊素材與730檔備份保留。
固定[017-b26-runtime.json](docs/re/017-b26-runtime.json)綁定b26當時的191479d／2c58d1d，地形診斷保存於8a0c26b；目前程式及完整正常驗收以上方Go演算法批次為準。
其他地下城圖組、野外地形及未觀測正常遭遇仍未完成，不重新開啟G6探索。

繪圖仍是Go演算法與固定語意混合。已列跨度、圖元、圖像列、縮放、翻轉、轉場及部分色盤／等待由Go運算；其他VGA條件或失配仍保留原版路徑。
原盤點90項函式／helper與1432區塊只是來源索引，不是Go完成數。
固定語意雖跳過通用x86解碼與CPU.Step，仍不能稱原生Go繪圖；分類與限制以目前狀態表及[規格025](docs/spec/025-direct-render-commands.md#bounded-memory-ready)為準。
不新增已停止的G6探索，也不把未驗玩家場景、GPU／SIMD、片尾或動畫HD列為已完成。

| 項目 | 何時決定 |
|---|---|
| 片尾字幕與片頭片尾動畫HD：來源、辨識及呈現方式 | G9 |
| `pool4.cfg` 第 6 行 `F` 與後續 `0` 欄位語意 | 需要時 |
| 攻略摘要尚留英文的物品名（例 Silver Shield）、未定專名及來源歧義 | G8 後續校讀；地名與已定專名已對齊，見研究 013 |

## 交接（2026-10-26）

### 階段

| 階段 | Issue | 狀態 | 依據 |
|---|---|---|---|
| G0–G3 | #1–#4 | 完成 | docs/re/001–004 |
| G4 說明書 | #5 | 完成：手札 99 則錯譯校正、其餘章節名稱對照；正文不對照錯譯（使用者決定） | docs/re/005、006 |
| G5 譯文與字型 | #6 | 完成 | `tools/g5-text-receipt.sh` |
| G6 覆繪接入 | #7 | 完成：冷開機、建角、存讀檔、選單收據；自動探索兩輪（使用者指示到此為止） | docs/re/007、010 |
| G7 手札與查驗題 | #8 | 完成：查驗題 54 題 CONFORMED；手札實機觸發第 57 則、面板自動開出 | docs/re/008、009 |
| G8 攻略與地圖 | #9 | 攻略 78 段，地名與已定專名對齊；32 塊 GEO 雜湊、24 個 GEO 名稱、東西菲蘭 5 個攻略點共 7 格。兩城 2／3 倍 A/B 通過；已有正常冷開機讀檔與面板切換收據。其餘 8 個名稱、未觀測地點與共用場景未驗，#9 保持開啟 | docs/re/011–013；規格 008、009 READY |
| G9 HD | #10、#18–#20、#23、#24、#26–#44、#48 | PIC81狀態／78項、238張怪物、7張BIGPIC、地下城第5組25張地形及菲蘭121格街景已接入。CPIC244個來源由238張重畫與六組別名涵蓋；本組純色格保留。現行776檔同b26受測主題，未觀測遇怪與其他圖組仍未知。BIGPIC244、開場雙語、街景及片頭字幕有正常抽樣；其餘天空仍為素材，動畫HD與片尾未完成。最新Go演算法版本另有完整正常回歸，主題未改 | docs/re/014、017、018；[b26收據](docs/re/017-b26-runtime.json)、[最新程式回歸](docs/re/019-native-output-runtime.json)；規格010局部CONFORMED |
| README | #17 | 故事背景、原版 BIGPIC、HD 預覽與兩組繁中＋HD 實際 DEMO 截圖已寫 | README.md |
| 直接繪圖 | #45–#47、#49 | Go字模／共享函式／限定VGA矩形／圖片來源及列控制／正反向RLE／裁切翻轉／色號掃描／色盤陣列／淡出／完整RGB工作、固定x86語意區塊及原x86回退的混合實作，全部Go演算法尚未完成。已驗子集通過完整事件、四路DEMO、39對介面、六對新存讀檔及全部存檔；Algorithmic分開記錄已提交Go計畫 | [最新Go演算法收據](docs/re/019-native-clip-runtime.json)、規格025子集CONFORMED／整體DRAFT |
| README 補圖 | #25、#26、#30、#44、#48 | 地圖、工具列與下拉選單、原版／HD街景、遊戲封面、雙語海報、代表怪物及新增地形素材圖鑑、BIGPIC244正常對照共34張直接顯示；素材預覽與正常遊戲截圖分開標示，禮服及標題同核准預覽 | [街景](docs/screenshots/wall-receipt.json)、[封面](docs/screenshots/title-receipt.json)、[怪物](docs/screenshots/monster-hd-preview.json)、[地形](docs/screenshots/dungeon-terrain-hd-preview.json)、[BIGPIC](docs/re/014-bigpic-runtime.json) |
| 左右海報 | #21 | 固定加寬已接入，文字隨語系；中央與定版兩側逐像素一致，2／3 倍及 F12、原版狀態與 127 個 DOS 檔案比較通過。規格 023 CONFORMED，範圍為繁中／英文與已列抽樣 | docs/re/015；review/hd/poster/ |
| 堆疊物品顯示 | #22 | 正式詞表六項數量／加值回歸通過；既有正常物品 UI 的一般／反白 50 支箭 +2 已驗；未修改產品邏輯 | docs/re/016 |
| G10 | #11 | 為開場F8先補英文前端、專用字型及兩個已證實地圖名稱；2210格英文對話無缺字、閃現0，原版狀態與127檔一致。其他語系、英文手札／攻略、配樂、作弊及推廣片未完成 | 規格015局部CONFORMED；docs/re/014 |
| G11 | #12 | 未開始 | — |
| 規格 018–021 | #13–#16 | 018已驗收並啟用；019辨識本作CRT形式，工具列倍率待驗；020待正式驗收；021已列路徑非預期閃現0，其餘不外推 | 規格索引；研究014、017、018、020固定收據 |

規格狀態：001–003、006、023、024 已列範圍 CONFORMED；007–009 READY；010戰鬥、BIGPIC與開場、015英文前端局部CONFORMED，PIC局部READY，021局部抽樣已驗；004、005、013 DRAFT；其餘未開始或未驗（docs/spec/000-index.md）。

### 接手入口

| 要做什麼 | 從哪裡開始 |
|---|---|
| 從冷開機看中文畫面 | `tools/play-smoke.sh`（pod-play，xvfb）；送鍵腳本格式見腳本註解 |
| 從冷開機看已接入的 HD | `tools/g9-hd-play-receipt.sh demo`；`switch` 驗前端切換。需本機 `workplace/hd/theme/` 與已建置 pod-play；同狀態收據掛載契約見研究 014 |
| 從遊戲中的狀態接著做 | `workplace/probe/states/g6r-*.state`（建角、存讀檔各段，`tools/g6-stage.sh` 說明怎麼重播）；街上起點 `workplace/probe/g6-town/town.state` 與同目錄 `scratch/`；叛徒門觸發後 `workplace/probe/g6-explore/town1/end.state` |
| 盤點新地圖的缺譯 | `tools/g6-explore.sh <名稱> <狀態檔> [-yes …]`，再 `tools/g6-misses.sh <inv.tsv>` |
| 改譯文後 | `tools/font-dev.sh` 重建字型、`tools/font_coverage.py` 查缺字，再跑受影響的收據 |
| 重跑本作收據 | `tools/g6-create-receipt.sh`、`g6-saveload-receipt.sh`、`g6-menus-receipt.sh`、`g7-quiz-receipt.sh`、`g7-journal-receipt.sh`（約 20 分鐘）、`g8-guide-receipt.sh` |
| 地圖記憶體與呈現 | `tools/g8-map-evidence.sh`（八個定點取樣）、`g8-map-receipt.sh [east\|west]`（兩城 2／3 倍 A/B）、`g8-map-play-receipt.sh`（正常冷開機讀檔、原版查驗題、地圖與攻略互斥），細節 docs/re/012、013 |
| 地圖名稱與攻略點 | `text/map-titles.tsv`、`map-guide.zh-TW.tsv`，`tools/map_geo.py` 在 Docker 內從實際 GEO 清單產生 `map-geo.tsv`；`--check` 核對。只使用有證據的名稱、圖例與觸發格 |
| 地圖參考與攻略專名 | 研究 013、`tools/map_reference.py`、`map_match.py` 與 `docs/re/013-map-reference-crops.json`；`tools/walkthrough_terms.py --check` 查攻略已定專名，`walkthrough_build.py` 重建正式表，均在 Docker 內執行 |
| 片頭字幕 | 研究018及規格024；018-intro-timing.json／018-intro-runtime.json保存逐幀來源與正常視窗證據。八組繁中／英文已接入，5100格及F8／F2／F12通過；本機intro-timing-v3及g9-hd/intro-play-v1保存原始收據 |
| HD來源與去重 | 研究014及014-inventory-summary.json；`tools/hd_vga_inventory_summary.py`核對既有兩份catalog的908個狀態／887組可見圖樣，另列完全相同來源的別名。數字不含另行盤點的牆面及INTRO／FINAL，不當全遊戲完成率 |
| CPIC來源覆蓋 | `tools/hd_vga_cpic_coverage.py --inventory workplace/hd/inventory --theme workplace/hd/theme --out workplace/hd/cpic-coverage.json`，在Docker執行。驗尺寸、色號、使用RGB與成品SHA，`--require-complete`於缺項時失敗；只證明候選來源覆蓋，不證明正常遭遇 |
| README怪物素材圖鑑 | `tools/hd_vga_readme_preview.py`在Docker內從私人審閱的已驗收PNG製作原尺寸拼圖，預設輸出`docs/screenshots/monster-hd-preview.png`及逐件來源JSON；不當正常遭遇截圖 |
| BIGPIC原版來源 | `bash tools/bigpic-source-receipt.sh <名稱>`沿正常DEMO兩個空白鍵做26000格有界觀測；原始碼`tools/bigpic_source_probe.go`複製到隔離clone的workplace再建置，輸出在`workplace/probe/g9-hd/`。目前只有244完整色號與使用RGB已觀測，見[014-bigpic-source-proof.json](docs/re/014-bigpic-source-proof.json) |
| BIGPIC受測主題 | `tools/hd_vga_bigpic_theme.py --root . --theme workplace/hd/theme-b25-stage`於Docker內合併七張已驗SHA至既有theme.json，須先複製現行主題且不能覆寫已有BIGPIC；接入與正常驗收另依規格010，不以建置成功稱完成 |
| BIGPIC正常驗收 | `POD_BIGPIC_SAMPLE=1`讓`tools/vga-battle-play-receipt.sh`在既有7400格DEMO加入3650格大圖、F12及2倍抽樣；四路結束後先跑`tools/vga_battle_play_check.py --out <本次四路根目錄> --theme <受測主題> --original-baseline <舊四路根目錄>`，再跑`tools/bigpic_play_check.py --out <本次四路根目錄> --theme <受測主題> --proof docs/re/014-bigpic-source-proof.json`，均在Docker執行；結果見014-bigpic-runtime.json |

狀態檔、暫存目錄與截圖都在 `workplace/`（gitignore），換機器要另外搬。

### 下一步

1. G8：尚有 8 塊 GEO 未命名；其餘攻略點須取得本作觸發格證據，共用場景須先分辨，不能混用圖例。攻略物品名與未知專名繼續校讀。不追加 G6 自動探索。
2. G9 HD 與切換列（規格 010、011、020）、片頭片尾字幕。
3. 規格 018、019、021，G10 配樂與多語系，G11 封包驗收。

### 已知缺口與注意

- G6 沒覆蓋：商店購買清單、神殿治療、訓練升級、旅店、DEMO 以外的戰鬥、野外與地下城（docs/re/010）。
- 規格 007：冒險畫面按 F5 的路徑只有 pod-play 按鍵對照測試，沒有實機截圖。
- 規格 009：24 個 GEO 名稱有靜態牆面辨識，東西菲蘭有正常路徑快照及 2／3 倍呈現，沒有全路線驗收。四象限外的共用樓層只用整座塔名稱，蛛網據點只用通稱。正常前端收據使用 map／guide 腳本動作，實際 F7、V、Esc、F5、F6 路由由單元測試驗證。
- 冒險主選單 `Move Area …` 要先按 `M` 才能用方向鍵移動；遊戲沒掛 int 09h，送鍵用 BIOS 緩衝區（`-bios-key-names`），`-press` 送不進去（docs/re/009）。
- 自動探索會卡在只有 `Exit` 的事件選單（脫困鍵只有 Enter、Esc、空白）；長跑要給外層時限，盤點以每 100 鍵的快照為準。
- 從狀態重播時，狀態存檔前就印好的字（隊伍欄、底列選單）不會重印，截圖上是英文，不算缺譯（docs/re/007）。
- 主機不跑 python，分析一律進 Docker（WORKLOG 記過兩次誤帶）。

## dosgolem 基線

| 項目 | 值 |
|---|---|
| commit | 基底 `a6e0254`；正式 `b11290cb7d8754decf0924e1513b48aa17a783ea`，已推本專案 `dosgolem/pod`；執行檔 `7d51351c24d07ab6f1ec97d4a3dde7a663670dbec118a47843a246610b59f0af` 為乾淨Go 1.24.13建置。前版b484851以pod-play-demo-control保留 |
| Go／image | go1.24.13；`dosgolem-go-ebiten:1.24-r1` `sha256:8ed6bdee2bce…` |
| 冷啟動 | 0 種缺服務；mode 13h；`START.BAT` 程式鏈；片頭送空白鍵跳過，製作群後送 `p` 進建立隊伍主選單。收據 `tools/g2-startup-receipt.sh`，見 [docs/re/003](docs/re/003-dosgolem-startup.md) |
| adapter | `apps/pod/`（由 `apps/sosb/` 複製，路徑 A／B／C 換本作常數與 13h 換算）；`cmd/pod-overlay-shot`；自動探索 `cmd/pod-explore`（由 `cmd/sosb-explore` 複製） |
| play 指令 | `cmd/pod-play`（由 `cmd/sosb-play` 複製）：照 `START.BAT` 以 `snd.com` 常駐、佇列 `CONTROL.EXE`、`RootTSRShrink`（規格 432）；透過規格 433 明示 `C:\POD\SND.COM`，讓 `C:\POD\SAVE\` 正確對到存檔根。冷開機入口 `tools/play-smoke.sh`，正常讀檔收據 `g8-map-play-receipt.sh` |

## 中文化進度

| 路徑 | 狀態 |
|---|---|
| 印字程序 | resident 段 `1B2E`：逐字 `01DF`、路徑 A `0354`、路徑 B `0498`、路徑 D `030B`，與幽靈騎士同函式（docs/re/004） |
| 路徑 A／B／C | 規格 001–003 CONFORMED：冷開機、DEMO、建角、存讀檔、選單收據範圍外差異 0（docs/re/007），自動探索兩輪未命中只剩數字與引號（docs/re/010） |
| catalog | ui 1,357 筆（654 筆與銀色匕首之謎、625 筆與幽靈騎士 key 相同）、ECL 4,725 筆 |
| 手札譯文 | 99 則修正譯文 `text/journal.zh-TW.tsv`（90 則有修正）；錯譯表 354 條通過 `tools/errata_check.py`（docs/re/005） |
| 術語 | `text/glossary.tsv` 981 筆：原 949 筆加 G8 攻略地名與拼法別名 32 筆，研究 013；舊譯 `text/term-avoid.tsv` 38 筆。術語唯一、`tools/term_check.py`、攻略完整詞組對齊通過；物品與來源歧義另保留校讀 |
| 說明書其餘章節 | 名稱 159 列對照與裁定（`text/zh157-names.tsv`）；全冊 OCR 草稿未校讀（docs/re/006） |
| 介面與劇情譯文 | `text/ecl-text.zh-TW.tsv` 4,634 則（其餘 91 則是雜訊）；`text/ui-text.zh-TW.tsv` 980 則（654 沿用前作、326 新譯；377 則雜訊）；`tools/translation_merge.py` 0 錯誤。翻譯指引 `text/TRANSLATION-GUIDE.md` |
| 字型 | 源泉圓體開發字型 `workplace/font/pod-rounded.golemfnt`（16）、`-3x`（24），2,843 字，12 份正式顯示表缺字 0。DEMO 以正式譯文覆繪，開場到戰鬥 5 張截圖範圍外差異 0（`tools/g5-text-receipt.sh`）；DEMO 戰鬥畫面的隊員名、物品名與加值、施法訊息已覆繪（docs/re/007） |
| 查驗題 | 規格 006 CONFORMED：GAME.EXE 題目表 21 筆 54 題（`tools/quiz_table.py` → `text/quiz-answers.tsv`），改寫第 2、4、6 列並顯示答案（F9 切換），收據 `tools/g7-quiz-receipt.sh`（docs/re/008） |
| 手札面板 | 規格 007 READY：叛徒門實機觸發第 57 則，面板自動開出、翻閱模式可開（`tools/g7-journal-receipt.sh`，docs/re/009）；酒館另觸發第 9 則 |
| 攻略面板 | 規格 008 READY：《軟體世界》第 43–45 期自寫摘要 78 段（`text/walkthrough.zh-TW.tsv`，docs/re/011）；主選單三張、冒險畫面與地圖互斥收據無缺字（docs/re/012） |
| 地圖疊層 | 規格 009 READY：`DS:A9EA` GEO 指標、`DS:A9F3–A9F5` 位置方向，八個定點取樣已證實；32 塊 GEO、24 個 GEO 名稱、5 個攻略點共 7 格。東西菲蘭 2／3 倍 A/B 通過，原版狀態不變；F7／V／Esc、F5／F6 互斥測試通過，走過格另存前端；初始化的全零緩衝區略過（docs/re/012、013） |
| 跨畫面盤點 | G6 完成：自動探索兩輪（街上、店家門口）印字 1,857 次（第二輪計到第 400 鍵的快照），未命中只剩數字與引號；商店清單、神殿、訓練、戰鬥、野外未覆蓋（docs/re/010） |
| 片頭片尾字幕 | 片頭八組台詞來源、首次繪製／清除與正式雙語覆繪已驗，2／3倍及F8／F12通過、閃現0；片尾與完整動畫HD未驗（研究018、規格024） |

## 原始輸入

| 檔案 | 狀態 |
|---|---|
| `pools-of-darkness.zip` | 已安裝目錄，v1.10；`START.BAT` → `STARTUP.EXE`（設定 `pool4.cfg`）→ `CONTROL.EXE`（`VACE.EXE` 片頭、`GAME.EXE`、片尾）；`GAME.OVR` 魔數 `FBOV`；查驗題問手札頁碼、標題後第 N 字；`SAVE/` 是 1995 年玩家資料。見 [docs/re/001](docs/re/001-input-baseline.md) |
| `珍157-黑暗之池.rar` | 107 張 JPG：操作手冊 63 頁、冒險者手冊 110 頁，無缺頁；旅行手札 99 則（第 1–100 則，無第 13 則）在冒險者手冊 p.55–106；操作手冊附錄二是查驗題答案表。見 [docs/re/002](docs/re/002-reference-sources.md) |
| `walkthrough/` | 《軟體世界》第 43–45 期「完全攻略」三回，第 1–24 章，作者 Steeler；未轉錄。見 [docs/re/002](docs/re/002-reference-sources.md) |
| 英文手冊與手札 | mocagh.org 的 journal（規則＋手札 99 則，無第 13 則）、manual（Rulebook）、refcard、hintbook，OCR 品質中等。見 [docs/re/002](docs/re/002-reference-sources.md) |

## Glossary

- 黑暗之池：本作中文題名，與珍157 相同，使用者指定。
- 珍157：軟體世界珍藏版第 157 號，本作的官方繁中說明書，錯譯修正的對象。
- 操作手冊、冒險者手冊：珍157 的兩冊。操作手冊對應英文 Rulebook 加查驗題答案表；冒險者手冊對應英文 Adventurer's Journal，含旅行手札。
- 密碼表：珍157 操作手冊附錄二，查驗題的頁數、字數、英文標題與答案對照。_Avoid_：答案表
- 錯譯表：`text/errata.tsv`，記錄珍157 譯文與修正譯文的對照與證據。
- 手札：說明書中 Adventurer's Journal 的段落，遊戲內以編號引用。_Avoid_：日誌、日記
- pod：本作遊戲短名，用於 dosgolem `apps/pod/` 與程式識別。
- 銀色匕首之謎：第一參考專案，本機 `/home/anr2/cht/golden_box/secret-of-the-sliver-blader/`，短名 `sosb`。
- 克萊恩英豪：銀色匕首之謎的參考專案，規格 001–017 的原始版本，本機 `/home/anr2/cht/golden_box/champions_of_krynn/`，短名 `cok`。
- 覆繪：在原版已輸出的畫面上清除原文墨跡、重畫中文。_Avoid_：替換、打字串補丁
- 收據：dosgolem 以固定輸入重播產生、可重現的執行證據。_Avoid_：截圖證明
- 閃現：原版字或原版圖先原樣呈現、8 格內才被譯文或 HD 換掉（銀色匕首之謎規格 021 §2）。_Avoid_：閃爍（只用於使用者描述的現象）
- HD 批次：一組交給 Codex 重繪的原版圖塊，含請求與輸出兩個目錄。

### 待定譯名

尚無。

### Flagged ambiguities

- 本機目錄名 `pool-of-darkness` 是單數，官方名與 repo `pools-of-darkness-cht` 是複數，指同一專案，不改名。
