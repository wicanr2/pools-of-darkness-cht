// Spec025/#49: original glyph output and complete-machine comparisons.
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func glyph5aa4WholeBoundary(t *testing.T, v *view, budget uint64) int {
	t.Helper()
	a, b := callMachine(v), callMachine(v)
	runCall(t, a, v, false)
	plans := 0
	for b.Steps < 1000000 {
		if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
			break
		}
		w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: v.dos}
		p := Plan(w, budget)
		if p != nil {
			compare(t, w, p)
			if !p.Algorithmic && glyph5aa4FixedPixels(p.Writes, p.Transfers) {
				t.Fatal("fixed boundary glyph pixel")
			}
			if p.Algorithmic {
				plans++
			}
			if e := b.CommitHostBlock(p); e != nil {
				t.Fatal(e)
			}
		} else {
			if w.CS() == 0x48a1 && (w.IP() == 0x5c02 || w.IP() == 0x5c2c) {
				t.Fatalf("CPU boundary glyph pixel at%x", w.IP())
			}
			if e := b.Step(); e != nil {
				t.Fatal(e)
			}
		}
	}
	var sa, sb bytes.Buffer
	if e := a.SaveState(&sa); e != nil {
		t.Fatal(e)
	}
	if e := b.SaveState(&sb); e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
		t.Fatalf("glyph boundary SaveState differs steps%d/%d", a.Steps, b.Steps)
	}
	return plans
}
func TestNativeGlyph5AA4Boundaries(t *testing.T) {
	cases, plans := 0, 0
	for _, kind := range []string{"corner", "full", "eight-parents", "source-end", "entry-DF"} {
		for _, mapped := range []bool{false, true} {
			for _, dos := range []bool{false, true} {
				width, height := uint16(7), uint16(2)
				if kind == "corner" {
					width, height = 1, 1
				}
				if kind == "full" {
					width, height = 320, 200
				}
				v := glyph5aa4Seed(t, width, height, mapped)
				v.dos = dos
				if kind == "entry-DF" {
					v.s.Flags |= cpu.DF
				}
				if kind == "source-end" {
					copy(v.mem[0x7fff0:0x80000], v.mem[0x70500:0x70510])
					putRowWord(v, 0x70100+0x108+2*65, 0xfef0)
				}
				if kind == "full" || kind == "corner" {
					for _, index := range []uint16{0, 4} {
						for off, value := range map[uint16]uint16{0xbd: 0xa000, 0x2bd: 0, 0x4bd: 0, 0x6bd: 0, 0x8bd: 319, 0xabd: 199, 0xcbd: 0x80} {
							putRowWord(v, 0x48a10+uint32(off+index), value)
						}
					}
					x, y := uint16(0), uint16(0)
					if kind == "corner" {
						x, y = 319, 199
					}
					putRowWord(v, 0x60102, x)
					putRowWord(v, 0x60104, y)
				}
				if kind == "eight-parents" {
					for i := uint16(0); i <= 8; i++ {
						seg, flags := uint16(0xa000), uint16(0x80)
						if i < 8 {
							seg, flags = (i+1)*2, 0xc0
						}
						for off, value := range map[uint16]uint16{0xbd: seg, 0x2bd: 0, 0x4bd: 0, 0x6bd: 0, 0x8bd: 319, 0xabd: 199, 0xcbd: flags} {
							putRowWord(v, 0x48a10+uint32(off+i*2), value)
						}
					}
				}
				t.Logf("checking kind=%s mapped=%v dos=%v", kind, mapped, dos)
				plans += glyph5aa4WholeBoundary(t, v, 1024)
				cases++
			}
		}
	}
	t.Logf("glyph boundary/full canvas/8 parents/segment end/DF SaveState cases=%d Go plans=%d", cases, plans)
}

func glyph5aa4FirstSpan(t *testing.T, mapped bool) *view {
	t.Helper()
	v := glyph5aa4Seed(t, 7, 3, mapped)
	m := original(v)
	entry := uint16(0x5c02)
	if mapped {
		entry = 0x5c2c
	}
	for m.Steps < 10000 {
		if m.CPU.Seg[cpu.CS] == 0x48a1 && m.CPU.IP == entry {
			return &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}}
		}
		if e := m.Step(); e != nil {
			t.Fatal(e)
		}
	}
	t.Fatal("no glyph span")
	return nil
}
func TestNativeGlyph5AA4Refusals(t *testing.T) {
	cases := 0
	for _, mapped := range []bool{false, true} {
		for _, bad := range []string{"budget", "code", "mode", "CS", "ES", "DF", "stack-wrap", "stack-driver", "context-driver", "context-stack", "font-driver", "font-stack", "font-segment", "char", "width-zero", "height-zero", "remaining", "slot", "parent-cycle", "surface", "context-x", "context-font", "CX", "DX", "BX", "SI", "DI"} {
			v := glyph5aa4FirstSpan(t, mapped)
			if planGlyph5aa4Spans(v, 1) == nil {
				t.Fatal("good span refused")
			}
			budget := uint64(1)
			word := func(off, value uint16) { putRowWord(v, 0x48a10+uint32(off), value) }
			switch bad {
			case "budget":
				budget = 0
			case "code":
				v.mem[0x48a10+0x5aa4] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "CS":
				v.s.Seg[cpu.CS] = 0x48a0
			case "ES":
				v.s.Seg[cpu.ES] = 0xb000
			case "DF":
				v.s.Flags |= cpu.DF
			case "stack-wrap":
				v.s.R[cpu.SP] = 0xfffa
			case "stack-driver":
				v.s.Seg[cpu.SS] = 0x48a1
				v.s.R[cpu.SP] = 0x1100
			case "context-driver":
				word(0x110b, 0x1100)
				word(0x110d, 0x48a1)
			case "context-stack":
				word(0x110b, v.SP())
				word(0x110d, v.SS())
			case "font-driver":
				word(0x1117, 0x48a1)
				putRowWord(v, 0x60108, 0x48a1)
			case "font-stack":
				word(0x1115, v.SP())
				word(0x1117, v.SS())
				putRowWord(v, 0x60106, v.SP())
				putRowWord(v, 0x60108, v.SS())
			case "font-segment":
				v.s.Seg[cpu.DS]++
			case "char":
				putRowWord(v, uint32(v.SS())*16+uint32(v.SP())+12, 256)
			case "width-zero":
				putRowWord(v, 0x70500, 0)
			case "height-zero":
				putRowWord(v, 0x70102, 0)
			case "remaining":
				word(0x11ff, 0)
			case "slot":
				word(0x110f, 256)
				putRowWord(v, 0x60100, 256)
			case "parent-cycle":
				word(0xbd, 0)
			case "surface":
				word(0xbd+4, 0xb000)
			case "context-x":
				word(0x1111, 319)
				putRowWord(v, 0x60102, 319)
			case "context-font":
				putRowWord(v, 0x60106, 0x101)
			case "CX":
				v.s.R[cpu.CX] = 0
			case "DX":
				v.s.R[cpu.DX]++
			case "BX":
				v.s.R[cpu.BX]++
			case "SI":
				v.s.R[cpu.SI]++
			case "DI":
				v.s.R[cpu.DI]++
			}
			before := append([]byte(nil), v.mem...)
			state := v.s
			if p := planGlyph5aa4Spans(v, budget); p != nil {
				t.Fatalf("bad glyph accepted %s mapped%v", bad, mapped)
			}
			if state != v.s || !bytes.Equal(before, v.mem) {
				t.Fatal("refusal changed input")
			}
			cases++
		}
	}
	t.Logf("glyph unproved/alias states refused=%d unchanged inputs", cases)
}
