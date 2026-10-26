# 研究索引

- [Go矩形填色首批接入](019-draw-interception.md#2026-10-26-go矩形填色首批接入)：#49，18組原版條件、194個計畫比較，完整事件、四路DEMO及正常介面／存讀檔；[固定收據](019-native-fill-runtime.json)。其餘VGA仍保留固定語意區塊，不稱全部Go演算法完成。

- [Go繪圖實作分類查核](019-draw-interception.md#2026-10-26-原生go繪圖要求的查核與勘誤)：#49，[正式執行檔與來源查核](019-native-go-audit.json)。目前為Go字模／共享搬移填色、VGA固定x86語意區塊及原版回退的混合實作，全部原生Go繪圖演算法尚未完成。

- [019 直接繪圖攔截](019-draw-interception.md)：#45–#47，VGA繪圖、字模、搬移、填色、裁切、鏡像與動畫；[完整狀態與事件](019-native-runtime.json)、[四路DEMO](019-demo-runtime.json)、[39對前端截圖](019-ui-runtime.json)、[填色與成本](019-rgb-cost.json)、[字模完整接管](019-whole-call-runtime.json)、[VGA全部繪圖、正常介面與新存檔往返](019-all-render-runtime.json)。B000、其他模式與未列玩家路線不外推完成。

- [018 片頭字幕來源](018-movie-subtitles.md)：#10，八組來源、逐幀淡出與雙語覆繪；正常5100格及切換通過，字幕規格024局部CONFORMED，動畫HD仍為DRAFT。

- [017 VGA 戰鬥圖](017-vga-battle-hd.md)：#23、#24及CPIC續批，完整244個來源狀態／238張重畫的目錄、覆蓋與來源診斷；正常四路抽樣的鏡像、透明區域、色盤、原版狀態及存檔比較。未觀測正常遭遇仍未知。
- [b26地下城第5組地形](017-vga-battle-hd.md#b26-地下城第5組地形)：#48，新增23張、沿用2張，純色格保留原色；[來源診斷與正常冷啟動驗收](017-b26-runtime.json)通過，25個條件式候選已啟用。
- [014 BIGPIC補證與接入](014-vga-hd.md#bigpic正常來源補證)：#44，244正常DEMO的完整色號／使用RGB與位置；七張條件式圖面、局部改寫及清除、完整四路與原版終點收據。

- [016 堆疊物品數量與反白列](016-inventory-quantity.md)：#22，正式詞表與既有正常選單收據，六項數量／加值回歸通過。

- [015 左右海報來源與版面示意](015-side-posters.md)：#21，原版來源、固定加寬接入、繁中／英文、2／3 倍及正常路徑驗收。

| 編號 | 主題 | Issue |
|---|---|---|
| [001](001-input-baseline.md) | 原始輸入基線：壓縮檔、版本、啟動流程、驅動、`SAVE/`；逐檔雜湊 [001-input-files.tsv](001-input-files.tsv) | #2 |
| [002](002-reference-sources.md) | 珍157 說明書頁序、《軟體世界》攻略掃描、SSI 英文文件來源 | #2 |
| [003](003-dosgolem-startup.md) | dosgolem 基底、通用層修正（規格 430、431）、`START.BAT` 程式鏈、mode 13h 冷啟動、`SAVE/` 讀取、收據 | #3 |
| [004](004-output-paths.md) | 輸出路徑 A／B／C／D、EXEPACK 映像、mode 13h 換算、片頭字幕、文字 catalog、主選單覆繪收據 | #4 |
| [005](005-journal-errata.md) | 珍157 旅行手札 99 則校讀、錯譯表 354 條、譯名裁決；查驗題延到 G7 | #5 |
| [006](006-manual-chapters.md) | 珍157 說明書其餘章節：全冊 OCR 草稿、魔法寶物、怪物與法術名稱 159 列對照與裁定 | #5、#6 |
| [007](007-g6-normal-path.md) | G6 正常玩家路徑：pod-play 冷開機、DEMO、建角到進入遊戲、存讀檔；查驗題兩題；待查項 | #7 |
| [008](008-quiz-table.md) | 手冊查驗題：GAME.EXE 題目表 21 筆 54 題、加密與解碼、六段畫面、核對 | #8 |
| [009](009-journal-ref.md) | 手札引用：引擎組成的整句、路徑 B 參數、單元測試；實機觸發點待 G8 | #8 |
| [010](010-g6-cross-screen.md) | G6 跨畫面盤點：pod-explore 自動探索兩輪，未命中只剩數字與引號；未覆蓋項 | #7 |
| [011](011-walkthrough.md) | 遊戲內攻略：《軟體世界》第 43–45 期的自寫摘要、照抄檢查、面板資料與收據 | #9 |
| [012](012-map-overlay.md) | 地圖疊層：本作 GEO 指標、座標、方向、東菲蘭攻略點；`g8-map-evidence.sh`、`g8-map-receipt.sh`、`map_geo.py` 及檢查器 | #9 |
| [013](013-map-identification.md) | 24 塊 GEO 名稱與五個攻略點、[裁切表](013-map-reference-crops.json)、牆面比對與攻略專名對齊工具 | #9 |
| [014](014-vga-hd.md) | 本作 VGA 圖檔、色盤、[盤點計數與別名](014-inventory-summary.json)、BIGPIC／TITLE、正常街景天空及17種牆面來源、HD素材與實際展示核對 | #10、#17 |
