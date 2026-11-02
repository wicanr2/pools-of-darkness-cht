# 039 v1.0.0正式版重建與推廣影片

Issue：[#80](https://github.com/wicanr2/pools-of-darkness-cht/issues/80)。使用者2026-11-02指定正式進入v1.0.0、重新打包、重製推廣影片及重新命名安裝包。依既定規則，完整版號為`v.1.0.0-20261102`，日期採Asia/Taipei。

| 項目 | 目前狀態 |
|---|---|
| 基線 | main 6bed8400、引擎c118740；[038](038-input-and-quiz.md)的輸入、五語、2034主題與18字型範圍保持。 |
| 正式版標示 | 更新十二個封包名稱、程式內部版本、manifest的candidate=false及隨包說明。來源與存檔位置沿既有契約。 |
| 推廣影片 | 沿已驗63秒1080p分鏡，以正式版程式重擷取開場、音訊及正常讀檔／面板，更新開頭、結尾、版號與檔名。戰鬥與片尾可沿已驗來源，逐項記錄SHA與直接入口界線。 |
| 驗收 | 十二包與63秒影片全部通過本輪驗收，交付於dist-all/v.1.0.0-20261102。舊36包及原影片保持SHA。 |
| 發行方式 | 維持私人交付，正式Git標籤名稱為v.1.0.0-20261102。不上傳原版、完整包或影片，不建立公開Release。 |

重建入口沿[`release_prepare.py`](../../tools/release_prepare.py)、[`package_build.sh`](../../tools/package_build.sh)、[`package_payload.py`](../../tools/package_payload.py)、[`package_appimage.py`](../../tools/package_appimage.py)、[`release_verify.py`](../../tools/release_verify.py)與Linux／Wine實際封包抽驗。影片使用[`promo_capture_build.py`](../../tools/promo_capture_build.py)、[`promo_capture.sh`](../../tools/promo_capture.sh)、[`promo_player_capture.sh`](../../tools/promo_player_capture.sh)、[`promo_audio_metrics.py`](../../tools/promo_audio_metrics.py)、[`promo_assets.py`](../../tools/promo_assets.py)、[`promo_render.sh`](../../tools/promo_render.sh)與[`promo_verify.py`](../../tools/promo_verify.py)。

[`v1_deliver.py`](../../tools/v1_deliver.py)核對全部實際封包與影片收據、舊36包、原版125檔及舊影片，通過後才建立正式版交付目錄。

## 交付與驗收結果

[封包固定收據](039-v1-package-runtime.json)保存十二包的大小、SHA、逐檔manifest、內部程式及平台識別。Linux／Windows／macOS各有基本與HD的patch及本機完整版。正式版名稱、程式版本、macOS metadata及manifest一致，candidate=false；隨包說明已移除候選／試玩標示。三平台來源為main `9afeac8474a8c79b9ed741bf2fe44488263f9ed0`與引擎`c118740dc3cbee42881abf45a609f63375c474e7`。

Linux實際AppImage驗首次／重啟、五語、F12、兩倍率、基本／完整、缺原版拒絕及存讀檔；4700格另存B與3250格冷讀B的CPU、RAM、原版畫面、色盤與全部SAVE同v.0.1.3。Windows實際ZIP以Wine驗日文1250格與韓文重啟，快取及封包保持；macOS只驗兩架構及系統依賴。2034檔主題與18份字型由乾淨Git重建，SHA保持。

[影片固定收據](039-v1-promo-runtime.json)保存新63秒、1920×1080、25fps、H.264／AAC成片，1575格及完整結尾。正式版重新擷取5100格開場與4800格正常讀檔／攻略／地圖，原版開場CPU、RAM、畫面及色盤同031基準，85秒48kHz立體聲PCM與原音源SHA相同，含6660筆OPL事件。其他戰鬥及片尾取自已驗來源，片尾仍明列播放器直接入口。

九張編碼成片樣本與完整結尾已目視，649個文字矩形通過安全區。音訊平均−20.7dB、峰值−5.1dB，原始PCM無削波；未檢出非預期靜音、黑場或操作段兩秒凍結，七段靜態卡片另列為預期。音訊為技術檢查，未另作人耳試聽。

十二包共1,186,619,064 bytes，影片6,812,591 bytes，合計1,193,431,655 bytes。影片SHA為`1afd4d3ed6366b73e827d39ca506be408d2a975dc4cd1696a5c26791c1f47959`。完整逐項清單在同版號SHA256SUMS.json，所有交付只留本機。正式Linux程式SHA為`303de6552d2c31af4714c625603f579fcac7f43872ab3438a207296d46064201`，已更新開發入口並回讀v.1.0.0版號。

原生Windows及macOS真機未驗、新編配樂未經試聽、日韓未經母語者審校等限制維持。正式版號依使用者定案，不把未走劇情、完整通關或全部原版行為列為本輪已驗事實。
