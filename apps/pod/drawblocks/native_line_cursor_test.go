package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func lineCursorStates(t *testing.T, xy [4]uint16, stride uint16) []*view {
	return lineCursorStatesFrom(t, axisSeed(t, xy, stride))
}
func lineCursorStatesFrom(t *testing.T, v *view) []*view {
	m := original(v)
	stack := uint32(v.SS())*16 + uint32(v.SP())
	dx := int32(v.Read16(stack+10)) - int32(v.Read16(stack+6))
	if dx < 0 {
		dx = -dx
	}
	dy := int32(v.Read16(stack+12)) - int32(v.Read16(stack+8))
	if dy < 0 {
		dy = -dy
	}
	major := dx
	if dy > dx {
		major = dy
	}
	out := []*view{}
	index := int32(0)
	for m.Steps < 10000 {
		ip := m.CPU.IP
		if m.CPU.Seg[cpu.CS] == 0x48a1 && (ip == 0x309c || ip == 0x30b9 || ip == 0x3206 || ip == 0x321c) {
			if index == 0 || index == major/2 || index == major {
				out = append(out, &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}})
			}
			if index == major {
				break
			}
			index++
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if len(out) != 3 {
		t.Fatal("line cursor states", len(out))
	}
	return out
}

func TestNativeLineCursorOriginal(t *testing.T) {
	cases := 0
	for _, xy := range [][4]uint16{{20, 10, 30, 31}, {30, 10, 20, 31}, {20, 10, 40, 20}, {40, 10, 20, 20}} {
		for _, stride := range []uint16{160, 320} {
			for _, seed := range lineCursorStates(t, xy, stride) {
				for _, dos := range []bool{false, true} {
					for _, budget := range []uint64{1, 6, 7, 9, 10, 11, 27, 128, 1024} {
						v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
						p := planLineCursor(v, budget)
						if p == nil {
							if budget >= 11 {
								t.Fatal("line cursor rejected", xy, stride, seed.CX(), dos, budget)
							}
							continue
						}
						compare(t, v, p)
						cases++
					}
				}
			}
		}
	}
	t.Log("line cursor original comparisons", cases)
}

func TestNativeLineCursorParents(t *testing.T) {
	cases := 0
	for _, links := range []uint16{1, 3} {
		for _, xy := range [][4]uint16{{25, 15, 35, 40}, {70, 15, 60, 40}, {25, 15, 70, 35}, {70, 15, 25, 35}} {
			v := axisSeed(t, xy, 320)
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
			for _, seed := range lineCursorStatesFrom(t, v) {
				for _, dos := range []bool{false, true} {
					for _, budget := range []uint64{11, 27, 128} {
						w := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
						p := planLineCursor(w, budget)
						if p == nil {
							t.Fatal("parent line cursor rejected", links, xy, seed.CX(), dos, budget)
						}
						compare(t, w, p)
						cases++
					}
				}
			}
		}
	}
	t.Log("line cursor parent comparisons", cases)
}

