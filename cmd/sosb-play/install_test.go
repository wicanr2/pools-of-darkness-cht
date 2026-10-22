package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// TestDefaultBladesCfgMatchesFirstRun 釘住預設設定檔與首次執行問答產出的檔案逐位元組相同
// （主專案 tools/g2-startup-receipt.sh 的 scratch/blades.cfg，docs/re/003）。
func TestDefaultBladesCfgMatchesFirstRun(t *testing.T) {
	const want = "00cbd89ee14e59022941a6a5fe3c5107cf8d26b27dc60c096174a091d55923b3"
	sum := sha256.Sum256(defaultBladesCfg())
	if got := hex.EncodeToString(sum[:]); got != want {
		t.Fatalf("blades.cfg SHA-256 %s，要 %s", got, want)
	}
}

// TestInstallKeepsExistingCfg 已有設定檔時不覆寫（玩家可能改過音效或存檔路徑）。
func TestInstallKeepsExistingCfg(t *testing.T) {
	dir := t.TempDir()
	own := []byte("E\r\nS\r\nN\r\n")
	if err := os.WriteFile(filepath.Join(dir, "BLADES.CFG"), own, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := installSoSB(dir); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "BLADES.CFG"))
	if string(b) != string(own) {
		t.Fatalf("既有設定檔被改寫：%q", b)
	}
	if _, err := os.Stat(filepath.Join(dir, "blades.cfg")); err == nil && string(b) != string(own) {
		t.Fatal("另寫了一份小寫檔名的設定檔")
	}
}
