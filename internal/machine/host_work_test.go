package machine

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func TestHostCallTargetsAreOwnedAndScoped(t *testing.T) {
	m := New()
	m.IRQ0Every = 0
	m.VGAFrameEvery = 0
	m.CPU.Seg = [4]uint16{0, 0x1000, 0x8000, 0x1000}
	m.CPU.R[cpu.SP] = 0x1000
	m.CPU.IP = 0x100
	copy(m.Mem[0x10100:], []byte{0xe8, 0xfd, 0})
	calls := 0
	m.ObserveHostCalls(func() { calls++ })
	targets := []uint32{0x10200}
	m.SetHostCallTargets(targets)
	targets[0] = 0
	if e := m.Step(); e != nil {
		t.Fatal(e)
	}
	if calls != 1 {
		t.Fatal("target slice aliased")
	}
	m.CPU.IP = 0x100
	m.SetHostCallTargets([]uint32{0x10201})
	if e := m.Step(); e != nil {
		t.Fatal(e)
	}
	if calls != 1 {
		t.Fatal("unrelated CALL reached provider")
	}
	m.ObserveHostCalls(nil)
	if m.CPU.HostCalls || len(m.hostCallTargets) != 0 {
		t.Fatal("callback retained")
	}
}

func TestHostWorkBudgetAndRefusals(t *testing.T) {
	m := New()
	m.Steps = 100
	m.IRQ0Every = 50
	m.nextIRQ0 = 137
	m.VGAFrameEvery = 100
	m.nextFrame = 190
	if m.HostWorkBudget(1000) != 37 {
		t.Fatal("IRQ")
	}
	m.periodic.on = true
	m.periodic.next = 121
	if m.HostWorkBudget(1000) != 21 {
		t.Fatal("periodic")
	}
	m.nextFrame = 115
	if m.HostWorkBudget(1000) != 15 {
		t.Fatal("frame")
	}
	m.RunLimit = 109
	if m.HostWorkBudget(1000) != 9 {
		t.Fatal("run limit")
	}
	cases := []func(*Machine){func(m *Machine) { m.CycleClock = true }, func(m *Machine) { m.CPU.Halted = true }, func(m *Machine) { m.CPU.Flags |= cpu.TF }, func(m *Machine) { m.irq0Pending = true }, func(m *Machine) { m.cbQueue = []QueuedCall{{}} }, func(m *Machine) { m.keyQueue = []KeyEvent{{}} }, func(m *Machine) { m.timedKeys = []timedKey{{}} }, func(m *Machine) { m.wordWatches = []wordWatch{{}} }, func(m *Machine) { m.breaks = map[int]uint32{1: 1} }, func(m *Machine) { m.Coverage = make([]bool, 1) }, func(m *Machine) { m.TraceSegs = true }, func(m *Machine) { m.WatchDSOn = true }, func(m *Machine) { m.onRead = func(uint32, byte) {} }, func(m *Machine) { m.accelCredit = 1 }}
	for i, change := range cases {
		m := New()
		m.IRQ0Every = 0
		m.VGAFrameEvery = 0
		change(m)
		if m.HostWorkBudget(1) != 0 {
			t.Fatal("unsupported mode", i)
		}
	}
}
func TestHostCallsPreserveRunnerBudgetAndEvents(t *testing.T) {
	for _, kind := range []string{"near", "far", "indirect-near", "indirect-far"} {
		for _, budget := range []uint64{3, 4, 5} {
			for _, frameAt := range []uint64{0, 2, 4} {
				for _, dosCost := range []bool{false, true} {
					fixture := func() *Machine {
						m := New()
						m.IRQ0Every = 0
						m.VGAFrameEvery = frameAt
						m.nextFrame = frameAt
						m.RunLimit = budget
						m.CPU.Seg = [4]uint16{0, 0x1000, 0x8000, 0x1000}
						m.CPU.R[cpu.SP] = 0x1000
						m.CPU.R[cpu.AX] = 0x200
						m.CPU.IP = 0x100
						m.CPU.DOSBoxCost = dosCost
						for i := 0; i < 32; i++ {
							m.Mem[0x10100+i] = 0x90
						}
						ret := byte(0xc3)
						switch kind {
						case "near":
							copy(m.Mem[0x10100:], []byte{0xe8, 0xfd, 0})
						case "far":
							copy(m.Mem[0x10100:], []byte{0x9a, 0, 2, 0, 0x10})
							ret = 0xcb
						case "indirect-near":
							copy(m.Mem[0x10100:], []byte{0xff, 0xd0})
						case "indirect-far":
							copy(m.Mem[0x10100:], []byte{0xff, 0x1e, 0, 4})
							copy(m.Mem[0x10400:], []byte{0, 2, 0, 0x10})
							ret = 0xcb
						}
						copy(m.Mem[0x10200:], []byte{0x90, 0x90, ret})
						return m
					}
					a, b := fixture(), fixture()
					var af, bf []uint64
					a.SetOnFrame(func() { af = append(af, a.Steps) })
					b.SetOnFrame(func() { bf = append(bf, b.Steps) })
					b.ObserveHostCalls(func() {
						if b.CPU.IP != 0x200 {
							return
						}
						after := HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.Read16(cpu.Linear(b.CPU.Seg[cpu.SS], b.CPU.R[cpu.SP])), Flags: b.CPU.Flags}
						cycles := uint64(16)
						after.R[cpu.SP] += 2
						if kind == "far" || kind == "indirect-far" {
							after.Seg[cpu.CS] = b.Read16(cpu.Linear(b.CPU.Seg[cpu.SS], b.CPU.R[cpu.SP]+2))
							after.R[cpu.SP] += 2
							cycles = 28
						}
						if dosCost {
							cycles = 3
						}
						p := &HostWorkPlan{After: after, Instructions: 3, Cycles: cycles, LastCS: 0x1000, LastIP: 0x202}
						if b.ValidateHostWork(p) == nil {
							if e := b.CommitHostWork(p); e != nil {
								t.Fatal(e)
							}
						}
					})
					_, ea := a.RunUntilObserved(nil, budget, nil)
					_, eb := b.RunUntilObserved(nil, budget, nil)
					if ea != nil || eb != nil {
						t.Fatal(ea, eb)
					}
					if a.Steps != b.Steps || a.CPU.R != b.CPU.R || a.CPU.Seg != b.CPU.Seg || a.CPU.IP != b.CPU.IP || a.CPU.Flags != b.CPU.Flags || a.CPU.Cycles != b.CPU.Cycles || !bytes.Equal(a.Mem, b.Mem) || len(af) != len(bf) {
						t.Fatalf("CALL mismatch %s budget%d event%d dos%v", kind, budget, frameAt, dosCost)
					}
					for i := range af {
						if af[i] != bf[i] {
							t.Fatal("frame time")
						}
					}
					if budget < 4 || frameAt == 2 {
						if b.HostCallCommits != 0 {
							t.Fatal("crossed boundary")
						}
					} else if b.HostCallCommits != 1 {
						t.Fatal("eligible call was not used")
					}
				}
			}
		}
	}
}
func TestHostWorkAtomicValidationAndContext(t *testing.T) {
	m := New()
	m.SetVideoMode(0x13)
	m.IRQ0Every = 0
	m.VGAFrameEvery = 0
	m.SetNativeVideo(true)
	m.Steps = 10
	p := &HostWorkPlan{After: HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, Instructions: 3, Cycles: 6, LastCS: 0x1000, LastIP: 0x202, Writes: []HostWrite{{Address: 0xa0010, Value: 7, Step: 2, CS: 0x1000, IP: 0x200}}}
	var writes []VideoWrite
	var commands []VideoCommand
	m.ObserveVideoWrites(func(w VideoWrite) { writes = append(writes, w) })
	m.ObserveVideoCommands(func(c VideoCommand) { commands = append(commands, c) })
	for _, which := range []string{"zero", "order", "BDA", "outside", "time", "watch", "mode", "overflow"} {
		q := *p
		q.Writes = append([]HostWrite(nil), p.Writes...)
		switch which {
		case "zero":
			q.Instructions = 0
		case "order":
			q.Writes = append(q.Writes, HostWrite{Address: 0x1000, Step: 1})
		case "BDA":
			q.Writes[0].Address = 0x449
		case "outside":
			q.Writes[0].Address = 0xaffff
		case "time":
			q.Writes[0].Step = 4
		case "watch":
			m.WatchWrites(0xa0000, 0xaffff, func(uint32, byte, byte) {})
		case "mode":
			m.planarOn = true
		case "overflow":
			q.Cycles = ^uint64(0)
			m.CPU.Cycles = 1
		}
		before := append([]byte(nil), m.Mem...)
		steps := m.Steps
		if m.CommitHostWork(&q) == nil || !bytes.Equal(before, m.Mem) || m.Steps != steps || len(writes) != 0 {
			t.Fatal("invalid plan committed", which)
		}
		m.WatchWrites(0, 0, nil)
		m.planarOn = false
		m.CPU.Cycles = 0
	}
	if e := m.CommitHostWork(p); e != nil {
		t.Fatal(e)
	}
	if len(writes) != 1 || writes[0].Step != 12 || writes[0].CS != 0x1000 || writes[0].IP != 0x200 || len(commands) != 1 || !commands[0].Native || commands[0].Step != 12 || m.hostEffect != nil {
		t.Fatal("semantic write context", writes, commands)
	}
}
