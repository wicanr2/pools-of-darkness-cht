# 014 VGA 圖檔與 HD 首批

日期：2026-10-24。Issue：[#10](https://github.com/wicanr2/pools-of-darkness-cht/issues/10)、[#17](https://github.com/wicanr2/pools-of-darkness-cht/issues/17)。
狀態：研究與畫風原型中，未作全量或執行期 HD 完成聲明。

## 範圍與來源

使用者已授權 Codex 全程負責 HD，先做第一批並列圖確認畫風。維持 3 倍、風格 B 與原版配色。
不把前作 EGA 圖檔解碼當作本作 VGA 的證據。

| 輸入 | SHA-256 | 本作容器區塊數 |
|---|---|---|
| `BIGPIC1.DAX` | `bc3565c62c1654ee757d1f98f64bbb78e06bcd551c2e8593c41feb130b9f0f6b` | 7 |
| `PIC1.DAX` | `a3cf36ca3d51ca7960fd73b630489f019c5ad984808ad08bcf91186f67401ab7` | 49 |
| `TITLE.DAX` | `8a53adf4044527df86d2d983f6c4eeee440c3d09773e7cbd4166279ccafde757` | 3 |

原檔為 `workplace/input/dos/pod/` 的唯讀研究輸入。解碼圖與索引只寫 `workplace/hd/`。
工具：[dax_vga.py](../../tools/dax_vga.py)。原版圖片與重繪預覽、遊戲實際截圖分別標示。
首批請求工具：[hd_vga_request.py](../../tools/hd_vga_request.py)，輸出 `workplace/hd/requests/b01/`；生成結果放 `workplace/hd/out/b01/`。
前處理：[hd_vga_prepare.py](../../tools/hd_vga_prepare.py)；數值驗收：[hd_vga_validate.py](../../tools/hd_vga_validate.py)，
公式沿用前作並保留來源雜湊於 [hd_metrics.py](../../tools/hd_metrics.py)。VGA 使用各來源自己的 256 色 RGB 色盤，不套用 EGA 16 色表。

## 外部格式線索

參考 Gold Box Explorer 原作者的固定版本程式，僅作格式線索，不執行下載的程式：

- commit `4fd42bd4617eb8ae28654c3b548ea0910d878ac0`。
- [VgaBlock.cs](https://github.com/simeonpilgrim/goldboxexplorer/blob/4fd42bd4617eb8ae28654c3b548ea0910d878ac0/Common/Plugins/Dax/VgaBlock.cs)，
  SHA-256 `9b16215e041682f7f9ff1f7310e9b95bad9ff8eb9fdbae3ec624eb33ec92b8e0`。
- [EgaVgaPalette.cs](https://github.com/simeonpilgrim/goldboxexplorer/blob/4fd42bd4617eb8ae28654c3b548ea0910d878ac0/Common/Plugins/Dax/EgaVgaPalette.cs)，
  SHA-256 `cebfe319fa433ba4eca735f0b5183bbf58385a0af4bb593e1ce5678329d1e960`。
- 原作者對本作雙層壓縮圖片的[說明](https://forums.goldbox.games/index.php?topic=1033.30)，2010-08-07，回覆 43。

快照在 `workplace/hd/source/`；來源原碼不進本專案版控。解碼器以本專案的 Python 實作，未複製外部程式。

## 第一階段格式候選

以下偏移均為「外層 DAX RLE 解開後的區塊內偏移」，不是 EXE 位址或原檔絕對偏移。
外部程式提供線索，本作仍須以原版實際畫面核對。

| 偏移／資料 | 候選意義 | 等級 |
|---|---|---|
| `0`、`1` | 高度、以 8 像素為單位的寬度，各一 byte | 強推論；本作 BIGPIC 為 304×120、TITLE 為 320×200 |
| `6..7` | 連續影像數，little endian u16 | 強推論 |
| `8`、`9` | 色盤起始索引、色數減一 | 強推論 |
| `10` 起 | 每色 RGB 各 6 bit，三 bytes | 強推論；值域核對 |
| 區塊尾端 `影像數×寬×高` bytes | 每像素 8 bit 色號，逐列排列 | 強推論；待畫面比對 |

色盤與影像之間的資料保留為未知，不拿來猜動畫語意。PIC1 使用下節的格式，不套用此表。
解碼圖能產生只證明格式候選可讀，不代表正常玩家路徑或原版畫面已驗。

## 驗證與展示

盤點、同色號前作素材比對、執行期來源辨識、首批生成與收據於本篇追加。
正式接入依[規格 010](../spec/010-hd-theme.md)的狀態閘門。

### PIC 格式與執行畫面核對

外部線索另讀固定版本的 `VgaSpriteBlock.cs`，SHA-256
`b03426294a0ad56dcd5a96964012155f11129d385c742ef0be5f14d4d459bd43`。
以下均為外層 DAX 解壓後偏移。高度為 offset 0 的 u16，寬度為 offset 2 的 u16 乘 8；
offset 8 是額外資料數、9 是色盤起點、10 是色數，RGB6 資料從 11 開始。
第二層 RLE 從 `132+3×額外資料數+3×色數` 開始；literal 長度為控制值加一，repeat 為 257 減控制值。
基準幀編號在 `127+3×色數`；其餘幀與該幀逐 byte XOR。49 塊共解出 81 個 88×88 狀態。
動畫的播放時間尚未驗證；呈現層依原版當下色號選圖，不自行推進動畫。

RGB6 轉 RGB8 採 `(v<<2)|(v>>4)`。圖內未提供的色號，取原版已觀測的基礎色盤
`workplace/probe/g2-receipt/menu-a.pal`，SHA-256
`62ed0c9528b22f85a7ca8a4d313bc3eb6a80ddfde959b927d71980ab2f68a486`。

| 已知正常 DEMO 畫面 | 指令步數 | 原版矩形 | 結果與等級 |
|---|---|---|---|
| welcome，PIC1-131 | 350,000,000 | 24,24,88,88 | 色號與所有使用色的 RGB 完全一致，已證實 |
| fire，PIC1-043 第 0 幀 | 700,000,000 | 24,24,88,88 | 色號與所有使用色的 RGB 完全一致，已證實 |
| credits，TITLE 第 3 塊 | 300,000,000 | 0,0,320,192 | 上方 192 列色號一致；底部 8 列被原版改寫，不能直接用完整圖作 watcher |

收據在 `workplace/probe/g9-hd/pic-source-proof.json`，使用 dosgolem `ad17b65`、Go 1.24.13。
從既有正常冷開機 300M 狀態重播，330M 送空白鍵；沒有狀態或座標注入。
349M、699M 另存狀態供同狀態測試加速；該加速測試不取代首次繪製驗收。
解碼器、請求包、前處理與驗收在 `sosb-img:r1`，Python 3.13.16、Pillow 11.0.0 執行。

### 畫風與數值閘門

使用者已確認 b01 港灣城市與魔法戰鬥畫風，授權全量。原始生成圖、完整提示詞與 SHA-256 保存在
[HD 審閱頁](../../review/hd/README.md)。生成方式為內建 imagegen，無濾鏡放大代替重畫。

| 圖 | 結構重疊 | 色相中位差 | 驗收 |
|---|---:|---:|---|
| BIGPIC1-243 | 0.4777 | 9.91° | 通過 |
| BIGPIC1-246 | 0.4910 | 12.26° | 通過 |
| PIC1-131 | 0.4243 | 3.98° | 通過 |
| PIC1-043 第 0 幀 | 0.4560 | 17.25° | 通過 |

沿用前作閾值，未放寬。正對照為來源最近鄰，分數 1、色相差 0；243／246 交換的負對照分數低於 0.19、
色相差大於 123°，均拒絕。實際數值在 `workplace/hd/out/b01/metric-probe.json` 與各批 `validation*.json`。
目前只驗收上述四張，不能將 b02 的兩張通過寫成 81 個狀態已完成。
主題建置入口為 [hd_vga_theme.py](../../tools/hd_vga_theme.py)。
同狀態收據入口為 [hd_vga_receipt.py](../../tools/hd_vga_receipt.py)。在容器內將原版掛 `/orig:ro`、
dosgolem binaries 掛 `/app:ro`、正式譯文掛 `/text:ro`、字型掛 `/font:ro`、主題掛 `/theme:ro`、
`workplace/probe/g9-hd` 掛 `/state:ro`、獨立收據目錄掛 `/out` 後執行。
既有 `pod-overlay-shot` 的 `diff_outside_stamps` 未包含 Layer.Art 與 HD Hold，會將圖面誤報為越界；
本收據直接比較兩側 PNG，依已證實的圖面矩形判定，並保留原始報告供回查。

首批兩張的 10 次執行、6 組比較通過：3 倍只改圖框，2 倍完全相同，將來源色號 XOR 255 後完全不繪 HD。
所有比較的原版記憶體、暫存器、色號、原版 PNG 與 scratch 檔案一致，圖框外差異 0。
收據為 `workplace/probe/g9-hd/receipt/summary.json`。另從正常冷開機的 300M 狀態重播到 700M，
保留中途印字掛勾，產生實際繁中與 HD 截圖 `workplace/probe/g9-hd/demo-hd/`；已目視核對。
README 展示副本與來源雜湊列於 [screenshots/manifest.json](../screenshots/manifest.json)。
此步不是前端冷啟動或首次繪製閃現驗收，規格維持局部 READY。

後續正常冷啟動抽樣已完成：2050 格觀測 PIC1-131，5600 格觀測 PIC1-043 第 0 幀，兩張 HD 圖框
與最終素材逐像素一致，整張截圖也與上述重播版完全一致；5900 格與 7400 格兩條 DEMO 路線閃現 0。
輸入、主題、截圖與同狀態收據雜湊保存於 [014-hd-runtime-receipt.json](014-hd-runtime-receipt.json)。
可重播入口為 [g9-hd-play-receipt.sh](../../tools/g9-hd-play-receipt.sh)，`demo` 為冷啟動 DEMO、
`switch` 為 F12／F2 對應動作。實體功能鍵與腳本送入 DOS 的鍵名分開；`scale` 專供前端倍率動作。
前作 demo 控制簽章在本作命中 0，控制未啟用，歸規格 018，不能據此宣稱逾時或中止功能完成。

切換收據通過：正常冷啟動 2210 格，不切換／切換的最終原版步數、記憶體、色號一致。
F12 對應的 `toggle` 顯示英文原圖，再切回後整張 HD 畫面相同；`scale` 切 2 倍時圖框為原版，
回 3 倍後整張 HD 畫面相同。鍵盤對應與 toolbar 共用動作的測試通過，沒有把腳本的 `F2` 鍵名改成主機動作。
原版英文字的列印進度在不同步數會不同，不將兩個時間點的整張原版圖當成同狀態比較。
dosgolem 為 `e5c94ff`，只新增收據動作，沒有改呈現與 DOS 行為。
自動檢查器為 [hd_vga_switch_check.py](../../tools/hd_vga_switch_check.py)，與冷啟動腳本一起使用。

### 其他圖族盤點

同一 VGA 區塊候選格式另讀 CPIC 244、CBODY 128、CHEAD 28、COMSPR 44、DUNGCOM 156、WILDCOM 120、
SKYGRND 4、BACK 6，雙層 RLE 另讀 SPRIT 87，共 817 個影像狀態。
清單為 `workplace/hd/inventory/inventory-other.json`，每件保留檔案／色號／色盤 SHA-256 與區塊偏移。
尺寸與色號去重後 790 種；這是候選盤點，尚未驗執行期位置、動態色盤與透明度，不能直接接入。
`--archives` 只接受已辨識圖族；牆面定義與 8×8 資料不套用此格式。大量盤點使用 `--no-atlas`，避免巨幅總覽。

### b02 基準圖完成

49 個基準畫面均通過尺寸、RGBA、結構與色相驗收，逐件目視核對，最終 PNG SHA-256 保存於
`workplace/hd/out/b02/visual-review.json`。後續變更雜湊必須重新目視。
PIC1-151 與 PIC1-143 第 0 幀的色號及所有使用色 RGB 完全相同，沿用同一張生圖，未用全色盤雜湊差異誤判。
32 個其餘動畫狀態另作生圖與遮罩驗收，基準圖完成不代表整批 81 張完成。
[hd_vga_review.py](../../tools/hd_vga_review.py) 只匯出同時通過數值及目視雜湊的圖，
正式並列頁見 [b02 基準圖](../../review/hd/b02-base/README.md)。

後續 32 個狀態已全部完成，共 81 張通過數值及逐件目視。PIC1-137 第 4 幀經使用者確認加完整紫色衣物，
輪廓與姿勢仍依原圖，結構 0.535、色相 3.4°、遮罩外差異 0。
最終去重另確認 `044@5／044@7`、`047@1／047@3`、`143@0／151` 三組原版色號及用色 RGB 完全相同。
前兩組原先各生成一張，建置器拒絕同色號對應不同 HD 後，已依來源證據改共用同一重畫。
沿用工具為 [hd_vga_deduplicate.py](../../tools/hd_vga_deduplicate.py)，結果在 `out/b02/deduplication.json`。
完整主題 78 項，provenance 分開兩項 observed 與其餘完整色號條件式候選，不把候選稱作已實機觀測。
全量素材完成；完整主題的正常路徑驗收待重跑。

### 完整 PIC 主題與剩餘 BIGPIC 素材

完整 78 項 PIC 主題已重跑原先兩個觀測狀態的 10 次執行、6 組比較，全部通過。
記憶體、暫存器、色號、原版 PNG 及 scratch 相同；3 倍差異只落於圖框，2 倍及失配負對照不畫 HD。
收據保存在 `workplace/probe/g9-hd/full-receipt/summary.json`。這仍只覆蓋兩個觀測狀態，不外推其他候選已實機看過。
正常冷啟動 5900 格、兩張觀測圖與切換抽樣亦通過，閃現 0；與原先兩項主題的中央截圖及原版狀態完全一致。
固定雜湊與比較結果見 [014-hd-full-runtime-receipt.json](014-hd-full-runtime-receipt.json)。

b03 其餘五張 BIGPIC 已完成，與首批合計七張；完整對照及逐件來源、提示詞與雜湊見
[b03 審閱頁](../../review/hd/b03-bigpic/README.md)。240 與 245 為地圖聚落標記變體，
以原版差異遮罩限制修改；紅色部分為屋頂，沒有改成火焰。245 遮罩外差異為 0。

242 巨像圖前三輪色相未過。追查實際 RGB 色盤後，確認第二次重畫提示將藍綠陰影誤寫為草綠；
訂正為赭金亮部、藍綠陰影，沿用第一輪構圖重畫。最終結構 0.667、色相中位差 29.2°，
其餘四張結構 0.554–0.747、色相 4.3–13.2°。仍用原閾值，沒有改量測公式或放寬標準。
失敗版本保留於本機 `out/b03/`；正式匯出僅使用已通過並目視的版本。BIGPIC 的執行期接入仍為 DRAFT。

### README 地圖、工具列與封面

使用者追加直接展示地圖、上方工具列及封面原版／HD。Issue [#25](https://github.com/wicanr2/pools-of-darkness-cht/issues/25)。
`tools/readme-ui-receipt.sh` 沿既有 G8 正常冷啟動讀槽 A 的路線，補 F7／V 對應動作、工具列及下拉選單，
4800 格終點的步數、記憶體與色號與原收據完全相同。沒有新探索或座標注入。
腳本新增 `mapall` 前端動作，與實體 V 共用 `actReveal`；腳本 V 仍送 DOS，解析回歸通過。
來源樹為 dosgolem `a73fb2e`；原始截圖留在 `workplace/probe/readme-ui/`。

README 工具列細節只裁切實際視窗的 `[240,0,960,36]`，未改像素；地圖為東菲蘭全圖模式，預設走過模式另保留。
[gallery-receipt.json](../screenshots/gallery-receipt.json)及[展示圖清單](../screenshots/manifest.json)保存來源、裁切與輸出雜湊。
匯出工具為 [readme_gallery.py](../../tools/readme_gallery.py)。

封面來源為 TITLE 第 2 塊，完整原版圖已解碼。原服裝重畫被生圖服務在輸出階段以 sexual 分類拒絕，
request ID `c5f4b695-469f-42d8-af35-6cde7382b6b2`。另做完整藍紫色禮服示意，使用者明確確認採用。
不把單次服務拒絕當成原版素材不合法或內容判定。
禮服與分離的語系標題為核准呈現例外，來源、完整提示詞及雜湊見 [封面審閱](../../review/hd/b06-cover/README.md)。
由 [cover_prototype.py](../../tools/cover_prototype.py) 及既有 Chromium 匯出繁中／英文 960×600，
正式文字為 `text/cover.zh-TW.tsv`、`cover.en.tsv`，源泉圓體 Bold，成品 RGBA、alpha 255，已逐張目視。
此處只完成 README 素材預覽；TITLE 的繪製時機、底列覆寫與正常前端接入仍 DRAFT，未冒稱實際遊戲內 HD 封面。

後續接入已完成，見下方「開場與英文正常路徑驗收」。上段DRAFT為當時素材預覽階段，保留原始紀錄。

### TITLE 完整來源補證

同一正常冷啟動DEMO以 `-shot-source-dir` 在F10動作匯出原始色號及配對色盤。
F10在本次推進前執行，檔名1000／1100／1200分別是來源格999／1099／1199。
TITLE1、2、3的完整320×200均與原始來源逐色號及用色RGB完全相同，位置 `[0,0,320,200]` 已證實。
先前300M時credits底列被改寫的觀察仍正確，只是較晚的另一狀態，不能據此否定早期完整畫面。

工具：[hd_vga_source_frames.py](../../tools/hd_vga_source_frames.py)；固定報告
[014-title-source-proof.json](014-title-source-proof.json)。dosgolem來源擷取程式已保存在`f08d961`，
原始擷取於其前一階段未提交的相同只讀匯出實作，Go1.24.13。
來源收據 `workplace/probe/g9-hd/vga-transition/sources/`。TITLE接入另由#26推進，不把來源已解出算成執行期HD完成。

TITLE每塊提供完整256色色盤，外層解壓後偏移8為0、9為255；RGB8雜湊與上述正常完整畫格相同。
文字來源：[title.zh-TW.tsv](../../text/title.zh-TW.tsv)、[title.en.tsv](../../text/title.en.tsv)，
封面沿既有cover語系表。轉錄人名只依原圖，不補履歷或其他版本名稱。
底圖、文字及字型分離的審閱原型由[title_prototype.py](../../tools/title_prototype.py)與
[poster_capture.mjs](../../tools/poster_capture.mjs)的`titles`模式建立，結果在[開場素材審閱](../../review/hd/b07-title/README.md)。
初版CSS文字陰影遮住金色填色，已保存原型後移除該重複陰影；四張語系圖均目視，文字邊界通過。
新語系文字納入字集後仍2843字、補字0，兩個golemfnt雜湊不變。

### 英文前端及專用字型

首次正常F8收據仍是繁中，原因為本作尚無英文host表及字型，只有一個語言包時F8依既有契約不切換。
不是開場圖層未刷新。補[host-text.en.tsv](../../text/host-text.en.tsv)及兩個已證實菲蘭名稱的
[map-titles.en.tsv](../../text/map-titles.en.tsv)，其他英文地圖名顯示未知；英文手札與攻略仍未交付。
查驗題格式採本作`quiz.page`／`quiz.word`，不沿用前作不同題型的`quiz.entry`。

英文遊戲文字字型沿前作生成器[en_font.py](../../tools/en_font.py)，開發入口[en-font-dev.sh](../../tools/en-font-dev.sh)。
來源為前作唯讀`workplace/fonts-src/courier/courierprime__CourierPrime-Bold.ttf`，
複製到本作`workplace/fonts-src/courier/CourierPrime-Bold.ttf`。來源SHA-256
`ff1f38786c849d1c41fa8e447960abdb2bd75fdfb0cfcdeb524fad65a5af3638`，
同行OFL授權保存為`OFL.txt`，SHA-256 `9a755af092b494944c99f471be6fddd19b006a448fefdc4717e4ee0aa09a97b0`。
前作生成器原始SHA-256 `eb54a8cc7f18579bad98bda908f25c65bb7ca5ceab87f8e8e00f22e629483fc2`，本作只將說明中的輸出名改為pod。

| 字型 | 字數 | SHA-256 |
|---|---:|---|
| pod-en.golemfnt，16×16 | 96 | `121e92b01c9c907aa66b8e6e78155d0a368102e599bbb9784feaec5c7f2234e0` |
| pod-en-3x.golemfnt，24×24 | 96 | `c701ce90a4b3009e6126b0ccfc0dc042dc17c22dbf59c07b1e41c90678481db9` |

繁中字集初次重建漏將新TITLE表加入明確清單；已補工具後重新建置，仍2843字、補字0、兩個雜湊不變。
不得將先前漏收表的結果當成新表已驗覆蓋。英文前端、字型尺寸、本作問句行寬與原文輸入不變測試均通過。
載入鏈及驗收範圍見[規格015](../spec/015-multilingual.md)。

### 開場與英文正常路徑驗收

dosgolem `eb72e98`加入獨立語系圖層，`f3d019e`補主字型旁的語系專用字型載入，均已推送私人分支。
本作固定三張來源及完整色盤，以既有`theme.json`分置`localized/base`、`zh-TW`、`en`。
未知語系用無字底圖；未知色盤回退。底列改寫按8×8格撤除，F8清除前一語系保留。
PNG與文字例外不冒稱原版像素相同，原版色號／CPU／記憶體／存檔則不改。

建置入口：[title_theme.py](../../tools/title_theme.py)接受`--root`專案根與`--theme`輸出主題。
它先核對請求、原始生圖、目視成品、底圖、正式文案與字型雜湊，再匯出三組既有格式清單。
原版來源與完整提示詞、權利分類保存於[開場素材清單](../../review/hd/b07-title/manifest.json)，服裝核准沿b06清單。
實際執行檔SHA-256 `2654ff7615532378865cde61f761740e9a3e8f1d3c1828fd993be68c658870cd`。

| 驗收 | 結果 |
|---|---|
| 三路正常冷啟動 | HD／無HD／切換，1250格、120312500步 |
| 原版CPU、記憶體、色號、色盤 | 三路完全相同 |
| 127個DOS檔案 | 三路完全相同，前端語系偏好另計 |
| 三張開場的繁中／英文 | 與已驗收PNG逐像素相同 |
| F8／F2／F12 | 語系、2倍原圖、原版及返回均通過 |
| 17張完整視窗 | 中央與畫布相同、兩側與定版語系海報逐像素相同 |
| 首次繪製量尺 | HD冷啟動1250格，閃現0；刻意顯示原版的切換另驗像素 |
| 真實來源負對照 | 未知語系、改色、來源失配、credits底列改寫通過 |
| 現行主題 | 與最終測試版本213檔完全相同 |

固定收據：[014-title-runtime.json](014-title-runtime.json)，原始資料`workplace/probe/g9-hd/title-play-v3/`。
入口：[title-play-receipt.sh](../../tools/title-play-receipt.sh)與[title_play_check.py](../../tools/title_play_check.py)。
README原版封面另與無HD正常截圖相同，HD服裝及標題與使用者核准預覽完全相同；
來源見[title-receipt.json](../screenshots/title-receipt.json)。
地圖與工具列也沿相同G8讀槽A路線刷新到`readme-ui-v2`，4800格原版終點與既有收據相同，沒有增加探索。

英文專用字型另沿既有2210格DEMO對話，A七次、B一次均覆繪，沒有缺字，閃現0。
原版212712500步、記憶體`407acf3df80edac1df653803d20563f2e233633ca755a4101b1bc84ae4a7af0e`、
色號`fdcb9523281066334cc302987e63e467e3a703c6739983fe5b9a56ce68335102`及127個DOS檔案同既有繁中基準。
固定收據：[014-english-front-runtime.json](014-english-front-runtime.json)，
檢查入口[english_front_check.py](../../tools/english_front_check.py)，正常畫面已目視。
本項只涵蓋英文前端、開場與DEMO抽樣，英文手札、攻略及其他語系尚未交付。

### 正常街景的天空與牆面來源

2026-10-25，沿相同G8讀槽A路線重跑`readme-source-v3`，4800格、462000000步。
記憶體`0e6e8e53b475dfc58edd6841ba5aa55960bdb7c790c0b54d5710413ffa484973`及色號
`df3c6a4c0c2247dda3d7872e31be58af78e097c4200f418ecb6fc7dcf681dd5d`與原收據相同。
只增加配對來源匯出，不增加探索。下列偏移是DAX解壓後區塊內偏移，不是EXE位址。

- `8X8D1.DAX`六塊各16818 bytes，表頭8×8、256張圖，RGB6色盤32–143。
  像素從434起，每張64 bytes；色盤在346結束，中間88 bytes用途未知，不據此推繪圖規則。
- 本次配對色盤只符合第1塊牆面及`SKYGRND.DAX`第1塊的144–175區段。
  天空1／2／3雖色號相同，色盤不同，只能選實際吻合者；其他區塊不因像素相同共用錯色圖。
- 視窗`[24,24,88,88]`的121格可完全重建。90格只有天空／地面，31格疊牆面，使用17種8×8圖塊。
  牆面色號13取底下天空／地面，各格來源色號及提供區段RGB均吻合，沒有歧義或未解格。
  此為本次呈現組合的已證實契約；未逆推WALLDEF內部座標算法或宣稱其他街景均已驗。

入口：[vga_wall_source_check.py](../../tools/vga_wall_source_check.py)，四個參數`--orig`、`--capture`、`--binary`、`--out`。
原始收據`workplace/probe/readme-source-v3/`，輸出`workplace/probe/g9-hd/vga-wall-source/`，
固定輸入雜湊、17種來源偏移與121格明細見[014-wall-source-proof.json](014-wall-source-proof.json)。
工具沿Python3.13.16、Pillow11.0與NumPy容器；牆面預覽把13留底畫成灰色，原始idx不改。
這次只補來源證據，沒有把原版濾鏡放大當HD，也沒有接入3D牆面。

天空／地面四張已在b10逐張以內建生圖重畫，264×264不透明RGBA，
構圖0.455–0.832、色相差4.4–27.1度及最終尺寸目視通過。來源色盤各自綁定，不自行推定日夜切換時間。
素材與完整提示詞見[b10審閱](../../review/hd/b10-sky/README.md)，Issue [#29](https://github.com/wicanr2/pools-of-darkness-cht/issues/29)。
本批只完成素材，產品的牆面組合、逐格撤除與色盤守衛仍需另有READY契約。

### 菲蘭起點街景接入驗收

後續依規格010接入17種牆面及天空第1塊。`b47a617`以實際來源色號先分流，保留原EGA名稱相容性。
真實121格RGBA、13留底、逐格遮擋、三段色盤及錯誤PNG邊界、單色非場景、2倍與F12負對照通過。
實際戰場來源與擴充怪物候選的診斷仍通過，沒有把VGA街景套到戰場。

最初4800格正常路徑量尺發現9格閃字。配對來源證實來自查驗題列替換，街景候選命中0。
修正與正對照見[研究008](008-quiz-table.md)及[規格006](../spec/006-quiz-rendering.md)。
`dd3e417`繼承原文字列的提早顯示及色彩，逐格診斷舊版9、修正版0。

正常`wall-play-v3`兩路完整結束後重跑檢查，CPU、記憶體、色號、色盤及54個DOS檔案同既有G8基準。
HD改變67089像素，3D視窗外0，最後街景與已驗收逐件組合完全相同，完整4800格閃現0。
固定收據[014-wall-runtime.json](014-wall-runtime.json)，正式主題275檔與同一執行檔已啟用。
README街景、地圖、工具列八圖由同一正常收據匯出，來源及裁切記於[wall-receipt.json](../screenshots/wall-receipt.json)。
本次只驗菲蘭起點，不宣稱其他牆面、天空色盤或WALLDEF內部演算法完成。
