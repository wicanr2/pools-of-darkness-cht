package machine

import "github.com/wicanr2/dosgolem/internal/cpu"

type VideoCommandKind uint8

const (
	VideoPixel VideoCommandKind = iota + 1
	VideoCopy
	VideoFill
)

// VideoCommand owns source bytes detached from guest RAM. Backward changes
// item order; the two bytes of a word retain little-endian order.
type VideoCommand struct {
	Kind             VideoCommandKind
	Step             uint64
	CS, IP           uint16
	Offset           uint32
	Count, Value     uint16
	Width            uint8
	Backward, Native bool
	Host             bool // effects committed by spec436, not a decoded string opcode
	Source           uint32
	Data             []byte
}

func (m *Machine) SetNativeVideo(on bool)                     { m.nativeVideo = on; m.CPU.NativeStrings = on }
func (m *Machine) ObserveVideoCommands(fn func(VideoCommand)) { m.onVideoCommand = fn }
func stringBounds(off uint16, width uint8, count uint16, backward bool) (int, int) {
	lo, hi := int(off), int(off)+int(width)*int(count)
	if backward {
		lo, hi = int(off)-(int(count)-1)*int(width), int(off)+int(width)
	}
	return lo, hi
}
func (m *Machine) WriteStringCommand(c cpu.StringCommand) bool {
	if !m.nativeVideo || m.planarOn || m.Unchained() || m.Peek8(0x449) != 0x13 || (c.Width != 1 && c.Width != 2) {
		return false
	}
	if c.Count == 0 {
		return true
	}
	lo, hi := stringBounds(c.DestOff, c.Width, c.Count, c.Backward)
	base := uint32(c.DestSeg) << 4
	if lo < 0 || hi > 65536 || base+uint32(lo) < vgaLo || base+uint32(hi) > vgaLo+64000 {
		return false
	}
	// Unrelated game-state guards do not observe this instruction's stores.
	// A matching guard may inspect intermediate SI/DI/CX: retain stringOnce.
	dlo, dhi := base+uint32(lo), base+uint32(hi)-1
	if m.onWrite != nil && m.watchLo <= dhi && m.watchHi >= dlo {
		return false
	}
	for _, w := range m.writeWatches {
		if w.lo <= dhi && w.hi >= dlo {
			return false
		}
	}
	var data []byte
	if c.Copy {
		slo, shi := stringBounds(c.SourceOff, c.Width, c.Count, c.Backward)
		sbase := uint32(c.SourceSeg) << 4
		if slo < 0 || shi > 65536 || sbase+uint32(shi) > vgaLo {
			return false
		}
		if m.onRead != nil && m.rWatchLo <= sbase+uint32(shi)-1 && m.rWatchHi >= sbase+uint32(slo) {
			return false
		}
		data = make([]byte, int(c.Width)*int(c.Count))
		for i := 0; i < int(c.Count); i++ {
			off := int(c.SourceOff) + i*int(c.Width)
			if c.Backward {
				off = int(c.SourceOff) - i*int(c.Width)
			}
			copy(data[i*int(c.Width):(i+1)*int(c.Width)], m.Mem[int(sbase)+off:int(sbase)+off+int(c.Width)])
		}
	}
	m.inVideoCommand = true
	defer func() { m.inVideoCommand = false }()
	for i := 0; i < int(c.Count); i++ {
		off := int(c.DestOff) + i*int(c.Width)
		if c.Backward {
			off = int(c.DestOff) - i*int(c.Width)
		}
		for b := 0; b < int(c.Width); b++ {
			value := byte(c.Value >> (8 * b))
			if c.Copy {
				value = data[i*int(c.Width)+b]
			}
			m.Write8(base+uint32(off+b), value)
		}
	}
	if m.onVideoCommand != nil {
		cs, ip := m.CPU.OpAddr()
		kind := VideoFill
		if c.Copy {
			kind = VideoCopy
		}
		m.onVideoCommand(VideoCommand{Kind: kind, Step: m.Steps, CS: cs, IP: ip, Offset: base + uint32(c.DestOff) - vgaLo, Count: c.Count, Value: c.Value, Width: c.Width, Backward: c.Backward, Native: true, Source: cpu.Linear(c.SourceSeg, c.SourceOff), Data: data})
	}
	return true
}
func (m *Machine) commandPixel(a uint32, v byte) {
	if m.onVideoCommand == nil || m.inVideoCommand || !m.nativeVideo || a < vgaLo || a >= vgaLo+64000 || m.planarOn || m.Unchained() || m.Peek8(0x449) != 0x13 {
		return
	}
	step, cs, ip := m.videoWriteContext()
	m.onVideoCommand(VideoCommand{Kind: VideoPixel, Step: step, CS: cs, IP: ip, Offset: a - vgaLo, Count: 1, Width: 1, Value: uint16(v), Native: m.hostEffect != nil, Host: m.hostEffect != nil})
}

// hostVideoGroup batches only writes belonging to the same original atomic
// instruction. VideoWrite still observes every byte in original order.
func (m *Machine) hostVideoGroup(ws []HostWrite) (int, VideoCommand) {
	if len(ws) < 2 || !m.nativeVideo || m.onVideoCommand == nil || ws[0].Address < vgaLo {
		return 0, VideoCommand{}
	}
	a := ws[0]
	width, back := 1, false
	if ws[1].Address+1 == a.Address {
		back = true
	} else if len(ws) > 2 && ws[1].Address == a.Address+1 && ws[2].Address+2 == a.Address {
		width, back = 2, true
	}
	n := 0
	for i, w := range ws {
		if i >= 65535*width || w.Step != a.Step || w.CS != a.CS || w.IP != a.IP {
			break
		}
		delta := (i / width) * width
		if back {
			delta = -delta
		}
		if int64(w.Address) != int64(a.Address)+int64(delta+i%width) {
			break
		}
		n++
	}
	n -= n % width
	if n < 2 {
		return 0, VideoCommand{}
	}
	value := uint16(a.Value)
	if width == 2 {
		value |= uint16(ws[1].Value) << 8
	}
	c := VideoCommand{Kind: VideoFill, Step: m.Steps + a.Step, CS: a.CS, IP: a.IP, Offset: a.Address - vgaLo, Count: uint16(n / width), Width: uint8(width), Value: value, Backward: back, Native: true, Host: true}
	for i, w := range ws[:n] {
		if w.Value != byte(value>>uint(8*(i%width))) {
			c.Kind = VideoCopy
			break
		}
	}
	if c.Kind == VideoCopy {
		c.Data = make([]byte, n)
		for i, w := range ws[:n] {
			c.Data[i] = w.Value
		}
	}
	return n, c
}
