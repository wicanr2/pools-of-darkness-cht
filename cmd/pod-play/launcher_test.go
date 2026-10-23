package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withFakeOriginals 把必要檔換成測試內容（不需原版）。
func withFakeOriginals(t *testing.T) map[string][]byte {
	t.Helper()
	files := map[string][]byte{"CONTROL.EXE": []byte("fake control"), "GAME.OVR": []byte("fake ovr")}
	old := requiredOriginals
	t.Cleanup(func() { requiredOriginals = old })
	requiredOriginals = nil
	for _, n := range []string{"CONTROL.EXE", "GAME.OVR"} {
		s := sha256.Sum256(files[n])
		requiredOriginals = append(requiredOriginals, struct{ name, sha string }{n, hex.EncodeToString(s[:])})
	}
	return files
}

func TestPrepareGameImportsOnceCaseInsensitive(t *testing.T) {
	files := withFakeOriginals(t)
	orig, data := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(orig, "control.exe"), files["CONTROL.EXE"], 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(orig, "Game.Ovr"), files["GAME.OVR"], 0o644); err != nil {
		t.Fatal(err)
	}
	game, start, err := prepareGame(orig, data)
	if err != nil {
		t.Fatal(err)
	}
	if game != filepath.Join(data, "game") || start != "control.exe" {
		t.Fatalf("game=%s start=%s", game, start)
	}
	// 第二次不再需要原版
	if _, _, err := prepareGame(filepath.Join(orig, "沒有"), data); err != nil {
		t.Fatalf("已匯入後應只核對 game/：%v", err)
	}
	// game/ 被改壞時要求重新匯入
	if err := os.WriteFile(filepath.Join(game, "Game.Ovr"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := prepareGame(orig, data); err == nil || !strings.Contains(err.Error(), "核對失敗") {
		t.Fatalf("game/ 雜湊不符時應失敗：%v", err)
	}
}

func TestPrepareGameMissingOriginal(t *testing.T) {
	withFakeOriginals(t)
	data := t.TempDir()
	_, _, err := prepareGame(filepath.Join(data, "nowhere"), data)
	if err == nil || !strings.Contains(err.Error(), "找不到原版") {
		t.Fatalf("缺原版時應給說明：%v", err)
	}
	if _, err := os.Stat(filepath.Join(data, "game")); err == nil {
		t.Fatal("失敗時不得留下 game/")
	}
}

func TestPrepareGameHashMismatch(t *testing.T) {
	files := withFakeOriginals(t)
	orig, data := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(orig, "CONTROL.EXE"), files["CONTROL.EXE"], 0o644)
	os.WriteFile(filepath.Join(orig, "GAME.OVR"), []byte("other version"), 0o644)
	_, _, err := prepareGame(orig, data)
	if err == nil || !strings.Contains(err.Error(), "SHA-256 不符") {
		t.Fatalf("版本不符時應說明雜湊：%v", err)
	}
}

func TestRequiredOriginalsMatchBaseline(t *testing.T) {
	// 主專案 docs/re/001
	want := map[string]string{
		"CONTROL.EXE": "60e53bcf4266fb4e68e4fc22990159d1c1af978e9d10068708cfd11d49935366",
		"GAME.EXE":    "35ad635e558852925b7028a921aacafc2df8e84be24a8a1136c6267471948d69",
		"GAME.OVR":    "885d6a07a3658f0688dc9d0630dfa984683a3d1d1e515a8e9909d29f62012485",
		"ADLIB.COM":   "7c39949b5c740d92449110e896f08f2d11b16b8602d7d3d6d59a9f250d574286",
	}
	if len(requiredOriginals) != len(want) {
		t.Fatalf("必要檔 %d 個，要 %d 個", len(requiredOriginals), len(want))
	}
	for _, r := range requiredOriginals {
		if want[r.name] != r.sha {
			t.Fatalf("%s 雜湊與 docs/re/001 不符", r.name)
		}
	}
}
