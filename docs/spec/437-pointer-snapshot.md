# 437 雙鍵指標快照

狀態：CONFORMED，功能抽樣。2026-11-02。依據：host既有MouseOutput及DOS MoveMouse／PressMouse／ReleaseMouse介面。

新增可選CapturedUpdate.Pointer，含同一畫格位置、左右鍵位元與有效目標。原有MouseBridge單左鍵事件介面維持。

- 純規劃先驗證目標、位元及layout，提交前不呼叫DOS。
- 有效遊戲區才送位置及新按下。按住跨入工具列、面板、區外或失焦時釋放；仍按住返回不能補按，必須先放開再按。
- 兩鍵獨立追蹤。重複位置不送Move；清理不製造額外按下。
- 同批不允許混用舊PointerDown／Up與新快照，避免重複提交。
- 使用owner同一generation與layout檢查，錯誤整批不提交。
- 本契約只處理邏輯像素，不含遊戲位址、倍率規則或遊戲名稱。

驗證：純規劃涵蓋兩鍵、拖出返回、失焦、無效位元與面板；session涵蓋原子拒絕與原有交易回歸。

消費端正常冷開機的2／3倍實際視窗左右鍵、海報／工具列隔離及拖曳抽樣通過，原版輸入座標亦已獨立重驗。收據保存在pools-of-darkness-cht的docs/re/038-input-runtime.json；不代表其他程式或原生Windows／macOS已驗。
