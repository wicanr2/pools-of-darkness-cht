package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func TestNativeSteepOriginal(t *testing.T) {
	cases := 0
	for _, stride := range []uint16{160, 320} {
		for _, xy := range [][4]uint16{{0, 0, 1, 2}, {1, 2, 0, 0}, {1, 0, 0, 2}, {0, 0, 10, 199}, {20, 10, 30, 31}, {30, 31, 20, 10}, {20, 10, 30, 30}, {50, 50, 52, 53}, {0, 0, 99, 199}} {
			for _, dos := range []bool{false, true} {
				for _, flags := range []uint16{2, 0x2d7, 0xed7} {
					v := axisSeed(t, xy, stride)
					v.dos = dos
					v.s.Flags = flags
					p := planSteepLine(v, 10000)
					if p == nil {
						t.Fatal("steep rejected", stride, xy, dos)
					}
					compare(t, v, p)
					if planSteepLine(v, p.Instructions-1) != nil {
						t.Fatal("short steep budget")
					}
					compare(t, v, planSteepLine(v, p.Instructions))
					cases++
				}
			}
		}
	}
	t.Log("steep original cases", cases, "independent plan comparisons", cases*2)
}

func TestNativeSteepParents(t *testing.T) {
	cases := 0
	for _, links := range []uint16{1, 3} {
		for _, xy := range [][4]uint16{{25, 15, 26, 17}, {26, 17, 25, 15}, {70, 15, 60, 40}, {60, 40, 70, 15}, {30, 15, 40, 40}, {40, 40, 30, 15}} {
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
				p := planSteepLine(v, 10000)
				if p == nil {
					t.Fatal("steep parent rejected", links, xy, dos)
				}
				compare(t, v, p)
				cases++
			}
		}
	}
	t.Log("steep parent original comparisons", cases)
}

func TestNativeSteepRefusals(t *testing.T) {
	cases := 0
	for _, kind := range []string{"budget", "code", "mode", "CS", "SP-low", "SP-high", "stack-code", "stack-video", "slot", "root", "bounds", "cycle", "odd-parent", "horizontal", "vertical", "shallow", "outside-x", "outside-y", "negative-x", "negative-y"} {
		v := axisSeed(t, [4]uint16{10, 10, 20, 30}, 320)
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
		case "horizontal":
			putRowWord(v, 0x7090c, 10)
		case "vertical":
			putRowWord(v, 0x7090a, 10)
		case "shallow":
			putRowWord(v, 0x7090a, 100)
		case "outside-x":
			putRowWord(v, 0x7090a, 320)
		case "outside-y":
			putRowWord(v, 0x7090c, 200)
		case "negative-x":
			putRowWord(v, 0x70906, 0xffff)
		case "negative-y":
			putRowWord(v, 0x70908, 0xffff)
		}
		before := append([]byte(nil), v.mem...)
		regs := v.s
		if planSteepLine(v, b) != nil {
			t.Fatal("invalid steep accepted", kind)
		}
		if string(before) != string(v.mem) || regs != v.s {
			t.Fatal("refusal mutated")
		}
		cases++
	}
	t.Log("steep refusals", cases)
}

func TestNativeShallowOriginal(t *testing.T) {
	cases := 0
	for _, stride := range []uint16{160, 320} {
		for _, xy := range [][4]uint16{{0, 0, 1, 1}, {1, 1, 0, 0}, {1, 0, 0, 1}, {50, 10, 10, 20}, {0, 0, 159, 10}, {159, 20, 0, 30}, {0, 0, 99, 99}, {20, 10, 40, 20}, {40, 20, 20, 10}} {
			for _, dos := range []bool{false, true} {
				for _, flags := range []uint16{2, 0x2d7, 0xed7} {
					v := axisSeed(t, xy, stride)
					v.dos = dos
					v.s.Flags = flags
					p := planShallowLine(v, 10000)
					if p == nil {
						t.Fatal("shallow rejected", stride, xy, dos)
					}
					compare(t, v, p)
					if planShallowLine(v, p.Instructions-1) != nil {
						t.Fatal("short shallow budget")
					}
					compare(t, v, planShallowLine(v, p.Instructions))
					cases++
				}
			}
		}
	}
	t.Log("shallow original cases", cases, "independent plan comparisons", cases*2)
}

func TestNativeShallowParents(t *testing.T) {
	cases := 0
	for _, links := range []uint16{1, 3} {
		for _, xy := range [][4]uint16{{25, 15, 27, 16}, {27, 16, 25, 15}, {70, 15, 45, 25}, {45, 25, 70, 15}, {30, 15, 55, 25}, {55, 25, 30, 15}} {
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
				p := planShallowLine(v, 10000)
				if p == nil {
					t.Fatal("shallow parent rejected", links, xy, dos)
				}
				compare(t, v, p)
				cases++
			}
		}
	}
	t.Log("shallow parent original comparisons", cases)
}

func TestNativeShallowRefusals(t *testing.T) {
	cases := 0
	for _, kind := range []string{"budget", "code", "mode", "CS", "SP-low", "SP-high", "stack-code", "stack-video", "slot", "root", "bounds", "cycle", "odd-parent", "horizontal", "vertical", "steep", "outside-x", "outside-y", "negative-x", "negative-y"} {
		v := axisSeed(t, [4]uint16{10, 10, 30, 20}, 320)
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
		case "horizontal":
			putRowWord(v, 0x7090c, 10)
		case "vertical":
			putRowWord(v, 0x7090a, 10)
		case "steep":
			putRowWord(v, 0x7090a, 11)
		case "outside-x":
			putRowWord(v, 0x7090a, 320)
		case "outside-y":
			putRowWord(v, 0x7090c, 200)
		case "negative-x":
			putRowWord(v, 0x70906, 0xffff)
		case "negative-y":
			putRowWord(v, 0x70908, 0xffff)
		}
		before := append([]byte(nil), v.mem...)
		regs := v.s
		if planShallowLine(v, b) != nil {
			t.Fatal("invalid shallow accepted", kind)
		}
		if string(before) != string(v.mem) || regs != v.s {
			t.Fatal("refusal mutated")
		}
		cases++
	}
	t.Log("shallow refusals", cases)
}
