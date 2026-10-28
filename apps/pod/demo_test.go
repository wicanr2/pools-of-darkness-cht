package pod

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func (m fakeMem) Peek8(a uint32) uint8    { return m[a] }
func (m fakeMem) Poke8(a uint32, v uint8) { m[a] = v }

func demoMem(t *testing.T) (fakeMem, *DemoControl) {
	t.Helper()
	m := make(fakeMem, 0xa0000)
	copy(m[0x124a:], demoMainSignature.Bytes)
	copy(m[0x124a+demoEnterOffset:], []byte{0x9a, 0x25, 0, 0xef, 0x16})
	copy(m[0x124a+demoReturnOffset:], []byte{0x9a, 0x2a, 0, 0x28, 0x17})
	raw, err := hex.DecodeString("8bdc368b57040bd27416b840008ec033ff268a058b1e82fb8bcbe806004a75f8ca0200263a057502e2f9c3")
	if err != nil {
		t.Fatal(err)
	}
	copy(m[0x4000:], raw)
	copy(m[0x5000:], demoSoundSignature.Bytes)
	copy(m[0x50a3:], []byte{0x89, 0xec, 0x5d, 0xca, 2, 0})
	d, err := NewDemoControl(m)
	if err != nil {
		t.Fatal(err)
	}
	return m, d
}

func TestDemoControlLocatesMainCalls(t *testing.T) {
	_, d := demoMem(t)
	if d.enter != 0x1310 || d.ret != 0x1315 {
		t.Fatalf("enter=%x ret=%x", d.enter, d.ret)
	}
	for _, at := range []uint32{d.enter, d.enter + 3, d.ret, 0x4009, 0x5006, 0x50a3} {
		m, _ := demoMem(t)
		m[at] ^= 1
		if _, err := NewDemoControl(m); err == nil {
			t.Fatalf("接受來源失配%x", at)
		}
	}
	m, _ := demoMem(t)
	copy(m[0x6000:], demoMainSignature.Bytes)
	if _, err := NewDemoControl(m); err == nil {
		t.Fatal("接受主簽章雙重命中")
	}
}

func TestDemoRunningOnlyInsideDemo(t *testing.T) {
	m, d := demoMem(t)
	var none *DemoControl
	if none.Running() || d.Running() {
		t.Fatal("開機被當成demo")
	}
	d.Before(0, d.enter, m)
	d.RequestAbort()
	before := append([]byte(nil), m...)
	if lo, hi := d.GuardRange(m); lo <= hi {
		t.Fatal("正式遊戲仍監看demo欄位")
	}
	if !bytes.Equal(before, m) || d.request || d.GuardWrite(demoMenuTimeout, 0, 1) {
		t.Fatal("正式遊戲被修改")
	}
	d.Before(0, d.ret, m)
	m[demoFlag] = 1
	d.Before(1, d.enter, m)
	if !d.Running() {
		t.Fatal("未辨識demo")
	}
	d.Before(2, d.ret, m)
	if d.Running() {
		t.Fatal("返回後仍在demo")
	}
}

func TestDemoBlocksMenuTimeout(t *testing.T) {
	m, d := demoMem(t)
	m[demoMenuTimeout] = 0xb8
	m[demoMenuTimeout+1] = 0xb
	lo, hi := d.GuardRange(m)
	if lo != demoMenuTimeout || hi != demoMenuTimeout+3 {
		t.Fatal(lo, hi)
	}
	for a := lo; a <= hi; a++ {
		if m[a] != 0 {
			t.Fatal("選單逾時未清除")
		}
	}
	if !d.GuardWrite(lo, 0, 1) || !d.GuardWrite(lo+1, 0, 1) || d.GuardWrite(lo, 1, 0) || d.GuardWrite(demoSideCounts, 0, 1) {
		t.Fatal("監看範圍錯誤")
	}
}

func TestDemoAbortWritesAndRestoresWaits(t *testing.T) {
	m, d := demoMem(t)
	m[demoFlag] = 1
	d.Before(100, d.enter, m)
	m[demoSideCounts], m[demoSideCounts+1] = 3, 4
	m[bdaKeyHead], m[bdaKeyTail], m[bdaKeyTail+1] = 0x24, 0x28, 1
	d.RequestAbort()
	d.RequestAbort()
	lo, hi := d.GuardRange(m)
	if lo != demoECLStop || hi != demoSideCounts+1 {
		t.Fatal(lo, hi)
	}
	if d.Aborts != 1 || m[demoECLStop] != 1 || m[demoGameStop] != 1 || m[demoSideCounts] != 0 || m[demoSideCounts+1] != 0 || m[bdaKeyTail] != 0x24 || m[bdaKeyTail+1] != 0 {
		t.Fatal("停止或佇列不符")
	}
	for _, p := range d.patches {
		if !demoBytesMatch(m.Read8, p.at, p.after) {
			t.Fatalf("等待補丁未套入%x", p.at)
		}
	}
	d.RequestAbort()
	d.GuardRange(m)
	if d.Aborts != 1 {
		t.Fatal("重複按鍵重啟中斷")
	}
	if !d.GuardWrite(demoSideCounts, 0, 2) || d.GuardWrite(demoSideCounts, 2, 0) {
		t.Fatal("存活數守衛錯誤")
	}
	d.Before(150, d.ret, m)
	d.GuardRange(m)
	if d.AbortReturns != 1 || d.LastAbortSteps != 50 || d.patched || d.Err != nil {
		t.Fatalf("返回／還原不符%+v", d)
	}
	for _, p := range d.patches {
		if !demoBytesMatch(m.Read8, p.at, p.before) {
			t.Fatalf("未恢復原始碼%x", p.at)
		}
	}
	if d.GuardWrite(demoSideCounts, 0, 1) {
		t.Fatal("返回後仍守住存活數")
	}
	if _, err := NewDemoControl(m); err != nil {
		t.Fatalf("還原後不能重新辨識%v", err)
	}
}

func TestDemoPatchMismatchDoesNotPartiallyPatch(t *testing.T) {
	m, d := demoMem(t)
	m[demoFlag] = 1
	d.Before(1, d.enter, m)
	bad := d.patches[2].at
	m[bad] ^= 1
	before := append([]byte(nil), m[:0x6000]...)
	d.RequestAbort()
	d.GuardRange(m)
	// BIOS queue clearing is outside this code-only comparison.
	copy(before[bdaKeyTail:bdaKeyTail+2], m[bdaKeyTail:bdaKeyTail+2])
	if d.Err == nil || d.patched || !bytes.Equal(before, m[:0x6000]) {
		t.Fatal("失配仍部分修改程式碼")
	}
}

func TestDemoRefusesUndeclaredAddress(t *testing.T) {
	m, d := demoMem(t)
	d.poke(m, 0x15295, 1)
	if d.Err == nil || m[0x15295] != 0 {
		t.Fatal("寫入未宣告位址")
	}
	m, d = demoMem(t)
	d.poke(m, d.patches[0].at, 0x42)
	if d.Err == nil || m[d.patches[0].at] != 0x26 {
		t.Fatal("接受未宣告的程式碼byte")
	}
}
