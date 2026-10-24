package pod

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dosgolem/xlate"
)

// 以真實原始畫格重播顯示交易，不執行或注入原版CPU；正常路徑另驗。
func TestQuizSourceTransition(t *testing.T) {
	dir, fontPath, textDir := os.Getenv("DOSGOLEM_POD_QUIZ_TRANSITION"), os.Getenv("DOSGOLEM_POD_QUIZ_FONT"), os.Getenv("DOSGOLEM_POD_QUIZ_TEXT")
	if dir == "" || fontPath == "" || textDir == "" {
		t.Skip("未提供查驗題逐格來源與正式字型")
	}
	font, err := LoadFontScaled(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	NewLive(font, nil, nil) // 登記正式半形對字與3倍字模，保持與前端相同。
	read := func(path string) []byte {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	host, err := LoadHostText(read(filepath.Join(textDir, "host-text.zh-TW.tsv")))
	if err != nil {
		t.Fatal(err)
	}
	format, ok := QuizFormatOf(host)
	if !ok {
		t.Fatal("正式查驗格式缺項")
	}
	answers, err := LoadQuizAnswersFile(filepath.Join(textDir, "quiz-answers.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	for _, legacy := range []bool{true, false} {
		p, mem, layer := newQuizA(t)
		p.Font = font
		p.QuizFormat = format
		p.Answers = answers
		layer.W = 320
		layer.H = 200
		audit := xlate.NewFlashAudit(320, 200, 8)
		for f := 3503; f <= 3506; f++ {
			prefix := filepath.Join(dir, "sources", fmt.Sprintf("pod-f%06d", f))
			idx, rgbpal := read(prefix+".idx"), read(prefix+".pal")
			if len(idx) != 64000 || len(rgbpal) != 768 {
				t.Fatal("配對來源長度")
			}
			original := append([]byte(nil), idx...)
			var pal [256][3]uint8
			for i := range pal {
				copy(pal[i][:], rgbpal[i*3:i*3+3])
			}
			p.Pal = pal
			if f == 3503 {
				printA(p, mem, 1, quizHead, 2, 3)
			}
			if f == 3504 {
				// 上一格只有提早顯示的譯文；這格原文完成才返回並定色。
				p.Before(2, 0x34CF, 0x1605, 0x1CBD, 0x3F10, mem)
				// 頁碼已出現在採樣畫面，先處理它對舊列尾端的清除。
				printA(p, mem, 2, " (1)         ", 2, 17)
			}
			p.PrepareFrame(idx, 320, 200)
			rgb := make([]byte, 64000*3)
			for n, c := range idx {
				copy(rgb[n*3:n*3+3], pal[c][:])
			}
			layer.Frame(idx, rgb)
			if f == 3504 {
				// 真實畫面在這一格已顯示答案，頁碼／字序替換尚未經下一次Frame採樣。
				printA(p, mem, 3, quizAsk, 4, 3)
				printA(p, mem, 4, "1     ", 4, 19)
				printA(p, mem, 5, quizAfter, 6, 3)
				printA(p, mem, 6, "CHARACTERS AND PARTIES", 8, 4)
				if p.QuizShown == nil {
					t.Fatal("未辨識同一真實題目")
				}
				if legacy {
					for _, row := range p.QuizShown.rows[:2] {
						row.Early = false
						row.BG = [3]uint8{}
						row.FG = [3]uint8{}
					}
				}
			}
			base := make([]byte, 960*600*4)
			for y := 0; y < 600; y++ {
				for x := 0; x < 960; x++ {
					c := pal[idx[y/3*320+x/3]]
					n := (y*960 + x) * 4
					copy(base[n:n+4], []byte{c[0], c[1], c[2], 255})
				}
			}
			out := append([]byte(nil), base...)
			layer.Draw(out, 3, nil)
			audit.Observe(f-1, idx, base, out, 3)
			if !bytes.Equal(original, idx) {
				t.Fatal("改動原版色號")
			}
		}
		t.Logf("舊列替換%v：閃現%d，事件%+v", legacy, audit.Count(), audit.Events())
		if legacy && audit.Count() != 9 || !legacy && audit.Count() != 0 {
			t.Errorf("舊行為%v：閃現%d", legacy, audit.Count())
		}
	}
}
