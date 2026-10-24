package pod

import (
	"bytes"
	"os"
	"testing"
)

// 使用正式詞表核對使用者回報的前作漏譯形狀。一般／反白列都要保留數量與加值。
func TestInventoryQuantityWithOfficialCatalog(t *testing.T) {
	dir := os.Getenv("DOSGOLEM_POD_TEXT_DIR")
	if dir == "" {
		t.Skip("需本專案正式 text/；設定 DOSGOLEM_POD_TEXT_DIR")
	}
	tr, _, err := LoadTextDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ original, want string }{
		{" No  30 Arrows +1 ", " 否 箭 +1 30"},
		{"No 30 Arrows +1", "否 箭 +1 30"},
		{" Yes  50 Arrows +2 ", " 是 箭 +2 50"},
		{"Yes  50 Arrows +2", "是 箭 +2 50"},
		{"30 Arrows +1", "箭 +1 30"},
		{" No  030 Arrows +1 ", " 否 箭 +1 030"},
	}
	for _, tc := range cases {
		t.Run(tc.original, func(t *testing.T) {
			input := []byte(tc.original)
			before := append([]byte(nil), input...)
			e, ok := tr.Lookup(input)
			if !ok || e.Translation != tc.want {
				t.Fatalf("%q → %q, 命中 %v；要 %q", input, e.Translation, ok, tc.want)
			}
			if !bytes.Equal(input, before) {
				t.Fatal("顯示查找修改原文字串")
			}
			if len(PackHalfWidth([]rune(e.Translation))) > len(input) {
				t.Fatal("譯文超出原版列寬")
			}
		})
	}
}
