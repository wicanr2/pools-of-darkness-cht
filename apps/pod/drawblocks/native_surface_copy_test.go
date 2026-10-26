package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func copySurfaceSeed(t *testing.T, width, height uint16, backX, backY bool) *view {
	v := axisSeed(t, [4]uint16{20, 10, 30, 20}, 320)
	v.s.IP = 0x3f5e
	stack := uint32(v.SS())*16 + uint32(v.SP())
	putRowWord(v, stack+4, 0)
	putRowWord(v, stack+6, 1)
	sx, sy, dx, dy := uint16(20), uint16(10), uint16(18), uint16(8)
	if backX {
		dx = 22
	}
	if backY {
		dy = 12
	}
	for i := uint16(0); i < 3; i++ {
		left, top, right, bottom, seg, flags := sx, sy, sx+width-1, sy+height-1, uint16(4), uint16(0xc0)
		if i == 1 {
			left, top, right, bottom = dx, dy, dx+width-1, dy+height-1
		}
		if i == 2 {
			left, top, right, bottom, seg, flags = 0, 0, 319, 199, 0xa000, 0x80
		}
		for _, w := range [][2]uint16{{0xbd, seg}, {0x2bd, 0}, {0x4bd, left}, {0x6bd, top}, {0x8bd, right}, {0xabd, bottom}, {0xcbd, flags}} {
			putRowWord(v, 0x48a10+uint32(w[0]+i*2), w[1])
		}
	}
	for i := uint32(0); i < 65536; i++ {
		v.mem[0xa0000+i] = byte(i*37 + i/320)
	}
	return v
}
func copyRowState(t *testing.T, v *view, skip int) *view {
	m := original(v)
	index := 0
	for m.Steps < 10000 {
		if m.CPU.Seg[cpu.CS] == 0x48a1 && m.CPU.IP == 0x415e {
			if index == skip {
				return &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}}
			}
			index++
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("copy row absent")
	return nil
}
func TestNativeCopyRows(t *testing.T) {
	v := copySurfaceSeed(t, 1, 1, false, false)
	var b [593]byte
	copyView(v, 0x48a10+0x3f5e, b[:])
	if sha256.Sum256(b[:]) != surfaceCopyCode {
		t.Fatal("source fingerprint changed")
	}
	cases := 0
	for _, width := range []uint16{1, 2, 3, 7, 8, 9, 20} {
		for _, height := range []uint16{1, 2, 5} {
			for _, bx := range []bool{false, true} {
				for _, by := range []bool{false, true} {
					for _, skip := range []int{0, int(height) - 1} {
						seed := copyRowState(t, copySurfaceSeed(t, width, height, bx, by), skip)
						for _, dos := range []bool{false, true} {
							for _, budget := range []uint64{17, 24, 51, 1000} {
								w := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
								p := planSurfaceCopyRows(w, budget)
								if p == nil {
									if budget >= 24 {
										t.Fatal("copy row refused", width, height, bx, by, skip, budget)
									}
									continue
								}
								t.Run("copy", func(t *testing.T) { compare(t, w, p) })
								if t.Failed() {
									t.Fatal("copy row mismatch", width, height, bx, by, skip, budget, dos, p.Instructions, p.Cycles)
								}
								cases++
							}
						}
					}
				}
			}
		}
	}
	t.Log("copy row independent comparisons", cases)
}
