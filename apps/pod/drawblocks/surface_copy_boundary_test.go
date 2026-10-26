package drawblocks

import (
	"bytes"
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func TestNativeCopyBoundaries(t *testing.T) {
	seed := copySurfaceSeed(t, 1, 1, false, false)
	var code [593]byte
	copyView(seed, 0x48a10+0x3f5e, code[:])
	if sha256.Sum256(code[:]) != surfaceCopyCode {
		t.Fatal("source fingerprint changed")
	}
	cases := 0
	t.Logf("copy source593 sha256 %x", surfaceCopyCode)
	for _, links := range []uint16{1, 3} {
		for _, larger := range []uint16{0, 1} {
			for _, bx := range []bool{false, true} {
				for _, by := range []bool{false, true} {
					for _, dos := range []bool{false, true} {
						v := copySurfaceSeed(t, 8, 5, bx, by)
						v.dos = dos
						for i := uint16(2); i <= links+1; i++ {
							seg, flags := uint16(0xa000), uint16(0x80)
							if i < links+1 {
								seg, flags = (i+1)*2, 0xc0
							}
							for _, w := range [][2]uint16{{0xbd, seg}, {0x2bd, 0}, {0x4bd, 0}, {0x6bd, 0}, {0x8bd, 319}, {0xabd, 199}, {0xcbd, flags}} {
								putRowWord(v, 0x48a10+uint32(w[0]+i*2), w[1])
							}
						}
						for _, off := range []uint16{0x8bd, 0xabd} {
							at := 0x48a10 + uint32(off+larger*2)
							putRowWord(v, at, v.Read16(at)+3)
						}
						p := planSurfaceCopyPrefix(v)
						if p == nil {
							t.Fatal("copy parent prefix refused")
						}
						compare(t, v, p)
						cases++
						row := copyRowState(t, v, 0)
						row.dos = dos
						p = planSurfaceCopyRows(row, 1000)
						if p == nil {
							t.Fatal("copy parent row refused")
						}
						compare(t, row, p)
						cases++
					}
				}
			}
		}
	}
	t.Log("copy parent and unequal surface comparisons", cases)
}
func TestNativeCopyRefusals(t *testing.T) {
	seed := copySurfaceSeed(t, 1, 1, false, false)
	var code [593]byte
	copyView(seed, 0x48a10+0x3f5e, code[:])
	if sha256.Sum256(code[:]) != surfaceCopyCode {
		t.Fatal("source fingerprint changed")
	}
	cases := 0
	for _, bx := range []bool{false, true} {
		base := copyRowState(t, copySurfaceSeed(t, 8, 5, bx, true), 0)
		for _, kind := range []string{"budget", "mode", "CS", "IP", "DS", "ES", "DF", "code", "SP", "BP", "stack-video", "slot", "width", "height", "remaining", "direction", "reverse", "SI", "DI", "source-table", "root-stride", "scratch"} {
			v := &view{mem: append([]byte(nil), base.mem...), s: base.s}
			budget := uint64(1000)
			stack := uint32(v.SS())*16 + uint32(v.SP())
			switch kind {
			case "budget":
				budget = 0
			case "mode":
				v.mem[0x449] = 0xd
			case "CS":
				v.s.Seg[cpu.CS]++
			case "IP":
				v.s.IP++
			case "DS":
				v.s.Seg[cpu.DS]--
			case "ES":
				v.s.Seg[cpu.ES]--
			case "DF":
				v.s.Flags ^= cpu.DF
			case "code":
				v.mem[0x48a10+0x3f5e] ^= 1
			case "SP":
				v.s.R[cpu.SP] = 65535
			case "BP":
				v.s.R[cpu.BP]++
			case "stack-video":
				v.s.Seg[cpu.SS] = 0xa000
			case "slot":
				putRowWord(v, stack+12, 256)
			case "width":
				putRowWord(v, 0x48a10+0x1135, 0)
			case "height":
				putRowWord(v, 0x48a10+0xecd, 0)
			case "remaining":
				putRowWord(v, 0x48a10+0x1133, 65535)
			case "direction":
				putRowWord(v, 0x48a10+0x1147, 0)
			case "reverse":
				putRowWord(v, 0x48a10+0x1139, 2)
			case "SI":
				v.s.R[cpu.SI]++
			case "DI":
				v.s.R[cpu.DI]++
			case "source-table":
				putRowWord(v, 0x48a10+0x8bd, 0)
			case "root-stride":
				putRowWord(v, 0x48a10+0xedd, 0)
			case "scratch":
				putRowWord(v, 0x48a10+0x1141, 65535)
			}
			before := append([]byte(nil), v.mem...)
			regs := v.s
			if planSurfaceCopyRows(v, budget) != nil {
				t.Fatal("copy invalid accepted", bx, kind)
			}
			if !bytes.Equal(before, v.mem) || v.s != regs {
				t.Fatal("copy refusal mutated")
			}
			cases++
		}
	}
	t.Log("copy row refusals", cases)
}
