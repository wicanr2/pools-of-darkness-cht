package drawblocks

import (
	"bytes"
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func TestNativeMirrorPairsOriginal(t *testing.T) {
	seed := rleSeed(t)
	plans := 0
	for _, base := range []uint16{0x6c9e, 0x6d42, 0x6df3} {
		for stripe := uint16(0); stripe < 4; stripe++ {
			for _, count := range []uint16{1, 2, 3, 4, 5, 17, 64} {
				for _, dos := range []bool{false, true} {
					v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
					v.s.IP = base + stripe*12
					v.s.Seg[cpu.DS] = 0x5000
					v.s.Seg[cpu.ES] = 0xa000
					v.s.R[cpu.SI] = 0x100
					v.s.R[cpu.DI] = count*2 - 2
					v.s.R[cpu.CX] = count
					v.s.Flags &^= cpu.DF
					for x := uint16(0); x < count*2; x++ {
						v.mem[0x50100+int(x)] = byte(x*11 + 3)
					}
					for _, budget := range []uint64{6, 7, 13, 64, 1024} {
						p := planMirrorPairs(v, budget)
						if p == nil {
							if stripe == 3 && count > 1 && budget == 6 {
								continue
							}
							t.Fatal("valid mirrored pairs rejected")
						}
						compare(t, v, p)
						plans++
					}
				}
			}
		}
	}
	t.Log("mirror pair independent original comparisons", plans)
}

func TestNativeMirrorImageStatesOriginal(t *testing.T) {
	seed := rleSeed(t)
	states := []*view{}
	seen := map[uint16]int{}
	for _, x := range []int16{0, 7, -7} {
		v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
		stack := uint32(v.SS())*16 + uint32(v.SP())
		v.mem[stack+6], v.mem[stack+7] = byte(x), byte(uint16(x)>>8)
		v.mem[stack+10], v.mem[stack+11] = 1, 0
		m := original(v)
		retIP, retCS := v.Read16(stack), v.Read16(stack+2)
		for m.Steps < 1000000 {
			if m.Steps > 0 && m.CPU.IP == retIP && m.CPU.Seg[cpu.CS] == retCS {
				break
			}
			for _, base := range []uint16{0x6c9e, 0x6d42, 0x6df3} {
				if m.CPU.Seg[cpu.CS] == 0x48a1 && m.CPU.IP == base && seen[base] < 8 {
					seen[base]++
					states = append(states, &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}})
				}
			}
			if err := m.Step(); err != nil {
				t.Fatal(err)
			}
		}
	}
	if len(seen) != 3 {
		t.Fatal("controlled original mirror paths missing", seen)
	}
	plans := 0
	for _, v := range states {
		for _, dos := range []bool{false, true} {
			v.dos = dos
			for _, budget := range []uint64{6, 64, 1024} {
				p := planMirrorPairs(v, budget)
				if p == nil {
					t.Fatal("controlled original mirror state rejected")
				}
				compare(t, v, p)
				plans++
			}
		}
	}
	t.Log("controlled original image mirror state comparisons", plans, "states", len(states), "paths", seen)
}

func TestNativeMirrorPairRefusals(t *testing.T) {
	seed := rleSeed(t)
	for _, kind := range []string{"budget", "code", "mode", "direction", "source", "destination", "row", "counter", "entry"} {
		t.Run(kind, func(t *testing.T) {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			v.s.IP = 0x6c9e
			v.s.Seg[cpu.DS] = 0x5000
			v.s.Seg[cpu.ES] = 0xa000
			v.s.R[cpu.SI] = 0x100
			v.s.R[cpu.DI] = 100
			v.s.R[cpu.CX] = 2
			v.s.Flags &^= cpu.DF
			budget := uint64(100)
			switch kind {
			case "budget":
				budget = 5
			case "code":
				v.mem[0x48a10+0x6c9e] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "direction":
				v.s.Flags |= cpu.DF
			case "source":
				v.s.Seg[cpu.DS] = 0xa000
			case "destination":
				v.s.R[cpu.DI] = 63999
			case "row":
				v.s.R[cpu.DI] = 320
			case "counter":
				v.s.R[cpu.CX] = 0
			case "entry":
				v.s.IP++
			}
			if planMirrorPairs(v, budget) != nil {
				t.Fatal("unsupported mirror accepted")
			}
		})
	}
}

func TestNativeMirrorControlledWholeCalls(t *testing.T) {
	seed := rleSeed(t)
	cases, plans := 0, 0
	for _, x := range []int16{-7, 0, 7} {
		for _, flags := range []byte{1, 3} {
			for _, dos := range []bool{false, true} {
				v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
				stack := uint32(v.SS())*16 + uint32(v.SP())
				v.mem[stack+6], v.mem[stack+7] = byte(x), byte(uint16(x)>>8)
				v.mem[stack+10], v.mem[stack+11] = flags, 0
				for _, budget := range []uint64{6, 1024} {
					old, newer := callMachine(v), callMachine(v)
					runCall(t, old, v, false)
					retIP, retCS := v.Read16(stack), v.Read16(stack+2)
					used := 0
					for newer.Steps < 1000000 {
						if newer.Steps > 0 && newer.CPU.IP == retIP && newer.CPU.Seg[cpu.CS] == retCS {
							break
						}
						w := &view{mem: newer.Mem, s: machine.HostRegisters{R: newer.CPU.R, Seg: newer.CPU.Seg, IP: newer.CPU.IP, Flags: newer.CPU.Flags}, dos: dos}
						if p := Plan(w, budget); p != nil {
							if p.Algorithmic && mirrorPairBase(w.IP()) != 0 {
								used++
							}
							if err := newer.CommitHostBlock(p); err != nil {
								t.Fatal(err)
							}
						} else if err := newer.Step(); err != nil {
							t.Fatal(err)
						}
					}
					if used == 0 || newer.CPU.R != old.CPU.R || newer.CPU.Seg != old.CPU.Seg || newer.CPU.IP != old.CPU.IP || newer.CPU.Flags != old.CPU.Flags || newer.Steps != old.Steps || newer.CPU.Cycles != old.CPU.Cycles || !bytes.Equal(newer.Mem, old.Mem) {
						t.Fatalf("controlled whole mirror differs x%d flags%d dos%v budget%d used%d", x, flags, dos, budget, used)
					}
					cases++
					plans += used
				}
			}
		}
	}
	t.Log("controlled whole mirror original cases", cases, "committed mirror plans", plans)
}
