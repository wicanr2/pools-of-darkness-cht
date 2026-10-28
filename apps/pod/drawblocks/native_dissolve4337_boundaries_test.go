// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func runDissolve4337Original(t *testing.T, m *machine.Machine) {
	t.Helper()
	m.RunLimit = 12000000
	for m.Steps < 12000000 {
		if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
			return
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("dissolve original bounded call did not return")
}
func runDissolve4337Native(t *testing.T, m *machine.Machine, v *view, budget uint64) (plans, waits, metadata int) {
	t.Helper()
	m.RunLimit = 12000000
	for m.Steps < 12000000 {
		if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
			return
		}
		w := &view{mem: m.Mem, s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: v.dos}
		p := dissolve4337(w, budget)
		if p != nil {
			plans++
		} else {
			snapshot, known := m.VGAStatusSnapshot()
			p = planWait(dissolve4337WaitView{w, snapshot, known}, budget)
			if p != nil {
				waits++
			}
		}
		if p != nil {
			if err := m.CommitHostBlock(p); err != nil {
				t.Fatalf("%v at%x steps%d N%d C%d allowed%d flags%x halted%v IRQevery%d frameEvery%d", err, w.IP(), m.Steps, p.Instructions, p.Cycles, m.HostWorkBudget(p.Instructions), m.CPU.Flags, m.CPU.Halted, m.IRQ0Every, m.VGAFrameEvery)
			}
		} else {
			if w.CS() == 0x48a1 && w.IP() != 0x45ca && w.IP() != 0x45cb {
				t.Fatalf("dissolve bounded native fell to CPU IP%x", w.IP())
			}
			if err := m.Step(); err != nil {
				t.Fatal(err)
			}
			metadata++
		}
	}
	t.Fatal("dissolve native bounded call did not return")
	return
}
func dissolve4337BoundarySeed(t *testing.T, box [8]uint16, parents, df, dos bool) *view {
	v := dissolve4337Seed(t, box[2]-box[0]+1, box[3]-box[1]+1)
	v.dos = dos
	if df {
		v.s.Flags |= cpu.DF
	}
	for i, off := range []uint16{0x4bd, 0x6bd, 0x8bd, 0xabd} {
		putRowWord(v, 0x48a10+uint32(off), box[i])
		putRowWord(v, 0x48a10+uint32(off+2), box[4+i])
	}
	if parents {
		for slot := uint16(2); slot <= 9; slot++ {
			seg, flags := uint16(0xa000), uint16(0x80)
			if slot < 9 {
				seg, flags = (slot+1)*2, 0xc0
			}
			for _, w := range [][2]uint16{{0xbd, seg}, {0x2bd, 0}, {0x4bd, 0}, {0x6bd, 0}, {0x8bd, 319}, {0xabd, 199}, {0xcbd, flags}} {
				putRowWord(v, 0x48a10+uint32(w[0]+slot*2), w[1])
			}
		}
	}
	return v
}
func TestNativeDissolve4337Boundaries(t *testing.T) {
	cases, plans, waits, metadata := 0, 0, 0, 0
	for _, box := range [][8]uint16{{0, 0, 319, 199, 0, 0, 319, 199}, {319, 198, 319, 199, 0, 0, 0, 1}, {0, 0, 1, 0, 318, 199, 319, 199}, {10, 10, 26, 18, 303, 191, 319, 199}} {
		for _, parents := range []bool{false, true} {
			for _, df := range []bool{false, true} {
				for _, dos := range []bool{false, true} {
					v := dissolve4337BoundarySeed(t, box, parents, df, dos)
					a, b := callMachine(v), callMachine(v)
					runDissolve4337Original(t, a)
					np, nw, nm := runDissolve4337Native(t, b, v, 4096)
					var sa, sb bytes.Buffer
					if err := a.SaveState(&sa); err != nil {
						t.Fatal(err)
					}
					if err := b.SaveState(&sb); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
						t.Fatalf("dissolve boundary differs box%v parents%v df%v dos%v", box, parents, df, dos)
					}
					cases++
					plans += np
					waits += nw
					metadata += nm
				}
			}
		}
	}
	t.Log("dissolve4337 boundary parent machine comparisons", cases, "algorithmic plans", plans, "shared wait plans", waits, "CPU control steps", metadata)
}
