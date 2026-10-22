package machine

import (
	"crypto/sha256"
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
)

// delayProgram：ES:DI＝0040:0000、AL＝該位元組，反覆以 CX＝FFFF 呼叫 DelayLoop（0113h）。
var delayProgram = []byte{
	0xB8, 0x40, 0x00, // 0100 mov ax,40h
	0x8E, 0xC0, // 0103 mov es,ax
	0x33, 0xFF, // 0105 xor di,di
	0x26, 0x8A, 0x05, // 0107 mov al,es:[di]
	0xB9, 0xFF, 0xFF, // 010A mov cx,FFFFh
	0xE8, 0x03, 0x00, // 010D call 0113h
	0xEB, 0xF8, // 0110 jmp 010Ah
	0x90,                                                                   // 0112
	0xBB, 0x04, 0x00, 0x4B, 0x75, 0xFD, 0x26, 0x3A, 0x05, 0xE1, 0xF5, 0xC3, // 0113 DelayLoop
}

type delayState struct {
	regs                 [8]uint16
	segs                 [4]uint16
	ip, flags            uint16
	steps, cycles, ticks uint64
	mem                  [32]byte
}

func stateOf(m *Machine) delayState {
	c := m.CPU
	return delayState{regs: c.R, segs: c.Seg, ip: c.IP, flags: c.Flags, steps: m.Steps, cycles: c.Cycles,
		ticks: m.Ticks, mem: sha256.Sum256(m.Mem[:])}
}

func delayMachines(t *testing.T) (plain, fast *Machine, loop uint32) {
	t.Helper()
	plain, fast = tinyProgram(t, delayProgram...), tinyProgram(t, delayProgram...)
	base := uint32(fast.CPU.Seg[cpu.CS]) << 4
	fast.Write16(base+0x200, 1000) // 假的 DelayCnt
	fast.SetTPDelay(base+0x113, base+0x200)
	return plain, fast, base + 0x113
}

func runChunk(t *testing.T, m *Machine, n uint64) {
	t.Helper()
	m.RunLimit = m.Steps + n
	if _, err := m.RunUntil(nil, n); err != nil {
		t.Fatal(err)
	}
	m.RunLimit = 0
}

func TestTPDelayAccelerationIsExact(t *testing.T) {
	plain, fast, _ := delayMachines(t)
	plain.Write16(uint32(plain.CPU.Seg[cpu.CS])<<4+0x200, 1000)
	for i, n := range []uint64{5, 37, 100003, 250000, 999, 1500000, 77777, 2000000} {
		runChunk(t, plain, n)
		runChunk(t, fast, n)
		if a, b := stateOf(plain), stateOf(fast); a != b {
			t.Fatalf("第 %d 段後狀態不同\n逐條 %+v\n快轉 %+v", i, a, b)
		}
	}
	if fast.DelayAccelerated == 0 || plain.Ticks == 0 {
		t.Fatalf("快轉 %d 步、計時中斷 %d 次，測試沒有涵蓋快轉跨中斷", fast.DelayAccelerated, plain.Ticks)
	}
}

func TestTPDelayNeedsEqualByteAndRunLimit(t *testing.T) {
	_, fast, _ := delayMachines(t)
	// 0107 改成 mov al,77h：AL 與 ES:DI 不同，迴圈第一圈就結束，不能快轉
	fast.WriteBytes(uint32(fast.CPU.Seg[cpu.CS])<<4+0x107, []byte{0xB0, 0x77, 0x90})
	fast.Write8(0x400, 0x5A)
	if _, err := fast.RunUntil(nil, 50000); err != nil {
		t.Fatal(err)
	}
	if fast.DelayAccelerated != 0 {
		t.Fatal("沒設 RunLimit 仍快轉")
	}
	runChunk(t, fast, 50000)
	if fast.DelayAccelerated != 0 {
		t.Fatal("ES:DI 不等於 AL 仍快轉")
	}
}

