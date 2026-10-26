package drawblocks

import (
	"bytes"
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func imageRowSeed(t *testing.T) *view {
	v := imageSpanSeed(t)
	v.s.IP = 0x6850
	putRowWord(v, 0x7080a, 0)
	for off, n := range map[uint32]uint16{0xec1: 0, 0xec3: 0, 0xec7: 199, 0xed3: 0, 0xed5: 0, 0xedd: 320, 0x1250: 0xa55a, 0x120c: 0x55aa} {
		putRowWord(v, 0x48a10+off, n)
	}
	v.mem[0x50100] = 10
	return v
}

func putRowWord(v *view, address uint32, n uint16) {
	v.mem[address], v.mem[address+1] = byte(n), byte(n>>8)
}

func TestNativeImageRowOriginal(t *testing.T) {
	plans := 0
	for _, stride := range []uint16{1, 16, 255, 319, 320} {
		for _, y := range []byte{0, 1, 127, 198, 199, 200, 254, 255} {
			for _, x := range []uint16{0, stride - 1} {
				for _, dos := range []bool{false, true} {
					v := imageRowSeed(t)
					v.dos = dos
					v.mem[0x50100] = y
					putRowWord(v, 0x48a10+0xedd, stride)
					putRowWord(v, 0x48a10+0xec1, x)
					p := planImageRow(v, 23)
					if p == nil {
						t.Fatalf("valid row rejected %d %d %d", stride, y, x)
					}
					compare(t, v, p)
					plans++
				}
			}
		}
	}
	for _, c := range [][5]int16{{0, -5, 0, 199, 0}, {5, -5, 0, 199, 0}, {10, 0, 20, 100, 0}, {10, 30, 20, 100, 20}, {200, 0, 20, 199, 20}, {1, 32766, 0, 32767, 32600}, {0, -32768, -32767, 0, 0}} {
		for _, dos := range []bool{false, true} {
			v := imageRowSeed(t)
			v.dos = dos
			v.mem[0x50100] = byte(c[0])
			putRowWord(v, 0x7080a, uint16(c[1]))
			putRowWord(v, 0x48a10+0xec3, uint16(c[2]))
			putRowWord(v, 0x48a10+0xec7, uint16(c[3]))
			putRowWord(v, 0x48a10+0xed5, uint16(c[4]))
			p := planImageRow(v, 23)
			if p == nil {
				t.Fatal("signed row rejected", c)
			}
			compare(t, v, p)
			plans++
		}
	}
	t.Log("image row synthetic original comparisons", plans)
}

func TestNativeImageRowEntryStates(t *testing.T) {
	m := original(rleSeed(t))
	states := []*view{}
	for m.Steps < 1000000 && len(states) < 64 {
		if m.CPU.Seg[cpu.CS] == 0x48a1 && m.CPU.IP == 0x6850 {
			states = append(states, &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}})
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if len(states) != 64 {
		t.Fatal("image row states absent", len(states))
	}
	plans := 0
	for _, v := range states {
		for _, dos := range []bool{false, true} {
			v.dos = dos
			for _, budget := range []uint64{23, 64, 1024} {
				p := planImageRow(v, budget)
				if p == nil {
					t.Fatalf("entry row rejected %+v", v.s)
				}
				compare(t, v, p)
				plans++
			}
		}
	}
	t.Log("image row original entry-state comparisons", plans)
}

func TestNativeImageRowRefusals(t *testing.T) {
	for _, kind := range []string{"budget", "code", "mode", "direction", "surface", "source", "source-wrap", "source-low", "argument", "argument-wrap", "source-alias", "argument-alias", "source-row-alias", "argument-row-alias", "signed-overflow", "bounds", "stride-zero", "stride-wide", "x-negative", "x-outside", "y-negative", "y-outside"} {
		t.Run(kind, func(t *testing.T) {
			v := imageRowSeed(t)
			budget := uint64(23)
			switch kind {
			case "budget":
				budget = 22
			case "code":
				v.mem[0x48a10+0x6850] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "direction":
				v.s.Flags |= cpu.DF
			case "surface":
				v.s.Seg[cpu.ES] = 0x9000
			case "source":
				v.s.Seg[cpu.DS] = 0xa000
			case "source-wrap":
				v.s.R[cpu.SI] = 65535
			case "source-low":
				v.s.Seg[cpu.DS] = 0
			case "argument":
				v.s.Seg[cpu.SS] = 0xa000
			case "argument-wrap":
				v.s.R[cpu.BP] = 65525
			case "source-alias":
				v.s.Seg[cpu.DS] = 0x48a1
				v.s.R[cpu.SI] = 0x1250
			case "argument-alias":
				v.s.Seg[cpu.SS] = 0x48a1
				v.s.R[cpu.BP] = 0x1246
			case "source-row-alias":
				v.s.Seg[cpu.DS] = 0x48a1
				v.s.R[cpu.SI] = 0x120c
			case "argument-row-alias":
				v.s.Seg[cpu.SS] = 0x48a1
				v.s.R[cpu.BP] = 0x1202
			case "signed-overflow":
				putRowWord(v, 0x7080a, 32767)
			case "bounds":
				putRowWord(v, 0x48a10+0xec3, 200)
			case "stride-zero":
				putRowWord(v, 0x48a10+0xedd, 0)
			case "stride-wide":
				putRowWord(v, 0x48a10+0xedd, 321)
			case "x-negative":
				putRowWord(v, 0x48a10+0xed3, 1)
			case "x-outside":
				putRowWord(v, 0x48a10+0xec1, 320)
			case "y-negative":
				putRowWord(v, 0x48a10+0xed5, 11)
			case "y-outside":
				putRowWord(v, 0x7080a, 190)
				putRowWord(v, 0x48a10+0xec7, 200)
			}
			before := append([]byte(nil), v.mem...)
			if planImageRow(v, budget) != nil {
				t.Fatal("unproved image row accepted")
			}
			if !bytes.Equal(before, v.mem) {
				t.Fatal("refusal mutated memory")
			}
		})
	}
}
