# 016 堆疊物品數量與反白列

日期：2026-10-24。Issue：[#22](https://github.com/wicanr2/pools-of-darkness-cht/issues/22)。

使用者提供前作截圖，選取列顯示 `NO 30 ARROWS +1`，要求本作注意數量顯示。本輪只核對本作，未修改前作。

本作既有正常選單收據已包含以下兩個輸出，均命中 `ready`：

| 步數 | 執行期呼叫端 CS:IP | 字串 | 路徑 |
|---:|---|---|---|
| 1,651,119,498 | 46F0:1820 | ` Yes  50 Arrows +2 ` | 一般列，底色 0、前景 10 |
| 1,651,258,268 | 46F0:19C5 | `Yes  50 Arrows +2` | 反白列，底色 15、前景 0 |

來源：`workplace/probe/g6-menus-receipt/items-inv.tsv` 與 `items-shot.json`，
原版執行期位址，不是 EXE 檔案偏移。該收據的畫面範圍外差異為 0；截圖已目視，選取列顯示「是 箭 +2 50」。
這只證實實際觀測的 50 支、加值 +2；不把前作照片當成本作缺陷。

本作正式術語 `text/glossary.tsv` 已有 Arrow／Arrows →「箭」，複數條目明記用於帶加值的物品列。
新增 `TestInventoryQuantityWithOfficialCatalog`，從正式 `text/` 載入譯文，驗 30 支 +1、50 支 +2、
一般與反白字串、獨立品名及前導零數量；六項均通過，輸入 bytes 未改，譯文未超出原列寬。
測試命令在既有 Go Docker 內以 `DOSGOLEM_POD_TEXT_DIR=/text` 執行；缺正式詞表時明確 skip。

本輪沒有改譯文、數量規則或遊戲程式，只加預防回歸。測試最初誤用「箭矢」作期望，已按正式詞表修正為「箭」。
`30 支 +1` 是查找層回歸輸入；原版 UI 實際收據仍限上述 `50 支 +2`。
