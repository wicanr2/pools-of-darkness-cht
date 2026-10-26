package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func paletteAdvanceStates(t *testing.T) []*view {
	var seed *view
	for _, v := range drawCalls(t) {
		if v.CS() == 0x48a1 && v.IP() == 0x4944 {
			seed = v
			break
		}
	}
	if seed == nil {
		t.Fatal("fade entry absent")
	}
	m := original(seed)
	out := []*view{}
	for m.Steps < 1000000 {
		if m.CPU.Seg[cpu.CS] == 0x48a1 && m.CPU.IP == 0x4a72 {
			y := m.CPU.R[cpu.BP]
			if y == 0 || y == 99 || y == 199 {
				out = append(out, &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}})
			}
			if y == 199 {
				break
			}
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if len(out) != 3 {
		t.Fatal("fade advance states", len(out))
	}
	return out
}

func TestNativePaletteAdvanceOriginal(t *testing.T) {
	cases := 0
	for _, v := range paletteAdvanceStates(t) {
		for _, stride := range []uint16{160, 320} {
			for _, dos := range []bool{false, true} {
				for _, entry := range []uint16{0x4a72, 0x4a31} {
					if v.BP() == 199 && entry == 0x4a31 {
						continue
					}
					w := &view{mem: append([]byte(nil), v.mem...), s: v.s, dos: dos}
					putRowWord(w, 0x48a10+0xedd, stride)
					putRowWord(w, 0x48a10+0x1135, stride)
					if entry == 0x4a31 {
						w.s.IP = entry
						w.s.R[cpu.SP] += 2
						w.s.R[cpu.BP]++
						w.s.Seg[cpu.DS] = 0x48a1
					}
					p := planPaletteAdvance(w, 100)
					if p == nil {
						t.Fatal("advance rejected", v.BP(), stride, dos, entry)
					}
					compare(t, w, p)
					cases++
				}
			}
		}
	}
	t.Log("palette row advance original comparisons", cases)
}

func TestNativePaletteAdvanceGeometry(t *testing.T) {
	seed := paletteAdvanceStates(t)[0]
	cases := 0
	for _, stride := range []uint16{1, 8, 160, 320} {
		for _, origin := range [][2]uint16{{0, 0}, {20, 10}} {
			for _, row := range []uint16{0, 1, 99} {
				for _, dos := range []bool{false, true} {
					for _, entry := range []uint16{0x4a72, 0x4a31} {
						v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
						left, top := origin[0], origin[1]
						if stride == 320 {
							left = 0
						}
						x := uint16(0)
						if stride > 3 {
							x = 3
						}
						putRowWord(v, 0x48a10+0xedd, stride)
						putRowWord(v, 0x48a10+0x1135, stride-x)
						putRowWord(v, 0x48a10+0xed3, left)
						putRowWord(v, 0x48a10+0xec1, left+x)
						putRowWord(v, 0x48a10+0xed5, top)
						putRowWord(v, 0x48a10+0xec7, top+150)
						v.s.R[cpu.BP] = top + row
						if entry == 0x4a31 {
							v.s.IP = entry
							v.s.R[cpu.SP] += 2
							v.s.R[cpu.BP]++
							v.s.Seg[cpu.DS] = 0x48a1
						}
						p := planPaletteAdvance(v, 100)
						if p == nil {
							t.Fatal("valid geometry rejected", stride, origin, row, dos, entry)
						}
						compare(t, v, p)
						if planPaletteAdvance(v, p.Instructions-1) != nil {
							t.Fatal("short budget accepted")
						}
						compare(t, v, planPaletteAdvance(v, p.Instructions))
						cases++
					}
				}
			}
		}
	}
	t.Log("palette row geometry original cases", cases, "plan comparisons", cases*2)
}

