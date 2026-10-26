package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func regionSeed(t *testing.T, xy [4]uint16, stride uint16, sub bool) *view {
	v := pointSeed(t, false, 0, 0, stride)
	v.s.IP = 0x3338
	for i, n := range []uint16{0x1234, 0x4567, 0, xy[0], xy[1], xy[2], xy[3], 0x8a2b} {
		putRowWord(v, 0x70900+uint32(i*2), n)
	}
	if sub {
		for _, w := range [][2]uint16{{0x4bd, 10}, {0x6bd, 12}, {0x8bd, 50}, {0xabd, 30}} {
			putRowWord(v, 0x48a10+uint32(w[0]), w[1])
		}
	}
	return v
}

func TestNativeRegionOriginal(t *testing.T) {
	cases := 0
	for _, sub := range []bool{false, true} {
		for _, stride := range []uint16{160, 320} {
			for _, xy := range [][4]uint16{{0, 0, 0, 0}, {0, 0, 1, 1}, {17, 19, 20, 22}, {0, 0, 319, 199}, {0xffff, 0xffff, 320, 200}, {0xffff, 0, 0xffff, 20}, {320, 0, 321, 0}, {0, 200, 0, 200}, {9, 11, 51, 31}, {10, 12, 50, 30}} {
				for _, dos := range []bool{false, true} {
					v := regionSeed(t, xy, stride, sub)
					v.dos = dos
					v.s.Flags = 0xed7
					p := planRegion(v, 10000)
					if p == nil {
						t.Fatal("region rejected", sub, stride, xy, dos)
					}
					compare(t, v, p)
					if planRegion(v, p.Instructions-1) != nil {
						t.Fatal("short budget")
					}
					cases++
				}
			}
		}
	}
	t.Log("region original comparisons", cases)
}

func TestNativeRegionParents(t *testing.T) {
	cases := 0
	for _, links := range []uint16{1, 3} {
		for _, xy := range [][4]uint16{{9, 11, 51, 31}, {10, 12, 50, 30}, {20, 20, 21, 21}} {
			for _, dos := range []bool{false, true} {
				v := regionSeed(t, xy, 320, true)
				v.dos = dos
				for i := uint16(0); i <= links; i++ {
					seg, flags := uint16(0xa000), uint16(0x80)
					left, top, right, bottom := uint16(0), uint16(0), uint16(319), uint16(199)
					if i < links {
						seg, flags = (i+1)*2, 0xc0
						left, top, right, bottom = 10, 12, 50, 30
					}
					for _, w := range [][2]uint16{{0xbd, seg}, {0x2bd, 0}, {0x4bd, left}, {0x6bd, top}, {0x8bd, right}, {0xabd, bottom}, {0xcbd, flags}} {
						putRowWord(v, 0x48a10+uint32(w[0]+2*i), w[1])
					}
				}
				p := planRegion(v, 10000)
				if p == nil {
					t.Fatal("parent rejected")
				}
				compare(t, v, p)
				cases++
			}
		}
	}
	t.Log("region parent comparisons", cases)
}

func TestNativeRegionRefusals(t *testing.T) {
	cases := 0
	for _, kind := range []string{"budget", "code", "mode", "CS", "SP-low", "SP-high", "stack-driver", "stack-video", "slot", "root", "bounds", "cycle", "odd-parent", "reverse-x", "reverse-y"} {
		v := regionSeed(t, [4]uint16{0, 0, 1, 1}, 320, false)
		b := uint64(10000)
		switch kind {
		case "budget":
			b = 1
		case "code":
			v.mem[0x48a10+0x3338] ^= 1
		case "mode":
			v.mem[0x449] = 0xd
		case "CS":
			v.s.Seg[cpu.CS] = 0x4800
		case "SP-low":
			v.s.R[cpu.SP] = 1
		case "SP-high":
			v.s.R[cpu.SP] = 65535
		case "stack-driver":
			v.s.Seg[cpu.SS] = 0x48a1
			v.s.R[cpu.SP] = 0x7000
		case "stack-video":
			v.s.Seg[cpu.SS] = 0xa000
		case "slot":
			putRowWord(v, 0x70904, 256)
		case "root":
			putRowWord(v, 0x48a10+0xbd, 0x9000)
		case "bounds":
			putRowWord(v, 0x48a10+0x8bd, 320)
		case "cycle":
			putRowWord(v, 0x48a10+0xcbd, 0xc0)
			putRowWord(v, 0x48a10+0xbd, 0)
		case "odd-parent":
			putRowWord(v, 0x48a10+0xcbd, 0xc0)
			putRowWord(v, 0x48a10+0xbd, 1)
		case "reverse-x":
			putRowWord(v, 0x70906, 2)
		case "reverse-y":
			putRowWord(v, 0x70908, 2)
		}
		before := append([]byte(nil), v.mem...)
		regs := v.s
		if planRegion(v, b) != nil {
			t.Fatal("invalid region accepted", kind)
		}
		if string(before) != string(v.mem) || regs != v.s {
			t.Fatal("refusal mutated")
		}
		cases++
	}
	t.Log("region refusals", cases)
}
