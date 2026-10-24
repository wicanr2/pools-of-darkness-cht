# 015 左右海報來源與版面示意

日期：2026-10-24。Issue：[#21](https://github.com/wicanr2/pools-of-darkness-cht/issues/21)。狀態：DRAFT，尚未接入。

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

使用者已於 2026-10-24 選定固定加寬，寬版繁中原圖與雜湊保存在審閱頁，正式契約見[規格 022](../spec/022-side-posters.md)。
Chromium 初次 CLI 抓圖的視窗高度包含外框，底部裁切，改用 CDP 明示 viewport 後重抓；定版圖為完整版本。
中央 960×600 與輸入截圖逐像素一致，繁中與英文文字切換均已目視。
