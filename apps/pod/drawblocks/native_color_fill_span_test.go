package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func colorFillSeed(mem []byte, color, count, offset, segment uint16, dos bool) *view {
	v := &view{mem: append([]byte(nil), mem...), dos: dos, s: machine.HostRegisters{R: [8]uint16{0xab77, 23, 45, 12, 0xe000, 0xd000, 0x440, 0x880}, Seg: [4]uint16{0x6000, 0x1ed4, 0x8000, 0x6000}, IP: 0x1a7f, Flags: 0xfed7}}
	for off, value := range map[uint32]uint16{0: 0x1234, 2: 0x4567, 4: color, 6: count, 8: offset, 10: segment} {
		putRowWord(v, 0x8e000+off, value)
	}
	return v
}
func colorFillPixel(t *testing.T, v *view) *view {
	m := callMachine(v)
	for i := 0; i < 5; i++ {
		if e := m.Step(); e != nil {
			t.Fatal(e)
		}
	}
	return &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: v.dos}
}
func TestColorFillSpanOriginal(t *testing.T) {
	mem := input(t)
	n, whole, g := 0, 0, 0
	for _, c := range []uint16{0, 15, 128, 0xab2b, 0xffff} {
		for _, dest := range [][3]uint16{{0, 0, 0xa000}, {1, 63999, 0xa000}, {128, 42104, 0xa000}, {257, 16, 0x9fff}, {64000, 0, 0xa000}} {
			for _, dos := range []bool{false, true} {
				v := colorFillSeed(mem, c, dest[0], dest[1], dest[2], dos)
				w := colorFillPixel(t, v)
				p := planColorFillSpan(w, 1)
				if p == nil {
					t.Fatal("valid span refused", c, dest, dos)
				}
				compare(t, w, p)
				n++
				for _, budget := range []uint64{1, 2, 6, 7, 1024} {
					a, b := callMachine(v), callMachine(v)
					runCall(t, a, v, false)
					for b.Steps < 1000 {
						if b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
							break
						}
						w := paletteMachineView{b}
						p := Plan(w, budget)
						if p != nil {
							if colorFillFixedPixels(p.Writes, p.Transfers) && !p.Algorithmic {
								t.Fatal("fixed VGA pixels")
							}
							if p.Algorithmic {
								g++
							}
							if e := b.CommitHostBlock(p); e != nil {
								t.Fatal(e)
							}
						} else {
							if w.CS() == 0x1ed4 && w.IP() == 0x1a8e {
								t.Fatal("CPU VGA span")
							}
							if e := b.Step(); e != nil {
								t.Fatal(e)
							}
						}
					}
					if b.CPU.Seg[cpu.CS] != 0x4567 || b.CPU.IP != 0x1234 {
						t.Fatal("no return")
					}
					paletteMachinesEqual(t, a, b)
					whole++
				}
			}
		}
	}
	t.Logf("color-fill ordered original spans=%d whole SaveState+PortLog=%d Go plans=%d CPU/fixed VGA pixels=0", n, whole, g)
}
func TestColorFillSpanRefusals(t *testing.T) {
	mem := input(t)
	changes := []func(*view){func(v *view) { v.mem[0x449] = 0xd }, func(v *view) { v.mem[0x207bf] ^= 1 }, func(v *view) { v.s.Seg[cpu.CS]++ }, func(v *view) { v.s.IP++ }, func(v *view) { v.s.R[cpu.BX]++ }, func(v *view) { v.s.R[cpu.AX]++ }, func(v *view) { v.s.R[cpu.CX]++ }, func(v *view) { v.s.R[cpu.DI]++ }, func(v *view) { v.s.Seg[cpu.ES]++ }, func(v *view) { v.s.Flags |= cpu.DF }, func(v *view) { v.s.R[cpu.SP] = 0xfffa; v.s.R[cpu.BX] = 0xfffa }, func(v *view) { v.s.Seg[cpu.SS] = 0xa000 }, func(v *view) { putRowWord(v, 0x8e00a, 0x6000); v.s.Seg[cpu.ES] = 0x6000 }, func(v *view) { putRowWord(v, 0x8e008, 63999); v.s.R[cpu.DI] = 63999 }, func(v *view) { putRowWord(v, 0x8e008, 65535); v.s.R[cpu.DI] = 65535 }}
	for _, change := range changes {
		v := colorFillPixel(t, colorFillSeed(mem, 0xab2b, 128, 42104, 0xa000, false))
		change(v)
		before := append([]byte(nil), v.mem...)
		state := v.s
		if planColorFillSpan(v, 1) != nil {
			t.Fatal("bad span accepted")
		}
		if !bytes.Equal(before, v.mem) || state != v.s {
			t.Fatal("refusal mutated input")
		}
	}
	v := colorFillPixel(t, colorFillSeed(mem, 0xab2b, 128, 42104, 0xa000, false))
	if planColorFillSpan(v, 0) != nil {
		t.Fatal("zero budget")
	}
	for _, w := range []CPUView{primitiveUnknownModel{v}, primitive8086Model{v}, struct{ CPUView }{v}} {
		if planColorFillSpan(w, 1) != nil {
			t.Fatal("unverified model accepted")
		}
	}
	t.Logf("color-fill unchanged refusals=%d", len(changes)+4)
}
