package pod

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type glyphView struct {
	memory []byte
	before machine.HostRegisters
	step   uint64
	dos    bool
}

func (v *glyphView) Read8(a uint32) byte    { return v.memory[a&0xfffff] }
func (v *glyphView) Read16(a uint32) uint16 { return uint16(v.Read8(a)) | uint16(v.Read8(a+1))<<8 }
func (v *glyphView) Steps() uint64          { return v.step }
func (v *glyphView) AX() uint16             { return v.before.R[0] }
func (v *glyphView) CX() uint16             { return v.before.R[1] }
func (v *glyphView) DX() uint16             { return v.before.R[2] }
func (v *glyphView) BX() uint16             { return v.before.R[3] }
func (v *glyphView) SP() uint16             { return v.before.R[4] }
func (v *glyphView) BP() uint16             { return v.before.R[5] }
func (v *glyphView) SI() uint16             { return v.before.R[6] }
func (v *glyphView) DI() uint16             { return v.before.R[7] }
func (v *glyphView) ES() uint16             { return v.before.Seg[0] }
func (v *glyphView) CS() uint16             { return v.before.Seg[1] }
func (v *glyphView) SS() uint16             { return v.before.Seg[2] }
func (v *glyphView) DS() uint16             { return v.before.Seg[3] }
func (v *glyphView) IP() uint16             { return v.before.IP }
func (v *glyphView) Flags() uint16          { return v.before.Flags }
func (v *glyphView) DOSBoxCost() bool       { return v.dos }
func glyphViews(t *testing.T) []*glyphView {
	t.Helper()
	dir := os.Getenv("DOSGOLEM_POD_GLYPH_CALLS")
	if dir == "" {
		t.Skip("未提供原版正常呼叫")
	}
	var manifest struct {
		Captures []struct {
			R         [8]uint16
			Seg       [4]uint16
			IP, Flags uint16
			Step      uint64
			Memory    string
		}
	}
	raw, e := os.ReadFile(filepath.Join(dir, "calls.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(raw, &manifest); e != nil {
		t.Fatal(e)
	}
	out := []*glyphView{}
	for _, c := range manifest.Captures {
		mem, e := os.ReadFile(filepath.Join(dir, c.Memory))
		if e != nil {
			t.Fatal(e)
		}
		out = append(out, &glyphView{memory: mem, before: machine.HostRegisters{R: c.R, Seg: c.Seg, IP: c.IP, Flags: c.Flags}, step: c.Step})
	}
	if len(out) != 32 {
		t.Fatal("capture count", len(out))
	}
	return out
}
func glyphMachine(v *glyphView, model cpu.Model) *machine.Machine {
	m := machine.New()
	m.SetVideoMode(0x13)
	copy(m.Mem, v.memory)
	m.IRQ0Every = 0
	m.VGAFrameEvery = 0
	m.CPU.R = v.before.R
	m.CPU.Seg = v.before.Seg
	m.CPU.IP = v.before.IP
	m.CPU.Flags = v.before.Flags
	m.CPU.Cycles = 0
	m.Steps = v.step
	m.CPU.Model = model
	m.CPU.DOSBoxCost = v.dos
	m.CPU.Hi = [8]uint16{1, 2, 3, 4, 5, 6, 7, 8}
	return m
}
func TestWholeGlyphIndependentOriginal(t *testing.T) {
	for _, v := range glyphViews(t) {
		for _, model := range []cpu.Model{cpu.Model8086, cpu.Model80186, cpu.Model80386} {
			for _, dosCost := range []bool{false, true} {
				v.dos = dosCost
				p := PlanGlyphCall(v, 4096)
				if p == nil {
					t.Fatal("real call refused")
				}
				a, b := glyphMachine(v, model), glyphMachine(v, model)
				var originalWrites []machine.HostWrite
				var av, bv []machine.VideoWrite
				a.WatchWrite(0, 0xfffff, func(m *machine.Machine, addr uint32, old, value byte) {
					cs, ip := m.CPU.OpAddr()
					originalWrites = append(originalWrites, machine.HostWrite{Address: addr, Value: value, Step: m.Steps - v.step, CS: cs, IP: ip})
				})
				a.ObserveVideoWrites(func(w machine.VideoWrite) { av = append(av, w) })
				b.ObserveVideoWrites(func(w machine.VideoWrite) { bv = append(bv, w) })
				for n := uint64(0); n < p.Instructions; n++ {
					if e := a.Step(); e != nil {
						t.Fatal(e)
					}
				}
				if e := b.CommitHostWork(p); e != nil {
					t.Fatal(e)
				}
				if a.CPU.R != b.CPU.R || a.CPU.Seg != b.CPU.Seg || a.CPU.IP != b.CPU.IP || a.CPU.Flags != b.CPU.Flags || a.CPU.Hi != b.CPU.Hi || a.CPU.Cycles != b.CPU.Cycles || a.Steps != b.Steps || !bytes.Equal(a.Mem, b.Mem) || !reflect.DeepEqual(originalWrites, p.Writes) || !reflect.DeepEqual(av, bv) {
					t.Fatalf("whole glyph differs model%d dos%v", model, dosCost)
				}
				ac, ai := a.CPU.OpAddr()
				bc, bi := b.CPU.OpAddr()
				if ac != bc || ai != bi {
					t.Fatal("last instruction identity differs")
				}
			}
		}
	}
}
func TestWholeGlyphRefusalsAndEmptyReturn(t *testing.T) {
	base := glyphViews(t)[0]
	p := PlanGlyphCall(base, 4096)
	if p == nil {
		t.Fatal("initial")
	}
	for _, which := range []string{"code", "mode", "x", "y", "stack", "alias", "font-video", "entry", "budget"} {
		v := *base
		v.memory = append([]byte(nil), base.memory...)
		sp := int(v.SS())*16 + int(v.SP())
		slot := int(v.DS())*16 + 0x9d86
		budget := uint64(4096)
		switch which {
		case "code":
			v.memory[0x1c486] ^= 1
		case "mode":
			v.memory[0x449] = 3
		case "x":
			binary.LittleEndian.PutUint16(v.memory[sp+12:], 40)
		case "y":
			binary.LittleEndian.PutUint16(v.memory[sp+10:], 25)
		case "stack":
			v.before.R[4] = 4
		case "entry":
			v.before.IP++
		case "budget":
			budget = p.Instructions - 1
		case "font-video":
			binary.LittleEndian.PutUint16(v.memory[slot+2:], 0xa000)
		case "alias":
			v.memory[sp+4] = 0
			binary.LittleEndian.PutUint16(v.memory[slot:], v.SP()-6)
			binary.LittleEndian.PutUint16(v.memory[slot+2:], v.SS())
		}
		before := append([]byte(nil), v.memory...)
		if PlanGlyphCall(&v, budget) != nil || !bytes.Equal(before, v.memory) {
			t.Fatal("unsafe plan", which)
		}
	}
	original, empty := glyphMachine(base, cpu.Model80386), glyphMachine(base, cpu.Model80386)
	for n := uint64(0); n < p.Instructions; n++ {
		if e := original.Step(); e != nil {
			t.Fatal(e)
		}
	}
	empty.CPU.IP = 0x1200
	if e := empty.Step(); e != nil {
		t.Fatal(e)
	}
	if bytes.Equal(original.Mem, empty.Mem) || original.CPU.R == empty.CPU.R {
		t.Fatal("empty return negative control did not detect missing effects")
	}
}

func TestWholeGlyphBoundaryStates(t *testing.T) {
	base := glyphViews(t)[0]
	for _, pos := range [][2]uint16{{0, 0}, {39, 0}, {0, 24}, {39, 24}} {
		for _, fill := range []byte{0, 255} {
			for _, dosCost := range []bool{false, true} {
				v := *base
				v.memory = append([]byte(nil), base.memory...)
				v.dos = dosCost
				stack := int(v.SS())*16 + int(v.SP())
				slot := int(v.DS())*16 + 0x9d86
				v.memory[stack+4] = 0
				v.memory[stack+6] = 255
				v.memory[stack+8] = 128
				binary.LittleEndian.PutUint16(v.memory[stack+10:], pos[1])
				binary.LittleEndian.PutUint16(v.memory[stack+12:], pos[0])
				binary.LittleEndian.PutUint16(v.memory[slot:], 0xfff8)
				binary.LittleEndian.PutUint16(v.memory[slot+2:], 0x6000)
				for i := 0; i < 8; i++ {
					v.memory[0x6fff8+i] = fill
				}
				p := PlanGlyphCall(&v, 4096)
				if p == nil {
					t.Fatal("valid boundary rejected")
				}
				a, b := glyphMachine(&v, cpu.Model80386), glyphMachine(&v, cpu.Model80386)
				for n := uint64(0); n < p.Instructions; n++ {
					if e := a.Step(); e != nil {
						t.Fatal(e)
					}
				}
				if e := b.CommitHostWork(p); e != nil {
					t.Fatal(e)
				}
				if a.CPU.R != b.CPU.R || a.CPU.Flags != b.CPU.Flags || a.CPU.Cycles != b.CPU.Cycles || !bytes.Equal(a.Mem, b.Mem) {
					t.Fatal("boundary differs", pos, fill, dosCost)
				}
			}
		}
	}
}
