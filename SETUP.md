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
| `assets` | 重生664份原版辨識索引，重建1357檔主題及6份字型，每檔核對已驗SHA-256 |
| `build` | 從本儲存庫`dosgolem/pod`取得[固定commit](tools/rebuild-engine.commit)，編譯`bin/pod-play`，保存建置資訊與SHA-256 |
| `smoke` | 正常冷啟動7400格DEMO，產生`smoke/`的截圖、來源、紀錄與獨立可寫客體資料 |

`smoke`還會自動比較CPU、記憶體、原版索引、色盤、最終畫面及127個客體檔案，要求閃現0。這是嚴格重播基準，原版目錄的既有檔案也必須相符；不同版本或不同資料集合會明確失敗，不會宣稱對拍通過。

主題在`theme/`，字型在`font/`，素材重建結果在`assets-receipt.json`。重建不改正在使用的正式主題。
新程式以`-trimpath -buildvcs=false`建置，雜湊可能與歷史開發執行檔不同；玩家可見行為另以正常路徑核對，不以歷史執行檔雜湊相等作假設。
目前只驗證Linux容器重建，Windows與macOS正式封包仍依[#12](https://github.com/wicanr2/pools-of-darkness-cht/issues/12)追蹤。
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

## 驗證與目前限制

跨機重建的驗證結果由[CONTEXT.md](CONTEXT.md)索引，過程記在[WORKLOG.md](WORKLOG.md)。
初次[跨機重建收據](docs/re/025-rebuild-runtime.json)保存新編譯程式的7400格正常路徑；目前[遭遇圖收據](docs/re/017-b41-runtime.json)驗證1357檔主題與6份字型由乾淨Git檔案重生，固定引擎672623e乾淨重編與受測程式逐位元相同。[正常增量比較](docs/re/017-b41-play.json)的CPU、記憶體、來源、127檔及最終畫面同前批、閃現0，因此重建終點畫面基準仍綁[b38正常收據](docs/re/017-b38-play.json)。正常近遠景的變化另由本批配對截圖驗證，不以相同終點取代。
這個入口保存目前已完成的素材，不代表尚未驗證的地形、動畫、語系、配樂或平台已完成。
新增HD素材或正式字型時，須同步更新重建清單，並以乾淨Git檔案重跑`assets`，核對新正式主題及字型雜湊後才提交。

維護重建清單時，先將已驗收PNG加入Git暫存，使用`git ls-files -z`保存檔案清單。
在Docker內執行`tools/rebuild_theme_snapshot.py --root <repo> --theme <已驗候選> --tracked-list <Git檔案清單> --batch bNN`，
更新原樣主題JSON與來源SHA。工具拒絕未存Git的PNG及被改動的既有索引；不會複製原版索引進Git。
再以乾淨Git檔案執行`assets`，逐檔核對候選，正常路徑通過後才提交。原版版本、引擎或正常終點改變時須先重驗基準，不能只更新雜湊。
