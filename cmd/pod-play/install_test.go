package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// TestDefaultPool4CfgMatchesFirstRun 釘住預設設定檔與首次執行問答產出的檔案逐位元組相同
// （主專案 tools/g2-startup-receipt.sh 的 scratch/pool4.cfg，docs/re/003）。
func TestDefaultPool4CfgMatchesFirstRun(t *testing.T) {
	const want = "6161a336828e889c6377bd4fa327376f691a57fc3cd4f2eb6f74db06908d03bd"
	b := defaultPool4Cfg()
	sum := sha256.Sum256(b)
	if got := hex.EncodeToString(sum[:]); got != want || len(b) != 85 {
		t.Fatalf("pool4.cfg SHA-256 %s（%d bytes），要 %s（85 bytes）", got, len(b), want)
	}
}

// TestInstallKeepsExistingCfg 已有設定檔與 snd.com 時不覆寫（玩家可能改過音效或存檔路徑）。
func TestInstallKeepsExistingCfg(t *testing.T) {
	dir := t.TempDir()
	own := []byte("1\r\nV\r\nN\r\n")
	snd := []byte("sblaster")
	if err := os.WriteFile(filepath.Join(dir, "POOL4.CFG"), own, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SND.COM"), snd, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := installPod(dir); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "POOL4.CFG")); string(b) != string(own) {
		t.Fatalf("既有設定檔被改寫：%q", b)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "SND.COM")); string(b) != string(snd) {
		t.Fatalf("既有 snd.com 被改寫：%q", b)
	}
	if _, err := os.Stat(filepath.Join(dir, "pool4.cfg")); err == nil {
		if b, _ := os.ReadFile(filepath.Join(dir, "pool4.cfg")); string(b) != string(own) {
			t.Fatal("另寫了一份小寫檔名的設定檔")
		}
	}
}

// TestInstallCopiesAdlibDriver 沒有 snd.com 時照 STARTUP.EXE 的做法把 ADLIB.COM 複製過去（G2 收據以 cmp 證實兩者相同）。
func TestInstallCopiesAdlibDriver(t *testing.T) {
	dir := t.TempDir()
	drv := []byte("adlib driver")
	if err := os.WriteFile(filepath.Join(dir, "ADLIB.COM"), drv, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := installPod(dir); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "snd.com")); err != nil || string(b) != string(drv) {
		t.Fatalf("snd.com=%q err=%v", b, err)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "pool4.cfg")); string(b) != string(defaultPool4Cfg()) {
		t.Fatal("沒有寫入預設 pool4.cfg")
	}
}

// TestInstallWithoutDriverFails 缺 ADLIB.COM 時回錯，不留下半套。
func TestInstallWithoutDriverFails(t *testing.T) {
	dir := t.TempDir()
	if err := installPod(dir); err == nil {
		t.Fatal("缺 ADLIB.COM 應回錯")
	}
}
