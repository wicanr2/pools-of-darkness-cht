# 031 片頭BMA來源與繪製中途狀態

日期：2026-11-01。Issue：[74](https://github.com/wicanr2/pools-of-darkness-cht/issues/74)，母項[#10](https://github.com/wicanr2/pools-of-darkness-cht/issues/10)。來源階段已完成，正式動畫HD仍DRAFT，#74保持開啟。沿[018的正常來源](018-movie-subtitles.md#片頭美術區盤點入口)，本輪不改原版、正式程式、主題或舊九包。

研究工具：[唯讀BMA解碼器](../../tools/intro_bma.py)、[格式邊界與真實輸入檢查](../../tools/intro_bma_test.py)、[正常來源與Go快路徑探針](../../tools/movie_bma_probe.go)、[來源核對及美術請求](../../tools/intro_bma_inventory.py)。這些工具只供來源診斷，不能當作正式播放器。輸入唯讀掛載，產物在`workplace/probe/intro-hd-v1/`。

容器編排入口：[intro-bma-receipt.sh](../../tools/intro-bma-receipt.sh)。需要合法原版、既有G2設定與驅動、固定Go工具映像及模組快取；沒有輸入時明確失敗，不改用狀態注入。

固定證據：[來源及兩路正常重播](031-intro-bma-source.json)、[98種完整畫面請求](031-intro-bma-requests.json)。請求內只保存檔名、位移、雜湊及提示詞，原版索引、色盤與參考PNG留在本機。六張既有HD只沿用已接受的PNG，沒有新HD成品或新主題格式。

## 輸入與工具

| 輸入 | SHA-256 |
|---|---|
| `INTRO.BMA`，648158 bytes | `12f3944652b2ae77feb3c0b16435cb2a3d855c26c388e6d6f97c277abedcb94f` |
| `INTRO.ACF` | `f005e97b9d704bdfee940415d29ed76ce130a4151ffb420abcf38aeb3b6b417f` |
| `VACE.EXE` | `30fd9c2baf0d38ce9a019fefd410ce5216983763e4a1cb30bd268dd08f6ad683` |
| `VGA.DRV` | `eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f` |

Go 1.24.13、`dosgolem-go-ebiten:1.24-r1`（`8ed6bdee2bce`），引擎`0e804eadf3f7c9d262938d57885f2e2ac0c6e595`。影像工具為`sosb-img:r1`（`8b9012adb991`）、Python 3.13.16、Pillow 11.0.0。研究探針SHA為`1a9897a5c827f4ac2b2bb975137b517f6a79a888c2bbf1053c1de5e5144868f7`。
原版125檔集合與G2的`pool4.cfg`、`snd.com`逐檔同018。兩側均沿SND／CONTROL、不送鍵、50%時脈及每回合96250步，沒有快照、座標注入或新亂數規則。

## 已證實的BMA格式

位移基準是`INTRO.BMA`原始檔案位移，整數採小端序。14個區塊連續解析至648158，沒有尾端餘資料。各塊完整SHA、原始位移與圖數保存在固定來源收據。

- 區塊先有包含自身的32位元長度、16位元圖數，再接每圖一個32位元相對位移。第一圖從`6+4×圖數`開始，目錄嚴格遞增。本作共133圖。
- 圖面先有16位元寬、高，本作全部320×200。接著逐列讀8位元y；255結束圖面。每段先讀16位元x及兩個8位元長度，x高位表示本列末段，其餘15位為x。兩個長度分別是展開像素數與壓縮bytes。
- RLE控制byte右移一位再加一為長度；最低位1表示重複下一色號，0表示讀取該數量的連續色號。空更新仍有寬高與255終止符，不能清空上一張圖。
- 本次正常INTRO依檔案順序各呼叫一圖，包含空更新，共133次。依序套用跨度後，上方168列共有98種完整索引。這個順序有正常呼叫證據，沒有據此建立通用ACF播放器。

解碼契約沿[既有原版RLE與區塊證據](019-draw-interception.md)，本輪再以原始BMA及正常畫格逐點核對。格式範圍、壓縮長度、展開長度、列末與圖末都須完整，未知或損壞資料直接拒絕。

## 正常入口與分段畫面

圖面入口為執行期`48A1:6733`；同一段255 bytes在`VGA.DRV`檔案位移`0x6733`，SHA為`d9ebc163f106be514229472090f2497771df8e36ab5aeca61681f758d32baa54`。兩種位址基準分開記錄，沒有覆寫原始名稱或operand。

在入口以`SS×16+SP`定位堆疊：+4表面、+6／+8座標、+10方向旗標、+12／+14來源遠指標、+16圖格。來源線性RAM位址另外保存；以完整BMA區塊SHA回查檔案位移，不能把RAM值當檔案位移。本次座標與方向均0，未驗非零方向。

畫面有相同邊框時不能只靠圖面SHA挑場景。例：1623回合的邊框前綴同時符合數張完整場景，實際呼叫已明確指向區塊4、圖0。133次呼叫的完整來源、圖格、時點及次序，在原版與Go路徑完全相同。

| 172筆正常上方區域轉換 | 筆數 | 證據 |
|---|---:|---|
| 完整BMA圖面 | 136 | 逐點等於該次呼叫的累積圖面 |
| 繪製前綴 | 28 | 逐點符合該圖的有序像素寫入前綴，其餘像素完整符合前一圖或已觀測清空底圖 |
| 清除前綴 | 1 | 3767回合的前35200點為0，其餘逐點保留前一完整圖 |
| 單色狀態 | 6 | 全53760點相同，色號另記 |
| BMA範圍外 | 1 | 018已列出的後續TITLE，沒有包裝成BMA畫面 |

以上是172次轉換、163種索引與使用RGB狀態、129種索引；98是完整BMA索引的數量，分母不同。索引與使用RGB分別核對，不能把全黑淡出當成圖面已清除。色盤變體全部保存在來源／請求索引，預覽選圖不等於正式淡出或調色契約。

## Go入口與驗證結果

- 原版及Go各5100格、490875000步，CPU／RAM、原始索引／色盤、172筆轉換與127個客體檔案同018。163對原始來源完全相同。
- Go路徑實際提交223064個計畫、7904409步，提案與提交數相同；133次入口均經過Go區塊入口。原版路徑兩項提交數為0，不能只看`go-draw`旗標判定已驗。
- 本次133次來源先由`BeforeStep`記錄，Go區塊入口也呼叫同一觀測並按步數去重。現行`Live.BeforeStep`在GAME尚未ready時會提早返回；未來產品掛勾須涵蓋VACE時期，不能把觀測放到該返回點之後。這是接入限制，本輪未修改Live。
- 六組格式／真實輸入測試通過，含22個損壞格式子案例及前綴外改點拒絕。六種來源負對照涵蓋未知區塊、圖格、次序、方向、索引與使用RGB。
- 最初研究探針漏開`NativeHostBlocks`，實際提交量0被檢查拒絕；補上session開關及提交量後，以同一新程式重跑兩側。舊輸出留在`final-*`，正式證據只取`checked-raw`、`checked-go`。這是驗證腳本缺項。

本輪是正常原版來源與Go繪圖相容性研究。未接入新HD，不把無頭來源收據說成新的正式GUI或HD閃現驗收；既有正式程式仍`7109e14`。

## 素材需求與重建

| 已接受素材 | BMA區塊／圖格 |
|---|---|
| `intro-council` | 1／13 |
| `intro-phlan` | 2／0 |
| `intro-horned` | 4／0 |
| `intro-heroes` | 6／0 |
| `intro-followers` | 7／0 |
| `intro-dreadlord` | 9／0 |

98種完整索引中，以上六種已有接受的960×504成品；其餘92種仍需相位素材或經證實的圖層合成。92不等於必須再生92張整幅畫，後續可沿原始相位證明素材共用。不能只以場景相同就沿用一張靜態圖。

另外由乾淨引擎Git匯出重建同一研究探針，SHA逐位元同受測檔；Go overlay能提供不存在於Git中的虛擬入口，不需要舊`workplace/movieprobe/`目錄。

本機請求在`workplace/probe/intro-hd-v1/requests-v1/`；由乾淨Git樹`249cb4ef56806605878844b28a5f19cbf97157f9`取回工具及已接受PNG，再用明示原版與兩路來源重生的`requests-clean/`，全部393檔逐位元相同。請求manifest SHA為`945c63d7a7393568f742af2b67f067b191d2e6062514bb6a8bcd3b1df15bc107`。393檔包括98組索引、色盤、1倍及3倍原版參考，以及manifest；參考圖不稱為HD成品。

重播入口：`bash tools/intro-bma-receipt.sh <新名稱> [合法G2種子目錄]`。預設種子在`workplace/probe/g2-receipt/scratch/`，產生方式沿[既有G2啟動證據](003-dosgolem-startup.md)及`tools/g2-startup-receipt.sh`；雜湊不同時拒絕。編排入口完成語法檢查，各容器步驟已分別實際執行；本輪未另跑整支編排第二次。
直接核對可在影像映像內執行`python3 -I tools/intro_bma_test.py -v`；來源盤點的全部參數見`intro_bma_inventory.py --help`。缺原版時格式測試明確skip，不能算來源通過。

較早018、規格010與CONTEXT的BMA未知已有回填，固定來源收據內的resolution紀錄保留不可變輸入鍵及必要勘誤標記。盤點工具會拒絕缺少標記或回鏈的舊入口。後續仍需HD相位素材、正式呈現及GUI／閃現抽驗；FINAL、通用ACF語意及其他原案功能未完成。


## 片頭HD接入驗收

2026-11-01追加，Issue [#74](https://github.com/wicanr2/pools-of-darkness-cht/issues/74)。前文保留來源階段紀錄；本節接續[規格010](../spec/010-hd-theme.md#bma片頭動畫接入)的READY審查與正式接入。引擎`50c93fb2fd1260572e031787e61e69659cec1d62`，程式SHA-256 `82b9c9ee2356cee690f747895a58e60ed485d1d02bae400d950e608903a83ba0`。

| 項目 | 結果與證據等級 |
|---|---|
| 原版來源 | 沿前述固定BMA／ACF／VACE／VGA雜湊與dosgolem位址空間，133次來源呼叫全部綁定，已證實 |
| 顯示 | 97種敘事相位，960×504不透明；六張b44主畫面RGBA逐點不變。新增七份生成背景／物件／動作圖；[逐相位成品與SHA](../../review/hd/b46-intro/manifest.json)、[數值驗收](../../review/hd/b46-intro/validation.json)與[提示詞](../../review/hd/b46-intro/generation.json)另列 |
| 相位／布幕 | 原版BMA呼叫及完整前後索引決定圖面；有序寫入表處理分段來源。BMA11用原版可見人物色段與布幕遮罩，不對人物拉伸；素材遮擋屬美術近似，來源與播放順序已證實 |
| 明暗 | 僅接受正常原版色盤SHA，亮度由當前與完整來源色盤導出；HD色調屬美術近似。未深入硬體driver，客體色盤與時序不改 |
| 真實來源守衛 | 131個完整、26個繪製前綴與1個清除前綴通過。未知像素／色盤、亂序寫入、來源區塊／程式碼失配、非法路徑／清單及2倍負對照通過；下方字幕不覆蓋 |
| 正常入口 | SND／CONTROL冷開機，不送DOS鍵；原版舊程式、HD候選與F2／F8／F12各5100回合、490875000步，原版終點及127個DOS檔案相同 |
| 畫面 | 33張正常擷取，其中31張完整HD與PNG／亮度逐點相同；全黑清除與TITLE轉場另驗。二倍美術、F12整幅與原版RGB逐點相同，F8不改美術 |
| 量尺 | 原始HD40／切換599；40格發生於3767→3769，來源及實際合成RGB逐格全0。其餘559格只在706→711明示F2／F12區間，二倍與原版截圖已驗。原始記錄完整保存，非預期可見閃現0；沒有調量尺、門檻或黑色像素 |
| 保存與重建 | PNG／提示詞／清單進私人Git。850份索引與133份有序寫入表由自備原版重生；乾淨Git樹3684bb4依SETUP跑`tools/rebuild.sh assets`，1866檔主題與十份字型逐檔相同。引擎乾淨Git重編程式SHA相同 |
| 現行與回復 | 現行已換成受測程式與主題；舊程式`pod-play-before-intro-hd`、舊主題`workplace/hd/theme-before-intro-hd/`保留。舊九包逐包SHA同026，不覆寫 |

固定[正常收據](031-intro-hd-runtime.json)保存來源、程式與主題SHA、全部量尺格子分類及乾淨重建結果。素材[審閱頁](../../review/hd/b46-intro/README.md)有逐相位拼圖；README另保存實際正常視窗。

### 重現

- 影像前處理：[intro_hd_art.py](../../tools/intro_hd_art.py)讀既有請求與生成原圖，`--root <repo> --out <新相位目錄>`。工具鏈[pod-intro-img:r2](../../tools/docker/intro-img.Dockerfile)固定Python3.13.16、Pillow11.0.0、NumPy2.1.3及OpenCV4.12.0.88；來源引導變形只定位手繪材質，不將原版RGB濾鏡放大作主畫面。
- 數值：[intro_hd_validate.py](../../tools/intro_hd_validate.py)指定`--root --artwork --report`；尺寸、alpha、結構0.4、色相30度門檻不變。
- 來源主題：[intro_hd_theme.py](../../tools/intro_hd_theme.py)指定`--root --original --artwork --capture --out`。原始索引、色盤及有序位置只留本機。
- 正常視窗：[intro-hd-play-receipt.sh](../../tools/intro-hd-play-receipt.sh)接受`hd`、`base`或`switch`及新收據名。現行程式／主題為預設，可用`POD_INTRO_BIN`／`POD_INTRO_THEME`明示候選；全部工作在Docker內。
- 比對：[intro_hd_play_check.py](../../tools/intro_hd_play_check.py)指定`--base --hd --switch --theme --proof`。本次基準在`workplace/probe/intro-hd-final-v1/base`，最終HD／切換在`workplace/probe/intro-hd-final-v2/`。保留原始量尺，只接受已固定且逐格證明的全黑或明示切換，其他記錄失敗。
- 跨機：[SETUP](../../SETUP.md)的`tools/rebuild.sh assets`從Git PNG、JSON、字型與自備DAX／INTRO.BMA／INTRO.ACF重建，不需要生圖服務或舊workplace。

SSI標誌全程保留原版，FINAL另計。其他速度、未觀測色盤與其他原版版本維持未知並安全回退。此結論不外推全遊戲HD或零閃現；#10與#16母項保持開啟。
