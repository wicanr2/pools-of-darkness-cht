# 012 三平台候選封包

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
