// Disposable whole-call model; compare against an independent original CPU.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"math/bits"
	"os"
	"path/filepath"
	"reflect"
)

type state struct {
	R         [8]uint16
	Seg       [4]uint16
	IP, Flags uint16
}
type capture struct {
	R                                [8]uint16
	Seg                              [4]uint16
	IP, Flags                        uint16
	Step                             uint64
	Memory, MemorySHA256, CodeSHA256 string
}
type effect struct {
	Address uint32
	Value   byte
	Step    uint64
	CS, IP  uint16
}
type plan struct {
	After                           state
	Instructions, Cycles, DOSCycles uint64
	Writes                          []effect
}

func glyphPlan(mem []byte, b state) (*plan, error) {
	fail := func() (*plan, error) { return nil, fmt.Errorf("unsupported glyph call") }
	if len(mem) != (1<<20) || b.Seg[cpu.CS] != 0x1b2e || b.IP != 0x11a6 || mem[0x449] != 0x13 {
		return fail()
	}
	if fmt.Sprintf("%x", sha256.Sum256(mem[0x1c486:0x1c4e3])) != "d1c978b914f28d1a978a2cae1b7a1c52f4b5d233176e8bd6c3029b05c9b4958b" {
		return fail()
	}
	sp := int(b.R[cpu.SP])
	sb := int(b.Seg[cpu.SS]) * 16
	ds := int(b.Seg[cpu.DS]) * 16
	if sp < 6 || sp+14 > 65536 || sb+sp+14 > 0xa0000 || ds+0x9d8a > 0xa0000 {
		return fail()
	}
	word := func(i int) uint16 { return binary.LittleEndian.Uint16(mem[i:]) }
	stack := sb + sp
	ch, bg, fg := mem[stack+4], mem[stack+6], mem[stack+8]
	y, x := word(stack+10), word(stack+12)
	if x >= 40 || y >= 25 {
		return fail()
	}
	baseOff := int(word(ds + 0x9d86))
	off := (baseOff + int(ch)*8) & 65535
	seg := word(ds + 0x9d88)
	src := int(seg)*16 + off
	if off+8 > 65536 || src+8 > 0xa0000 {
		return fail()
	}
	overlap := func(a, b, c, d int) bool { return a < d && c < b }
	if overlap(stack-6, stack+14, src, src+8) || overlap(stack-6, stack+14, 0x1c486, 0x1c4e3) || overlap(stack-6, stack+14, ds+0x9d86, ds+0x9d8a) {
		return fail()
	}
	dest := (int(y)*320 + int(x)) * 8
	p := &plan{After: b, Writes: make([]effect, 0, 84)}
	emit := func(addr int, value byte, n uint64, ip uint16) {
		p.Writes = append(p.Writes, effect{uint32(addr), value, n, b.Seg[cpu.CS], ip})
	}
	push := func(addr int, value uint16, n uint64, ip uint16) {
		emit(addr, byte(value), n, ip)
		emit(addr+1, byte(value>>8), n, ip)
	}
	push(stack-2, b.R[cpu.BP], 1, 0x11a6)
	push(stack-4, b.Seg[cpu.DS], 4, 0x11aa)
	n := uint64(21)
	ones := 0
	for row := 0; row < 8; row++ {
		n++
		push(stack-6, uint16(8-row), n, 0x11d7)
		n += 2
		for col := 0; col < 8; col++ {
			on := mem[src+row]&(0x80>>col) != 0
			value, ip := bg, uint16(0x11ea)
			if on {
				value, ip = fg, 0x11f1
				ones++
			}
			n += 6
			emit(0xa0000+dest+row*320+col, value, n, ip)
			if !on {
				n++
			}
			n += 2
		}
		n += 4
	}
	n += 3
	p.Instructions = n
	p.Cycles = 2620 - uint64(2*ones)
	p.DOSCycles = n + 64
	r := &p.After.R
	r[cpu.AX] = uint16(bg)
	if mem[src+7]&1 != 0 {
		r[cpu.AX] = uint16(fg)
	}
	r[cpu.CX] = 0
	r[cpu.DX] = 0
	r[cpu.BX] = 0
	r[cpu.SI] = uint16(off + 8)
	r[cpu.DI] = uint16(dest + 2560)
	r[cpu.SP] += 14
	p.After.Seg[cpu.ES] = 0xa000
	p.After.IP = word(stack)
	p.After.Seg[cpu.CS] = word(stack + 2)
	a, add := uint16(dest+7*320+8), uint16(312)
	v := a + add
	f := b.Flags &^ uint16(0xcd5|cpu.DF)
	if uint32(a)+uint32(add) > 65535 {
		f |= cpu.CF
	}
	if (a^add^v)&16 != 0 {
		f |= cpu.AF
	}
	if v == 0 {
		f |= cpu.ZF
	}
	if v&0x8000 != 0 {
		f |= cpu.SF
	}
	if (^(a^add)&(a^v))&0x8000 != 0 {
		f |= cpu.OF
	}
	if bits.OnesCount8(byte(v))%2 == 0 {
		f |= cpu.PF
	}
	p.After.Flags = f
	return p, nil
}
func must(e error) {
	if e != nil {
		panic(e)
	}
}
func read(p string) []byte { b, e := os.ReadFile(p); must(e); return b }
func main() {
	root := flag.String("root", "", "call capture")
	out := flag.String("out", "", "new report")
	flag.Parse()
	if _, e := os.Stat(*out); !os.IsNotExist(e) {
		panic("report exists")
	}
	var input struct{ Captures []capture }
	must(json.Unmarshal(read(filepath.Join(*root, "calls.json")), &input))
	rows := []any{}
	for _, c := range input.Captures {
		mem := read(filepath.Join(*root, c.Memory))
		if fmt.Sprintf("%x", sha256.Sum256(mem)) != c.MemorySHA256 {
			panic("memory changed")
		}
		before := state{c.R, c.Seg, c.IP, c.Flags}
		p, e := glyphPlan(mem, before)
		must(e)
		for _, model := range []cpu.Model{cpu.Model8086, cpu.Model80186, cpu.Model80386} {
			for _, dosCost := range []bool{false, true} {
				m := machine.New()
				m.SetVideoMode(0x13)
				copy(m.Mem, mem)
				m.IRQ0Every = 0
				m.VGAFrameEvery = 0
				m.Steps = c.Step
				m.CPU.R = c.R
				m.CPU.Seg = c.Seg
				m.CPU.IP = c.IP
				m.CPU.Flags = c.Flags
				m.CPU.Model = model
				m.CPU.DOSBoxCost = dosCost
				m.CPU.Cycles = 0
				writes := []effect{}
				m.WatchWrite(0, 0xfffff, func(m *machine.Machine, a uint32, old, v byte) {
					cs, ip := m.CPU.OpAddr()
					writes = append(writes, effect{a, v, m.Steps - c.Step, cs, ip})
				})
				count := uint64(0)
				costs := map[uint16]uint64{}
				for count < 2000 {
					ip, cy := m.CPU.IP, m.CPU.Cycles
					must(m.Step())
					count++
					costs[ip] = m.CPU.Cycles - cy
					if m.CPU.IP == p.After.IP && m.CPU.Seg[cpu.CS] == p.After.Seg[cpu.CS] {
						break
					}
				}
				after := state{m.CPU.R, m.CPU.Seg, m.CPU.IP, m.CPU.Flags}
				wantMem := append([]byte(nil), mem...)
				for _, w := range p.Writes {
					wantMem[w.Address] = w.Value
				}
				cycles := p.Cycles
				if dosCost {
					cycles = p.DOSCycles
				}
				if count != p.Instructions || cycles != m.CPU.Cycles || after != p.After || !bytes.Equal(wantMem, m.Mem) || !reflect.DeepEqual(writes, p.Writes) {
					fmt.Printf("case %s model%d dos%v count%d/%d cycles%d/%d CPU%+v/%+v RAM%v writes%v costs%v\n", c.Memory, model, dosCost, count, p.Instructions, m.CPU.Cycles, cycles, after, p.After, bytes.Equal(wantMem, m.Mem), reflect.DeepEqual(writes, p.Writes), costs)
					panic("whole-call model mismatch")
				}
				rows = append(rows, struct {
					Source               string
					Model                cpu.Model
					DOSCost              bool
					Instructions, Cycles uint64
					Writes               int
					MemorySHA256         string
				}{c.Memory, model, dosCost, count, cycles, len(writes), fmt.Sprintf("%x", sha256.Sum256(wantMem))})
			}
		}
	}
	b, e := json.MarshalIndent(struct {
		Scope string
		Cases []any
	}{"真實正常DEMO入口的獨立函式級原版執行；三模型與兩種週期，完整RAM／CPU／返回／堆疊／有序寫入；不取代正常玩家驗收", rows}, "", "  ")
	must(e)
	must(os.WriteFile(*out, append(b, '\n'), 0600))
	fmt.Println("matched", len(rows))
}
