package pod

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"os"
	"path/filepath"
	"testing"
)

func testGlyphTextFont(t *testing.T) *glyphText {
	t.Helper()
	root := os.Getenv("POD_ROOT")
	if root == "" {
		t.Fatal("POD_ROOT required")
	}
	font, e := LoadFontScaled(filepath.Join(root, "workplace/font/pod-en.golemfnt"))
	if e != nil {
		t.Fatal(e)
	}
	return newGlyphText(font)
}
func TestGlyphTextOriginalCells(t *testing.T) {
	cases, changed, details := 0, 0, 0
	for _, seed := range glyphViews(t) {
		ch := seed.Read8(uint32(seed.SS())*16 + uint32(seed.SP()) + 4)
		if ch < 0x20 || ch > 0x7e {
			continue
		}
		g := testGlyphTextFont(t)
		m := glyphMachine(seed, cpu.Model80386)
		m.ObserveVideoWrites(func(w machine.VideoWrite) { g.write(w.Offset) })
		stack := uint32(seed.SS())*16 + uint32(seed.SP())
		ip, cs := seed.Read16(stack), seed.Read16(stack+2)
		partial := false
		for n := 0; n < 1000; n++ {
			g.before(residentViewAt(m, seed.dos))
			if n > 0 && m.CPU.IP == ip && m.CPU.Seg[cpu.CS] == cs {
				break
			}
			if e := m.Step(); e != nil {
				t.Fatal(e)
			}
			if !partial && len(g.cells) == 1 {
				partial = true
				g.frame(m.Indexed())
				c := g.cells[g.call.key]
				if !c.pending {
					t.Fatal("need pending original glyph")
				}
				var pal [256][3]byte
				pal[c.bg] = [3]byte{2, 3, 4}
				pal[c.fg] = [3]byte{220, 221, 222}
				out := make([]byte, 960*600*4)
				if g.draw(out, 3, pal) != 1 {
					t.Fatal("pending must mask original")
				}
				for y := c.y * 3; y < (c.y+8)*3; y++ {
					for x := c.x * 3; x < (c.x+8)*3; x++ {
						if !bytes.Equal(out[(y*960+x)*4:(y*960+x)*4+3], pal[c.bg][:]) {
							t.Fatal("pending leaked original")
						}
					}
				}
			}
		}
		g.frame(m.Indexed())
		if len(g.cells) != 1 || g.call != nil {
			t.Fatal("missing complete glyph", ch, len(g.cells))
		}
		var cell *glyphTextCell
		for _, c := range g.cells {
			cell = c
		}
		var pal [256][3]byte
		pal[cell.bg] = [3]byte{2, 3, 4}
		pal[cell.fg] = [3]byte{220, 221, 222}
		for _, scale := range []int{2, 3} {
			out := bytes.Repeat([]byte{17, 18, 19, 255}, 320*200*scale*scale)
			before := append([]byte(nil), out...)
			if g.draw(out, scale, pal) != 1 {
				t.Fatal("missing native font", scale)
			}
			for y := 0; y < 200*scale; y++ {
				for x := 0; x < 320*scale; x++ {
					at := (y*320*scale + x) * 4
					inside := x >= cell.x*scale && x < (cell.x+8)*scale && y >= cell.y*scale && y < (cell.y+8)*scale
					if !inside && !bytes.Equal(out[at:at+4], before[at:at+4]) {
						t.Fatal("font escaped cell")
					}
				}
			}
			detailed := false
			for y := 0; y < 8*scale; y++ {
				for x := 0; x < 8*scale; x++ {
					ax, ay := cell.x*scale+x, cell.y*scale+y
					bx, by := cell.x*scale+(x/scale)*scale, cell.y*scale+(y/scale)*scale
					if !bytes.Equal(out[(ay*320*scale+ax)*4:(ay*320*scale+ax)*4+3], out[(by*320*scale+bx)*4:(by*320*scale+bx)*4+3]) {
						detailed = true
					}
				}
			}
			if detailed {
				details++
			}
			if !bytes.Equal(out, before) {
				changed++
			}
			cases++
		}
		bad := append([]byte(nil), m.Indexed()...)
		bad[cell.y*320+cell.x] ^= 1
		g.frame(bad)
		if len(g.cells) != 0 {
			t.Fatal("foreign overwrite retained font")
		}
	}
	if cases == 0 || changed != cases || details == 0 {
		t.Fatal("no actual high resolution drawing", cases, changed, details)
	}
	t.Logf("native glyph font original-cell draws=%d two scales native-detail cases=%d pending masks and source invalidation passed", cases, details)
}

func TestGlyphTextGuards(t *testing.T) {
	base := glyphViews(t)[0]
	cases := 0
	for _, kind := range []string{"valid", "icon", "extended", "source", "code", "mode", "x", "y", "stack"} {
		v := *base
		v.memory = append([]byte(nil), base.memory...)
		stack := uint32(v.SS())*16 + uint32(v.SP())
		v.memory[stack+4] = 'A'
		switch kind {
		case "icon":
			v.memory[stack+4] = 12
		case "extended":
			v.memory[stack+4] = 128
		case "source":
			table := uint32(v.DS())*16 + 0x9d86
			src := uint32(v.Read16(table+2))*16 + uint32(v.Read16(table)) + 256
			v.memory[src] ^= 1
		case "code":
			v.memory[0x1c486] ^= 1
		case "mode":
			v.memory[0x449] = 3
		case "x":
			v.memory[stack+12] = 40
		case "y":
			v.memory[stack+10] = 25
		case "stack":
			v.before.R[cpu.SP] = 65530
		}
		before := append([]byte(nil), v.memory...)
		g := testGlyphTextFont(t)
		g.before(&v)
		if (g.call != nil) != (kind == "valid") {
			t.Fatal("glyph source acceptance", kind)
		}
		if !bytes.Equal(before, v.memory) {
			t.Fatal("presentation touched guest")
		}
		cases++
	}
	t.Logf("native glyph source/icon/frame acceptance and refusal cases=%d guest unchanged", cases)
}
