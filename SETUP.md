# 在另一台機器重建

本入口重建目前已驗收的繁中、HD與Linux程式。需要Git、Docker，以及使用者合法持有的DOS原版資料目錄。
GitHub儲存庫維持私人。原版EXE、DAX、存檔與掃描不在Git內；重建不需要舊機器的`workplace/`、前作目錄、聊天紀錄或憑證。

## 取得與重建

```sh
git clone https://github.com/wicanr2/pools-of-darkness-cht.git
cd pools-of-darkness-cht
bash tools/rebuild.sh assets /absolute/path/to/original
bash tools/rebuild.sh build /absolute/path/to/original
bash tools/rebuild.sh smoke /absolute/path/to/original
```

首次需要網路取得Docker基底映像、系統套件、固定引擎commit與Go模組。素材解碼、字型轉換、編譯與遊戲測試都在Docker內執行。
腳本沿用已有的`sosb-img:r1`及`dosgolem-go-ebiten:1.24-r1`；找不到時，分別由[影像Dockerfile](tools/docker/img.Dockerfile)及[Go Dockerfile](tools/docker/rebuild-go.Dockerfile)建置。
原版唯讀掛載，重建輸出預設在`workplace/rebuild/current/`，已有輸出時拒絕覆寫。可用第三個參數指定另一個`workplace/`子目錄。

| 步驟 | 輸出與檢查 |
|---|---|
| `assets` | 重生751份原版辨識索引，重建1536檔主題及十份字型，每檔核對已驗SHA-256 |
| `build` | 從本儲存庫`dosgolem/pod`取得[固定commit](tools/rebuild-engine.commit)，編譯`bin/pod-play`，保存建置資訊與SHA-256 |
| `smoke` | 正常冷啟動7400格DEMO，產生`smoke/`的截圖、來源、紀錄與獨立可寫客體資料 |

`smoke`還會自動比較CPU、記憶體、原版索引、色盤、最終畫面及127個客體檔案，要求閃現0。這是嚴格重播基準，原版目錄的既有檔案也必須相符；不同版本或不同資料集合會明確失敗，不會宣稱對拍通過。

