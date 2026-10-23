package pod

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/wicanr2/dosgolem/xlate"
)

// 規格 006：手冊查驗題。路徑 A 依序印六段（主專案 docs/re/008）：
// 「 journal, page 」、頁碼、「 what is word # 」、字序、「 after the heading 」、標題；
// 第 6 段時把第 2、4、6 列改寫成中文並附答案，標題留英文。

// QuizKey 是答案表的索引：英文手札頁碼、字序、標題。
type QuizKey struct {
	Page, Word int
	Heading    string
}

// LoadQuizAnswers 讀 text/quiz-answers.tsv（page、word、heading、answer）。
func LoadQuizAnswers(b []byte) (map[QuizKey]string, error) {
	rows, err := readTSV(b, []string{"page", "word", "heading", "answer"})
	if err != nil {
		return nil, fmt.Errorf("pod: 查驗答案表：%w", err)
	}
	m := make(map[QuizKey]string, len(rows))
	for i, r := range rows {
		var k [2]int
		for j := range k {
			if k[j], err = strconv.Atoi(r[j]); err != nil {
				return nil, fmt.Errorf("pod: 查驗答案表第 %d 列：%w", i+2, err)
			}
		}
		m[QuizKey{k[0], k[1], r[2]}] = r[3]
	}
	return m, nil
}

// LoadQuizAnswersFile 讀答案表檔案。
func LoadQuizAnswersFile(path string) (map[QuizKey]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadQuizAnswers(b)
}

// 六段的原文與相對位置（主專案 docs/re/008「畫面」）：頁碼接在第 1 段的第 14 格，字序接在第 3 段之後，
// 第 3、5 段在第 1 段下方 2、4 列，標題在下方 6 列、右移 1 格。
var (
	quizHead    = " journal, page "
	quizAsk     = " what is word # "
	quizAfter   = " after the heading "
	quizPage    = regexp.MustCompile(`^ \(([0-9]+)\) *$`)
	quizWord    = regexp.MustCompile(`^([0-9]+) *$`)
	quizHeading = regexp.MustCompile(`^[A-Z0-9]`)
)

const quizPageCol = 14

type quizTrack struct {
	stage            int
	row, col         int
	page, word       int
	pageLen, wordLen int
	head, ask, after *xlate.Stamp
	bg               uint8
}

// QuizFormat 是改寫後三列的格式（text/host-text.zh-TW.tsv 的 quiz.page、quiz.word、quiz.answer），
// 佔位 {page}、{word}、{heading}、{answer}。
type QuizFormat [3]string

// LoadHostText 讀 text/host-text.zh-TW.tsv（key、translation、note）：前端自己畫的固定中文。
func LoadHostText(b []byte) (map[string]string, error) {
	rows, err := readTSV(b, []string{"key", "translation"})
	if err != nil {
		return nil, fmt.Errorf("pod: 前端文字表：%w", err)
	}
	m := make(map[string]string, len(rows))
	for _, r := range rows {
		m[r[0]] = r[1]
	}
	return m, nil
}

// QuizFormatOf 從前端文字表取查驗題格式；缺任一項回 false。
func QuizFormatOf(host map[string]string) (QuizFormat, bool) {
	var f QuizFormat
	for i, k := range []string{"quiz.page", "quiz.word", "quiz.answer"} {
		if f[i] = host[k]; f[i] == "" {
			return f, false
		}
	}
	return f, true
}

// lines 是改寫後的三列（規格 006 §2.2）；答案關閉或表中沒有時第 3 列空白。
func (f QuizFormat) lines(k QuizKey, answer string, show bool) [3]string {
	r := strings.NewReplacer("{page}", strconv.Itoa(k.Page), "{word}", strconv.Itoa(k.Word),
		"{heading}", k.Heading, "{answer}", answer)
	l := [3]string{r.Replace(f[0]), r.Replace(f[1])}
	if show && answer != "" {
		l[2] = r.Replace(f[2])
	}
	return l
}

func fitText(s string, cells int) ([]rune, bool) {
	r := PackHalfWidth([]rune(s))
	if len(r) > cells {
		return nil, false
	}
	for len(r) < cells {
		r = append(r, ' ')
	}
	return r, true
}

