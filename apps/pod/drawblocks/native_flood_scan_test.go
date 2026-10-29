package drawblocks

import (
	"bytes"
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func floodScanSeed(t *testing.T, backward bool, start, count uint16) *view {
	v := pointSeed(t, true, 5, 5, 320)
	v.s.IP = 0x5602
	v.s.R[cpu.AX], v.s.R[cpu.DI], v.s.R[cpu.CX] = 0xab27, start, count
	v.s.Seg[cpu.ES] = 0xa000
	putRowWord(v, 0x48a10+0xedd, 320)
	putRowWord(v, 0x48a10+0xedf, 200)
	v.s.Flags = 0x893
	if backward {
		v.s.IP, v.s.Flags = 0x55cd, v.s.Flags|cpu.DF
	}
	for i := 0xa0000; i < 0xafa00; i++ {
		v.mem[i] = 0x27
	}
	return v
}

func TestNativeFloodScanOriginal(t *testing.T) {
	cases := 0
	for _, width := range []uint16{32, 160, 320} {
		for _, backward := range []bool{false, true} {
			for _, row := range []uint16{0, 99, 199} {
				for _, count := range []uint16{0, 1, 2, 19, width} {
					start := row * width
					if backward {
						start += width - 1
					}
					for _, mismatch := range []int{-1, 0, int(count) / 2, int(count) - 1} {
						for _, dos := range []bool{false, true} {
							v := floodScanSeed(t, backward, start, count)
							putRowWord(v, 0x48a10+0xedd, width)
							v.dos = dos
							if mismatch >= 0 && mismatch < int(count) {
								at := int(start) + mismatch
								if backward {
									at = int(start) - mismatch
								}
								v.mem[0xa0000+at] = 0x80
							}
							before, regs := append([]byte(nil), v.mem...), v.s
							p := Plan(v, 1)
							if p == nil || !p.Algorithmic || p.Instructions != 1 || len(p.Writes)+len(p.Transfers) != 0 {
								t.Fatal("row scan not native", backward, row, count, mismatch)
							}
							if !bytes.Equal(before, v.mem) || v.s != regs {
								t.Fatal("planner mutated")
							}
							compare(t, v, p)
							cases++
						}
					}
				}
			}
		}
	}
	t.Log("original row scan comparisons", cases)
}

func TestNativeFloodScanRefusals(t *testing.T) {
	for _, backward := range []bool{false, true} {
		for _, kind := range []string{"budget", "mode", "segment", "direction", "count", "row", "end", "code", "cs", "ip", "width-zero", "width-high", "height-zero", "height-high"} {
			v := floodScanSeed(t, backward, 100, 20)
			budget := uint64(1)
			switch kind {
			case "budget":
				budget = 0
			case "mode":
				v.mem[0x449] = 0xd
			case "segment":
				v.s.Seg[cpu.ES] = 0x9000
			case "direction":
				v.s.Flags ^= cpu.DF
			case "count":
				v.s.R[cpu.CX] = 321
			case "row":
				v.s.R[cpu.DI] = 319
				if backward {
					v.s.R[cpu.DI] = 0
				}
			case "end":
				v.s.R[cpu.DI] = 64000
			case "code":
				v.mem[0x48a10+0x5593] ^= 1
			case "cs":
				v.s.Seg[cpu.CS]--
			case "ip":
				v.s.IP++
			case "width-zero":
				putRowWord(v, 0x48a10+0xedd, 0)
			case "width-high":
				putRowWord(v, 0x48a10+0xedd, 321)
			case "height-zero":
				putRowWord(v, 0x48a10+0xedf, 0)
			case "height-high":
				putRowWord(v, 0x48a10+0xedf, 201)
			}
			before, regs := append([]byte(nil), v.mem...), v.s
			if planFloodScan(v, budget) != nil {
				t.Fatal("invalid scan accepted", kind)
			}
			if !bytes.Equal(before, v.mem) || v.s != regs {
				t.Fatal("refusal mutated")
			}
		}
	}
	// Fixed prefixes stop before scanning, even when ES is an unknown source.
	for _, ip := range []uint16{0x55be, 0x55f4} {
		v := floodScanSeed(t, false, 100, 20)
		v.s.IP, v.s.Seg[cpu.ES] = ip, 0x9000
		if p := Plan(v, 1024); p != nil {
			t.Fatal("fixed scan block accepted", ip)
		}
	}
	t.Log("scan refusals", 30)
}

func TestNativeFloodFillWholeOriginal(t *testing.T) {
	cases, scans := 0, uint64(0)
	for _, budget := range []uint64{1, 13, 1024} {
		for _, stride := range []uint16{32, 160, 320} {
			for _, shape := range []string{"rectangle", "split", "hole", "narrow", "same", "outside", "left-edge", "right-edge"} {
				for _, dos := range []bool{false, true} {
					v := pointSeed(t, false, 5, 5, stride)
					v.s.IP, v.dos = 0x5414, dos
					putRowWord(v, 0x7090a, 0x002a)
					for y := 2; y <= 9; y++ {
						for x := 2; x <= 12; x++ {
							v.mem[0xa0000+y*int(stride)+x] = 0x27
						}
					}
					switch shape {
					case "split":
						for y := 2; y <= 9; y++ {
							v.mem[0xa0000+y*int(stride)+8] = 0x44
						}
					case "hole":
						for y := 6; y <= 8; y++ {
							for x := 7; x <= 9; x++ {
								v.mem[0xa0000+y*int(stride)+x] = 0x44
							}
						}
					case "narrow":
						for y := 2; y <= 9; y++ {
							for x := 2; x <= 12; x++ {
								if x != 5 {
									v.mem[0xa0000+y*int(stride)+x] = 0x44
								}
							}
						}
					case "same":
						putRowWord(v, 0x7090a, 0x0027)
					case "outside":
						putRowWord(v, 0x70906, 0xffff)
					case "left-edge":
						for y := 2; y <= 9; y++ {
							for x := 0; x < 7; x++ {
								v.mem[0xa0000+y*int(stride)+x] = 0x27
							}
						}
					case "right-edge":
						putRowWord(v, 0x70906, stride-2)
						for y := 2; y <= 9; y++ {
							for x := int(stride) - 7; x < int(stride); x++ {
								v.mem[0xa0000+y*int(stride)+x] = 0x27
							}
						}
					}
					a, b := callMachine(v), callMachine(v)
					runCall(t, a, v, false)
					used := false
					for b.Steps < 1000000 {
						if used && b.CPU.IP == 0x1234 && b.CPU.Seg[cpu.CS] == 0x4567 {
							break
						}
						used = true
						w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
						p := Plan(w, budget)
						if w.CS() == 0x48a1 && (w.IP() == 0x55cd || w.IP() == 0x5602 || w.IP() == 0x2cf4) && (p == nil || !p.Algorithmic) {
							t.Fatal("pixel read fell back", w.IP(), budget)
						}
						if p != nil {
							if w.CS() == 0x48a1 && (w.IP() == 0x55cd || w.IP() == 0x5602) {
								if !p.Algorithmic {
									t.Fatal("fixed search")
								}
								scans++
							}
							if err := b.CommitHostBlock(p); err != nil {
								t.Fatal(err)
							}
						} else {
							if err := b.Step(); err != nil {
								t.Fatal(err)
							}
						}
					}
					if b.CPU.IP != 0x1234 || b.CPU.Seg[cpu.CS] != 0x4567 {
						t.Fatal("fill did not return")
					}
					if a.CPU.R != b.CPU.R || a.CPU.Seg != b.CPU.Seg || a.CPU.IP != b.CPU.IP || a.CPU.Flags != b.CPU.Flags || a.CPU.Cycles != b.CPU.Cycles || a.Steps != b.Steps || !bytes.Equal(a.Mem, b.Mem) {
						t.Fatal("whole flood fill differs", shape, dos)
					}
					paletteMachinesEqual(t, a, b)
					cases++
				}
			}
		}
	}
	if scans == 0 {
		t.Fatal("no scan reached")
	}
	t.Log("whole flood calls", cases, "Go scans", scans)
}

func TestNativePointReadBoundaryOriginal(t *testing.T) {
	cases := 0
	for _, stride := range []uint16{32, 160, 320} {
		for _, xy := range [][2]uint16{{0, 0}, {1, 99}, {stride - 1, 199}, {stride / 2, 25}} {
			for _, dos := range []bool{false, true} {
				for _, df := range []bool{false, true} {
					v := pointSeed(t, true, xy[0], xy[1], stride)
					v.dos = dos
					if df {
						v.s.Flags |= cpu.DF
					}
					v.mem[0xa0000+uint32(xy[1])*uint32(stride)+uint32(xy[0])] = 0xe7
					m := original(v)
					for m.CPU.IP != 0x2cf4 && m.Steps < 100 {
						if err := m.Step(); err != nil {
							t.Fatal(err)
						}
					}
					if m.CPU.IP != 0x2cf4 {
						t.Fatal("point boundary not reached")
					}
					w := &view{mem: m.Mem, s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: dos}
					p := Plan(w, 1)
					if p == nil || !p.Algorithmic {
						t.Fatal("point read not native")
					}
					compare(t, w, p)
					for _, kind := range []string{"budget", "code", "surface", "pixel", "ax", "stack"} {
						q := &view{mem: append([]byte(nil), w.mem...), s: w.s, dos: dos}
						budget := uint64(1)
						switch kind {
						case "budget":
							budget = 0
						case "code":
							q.mem[0x48a10+0x2c04] ^= 1
						case "surface":
							putRowWord(q, 0x48a10+0xedd, 0)
						case "pixel":
							q.s.R[cpu.BX]++
						case "ax":
							q.s.R[cpu.AX] |= 0x100
						case "stack":
							q.s.R[cpu.SP]++
						}
						before, regs := append([]byte(nil), q.mem...), q.s
						if planPointReadPixel(q, budget) != nil {
							t.Fatal("bad point boundary accepted", kind)
						}
						if !bytes.Equal(before, q.mem) || q.s != regs {
							t.Fatal("point refusal mutated")
						}
					}
					cases++
				}
			}
		}
	}
	t.Log("point boundaries", cases, "refusals", cases*6)
}
