// Typed Go image rows and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func bitmap51RowSeed(t *testing.T, width, rows uint16, planar, compressed bool) *view {
	v := axisSeed(t, [4]uint16{20, 10, 30, 20}, 320)
	originalSP := v.SP()
	frame := uint32(v.SS())*16 + uint32(originalSP) - 8
	for _, w := range [][2]uint16{{0, v.DI()}, {2, v.SI()}, {4, v.DS()}, {6, v.BP()}, {8, 0x1234}, {10, 0x4567}, {12, 0}, {14, 0x100}, {16, 0x8000}} {
		putRowWord(v, frame+uint32(w[0]), w[1])
	}
	v.s.R[cpu.SP] = originalSP - 8
	v.s.R[cpu.BP] = originalSP - 2
	v.s.Seg[cpu.DS] = 0x8000
	v.s.Seg[cpu.ES] = 0x9000
	v.s.R[cpu.SI] = 0x200
	v.s.IP = 0x5373
	v.s.Flags &^= cpu.DF
	stride := ((width + 15) / 16) * 2
	rowBytes := stride * 8
	s := rectangleSurface{2, 0, 20, 10, 20 + width - 1, 10 + rows - 1, 0xc0}
	root := rectangleSurface{0xa000, 0, 0, 0, 319, 199, 0x80}
	for _, d := range []struct {
		index uint16
		s     rectangleSurface
	}{{0, s}, {2, root}} {
		for index, value := range []uint16{d.s.seg, d.s.extra, d.s.left, d.s.top, d.s.right, d.s.bottom, d.s.flags} {
			off := []uint16{0xbd, 0x2bd, 0x4bd, 0x6bd, 0x8bd, 0xabd, 0xcbd}[index]
			putRowWord(v, 0x48a10+uint32(off+d.index), value)
		}
	}
	for index, value := range []uint16{s.seg, s.extra, s.left, s.top, s.right, s.bottom, s.flags, width, rows, root.seg, root.extra, root.left, root.top, root.right, root.bottom, root.flags, 320, 200} {
		putRowWord(v, 0x48a10+0xebd+uint32(index)*2, value)
	}
	kind, compression := uint16(1), uint16(0)
	if planar {
		kind = 0
	}
	if compressed {
		compression = 1
		v.s.Seg[cpu.DS] = 0x48a1
		v.s.R[cpu.SI] = 0x1252
	}
	for _, w := range [][2]uint16{{0x11f3, compression}, {0x11f5, kind}, {0x11f7, stride}, {0x1135, rowBytes}, {0x11fb, width}, {0x11ff, rows}, {0x1201, 0x200 + rowBytes}, {0x1214, 0x7b}} {
		putRowWord(v, 0x48a10+uint32(w[0]), w[1])
	}
	for i := uint16(0); i < rowBytes; i++ {
		v.mem[uint32(v.DS())*16+uint32(v.SI()+i)] = byte(i*37 + width)
	}
	return v
}
func TestBitmap51Chunky(t *testing.T) {
	cases := 0
	for _, width := range []uint16{1, 2, 7, 8, 9, 16, 17, 32, 160, 300} {
		for _, rows := range []uint16{1, 3} {
			for _, compressed := range []bool{false, true} {
				for _, dos := range []bool{false, true} {
					v := bitmap51RowSeed(t, width, rows, false, compressed)
					v.dos = dos
					p := planBitmap51Chunky(v, 1000)
					if p == nil {
						t.Fatal("bitmap chunky refused", width, rows, compressed, dos)
					}
					t.Run("row", func(t *testing.T) { compare(t, v, p) })
					if t.Failed() {
						t.Fatal("bitmap chunky differs", width, rows, compressed, dos, p.Instructions, p.Cycles)
					}
					cases++
				}
			}
		}
	}
	t.Log("bitmap51 chunky row original comparisons", cases)
}
