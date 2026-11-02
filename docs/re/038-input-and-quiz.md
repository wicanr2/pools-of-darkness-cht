# 038 遊戲滑鼠與查驗題

日期：2026-11-02。Issue：[#79](https://github.com/wicanr2/pools-of-darkness-cht/issues/79)。規格：[026](../spec/026-input-and-quiz.md)。

| 項目 | 目前狀態 |
|---|---|
| 基線 | main d9c17f2、dosgolem/pod 3fdeeb8；既有v.0.1.2封包不改。 |
| 已確認缺口 | 前端只送左鍵按下／放開，沒有持續移動及右鍵；查驗題只顯示答案。 |
| 原版滑鼠 | INT33水平回報由1D8C:00B1除2，垂直直接使用；維持320×200邏輯像素。沒有前作8倍範圍設定。前端使用系統游標，沒有自畫盾劍游標的覆繪問題。 |
| 查驗題 | 原版54題、答案DS:7DEE、輸入DS:7EEE及resident輸入1B2E:0711已由[008](008-quiz-table.md)定位；本輪補啟動時序與自動輸入收據。 |
| 目前程式 | dosgolem/pod c118740dc3cbee42881abf45a609f63375c474e7，已推送本專案私人分支。 |
| 功能驗收 | 正常冷開機自動3505格填YOU、手動關輔助、F12與五語、答錯換題4200格填SURVIVE，以及2／3倍真實滑鼠通過。自動／手動4300格CPU、RAM、畫面、色盤及存檔一致。 |
| 封包 | v.0.1.3-20261102準備中，既有24包雜湊保持；封包完成以038-input-package.json為準。 |

所有分析在Docker內執行。正式原版輸入與前作只讀，研究輸出位於workplace/input-fix-v1。

重生入口：[`tools/input_receipt.py`](../../tools/input_receipt.py)，在既有Go／Ebiten容器中掛載/repo及/original唯讀、/out為本輪研究目錄，先建置/out/pod-play，逐項執行auto、manual、raw、retry。

實際滑鼠入口：[`tools/input_mouse_gui.py`](../../tools/input_mouse_gui.py)，使用既有yuan-wine-verify:noble-r2的Xvfb、xdotool及Linux圖形函式庫，在視窗中實按左右鍵與拖曳，分別抽驗2／3倍及B版海報。

彙整入口：[`tools/input_finalize.py`](../../tools/input_finalize.py)。[原版來源038](038-input-source.json)保存工具、位址基準、bytes及輸入雜湊；[正常收據038](038-input-runtime.json)保存五種路線、同狀態及舊24包保護。

新版封包入口：[`tools/input_release.py`](../../tools/input_release.py)，先經release_prepare.py從明示main與引擎commit匯出乾淨來源，重建2034檔／18字型及十二包，再以release_verify.py的content-only核對完整封包內容。Linux實際匯入、五語開場、另存B及冷讀B、缺原版拒絕與Windows Wine另存收據，全部通過才建立新的dist-all版本。舊推廣影片保留原版號，不重標為新版。

限制：失焦清理由host／session測試涵蓋，實際視窗收據涵蓋左右鍵、區外、海報／工具列與拖曳。原生Windows及macOS滑鼠未實測。不因單元測試通過就聲稱全遊戲原版一致。

勘誤：008把02B0–02C7概括為輸入階段。本輪逐bytes核對，遠呼叫`9a11072e1b`實際位於overlay:02B9，返回02BE；02C7屬其後的字串複製。首輪把範圍末端誤當返回點，守衛因此拒絕，沒有誤輸入；修正為原始callsite後重跑。原版1D8C:0072滑鼠程序以2除INT33水平回報，垂直直接使用。1D1F:0138–0154將左鍵1轉Enter、右鍵2轉Esc；單次輪詢即放開可能在原版查詢／讀取兩階段間漏掉，正式GUI以實際按住120ms抽驗。
