# 010 G6 跨畫面盤點

日期：2026-10-24。Issue：[#7](https://github.com/wicanr2/pools-of-darkness-cht/issues/7)。接續 [007](007-g6-normal-path.md)（冷開機、建角、存讀檔、選單）。
使用者指示盤點以自動探索兩輪為止（CONTEXT 決策 2026-10-24）。

## 工具

| 工具 | 做法 |
|---|---|
| dosgolem `cmd/pod-explore` | 由 `cmd/sosb-explore` 複製：每走一步從路徑 A 的座標列（`x,y D hh:mm`）讀位置與方向，深度優先把每格四個方向都試過；冒險主選單 `Move Area …` 先按 `M` 進移動模式；`Yes No` 依 `-yes` 決定，預設 N；其他選單選能離開的選項；連續不動時依序試 Enter、Esc、空白。本作戰鬥選單有 `Quick`，`-fight` 時按 Q |
| [`tools/g6-explore.sh`](../../tools/g6-explore.sh) `<名稱> <狀態檔>` | 包裝 pod-explore；狀態檔同目錄要有 `scratch/`，複製一份再跑；輸出 `workplace/probe/g6-explore/<名稱>/`（`explore.log`、`inv.tsv`、`end.state`） |
| [`tools/g6-misses.sh`](../../tools/g6-misses.sh) `<inv.tsv>…` | 盤點中未命中的印字，排除數字、骰子、單一字母與標點 |
| [`tools/g6-service.sh`](../../tools/g6-service.sh) `<名稱> "<鍵…>"` | 由前作移植的定鍵路線（pod-overlay-shot），本輪沒用到 |

起點 `workplace/probe/g6-town/town.state` 是存讀檔收據 `lanswer` 段的狀態（讀檔、作答後回到菲蘭城街上，第 1,500,000,000 道指令），
可寫層是同一收據結束時的暫存目錄。盤點含原文，只留工作區。
探索器每送 100 鍵寫一次 `inv.tsv` 與 `end.state`；被外層時限終止時，盤點停在最後一次快照。
`tools/g7-journal-receipt.sh` 以 pod-overlay-shot 的 `-keys-at` 重播探索紀錄，依賴兩者在同一步數送鍵的效果相同；dosgolem 通用層改動按鍵排程時，這份收據會先壞。

## 結果

| 輪 | 起點 | 送鍵 | 印字（A／B／C／G） | 覆繪 | 未命中 | 結束 |
|---|---|---:|---|---:|---|---|
| town1 | 街上，`-until 'journal entry'` | 247 | 770（448／65／249／8） | 740 | 30：告示的引號、防護與生命力數字 | (0,4) 叛徒門觸發第 57 則手札後停下，造訪 73 格 |
| town2 | town1 結束處，`-yes` 店家與神殿等字 | 410（盤點快照在第 400 鍵） | 1,087（556／54／422／55），到第 400 鍵為止 | 1,055 | 32：同上 | 卡在酒館門口的 `Exit` 事件選單，到外層時限 5,400 秒被終止；盤點取最後一次快照 |

`tools/g6-misses.sh` 對兩輪的 `inv.tsv` 都是空的：遇到的介面字、對白、告示、組句都已覆繪。

經過的地點（路徑 B 的告示與對白）：客運碼頭、提爾修院（牧師對白）、訓練所門口、渡船停駛、叛徒門（Traitor's Gate；守衛、手札第 57 則）、
瓦海登公園（Valhegen Park）與噴泉、歐洛神秘器物店（店員問「要買嗎」）、舊公會軍械庫門口、擲巨魔酒館（手札第 9 則、問要不要與雙頭巨人比試）。

## 未覆蓋

- 商店購買清單、神殿治療、訓練所升級、旅店：器物店的提問前四段路徑 B 沒有 `-yes` 列的字，答了 No；其他門口只看到告示。
- 戰鬥：兩輪都沒遇到；DEMO 的戰鬥畫面見 007。
- 野外地圖、地下城、其他次元。
- town2 卡住的原因（推測）：酒館門口的事件選單只有 `Exit`，選了之後同一格的事件又觸發；探索器的脫困鍵只有 Enter、Esc、空白，沒有轉向離開。
