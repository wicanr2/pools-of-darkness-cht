package main

import (
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
	t.Log("首次重建、唯讀重用與快取改動拒絕通過；不呼叫正式玩家分支")
}
