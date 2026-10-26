package machine

import (
	"bytes"
	"reflect"
	"testing"
)

func TestHostVideoBatchesPreserveByteOrderAndContext(t *testing.T) {
	for _, kind := range []string{"fill", "copy", "back-byte", "back-word", "sparse"} {
		m := New()
		m.SetVideoMode(0x13)
		m.IRQ0Every = 0
		m.VGAFrameEvery = 0
		m.SetNativeVideo(true)
		p := &HostWorkPlan{After: HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: 12, Flags: m.CPU.Flags}, Instructions: 3, Cycles: 6, LastCS: 0x1234, LastIP: 10}
		for i := 0; i < 12; i++ {
			off := 100 + i
			value := byte(9)
			switch kind {
			case "copy":
				value = byte(i)
			case "back-byte":
				off = 100 - i
				value = byte(i)
			case "back-word":
				off = 100 - (i/2)*2 + i%2
				value = byte(i)
			case "sparse":
				off = 100 + i*2
				value = byte(i)
			}
			p.Writes = append(p.Writes, HostWrite{Address: vgaLo + uint32(off), Value: value, Step: 2, CS: 0x1234, IP: 8})
		}
		var observed []HostWrite
		var commands []VideoCommand
		canvas := make([]byte, 64000)
		m.ObserveVideoWrites(func(w VideoWrite) {
			observed = append(observed, HostWrite{Address: vgaLo + w.Offset, Value: w.Value, Step: w.Step, CS: w.CS, IP: w.IP})
		})
		m.ObserveVideoCommands(func(c VideoCommand) {
			commands = append(commands, c)
			if !c.Host || !c.Native || c.Step != 2 || c.CS != 0x1234 || c.IP != 8 {
				t.Fatal("context")
			}
			for i := 0; i < int(c.Count); i++ {
				off := int(c.Offset) + i*int(c.Width)
				if c.Backward {
					off = int(c.Offset) - i*int(c.Width)
				}
				for b := 0; b < int(c.Width); b++ {
					v := byte(c.Value >> uint(8*b))
					if c.Kind == VideoCopy {
						v = c.Data[i*int(c.Width)+b]
					}
					canvas[off+b] = v
				}
			}
		})
		if err := m.CommitHostBlock(p); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(observed, p.Writes) || !bytes.Equal(canvas, m.Indexed()) {
			t.Fatal("batch changed original writes", kind)
		}
		want := 1
		if kind == "sparse" {
			want = 12
		}
		if len(commands) != want {
			t.Fatal("not batched", kind, len(commands))
		}
		if kind == "back-word" && (commands[0].Width != 2 || !commands[0].Backward) {
			t.Fatal("backward word")
		}
		// Already-published command payload does not borrow the proposal.
		p.Writes[0].Value ^= 255
		if commands[0].Kind == VideoCopy && commands[0].Data[0] == p.Writes[0].Value {
			t.Fatal("payload aliased")
		}
	}
}
