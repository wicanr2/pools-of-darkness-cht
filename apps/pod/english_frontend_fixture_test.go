package pod

import (
	"bytes"
	"os"
	"testing"
)

func TestFormalEnglishFrontendAndFont(t *testing.T) {
	dir, path := os.Getenv("DOSGOLEM_POD_TEXT_DIR"), os.Getenv("DOSGOLEM_POD_EN_FONT")
	if dir == "" || path == "" {
		t.Skip("未提供本作英文前端及專用字型")
	}
	font, err := LoadFontScaled(path)
	if err != nil {
		t.Fatal(err)
	}
	if font.W != 16 || font.H != 16 || font.Scaled[3].Font == nil || font.Scaled[3].Font.W != 24 {
		t.Fatal("英文2/3倍字型不符")
	}
	pack, err := LoadLangPack(dir, EnglishLang, font)
	if err != nil {
		t.Fatal(err)
	}
	lines := pack.QuizFormat.lines(QuizKey{Page: 104, Word: 10}, "ANSWER", true)
	for n, line := range lines {
		if len([]rune(line)) > []int{27, 22, 19}[n] {
			t.Fatal("本作英文查驗問句超寬", line)
		}
	}
	original := []byte("50 Arrows +2")
	before := append([]byte(nil), original...)
	entry, ok := pack.UI.Lookup(original)
	if !ok || entry.Translation != string(original) || !bytes.Equal(before, original) {
		t.Fatal("英文原文顯示查表修改語意輸入")
	}
}
