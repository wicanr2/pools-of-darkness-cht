package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func pointSeed(t *testing.T, read bool, x, y uint16, stride uint16) *view {
	v := imageSpanSeed(t)
	v.s.IP = 0x2b10
	if read {
		v.s.IP = 0x2c04
	}
	v.s.R[cpu.SP] = 0x900
	for i, n := range []uint16{0x1234, 0x4567, 0, x, y, 0xab2a} {
		putRowWord(v, 0x70900+uint32(2*i), n)
	}
	for _, w := range [][2]uint16{{0xbd, 0xa000}, {0x2bd, 0}, {0x4bd, 0}, {0x6bd, 0}, {0x8bd, stride - 1}, {0xabd, 199}, {0xcbd, 0x80}} {
		putRowWord(v, 0x48a10+uint32(w[0]), w[1])
	}
	return v
}

func TestNativePointCosts(t *testing.T) {
	for _, read := range []bool{false, true} {
		for _, xy := range [][2]uint16{{0, 0}, {17, 99}, {0xffff, 0}, {320, 0}, {0, 0xffff}, {0, 200}} {
			for _, stride := range []uint16{160, 320} {
				v := pointSeed(t, read, xy[0], xy[1], stride)
				m := original(v)
				for m.Steps < 1000 {
					if m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
						break
					}
					if err := m.Step(); err != nil {
						t.Fatal(err)
					}
				}
				t.Logf("point read%v xy%v stride%d N%d C%d flags%x", read, xy, stride, m.Steps, m.CPU.Cycles, m.CPU.Flags)
			}
		}
	}
}

func TestNativePointOriginal(t *testing.T) {
	cases := 0
	for _, read := range []bool{false, true} {
		for _, stride := range []uint16{160, 320} {
			for _, xy := range [][2]uint16{{0, 0}, {17, 99}, {159, 199}, {319, 199}, {0xffff, 0}, {320, 0}, {0, 0xffff}, {0, 200}} {
				for _, flags := range []uint16{2, 0x2d7, 0xed7} {
					for _, dos := range []bool{false, true} {
						v := pointSeed(t, read, xy[0], xy[1], stride)
						v.s.Flags = flags
						v.dos = dos
						for i := uint32(0); i < 64000; i++ {
							v.mem[0xa0000+i] = byte(i*17 + 21)
						}
						p := planPoint(v, 200)
						if p == nil {
							t.Fatal("point rejected", read, xy, stride)
						}
						compare(t, v, p)
						if planPoint(v, p.Instructions-1) != nil {
							t.Fatal("short budget")
						}
						compare(t, v, planPoint(v, p.Instructions))
						cases += 2
					}
				}
			}
		}
	}
	t.Log("point independent original comparisons", cases)
}

func TestNativePointSurfaces(t *testing.T) {
	cases := 0
	for _, read := range []bool{false, true} {
		for _, links := range []uint16{0, 1, 3} {
			for _, xy := range [][2]uint16{{20, 10}, {50, 30}, {19, 10}, {51, 30}, {20, 9}, {50, 31}} {
				for _, dos := range []bool{false, true} {
					v := pointSeed(t, read, xy[0], xy[1], 160)
					v.dos = dos
					for i := uint16(0); i <= links; i++ {
						seg, flags := uint16(0xa000), uint16(0x80)
						if i < links {
							seg, flags = (i+1)*2, 0xc0
						}
						for _, w := range [][2]uint16{{0xbd, seg}, {0xcbd, flags}, {0x4bd, 20}, {0x6bd, 10}, {0x8bd, 50}, {0xabd, 30}} {
							putRowWord(v, 0x48a10+uint32(w[0]+2*i), w[1])
						}
					}
					p := planPoint(v, 200)
					if p == nil {
						t.Fatal("surface rejected")
					}
					compare(t, v, p)
					cases++
				}
			}
		}
	}
	t.Log("point surface original comparisons", cases)
}

func TestNativePointRefusals(t *testing.T) {
	cases := 0
	for _, read := range []bool{false, true} {
		for _, kind := range []string{"budget", "code", "mode", "CS", "SP-low", "SP-high", "stack-video", "stack-driver", "slot", "root", "bounds", "cycle", "chain-odd"} {
			v := pointSeed(t, read, 1, 1, 320)
			b := uint64(200)
			switch kind {
			case "budget":
				b = 1
			case "code":
				v.mem[0x48a10+uint32(v.IP())] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "CS":
				v.s.Seg[cpu.CS] = 0x4800
			case "SP-low":
				v.s.R[cpu.SP] = 1
			case "SP-high":
				v.s.R[cpu.SP] = 65535
			case "stack-video":
				v.s.Seg[cpu.SS] = 0xa000
			case "stack-driver":
				v.s.Seg[cpu.SS] = 0x48a1
				v.s.R[cpu.SP] = 0x7000
			case "slot":
				putRowWord(v, 0x70904, 256)
			case "root":
				putRowWord(v, 0x48a10+0xbd, 0x9000)
			case "bounds":
				putRowWord(v, 0x48a10+0x8bd, 320)
			case "cycle":
				putRowWord(v, 0x48a10+0xcbd, 0xc0)
				putRowWord(v, 0x48a10+0xbd, 0)
			case "chain-odd":
				putRowWord(v, 0x48a10+0xcbd, 0xc0)
				putRowWord(v, 0x48a10+0xbd, 1)
			}
			before := append([]byte(nil), v.mem...)
			regs := v.s
			if planPoint(v, b) != nil {
				t.Fatal("invalid point accepted", read, kind)
			}
			if string(before) != string(v.mem) || regs != v.s {
				t.Fatal("refusal mutated")
			}
			cases++
		}
	}
	t.Log("point refusals", cases)
}
