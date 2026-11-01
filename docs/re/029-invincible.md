# 029 本作無敵欄位與收尾證據

日期：2026-11-01。Issue：[72](https://github.com/wicanr2/pools-of-darkness-cht/issues/72)。契約：[規格017](../spec/017-invincible.md)。本研究只追F1必需的玩家資料，沒有追加G6探索。

來源為本作v1.10 GAME.EXE／GAME.OVR，SHA列於規格017。dosgolem引擎基準2ec2dfa，probe由既有正常街上與正常DEMO戰場快照匯出640KiB RAM。原始快照、RAM及一次性IDA資料庫放`workplace/probe/invincible-v1/`，不入Git。IDA沿已用`ida-pro-9.4-idapython:py312-v1`，映像ID`4ac62de83339`，版本實測9.4，16-bit輸入與輸出SHA核對。工具能力入口為`/home/anr2/ida_94_official/knowledge-base/ida-94-tools.md`。

## 已證實欄位

位址分兩種：runtime段位移依正常SND／CONTROL鏈；OVR位移是原始GAME.OVR檔案位移。IDA EA另記資料庫基底，不把EA當作runtime或OVR位移。原函式名只作導覽，沒有推測性改名。

規格017列出串列、一般／戰鬥HP、健康、陣營及效果的最小讀寫端。正常街上全隊頭指向ROLAND，線性0x6FF80，一般HP40、健康0、陣營0；目前唯一效果為0x5F。正常戰場三名隊員為SIR CYNEHELM、ALEFRAED、KRISTKA，一般／戰鬥HP分別162／162、124／124、80／80。其後串敵方，敵方陣營1，有死亡敵方戰鬥HP0、健康6、可行動旗標0。這些值只證明該狀態的形狀，不是全遊戲隊員數上限。

健康顯示消費端0x23933取`[di+1EDh]`，以13為步長索引`DS:1144`；0x6D01–0x6D22的原版設定端按健康是否0計算可行動旗標。這些證據否定舊程式把`+0x188`當本作健康欄位的推定。

## 有害效果定位

| 類型 | 本作原始來源 | 等級與限制 |
|---|---|---|
| 0x21 | OVR0x2CBD1附近傳0x21後顯示失明效果；原版`is blind`在0x2CC01，解除分支在0x2CA33附近 | 已證實本作失明類型；其他施法者／地圖未窮舉 |
| 0x2A | `is slowed`在0x2E68C附近；0x2E725傳0x2A，後續0x2E756以0x2A查效果串列 | 已證實本作緩慢類型 |
| 0x34 | `paralyzed`在0x1D6AB，該函式0x1D886傳0x34；`paralyzed by poison`在0x1AA43，0x1AA95傳0x34 | 已證實本作麻痺類型 |
| 0x35 | `falls asleep`在0x2C09B，所屬函式0x2C16E查0x35；0x1AF52另有同類型呼叫 | 已證實本作睡眠類型 |
| 0x37 | `is not poisoned.`在0x3C41附近，0x3C62傳0x37查效果；`collapses from poison`在0x1B049附近，後續傳0x37 | 已證實本作中毒類型 |

舊程式註解把同引擎其他作品的效果名稱寫成本作事實，沒有足夠來源。舊0x1F、0x33不納入本作刪除表；只保留它們曾被當候選的歷史。本作效果頭是`+0x143`，不是舊`+0xE3`；一般HP與戰鬥HP分開，不能只換一個位移便啟用。

## 驗證狀態

欄位證據審查後進入READY，實作及正常抽驗已完成。F1功能依規格017升CONFORMED抽樣，完整結果見[固定執行期收據](029-invincible-runtime.json)。參數或工具失敗另記WORKLOG，不當作產品缺陷。

來源固定收據：[029來源](029-invincible-source.json)。重生工具：[IDA窄範圍匯出](../../tools/ida_pod_invincible.py)、[資料與位元組核對](../../tools/invincible_source.py)、[受控原版傷害呼叫](../../tools/invincible_oracle_test.go)、[正常玩家路線](../../tools/invincible_gui.sh)。受控呼叫明示以街上已載入的原版程式區塊搭配戰場角色資料，不冒稱原始戰場快照直接可執行該overlay或正常玩家操作。

收據由[核對工具](../../tools/invincible_check.py)重生。正常路線在Go工具容器中執行`bash /tools/invincible_gui.sh off`、`city-on`及`demo`；`audit`與`baseline`只增加量尺疑點相鄰擷取。專案唯讀掛`/project`，本專案工作目錄可寫掛`/out`並先放候選`pod-play`；前版程式需為028的218b60f，保留在`workplace/dosgolem/workplace/bin/pod-play-before-invincible`。每次使用新輸出目錄，不覆寫既有收據。

IDA沿既有非root映像，唯讀掛原版到`/orig`、工具到`/tools`，輸出掛`/out`，執行`idat -A -c -pmetapc -b16 -o/out/final-source.i64 -L/out/ida-final.log -S/tools/ida_pod_invincible.py /orig/GAME.OVR`。來源工具以`--root /project --case /out --receipt /out/source.json`核對。精確重生舊來源收據需本機已記SHA的兩份正常快照及診斷RAM；它們不散布，跨機建置F1不依賴它們。

受控測試以Go overlay把`/tools/invincible_oracle_test.go`映到`/engine/apps/pod/invincible_oracle_test.go`，在引擎根目錄以`POD_CHEAT_CASE=/out go test -overlay /out/oracle-overlay.json ./apps/pod -run TestInvincibleOriginalDamage -count=1 -v`執行。原版檔與上述RAM缺少時不宣稱通過。收據核對為`python3 -I /tools/invincible_check.py --root /project --out /out --receipt /out/runtime.json`，資源、離線及擁有權規則沿AGENTS與SETUP。

## 長戰鬥量尺限制

8191→8192的13格原先全為色號32空地，鏡頭移動後是新牆面，不能當作同一張原版圖先露出。8195→8196的列4欄4–6、列5欄5–6，以及8898→8899的列10欄15共6格，在前版及F1候選的來源與呈現完全相同，保留為既有遮擋疑點，歸#16追蹤。沒有修改繪圖算法或放寬量尺；固定收據保留全部19筆格座標與每格來源改變數量，不外推整段戰鬥零閃現。
