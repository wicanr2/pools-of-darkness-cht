package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	machine "github.com/wicanr2/dosgolem/internal/machine"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type paletteMachineView struct{ m *machine.Machine }

func (v paletteMachineView) Steps() uint64 { return v.m.Steps }

func (v paletteMachineView) Read8(a uint32) byte { return v.m.Mem[a&0xfffff] }
func (v paletteMachineView) Read16(a uint32) uint16 {
	return uint16(v.Read8(a)) | uint16(v.Read8(a+1))<<8
}
func (v paletteMachineView) AX() uint16                  { return v.m.CPU.R[0] }
func (v paletteMachineView) CX() uint16                  { return v.m.CPU.R[1] }
func (v paletteMachineView) DX() uint16                  { return v.m.CPU.R[2] }
func (v paletteMachineView) BX() uint16                  { return v.m.CPU.R[3] }
func (v paletteMachineView) SP() uint16                  { return v.m.CPU.R[4] }
func (v paletteMachineView) BP() uint16                  { return v.m.CPU.R[5] }
func (v paletteMachineView) SI() uint16                  { return v.m.CPU.R[6] }
func (v paletteMachineView) DI() uint16                  { return v.m.CPU.R[7] }
func (v paletteMachineView) ES() uint16                  { return v.m.CPU.Seg[0] }
func (v paletteMachineView) CS() uint16                  { return v.m.CPU.Seg[1] }
func (v paletteMachineView) SS() uint16                  { return v.m.CPU.Seg[2] }
func (v paletteMachineView) DS() uint16                  { return v.m.CPU.Seg[3] }
func (v paletteMachineView) IP() uint16                  { return v.m.CPU.IP }
func (v paletteMachineView) Flags() uint16               { return v.m.CPU.Flags }
func (v paletteMachineView) DOSBoxCost() bool            { return v.m.CPU.DOSBoxCost }
func (v paletteMachineView) CPUModel() (cpu.Model, bool) { return v.m.CPU.Model, true }

func paletteMachine(t *testing.T, params [4]uint16, dos bool) *machine.Machine {
	t.Helper()
	root := os.Getenv("POD_ROOT")
	if root == "" {
		t.Skip("original fixture absent")
	}
	mem, e := os.ReadFile(filepath.Join(root, "workplace/probe/draw-intercept-v1/legacy-demo/writer-48a1-6903.mem"))
	if e != nil {
		t.Fatal(e)
	}
	m := machine.New()
	m.SetVideoMode(0x13)
	copy(m.Mem, mem)
	m.IRQ0Every = 0
	m.VGAFrameEvery = 0
	m.CPU.Model = cpu.Model80386
	m.CPU.DOSBoxCost = dos
	m.CPU.R = [8]uint16{0xab77, 23, 45, 12, 0xe000, 0xd000, 0x440, 0x880}
	m.CPU.Seg = [4]uint16{0x6000, 0x48a1, 0x8000, 0x6000}
	m.CPU.IP = 0x2aa8
	m.CPU.Flags = 0xfad7
	for i, value := range []uint16{0x1234, 0x4567, params[0], params[1], params[2], params[3]} {
		a := 0x8e000 + i*2
		m.Mem[a] = byte(value)
		m.Mem[a+1] = byte(value >> 8)
	}
	return m
}
func paletteDevicePhase(m *machine.Machine, phase int) {
	m.Out8(0x3c8, 17)
	for i := 0; i < phase; i++ {
		m.Out8(0x3c9, byte(i+11))
	}
	m.Out8(0x3c7, 8)
	m.In8(0x3c9)
}
func paletteMachinesEqual(t *testing.T, a, b *machine.Machine) {
	t.Helper()
	var sa, sb bytes.Buffer
	if e := a.SaveState(&sa); e != nil {
		t.Fatal(e)
	}
	if e := b.SaveState(&sb); e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(sa.Bytes(), sb.Bytes()) || !reflect.DeepEqual(a.PortLog, b.PortLog) {
		t.Fatalf("RGB bounded state differs steps%d/%d cycles%d/%d IP%x/%x ports%d/%d", a.Steps, b.Steps, a.CPU.Cycles, b.CPU.Cycles, a.CPU.IP, b.CPU.IP, len(a.PortLog), len(b.PortLog))
	}
}
func TestBoundedDACPrefixes(t *testing.T) {
	cases := 0
	for _, params := range [][4]uint16{{0, 0, 0, 0}, {255, 63, 63, 63}, {127, 1, 17, 31}} {
		for _, dos := range []bool{false, true} {
			for phase := 0; phase < 3; phase++ {
				for prefix := 0; prefix < 16; prefix++ {
					for _, budget := range []uint64{1, 2, 5, 16} {
						a, b := paletteMachine(t, params, dos), paletteMachine(t, params, dos)
						for i := 0; i < prefix; i++ {
							if e := a.Step(); e != nil {
								t.Fatal(e)
							}
							if e := b.Step(); e != nil {
								t.Fatal(e)
							}
						}
						paletteDevicePhase(a, phase)
						paletteDevicePhase(b, phase)
						a.PortLog = nil
						b.PortLog = nil
						p := planPaletteOutputBounded(paletteMachineView{b}, budget)
						if p == nil {
							t.Fatalf("bounded RGB refused prefix%d budget%d", prefix, budget)
						}
						for i := uint64(0); i < p.Instructions; i++ {
							if e := a.Step(); e != nil {
								t.Fatal(e)
							}
						}
						if e := b.CommitHostBlock(p); e != nil {
							t.Fatal(e)
						}
						paletteMachinesEqual(t, a, b)
						cases++
					}
				}
			}
		}
	}
	t.Logf("bounded DAC original prefix/device-phase comparisons=%d", cases)
}
func TestBoundedDACWhole(t *testing.T) {
	cases, plans := 0, 0
	for _, params := range [][4]uint16{{0, 0, 0, 0}, {255, 63, 63, 63}, {127, 1, 17, 31}} {
		for _, dos := range []bool{false, true} {
			for phase := 0; phase < 3; phase++ {
				for _, budget := range []uint64{1, 2, 5, 16} {
					a, b := paletteMachine(t, params, dos), paletteMachine(t, params, dos)
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
							t.Fatal("RGB whole CPU fallback")
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
	t.Logf("bounded DAC whole original SaveState+PortLog cases=%d Go plans=%d CPU fallback=0", cases, plans)
}
