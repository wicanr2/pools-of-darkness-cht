package machine

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
)

func nativeFixture(code []byte, count uint16, backward bool, model cpu.Model, cost bool) *Machine {
	m := New()
	m.SetVideoMode(0x13)
	for i := 0; i < 8192; i++ {
		m.Mem[0x30000+i] = byte(i*37 + 13)
	}
	copy(m.Mem[0x10000:], code)
	m.CPU.Seg = [4]uint16{0xa000, 0x1000, 0x8000, 0x3000}
	m.CPU.R = [8]uint16{0x1234, count, 0x5678, 0x9abc, 0xff00, 0x0100, 0x0400, 0x1000}
	m.CPU.IP = 0
	m.CPU.Model = model
	m.CPU.DOSBoxCost = cost
	m.CPU.Flags = 0xF293
	if backward {
		m.CPU.Flags |= cpu.DF
	}
	return m
}
func sameNativeState(t *testing.T, a, b *Machine) {
	t.Helper()
	if a.CPU.R != b.CPU.R || a.CPU.Seg != b.CPU.Seg || a.CPU.IP != b.CPU.IP || a.CPU.Flags != b.CPU.Flags || a.CPU.Cycles != b.CPU.Cycles || a.Steps != b.Steps || !bytes.Equal(a.Mem, b.Mem) {
		t.Fatal("native string differs from original CPU/RAM/accounting")
	}
}
func TestNativeVideoStringEquivalent(t *testing.T) {
	for _, model := range []cpu.Model{cpu.Model8086, cpu.Model80186, cpu.Model80386} {
		for _, cost := range []bool{false, true} {
			for _, back := range []bool{false, true} {
				for _, op := range []byte{0xa4, 0xa5, 0xaa, 0xab} {
					for _, prefix := range []byte{0, 0xf2, 0xf3} {
						for _, count := range []uint16{0, 1, 7, 8, 319} {
							code := []byte{op}
							if prefix != 0 {
								code = append([]byte{prefix}, code...)
							}
							a, b := nativeFixture(code, count, back, model, cost), nativeFixture(code, count, back, model, cost)
							b.SetNativeVideo(true)
							var wa, wb []VideoWrite
							var commands []VideoCommand
							a.ObserveVideoWrites(func(w VideoWrite) { wa = append(wa, w) })
							b.ObserveVideoWrites(func(w VideoWrite) { wb = append(wb, w) })
							b.ObserveVideoCommands(func(c VideoCommand) { commands = append(commands, c) })
							if err := a.Step(); err != nil {
								t.Fatal(err)
							}
							if err := b.Step(); err != nil {
								t.Fatal(err)
							}
							sameNativeState(t, a, b)
							if !reflect.DeepEqual(wa, wb) {
								t.Fatal("per-pixel original notifications changed")
							}
							n := int(count)
							if prefix == 0 {
								n = 1
							}
							if n == 0 {
								if len(commands) != 0 {
									t.Fatal("zero-count command wrote pixels")
								}
							} else if len(commands) != 1 || !commands[0].Native {
								t.Fatal("native primitive not intercepted")
							}
						}
					}
				}
			}
		}
	}
}
func TestNativeVideoFallbackAndPayloadIsolation(t *testing.T) {
	for _, kind := range []string{"source-read-watch", "write-watch", "source-video", "segment-wrap", "dest-outside", "non-13h"} {
		a, b := nativeFixture([]byte{0xf3, 0xa4}, 8, false, cpu.Model8086, false), nativeFixture([]byte{0xf3, 0xa4}, 8, false, cpu.Model8086, false)
		for _, m := range []*Machine{a, b} {
			switch kind {
			case "source-read-watch":
				m.WatchReads(0x30000, 0x3ffff, func(uint32, byte) {})
			case "write-watch":
				m.WatchWrites(0xa0000, 0xaffff, func(uint32, byte, byte) {})
			case "source-video":
				m.CPU.Seg[cpu.DS] = 0xa000
			case "segment-wrap":
				m.CPU.R[cpu.SI] = 0xfffc
			case "dest-outside":
				m.CPU.R[cpu.DI] = 63998
			case "non-13h":
				m.SetVideoMode(3)
			}
		}
		b.SetNativeVideo(true)
		native := 0
		b.ObserveVideoCommands(func(c VideoCommand) {
			if c.Native {
				native++
			}
		})
		if e := a.Step(); e != nil {
			t.Fatal(e)
		}
		if e := b.Step(); e != nil {
			t.Fatal(e)
		}
		sameNativeState(t, a, b)
		if native != 0 {
			t.Fatal(kind, "unsafe native path")
		}
	}
	m := nativeFixture([]byte{0xf3, 0xa4}, 8, false, cpu.Model8086, false)
	m.SetNativeVideo(true)
	m.ObserveVideoCommands(func(c VideoCommand) {
		if len(c.Data) > 0 {
			c.Data[0] ^= 255
		}
	})
	want := m.Mem[0x30400]
	if e := m.Step(); e != nil {
		t.Fatal(e)
	}
	if m.Mem[0xa1000] != want {
		t.Fatal("observer command payload aliases guest RAM")
	}
}
func TestNativeVideoInstructionEvents(t *testing.T) {
	code := []byte{0xf3, 0xa4, 0xb9, 0x04, 0x00, 0xf3, 0xab, 0x90, 0xeb, 0xfd}
	a, b := nativeFixture(code, 8, false, cpu.Model8086, false), nativeFixture(code, 8, false, cpu.Model8086, false)
	b.SetNativeVideo(true)
	var ea, eb []uint64
	a.VGAFrameEvery = 1
	b.VGAFrameEvery = 1
	a.IRQ0Every = 13
	b.IRQ0Every = 13
	a.nextIRQ0, b.nextIRQ0 = 13, 13
	a.SetOnFrame(func() { ea = append(ea, a.Steps, a.CPU.Cycles) })
	b.SetOnFrame(func() { eb = append(eb, b.Steps, b.CPU.Cycles) })
	for i := 0; i < 80; i++ {
		if e := a.Step(); e != nil {
			t.Fatal(e)
		}
		if e := b.Step(); e != nil {
			t.Fatal(e)
		}
		sameNativeState(t, a, b)
	}
	if !reflect.DeepEqual(ea, eb) || a.Ticks != b.Ticks || a.Ticks == 0 {
		t.Fatal("instruction event boundary changed")
	}
}