func TestNativeLineCursorRefusals(t *testing.T) {
	cases := 0
	for _, xy := range [][4]uint16{{20, 10, 30, 31}, {40, 10, 20, 20}} {
		seed := lineCursorStates(t, xy, 320)[1]
		for _, kind := range []string{"budget", "code", "mode", "CS", "DS", "ES", "DF", "SP", "stack-video", "stack-code", "slot", "CX", "BX", "DX", "AX", "SI", "BP", "DI", "parameter", "scratch", "left", "stride"} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			b := uint64(100)
			stack := uint32(v.SS())*16 + uint32(v.SP())
			switch kind {
			case "budget":
				b = 0
			case "code":
				v.mem[0x48a10+0x2cfc] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "CS":
				v.s.Seg[cpu.CS] = 0x4800
			case "DS":
				v.s.Seg[cpu.DS] = 0x5000
			case "ES":
				v.s.Seg[cpu.ES] = 0x9000
			case "DF":
				v.s.Flags ^= cpu.DF
			case "SP":
				v.s.R[cpu.SP] = 65535
			case "stack-video":
				v.s.Seg[cpu.SS] = 0xa000
			case "stack-code":
				v.s.Seg[cpu.SS] = 0x48a1
				v.s.R[cpu.SP] = 0x7000
			case "slot":
				putRowWord(v, stack+12, 256)
			case "CX":
				v.s.R[cpu.CX] = 0xffff
			case "BX":
				v.s.R[cpu.BX]++
			case "DX":
				v.s.R[cpu.DX]++
			case "AX":
				v.s.R[cpu.AX]++
			case "SI":
				v.s.R[cpu.SI]++
			case "BP":
				v.s.R[cpu.BP]++
			case "DI":
				v.s.R[cpu.DI]++
			case "parameter":
				putRowWord(v, stack+22, 0)
			case "scratch":
				putRowWord(v, 0x48a10+0x1137, 0xffff)
			case "left":
				putRowWord(v, 0x48a10+0xec1, 0xffff)
			case "stride":
				putRowWord(v, 0x48a10+0xedd, 0)
			}
			before := append([]byte(nil), v.mem...)
			regs := v.s
			if planLineCursor(v, b) != nil {
				t.Fatal("invalid line cursor accepted", seed.IP(), kind)
			}
			if string(before) != string(v.mem) || regs != v.s {
				t.Fatal("refusal mutated")
			}
			cases++
		}
	}
	t.Log("line cursor refusals", cases)
}

func TestNativeDiagonalPrefixes(t *testing.T) {
	cases := 0
	for _, xy := range [][4]uint16{{20, 10, 30, 31}, {30, 31, 20, 10}, {30, 10, 20, 31}, {20, 31, 30, 10}, {20, 10, 40, 20}, {40, 20, 20, 10}, {40, 10, 20, 20}, {20, 20, 40, 10}} {
		for _, stride := range []uint16{160, 320} {
			for _, dos := range []bool{false, true} {
				v := axisSeed(t, xy, stride)
				v.dos = dos
				v.s.Flags = 0xed7
				var prefix *machine.HostWorkPlan
				if plans := buildSteepLine(v); plans != nil {
					prefix = plans.prefix
				} else if plans := buildShallowLine(v); plans != nil {
					prefix = plans.prefix
				}
				if prefix == nil {
					t.Fatal("prefix absent", xy, stride, dos)
				}
				compare(t, v, prefix)
				cases++
			}
		}
	}
	t.Log("diagonal prefix original comparisons", cases)
}

func planDiagonalBounded(v CPUView, budget uint64) *machine.HostWorkPlan { return planLine(v, budget) }
func TestNativeDiagonalBoundedWhole(t *testing.T) {
	cases, plans := 0, 0
	for _, xy := range [][4]uint16{{20, 10, 30, 31}, {30, 31, 20, 10}, {30, 10, 20, 31}, {20, 31, 30, 10}, {20, 10, 40, 20}, {40, 20, 20, 10}, {40, 10, 20, 20}, {20, 20, 40, 10}} {
		for _, stride := range []uint16{160, 320} {
			for _, dos := range []bool{false, true} {
				for _, budget := range []uint64{11, 160, 1024, 10000} {
					v := axisSeed(t, xy, stride)
					v.dos = dos
					a, b := callMachine(v), callMachine(v)
					runCall(t, a, v, false)
					used := 0
					for b.Steps < 1000000 {
						if b.Steps > 0 && b.CPU.IP == 0x1234 && b.CPU.Seg[cpu.CS] == 0x4567 {
							break
						}
						w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
						if p := planDiagonalBounded(w, budget); p != nil {
							if err := b.CommitHostBlock(p); err != nil {
								t.Fatal(err)
							}
							used++
						} else if err := b.Step(); err != nil {
							t.Fatal(err)
						}
					}
					if used == 0 {
						t.Fatal("no bounded diagonal Go plan")
					}
					var sa, sb bytes.Buffer
					if err := a.SaveState(&sa); err != nil {
						t.Fatal(err)
					}
					if err := b.SaveState(&sb); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
						t.Fatal("bounded diagonal complete state differs", xy, stride, dos, budget, a.Steps, b.Steps)
					}
					cases++
					plans += used
				}
			}
		}
	}
	t.Log("bounded diagonal whole original comparisons", cases, "Go plans", plans)
}
