# 018 片頭字幕的原版來源

2026-10-25。Issue：[#10](https://github.com/wicanr2/pools-of-darkness-cht/issues/10)。
來源採樣已完成，字幕覆繪與動畫HD仍為DRAFT。前置：[研究003](003-dosgolem-startup.md)、[004](004-output-paths.md)。

## 正常原版路徑

沿既有G2的常駐`ADLIB.COM`、程式路徑`C:\POD\SND.COM`及`CONTROL.EXE`鏈，不送跳過鍵。
原版唯讀掛載，只有獨立scratch與畫格輸出可寫。每2000萬指令擷取一次，至12億指令，共59組320×200色號及768-byte色盤。
原始輸入、seed設定檔與探針SHA-256、完整命令及收據雜湊見[018-intro-source.json](018-intro-source.json)。
工具沿`dosgolem-go-ebiten:1.24-r1`映像，探針版本以收據中的二進位SHA為準，不以目前分支HEAD代替建置版本。

EXEC紀錄證實`VACE.EXE`正常exit=0，之後`GAME.EXE`已啟動。原版程式與存檔未改。
本收據不把後續GAME的自然DEMO列為字幕證據，也不把指令數當成正式顯示時間或亂數對拍。

## 採樣結論

- 字幕區`[0,168,320,32]`的RGB去重得到15組，其中第2–9組是八組開場字幕。
  其他組是版權、空白、TITLE、製作群與GAME畫面，不混為片頭台詞。
- 八組字幕皆已逐組目視校讀，英文逐字稿與繁中草稿只留本機`workplace/probe/intro-source-v1/subtitle-draft.json`。
  專名沿正式詞表的「恐怖魔王」「王國」，尚未加入正式顯示表或重建字型。
- 收據記錄每組首次與末次採樣指令數、配對色號與色盤SHA、字幕區RGB指紋及文字草稿SHA。
  這些範圍只說明採樣時可見，不是精確出現／清除時序；首次繪製、遮擋與語系切換仍未知。
- 已證實字幕在VACE畫面可見；字幕是否直接封在BMA畫格中仍沿研究004的強推論。
  沒有因此新增未經證實的檔案格式或修改原版動畫、判定與執行流程。

## 重生與下一個閘門

1. `bash tools/movie-source-receipt.sh <新名稱>`產生正常原版採樣，不覆寫舊收據。
2. 將已校讀草稿放進該本機收據目錄，再在`sosb-img:r1`內執行
   `python3 -I tools/movie_subtitle_inventory.py --capture <收據目錄> --proof docs/re/018-intro-source.json`。
   [盤點工具](../../tools/movie_subtitle_inventory.py)只輸出去重圖與來源證據，不修改遊戲或字型。
3. 先補首次繪製與清除的細格證據，再決定呈現層的辨識守衛與回退契約。
   達READY後才接入，另驗2／3倍、F8／F12、原版狀態與閃現0。

原始配對畫格、目視對照及英文字幕不加入Git。片尾`FINAL`、完整動畫HD及正式字幕接入尚未完成。
