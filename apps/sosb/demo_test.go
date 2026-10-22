package sosb

import "testing"

func (m fakeMem) Peek8(a uint32) uint8    { return m[a] }
func (m fakeMem) Poke8(a uint32, v uint8) { m[a] = v }

// demoMem 放一份主程式簽章在線性 0x124A（任意位置），兩個遠呼叫點為 9A。
func demoMem(t *testing.T) (fakeMem, *DemoControl) {
	t.Helper()
	m := make(fakeMem, 0xA0000)
	copy(m[0x124A:], demoMainSignature.Bytes)
	m[0x124A+demoEnterOffset], m[0x124A+demoReturnOffset] = 0x9A, 0x9A
	d, err := NewDemoControl(m)
	if err != nil {
		t.Fatal(err)
	}
	return m, d
}

func TestDemoControlLocatesMainCalls(t *testing.T) {
	_, d := demoMem(t)
	if d.enter != 0x1360 || d.ret != 0x1365 {
		t.Fatalf("進入點 %05X、返回點 %05X，要 01360、01365", d.enter, d.ret)
	}
	m := make(fakeMem, 0xA0000)
	copy(m[0x124A:], demoMainSignature.Bytes)
	m[0x124A+demoReturnOffset] = 0x9A
	if _, err := NewDemoControl(m); err == nil {
		t.Fatal("進入點不是遠呼叫仍綁定")
	}
	m[0x124A+demoEnterOffset], m[0x124A+demoReturnOffset] = 0x9A, 0x80 // 本作返回點是 cmp
	if _, err := NewDemoControl(m); err != nil {
		t.Fatal(err)
	}
	m[0x124A+demoEnterOffset] = 0
	copy(m[0x5000:], demoMainSignature.Bytes)
	if _, err := NewDemoControl(m); err == nil {
		t.Fatal("簽章兩處命中仍綁定")
	}
}

func TestDemoRunningOnlyInsideDemo(t *testing.T) {
	m, d := demoMem(t)
	if d.Running() {
		t.Fatal("開機時不應在 demo")
	}
	d.Before(0, d.enter, m) // 67E9＝0：正式遊戲
	if d.Running() {
		t.Fatal("正式遊戲被當成 demo")
	}
	d.Before(0, d.ret, m)
	m[demoFlag] = 1
	d.Before(0, d.enter, m)
	if !d.Running() {
		t.Fatal("demo 進入點後應在 demo")
	}
	d.Before(0, d.ret, m)
	if d.Running() {
		t.Fatal("返回點後（標題、選單）不應在 demo")
	}
	var none *DemoControl
	if none.Running() {
		t.Fatal("未綁定時不應在 demo")
	}
}

func TestDemoBlocksMenuTimeout(t *testing.T) {
	m, d := demoMem(t)
	m[demoMenuTimeout], m[demoMenuTimeout+1] = 0xB8, 0x0B // 綁定前留下的 3000
	lo, hi := d.GuardRange(m)
	if lo != demoMenuTimeout || hi != demoMenuTimeout+3 {
		t.Fatalf("平時監看 %05X..%05X，要只有選單逾時", lo, hi)
	}
	for a := demoMenuTimeout; a <= demoMenuTimeout+3; a++ {
		if m[a] != 0 {
			t.Fatalf("%05X=%d，逾時應清為 0", a, m[a])
		}
	}
	// 主程式寫 8A63=3000（低位元組先寫）：兩個位元組都擋下，清零照放行。
	if !d.GuardWrite(demoMenuTimeout, 0, 0xB8) || !d.GuardWrite(demoMenuTimeout+1, 0, 0x0B) || d.GuardWrite(demoMenuTimeout, 0xB8, 0) {
		t.Fatal("選單逾時的非零寫入要擋、歸零要放行")
	}
	if d.GuardWrite(demoSideCounts, 0, 1) {
		t.Fatal("沒有中止時不應擋存活數")
	}
	if d.TimeoutsCleared != 1 || d.Err != nil {
		t.Fatalf("擋下 %d 次、錯誤 %v", d.TimeoutsCleared, d.Err)
	}
}

func TestDemoAbortWritesAndGuards(t *testing.T) {
	m, d := demoMem(t)
	d.RequestAbort() // 不在 demo：無效
	m[demoFlag] = 1
	d.Before(0, d.enter, m)
	m[demoSideCounts], m[demoSideCounts+1] = 3, 4
	m[bdaKeyHead], m[bdaKeyTail] = 0x24, 0x28
	if lo, _ := d.GuardRange(m); lo != demoMenuTimeout {
		t.Fatal("沒有要求時就開始中止")
	}
	d.RequestAbort()
	lo, hi := d.GuardRange(m)
	if lo != demoSideCounts || hi != demoMenuTimeout+3 {
		t.Fatalf("監看 %05X..%05X", lo, hi)
	}
	if m[demoECLStop] != 1 || m[demoSideCounts] != 0 || m[demoSideCounts+1] != 0 || m[bdaKeyTail] != 0x24 {
		t.Fatalf("59B4=%d 87CC=%d 87CD=%d 尾=%#x", m[demoECLStop], m[demoSideCounts], m[demoSideCounts+1], m[bdaKeyTail])
	}
	if !d.GuardWrite(demoSideCounts, 0, 1) || d.GuardWrite(demoSideCounts, 1, 0) {
		t.Fatal("中止期間要擋非 0 寫入、放行歸零")
	}
	m[demoECLStop] = 0 // ECL 迴圈進入時清掉
	d.GuardRange(m)
	if m[demoECLStop] != 1 {
		t.Fatal("中止期間每回合都要重寫停止旗標")
	}
	d.Before(0, d.ret, m)
	if d.Aborts != 1 || d.AbortReturns != 1 || d.Err != nil {
		t.Fatalf("中止 %d、返回 %d、錯誤 %v", d.Aborts, d.AbortReturns, d.Err)
	}
	if lo, _ := d.GuardRange(m); lo != demoMenuTimeout || d.GuardWrite(demoSideCounts, 0, 1) {
		t.Fatal("返回後仍在監看或還原存活數")
	}
}

func TestDemoRefusesUndeclaredAddress(t *testing.T) {
	m, d := demoMem(t)
	d.poke(m, 0x15295, 1)
	if d.Err == nil || m[0x15295] != 0 {
		t.Fatal("寫入了未宣告的位址")
	}
}
