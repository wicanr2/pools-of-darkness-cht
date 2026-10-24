package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSiblingLanguageFontPreservesExplicitChoice(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pod-en.golemfnt")
	if err := os.WriteFile(path, []byte("字型探測測試"), 0600); err != nil {
		t.Fatal(err)
	}
	fonts := map[string]string{"en": "/manual/font.golemfnt"}
	discoverSiblingLangFonts(filepath.Join(dir, "pod-rounded.golemfnt"), fonts)
	if fonts["en"] != "/manual/font.golemfnt" {
		t.Fatal("覆蓋明示字型")
	}
	fonts = map[string]string{}
	discoverSiblingLangFonts(filepath.Join(dir, "pod-rounded.golemfnt"), fonts)
	if fonts["en"] != path || len(fonts) != 1 {
		t.Fatal("未限定為實際存在的同目錄字型", fonts)
	}
}
