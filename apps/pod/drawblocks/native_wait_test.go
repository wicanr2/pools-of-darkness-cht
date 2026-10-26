package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"reflect"
	"testing"
)

type waitView struct {
	*view
	snapshot machine.VGAStatusSnapshot
	known    bool
}

func (v waitView) VGAStatusSnapshot() (machine.VGAStatusSnapshot, bool) { return v.snapshot, v.known }
func (v waitView) Steps() uint64                                        { return v.snapshot.Steps }

func waitSeed(t *testing.T) *view {
	v := imageSpanSeed(t)
	v.s.IP = 0x2a90
	v.s.R[cpu.SP] = 0x900
	v.s.Seg[cpu.SS] = 0x7000
	v.s.Seg[cpu.ES] = 0x1000
	v.s.R[cpu.AX] = 0x1234
	putRowWord(v, 0x463, 0x3d4)
	putRowWord(v, 0x70900, 0x1234)
	putRowWord(v, 0x70902, 0x4567)
	return v
}

func waitMachine(v *view, phase bool, start uint64, preReads int) *machine.Machine {
	m := original(v)
	for i := 0; i < preReads; i++ {
		m.In8(0x3da)
	}
	m.Mode13Phase = phase
	if phase {
		m.VGAFrameEvery = 165000
	}
	m.Steps = start
	return m
}

func TestNativeWaitOriginal(t *testing.T) {
	cases := 0
	for _, phase := range []bool{false, true} {
		for _, reads := range []int{0, 15, 16, 31} {
			for _, dos := range []bool{false, true} {
				for _, budget := range []uint64{4, 5, 6, 7, 20, 128, 512} {
					v := waitSeed(t)
					v.dos = dos
					start := uint64(0)
					if phase {
						start = 151000
					}
					a, b := waitMachine(v, phase, start, reads), waitMachine(v, phase, start, reads)
					snapshot, known := b.VGAStatusSnapshot()
					job := planWait(waitView{view: v, snapshot: snapshot, known: known}, budget)
					if job == nil {
						t.Fatal("valid wait rejected")
					}
					for n := uint64(0); n < job.Instructions; n++ {
						if err := a.Step(); err != nil {
							t.Fatal(err)
						}
					}
					if err := b.CommitHostBlock(job); err != nil {
						t.Fatal(err)
					}
					if a.CPU.R != b.CPU.R || a.CPU.Seg != b.CPU.Seg || a.CPU.IP != b.CPU.IP || a.CPU.Flags != b.CPU.Flags || a.CPU.Cycles != b.CPU.Cycles || a.Steps != b.Steps || !bytes.Equal(a.Mem, b.Mem) || !reflect.DeepEqual(a.PortsIn, b.PortsIn) || !reflect.DeepEqual(a.VGA, b.VGA) {
						t.Fatalf("wait differs phase%v reads%d dos%v budget%d N%d C%d/%d R%v/%v IP%x/%x F%x/%x", phase, reads, dos, budget, job.Instructions, a.CPU.Cycles, b.CPU.Cycles, a.CPU.R, b.CPU.R, a.CPU.IP, b.CPU.IP, a.CPU.Flags, b.CPU.Flags)
					}
					var savedA, savedB bytes.Buffer
					if err := a.SaveState(&savedA); err != nil {
						t.Fatal(err)
					}
					if err := b.SaveState(&savedB); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(savedA.Bytes(), savedB.Bytes()) {
						t.Fatal("wait full machine state differs")
					}
					cases++
				}
			}
		}
	}
	t.Log("wait independent original plan comparisons", cases)
}