func TestSetClockPercentLive(t *testing.T) {
	m := New()
	if err := m.SetClockPercent(50); err != nil {
		t.Fatal(err)
	}
	m.Steps = 1000
	m.nextIRQ0, m.nextFrame = 1000+40000, 1000+10000
	m.SetTPDelay(0x500, 0x600) // DelayCnt 位址
	m.Write16(0x600, 525)
	m.SetTPDelay(0x500, 0x600)
	if err := m.SetClockPercentLive(200); err != nil {
		t.Fatal(err)
	}
	if m.ClockPercent() != 200 || m.nextIRQ0 != 1000+160000 || m.nextFrame != 1000+40000 ||
		m.VGAFrameEvery != ScaleSteps(DefaultVGAFrameEvery, 200) || m.IRQ0Base != ScaleSteps(DefaultIRQ0Every, 200) || m.nextKey != 0 {
		t.Fatalf("200%%：%d %d %d %d %d", m.nextIRQ0, m.nextFrame, m.VGAFrameEvery, m.IRQ0Base, m.nextKey)
	}
	if got := m.Read16(0x600); got != 2100 {
		t.Fatalf("DelayCnt %d，要 525×4＝2100", got)
	}
	for _, p := range []int{150, 333, 50} { // 依登記值重算，不累積誤差
		if err := m.SetClockPercentLive(p); err != nil {
			t.Fatal(err)
		}
	}
	if got := m.Read16(0x600); got != 525 {
		t.Fatalf("回到 50%% 後 DelayCnt %d，要 525", got)
	}
	before := clockFields(m)
	if err := m.SetClockPercentLive(50); err != nil || !eqU(clockFields(m), before) {
		t.Fatal("比例不變時改了欄位")
	}
	for _, p := range []int{9, 1601} {
		if m.SetClockPercentLive(p) == nil {
			t.Fatalf("p=%d 應失敗", p)
		}
	}
	m.CycleClock = true
	if m.SetClockPercentLive(100) == nil {
		t.Fatal("CycleClock 應失敗")
	}
}

// delayProgramTP5：ES:DI＝0000:0000、AL＝該位元組，反覆以 CX＝FFFF 呼叫 TP5 寫法的 DelayLoop（0113h）。
var delayProgramTP5 = []byte{
	0x33, 0xFF, // 0100 xor di,di
	0x8E, 0xC7, // 0102 mov es,di
	0x90, 0x90, 0x90, // 0104
	0x26, 0x8A, 0x05, // 0107 mov al,es:[di]
	0xB9, 0xFF, 0xFF, // 010A mov cx,FFFFh
	0xE8, 0x03, 0x00, // 010D call 0113h
	0xEB, 0xF8, // 0110 jmp 010Ah
	0x90,                                           // 0112
	0x26, 0x3A, 0x05, 0x75, 0x02, 0xE2, 0xF9, 0xC3, // 0113 DelayLoop（TP5）
}

func TestTPDelayAccelerationIsExactTP5(t *testing.T) {
	plain, fast := tinyProgram(t, delayProgramTP5...), tinyProgram(t, delayProgramTP5...)
	base := uint32(fast.CPU.Seg[cpu.CS]) << 4
	for _, m := range []*Machine{plain, fast} {
		m.Write16(base+0x200, 1000)
	}
	fast.SetTPDelay(base+0x113, base+0x200)
	if fast.tpd.form == nil || fast.tpd.form.perLoop != 3 {
		t.Fatal("沒有認出 TP5 寫法")
	}
	for i, n := range []uint64{5, 37, 100003, 250000, 999, 1500000, 77777, 2000000} {
		runChunk(t, plain, n)
		runChunk(t, fast, n)
		if a, b := stateOf(plain), stateOf(fast); a != b {
			t.Fatalf("第 %d 段後狀態不同\n逐條 %+v\n快轉 %+v", i, a, b)
		}
	}
	if fast.DelayAccelerated == 0 || plain.Ticks == 0 {
		t.Fatalf("快轉 %d 步、計時中斷 %d 次，測試沒有涵蓋快轉跨中斷", fast.DelayAccelerated, plain.Ticks)
	}
}
