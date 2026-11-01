package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestHDSourceCache(t *testing.T) {
	root, original, out := os.Getenv("POD_HD_ROOT"), os.Getenv("POD_HD_ORIGINAL"), os.Getenv("POD_HD_OUT")
	if root == "" || original == "" || out == "" {
		t.Skip("需要合法原版及正式主題素材")
	}
	raw, e := os.ReadFile(filepath.Join(root, "review/hd/runtime/manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	data := filepath.Join(out, "user-data")
	first, e := prepareHDSource(raw, root, original, data)
	if e != nil {
		t.Fatal(e)
	}
	originalFile := filepath.Join(first, "intro/source/b01-f00.ops")
	before, e := os.Stat(originalFile)
	if e != nil {
		t.Fatal(e)
	}
	again, e := prepareHDSource(raw, root, original, data)
	if e != nil || first != again {
		t.Fatal("快取未重用", e)
	}
	after, e := os.Stat(originalFile)
	if e != nil || before.ModTime() != after.ModTime() {
		t.Fatal("重啟改寫快取")
	}
	var manifest struct {
		Originals map[string]string `json:"original_archives"`
	}
	if e = json.Unmarshal(raw, &manifest); e != nil {
		t.Fatal(e)
	}
	copyGame := filepath.Join(out, "copied-game")
	if e = os.Mkdir(copyGame, 0755); e != nil {
		t.Fatal(e)
	}
	for name := range manifest.Originals {
		b, e := os.ReadFile(filepath.Join(original, name))
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(copyGame, name), b, 0644); e != nil {
			t.Fatal(e)
		}
	}
	if _, e = prepareHDSource(raw, root, copyGame, data); e != nil {
		t.Fatal("拒絕同一原版的快取", e)
	}
	if e = os.WriteFile(filepath.Join(copyGame, "INTRO.BMA"), []byte("wrong"), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e = prepareHDSource(raw, root, copyGame, data); e == nil {
		t.Fatal("來源改動仍接受快取")
	}
	if e = os.WriteFile(filepath.Join(first, "unlisted.dat"), []byte("unknown"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = prepareHDSource(raw, root, original, data); e == nil {
		t.Fatal("接受多餘快取檔")
	}
	if e = os.Remove(filepath.Join(first, "unlisted.dat")); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(originalFile, []byte("wrong"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = prepareHDSource(raw, root, original, data); e == nil {
		t.Fatal("接受已改動的快取")
	}
	t.Log("首次重建、唯讀重用、原版與快取改動拒絕通過")
}
