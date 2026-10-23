package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// installPod 讓遊戲樹可以在 dosgolem 開機，可重複呼叫：
//   - 原版沒有 pool4.cfg 與 snd.com。START.BAT 先跑 STARTUP.EXE，第一次在文字模式問音效、顯示卡、輸入裝置與存檔路徑，
//     寫出 pool4.cfg 並把選定的音效驅動複製成 snd.com（主專案 docs/re/003）。設定檔已存在時 STARTUP.EXE 不開檔、不等鍵就結束
//     （主專案 workplace/probe/g6-startup-existing），所以這裡直接寫入與玩家回答 AdLib、VGA、MOUSE、預設路徑相同的設定檔
//     （G2 收據產出的檔案，SHA-256 6161a336…），並把 ADLIB.COM 複製成 snd.com。已有就不動。
//   - 原版附的 SAVE/ 是 1995 年玩家資料，遊戲內讀不到（主專案 docs/re/003），不複製。
func installPod(root string) error {
	if _, err := findFold(root, "POOL4.CFG"); err != nil {
		if err := os.WriteFile(filepath.Join(root, "pool4.cfg"), defaultPool4Cfg(), 0o644); err != nil {
			return fmt.Errorf("寫入 pool4.cfg：%w", err)
		}
	}
	if _, err := findFold(root, "SND.COM"); err == nil {
		return nil
	}
	src, err := findFold(root, "ADLIB.COM")
	if err != nil {
		return fmt.Errorf("找不到 ADLIB.COM：%w", err)
	}
	b, err := os.ReadFile(filepath.Join(root, src))
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(root, "snd.com"), b, 0o644); err != nil {
		return fmt.Errorf("寫入 snd.com：%w", err)
	}
	return nil
}

// defaultPool4Cfg 是首次執行選 [2] Adlib、[3] VGA、[2] MOUSE、兩個存檔路徑按 Enter 的結果
// （程式目錄 C:\POD；主專案 docs/re/003）：五行設定、F、14 行 0，皆以 CRLF 結尾，共 85 bytes。
func defaultPool4Cfg() []byte {
	lines := []string{"2", "V", "M", `C:\POD\SAVE\`, `C:\SECRET\SAVE\`, "F"}
	for i := 0; i < 14; i++ {
		lines = append(lines, "0")
	}
	return []byte(strings.Join(lines, "\r\n") + "\r\n")
}