func TestNativePaletteAdvanceRefusals(t *testing.T) {
	seed := paletteAdvanceStates(t)[0]
	for _, kind := range []string{"budget", "code", "mode", "DF", "ES", "DS", "saved-DS", "SP", "stack-video", "stack-code", "bottom", "y", "stride-zero", "stride-large", "width-zero", "width-large", "left", "top"} {
		t.Run(kind, func(t *testing.T) {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			stack := uint32(v.SS())*16 + uint32(v.SP())
			budget := uint64(100)
			switch kind {
			case "budget":
				budget = 1
			case "code":
				v.mem[0x48a10+0x4a31] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "DF":
				v.s.Flags |= cpu.DF
			case "ES":
				v.s.Seg[cpu.ES] = 0xa000
			case "DS":
				v.s.Seg[cpu.DS] = 0x9000
			case "saved-DS":
				putRowWord(v, stack, 0)
			case "SP":
				v.s.R[cpu.SP] = 65535
			case "stack-video":
				v.s.Seg[cpu.SS] = 0xa000
			case "stack-code":
				v.s.Seg[cpu.SS] = 0x48a1
				v.s.R[cpu.SP] = 0x7000
			case "bottom":
				putRowWord(v, 0x48a10+0xec7, 200)
			case "y":
				v.s.R[cpu.BP] = 200
			case "stride-zero":
				putRowWord(v, 0x48a10+0xedd, 0)
			case "stride-large":
				putRowWord(v, 0x48a10+0xedd, 321)
			case "width-zero":
				putRowWord(v, 0x48a10+0x1135, 0)
			case "width-large":
				putRowWord(v, 0x48a10+0x1135, 321)
			case "left":
				putRowWord(v, 0x48a10+0xed3, 1)
			case "top":
				putRowWord(v, 0x48a10+0xed5, 2)
			}
			before := append([]byte(nil), v.mem...)
			regs := v.s
			if planPaletteAdvance(v, budget) != nil {
				t.Fatal("invalid advance admitted")
			}
			if string(before) != string(v.mem) || v.s != regs {
				t.Fatal("decline mutated input")
			}
		})
	}
}

func paletteFrameSeed(t *testing.T) *view {
	v := paletteAdvanceStates(t)[0]
	stack := uint32(v.SS())*16 + uint32(v.SP())
	v.s.R[cpu.SP] += 4
	v.s.R[cpu.BP] = v.Read16(stack + 2)
	v.s.IP = 0x4b55
	v.s.Seg[cpu.DS] = 0x48a1
	return v
}

func TestNativePaletteFrameOriginal(t *testing.T) {
	seed := paletteFrameSeed(t)
	cases := 0
	for _, es := range []uint16{0x48a1, 0x40, 0xa000} {
		for _, remaining := range []byte{0, 1, 63} {
			for _, count := range []uint16{1, 2, 256} {
				for _, flags := range []uint16{0x7202, 0x7ac7} {
					for _, dos := range []bool{false, true} {
						v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
						v.s.Flags = flags
						v.s.Seg[cpu.ES] = es
						v.mem[0x48a10+0x1174] = remaining
						putRowWord(v, 0x48a10+0x116f, count)
						p := planPaletteFrame(v, 100)
						if p == nil {
							t.Fatal("valid frame control rejected")
						}
						compare(t, v, p)
						if planPaletteFrame(v, p.Instructions-1) != nil {
							t.Fatal("short budget admitted")
						}
						compare(t, v, planPaletteFrame(v, p.Instructions))
						cases++
					}
				}
			}
		}
	}
	t.Log("palette frame control original cases", cases, "plan comparisons", cases*2)
}

func TestNativePaletteFrameRefusals(t *testing.T) {
	seed := paletteFrameSeed(t)
	for _, kind := range []string{"budget", "code", "mode", "DF", "DS", "counter", "count-zero", "count-large", "return-code", "stack-wrap", "stack-low", "stack-video", "stack-code"} {
		t.Run(kind, func(t *testing.T) {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			v.mem[0x48a10+0x1174] = 1
			putRowWord(v, 0x48a10+0x116f, 256)
			budget := uint64(100)
			switch kind {
			case "budget":
				budget = 1
			case "code":
				v.mem[0x48a10+0x4b55] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "DF":
				v.s.Flags |= cpu.DF
			case "DS":
				v.s.Seg[cpu.DS] = 0xa000
			case "counter":
				v.mem[0x48a10+0x1174] = 64
			case "count-zero":
				putRowWord(v, 0x48a10+0x116f, 0)
			case "count-large":
				putRowWord(v, 0x48a10+0x116f, 257)
			case "return-code":
				v.mem[0x48a10+0x1174] = 0
				v.mem[0x48a10+0x4bc6] ^= 1
			case "stack-wrap":
				v.mem[0x48a10+0x1174] = 0
				v.s.R[cpu.SP] = 65535
			case "stack-low":
				v.mem[0x48a10+0x1174] = 0
				v.s.R[cpu.SP] = 0
				v.s.Seg[cpu.SS] = 0
			case "stack-video":
				v.mem[0x48a10+0x1174] = 0
				v.s.Seg[cpu.SS] = 0xa000
			case "stack-code":
				v.mem[0x48a10+0x1174] = 0
				v.s.Seg[cpu.SS] = 0x48a1
				v.s.R[cpu.SP] = 0x7000
			}
			before := append([]byte(nil), v.mem...)
			regs := v.s
			if planPaletteFrame(v, budget) != nil {
				t.Fatal("invalid frame control admitted")
			}
			if string(before) != string(v.mem) || v.s != regs {
				t.Fatal("decline mutated state")
			}
		})
	}
}
