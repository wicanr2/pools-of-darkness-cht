package pod

import (
	"bytes"
	"fmt"
	"github.com/wicanr2/dosgolem/internal/machine"
)

type DrawWriterCount struct{ Commands, NativeBytes, FallbackBytes uint64 }
type DirectStats struct {
	Commands, NativeBytes, FallbackBytes, Frames, BootstrapFrames, Mismatches uint64
	RGBFrames, OriginalRGBPixels, SelectedRGBPixels, LegacyRGBFrames          uint64
	HostCalls, HostInstructions                                               uint64
	Writers                                                                   map[uint32]DrawWriterCount
}

// DirectCanvas is the CPU reference backend for typed index commands. It never
// reads a guest pointer. Committed pixels stop at the retrace frame boundary.
type DirectCanvas struct {
	pending, committed []byte
	ready, seenDraw    bool
	stats              DirectStats
}

func (d *DirectCanvas) RecordHostWork(calls, instructions uint64) {
	if d != nil {
		d.stats.HostCalls += calls
		d.stats.HostInstructions += instructions
	}
}

func NewDirectCanvas() *DirectCanvas {
	return &DirectCanvas{pending: make([]byte, 64000), stats: DirectStats{Writers: map[uint32]DrawWriterCount{}}}
}
func (d *DirectCanvas) Reset() {
	if d != nil {
		clear(d.pending)
		d.ready = false
		d.seenDraw = false
	}
}
func (d *DirectCanvas) Apply(c machine.VideoCommand) error {
	if d == nil {
		return nil
	}
	if c.Width != 1 && c.Width != 2 {
		return fmt.Errorf("draw command width")
	}
	if c.Kind != machine.VideoPixel && c.Kind != machine.VideoFill && c.Kind != machine.VideoCopy {
		return fmt.Errorf("draw command kind")
	}
	if c.Kind == machine.VideoPixel && (c.Width != 1 || c.Count != 1) {
		return fmt.Errorf("pixel command size")
	}
	if c.Count == 0 {
		return nil
	}
	lo, hi := int(c.Offset), int(c.Offset)+int(c.Width)*int(c.Count)
	if c.Backward {
		lo, hi = int(c.Offset)-(int(c.Count)-1)*int(c.Width), int(c.Offset)+int(c.Width)
	}
	if lo < 0 || hi > 64000 || c.Kind == machine.VideoCopy && len(c.Data) != int(c.Width)*int(c.Count) {
		return fmt.Errorf("draw command bounds/source")
	}
	for i := 0; i < int(c.Count); i++ {
		off := int(c.Offset) + i*int(c.Width)
		if c.Backward {
			off = int(c.Offset) - i*int(c.Width)
		}
		for b := 0; b < int(c.Width); b++ {
			v := byte(c.Value >> (8 * b))
			if c.Kind == machine.VideoCopy {
				v = c.Data[i*int(c.Width)+b]
			}
			d.pending[off+b] = v
		}
	}
	d.seenDraw = true
	d.stats.Commands++
	key := uint32(c.CS)<<16 | uint32(c.IP)
	w := d.stats.Writers[key]
	w.Commands++
	n := uint64(c.Count) * uint64(c.Width)
	if c.Native {
		d.stats.NativeBytes += n
		w.NativeBytes += n
	} else {
		d.stats.FallbackBytes += n
		w.FallbackBytes += n
	}
	d.stats.Writers[key] = w
	return nil
}
func (d *DirectCanvas) Frame(original []byte) error {
	if d == nil || len(original) != 64000 {
		return nil
	}
	d.stats.Frames++
	if !d.ready && !d.seenDraw {
		copy(d.pending, original)
		d.stats.BootstrapFrames++
	}
	if !bytes.Equal(original, d.pending) {
		d.stats.Mismatches++
		return fmt.Errorf("draw command canvas differs from original indexed frame")
	}
	d.ready = true
	d.committed = append(d.committed[:0], d.pending...)
	return nil
}
func (d *DirectCanvas) Stats() DirectStats {
	if d == nil {
		return DirectStats{}
	}
	s := d.stats
	s.Writers = map[uint32]DrawWriterCount{}
	for k, v := range d.stats.Writers {
		s.Writers[k] = v
	}
	return s
}
func (l *Live) EnableDirect() { l.Direct = NewDirectCanvas() }
func (l *Live) VideoCommand(c machine.VideoCommand) {
	if l.Direct != nil {
		if err := l.Direct.Apply(c); err != nil {
			panic(err)
		}
	}
}
