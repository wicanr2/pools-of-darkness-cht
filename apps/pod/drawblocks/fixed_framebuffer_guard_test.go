package drawblocks

import (
	"bytes"
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

// The original character-record routine is compatibility work, not a raster
// kernel. An unknown framebuffer alias must not turn it into "native drawing".
func TestFixedCharacterRecordFramebufferAlias(t *testing.T) {
	for _, address := range []uint32{0x70002, 0x9ffff, 0xa0000, 0xa0001, 0xaf9fe} {
		for _, df := range []bool{false, true} {
			for _, dos := range []bool{false, true} {
				v := &view{mem: input(t), dos: dos, s: machine.HostRegisters{
					R:   [8]uint16{0x7654, 13, 17, 19, 0xf000, 0xbeef, 23, 29},
					Seg: [4]uint16{0x1234, 0x48a1, 0x2000, 0x4321}, IP: 0x5e56, Flags: 0x246}}
				if df {
					v.s.Flags |= cpu.DF
				}
				stack := uint32(v.SS())*16 + uint32(v.SP())
				putRowWord(v, stack, 0x7777)
				putRowWord(v, stack+2, 0x3333)
				putRowWord(v, stack+4, 0x12ab)
				putRowWord(v, 0x48a10+0x110b, 0x100)
				putRowWord(v, 0x48a10+0x110d, 0x6000)
				putRowWord(v, 0x60114, uint16(address&15))
				putRowWord(v, 0x60116, uint16(address>>4))
				before := append([]byte(nil), v.mem...)
				registers := v.s
				p := Plan(v, 200)
				if !bytes.Equal(before, v.mem) || v.s != registers {
					t.Fatal("planner changed its input")
				}
				if address == 0x70002 {
					if p == nil || p.Algorithmic {
						t.Fatal("ordinary character record lost compatibility work")
					}
					compare(t, v, p)
					continue
				}
				if p != nil && fixedFramebufferWrites(p.Writes, p.Transfers) {
					t.Fatalf("fixed framebuffer alias accepted: %x", address)
				}
				// The declined pixels still have the unmodified original path.
				m := original(v)
				for n := 0; n < 16; n++ {
					if err := m.Step(); err != nil {
						t.Fatal(err)
					}
				}
				if m.Mem[address] != 0xab || m.Mem[address+1] != 0 {
					t.Fatal("original fallback did not preserve character record")
				}
			}
		}
	}
}
