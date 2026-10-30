# HD 圖片審閱

風格 B，沿用使用者確認的奇幻手繪與原版配色。
本頁為素材對照；執行期覆蓋範圍見[主審閱入口](../README.md)。

| 來源 key／原版 3 倍 | HD 3 倍 |
|---|---|
| dungcom-003@0<br>![原版](dungcom-003@0-original.png) | ![HD](dungcom-003@0.png) |
| dungcom-003@1<br>![原版](dungcom-003@1-original.png) | ![HD](dungcom-003@1.png) |
| dungcom-003@2<br>![原版](dungcom-003@2-original.png) | ![HD](dungcom-003@2.png) |
| dungcom-003@3<br>![原版](dungcom-003@3-original.png) | ![HD](dungcom-003@3.png) |
| dungcom-003@4<br>![原版](dungcom-003@4-original.png) | ![HD](dungcom-003@4.png) |
| dungcom-003@5<br>![原版](dungcom-003@5-original.png) | ![HD](dungcom-003@5.png) |
| dungcom-003@6<br>![原版](dungcom-003@6-original.png) | ![HD](dungcom-003@6.png) |
| dungcom-003@7<br>![原版](dungcom-003@7-original.png) | ![HD](dungcom-003@7.png) |
| dungcom-003@8<br>![原版](dungcom-003@8-original.png) | ![HD](dungcom-003@8.png) |
| dungcom-003@9<br>![原版](dungcom-003@9-original.png) | ![HD](dungcom-003@9.png) |
| dungcom-003@10<br>![原版](dungcom-003@10-original.png) | ![HD](dungcom-003@10.png) |
| dungcom-003@11<br>![原版](dungcom-003@11-original.png) | ![HD](dungcom-003@11.png) |
| dungcom-003@12<br>![原版](dungcom-003@12-original.png) | ![HD](dungcom-003@12.png) |
| dungcom-003@13<br>![原版](dungcom-003@13-original.png) | ![HD](dungcom-003@13.png) |
| dungcom-003@14<br>![原版](dungcom-003@14-original.png) | ![HD](dungcom-003@14.png) |
| dungcom-003@15<br>![原版](dungcom-003@15-original.png) | ![HD](dungcom-003@15.png) |
| dungcom-003@16<br>![原版](dungcom-003@16-original.png) | ![HD](dungcom-003@16.png) |
| dungcom-003@17<br>![原版](dungcom-003@17-original.png) | ![HD](dungcom-003@17.png) |
| dungcom-003@18<br>![原版](dungcom-003@18-original.png) | ![HD](dungcom-003@18.png) |
| dungcom-003@19<br>![原版](dungcom-003@19-original.png) | ![HD](dungcom-003@19.png) |
| dungcom-003@20<br>![原版](dungcom-003@20-original.png) | ![HD](dungcom-003@20.png) |
| dungcom-003@21<br>![原版](dungcom-003@21-original.png) | ![HD](dungcom-003@21.png) |
| dungcom-003@23<br>![原版](dungcom-003@23-original.png) | ![HD](dungcom-003@23.png) |
| dungcom-003@24<br>![原版](dungcom-003@24-original.png) | ![HD](dungcom-003@24.png) |
| dungcom-003@25<br>![原版](dungcom-003@25-original.png) | ![HD](dungcom-003@25.png) |

完整來源、提示詞與 SHA-256 見 [manifest.json](manifest.json)。

第9格與第5格的原始RGB只在右側32點不同。沿既有來源差異遮罩流程保留相同區域，只在來源差異遮罩外擴6個輸出像素內使用第9格AI重繪；遮罩外差異0，不表示動畫順序或時間。

下圖是原始RGB差異遮罩，白色為來源差異，前處理另向外擴6個輸出像素：

![第9格原始RGB差異遮罩](dungcom-003@9-mask.png)

來源與失敗版本的限制見[研究017](../../../docs/re/017-vga-battle-hd.md#b30-terrain)。
