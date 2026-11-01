package hdimport

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// 以Python解碼器產生的真實候選逐SHA核對；不靠Go解碼器自己建立預期。
func TestFinalArchiveRecipes(t *testing.T) {
	orig, theme, recipes := os.Getenv("POD_FINAL_ORIGINAL"), os.Getenv("POD_FINAL_THEME"), os.Getenv("POD_FINAL_RECIPES")
	if orig == "" || theme == "" || recipes == "" {
		t.Skip("requires original FINAL and Python candidate")
	}
	raw, err := os.ReadFile(recipes)
	if err != nil {
		t.Fatal(err)
	}
	var rows []Row
	if err = json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 146 {
		t.Fatal(len(rows))
	}
	wants := map[string]string{}
	for n := range rows {
		if rows[n].Kind == "repo-file" {
			rows[n].Source = rows[n].Path
		}
		wants[rows[n].Path] = rows[n].SHA
	}
	originals := map[string]string{}
	for _, name := range []string{"FINAL.BMA", "FINAL.ACF"} {
		b, e := os.ReadFile(filepath.Join(orig, name))
		if e != nil {
			t.Fatal(e)
		}
		originals[name] = Sum(b)
	}
	m := Manifest{Version: 1, Theme: rows, Originals: originals}
	encoded, _ := json.Marshal(m)
	out := filepath.Join(t.TempDir(), "final")
	report, err := Build(encoded, theme, orig, out)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Files) != 146 {
		t.Fatal(report)
	}
	for name, want := range wants {
		b, e := os.ReadFile(filepath.Join(out, name))
		if e != nil || Sum(b) != want {
			t.Fatal("different from Python", name, e)
		}
	}
	for _, name := range []string{"unknown-archive", "wrong-movie", "changed-acf", "negative-block"} {
		t.Run(name, func(t *testing.T) {
			var bad Manifest
			json.Unmarshal(encoded, &bad)
			switch name {
			case "unknown-archive":
				bad.Theme[0].Archive = "UNKNOWN.BMA"
			case "wrong-movie":
				bad.Theme[0].Archive = "INTRO.BMA"
			case "changed-acf":
				bad.Originals["FINAL.ACF"] = Sum([]byte("wrong"))
			case "negative-block":
				bad.Theme[0].Block = -1
			}
			b, _ := json.Marshal(bad)
			dest := filepath.Join(t.TempDir(), "negative")
			if _, e := Build(b, theme, orig, dest); e == nil {
				t.Fatal("invalid archive accepted")
			}
			if _, e := os.Stat(dest); !os.IsNotExist(e) {
				t.Fatal("invalid output published")
			}
		})
	}
}
