package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func TestBoundedDACFlags(t *testing.T) {
	cases, plans := 0, 0
	for _, flags := range []uint16{0xf002, 0xfed7} {
		for _, dos := range []bool{false, true} {
			for phase := 0; phase < 3; phase++ {
				for _, budget := range []uint64{1, 5, 16} {
					params := [4]uint16{127, 1, 17, 31}
					a, b := paletteMachine(t, params, dos), paletteMachine(t, params, dos)
					a.CPU.Flags = flags
					b.CPU.Flags = flags
					paletteDevicePhase(a, phase)
					paletteDevicePhase(b, phase)
					a.PortLog = nil
					b.PortLog = nil
					for i := 0; i < 16; i++ {
						if e := a.Step(); e != nil {
							t.Fatal(e)
						}
					}
					for b.Steps < 16 {
						p := planPaletteOutputBounded(paletteMachineView{b}, budget)
						if p == nil {
							t.Fatal("flags RGB whole CPU fallback")
						}
						if e := b.CommitHostBlock(p); e != nil {
							t.Fatal(e)
						}
						plans++
					}
					paletteMachinesEqual(t, a, b)
					cases++
				}
			}
		}
	}
	t.Logf("bounded DAC CF/IF/DF whole SaveState+PortLog cases=%d Go plans=%d", cases, plans)
}

type unknownPaletteModel struct{ paletteMachineView }

func (v unknownPaletteModel) CPUModel() (cpu.Model, bool) { return 0, false }
func TestBoundedDACRefusals(t *testing.T) {
	cases := 0
	for _, kind := range []string{"budget", "mode", "CS", "IP", "source", "model", "model-unknown", "SP-low", "SP-high", "stack-low", "stack-high", "stack-driver", "index", "red", "green", "blue", "resume-AX", "resume-DX", "resume-BP", "resume-CLI", "resume-flags", "resume-STI", "resume-POP"} {
		m := paletteMachine(t, [4]uint16{127, 1, 17, 31}, false)
		budget := uint64(16)
		unknown := false
		word := func(off uint32, value uint16) {
			m.Mem[0x8e000+off] = byte(value)
			m.Mem[0x8e000+off+1] = byte(value >> 8)
		}
		target := uint16(0)
		switch kind {
		case "resume-AX", "resume-DX":
			target = 0x2ab2
		case "resume-BP":
			target = 0x2aab
		case "resume-CLI":
			target = 0x2aac
		case "resume-flags":
			target = 0x2ab4
		case "resume-STI":
			target = 0x2ac1
		case "resume-POP":
			target = 0x2ac2
		}
		if target != 0 {
			for i := 0; i < 16 && m.CPU.IP != target; i++ {
				if e := m.Step(); e != nil {
					t.Fatal(e)
				}
			}
			if m.CPU.IP != target {
				t.Fatal("missing resume stage")
			}
		}
		switch kind {
		case "budget":
			budget = 0
		case "mode":
			m.Mem[0x449] = 3
		case "CS":
			m.CPU.Seg[cpu.CS]--
		case "IP":
			m.CPU.IP = 0xffff
		case "source":
			m.Mem[0x48a10+0x2aa8] ^= 1
		case "model":
			m.CPU.Model = cpu.Model8086
		case "model-unknown":
			unknown = true
		case "SP-low":
			m.CPU.R[cpu.SP] = 1
		case "SP-high":
			m.CPU.R[cpu.SP] = 0xfff5
		case "stack-low":
			m.CPU.Seg[cpu.SS] = 0
			m.CPU.R[cpu.SP] = 0x500
		case "stack-high":
			m.CPU.Seg[cpu.SS] = 0xa000
		case "stack-driver":
			m.CPU.Seg[cpu.SS] = 0x48a1
			m.CPU.R[cpu.SP] = 0x1200
		case "index":
			word(4, 256)
		case "red":
			word(6, 64)
		case "green":
			word(8, 64)
		case "blue":
			word(10, 64)
		case "resume-AX":
			m.CPU.R[cpu.AX] ^= 1
		case "resume-DX":
			m.CPU.R[cpu.DX]++
		case "resume-BP":
			m.CPU.R[cpu.BP]++
		case "resume-CLI":
			m.CPU.Flags |= cpu.IF
		case "resume-flags":
			m.CPU.Flags ^= cpu.PF
		case "resume-STI":
			m.CPU.Flags &^= cpu.IF
		case "resume-POP":
			m.CPU.R[cpu.BP]++
		}
		var before, after bytes.Buffer
		if e := m.SaveState(&before); e != nil {
			t.Fatal(e)
		}
		var view CPUView = paletteMachineView{m}
		if unknown {
			view = unknownPaletteModel{paletteMachineView{m}}
		}
		if p := planPaletteOutputBounded(view, budget); p != nil {
			t.Fatal("unproved RGB accepted", kind)
		}
		if e := m.SaveState(&after); e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(before.Bytes(), after.Bytes()) {
			t.Fatal("RGB refusal mutated machine", kind)
		}
		cases++
	}
	t.Logf("bounded DAC source/frame/model/resume refusals=%d unchanged", cases)
}
