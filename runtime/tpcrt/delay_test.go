package tpcrt

import "testing"

func memWithForm(f delayForm, at ...uint32) []byte {
	m := make([]byte, searchEnd)
	for _, a := range at {
		for i, b := range f.sig {
			if b >= 0 {
				m[a+uint32(i)] = uint8(b)
			}
		}
		m[a+f.countOperand], m[a+f.countOperand+1] = 0x8A, 0xFB
	}
	return m
}

func memWith(at ...uint32) []byte { return memWithForm(delayForms[0], at...) }

// TestFindDelayTP5 以 Secret of the Silver Blades 1990 鎮上記憶體的實際位置（線性 0CB2E、DelayCnt DS:A6E4）。
func TestFindDelayTP5(t *testing.T) {
	m := memWithForm(delayForms[1], 0xCB2E)
	m[0xCB2E+0x13], m[0xCB2E+0x14] = 0xE4, 0xA6
	loop, count, ok := FindDelay(func(a uint32) uint8 { return m[a] }, 0x0EF2)
	if !ok || loop != 0xCB4E || count != 0xEF20+0xA6E4 {
		t.Fatalf("loop=%05X count=%05X ok=%v", loop, count, ok)
	}
	tp7 := memWith(0x20000)
	for i := uint32(0x20000); i < 0x20000+uint32(len(delaySig)); i++ {
		m[i] = tp7[i]
	}
	if _, _, ok := FindDelay(func(a uint32) uint8 { return m[a] }, 0x0EF2); ok {
		t.Fatal("TP5 與 TP7 各一處命中仍回 ok")
	}
}

func TestFindDelay(t *testing.T) {
	m := memWith(0xA62C)
	loop, count, ok := FindDelay(func(a uint32) uint8 { return m[a] }, 0x0CDE)
	if !ok || loop != 0xA62C+0x21 || count != 0xCDE0+0xFB8A {
		t.Fatalf("loop=%05X count=%05X ok=%v", loop, count, ok)
	}
	m = memWith(0xA62C, 0x20000)
	if _, _, ok := FindDelay(func(a uint32) uint8 { return m[a] }, 0x0CDE); ok {
		t.Fatal("兩處命中仍回 ok")
	}
	m = make([]byte, searchEnd)
	if _, _, ok := FindDelay(func(a uint32) uint8 { return m[a] }, 0x0CDE); ok {
		t.Fatal("沒有命中仍回 ok")
	}
}
