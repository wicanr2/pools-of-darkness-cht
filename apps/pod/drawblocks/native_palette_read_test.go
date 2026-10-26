package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"reflect"
	"testing"
)

type paletteReadView struct {
	*view
	colors [768]byte
	known  bool
}

func (v paletteReadView) PaletteRGB(index byte) ([3]byte, bool) {
	at := int(index) * 3
	return [3]byte{v.colors[at], v.colors[at+1], v.colors[at+2]}, v.known
}

func paletteReadSeed(t *testing.T) *view {
	v := imageSpanSeed(t)
	v.s.IP = 0x2aeb
	v.s.Seg[cpu.DS] = 0x48a1
	v.s.Seg[cpu.SS] = 0x7000
	v.s.R[cpu.SP] = 0x900
	v.s.R[cpu.BP] = 0x800
	for i, n := range []uint16{0x1234, 0x4567, 17, 0x1179, 0x117b, 0x117d} {
		putRowWord(v, 0x70900+uint32(i)*2, n)
	}
	return v
}

func paletteReadFixture(m *machine.Machine, phase int) {
	for i := range m.DAC {
		m.DAC[i] = byte((i*11 + 7) % 64)
	}
	m.Out8(0x3c7, 200)
	for i := 0; i < phase; i++ {
		m.In8(0x3c9)
	}
	m.PortLog = nil
}

func TestNativePaletteReadOriginal(t *testing.T) {
	cases := 0
	for _, index := range []uint16{0, 17, 255} {
		for _, phase := range []int{0, 1, 2} {
			for _, flags := range []uint16{0x7246, 0x7603} {
				for _, dos := range []bool{false, true} {
					v := paletteReadSeed(t)
					v.dos = dos
					v.s.Flags = flags
					putRowWord(v, 0x70904, index)
					a, b := original(v), original(v)
					paletteReadFixture(a, phase)
					paletteReadFixture(b, phase)
					w := paletteReadView{view: v, colors: a.DAC, known: true}
					job := planPaletteRead(w, 20)
					if job == nil {
						t.Fatal("valid RGB read rejected")
					}
					var writesA []machine.HostWrite
					a.WatchWrite(0, 0xfffff, func(m *machine.Machine, at uint32, old, val byte) {
						cs, ip := m.CPU.OpAddr()
						writesA = append(writesA, machine.HostWrite{Address: at, Value: val, Step: m.Steps, CS: cs, IP: ip})
					})
					for i := 0; i < 20; i++ {
						if err := a.Step(); err != nil {
							t.Fatal(err)
						}
					}
					if err := b.CommitHostBlock(job); err != nil {
						t.Fatal(err)
					}
					if a.CPU.R != b.CPU.R || a.CPU.Seg != b.CPU.Seg || a.CPU.IP != b.CPU.IP || a.CPU.Flags != b.CPU.Flags || a.CPU.Cycles != b.CPU.Cycles || a.Steps != b.Steps || !bytes.Equal(a.Mem, b.Mem) {
						t.Fatal("RGB read CPU/RAM/time differs")
					}
					if a.DAC != b.DAC || !reflect.DeepEqual(a.VGA, b.VGA) || !reflect.DeepEqual(a.Ports, b.Ports) || !reflect.DeepEqual(a.PortsIn, b.PortsIn) || !reflect.DeepEqual(a.PortLog, b.PortLog) || !reflect.DeepEqual(writesA, expandWrites(job)) {
						t.Fatal("RGB read device/ordered effects differ", a.PortLog, b.PortLog)
					}
					var stateA, stateB bytes.Buffer
					if err := a.SaveState(&stateA); err != nil {
						t.Fatal(err)
					}
					if err := b.SaveState(&stateB); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(stateA.Bytes(), stateB.Bytes()) {
						t.Fatal("RGB read full saved machine differs")
					}
					cases++
				}
			}
		}
	}
	t.Log("RGB read independent original comparisons", cases)
}

