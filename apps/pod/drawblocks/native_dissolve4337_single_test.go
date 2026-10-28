// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeDissolve4337SinglePixelContinuation(t *testing.T) {
	cases := 0
	for _, dos := range []bool{false, true} {
		for _, df := range []bool{false, true} {
			v := dissolve4337Seed(t, 1, 1)
			v.dos = dos
			if df {
				v.s.Flags |= cpu.DF
			}
			a, b := callMachine(v), callMachine(v)
			a.RunLimit, b.RunLimit = 10000, 10000
			for a.Steps < 10000 {
				if a.Steps > 0 && a.CPU.Seg[cpu.CS] == 0x4567 && a.CPU.IP == 0x1234 {
					t.Fatal("one-pixel original unexpectedly returned")
				}
				if err := a.Step(); err != nil {
					t.Fatal(err)
				}
			}
			for b.Steps < 10000 {
				if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
					t.Fatal("one-pixel native changed termination")
				}
				w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
				budget := uint64(4096)
				if 10000-b.Steps < budget {
					budget = 10000 - b.Steps
				}
				p := dissolve4337(w, budget)
				if p != nil {
					if err := b.CommitHostBlock(p); err != nil {
						t.Fatal(err)
					}
				} else {
					if w.CS() == 0x48a1 && (w.IP() == 0x4569 || w.IP() == 0x45bb) {
						t.Fatal("one-pixel CPU raster")
					}
					if err := b.Step(); err != nil {
						t.Fatal(err)
					}
				}
			}
			var sa, sb bytes.Buffer
			if err := a.SaveState(&sa); err != nil {
				t.Fatal(err)
			}
			if err := b.SaveState(&sb); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
				t.Fatal("one-pixel continuation changed full state", dos, df)
			}
			if b.Mem[0x48a10+0x1171] != 1 || b.Mem[0x48a10+0x1172] != 0 {
				t.Fatal("one-pixel remaining changed")
			}
			cases++
		}
	}
	t.Log("dissolve4337 one-pixel bounded continuation comparisons", cases, "cutoff", 10000, "seed", 1)
}
