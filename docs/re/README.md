# 研究索引

- [Go繪圖收尾與固定畫布讀寫守衛](../spec/025-direct-render-commands.md#draw-closure-conformed)：#49已支援VGA核心CONFORMED。[收尾收據](023-draw-closure-runtime.json)、[重生入口](../../tools/draw-closure-receipt.sh)、[來源觀測器](../../tools/draw_closure_probe.py)、[檢查器](../../tools/draw_closure_check.py)。六路15,073,225 bytes全為型別化Go，原x86像素寫入0；填色搜尋與讀點另有受控原版完整呼叫，來源570f8a3已啟用。以下較早各批的DRAFT與剩餘計數保留歷史範圍，現況以上述收尾為準。

- [CPU速度互動效能與Draw量測修正](../spec/019-cpu-speed.md#效能量測修正與互動驗收ready)：[2CPU固定收據](022-speed-interactive-runtime.json)、[重生入口](../../tools/speed-interactive-receipt.sh)、[檢查器](../../tools/speed_interactive_check.py)。正常讀檔4800格同狀態、六種倍率及關窗通過；2CPU抽樣約14–18fps、實際停於1x，30fps尚未達成，保留效能限制。
  [4CPU對照](022-speed-interactive-4cpu-runtime.json)沒有CPU節流，仍在繪製低於30fps時提高倍率；修正契約見[繪製回饋](../spec/019-cpu-speed.md#繪製間隔回饋ready)。
  [修正後3倍HD](022-speed-feedback-hd-runtime.json)保留1x效能限制，已核對完整Draw時間戳與回饋，低於30fps時不再升速。
  [修正後預設2倍](022-speed-feedback-2x-runtime.json)另驗相同路線；30fps目標與1x下限的限制各自記錄。

- [工具列字級、固定倍率與1x快轉](../spec/020-toolbar-font.md)：#14、#15，[固定收據](021-toolbar-runtime.json)、[重生入口](../../tools/toolbar-acceptance-receipt.sh)、[檢查器](../../tools/toolbar_acceptance_check.py)。2／3倍三級原生字模、英文抽樣、六種倍率、偏好重啟及105875000步的狀態隔離通過；速度的互動效能仍待驗。[README字級並列](../screenshots/toolbar-font-sizes.png)由[擷取工具](../../tools/toolbar_font_preview.py)保留原始像素。

- [圖像、RGB與填色工作短續行](../spec/025-direct-render-commands.md#drawing-job-boundary-ready)：#49，[固定收據](019-native-drawing-jobs-runtime.json)、[檢查器](../../tools/native_drawing_jobs_check.py)。2523次原版邊界、172種拒絕及32次selector通過；正常111步固定與91步CPU準備改Go，僅剩14步BIOS控制。完整回歸、狀態事件與30張正常PNG相同並啟用，全部Go仍DRAFT。

- [Go畫布與子表面初始化](../spec/025-direct-render-commands.md#surface-setup-ready)：#49，[固定收據](019-native-surface-setup-runtime.json)、[檢查器](../../tools/native_surface_setup_check.py)、[固定控制觀測器產生器](../../tools/build_fixed_control_probe.py)。原版72次畫布、48次子表面與28種拒絕通過，正常893步改為Go，剩餘18份125步控制逐筆不變。完整回歸及30張正常PNG相同並啟用，全部Go仍DRAFT。

- [共享VGA填色一步續行](../spec/025-direct-render-commands.md#color-fill-span-ready)：#49，[固定收據](019-native-color-fill-runtime.json)、[檢查器](../../tools/native_color_fill_check.py)。50次有序跨度、250次完整呼叫及19種拒絕通過；正常新增1920 bytes Go填色，該路線3343240 bytes命令全由HostWork提交。完整回歸、46份原版入口及30張正常PNG通過並啟用，全部Go仍DRAFT。

- [常駐8×8字模與原生16／24像素字型](../spec/025-direct-render-commands.md#resident-glyph-ready)：#49，[固定收據](019-native-resident-glyph-runtime.json)、[總檢查器](../../tools/native_resident_glyph_check.py)、[字格影像檢查器](../../tools/native_resident_glyph_ui_check.py)。192次完整原版與正常33728點Go接管；14組基準截圖差異只在認證ASCII格、13組直接／舊合成相同，含2／3倍、英文與F12。全部Go仍DRAFT。

- [整個表面填色的一步跨度](../spec/025-direct-render-commands.md#surface-fill-spans-ready)：#49，[固定收據](019-native-surface-fill-spans-runtime.json)、[檢查器](../../tools/native_surface_fill_spans_check.py)。原版有序邊界、70次完整呼叫、全部回歸及11對正常截圖通過；正常實際新增兩個Go計畫、320 bytes填色，全部Go仍DRAFT。

- [單點、填色與斜線單步](../spec/025-direct-render-commands.md#primitive-spans-ready)：#49，[固定收據](019-native-primitive-spans-runtime.json)、[檢查器](../../tools/native_primitive_spans_check.py)。原版有序邊界、112次完整呼叫、完整回歸及11對正常截圖通過並啟用；新像素出口正常未觀測，全部Go仍DRAFT。

- [橫直線單步與寬1縮放](../spec/025-direct-render-commands.md#axis-spans-ready)：#49，[固定收據](019-native-axis-scale-runtime.json)、[檢查器](../../tools/native_axis_scale_check.py)。3504次線段邊界、零半徑與寬1完整狀態、全部回歸及11對正常截圖通過並啟用；新入口正常未觀測，全部Go仍DRAFT。

- [表面搬移的一步跨度](../spec/025-direct-render-commands.md#surface-copy-span-ready)：#49，[固定收據](019-native-surface-spans-runtime.json)、[跨度檢查器](../../tools/native_surface_spans_check.py)。400次完整原版、320次邊界比較、12次畫布／父表面邊界、完整回歸及11對正常截圖通過並啟用；像素回退0限受控條件，新入口正常未觀測，全部Go仍DRAFT。

- [Go淡入與分段DAC](../spec/025-direct-render-commands.md#palette-in-ready)：#49，[固定收據](019-native-palettein-runtime.json)、[色盤檢查器](../../tools/native_palette_check.py)。原版對拍、全部回歸、正常HD與地圖工具列11對截圖通過並啟用；固定工作1062降至1018步，淡入新入口正常未觀測，全部Go仍DRAFT。

- [變寬字模像素核心](../spec/025-direct-render-commands.md#glyph-5aa4-ready)：#49，[正式收據](019-native-glyph-runtime.json)。Go字形列與透明映射已在受控原版條件驗收啟用；正常HD及地圖工具列11對截圖相同，新入口在此正常demo未觀測，其他繪圖仍DRAFT。

- [有界共享跨度與RLE續行](019-draw-interception.md#2026-10-28-有界共享跨度與rle續行)：#49，[契約](../spec/025-direct-render-commands.md#bounded-memory-ready)、[正式收據](019-native-bounded-runtime.json)。一步預算的清除、複製與封包輸出已驗收啟用；正常固定語意2576降至1062，四路demo、39對介面、6對存讀檔與demo中斷通過，整體Go仍DRAFT。

- [Go像素轉場](019-draw-interception.md#2026-10-28-go像素轉場與hd逐格呈現接入)：#49，[正式收據](019-native-dissolve-runtime.json)。Go表面／LFSR及有界像素批次、HD辨識快取與四格保留已驗收啟用，正常閃現0；其他色盤、字模與未驗條件仍待完成，全部Go閘門DRAFT。

- [021 HD逐格呈現](021-hd-frame-presentation.json)：明確按D的demo顯示三格保留不足；完整逐格來源重現一格閃現，四格原型為0。HD辨識快取前後原版狀態及共有截圖相同；正式驗收尚待完成。

- [020 Demo控制驗收](020-demo-control-runtime.json)：本作簽章、A7CE／A7CF、等待取消與返回還原，正常冷開機、戰鬥中斷、重播及F12／2倍通過；[規格018](../spec/018-demo-control.md#demo-conformed) CONFORMED。

- [Go表面原地翻轉](019-draw-interception.md#2026-10-27-go表面原地翻轉接入)：#49，[固定收據](019-native-surfaceflip-runtime.json)。成對列與短預算輸出已驗收啟用，保留同列相依及奇數高度行為；像素轉場研究中，全部Go閘門仍DRAFT。

- [Go表面整數縮放](019-draw-interception.md#2026-10-27-go表面整數縮放接入)：#49，[固定收據](019-native-scale-runtime.json)。取樣、補色與重複列及短預算輸出已驗收啟用；原地翻轉READY待接入，全部Go閘門仍DRAFT。

- [Go圖像列](019-draw-interception.md#2026-10-27-go圖像列輸出接入)：#49，[固定收據](019-native-bitmap-runtime.json)。八平面組色、直接色號列與單步輸出已驗收啟用；原版資料準備及其他未驗路徑保留，全部Go閘門仍DRAFT。

- [Go正半徑填滿橢圓](019-draw-interception.md#2026-10-27-go正半徑填滿橢圓接入)：#49，[固定收據](019-native-ellipsefill-runtime.json)。對稱區間及單步搬移通過受控原版比較與完整正常回歸並啟用；其他圖像列研究中，全部Go閘門仍DRAFT。

- [Go正半徑橢圓外框](019-draw-interception.md#2026-10-27-go正半徑橢圓外框接入)：#49，[固定收據](019-native-ellipse-runtime.json)。兩段中點幾何及單步像素通過受控原版比較與完整正常回歸並啟用；填滿橢圓READY待接，全部Go閘門仍DRAFT。

- [Go表面矩形搬移](019-draw-interception.md#2026-10-27-go表面矩形搬移接入)：#49，[固定收據](019-native-copy-runtime.json)。重疊、四方向與奇數尾byte通過受控原版比較及完整正常回歸並啟用；[斜線計畫分類勘誤](019-draw-interception.md#斜線計畫分類勘誤)另列，全部Go閘門仍DRAFT。

- [Go斜線與裁切游標](019-draw-interception.md#2026-10-27-go斜線與裁切游標接入)：#49，[固定收據](019-native-line-runtime.json)。陡／緩斜線與相交裁切通過受控原版比較、完整正常回歸並啟用；矩形搬移研究中，全部Go閘門仍DRAFT。

- [Go軸向直線與矩形續行](019-draw-interception.md#2026-10-27-go軸向直線與矩形續行接入)：#49，[固定收據](019-native-axis-runtime.json)。312次軸線計畫、16次機器提交、92次列計畫與完整受控呼叫及正常回歸通過並啟用；陡／緩斜線原型各240次比較READY，全部Go閘門仍DRAFT。

- [Go單點與指定範圍矩形](019-draw-interception.md#2026-10-26-go單點與指定範圍矩形接入)：#49，[固定收據](019-native-primitive-runtime.json)。548個受控原版計畫、32個整台機器提交比較與完整正常回歸通過並啟用；新入口未在正常DEMO觀測。全部Go閘門仍DRAFT。

- [Go色盤資料流程與回掃等待](019-draw-interception.md#2026-10-26-go色盤資料流程與回掃等待接入)：#49，[固定收據](019-native-control-runtime.json)。完整事件、四路DEMO、39對介面及六對存讀檔通過並啟用；固定語意區塊2576步，全部Go閘門仍DRAFT。

- [Go圖片來源、裁切與翻轉](019-draw-interception.md#2026-10-26-go圖片來源裁切與翻轉接入)：#49，[固定收據](019-native-clip-runtime.json)。完整正常事件、四路DEMO、39對介面及六對新存讀檔通過並啟用；鏡像另有受控原圖比較，固定語意區塊仍20829步，全部Go閘門DRAFT。

- [Go完整RGB輸出](019-draw-interception.md#2026-10-26-go完整rgb輸出接入)：#49，[固定收據](019-native-output-runtime.json)。198次原版隔離比較、14種拒絕、通用15種拒絕及完整正常回歸通過，已啟用；固定語意區塊仍24451步，全部Go閘門DRAFT。

- [Go圖片列控制](019-draw-interception.md#2026-10-26-go圖片列控制接入)：#49，[固定收據](019-native-row-runtime.json)。558個獨立原版計畫、22種拒絕及完整正常回歸通過，已啟用；固定語意區塊仍100421步，全部Go閘門DRAFT。

- [Go色盤陣列準備](019-draw-interception.md#2026-10-26-go色盤陣列準備接入)：#49，[固定收據](019-native-setup-runtime.json)。530次原版計畫比較、21種拒絕與正常回歸通過；已啟用使用色整理、目標RGB填入及誤差初始化，其他固定路徑仍保留。

- [Go逐色RGB工作接入](019-draw-interception.md#2026-10-26-go逐色rgb工作接入)：#49，[固定收據](019-native-color-runtime.json)。32個獨立原版工作、完整事件、四路DEMO及正常介面／存讀檔通過；正式44份直接見證加三份分割補證，固定語意區塊仍117416步，未稱全部Go完成。

- [Go淡出色差與插值](019-draw-interception.md#2026-10-26-go淡出色差與插值接入)：#49，496個獨立原版計畫與正常回歸通過；[固定收據](019-native-fade-runtime.json)。正式44份直接入口與三份分割補證同原版47份，分類取正式未分割版。

- [入口見證觀測勘誤與補充工具](019-draw-interception.md#2026-10-26-入口見證數量與觀測勘誤)：前四批直接見證實為42份，原版45份；三個主機區塊內部helper入口由分割觀測探針補證，正式分類仍用未分割版。

- [Go圖片區塊與鏡像像素對](019-draw-interception.md#2026-10-26-go圖片區塊與鏡像像素對接入)：#49，[固定收據](019-native-image-runtime.json)。區塊解析82個原版計畫比較及正常路線通過；鏡像948個原版計畫、24個受控完整呼叫通過，正常DEMO未觀測鏡像新入口。

- [Go使用色號掃描接入](019-draw-interception.md#2026-10-26-go使用色號掃描接入)：#49，252個獨立原版計畫、完整事件與正常回歸；[固定收據](019-native-palette-runtime.json)。鏡像像素對原型948次比較通過，READY後待接入。

- [Go圖片RLE像素循環接入](019-draw-interception.md#2026-10-26-go圖片rle像素循環接入)：#49，302個獨立原版計畫比較、完整事件、四路DEMO、39對介面與6對存讀檔；[固定收據](019-native-rle-runtime.json)。裁切、鏡像及其他固定區塊仍待Go演算法。

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