func TestNativePaletteReadRefusals(t *testing.T) {
	for _, kind := range []string{"budget", "code", "mode", "index", "snapshot", "RGB", "stack-wrap", "stack-low", "stack-video", "stack-code", "output-wrap", "output-low", "output-video", "output-code", "output-frame"} {
		t.Run(kind, func(t *testing.T) {
			v := paletteReadSeed(t)
			m := original(v)
			paletteReadFixture(m, 0)
			w := paletteReadView{view: v, colors: m.DAC, known: true}
			budget := uint64(20)
			switch kind {
			case "budget":
				budget = 19
			case "code":
				v.mem[0x48a10+0x2aeb] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "index":
				putRowWord(v, 0x70904, 256)
			case "snapshot":
				w.known = false
			case "RGB":
				w.colors[17*3] = 64
			case "stack-wrap":
				v.s.R[cpu.SP] = 65535
			case "stack-low":
				v.s.Seg[cpu.SS] = 0
				v.s.R[cpu.SP] = 0
			case "stack-video":
				v.s.Seg[cpu.SS] = 0xa000
			case "stack-code":
				v.s.Seg[cpu.SS] = 0x48a1
				v.s.R[cpu.SP] = 0x7000
			case "output-wrap":
				putRowWord(v, 0x70906, 65535)
			case "output-low":
				v.s.Seg[cpu.DS] = 0
				putRowWord(v, 0x70906, 0x100)
			case "output-video":
				v.s.Seg[cpu.DS] = 0xa000
			case "output-code":
				putRowWord(v, 0x70906, 0x6733)
			case "output-frame":
				v.s.Seg[cpu.DS] = 0x7000
				putRowWord(v, 0x70906, 0x8fe)
			}
			before := append([]byte(nil), v.mem...)
			regs := v.s
			colors := w.colors
			if planPaletteRead(w, budget) != nil {
				t.Fatal("invalid RGB read admitted")
			}
			if string(before) != string(v.mem) || v.s != regs || w.colors != colors {
				t.Fatal("decline changed source")
			}
		})
	}
	v := paletteReadSeed(t)
	if planPaletteRead(v, 20) != nil {
		t.Fatal("missing snapshot reader admitted")
	}
}

type paletteRuntimeView struct {
	paletteReadView
	status      machine.VGAStatusSnapshot
	statusKnown bool
}

func (v paletteRuntimeView) VGAStatusSnapshot() (machine.VGAStatusSnapshot, bool) {
	return v.status, v.statusKnown
}
func (v paletteRuntimeView) Steps() uint64 { return v.status.Steps }

func TestNativePaletteReadWholeCalls(t *testing.T) {
	var seed *view
	for _, v := range drawCalls(t) {
		if v.CS() == 0x48a1 && v.IP() == 0x4944 {
			seed = v
			break
		}
	}
	if seed == nil {
		t.Fatal("normal fade entry absent")
	}
	cases, reads := 0, 0
	for _, dos := range []bool{false, true} {
		for _, budget := range []uint64{19, 20, 1024} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
			a, b := callMachine(v), callMachine(v)
			runCall(t, a, v, false)
			stack := uint32(v.SS())*16 + uint32(v.SP())
			retIP, retCS := v.Read16(stack), v.Read16(stack+2)
			used, statusUsed := 0, 0
			for b.Steps < 1000000 {
				if b.Steps > 0 && b.CPU.IP == retIP && b.CPU.Seg[cpu.CS] == retCS {
					break
				}
				w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
				snapshot, known := b.VGAStatusSnapshot()
				if p := Plan(paletteRuntimeView{paletteReadView: paletteReadView{view: w, colors: b.DAC, known: true}, status: snapshot, statusKnown: known}, budget); p != nil {
					if len(p.PaletteReads) > 0 {
						used += len(p.PaletteReads)
					}
					statusUsed += len(p.VideoStatusReads)
					if err := b.CommitHostBlock(p); err != nil {
						t.Fatal(err)
					}
				} else if err := b.Step(); err != nil {
					t.Fatal(err)
				}
			}
			if b.CPU.R != a.CPU.R || b.CPU.Seg != a.CPU.Seg || b.CPU.IP != a.CPU.IP || b.CPU.Flags != a.CPU.Flags || b.CPU.Cycles != a.CPU.Cycles || b.Steps != a.Steps || !bytes.Equal(b.Mem, a.Mem) || a.DAC != b.DAC || !reflect.DeepEqual(a.VGA, b.VGA) || !reflect.DeepEqual(a.Ports, b.Ports) || !reflect.DeepEqual(a.PortsIn, b.PortsIn) || !reflect.DeepEqual(a.PortLog, b.PortLog) {
				t.Fatalf("whole RGB flow differs dos%v budget%d reads%d", dos, budget, used)
			}
			if budget < 20 && used != 0 || budget >= 20 && used == 0 {
				t.Fatal("RGB read budget path missing", budget, used)
			}
			if statusUsed == 0 {
				t.Fatal("Go wait path missing", budget)
			}
			var savedA, savedB bytes.Buffer
			if err := a.SaveState(&savedA); err != nil {
				t.Fatal(err)
			}
			if err := b.SaveState(&savedB); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(savedA.Bytes(), savedB.Bytes()) {
				t.Fatal("whole RGB machine snapshot differs")
			}
			cases++
			reads += used
		}
	}
	t.Log("whole original fade with RGB read cases", cases, "committed RGB reads", reads)
}
