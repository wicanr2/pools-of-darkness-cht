# 039 v1.0.0正式版重建與推廣影片

Issue：[#80](https://github.com/wicanr2/pools-of-darkness-cht/issues/80)。使用者2026-11-02指定正式進入v1.0.0、重新打包、重製推廣影片及重新命名安裝包。依既定規則，完整版號為`v.1.0.0-20261102`，日期採Asia/Taipei。

| 項目 | 目前狀態 |
|---|---|
| 基線 | main 6bed8400、引擎c118740；[038](038-input-and-quiz.md)的輸入、五語、2034主題與18字型範圍保持。 |
| 正式版標示 | 更新十二個封包名稱、程式內部版本、manifest的candidate=false及隨包說明。來源與存檔位置沿既有契約。 |
| 推廣影片 | 沿已驗63秒1080p分鏡，以正式版程式重擷取開場、音訊及正常讀檔／面板，更新開頭、結尾、版號與檔名。戰鬥與片尾可沿已驗來源，逐項記錄SHA與直接入口界線。 |
| 驗收 | 進行中。新版本不覆寫舊36包與影片。 |
| 發行方式 | 維持私人交付。驗收後建立同版號私人Git tag，不上傳原版、完整包或影片，不建立公開Release。 |

重建入口沿[`release_prepare.py`](../../tools/release_prepare.py)、[`package_build.sh`](../../tools/package_build.sh)、[`package_payload.py`](../../tools/package_payload.py)、[`package_appimage.py`](../../tools/package_appimage.py)、[`release_verify.py`](../../tools/release_verify.py)與Linux／Wine實際封包抽驗。影片使用[`promo_capture_build.py`](../../tools/promo_capture_build.py)、[`promo_capture.sh`](../../tools/promo_capture.sh)、[`promo_player_capture.sh`](../../tools/promo_player_capture.sh)、[`promo_audio_metrics.py`](../../tools/promo_audio_metrics.py)、[`promo_assets.py`](../../tools/promo_assets.py)、[`promo_render.sh`](../../tools/promo_render.sh)與[`promo_verify.py`](../../tools/promo_verify.py)。

[`v1_deliver.py`](../../tools/v1_deliver.py)核對全部實際封包與影片收據、舊36包、原版125檔及舊影片，通過後才建立正式版交付目錄。

原生Windows及macOS真機未驗、新編配樂未經試聽、日韓未經母語者審校等限制維持。正式版號依使用者定案，不把未走劇情、完整通關或全部原版行為列為本輪已驗事實。
