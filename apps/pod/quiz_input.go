package pod

// 本作原版1B2E:0711輸入程序。位置以已綁定路徑A的resident基址換算。
// 證據：主專案008、038；只觀察，鍵盤由前端交給session。
type quizInput struct {
	shown               *QuizShown
	ss, sp, ds, retCS   uint16
	active, ready, done bool
	answer              string
}

func pascalText(v CPUView, at uint32, max int) string {
	n := int(v.Read8(at))
	if n > max {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = v.Read8(at + 1 + uint32(i))
	}
	return string(b)
}

func (l *Live) observeQuizInput(v StepCPU) {
	q := &l.quizInput
	if q.active && v.SS() == q.ss && v.CS() == q.retCS && v.IP() == 0x2be {
		q.active, q.ready = false, false
	}
	base := l.A.Entry - 0x354
	if linear(v.CS(), v.IP()) == base+0x711 {
		*q = quizInput{}
		s := l.A.QuizShown
		sp, ss := v.SP(), v.SS()
		ret, cs := v.Read16(linear(ss, sp)), v.Read16(linear(ss, sp+2))
		// maxLength=20、固定查驗呼叫端、答案與正式表一致。
		if s == nil || s.Answer == "" || ret != 0x2be || v.Read8(linear(ss, sp+4)) != 20 || pascalText(v, linear(v.DS(), 0x7dee), 20) != s.Answer {
			return
		}
		sig := []byte{0x55, 0x89, 0xe5, 0x81, 0xec, 0x2c, 0x03, 0xc4, 0x7e, 0x0c}
		for i, b := range sig {
			if v.Read8(base+0x711+uint32(i)) != b {
				return
			}
		}
		call := linear(cs, ret-5)
		if v.Read8(call) != 0x9a || v.Read16(call+1) != 0x711 || v.Read16(call+3) != v.CS() {
			return
		}
		*q = quizInput{shown: s, ss: ss, sp: sp, ds: v.DS(), retCS: cs, active: true, answer: s.Answer}
	}
	if q.active && !q.done && linear(v.CS(), v.IP()) == base+0x771 && v.SS() == q.ss {
		// 原版在SS:BP-12Bh建立編輯字串，BP=入口SP-2。
		if v.Read8(linear(q.ss, q.sp-2-0x12b)) != 0 {
			q.done, q.ready = true, false
		} else if pascalText(v, linear(q.ds, 0x7dee), 20) == q.answer {
			q.ready = true
		}
	}
}

// TakeQuizAnswer 回傳一次待輸入的原文，不含Enter。玩家先輸入時不追加。
func (l *Live) TakeQuizAnswer(playerInput bool) string {
	q := &l.quizInput
	if !q.active || !q.ready || q.done || l.hideAnswers {
		return ""
	}
	q.done, q.ready = true, false
	if playerInput {
		return ""
	}
	return q.answer
}
