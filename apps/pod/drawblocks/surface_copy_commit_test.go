package drawblocks

import (
	"bytes"
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeCopyWhole(t *testing.T) {
	seed := copySurfaceSeed(t, 1, 1, false, false)
	var code [593]byte
	copyView(seed, 0x48a10+0x3f5e, code[:])
	if sha256.Sum256(code[:]) != surfaceCopyCode {
		t.Fatal("source fingerprint changed")
	}
	cases, plans := 0, 0
	for _, width := range []uint16{1, 3, 8, 20} {
		for _, height := range []uint16{1, 5} {
			for _, bx := range []bool{false, true} {
				for _, by := range []bool{false, true} {
					for _, dos := range []bool{false, true} {
						for _, budget := range []uint64{24, 256, 10000} {
							v := copySurfaceSeed(t, width, height, bx, by)
							v.dos = dos
							a, b := callMachine(v), callMachine(v)
							runCall(t, a, v, false)
							used := 0
							for b.Steps < 1000000 {
								if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
									break
								}
								w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
								var p *machine.HostWorkPlan
								p = Plan(w, budget)
								if p != nil {
									if err := b.CommitHostBlock(p); err != nil {
										t.Fatal(err)
									}
									used++
								} else {
									if w.CS() == 0x48a1 && (w.IP() == 0x4169 || w.IP() == 0x4177) {
										t.Fatal("copy pixels executed by CPU")
									}
									if err := b.Step(); err != nil {
										t.Fatal(err)
									}
								}
							}
							if used == 0 {
								t.Fatal("copy no Go plans")
							}
							var sa, sb bytes.Buffer
							if err := a.SaveState(&sa); err != nil {
								t.Fatal(err)
							}
							if err := b.SaveState(&sb); err != nil {
								t.Fatal(err)
							}
							if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
								t.Fatal("copy whole differs", width, height, bx, by, dos, budget)
							}
							cases++
							plans += used
						}
					}
				}
			}
		}
	}
	t.Log("copy whole machine comparisons", cases, "Go plans", plans)
}