// observeQuiz 在每次路徑 A 呼叫後呼叫；st 是這次加入的 Stamp（未覆繪為 nil）。
// 六段中間夾了別的呼叫，或任何一段對不上，就放棄這一組。
func (p *PathA) observeQuiz(text []byte, row, col, bg int, st *xlate.Stamp) {
	q := &p.quiz
	s := string(text)
	switch {
	case q.stage == 1 && row == q.row && col == q.col+quizPageCol && quizPage.MatchString(s):
		q.page, _ = strconv.Atoi(quizPage.FindStringSubmatch(s)[1])
		q.pageLen, q.stage = len(s), 2
		return
	case q.stage == 2 && st != nil && s == quizAsk && row == q.row+2 && col == q.col:
		q.ask, q.stage = st, 3
		return
	case q.stage == 3 && row == q.row+2 && col == q.col+len(quizAsk) && quizWord.MatchString(s):
		q.word, _ = strconv.Atoi(quizWord.FindStringSubmatch(s)[1])
		q.wordLen, q.stage = len(s), 4
		return
	case q.stage == 4 && st != nil && s == quizAfter && row == q.row+4 && col == q.col:
		q.after, q.stage = st, 5
		return
	case q.stage == 5 && row == q.row+6 && col == q.col+1 && quizHeading.MatchString(s) && p.QuizFormat[0] != "":
		if st != nil { // 標題留英文（規格 006 §2.2）
			p.Layer.Clear(st.Rect())
		}
		p.finishQuiz(QuizKey{q.page, q.word, s})
	}
	*q = quizTrack{}
	if s == quizHead && st != nil {
		*q = quizTrack{stage: 1, row: row, col: col, head: st, bg: uint8(bg)}
	}
}

// widen 以同原點、較長的 Stamp 取代 st，涵蓋到後面數字段的結尾。
func (p *PathA) widen(st *xlate.Stamp, cells int, text []rune, key string) *xlate.Stamp {
	w := &xlate.Stamp{Key: key, X: st.X, Y: st.Y, Cells: cells, CellW: 8, CellH: 8,
		Font: p.Font, Text: text, State: xlate.Pending}
	p.Layer.Replace(w)
	if p.bgOf == nil {
		p.bgOf = map[*xlate.Stamp]uint8{}
	}
	p.bgOf[w] = p.quiz.bg
	return w
}

func (p *PathA) finishQuiz(k QuizKey) {
	q := &p.quiz
	ans := p.Answers[k]
	lines := p.QuizFormat.lines(k, ans, !p.HideAnswers)
	headCells := quizPageCol + q.pageLen
	askCells := len(quizAsk) + q.wordLen
	t0, ok0 := fitText(lines[0], headCells)
	t1, ok1 := fitText(lines[1], askCells)
	t2, ok2 := fitText(lines[2], q.after.Cells)
	if !ok0 || !ok1 || !ok2 {
		return
	}
	head := p.widen(q.head, headCells, t0, "quiz.page")
	ask := p.widen(q.ask, askCells, t1, "quiz.word")
	q.after.Text = t2
	delete(p.src, q.after) // 第 6 列的內容由查驗題格式決定，換語言時不照 catalog 重譯
	p.QuizShown = &QuizShown{Key: k, Answer: ans, rows: [3]*xlate.Stamp{head, ask, q.after}}
}

// QuizShown 記最近一次改寫的查驗題，供切換答案顯示與換語言。
type QuizShown struct {
	Key    QuizKey
	Answer string
	rows   [3]*xlate.Stamp // 第 2、4、6 列
}

// SetHideAnswers 切換答案顯示（只改覆繪，規格 006 §2.2）。
func (p *PathA) SetHideAnswers(hide bool) {
	p.HideAnswers = hide
	if s := p.QuizShown; s != nil {
		if t, ok := fitText(p.QuizFormat.lines(s.Key, s.Answer, !hide)[2], s.rows[2].Cells); ok {
			s.rows[2].Text = t
		}
	}
}

// relangQuiz 換語言後以新格式重寫最近一次改寫的查驗題三列；放不下的列保持原樣。
func (p *PathA) relangQuiz() {
	s := p.QuizShown
	if s == nil || p.QuizFormat[0] == "" {
		return
	}
	lines := p.QuizFormat.lines(s.Key, s.Answer, !p.HideAnswers)
	for i, st := range s.rows {
		if t, ok := fitText(lines[i], st.Cells); ok {
			st.Text, st.Font = t, p.Font
		}
	}
}
