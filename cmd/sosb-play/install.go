package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// installSoSB 讓遊戲樹可以在 dosgolem 開機，可重複呼叫：
//   - 原版壓縮檔沒有 blades.cfg，程式第一次執行時在文字模式問顯示卡、音效、輸入裝置與存檔路徑（主專案 docs/re/003）。
//     缺檔時寫入與玩家回答 EGA、AdLib、MOUSE、預設路徑相同的設定檔（G2 收據產出的檔案，SHA-256 00cbd89e…），
//     已有就不動。
//   - 有 SAVE/ 子目錄時，比照安裝程式把預建隊伍複製到根目錄；dosgolem 在存檔路徑對不上目錄時依檔名在根目錄找
//     （dosgolem 規格 198 §2.2），已存在的檔不覆寫。本作原版的預建隊伍本來就在根目錄。
func installSoSB(root string) error {
	if _, err := findFold(root, "BLADES.CFG"); err != nil {
		if err := os.WriteFile(filepath.Join(root, "blades.cfg"), defaultBladesCfg(), 0o644); err != nil {
			return fmt.Errorf("寫入 blades.cfg：%w", err)
		}
	}
	saveDir, err := findFold(root, "SAVE")
	if err != nil {
		return nil
	}
	ents, err := os.ReadDir(filepath.Join(root, saveDir))
	if err != nil {
		return nil
	}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		if _, err := findFold(root, e.Name()); err == nil {
			continue
		}
		b, err := os.ReadFile(filepath.Join(root, saveDir, e.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(root, strings.ToUpper(e.Name())), b, 0o644); err != nil {
			return fmt.Errorf("複製預建隊伍 %s：%w", e.Name(), err)
		}
	}
	return nil
}

// defaultBladesCfg 是首次執行選 [2] EGA、[1] AdLib、[2] MOUSE、存檔路徑按 Enter 的結果
// （程式目錄 C:\SECRET；主專案 docs/re/003）：四行設定、F、14 行 0，皆以 CRLF 結尾。
func defaultBladesCfg() []byte {
	lines := []string{"E", "A", "M", `C:\SECRET\SAVE\`, "F"}
	for i := 0; i < 14; i++ {
		lines = append(lines, "0")
	}
	return []byte(strings.Join(lines, "\r\n") + "\r\n")
}
