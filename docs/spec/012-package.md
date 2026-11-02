# 012 三平台候選封包

## 五語與片尾HD候選CONFORMED

Issue：[78](https://github.com/wicanr2/pools-of-darkness-cht/issues/78)。使用者2026-11-02指定不要求完整通關，已完成HD先打包與製作推廣影片。新候選固定`v.0.1.2-20261102`，來源沿目前已驗五語、2034檔主題、18字型及引擎`3fdeeb8`；不建立tag或公開Release，不覆寫舊十二包。

沿既有格式及三平台工具鏈，基本／HD patch各三包，基本／HD本機完整版各三包，共十二包。patch不附原版或1091份辨識索引／有序表，HD patch首次匯入自備DAX、INTRO／FINAL的BMA與ACF，重生全部2034檔。完整版只附47個原版遊戲檔，不附歷史SAVE，只放本機full-local。

版本、架構、manifest、逐檔SHA、五語與18字型、HD全量重建及原版權利檢查沿既有契約。實際Linux封包抽驗首次／重啟、五語／倍率、正常demo及存讀檔，Windows以Wine抽驗實際ZIP，macOS驗universal兩架構及依賴，原生Windows與macOS未驗仍明列。來源、原版與舊包SHA保持，封包通過後才移入dist-all。推廣片依[規格016](016-promo-video.md)分開驗證。

本機候選範圍CONFORMED。十二包共1,186,472,427 bytes，實際解開全部封包核對內容、架構、版本、字型／五語、HD與巢狀來源。Linux首次／重啟、十張五語／倍率／F12、DEMO中止與存讀檔通過；Windows實際ZIP以Wine驗日文及韓文重啟。新舊AppImage以相同原版目錄與種子重跑，存檔／冷讀終點及84份SAVE一致，歷史A/B十份資料以記錄的DOS大小寫對應吻合；不同輸入的歷史存檔終點不冒稱相同。原版125輸入與舊十二包保持。固定[036收據](../re/036-package-promo-runtime.json)及[交付入口](../re/036-package-promo.md)保存完整範圍；沒有tag或公開Release。

以下保留舊候選的來源與驗收，不能作為本版本已通過證據。

## HD patch 原版辨識資料匯入

Issue [#75](https://github.com/wicanr2/pools-of-darkness-cht/issues/75)。本節的檔案重建、前端與本機三平台候選範圍CONFORMED，固定[收據033](../re/033-hd-patch-runtime.json)。使用者2026-11-01確認B版HD海報；既有九包保留原驗收與SHA，不覆寫。

- 沿用`review/hd/runtime/manifest.json`版本1及目前1866檔主題、十份字型。輸入、解碼與輸出SHA來自既有`tools/rebuild_assets.py`及研究014／017／031，沒有新遊戲資料格式或規則。
- patch附已接受的PNG、JSON與原版SHA／解碼配方。850份原版索引及133份BMA有序寫入表由玩家自備DAX／INTRO.BMA重生，原始輸入逐檔核對既有`original_archives`。INTRO.ACF只核對SHA，不改其內容。
- Go核心放本作`apps/pod/hdimport`，只做檔案解碼與驗證，不接管CPU或改寫客體檔案。DAX依既有9-byte目錄與外層RLE；VGA區塊、sprite第二層RLE／XOR、434-byte牆面表頭及BMA均沿既有來源與邊界，不猜未知格式。
- 相對路徑正規化、重複輸出、未知kind／decoder、越界、雜湊、符號連結逃出來源根及既有輸出均須拒絕。驗完全部輸入及重生結果才建立輸出；失敗不可留下可被載入的半套主題，不覆寫原版、發行目錄或存檔。
- 前端接入沿既有首次原版匯入，在使用者資料目錄另建HD快取；現有完整HD與開發模式優先沿原路徑，基本patch不載入新分支。具體海報來源確認後才接入正式候選。
- 核心驗收以目前合法原版及正式清單重生全部1866檔，每檔同受測主題；原版與來源樹不變。另驗路徑、缺件、改SHA、非法kind／decoder及解析邊界。正常封包驗收須等海報定案後，依前節三平台與存讀檔閘門完成，不以核心測試宣稱patch已交付。

核心檔案級驗證通過，固定[收據032](../re/032-hd-patch-core.json)：來源main461b274、核心f09cf68；全部1866檔逐SHA相同，850份索引與133份有序表均由原版重生。首次重建、第二次不改檔案的快取重用、多餘檔案／改位元組拒絕通過；來源125檔與正式程式／主題不變。Windows amd64及macOS兩架構的核心測試程式可編譯，沒有執行或整包啟動聲明。

重現入口[hd-import-core-receipt.sh](../../tools/hd-import-core-receipt.sh)接受一個新收據名稱，在Docker內從固定main461b274與引擎f09cf68的Git匯出重跑。原版唯讀、輸出在workplace，需已有合法原版與Go模組快取，沿AGENTS資源及權限限制。本次已執行乾淨來源全部命令，編排最後加入相同main固定值與Go版本檢查，語法另驗；不需要第二輪相同輸出。

核心階段當時海報A／B待確認，原型在`workplace/package-hd-v1/poster-comparison.png`；`prepareHDSource`只供測試。後續由下方B版契約與033收據接續，正式main首次匯入及實際封包驗收已完成；不回寫032當時的檔案級範圍。

## 原九包候選驗收

狀態：本機九包候選範圍CONFORMED，正式原案交付仍未完成。Issue：[12](https://github.com/wicanr2/pools-of-darkness-cht/issues/12)、[69](https://github.com/wicanr2/pools-of-darkness-cht/issues/69)。固定證據：[026](../re/026-package-runtime.json)。

候選識別為`v.0.1.0-20261101`。不建立tag或公開Release，不宣稱配樂、多語系、無敵及動畫HD已完成。既有Go繪圖、繁中、已啟用HD與demo中止維持原範圍。

| 項目 | 契約 |
|---|---|
| 來源 | 固定本專案dosgolem/pod提交，以乾淨Git匯出建置；只修正啟動器殘留的1.2版提示為原版1.10，不改原版判定 |
| 平台 | Linux x86_64 AppImage、Windows x86_64 zip、macOS amd64與arm64 universal zip |
| 版本 | 以ldflags注入相同完整版號及引擎commit，程式、封包名與manifest一致 |
| 本機完整版 | 兩種呈現，各三平台；原版檔47個，只放本機`dist-all/<版本>/full-local/`，排除歷史玩家SAVE，不上傳 |
| 基本patch | 不含原版遊戲或擷取資產，玩家自備原版，原版呈現加繁中及高解析字型。HD patch仍待原版索引匯入與海報權利處理，不冒稱已交付 |
| 存檔 | 首次原版匯入與之後存檔用既有使用者資料目錄，發行目錄與自備原版只讀 |
| 資產 | 目前1525檔主題及6份字型，依固定重建清單逐檔核對。b44六張動畫素材不冒充正式接入 |
| 授權 | 隨包附RRSAL-1.0、字型及引擎／第三方授權；本機包的原版權利邊界明列 |
| 工具 | 既有Go1.24.13與macOS SDK15.5映像，固定官方AppImage type2 runtime及現有mksquashfs，不新增重複映像 |
| 驗收 | 乾淨來源、版本、ELF／PE／Mach-O架構、universal兩架構、封包完整性、檔案清單與SHA-256、首次匯入、重啟與存檔位置、缺原版拒絕、正常DEMO抽驗 |
| 限制 | Linux容器及可用Wine實際驗證；macOS僅結構／依賴檢查，真機未驗。未通過項目不能寫成已驗 |

目前前端`cmd/pod-play/launcher.go`、`install.go`已有匯入、雜湊、資源搜尋及存檔路徑，另有既有前端測試。HD與字型來源沿[重建入口](../../SETUP.md)、[固定主題清單](../../review/hd/runtime/manifest.json)及[b43收據](../re/017-b43-runtime.json)。

候選輸出依既定`dist-all/`契約。九包保存在`dist-all/v.0.1.0-20261101/`，逐包SHA、manifest與實際驗收範圍由固定收據記錄。基本patch三包，基本與HD本機完整版各三包；不含HD patch。

編排入口為[package.sh](../../tools/package.sh)，使用及工具映像限制見[SETUP](../../SETUP.md#候選封包)。本輪逐步執行下列容器工具通過；編排入口僅完成語法檢查，沒有另跑第二整輪。

容器內編譯入口為[package_build.sh](../../tools/package_build.sh)。固定AppImage runtime、MIT及內含程式庫授權、對應來源與所有SHA保存在[工具來源清單](../../tools/vendor/appimage/manifest.json)，採固定20251108版，不依賴浮動continuous下載。

[package_payload.py](../../tools/package_payload.py)建立三平台payload與ZIP，驗證版號、二進位架構、固定主題／字型、原版檔數、授權及patch無原版資料。Windows README採UTF-8 BOM與CRLF，ZIP以zipfile產生並檢查非ASCII檔名旗標。圖示是本專案SVG，不引用原版美術。

[package_appimage.py](../../tools/package_appimage.py)以固定runtime及`mksquashfs -mkfs-time 0 -all-time 0`組合AppImage，核對runtime偏移與SquashFS魔術值，彙整各包SHA與大小。產生後仍須解開驗證與正常啟動，不能直接登錄為完成交付。

Linux實際封包入口為[package_smoke_linux.sh](../../tools/package_smoke_linux.sh)。base及hd沿既有3800格DEMO對話中止與正常進入腳本；另驗不附原版的patch沿用已匯入資料。save模式使用既有正常G6自建槽A存檔作測試輸入，以正式前端另存槽B再冷啟動讀B，檢查檔案只寫入使用者資料目錄，不注入執行期狀態。

Windows的[Wine抽驗入口](../../tools/package_smoke_windows.py)從實際ZIP解開，驗版本、1250格正常開場與三張截圖、首次匯入至Windows使用者資料目錄。Wine與原生Windows真機分開記錄；不能用此結果稱為macOS真機通過。

[package_verify.py](../../tools/package_verify.py)解開實際九包，核對內含manifest、SHA、字型、主題、原版47檔及patch外洩邊界，再核對Linux同狀態、demo中止、冷存讀檔與Windows Wine收據。基本版量尺原始23格另依固定相鄰畫格分類：2804／2805回合原版索引與色盤仍是同一紅色背景，中文題名提前出現，尚未有後續10／13／15色號字形；保留原計數、來源與原尺寸目視證據，不改量尺。HD原始計數0，兩者非預期0。來源或事件不同就拒絕，不能用此分類排除其他閃現。

## B版海報與HD patch接入READY

使用者以「ok 繼續囉 B」確認[核准原型](../../review/hd/b47-hd-posters/README.md)，三倍兩側合成須逐像素相同；2倍維持既定版面、語系文字與中央畫布。四張背景改用已核准生圖，原版盒繪與比較圖不進patch。

發行模式先沿原版匯入取得使用者game目錄，再讀`hd-source/manifest.json`及其repo-file素材，於使用者`hd-cache/<清單SHA>/`重建完整主題。現有完整HD優先，明示-hd及開發-save模式不改；無hd-source的基本包沿原路徑。重啟驗完整快取，來源缺件／SHA不符／快取改檔拒絕，不寫入自備原版或發行目錄。

新候選固定v.0.1.1-20261101，只產三平台HD patch；不附原版EXE／DAX／BMA／ACF或idx／ops，後兩者首次匯入重生。PNG、JSON、十份字型、全部三語TSV及音譯資料、授權與原創來源依固定清單保存。封包掃描每個成員及巢狀引擎來源，原版檔雜湊與擷取副檔名不得出現。Linux實際AppImage驗首次、重啟、demo中斷、另存／冷讀；Windows實際ZIP以Wine抽驗。macOS只驗universal架構與依賴，真機未驗保持。舊九包與正式已驗程式在候選通過前保留。

<a id="hd-patch-conformed"></a>

## B版HD patch候選CONFORMED

2026-11-01，固定[033收據](../re/033-hd-patch-runtime.json)。來源main `7ce48d7`與引擎`661f7dc`，從乾淨Git匯出建置三包。交付在`dist-all/v.0.1.1-20261101/patch/`，合計396,377,338 bytes；SHA與實際封包抽樣在同版本`SHA256SUMS.json`及`smoke/`，不建立tag或公開Release。

| 閘門 | 結果 |
|---|---|
| 完整素材 | 1866檔主題、十份字型逐SHA相同；983份原版辨識資料首次匯入重生 |
| 封包內容 | 三包逐檔與manifest相同，ELF／PE／universal及macOS依賴通過；巢狀來源掃描無原版檔案 |
| B版海報 | 實際1979格完整視窗逐像素同核准原型；三語與2／3倍文字、F12英文原版抽驗通過 |
| Linux首次／重啟 | 實際AppImage各5100格，22張完整視窗、來源、終點與客體檔案相同，1866檔快取未改內容或mtime |
| 片頭量尺 | 原始599保留，其中559為明示切換、40為原版及實際畫面全黑轉場；非預期可見0，沒有調量尺 |
| demo與存讀檔 | 3800格正常demo中止／返回各1，閃現0；4700格另存B及3250格冷讀B，全部SAVE雜湊保持 |
| 負對照 | 缺原版、快取改檔及原版INTRO.BMA改檔明確拒絕；還原後完整檔案相同 |
| 乾淨重建 | 7400格CPU／RAM、原版索引／色盤與127檔同既有基準，中央960×600同舊截圖，閃現0 |
| Windows | 實際ZIP以Wine驗1250格、三張截圖、首次匯入與重啟，快取及封包不變 |
| 限制 | Windows原生／macOS真機未驗，macOS未簽章。日韓、配樂試聽、FINAL HD及推廣片仍待完成，#12保留 |

重建入口[hd-patch.sh](../../tools/hd-patch.sh)固定以上兩側來源；[Linux腳本](../../tools/hd_patch_smoke_linux.sh)、[Wine腳本](../../tools/hd_patch_smoke_windows.py)與[獨立驗收器](../../tools/hd_patch_verify.py)處理實際封包，使用及工具依賴見[SETUP](../../SETUP.md#hd-patch候選)。
