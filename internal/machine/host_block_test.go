package machine

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"reflect"
	"testing"
)

func TestHostBlocksBudgetsEventsAndScope(t *testing.T) {
	for _, run := range []uint64{1, 2, 3, 4, 7} {
		for _, frame := range []uint64{0, 1, 2, 3} {
			for _, dos := range []bool{false, true} {
				fixture := func() *Machine {
					m := New()
					m.IRQ0Every = 0
					m.VGAFrameEvery = frame
					m.nextFrame = frame
					m.RunLimit = run
					m.CPU.Seg[1] = 0x1000
					m.CPU.IP = 0x100
					m.CPU.DOSBoxCost = dos
					for i := 0; i < 32; i++ {
						m.Mem[0x10100+i] = 0x90
					}
					return m
				}
				a, b := fixture(), fixture()
				var af, bf []uint64
				a.SetOnFrame(func() { af = append(af, a.Steps) })
				b.SetOnFrame(func() { bf = append(bf, b.Steps) })
				ranges := []HostCodeRange{{Lo: 0x10100, Hi: 0x10120}}
				b.ObserveHostBlocks(ranges, func() error {
					if b.HostWorkBudget(run-b.Steps) < 2 {
						return nil
					}
					after := HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP + 2, Flags: b.CPU.Flags}
					cycles := uint64(4)
					if dos {
						cycles = 2
					}
					return b.CommitHostBlock(&HostWorkPlan{After: after, Instructions: 2, Cycles: cycles, LastCS: 0x1000, LastIP: b.CPU.IP + 1})
				})
				ranges[0] = HostCodeRange{}
				if _, err := a.RunUntil(nil, run); err != nil {
					t.Fatal(err)
				}
				if _, err := b.RunUntilObserved(nil, run, nil); err != nil {
					t.Fatal(err)
				}
				if a.Steps != b.Steps || a.CPU.R != b.CPU.R || a.CPU.Seg != b.CPU.Seg || a.CPU.IP != b.CPU.IP || a.CPU.Flags != b.CPU.Flags || a.CPU.Cycles != b.CPU.Cycles || !bytes.Equal(a.Mem, b.Mem) || !reflect.DeepEqual(af, bf) {
					t.Fatalf("block differs budget%d frame%d dos%v", run, frame, dos)
				}
				ac, ai := a.CPU.OpAddr()
				bc, bi := b.CPU.OpAddr()
				if ac != bc || ai != bi {
					t.Fatal("last identity")
				}
				if b.HostCallCommits != 0 || b.HostBlockInstructions > b.Steps {
					t.Fatal("double count")
				}
				b.ObserveHostBlocks(nil, nil)
				if b.onHostBlock != nil || len(b.hostBlockRanges) != 0 {
					t.Fatal("leaked callback")
				}
			}
		}
	}
}
func TestHostBlockFailureBeforeOriginalInstruction(t *testing.T) {
	m := New()
	m.IRQ0Every = 0
	m.VGAFrameEvery = 0
	m.CPU.Seg[cpu.CS] = 0x1000
	m.CPU.IP = 0x100
	m.Mem[0x10100] = 0x40
	m.ObserveHostBlocks([]HostCodeRange{{Lo: 0x10100, Hi: 0x10101}}, func() error { return cpuError{} })
	if m.Step() == nil || m.Steps != 0 || m.CPU.R[cpu.AX] != 0 || m.CPU.IP != 0x100 {
		t.Fatal("original executed after provider failure")
	}
}

type cpuError struct{}

func (cpuError) Error() string { return "host failure" }
