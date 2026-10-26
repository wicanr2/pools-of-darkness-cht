package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func axisSeed(t *testing.T, xy [4]uint16, stride uint16) *view {
	v := regionSeed(t, xy, stride, false)
	v.s.IP = 0x2cfc
	return v
}
func TestNativeAxisOriginal(t *testing.T) {
	cases := 0
	for _, stride := range []uint16{160, 320} {
		for _, xy := range [][4]uint16{{0, 0, 0, 0}, {0, 0, 1, 0}, {0, 0, 159, 0}, {159, 10, 0, 10}, {0xffff, 20, 320, 20}, {17, 0, 17, 0}, {17, 0, 17, 1}, {17, 0, 17, 6}, {17, 0, 17, 7}, {17, 0, 17, 8}, {17, 199, 17, 0}, {17, 0xffff, 17, 200}} {
			for _, dos := range []bool{false, true} {
				for _, flags := range []uint16{2, 0x2d7, 0xed7} {
					v := axisSeed(t, xy, stride)
					v.dos = dos
					v.s.Flags = flags
					p := planAxisLine(v, 10000)
					if p == nil {
						t.Fatal("axis rejected", stride, xy, dos)
					}
					compare(t, v, p)
					if planAxisLine(v, p.Instructions-1) != nil {
						t.Fatal("short axis budget")
					}
					compare(t, v, planAxisLine(v, p.Instructions))
					cases++
				}
			}
		}
	}
	t.Log("axis original cases", cases, "independent plan comparisons", cases*2)
}

func TestNativeAxisParents(t *testing.T) {
	cases := 0
	for _, links := range []uint16{1, 3} {
		for _, xy := range [][4]uint16{{25, 15, 25, 15}, {0, 20, 200, 20}, {70, 25, 25, 25}, {30, 0, 30, 60}, {30, 40, 30, 15}, {30, 15, 30, 22}} {
			for _, dos := range []bool{false, true} {
				v := axisSeed(t, xy, 320)
				v.dos = dos
				for i := uint16(0); i <= links; i++ {
					seg, flags := uint16(0xa000), uint16(0x80)
					left, top, right, bottom := uint16(20), uint16(10), uint16(100), uint16(60)
					if i < links {
						seg, flags = (i+1)*2, 0xc0
						left, top, right, bottom = 25, 15, 70, 40
					}
					for _, w := range [][2]uint16{{0xbd, seg}, {0x2bd, 0}, {0x4bd, left}, {0x6bd, top}, {0x8bd, right}, {0xabd, bottom}, {0xcbd, flags}} {
						putRowWord(v, 0x48a10+uint32(w[0]+2*i), w[1])
					}
				}
				p := planAxisLine(v, 10000)
				if p == nil {
					t.Fatal("axis parent rejected", links, xy, dos)
				}
				compare(t, v, p)
				cases++
			}
		}
	}
	t.Log("axis parent original comparisons", cases)
}

func TestNativeAxisRefusals(t *testing.T) {
	cases := 0
	for _, kind := range []string{"budget", "code", "mode", "CS", "SP-low", "SP-high", "stack-code", "stack-video", "slot", "root", "bounds", "cycle", "odd-parent", "not-axis", "outside-x", "outside-y", "difference-x", "difference-y"} {
		v := axisSeed(t, [4]uint16{0, 0, 17, 0}, 320)
		b := uint64(10000)
		switch kind {
		case "budget":
			b = 1
		case "code":
			v.mem[0x48a10+0x2cfc] ^= 1
		case "mode":
			v.mem[0x449] = 0xd
		case "CS":
			v.s.Seg[cpu.CS] = 0x4800
		case "SP-low":
			v.s.R[cpu.SP] = 1
		case "SP-high":
			v.s.R[cpu.SP] = 65535
		case "stack-code":
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
		case "not-axis":
			putRowWord(v, 0x7090c, 1)
		case "outside-x":
			putRowWord(v, 0x70906, 320)
			putRowWord(v, 0x7090a, 320)
		case "outside-y":
			putRowWord(v, 0x70908, 200)
			putRowWord(v, 0x7090c, 200)
		case "difference-x":
			putRowWord(v, 0x70906, 0x8000)
			putRowWord(v, 0x7090a, 0x7fff)
		case "difference-y":
			putRowWord(v, 0x70906, 17)
			putRowWord(v, 0x70908, 0x8000)
			putRowWord(v, 0x7090c, 0x7fff)
		}
		before := append([]byte(nil), v.mem...)
		regs := v.s
		if planAxisLine(v, b) != nil {
			t.Fatal("invalid axis accepted", kind)
		}
		if string(before) != string(v.mem) || regs != v.s {
			t.Fatal("refusal mutated")
		}
		cases++
	}
	t.Log("axis refusals", cases)
}
