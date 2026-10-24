# HD 圖片審閱

新增需求：[左右海報](poster/README.md)，原版封面來源與定版原型已保存，固定加寬與語系切換已接入驗收。

[遊戲封面原版／HD](b06-cover/README.md)採使用者確認的禮服版，標題文字依語系渲染，已在正常冷啟動接入驗收。
[開場標誌與製作名單](b07-title/README.md)亦完成繁中／英文接入，三張開場、F8／F2／F12及17張完整海報視窗均通過。

Issue：[#18](https://github.com/wicanr2/pools-of-darkness-cht/issues/18)。狀態：使用者於 2026-10-24 確認「採用這批畫風，繼續全量」。
這些是 BIGPIC 資產的原版／重繪對照，尚非遊戲內 HD 截圖。

| 原版來源 | 原版 3 倍／HD 3 倍並列 |
|---|---|
| BIGPIC1 區塊 243，港灣城市 | [開啟完整比較](b01-bigpic1-243-comparison.png) |
| BIGPIC1 區塊 246，魔法戰鬥 | [開啟完整比較](b01-bigpic1-246-comparison.png) |

原始生圖、最終 912×360 預覽、完整提示詞、生成方式與 SHA-256 見 [b01-manifest.json](b01-manifest.json)。
畫風 B，原版配色及構圖為基準。只做全幅縮小，沒有裁切或用濾鏡生成 HD。
上述兩張已通過數值閘門；BIGPIC 執行期接入仍待驗證。

人物圖 PIC1-131、PIC1-043 第 0 幀已通過數值與同狀態比較，重畫及遊戲畫面均已目視核對。
來源、完整提示詞、原始生成圖與最終 PNG 雜湊見 [b02-manifest.json](b02-manifest.json)。
實際執行畫面見 [README](../../README.md#繁中與-hd-實際畫面)。
PIC 的 49 個基準畫面與 32 個動畫狀態均完成數值與目視驗收，原版並列、提示詞與逐件 SHA-256 見 [b02 全量圖](b02-base/README.md)。
三組原版完全相同的狀態共用重畫，主題登記 78 項；完整主題已通過正常 DEMO、切換及同狀態抽樣，未觀測項仍為條件式候選。

其餘五張 BIGPIC 已通過數值及逐件目視，見 [b03 大圖對照](b03-bigpic/README.md)。
加上首批兩張，七張 BIGPIC 素材均已完成；遊戲內出圖位置與接入仍待驗證。

[b04 VGA 怪物](b04-vga-actors/README.md)六個狀態與 [b05 VGA 地形](b05-vga-terrain/README.md)兩張石牆，
均完成數值與逐件目視。正常戰鬥、鏡像與透明背景的接入驗收見[研究 017](../../docs/re/017-vga-battle-hd.md)。

[b08 土鯊與摩安德碎塊](b08-cpic/README.md)六個狀態完成數值與目視，三個變體遮罩外差異0。
條件式候選已接入，既有正常DEMO與原版狀態、2倍／F12及閃現0通過，範圍見[研究017](../../docs/re/017-vga-battle-hd.md#b08-同格式候選擴充)。

本頁包含擷取的原版素材，只限目前私人儲存庫；公開前須依 AGENTS.md 移除並清理歷史。
