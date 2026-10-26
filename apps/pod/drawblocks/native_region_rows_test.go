package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func regionRowStates(t *testing.T, sub bool) *view {
	seed := regionSeed(t, [4]uint16{17, 19, 33, 22}, 320, sub)
	m := original(seed)
	for m.Steps < 1000 {
		if m.CPU.Seg[cpu.CS] == 0x48a1 && m.CPU.IP == 0x3418 {
			return &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}}
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("row state absent")
	return nil
}

func TestNativeRegionRowsOriginal(t *testing.T) {
	cases := 0
	for _, sub := range []bool{false, true} {
		seed := regionRowStates(t, sub)
		for _, y := range []uint16{19, 20, 22, 23} {
			for _, dos := range []bool{false, true} {
				for _, budget := range []uint64{7, 9, 24, 25, 26, 27, 28, 50, 250} {
					v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
					v.s.R[cpu.SI] = y
					p := planRegionRows(v, budget)
					if p == nil {
						if budget >= 27 {
							t.Fatal("row unexpectedly rejected", sub, y, dos, budget)
						}
						continue
					}
					compare(t, v, p)
					cases++
				}
			}
		}
	}
	t.Log("region row original comparisons", cases)
}

func TestNativeRegionRowsRefusals(t *testing.T) {
	seed := regionRowStates(t, false)
	cases := 0
	for _, kind := range []string{"budget", "mode", "code", "CS", "DS", "ES", "DF", "BP", "SP", "stack-video", "left", "stride", "color", "cursor", "reverse"} {
		v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
		b := uint64(250)
		stack := uint32(v.SS())*16 + uint32(v.SP())
		switch kind {
		case "budget":
			b = 0
		case "mode":
			v.mem[0x449] = 0xd
		case "code":
			v.mem[0x48a10+0x3338] ^= 1
		case "CS":
			v.s.Seg[cpu.CS] = 0x4800
		case "DS":
			v.s.Seg[cpu.DS] = 0x5000
		case "ES":
			v.s.Seg[cpu.ES] = 0x9000
		case "DF":
			v.s.Flags |= cpu.DF
		case "BP":
			v.s.R[cpu.BP]++
		case "SP":
			v.s.R[cpu.SP] = 65535
		case "stack-video":
			v.s.Seg[cpu.SS] = 0xa000
		case "left":
			putRowWord(v, 0x48a10+0x113b, 0)
		case "stride":
			putRowWord(v, 0x48a10+0xedd, 0)
		case "color":
			putRowWord(v, 0x48a10+0x1129, 0)
		case "cursor":
			v.s.R[cpu.SI] = 0
		case "reverse":
			putRowWord(v, stack+14, 100)
		}
		before := append([]byte(nil), v.mem...)
		regs := v.s
		if planRegionRows(v, b) != nil {
			t.Fatal("invalid region row accepted", kind)
		}
		if string(before) != string(v.mem) || regs != v.s {
			t.Fatal("refusal mutated")
		}
		cases++
	}
	t.Log("region row refusals", cases)
}

// ABI prefix plus a typed row cursor. The full Go plan supplies the already
// verified prefix notices; no pixel transfer from it is submitted twice.
func TestNativeRegionPrefixOriginal(t *testing.T) {
	cases := 0
	for _, sub := range []bool{false, true} {
		for _, dos := range []bool{false, true} {
			for _, xy := range [][4]uint16{{0, 0, 319, 199}, {17, 19, 33, 22}, {9, 11, 51, 31}} {
				for _, budget := range []uint64{70, 71, 80, 1024, 10000} {
					v := regionSeed(t, xy, 320, sub)
					v.dos = dos
					p := planRegionBounded(v, budget)
					if p == nil {
						if budget >= 80 {
							t.Fatal("prefix rejected", sub, dos, xy, budget)
						}
						continue
					}
					compare(t, v, p)
					cases++
				}
			}
		}
	}
	t.Log("region bounded prefix original comparisons", cases)
}

func TestNativeRegionBoundedWhole(t *testing.T) {
	cases, jobs := 0, 0
	for _, sub := range []bool{false, true} {
		for _, xy := range [][4]uint16{{0, 0, 319, 199}, {17, 19, 33, 22}, {9, 11, 51, 31}} {
			for _, dos := range []bool{false, true} {
				for _, budget := range []uint64{27, 71, 1024, 10000} {
					v := regionSeed(t, xy, 320, sub)
					v.dos = dos
					a, b := callMachine(v), callMachine(v)
					runCall(t, a, v, false)
					used := 0
					for b.Steps < 1000000 {
						if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
							break
						}
						w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
						if p := planRegionBounded(w, budget); p != nil {
							if err := b.CommitHostBlock(p); err != nil {
								t.Fatal(err)
							}
							used++
						} else if err := b.Step(); err != nil {
							t.Fatal(err)
						}
					}
					if used == 0 {
						t.Fatal("bounded Go rectangle path absent", sub, xy, dos, budget)
					}
					var sa, sb bytes.Buffer
					if err := a.SaveState(&sa); err != nil {
						t.Fatal(err)
					}
					if err := b.SaveState(&sb); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
						t.Fatal("bounded rectangle whole state differs", sub, xy, dos, budget, a.Steps, b.Steps)
					}
					cases++
					jobs += used
				}
			}
		}
	}
	t.Log("bounded rectangle whole original comparisons", cases, "Go plans", jobs)
}
