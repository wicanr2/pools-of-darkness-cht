package pod

import (
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/xlate"
)

const quizHostZh = "key\ttranslation\tnote\n" +
	"quiz.page\t英文手札第 {page} 頁\t\nquiz.word\t下列標題之後的第 {word} 個字？\t\nquiz.answer\t答案：{answer}\t\n"

func newQuizA(t *testing.T) (*PathA, fakeMem, *xlate.Layer) {
	t.Helper()
	m := make(fakeMem, 0x100000)
	plantSignature(m, 0x62AA, 0)
	tr := &Translator{Exact: []Lookuper{catalogOf(t, quizHead, "手札，頁", quizAsk, "請問字序 #", quizAfter, "於標題之後")}}
	layer := &xlate.Layer{}
	p, err := NewPathA(m, tr, &xlate.Font{W: 16, H: 16, Glyphs: map[rune][]byte{}}, layer)
	if err != nil {
		t.Fatal(err)
	}
	p.Answers = map[QuizKey]string{{4, 7, "ALIGNMENT"}: "LIVES"}
	host, err := LoadHostText([]byte(quizHostZh))
	if err != nil {
		t.Fatal(err)
	}
	var ok bool
	if p.QuizFormat, ok = QuizFormatOf(host); !ok {
		t.Fatal("格式不全")
	}
	return p, m, layer
}

func printA(p *PathA, m fakeMem, step uint64, s string, row, col uint8) {
	b := linear(0x5777, 8)
	m[b] = byte(len(s))
	copy(m[b+1:], s)
	callFrame(m, 0x1CBD, 0x3F00, 0x34CF, 0x1605, 0x5777, 8, 0, 10, row, col)
	p.Before(step, 0x05F4, 0x036A, 0x1CBD, 0x3F00, m)
}

func quizRows(l *xlate.Layer) map[int]string {
	rows := map[int]string{}
	for _, s := range l.Stamps {
		rows[s.Y/8] = strings.TrimRight(string(s.Text), " ")
	}
	return rows
}

// askQuiz 照原版的六段印一題（主專案 docs/re/008「畫面」）。
func askQuiz(p *PathA, m fakeMem, page, word, heading string) {
	printA(p, m, 1, quizHead, 2, 3)
	printA(p, m, 2, " ("+page+")         ", 2, 17)
	printA(p, m, 3, quizAsk, 4, 3)
	printA(p, m, 4, word+"     ", 4, 19)
	printA(p, m, 5, quizAfter, 6, 3)
	printA(p, m, 6, heading, 8, 4)
}

func TestQuizRewritesWithAnswer(t *testing.T) {
	p, m, l := newQuizA(t)
	askQuiz(p, m, "4", "7", "ALIGNMENT")
	r := quizRows(l)
	want := map[int]string{
		2: string(PackHalfWidth([]rune("英文手札第 4 頁"))),
		4: string(PackHalfWidth([]rune("下列標題之後的第 7 個字？"))),
		6: string(PackHalfWidth([]rune("答案：LIVES"))),
	}
	for row, w := range want {
		if r[row] != w {
			t.Errorf("第 %d 列 %q，要 %q", row, r[row], w)
		}
	}
	cells := map[int]int{2: 27, 4: 22, 6: 19}
	for _, s := range l.Stamps {
		if c, ok := cells[s.Y/8]; ok && s.Cells != c {
			t.Errorf("第 %d 列 %d 格，要 %d 格（延伸到數字段結尾）", s.Y/8, s.Cells, c)
		}
	}
	if _, ok := r[8]; ok {
		t.Errorf("標題列不該覆繪：%q", r[8])
	}
	p.SetHideAnswers(true)
	if r := quizRows(l); r[6] != "" {
		t.Errorf("隱藏答案後第 6 列應空白：%q", r[6])
	}
}

func TestQuizUnknownAnswerAndHidden(t *testing.T) {
	p, m, l := newQuizA(t)
	p.HideAnswers = true
	askQuiz(p, m, "4", "7", "ALIGNMENT")
	if r := quizRows(l); r[6] != "" || r[2] == "" {
		t.Errorf("隱藏答案：%q", r)
	}
	p, m, l = newQuizA(t)
	askQuiz(p, m, "4", "5", "ALIGNMENT")
	if r := quizRows(l); r[6] != "" || !strings.Contains(r[4], string(PackHalfWidth([]rune("第 5 個")))) {
		t.Errorf("表中無答案：%q", r)
	}
}

func TestQuizInterruptedKeepsNormalResult(t *testing.T) {
	p, m, l := newQuizA(t)
	printA(p, m, 1, quizHead, 2, 3)
	printA(p, m, 2, " (4)         ", 2, 17)
	printA(p, m, 3, "ANNA", 10, 3) // 中間插入別的呼叫
	printA(p, m, 4, quizAsk, 4, 3)
	printA(p, m, 5, "7     ", 4, 19)
	printA(p, m, 6, quizAfter, 6, 3)
	printA(p, m, 7, "ALIGNMENT", 8, 4)
	if r := quizRows(l); r[4] != string(PackHalfWidth([]rune("請問字序 #"))) || p.QuizShown != nil {
		t.Errorf("順序被打斷不該改寫：%q", r)
	}
}

func TestQuizHeadingStaysEnglish(t *testing.T) {
	p, m, l := newQuizA(t)
	p.Catalog.(*Translator).Exact = append(p.Catalog.(*Translator).Exact, catalogOf(t, "ALIGNMENT", "陣營"))
	askQuiz(p, m, "4", "7", "ALIGNMENT")
	if r := quizRows(l); r[8] != "" || r[6] != string(PackHalfWidth([]rune("答案：LIVES"))) {
		t.Errorf("標題要留英文：%q", r)
	}
}

func TestLoadQuizAnswers(t *testing.T) {
	m, err := LoadQuizAnswers([]byte("page\tword\theading\tanswer\n2\t3\tABILITY SCORES\tHAS\n"))
	if err != nil || m[QuizKey{2, 3, "ABILITY SCORES"}] != "HAS" {
		t.Fatalf("%v %v", m, err)
	}
	if _, err := LoadQuizAnswers([]byte("page\tentry\tword\tanswer\n")); err == nil {
		t.Fatal("前作的表頭要拒絕")
	}
}
