package drawblocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

type view struct {
	mem []byte
	s   machine.HostRegisters
	dos bool
}

func (v *view) Read8(a uint32) byte         { return v.mem[a&0xfffff] }
func (v *view) Read16(a uint32) uint16      { return uint16(v.Read8(a)) | uint16(v.Read8(a+1))<<8 }
func (v *view) Steps() uint64               { return 0 }
func (v *view) AX() uint16                  { return v.s.R[0] }
func (v *view) CX() uint16                  { return v.s.R[1] }
func (v *view) DX() uint16                  { return v.s.R[2] }
func (v *view) BX() uint16                  { return v.s.R[3] }
func (v *view) SP() uint16                  { return v.s.R[4] }
func (v *view) BP() uint16                  { return v.s.R[5] }
func (v *view) SI() uint16                  { return v.s.R[6] }
func (v *view) DI() uint16                  { return v.s.R[7] }
func (v *view) ES() uint16                  { return v.s.Seg[0] }
func (v *view) CS() uint16                  { return v.s.Seg[1] }
func (v *view) SS() uint16                  { return v.s.Seg[2] }
func (v *view) DS() uint16                  { return v.s.Seg[3] }
func (v *view) IP() uint16                  { return v.s.IP }
func (v *view) Flags() uint16               { return v.s.Flags }
func (v *view) DOSBoxCost() bool            { return v.dos }
func (v *view) CPUModel() (cpu.Model, bool) { return cpu.Model80386, true }
func (v *view) CopyRAM(a uint32, dst []byte) bool {
	if uint64(a)+uint64(len(dst)) > 0xa0000 {
		return false
	}
	copy(dst, v.mem[int(a):int(a)+len(dst)])
	return true
}
func input(t *testing.T) []byte {
	t.Helper()
	root := os.Getenv("POD_ROOT")
	if root == "" {
		t.Skip("original fixtures absent")
	}
	raw, err := os.ReadFile(filepath.Join(root, "workplace/probe/draw-intercept-v1/legacy-demo/writer-48a1-6903.mem"))
	if err != nil {
		t.Fatal(err)
	}
	for _, seg := range []int{0x1ed4, 0x1d1f} {
		var name string
		if seg == 0x1ed4 {
			name = "writer-1ed4-1a78.mem"
		} else {
			name = "writer-1d1f-0482.mem"
		}
		m, e := os.ReadFile(filepath.Join(root, "workplace/probe/draw-intercept-v1/legacy-demo", name))
		if e != nil {
			t.Fatal(e)
		}
		copy(raw[seg*16:seg*16+65536], m[seg*16:seg*16+65536])
	}
	raw[0x449] = 0x13
	return raw
}
func original(v *view) *machine.Machine {
	m := machine.New()
	m.SetVideoMode(0x13)
	copy(m.Mem, v.mem)
	m.IRQ0Every = 0
	m.VGAFrameEvery = 0
	m.CPU.R = v.s.R
	m.CPU.Seg = v.s.Seg
	m.CPU.IP = v.s.IP
	m.CPU.Flags = v.s.Flags
	m.CPU.DOSBoxCost = v.dos
	m.CPU.Model = cpu.Model80386
	return m
}
func compare(t *testing.T, v *view, p *machine.HostWorkPlan, models ...cpu.Model) {
	t.Helper()
	m := original(v)
	if len(models) > 0 {
		m.CPU.Model = models[0]
	}
	var writes []machine.HostWrite
	m.WatchWrite(0, 0xfffff, func(m *machine.Machine, a uint32, old, val byte) {
		cs, ip := m.CPU.OpAddr()
		writes = append(writes, machine.HostWrite{Address: a, Value: val, Step: m.Steps, CS: cs, IP: ip})
	})
	for n := uint64(0); n < p.Instructions; n++ {
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	want := machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}
	expanded := expandWrites(p)
	if want != p.After || m.CPU.Cycles != p.Cycles || !reflect.DeepEqual(writes, expanded) {
		t.Fatalf("%04x:%04x dos%v N%d\nstate want%+v got%+v\ncycles want%d got%d\nwrites want%d got%d first:%s", v.CS(), v.IP(), v.dos, p.Instructions, want, p.After, m.CPU.Cycles, p.Cycles, len(writes), len(p.Writes), diffWrites(writes, p.Writes))
	}
	predicted := append([]byte(nil), v.mem...)
	for _, w := range expanded {
		predicted[w.Address] = w.Value
	}
	if !bytes.Equal(predicted, m.Mem) {
		t.Fatal("full memory differs")
	}
	cs, ip := m.CPU.OpAddr()
	if cs != p.LastCS || ip != p.LastIP {
		t.Fatal("last instruction differs")
	}
}
func expandWrites(p *machine.HostWorkPlan) []machine.HostWrite {
	out := append([]machine.HostWrite(nil), p.Writes...)
	for _, s := range p.Transfers {
		for i := 0; i < int(s.Count); i++ {
			off := int(s.Address) + i*int(s.Width)
			if s.Backward {
				off = int(s.Address) - i*int(s.Width)
			}
			for b := 0; b < int(s.Width); b++ {
				v := byte(s.Value >> uint(8*b))
				if s.Data != nil {
					v = s.Data[i*int(s.Width)+b]
				}
				out = append(out, machine.HostWrite{Address: uint32(off + b), Value: v, Step: s.Step, CS: s.CS, IP: s.IP})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Step < out[j].Step })
	return out
}
func diffWrites(a, b []machine.HostWrite) string {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return fmt.Sprintf("%d:%+v / %+v", i, a[i], b[i])
		}
	}
	return "same prefix"
}
func TestEveryBlockIndependentCPU(t *testing.T) {
	mem := input(t)
	rng := rand.New(rand.NewSource(476025))
	keys := make([]int, 0, len(blocks))
	for at := range blocks {
		keys = append(keys, int(at))
	}
	sort.Ints(keys)
	cases, rejected := 0, 0
	tested := map[uint32]bool{}
	for _, a := range keys {
		at := uint32(a)
		b := blocks[at]
		for k := 0; k < 8; k++ {
			v := &view{mem: mem, s: machine.HostRegisters{R: [8]uint16{uint16(rng.Intn(256)), uint16(rng.Intn(65)), uint16(rng.Intn(256)), uint16(rng.Intn(128)), 0xe000, 0xe010, uint16(rng.Intn(512)), uint16(rng.Intn(512))}, Seg: [4]uint16{0xa000, 0x48a1, 0x8000, 0x6000}, Flags: uint16(0xf002 | rng.Intn(4096)&0xcd5)}, dos: k%2 != 0}
			switch {
			case at >= 0x48a10+0x2772:
				v.s.Seg[1] = 0x48a1
			case at >= 0x1ed40:
				v.s.Seg[1] = 0x1ed4
			default:
				v.s.Seg[1] = 0x1d1f
			}
			v.s.IP = uint16(at - uint32(v.CS())*16)
			e := &effect{v: v, dos: v.dos, s: v.s, budget: 4096, shadow: map[uint32]byte{}}
			b.run(e)
			if e.bad {
				rejected++
				continue
			}
			e.p.After = e.s
			compare(t, v, &e.p)
			cases++
			tested[at] = true
		}
	}
	// A second controlled RAM layout supplies valid far pointers for helpers
	// whose original driver tables were not initialized in the captured frame.
	alt := append([]byte(nil), mem...)
	for i := 0x500; i < 0xa0000; i++ {
		if !protected(uint32(i)) {
			alt[i] = 0x60
		}
	}
	alt[0x449] = 0x13
	for _, a := range keys {
		at := uint32(a)
		if tested[at] {
			continue
		}
		b := blocks[at]
		for k := 0; k < 32; k++ {
			v := &view{mem: alt, dos: k%2 != 0, s: machine.HostRegisters{R: [8]uint16{uint16(rng.Intn(64)), uint16(rng.Intn(33)), uint16(rng.Intn(64)), 0, 0xe000, 0xe010, uint16(rng.Intn(256)), uint16(rng.Intn(256))}, Seg: [4]uint16{0x6000, b.cs, 0x8000, 0x6000}, IP: uint16(at - uint32(b.cs)*16), Flags: uint16(0xf002 | rng.Intn(4096)&0xcd5)}}
			e := &effect{v: v, dos: v.dos, s: v.s, budget: 4096, shadow: map[uint32]byte{}}
			b.run(e)
			if e.bad {
				rejected++
				continue
			}
			e.p.After = e.s
			compare(t, v, &e.p)
			cases++
			tested[at] = true
		}
	}
	t.Logf("independent single blocks cases=%d tested=%d total=%d refused=%d", cases, len(tested), len(blocks), rejected)
	if len(tested) < len(blocks)*9/10 {
		t.Fatal("insufficient accepted coverage")
	}
	missing := []uint32{}
	for _, a := range keys {
		if !tested[uint32(a)] {
			missing = append(missing, uint32(a))
		}
	}
	report := map[string]any{"cases": cases, "blocks": len(blocks), "tested_blocks": len(tested), "refused_cases": rejected, "unaccepted_blocks": missing}
	raw, _ := json.MarshalIndent(report, "", "  ")
	if out := os.Getenv("POD_BLOCK_REPORT"); out != "" {
		if err := os.WriteFile(out, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
func TestPlanBudgetAndMismatch(t *testing.T) {
	mem := input(t)
	at := uint32(0x1ed40 + 0x1a7f)
	b := blocks[at]
	v := &view{mem: mem, s: machine.HostRegisters{Seg: [4]uint16{0xa000, 0x1ed4, 0x8000, 0x6000}, R: [8]uint16{0, 0, 0, 0, 0xe000}, IP: 0x1a7f, Flags: 0xf002}}
	// Normal FillChar arguments: value, count, destination far pointer.
	word := func(off, val uint16) { mem[0x80000+int(off)] = byte(val); mem[0x80000+int(off)+1] = byte(val >> 8) }
	word(0xe000, 0x1234)
	word(0xe002, 0x1111)
	word(0xe004, 7)
	word(0xe006, 13)
	word(0xe008, 0)
	word(0xe00a, 0xa000)
	p := Plan(v, 4096)
	if p == nil {
		t.Fatal("normal fill refused")
	}
	compare(t, v, p)
	if Plan(v, b.n-1) != nil || Plan(v, 0) != nil {
		t.Fatal("short budget accepted")
	}
	mem[at] ^= 1
	if Plan(v, 4096) != nil {
		t.Fatal("modified code accepted")
	}
	mem[at] ^= 1
	mem[0x449] = 0x0d
	if Plan(v, 4096) != nil {
		t.Fatal("wrong surface accepted")
	}
}