主題在`theme/`，字型在`font/`，素材重建結果在`assets-receipt.json`。重建不改正在使用的正式主題。
新程式以`-trimpath -buildvcs=false`建置，雜湊可能與歷史開發執行檔不同；玩家可見行為另以正常路徑核對，不以歷史執行檔雜湊相等作假設。
既有固定程式重建只驗證Linux容器。另有下方三平台候選，Windows與macOS正式驗收仍依[#12](https://github.com/wicanr2/pools-of-darkness-cht/issues/12)追蹤。
本輪沿用既有專用工具映像驗證乾淨來源，未另建映像；兩份Dockerfile的固定官方基底摘要已查證可取得。

## 素材與權利

| Git中的輸入 | 用途 |
|---|---|
| [主題重建清單](review/hd/runtime/manifest.json) | 固定全部檔案、來源、SHA-256及原版DAX版本 |
| [HD審閱素材](review/hd/README.md) | 沿用已接受PNG；缺少的開場／海報圖補存於[runtime](review/hd/runtime/README.md) |
| [字型來源與字集](font/README.md) | 開源字型及授權，逐位元組重生目前6份正式字型 |
| `text/`及`reference/` | 正式譯文、手札、攻略摘要與術語 |
| 本儲存庫`dosgolem/pod`分支 | 遊戲adapter與引擎來源，不依賴前作工作樹 |

海報含原版封面插畫，沿既有私人審閱例外保存，不因此取得公開散布權。生成素材不必再次呼叫生圖服務。
重建清單只保存原版資料的雜湊及選取方式，原始辨識索引在本機由DAX重生。原版版本不合、圖片遭改或字型來源不同時，重建會失敗並指出缺項。

目前另含簡中介面、手札、攻略、地圖、姓名音譯與開場文字，F8依繁中、英文、簡中循環。簡中有原生16／24及工具列12／22字模。正式TSV由固定OpenCC1.1.9及詞彙表產生；修改繁中後，可在`dosgolem-go-ebiten:1.24-r1`容器內執行`python3 -I tools/zhcn_text.py --root <repo> --out <repo>/text`重生，或加`--check`檢查。來源與重建範圍見[規格015](docs/spec/015-multilingual.md#簡體中文收尾)，OpenCC不在遊戲執行時轉換。

## 驗證與目前限制

跨機重建的驗證結果由[CONTEXT.md](CONTEXT.md)索引，過程記在[WORKLOG.md](WORKLOG.md)。
初次[跨機重建收據](docs/re/025-rebuild-runtime.json)保存新編譯程式的7400格正常路徑；目前[遭遇圖收據](docs/re/017-b43-runtime.json)驗證1525檔主題與6份字型由乾淨Git檔案重生，固定引擎56ebba3乾淨重編與受測程式逐位元相同。[正常增量比較](docs/re/017-b43-play.json)的CPU、記憶體、來源、127檔及最終畫面同前批、閃現0，因此重建終點畫面基準仍綁[b38正常收據](docs/re/017-b38-play.json)。既有041近遠景由前批配對截圖驗證；本批十四張既有截圖逐張相同，新正常事件仍未知。
簡中增量的[固定收據](docs/re/027-zhcn-runtime.json)另外保存正常讀檔／面板路線、既有字幕來源與1536檔／十份字型的乾淨重建。歷史7400格原版狀態與畫面預期未改，新增19份簡中TSV另列於清單，舊文字SHA保持原樣；新語系的抽樣範圍不能外推為全片頭或全部玩家路線。
這個入口保存目前已完成的素材，不代表尚未驗證的地形、動畫、日韓語系、英文面板、配樂或平台已完成。
新增HD素材或正式字型時，須同步更新重建清單，並以乾淨Git檔案重跑`assets`，核對新正式主題及字型雜湊後才提交。

## 候選封包

`v.0.1.0-20261101`是本機可試玩候選，沒有tag或公開Release。九包在`dist-all/v.0.1.0-20261101/`，總大小約314MB；`SHA256SUMS.json`記錄逐包雜湊，`smoke/`保存實際封包驗收。含原版的六包只留`full-local/`，不可上傳。`patch/`的三包為基本中文化，需自備原版；HD patch尚未完成。

| 平台 | 候選抽驗 |
|---|---|
| Linux x86_64 AppImage | 基本／HD首次匯入、正常demo中止返回、patch沿用匯入、缺原版拒絕、另存B與冷啟動讀B |
| Windows x86_64 ZIP | Wine內實際解包、首次匯入、版本與1250格正常開場；原生Windows未驗 |
| macOS universal ZIP | amd64與arm64、App結構、執行權限及系統依賴；未簽章、未公證，真機未驗 |

各包`README.txt`說明操作與存檔位置。Linux需要X11、OpenGL、ALSA與glibc2.36或相容環境；缺FUSE時用`APPIMAGE_EXTRACT_AND_RUN=1`。首次把合法DOS英文版1.10的47個遊戲檔放到程式旁`original/`，或用`-original`指定。歷史玩家SAVE不在包內；新存檔寫入使用者資料目錄。

重建入口為[tools/package.sh](tools/package.sh)，固定候選引擎`4a8ef9f4ff18c82d8c98595faf19fb1ee803c609`及主專案素材來源`4bdd711`，保留這九包的原始範圍。它從乾淨Git匯出素材與程式，只改啟動器的原版版本提示；後續簡中資料不會改寫這組既有候選。輸出目錄必須是本專案`workplace/`內的新路徑。

```sh
bash tools/package.sh build /absolute/path/to/original "$PWD/workplace/package-rebuild"
bash tools/package.sh smoke /absolute/path/to/original "$PWD/workplace/package-rebuild" /absolute/path/to/normal-save-A
bash tools/package.sh verify /absolute/path/to/original "$PWD/workplace/package-rebuild"
```

`build`只對下載Go模組開網路。`smoke`第四參數是經正常遊戲建立、只有槽A而沒有槽B的SAVE目錄，供測試另存B，不注入執行期狀態；開場與缺檔抽驗仍由正式前端啟動。個別步驟由[規格012](docs/spec/012-package.md)連結的容器腳本執行。本輪實際驗證的是各步驟，編排腳本完成語法檢查，尚未再整輪重跑。

候選工具映像須已存在：`sosb-img:r1`、`dosgolem-go-ebiten:1.24-r1`、`cok-osxcross:go1.24.13-sdk15.5-r1`、`kgb-cht-pkg:1.99.1-r2`，Wine抽驗另需`yuan-wine-verify:noble-r2`。前兩者來源見上方Dockerfile。後三者沿用主機既有工具鏈，macOS SDK不放Git；新機器需另備合法SDK及相同工具能力，完整三平台工具映像重建尚非已驗範圍。固定AppImage runtime及對應來源已[存檔](tools/vendor/appimage/manifest.json)，不需浮動下載。

[固定驗收收據](docs/re/026-package-runtime.json)保存九包大小與SHA、原版同狀態、Go繪圖計數、demo、存讀檔及平台限制。基本版量尺原始23格保留為題名與紅色背景轉換，HD原始0；此分類不代替其他畫面抽驗。配樂、多語系、無敵、完整動畫HD與HD patch仍未完成，不能把候選稱為原案完成版。

維護重建清單時，先將已驗收PNG加入Git暫存，使用`git ls-files -z`保存檔案清單。
在Docker內執行`tools/rebuild_theme_snapshot.py --root <repo> --theme <已驗候選> --tracked-list <Git檔案清單> --batch bNN`，
更新原樣主題JSON與來源SHA。工具拒絕未存Git的PNG及被改動的既有索引；不會複製原版索引進Git。
再以乾淨Git檔案執行`assets`，逐檔核對候選，正常路徑通過後才提交。原版版本、引擎或正常終點改變時須先重驗基準，不能只更新雜湊。
