# 015 左右海報來源與版面示意

日期：2026-10-24。Issue：[#21](https://github.com/wicanr2/pools-of-darkness-cht/issues/21)。狀態：繁中／英文與 2／3 倍已接入並驗收。

使用者指定在遊戲視窗兩側加海報，先查原版封面或海報；文字須隨目前語系切換。
本次命中視覺共同決策與文件職責路由，已讀 grilling 與文件職責入口。先展示可撤回原型，再確認版面。

## 已找到的來源

- [電腦遊戲博物館](https://www.mocagh.org/loadpage.php?getgame=pod)保存原版盒裝正反面。
  正面掃描 402×612，背面 412×632，適合作版本與構圖參考。
- [1991 年廣告掃描](https://videogamecomicads.blogspot.com/2016/09/pools-of-darkness-1991.html)，1210×1600，
  為廣告版面，不能稱作已證實的隨盒海報。
- [Steam 官方商店](https://store.steampowered.com/app/1904522/Pools_of_Darkness/)的背景圖為 1438×810，
  無英文標題與廣告文案，適合裁作兩側背景，另放目前語系的文字。
- [MobyGames 封面頁](https://www.mobygames.com/game/pools-of-darkness/cover-art/gameCoverId%2C659431/)可交叉辨識盒繪，
  本輪直接下載回 403，未假設已取得高清圖。

來源下載及逐檔 SHA-256 保存在 `workplace/hd/source/poster/sources.json`，HTML 來源快照與網址在同目錄 `pages.json`。
可回查的來源清單副本見 [015-poster-sources.json](015-poster-sources.json)。
只下載公開掃描與商店美術，不執行其程式碼；Docker 使用 `sosb-img:r1`、Python 3.13.16、Pillow 11.0.0，下載這批才開網路。
原始掃描維持本機參考，不放進一般發行包。

## 可撤回原型

[審閱入口](../../review/hd/poster/README.md)，由 [poster_prototype.py](../../tools/poster_prototype.py) 產生。
中央為既有截圖，沒有修改 dosgolem；海報採原版插畫裁切，標題與文案以獨立文字顯示。
繁中與英文可切換，示意用字型為既有源泉圓體 Bold。其他語系的正式接入依既定多語系流程另驗。

兩個候選：固定加寬到 1440×600，中央保持 960×600、兩側各 240；或維持 16:9，中央同尺寸、每側只餘 53。
後者不足以放完整海報與文案，示意只保留圖面。待使用者確認顯示方式、裁切與比例後，再寫正式 READY 契約。

使用者已於 2026-10-24 選定固定加寬，寬版繁中原圖與雜湊保存在審閱頁，正式契約見[規格 023](../spec/023-side-posters.md)。
Chromium 初次 CLI 抓圖的視窗高度包含外框，底部裁切，改用 CDP 明示 viewport 後重抓；定版圖為完整版本。
中央 960×600 與輸入截圖逐像素一致，繁中與英文文字切換均已目視。

## 圖層匯出與實作查驗

`tools/poster-build.sh` 沿用 Playwright 1.56.1 的 Chromium 1194，來源唯讀，輸出在本機 HD 主題的 `posters/`。
四張無字底圖與八張語系文字層沿用 PNG；正式文案仍只有 `text/poster.<語系>.tsv`。
`apps/pod/PosterSet` 合成兩側，`cmd/pod-play` 加寬 Layout、修正游標座標；DOS 畫布與截圖契約不變。
另加 `-shot-frame`／`-shot-frame-dir` 產生完整視窗收據，`-posters -` 可停用本機海報。

第一次直接匯出透明文字，Chromium 改用不同字緣。普通不透明原型重抓與定版完全一致，
Go 與 Pillow 都重現透明層的差異，據此確認問題在瀏覽器透明背景渲染，沒有放寬比較閾值。
改成從不透明語系圖與無字底圖取文字差分層；未改的像素全透明，改動像素保持不透明原型的顏色。
兩種倍率、兩個語系、兩側的 Go 合成都與不透明參照逐像素相同，3 倍參照也與使用者定版兩側完全相同。
匯出入口為 `poster_capture.mjs` 與 `poster_layers.py`，各輸入與輸出雜湊保存在 `posters.manifest.json`。

純呈現與輸入測試已通過中央 RGBA 不變、原始輸入未修改、語系返回、未知語系清字、缺檔與錯尺寸拒絕，
以及 2／3 倍左右邊界、海報不觸發中央切換列與 DOS 按下事件。

正常冷啟動 2210 格的不切換、F2／F12 對應動作切換、停用海報三路收據通過。
原版步數、記憶體、色號一致，127 個 DOS 檔案相同；中央及定版兩側像素相同。
5900 格 DEMO 兩張圖與舊兩項 HD 主題的中央截圖完全相同，原版狀態一致，閃現 0。
固定收據與素材雜湊見 [015-poster-runtime.json](015-poster-runtime.json)，實際視窗副本見 README。
本次使用 dosgolem `49f2935`、Go 1.24.13；來源、字型與文字來源雜湊同已定版記錄。

最終切換第一輪在已產出完整終點後驅動腳本結束為 143；保留該輪，另以相同容器與命令換乾淨輸出目錄重跑，
結束碼 0，畫面及原版狀態與前輪完全相同。這是收據執行環境紀錄，不列為遊戲缺陷。
海報規格原先誤用已保留給手札地圖頁的 022，已改為 023 並修復現行連結。
