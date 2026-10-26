package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeClippedCursor(t *testing.T) {
	cases := 0
	inputs := [][4]int32{{-10, 0, 20, 90}, {20, 0, -10, 90}, {-10, -10, 100, 100}, {100, -10, -10, 100}, {-10, 20, 100, 30}, {100, 20, -10, 30}, {30, -10, 50, 100}, {50, -10, 30, 100}, {-100, -50, 420, 250}, {420, -50, -100, 250}, {30, 100, 50, -10}, {100, 30, -10, 20}}
	for _, box := range [][4]int32{{0, 0, 319, 199}, {25, 15, 70, 40}, {0, 0, 10, 100}, {25, 15, 70, 150}, {25, 15, 200, 40}} {
		for _, xy := range inputs {
			v := clipLineSeed(t, xy, true)
			for _, w := range [][2]uint16{{0x4bd, uint16(box[0])}, {0x6bd, uint16(box[1])}, {0x8bd, uint16(box[2])}, {0xabd, uint16(box[3])}} {
				putRowWord(v, 0x48a10+uint32(w[0]), w[1])
			}
			count := len(originalLinePixels(t, v))
			if count == 0 {
				continue
			}
			m := original(v)
			index := 0
			for m.Steps < 10000 {
				if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
					break
				}
				ip := m.CPU.IP
				if m.CPU.Seg[cpu.CS] == 0x48a1 && (ip == 0x309c || ip == 0x30b9 || ip == 0x3206 || ip == 0x321c) {
					if index == 0 || index == count/2 || index == count-1 {
						for _, dos := range []bool{false, true} {
							for _, budget := range []uint64{11, 18, 128, 10000} {
								w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}, dos: dos}
								p := planLineCursor(w, budget)
								if p == nil {
									if budget >= 18 {
										t.Fatalf("clip cursor refused box%v xy%v index%d count%d IP%x CX%d scratch%x/%x", box, xy, index, count, ip, w.CX(), w.Read16(0x48a10+0x1139), w.Read16(0x48a10+0x1133))
									}
									continue
								}
								t.Run("cursor", func(t *testing.T) { compare(t, w, p) })
								if t.Failed() {
									t.Fatalf("clip cursor mismatch box%v xy%v index%d count%d budget%d dos%v", box, xy, index, count, budget, dos)
								}
								cases++
							}
						}
					}
					index++
				}
				if err := m.Step(); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	t.Log("clipped cursor original comparisons", cases)
}
func TestNativeClippedCursorRefusals(t *testing.T) {
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
