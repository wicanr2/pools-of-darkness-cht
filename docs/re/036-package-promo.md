# 036 五語HD封包與推廣影片

Issue：[78](https://github.com/wicanr2/pools-of-darkness-cht/issues/78)。使用者指定不要求完整通關，已完成HD先打包並製作推廣影片。候選版號為`v.0.1.2-20261102`，全部交付留本機，不建立tag或公開Release。

## 十二個實際封包

乾淨主專案來源為`28078cede0c0ab69cc4220591c6ec9d871ee6bd5`，引擎為`3fdeeb8ddd93ca84ee3626bc357dc2586c814883`。三平台沿既有Go1.24.13與macOS SDK15.5工具鏈建置；主題2034檔、18份字型全部由乾淨Git及自備原版重建，逐檔SHA同035受測版。

| 交付 | Linux | Windows | macOS | 權利 |
|---|---|---|---|---|
| 基本patch | x86_64 AppImage | x86_64 ZIP | universal ZIP | 自備原版 |
| HD patch | x86_64 AppImage | x86_64 ZIP | universal ZIP | 自備原版，首次重生1091份辨識資料 |
| 基本完整版 | x86_64 AppImage | x86_64 ZIP | universal ZIP | 僅本機，含47個原版遊戲檔 |
| HD完整版 | x86_64 AppImage | x86_64 ZIP | universal ZIP | 僅本機，含47個原版遊戲檔 |

全部包包含繁中、英文、日文、韓文與簡中，及18份原生字型。HD包包含97種片頭與43種片尾相位、B版海報及既有已驗HD。日韓為機器翻譯並經語境校讀，未經母語者審校；新編配樂試聽仍未完成。

[036固定收據](036-package-promo-runtime.json)逐包保存大小、SHA、內含manifest及程式SHA。獨立解開全部十二包，核對完整內容、ELF／PE／兩架構Mach-O、macOS系統依賴、版本與引擎識別、十八份字型、五語、授權與巢狀來源。六個patch無原版檔案、擷取索引或有序表；原版資料只在六個本機完整版，歷史SAVE不隨包附帶。

實際Linux HD patch首次／重啟各1250格，2034檔快取內容與mtime保持，十張三倍／二倍、五語及F12畫面逐像素核對；快取與FINAL.BMA改檔明確拒絕。基本patch以日文1250格驗證，兩種本機完整版各驗韓文啟動，缺原版拒絕。3800格DEMO中止及返回各1次，閃現0；4700格正常另存B與3250格冷讀B通過。Windows實際HD ZIP以Wine驗日文1250格、三張擷取與韓文重啟，2034檔快取與封包保持。原生Windows與macOS真機未驗，macOS未簽章。

測試種子只保留與033雜湊相同的五個正常自建槽A檔案。本次自備原版目錄另有78份歷史SAVE，首次匯入會保留；033當時僅提供47個遊戲檔，不能直接當作相同輸入。另以同一舊AppImage、相同原版目錄與種子重跑後，新舊包的存檔及冷讀CPU／RAM、索引／色盤完全一致，84份SAVE的原始路徑與SHA也完全相同。歷史A/B十份資料以逐檔記錄的DOS大小寫對應核對SHA，全部相同。歷史終點與本次舊包重跑的差異另存，不歸因於新版，也不推測未驗證的內部原因。封包本身仍只附47個原版遊戲檔，沒有歷史SAVE。

## 六十三秒推廣影片

[規格016](../spec/016-promo-video.md)及[037固定收據](037-promo-runtime.json)保存分鏡、來源、擷取、音訊與成品檢查。影片為1920×1080、25fps、H.264／AAC，共1575格、63秒，結尾完整。

影片依序展示題名、重返菲蘭、實際冒險、地圖／攻略、F12與F2操作、五語、戰鬥、片尾及交付提示。題名與字幕以源泉字型排在遊戲畫面外，648個文字矩形通過邊界檢查。影片使用原版實際前端畫面，素材取自本次擷取及034／035已驗收據；地圖只裁出中央畫面，沒有合成假UI。片尾明列原版播放器直接入口，不暗示已完整通關。

原版開場透過既有Nuked OPL混音器擷取85秒48kHz立體聲PCM及6660筆OPL事件。捕捉副本只加入樣本／事件寫檔，沒有替換音源、改啟動佇列或修改客體狀態；5100格終點完全同031原版基準。影片使用其中63秒、音量乘2並淡入2秒／淡出3秒，沒有自寫近似合成器或新編配樂。

FFprobe確認視訊與音訊都是63秒。平均音量−20.7dB、峰值−5.1dB，未檢出超過0.4秒的非預期靜音，原始PCM無削波。未檢出持續0.2秒的黑場；七段凍結均位於分鏡明列靜態卡片，F2／F12操作片段未檢出兩秒凍結。九張成品擷取已目視，字幕與結尾完整。音訊完成技術檢查，沒有另外宣稱人耳試聽通過。

## 交付與重現入口

交付根為`dist-all/v.0.1.2-20261102/`，`patch/`六包、`full-local/`六包、`promo/`影片與媒體驗收、`smoke/`實際封包收據。`SHA256SUMS.json`保存十二包與影片的大小及SHA。原版125個輸入與舊十二包保持。

- [乾淨來源準備](../../tools/release_prepare.py)、[平台建置](../../tools/package_build.sh)、[payload](../../tools/package_payload.py)、[AppImage](../../tools/package_appimage.py)。輸入固定main／engine，輸出先留workplace，使用既有Docker映像。
- [Linux實際封包抽樣](../../tools/release_smoke_linux.py)、[Wine抽樣](../../tools/release_smoke_windows.py)、[獨立封包驗收](../../tools/release_verify.py)。正常自建槽A的五個輸入SHA保存在036收據，不能替換成含歷史SAVE的目錄。
- [音源擷取副本](../../tools/promo_capture_build.py)、[正常開場擷取](../../tools/promo_capture.sh)、[字卡與分鏡](../../tools/promo_assets.py)、[FFmpeg合成](../../tools/promo_render.sh)、[影片驗收](../../tools/promo_verify.py)。實際媒體不入Git，原始畫面、PCM與中間檔留本機`workplace/package-promo-v1`。
- [交付工具](../../tools/release_deliver.py)只在兩份驗收通過後建立新版本，核對所有SHA與擁有權。完整影片使用原版畫面與音源，僅本機私人展示，公開散布權未取得；沒有公開上傳。
