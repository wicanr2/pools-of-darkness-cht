package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func TestResearchImageInitializeOriginal(t *testing.T) {
	cases := 0
	for _, flags := range []uint16{0, 1, 2, 3, 0xfffc} {
		for _, missing := range []bool{false, true} {
			for _, dos := range []bool{false, true} {
				v := rleSeed(t)
				v.dos = dos
				stack := uint32(v.SS())*16 + uint32(v.SP())
				putRowWord(v, stack+10, flags)
				if missing {
					putRowWord(v, stack+16, 65535)
				}
				p := planImageInitialize(v, 1024)
				if p == nil {
					t.Fatal("valid init rejected", flags, missing, dos)
				}
				compare(t, v, p)
				cases++
			}
		}
	}
	t.Log("source/surface initialization original comparisons", cases)
}

func TestResearchImageInitializeGeometry(t *testing.T) {
	cases := 0
	for _, links := range []int{0, 1, 2, 3} {
		for _, flags := range []uint16{0, 1, 2, 3} {
			for _, index := range []uint16{0, 1, 2} {
				for _, dos := range []bool{false, true} {
					v := rleSeed(t)
					v.dos = dos
					stack := uint32(v.SS())*16 + uint32(v.SP())
					fields := []uint32{0xbd, 0x2bd, 0x4bd, 0x6bd, 0x8bd, 0xabd, 0xcbd}
					root := make([]uint16, len(fields))
					for i, off := range fields {
						root[i] = v.Read16(0x48a10 + off)
					}
					slots := []uint16{4, 20, 22, 0}
					for level := 0; level <= links; level++ {
						slot := slots[level]
						for i, off := range fields {
							putRowWord(v, 0x48a10+off+uint32(slot), root[i])
						}
						if level < links {
							putRowWord(v, 0x48a10+0xbd+uint32(slot), slots[level+1])
							putRowWord(v, 0x48a10+0xcbd+uint32(slot), 64)
						}
					}
					putRowWord(v, stack+10, flags)
					putRowWord(v, stack+12, 0xfff0)
					putRowWord(v, stack+14, 0x4000)
					putRowWord(v, stack+16, index)
					putRowWord(v, 0x4fff4, 2)
					putRowWord(v, 0x4fff6, 0x0030)
					putRowWord(v, 0x4fff8, 1)
					putRowWord(v, 0x4fffa, 0xfffc)
					putRowWord(v, 0x4fffc, 0)
					p := planImageInitialize(v, 1024)
					if p == nil {
						t.Fatal("valid initialization rejected", links, flags, index, dos)
					}
					compare(t, v, p)
					cases++
				}
			}
		}
	}
	t.Log("surface chain / frame address original comparisons", cases)
}

func TestResearchImageInitializeRefusals(t *testing.T) {
	for _, kind := range []string{"budget", "code", "mode", "stack", "stack-video", "stack-code", "slot", "cycle", "root", "source-low", "source-video", "source-wrap", "source-scratch", "source-stack", "index-wrap", "relative-wrap"} {
		t.Run(kind, func(t *testing.T) {
			v := rleSeed(t)
			stack := uint32(v.SS())*16 + uint32(v.SP())
			budget := uint64(1024)
			switch kind {
			case "budget":
				budget = 1
			case "code":
				v.mem[0x48a10+0x6733] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "stack":
				v.s.R[cpu.SP] = 65530
			case "stack-video":
				v.s.Seg[cpu.SS] = 0xa000
			case "stack-code":
				v.s.Seg[cpu.SS] = 0x48a1
				v.s.R[cpu.SP] = 0x7000
			case "slot":
				putRowWord(v, stack+4, 256)
			case "cycle":
				putRowWord(v, 0x48a10+0xbd+4, 4)
			case "root":
				putRowWord(v, 0x48a10+0xbd, 0x9000)
			case "source-low":
				putRowWord(v, stack+12, 0)
				putRowWord(v, stack+14, 0)
			case "source-video":
				putRowWord(v, stack+14, 0xa000)
			case "source-wrap":
				putRowWord(v, stack+12, 65534)
			case "source-scratch":
				putRowWord(v, stack+14, 0x48a1)
				putRowWord(v, stack+12, 0xebd)
			case "source-stack":
				putRowWord(v, stack+14, v.SS())
				putRowWord(v, stack+12, v.SP()-8)
			case "index-wrap":
				putRowWord(v, stack+14, 0x5000)
				putRowWord(v, stack+12, 0)
				putRowWord(v, 0x50004, 65535)
				putRowWord(v, stack+16, 16382)
			case "relative-wrap":
				head := uint32(v.Read16(stack+14))*16 + uint32(v.Read16(stack+12))
				putRowWord(v, head+6, 65535)
				putRowWord(v, head+8, 65535)
			}
			before := append([]byte(nil), v.mem...)
			regs := v.s
			if planImageInitialize(v, budget) != nil {
				t.Fatal("invalid init accepted")
			}
			if string(before) != string(v.mem) || v.s != regs {
				t.Fatal("decline mutated state")
			}
		})
	}
}
