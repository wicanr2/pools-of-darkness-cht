# 015 多語系

狀態：英文前端與開場切換局部 CONFORMED（2026-10-24）；简中、日文、韓文及其他面板資料仍 DRAFT。
Issue：[#11](https://github.com/wicanr2/pools-of-darkness-cht/issues/11)、[#26](https://github.com/wicanr2/pools-of-darkness-cht/issues/26)。
沿用本作現有`apps/pod/lang.go`、`cmd/pod-play/lang.go`及前作規格015，不新增遊戲規則或語言存檔格式。

- 本輪先補英文`host-text.en.tsv`，placeholder與本作繁中表對應，問句使用本作頁碼／字序形式。
  不直接使用前作`quiz.entry`格式。玩家可見原文由既有identityLookup以英文字型重繪，原版比較及存檔不使用譯文。
- 英文字型沿前作`tools/en_font.py`與Courier Prime Bold，2倍16×16、3倍24×24，填滿原版方形字格。
  實際來源、授權、SHA及可重現入口掛於研究014。繁中字型與原版點陣字不混用。
- `-lang-fonts`仍有最高優先。開發模式另可讀主字型旁既有`pod-<語系>.golemfnt`，沒有該字型就不載該語系。
  發行模式沿既有`font/pod-<語系>.golemfnt`。目前提供繁中與英文，F8循環，F12原版獨立。
- 第一次依系統語言、之後記住選擇，沿現有偏好檔；不重開遊戲、不送額外DOS按鍵。
  面板沒有該語言資料時停用，不沿用上一語言內容。本輪不宣稱英文手札及攻略完成。
  英文地圖只先提供已證實的東西菲蘭名稱；其他名稱顯示未知，沒有英文攻略點，不露出繁中名稱。
- 驗收正式英文前端表、字型尺寸與覆蓋、語系載入與優先序、三張正常開場及F8／F2／F12，
  CPU、記憶體、色號、色盤與DOS檔案不變。正文、手札、攻略及其他語系後續範圍仍由#11追蹤。

驗收：dosgolem `f3d019e`，正式前端與專用字型、語系載入優先序、1250格三張開場的F8／F2／F12均通過。
另以2210格正常DEMO抽測英文原文的新字型，A／B路徑均覆繪、無缺字、閃現0；原版記憶體、色號與127檔同既有基準。
證據：[開場](../re/014-title-runtime.json)、[英文對話](../re/014-english-front-runtime.json)。其餘畫面未窮舉，英文手札與攻略未交付。