func TestNativeVideoGuardsAndRestoreOrder(t *testing.T) {
	m := nativeFixture([]byte{0xf3, 0xa4}, 8, false, cpu.Model8086, false)
	m.SetNativeVideo(true)
	m.WatchWrites(0x20000, 0x20010, func(uint32, byte, byte) { t.Fatal("unrelated write guard called") })
	m.WatchReads(0x20000, 0x20010, func(uint32, byte) { t.Fatal("unrelated read guard called") })
	m.WatchWrite(0x20020, 0x20030, func(*Machine, uint32, byte, byte) { t.Fatal("unrelated post guard called") })
	native := 0
	m.ObserveVideoCommands(func(c VideoCommand) {
		if c.Native {
			native++
		}
	})
	if e := m.Step(); e != nil {
		t.Fatal(e)
	}
	if native != 1 {
		t.Fatal("unrelated guards disabled native command")
	}
	m = nativeFixture([]byte{0xaa}, 1, false, cpu.Model8086, false)
	m.SetNativeVideo(true)
	var cmds []VideoCommand
	m.ObserveVideoCommands(func(c VideoCommand) { cmds = append(cmds, c) })
	active := false
	m.WatchWrite(0xa1000, 0xa1000, func(m *Machine, a uint32, old, new byte) {
		if !active {
			active = true
			m.Write8(a, old)
			active = false
		}
	})
	if e := m.Step(); e != nil {
		t.Fatal(e)
	}
	if len(cmds) != 2 || cmds[0].Native || cmds[0].Value != 0x34 || cmds[1].Value != 0 || m.Mem[0xa1000] != 0 {
		t.Fatal("restoring watcher command order", cmds)
	}
}
