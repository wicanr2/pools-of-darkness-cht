package hdimport

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRealTheme(t *testing.T) {
	root, original, out := os.Getenv("POD_HD_ROOT"), os.Getenv("POD_HD_ORIGINAL"), os.Getenv("POD_HD_OUT")
	if root == "" || original == "" || out == "" {
		t.Skip("需要自備合法原版及正式Git素材")
	}
	raw, e := os.ReadFile(filepath.Join(root, "review/hd/runtime/manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	var m Manifest
	if e = json.Unmarshal(raw, &m); e != nil {
		t.Fatal(e)
	}
	dest := filepath.Join(out, "rebuilt-theme")
	report, e := Build(raw, root, original, dest)
	if e != nil {
		t.Fatal(e)
	}
	if len(report.Files) != 1866 {
		t.Fatal("主題數量", len(report.Files))
	}
	generated := 0
	for _, row := range m.Theme {
		data, e := os.ReadFile(filepath.Join(dest, filepath.FromSlash(row.Path)))
		if e != nil || Sum(data) != row.SHA || report.Files[row.Path] != row.SHA {
			t.Fatal(row.Path, e)
		}
		if row.Kind != "repo-file" {
			generated++
		}
	}
	if generated != 983 {
		t.Fatal("重生來源數量", generated)
	}
	t.Logf("%d檔逐SHA相同，重生%d份，原版%d檔", len(report.Files), generated, report.OriginalFiles)
	encoded, _ := json.MarshalIndent(report, "", "  ")
	if e = os.WriteFile(filepath.Join(out, "import-core.json"), encoded, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = Build(raw, root, original, dest); e == nil {
		t.Fatal("覆寫既有輸出")
	}
	changes := []struct {
		name   string
		change func(*Manifest)
	}{
		{"version", func(m *Manifest) { m.Version = 2 }},
		{"path", func(m *Manifest) { m.Theme[0].Path = "../escaped.idx" }},
		{"absolute", func(m *Manifest) { m.Theme[0].Path = "C:/escaped.idx" }},
		{"duplicate", func(m *Manifest) { m.Theme[1].Path = m.Theme[0].Path }},
		{"kind", func(m *Manifest) { m.Theme[0].Kind = "unknown" }},
		{"decoder", func(m *Manifest) { m.Theme[0].Decoder = "unknown" }},
		{"frame", func(m *Manifest) { m.Theme[0].Frame = 65535 }},
		{"sha", func(m *Manifest) { m.Theme[0].SHA = Sum([]byte("wrong")) }},
		{"original", func(m *Manifest) { m.Originals["INTRO.BMA"] = Sum([]byte("wrong")) }},
	}
	for _, tc := range changes {
		t.Run(tc.name, func(t *testing.T) {
			var bad Manifest
			json.Unmarshal(raw, &bad)
			tc.change(&bad)
			encoded, _ := json.Marshal(bad)
			d := filepath.Join(out, "negative-"+tc.name)
			if _, e = Build(encoded, root, original, d); e == nil {
				t.Fatal("未拒絕不合法輸入")
			}
			if _, e = os.Stat(d); !os.IsNotExist(e) {
				t.Fatal("留下半套輸出")
			}
		})
	}
}

func TestMalformedSources(t *testing.T) {
	for _, data := range [][]byte{nil, {1}, {9, 0}, {9, 0, 1, 255, 255, 255, 255, 1, 0, 1, 0}} {
		if _, e := dax(data); e == nil {
			t.Fatal("接受壞DAX")
		}
	}
	for _, decoder := range []string{"wall-8x8", "block-vga", "sprite-vga", "unknown"} {
		if _, e := decodeFrames([]byte{1, 2, 3}, decoder); e == nil {
			t.Fatal("接受壞影像", decoder)
		}
	}
	for _, data := range [][]byte{nil, {1}, {10, 0, 0, 0, 1, 0, 10, 0, 0, 0}, {15, 0, 0, 0, 1, 0, 10, 0, 0, 0, 64, 1, 200, 0, 0}} {
		if _, e := bma(data); e == nil {
			t.Fatal("接受壞BMA")
		}
	}
}

func TestPaths(t *testing.T) {
	for _, name := range []string{"", "../x", "a/../../x", "/x", "C:/x", `a\b`, "a/./b", "a//b"} {
		if cleanName(name) {
			t.Fatal("接受壞路徑", name)
		}
	}
	if !cleanName("intro/source/b01-f01.ops") {
		t.Fatal("拒絕合法路徑")
	}
	root := t.TempDir()
	outside := t.TempDir()
	os.WriteFile(filepath.Join(outside, "secret"), []byte("private"), 0600)
	if e := os.Symlink(outside, filepath.Join(root, "escape")); e != nil {
		t.Fatal(e)
	}
	if _, e := read(root, "escape/secret"); e == nil {
		t.Fatal("符號連結逃出來源")
	}
}
