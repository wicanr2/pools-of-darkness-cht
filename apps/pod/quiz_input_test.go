package pod

import "testing"

type quizCPU struct {
	fakeMem
	cs, ip, ss, sp, ds uint16
}

func (v quizCPU) Steps() uint64 { return 1 }
func (v quizCPU) CS() uint16    { return v.cs }
func (v quizCPU) IP() uint16    { return v.ip }
func (v quizCPU) SS() uint16    { return v.ss }
func (v quizCPU) SP() uint16    { return v.sp }
func (v quizCPU) DS() uint16    { return v.ds }

func TestQuizInputOnlyVerifiedEditor(t *testing.T) {
	v := quizCPU{fakeMem: make(fakeMem, 0x100000), cs: 0x1b2e, ip: 0x711, ss: 0x3000, sp: 0xf000, ds: 0x2096}
	put := func(at uint32, w uint16) { v.fakeMem[at] = byte(w); v.fakeMem[at+1] = byte(w >> 8) }
	entry := linear(v.cs, 0x711)
	copy(v.fakeMem[entry:], []byte{0x55, 0x89, 0xe5, 0x81, 0xec, 0x2c, 0x03, 0xc4, 0x7e, 0x0c})
	put(linear(v.ss, v.sp), 0x2be)
	put(linear(v.ss, v.sp+2), 0x5509)
	put(linear(v.ss, v.sp+4), 20)
	call := linear(0x5509, 0x2b9)
	v.fakeMem[call] = 0x9a
	put(call+1, 0x711)
	put(call+3, v.cs)
	a := linear(v.ds, 0x7dee)
	v.fakeMem[a] = 5
	copy(v.fakeMem[a+1:], "LIVES")
	l := &Live{A: &PathA{Entry: linear(v.cs, 0x354), QuizShown: &QuizShown{Answer: "LIVES"}}}
	l.observeQuizInput(v)
	if s := l.TakeQuizAnswer(false); s != "" {
		t.Fatal("before editor", s)
	}
	v.ip = 0x771
	l.observeQuizInput(v)
	l.hideAnswers = true
	if s := l.TakeQuizAnswer(false); s != "" {
		t.Fatal("disabled", s)
	}
	l.hideAnswers = false
	if s := l.TakeQuizAnswer(false); s != "LIVES" {
		t.Fatal("ready", s)
	}
	if s := l.TakeQuizAnswer(false); s != "" {
		t.Fatal("duplicate", s)
	}
	// A new attempt is independently armed, but manual input wins.
	v.ip = 0x711
	l.observeQuizInput(v)
	v.ip = 0x771
	l.observeQuizInput(v)
	if s := l.TakeQuizAnswer(true); s != "" {
		t.Fatal("manual", s)
	}
	if s := l.TakeQuizAnswer(false); s != "" {
		t.Fatal("after manual", s)
	}
	v.ip = 0x711
	l.observeQuizInput(v)
	v.ip = 0x771
	l.observeQuizInput(v)
	v.cs, v.ip = 0x5509, 0x2be
	l.observeQuizInput(v)
	if s := l.TakeQuizAnswer(false); s != "" {
		t.Fatal("stale", s)
	}
	v.cs, v.ip = 0x1b2e, 0x711
	put(linear(v.ss, v.sp+4), 19)
	l.observeQuizInput(v)
	v.ip = 0x771
	l.observeQuizInput(v)
	if s := l.TakeQuizAnswer(false); s != "" {
		t.Fatal("wrong editor", s)
	}
	put(linear(v.ss, v.sp+4), 20)
	v.ip = 0x711
	l.observeQuizInput(v)
	v.fakeMem[linear(v.ss, v.sp-2-0x12b)] = 1
	v.ip = 0x771
	l.observeQuizInput(v)
	if s := l.TakeQuizAnswer(false); s != "" {
		t.Fatal("existing content", s)
	}
}
