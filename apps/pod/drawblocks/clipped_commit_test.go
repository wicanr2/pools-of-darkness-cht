package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeClippedWhole(t *testing.T) {
	cases, plans := 0, 0
	for _, box := range [][4]int32{{0, 0, 319, 199}, {25, 15, 70, 40}, {0, 0, 10, 100}} {
		for _, xy := range [][4]int32{{-10, 0, 20, 90}, {20, 0, -10, 90}, {30, -10, 50, 100}, {50, 100, 30, -10}, {-10, 20, 100, 30}, {100, 20, -10, 30}} {
			for _, dos := range []bool{false, true} {
				for _, budget := range []uint64{18, 256, 10000} {
					v := clipLineSeed(t, xy, true)
					v.dos = dos
					for _, w := range [][2]uint16{{0x4bd, uint16(box[0])}, {0x6bd, uint16(box[1])}, {0x8bd, uint16(box[2])}, {0xabd, uint16(box[3])}} {
						putRowWord(v, 0x48a10+uint32(w[0]), w[1])
					}
					a, b := callMachine(v), callMachine(v)
					runCall(t, a, v, false)
					used := 0
					rawPixels := 0
					for b.Steps < 1000000 {
						if b.Steps > 0 && b.CPU.IP == 0x1234 && b.CPU.Seg[cpu.CS] == 0x4567 {
							break
						}
						w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
						if p := Plan(w, budget); p != nil {
							if err := b.CommitHostBlock(p); err != nil {
								t.Fatal(err)
							}
							used++
						} else {
							if w.CS() == 0x48a1 && (w.IP() == 0x309c || w.IP() == 0x30b9 || w.IP() == 0x3206 || w.IP() == 0x321c) {
								rawPixels++
							}
							if err := b.Step(); err != nil {
								t.Fatal(err)
							}
						}
					}
					if count := len(originalLinePixels(t, v)); count > 0 && (used == 0 || rawPixels > 0) {
						t.Fatalf("clipped pixels not all Go box%v xy%v budget%d used%d raw%d", box, xy, budget, used, rawPixels)
					}
					var sa, sb bytes.Buffer
					if err := a.SaveState(&sa); err != nil {
						t.Fatal(err)
					}
					if err := b.SaveState(&sb); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
						t.Fatalf("clipped whole state differs box%v xy%v budget%d dos%v steps%d/%d", box, xy, budget, dos, a.Steps, b.Steps)
					}
					cases++
					plans += used
				}
			}
		}
	}
	t.Log("clipped complete machine comparisons", cases, "Go plans", plans)
}