func TestNativeWaitResumeOriginal(t *testing.T) {
	cases := 0
	for _, phase := range []bool{false, true} {
		for _, entry := range []uint16{0x2a9d, 0x2a9e, 0x2aa0, 0x2aa2, 0x2aa3, 0x2aa5, 0x2aa7} {
			for _, sample := range []byte{0, 9} {
				for _, dos := range []bool{false, true} {
					for _, budget := range []uint64{1, 2, 3, 8, 64} {
						v := waitSeed(t)
						v.dos = dos
						v.s.IP = entry
						v.s.R[cpu.DX] = 0x3da
						v.s.R[cpu.AX] = 0x1200 | uint16(sample)
						v.s.Seg[cpu.ES] = 0x40
						v.s.Flags = waitTestFlags(v.Flags(), sample)
						start := uint64(0)
						if phase {
							start = 151000
						}
						a, b := waitMachine(v, phase, start, 0), waitMachine(v, phase, start, 0)
						snapshot, known := b.VGAStatusSnapshot()
						job := planWait(waitView{view: v, snapshot: snapshot, known: known}, budget)
						if job == nil {
							t.Fatal("valid resume rejected", entry, budget)
						}
						for n := uint64(0); n < job.Instructions; n++ {
							if err := a.Step(); err != nil {
								t.Fatal(err)
							}
						}
						if err := b.CommitHostBlock(job); err != nil {
							t.Fatal(err)
						}
						var sa, sb bytes.Buffer
						if err := a.SaveState(&sa); err != nil {
							t.Fatal(err)
						}
						if err := b.SaveState(&sb); err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
							t.Fatalf("wait resume differs phase%v entry%x sample%d dos%v budget%d N%d C%d/%d R%v/%v IP%x/%x F%x/%x", phase, entry, sample, dos, budget, job.Instructions, a.CPU.Cycles, b.CPU.Cycles, a.CPU.R, b.CPU.R, a.CPU.IP, b.CPU.IP, a.CPU.Flags, b.CPU.Flags)
						}
						cases++
					}
				}
			}
		}
	}
	t.Log("wait resume independent original plan comparisons", cases)
}

func TestNativeWaitRefusals(t *testing.T) {
	for _, kind := range []string{"budget", "prefix-budget", "code", "mode", "BIOS", "stack-wrap", "stack-low", "stack-video", "entry", "CS", "port", "snapshot", "snapshot-empty", "snapshot-step-overflow"} {
		t.Run(kind, func(t *testing.T) {
			v := waitSeed(t)
			m := waitMachine(v, false, 0, 0)
			snapshot, known := m.VGAStatusSnapshot()
			w := waitView{view: v, snapshot: snapshot, known: known}
			budget := uint64(128)
			switch kind {
			case "budget":
				budget = 0
			case "prefix-budget":
				budget = 3
			case "code":
				v.mem[0x48a10+0x2a90] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "BIOS":
				putRowWord(v, 0x463, 0x3b4)
			case "stack-wrap":
				v.s.R[cpu.SP] = 65535
			case "stack-low":
				v.s.Seg[cpu.SS] = 0
				v.s.R[cpu.SP] = 0
			case "stack-video":
				v.s.Seg[cpu.SS] = 0xa000
			case "entry":
				v.s.IP = 0x2a91
			case "CS":
				v.s.Seg[cpu.CS]++
			case "port":
				v.s.IP = 0x2a9d
				v.s.R[cpu.DX] = 0x3ba
			case "snapshot":
				v.s.IP = 0x2a9d
				v.s.R[cpu.DX] = 0x3da
				w.known = false
			case "snapshot-empty":
				v.s.IP = 0x2a9d
				v.s.R[cpu.DX] = 0x3da
				w.snapshot = machine.VGAStatusSnapshot{}
			case "snapshot-step-overflow":
				v.s.IP = 0x2a9d
				v.s.R[cpu.DX] = 0x3da
				w.snapshot.Steps = ^uint64(0)
			}
			before := append([]byte(nil), v.mem...)
			regs := v.s
			if planWait(w, budget) != nil {
				t.Fatal("invalid wait admitted")
			}
			if !bytes.Equal(before, v.mem) || v.s != regs {
				t.Fatal("decline changed input")
			}
		})
	}
}
