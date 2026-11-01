# 015 多語系

狀態：英文前端、簡中與英文面板各依下節明列範圍CONFORMED；日文、韓文仍DRAFT。
Issue：[#11](https://github.com/wicanr2/pools-of-darkness-cht/issues/11)、[#26](https://github.com/wicanr2/pools-of-darkness-cht/issues/26)。
沿用本作現有`apps/pod/lang.go`、`cmd/pod-play/lang.go`及前作規格015，不新增遊戲規則或語言存檔格式。

- 英文`host-text.en.tsv`的placeholder與本作繁中表對應，問句使用本作頁碼／字序形式。
  不直接使用前作`quiz.entry`格式。玩家可見原文由既有identityLookup以英文字型重繪，原版比較及存檔不使用譯文。
- 英文字型沿前作`tools/en_font.py`與Courier Prime Bold，2倍16×16、3倍24×24，填滿原版方形字格。
  實際來源、授權、SHA及可重現入口掛於研究014。繁中字型與原版點陣字不混用。
- `-lang-fonts`仍有最高優先。開發模式另可讀主字型旁既有`pod-<語系>.golemfnt`，沒有該字型就不載該語系。
  發行模式沿既有`font/pod-<語系>.golemfnt`。目前提供繁中、英文及簡中，F8循環，F12原版獨立。
- 第一次依系統語言、之後記住選擇，沿現有偏好檔；不重開遊戲、不送額外DOS按鍵。
  面板沒有該語言資料時停用，不沿用上一語言內容。英文手札與攻略依下方英文面板收尾驗收。
  英文地圖有32鍵，24個已知名稱與八個未知名稱；七個攻略點沿同座標，未知仍顯示未知，不露出繁中名稱。
- 驗收正式英文前端表、字型尺寸與覆蓋、語系載入與優先序、三張正常開場及F8／F2／F12，
  CPU、記憶體、色號、色盤與DOS檔案不變。正文、手札、攻略及其他語系後續範圍仍由#11追蹤。

驗收：dosgolem `f3d019e`，正式前端與專用字型、語系載入優先序、1250格三張開場的F8／F2／F12均通過。
另以2210格正常DEMO抽測英文原文的新字型，A／B路徑均覆繪、無缺字、閃現0；原版記憶體、色號與127檔同既有基準。
證據：[開場](../re/014-title-runtime.json)、[英文對話](../re/014-english-front-runtime.json)。其餘畫面未窮舉；英文手札、攻略與地圖面板的後續證據見下節。

## 簡體中文收尾

狀態：CONFORMED，正常玩家路線抽樣。Issue：[70](https://github.com/wicanr2/pools-of-darkness-cht/issues/70)。沿用既有LangPack、顯示查表、語系圖層及偏好契約，不新增DOS資料格式或規則。固定證據：[027](../re/027-zhcn-runtime.json)。

正式來源是繁中TSV。固定官方OpenCC1.1.9的`tw2s.json`搭配[text/zh-CN-terms.tsv](../../text/zh-CN-terms.tsv)，只轉顯示欄，輸出獨立`*.zh-CN.tsv`。事件key、英文原文、模板比對端、校驗雜湊、座標、數字、查驗答案及ASCII內容保持原樣。手札、攻略、地圖與姓名音譯一併提供；地圖名稱仍保留原本未知狀態，不猜新名稱。非預設語系需明列手札引用格式，[前端補充表](../../text/host-extra.zh-CN.tsv)由繁中既有DefaultJournalRef轉換；地圖顯示表採載入器既有map、title欄序，來源雜湊保留為第三欄。

語系切換沿`apps/pod/lang.go`及`cmd/pod-play/lang.go`，重譯已保留的原文，保留面板頁碼、已讀狀態、探索格及原版存檔。音譯資料放既有`translit-zh-CN/`，內部檔名及比對欄維持原協定。

源泉圓體沿既有OFL來源，16用Medium、24用Bold；工具列12／22也用原生字模。補齊其他語系的工具列字級選擇，讓既定小／中／大契約適用簡中；不改原版／英文或Go繪圖演算法。

開場三張TITLE與海報沿既定版面及已核准無字底圖，只重排簡中TSV文字。新語系圖層逐件檢查尺寸、文字邊界與PNG雜湊，沒有已驗資料時不借用繁中圖層。片頭字幕沿host-text原有字幕key，不改原版播放順序。

驗證採抽樣：全部新TSV欄位／key／ASCII與placeholder不變、字型覆蓋及乾淨重建；正常開場與一條正常讀檔的F8／面板／倍率路線，同初始狀態對照繁中CPU、RAM、索引、色盤及客體檔案，保存原尺寸擷取。缺表或錯字型須拒絕，不能用字型放大代替原生24字模。該批不含英文面板，後續見下節；日／韓仍DRAFT，既有候選九包不改寫。

固定OpenCC wheel、對應來源、授權與SHA入口：[工具來源](../../tools/vendor/opencc/manifest.json)。轉換只在Docker的Python3.11內離線執行；[轉換入口](../../tools/zhcn_text.py)列明顯示欄及不變欄。

重建入口：[字模](../../tools/zhcn_font.py)、[既定版面重排](../../tools/zhcn_assets.py)、[固定Chromium擷取](../../tools/zhcn_capture.mjs)。成品與版面雜湊由[簡中素材清單](../../review/hd/zh-CN/assets.json)保存，正式1536檔主題及十份字型依[SETUP](../../SETUP.md)由Git重建。正常抽驗入口為[zhcn_gui.sh](../../tools/zhcn_gui.sh)，[失敗分支](../../tools/zhcn_negative.py)、[八組既有字幕來源](../../tools/zhcn_acceptance_test.go)及[收據核對](../../tools/zhcn_check.py)分開保存，不用局部字幕呈現冒稱新跑完整片頭。

驗收引擎2ec2dfa，乾淨重編SHA與受測程式相同。繁中／簡中各4800格、462000000步，原版CPU／RAM、索引／色盤及132個客體檔案相同，沒有改寫存檔。每側19張正常開場、字級、讀檔、地圖、攻略、語系切換與手札空清單擷取；八組字幕另用既有正常來源驗2／3倍，動畫區域不變。兩種語系量尺原始121格皆為4749→4750的明示2→3倍HD切換，11×11個格子與切換前後來源鎖定，非預期0。`-shot-original`會再畫當前語系工具列，RGBA差異1260點只在工具列，原版像素依相同IDX／PAL核對，不把工具列當作原版差異。

全部1536檔及十份字型由最終Git輸入乾淨重建；舊素材與原六份字型SHA未變。缺簡中前端表及錯誤3倍字模均由實際前端拒絕。該批不含英文面板，後續見下節；日韓、配樂、無敵、完整動畫HD及HD patch仍未完成，舊九包候選固定在原來源，不覆寫。

## 英文面板收尾

狀態：CONFORMED，正常玩家路線抽樣。Issue：[71](https://github.com/wicanr2/pools-of-darkness-cht/issues/71)。沿既有LangPack讀入英文手札、攻略及地圖，不新增原版條件或存檔格式。固定證據：[028](../re/028-english-panels-runtime.json)。

手札從G4已校讀的99則英文稿匯入，保留第1–100則但沒有第13則、原標題、正文、段落與來源頁碼。彎雙引號正規化為ASCII雙引號，破折號以ASCII雙連字號保留，不從繁中反譯。原因是現有PackHalfWidth會把含非ASCII標點的段落改為半形組合，與Courier全形英數字格不合；完整ASCII段落沿原英文字格，不裁字。英文原稿與手冊來源依私人儲存庫例外保存，不擴張散布權利。

攻略以目前78段自寫繁中摘要為來源翻譯，章節序、標號、數量／加值、路線與來源頁碼不變，不重抄雜誌或新增攻略推論。英文地圖名稱以既有已辨識名稱為翻譯來源，八個未知名稱仍為空；七個攻略點的map、n、x、y及證據欄保持原值。

正式顯示來源為`journal.en.tsv`、`walkthrough.en.tsv`、`map-titles.en.tsv`及`map-guide.en.tsv`。使用原生Courier16／24及原面板分頁，檢查字型覆蓋、標題安全寬度、每則所有正文能排入頁面及頁碼切換，不能以裁字通過。

手札標題最多38個ASCII字元；攻略標題最多33個，留出Guide前綴後不超過40欄。超長標題改用完整可讀的短標題，原標題全文置於正文第一段，保留來源資訊；正文依原面板動態分頁，不因中文歷史頁數常數截斷英文。

抽驗一條既有正常讀檔與面板路線，保留語系選擇、已讀清單、攻略頁碼及地圖探索狀態，對照繁中原版CPU、RAM、IDX、PAL及客體檔案。正常空清單與已有手札的局部呈現證據分開保存；不新增G6探索、原版亂數或遊戲規則逆向。日韓與其餘原案功能仍未完成，舊候選九包不改寫。

原文來源：[G4英文校讀稿](../../reference/journal-en.tsv)；正式英文資料：[手札](../../text/journal.en.tsv)、[攻略](../../text/walkthrough.en.tsv)、[地圖名稱](../../text/map-titles.en.tsv)、[攻略點](../../text/map-guide.en.tsv)。攻略是本專案自寫摘要的英譯，保留來源既有疑義，不當作新規則證據。Ohlo's Mystic Devices、Troll Toss Tavern及Cloister of Tyr拼寫以既有town1／town2正常輸出核對，沒有新增探索。

可重現入口：[資料檢查](../../tools/english_panels.py)、[局部分頁與字型](../../tools/english_panels_test.go)、[正常玩家路線](../../tools/english_panels_gui.sh)、[錯值負對照](../../tools/english_panels_negative.py)、[收據核對](../../tools/english_panels_check.py)。99則手札共161頁、78段攻略共128頁，最長各4／5頁，所有正文非空白字元完整，全部列與標題界內；代表性四則手札、三段攻略以2／3倍原尺寸擷取。

上述入口都在容器內執行。Go工具映像沿SETUP，將引擎唯讀掛到`/engine`、專案唯讀掛到`/project`、本專案新建的驗證工作目錄可寫掛到`/out`，工具目錄唯讀掛到`/tools`。局部測試的overlay內容為`{"Replace":{"/engine/apps/pod/english_panels_test.go":"/tools/english_panels_test.go"}}`，存到`/out/test/overlay.json`；在`/engine`執行`POD_EN_ROOT=/project POD_EN_OUT=/out go test -overlay /out/test/overlay.json ./apps/pod -run '^TestEnglishPanels$' -count=1`。模組快取唯讀、編譯快取用容器暫存，沿AGENTS的資源上限與離線規則。

正常路線執行`bash /tools/english_panels_gui.sh`，需要合法原版、既有正常建立的槽A及目前程式／主題／字型，實際路徑列於腳本；缺少時不能改用狀態注入冒充。資料及負對照分別執行`python3 -I /tools/english_panels.py --root /project --receipt /out/test/data.json`、`python3 -I /tools/english_panels_negative.py --root /project --out /out`。固定收據核對另需乾淨Git匯出的相同資料檢查結果`/out/test/clean-data.json`及來源樹`/out/main-tree.txt`，最後執行`python3 -I /tools/english_panels_check.py --root /project --out /out --receipt /out/runtime.json`。原始紀錄保留在本機，Git保存固定收據與重生工具。

正常英文路線4800格、462000000步、17張擷取，同027繁中基準的DOS輸入、CPU／RAM、IDX／PAL及132個客體檔案。F6翻頁、F8英→簡→繁→英、F7地圖、F5空清單與2／3倍通過。量尺原始121格仍為4749→4750的明示倍率切換，來源與11×11格座標鎖定，非預期0。新觸發手札未重播；正文樣本是局部排版，不能替代正常新觸發證據。程式2ec2dfa／218b60f、1536檔主題及十份字型未改。
