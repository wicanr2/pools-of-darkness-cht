package pod

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"github.com/wicanr2/dosgolem/xlate"
	"testing"
)

func TestGlyphTextCompositionAndOriginal(t *testing.T) {
	seed := glyphViews(t)[0]
	v := *seed
	v.memory = append([]byte(nil), seed.memory...)
	stack := uint32(v.SS())*16 + uint32(v.SP())
	v.memory[stack+4] = 'A'
	g := testGlyphTextFont(t)
	m := glyphMachine(&v, cpu.Model80386)
	m.ObserveVideoWrites(func(w machine.VideoWrite) { g.write(w.Offset) })
	ip, cs := v.Read16(stack), v.Read16(stack+2)
	for n := 0; n < 1000; n++ {
		g.before(residentViewAt(m, false))
		if n > 0 && m.CPU.IP == ip && m.CPU.Seg[cpu.CS] == cs {
			break
		}
		if e := m.Step(); e != nil {
			t.Fatal(e)
		}
	}
	g.frame(m.Indexed())
	if len(g.cells) != 1 {
		t.Fatal("need original A cell")
	}
	var c *glyphTextCell
	for _, cell := range g.cells {
		c = cell
	}
	l := NewLive(g.font, nil, nil)
	l.glyphText = g
	l.indexed = append([]byte(nil), m.Indexed()...)
	l.pal[c.bg] = [3]byte{0, 0, 0}
	l.pal[c.fg] = [3]byte{85, 255, 85}
	for _, scale := range []int{2, 3} {
		original := l.Compose(scale, false, nil)
		hires := l.Compose(scale, true, nil)
		if bytes.Equal(original, hires) {
			t.Fatal("high resolution not applied", scale)
		}
		report := l.GlyphTextReport(scale)
		if report.CellPixels != 8*scale || len(report.Cells) != 1 || report.Cells[0].Code != 'A' || report.Cells[0].Pending || report.Font == "" {
			t.Fatal("invalid font receipt", report)
		}
		if !bytes.Equal(original, l.Compose(scale, false, nil)) {
			t.Fatal("F12 changed original")
		}
		stamp := &xlate.Stamp{X: c.x, Y: c.y, CellW: 8, CellH: 8, Cells: 1, Font: g.font, Text: []rune{'Z'}, State: xlate.Shown, FG: [3]byte{255, 255, 0}, BG: [3]byte{0, 0, 128}}
		l.Layer.Stamps = []*xlate.Stamp{stamp}
		expected := append([]byte(nil), original...)
		l.Layer.Draw(expected, scale, nil)
		if !bytes.Equal(expected, l.Compose(scale, true, nil)) {
			t.Fatal("underlying ASCII covered translation")
		}
		l.Layer.Stamps = nil
	}
	g.reset()
	if len(l.GlyphTextReport(3).Cells) != 0 {
		t.Fatal("mode reset retained glyphs")
	}
	t.Log("native 16/24 font receipts, original toggle, translation priority and mode reset passed")
}
func TestResidentPixelRanges(t *testing.T) {
	for _, ip := range []uint16{0x11a6, 0x11e9, 0x11ea, 0x11eb, 0x11f0, 0x11f1, 0x11f2} {
		matches := 0
		for _, r := range DrawBlockRanges() {
			at := uint32(0x1b2e)*16 + uint32(ip)
			if r.CS == 0x1b2e && at >= r.Lo && at < r.Hi {
				matches++
			}
		}
		expected := 0
		if ip == 0x11ea || ip == 0x11f1 {
			expected = 1
		}
		if matches != expected {
			t.Fatalf("resident block range IP%x %d/%d", ip, matches, expected)
		}
	}
	if targets := GlyphHostCallTargets(); len(targets) != 1 || targets[0] != 0x1c486 {
		t.Fatal("changed CALL-only targets")
	}
}
