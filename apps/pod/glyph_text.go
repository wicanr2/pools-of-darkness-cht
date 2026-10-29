package pod

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/wicanr2/dosgolem/xlate"
)

const glyphTextSourceSHA = "72fb1adcd5698a5e9770b91a499c347c7c8b05c567b93a8cca1f7249c49b4a7e"

type glyphTextCell struct {
	x, y       int
	ch, bg, fg byte
	source     [64]byte
	pending    bool
}
type glyphTextCall struct {
	step   uint64
	ret    uint32
	ss, sp uint16
	key    int
	cell   *glyphTextCell
}
type glyphText struct {
	font  *xlate.Font
	cells map[int]*glyphTextCell
	call  *glyphTextCall
}

// SetGlyphFont selects an already rasterized 16/24-pixel display font.
// It never replaces the guest's 8x8 source table.
func (l *Live) SetGlyphFont(font *xlate.Font) {
	if l.glyphText == nil {
		l.glyphText = newGlyphText(font)
	} else {
		l.glyphText.font = font
	}
}

type GlyphTextCellReport struct {
	X, Y                         int
	Code, Background, Foreground byte
	Pending                      bool
	SourceSHA256                 string
}
type GlyphTextReport struct {
	Enabled           bool
	Font              string
	CellPixels        int
	ASCIISourceSHA256 string
	Cells             []GlyphTextCellReport
}

func (g *glyphText) nativeFont(scale int) *xlate.Font {
	if g == nil || g.font == nil || (scale != 2 && scale != 3) {
		return nil
	}
	font, k, dx, dy := g.font.ForScale(scale)
	if font == nil || k != 1 || dx != 0 || dy != 0 || font.W != 8*scale || font.H != 8*scale {
		return nil
	}
	return font
}

// GlyphTextReport describes the last presented source, for screenshot receipts.
func (l *Live) GlyphTextReport(scale int) GlyphTextReport {
	out := GlyphTextReport{Enabled: true, ASCIISourceSHA256: glyphTextSourceSHA}
	if l == nil || l.glyphText == nil {
		return out
	}
	g := l.glyphText
	font := g.nativeFont(scale)
	if font == nil {
		return out
	}
	out.Font, out.CellPixels = font.Name, font.W
	for key := 0; key < 1000; key++ {
		if c := g.cells[key]; c != nil {
			bitmap, known := font.Glyphs[rune(c.ch)]
			if !known || len(bitmap) != (font.W+7)/8*font.H {
				continue
			}
			digest := sha256.Sum256(c.source[:])
			out.Cells = append(out.Cells, GlyphTextCellReport{X: c.x, Y: c.y, Code: c.ch, Background: c.bg, Foreground: c.fg, Pending: c.pending, SourceSHA256: hex.EncodeToString(digest[:])})
		}
	}
	return out
}
func newGlyphText(font *xlate.Font) *glyphText {
	return &glyphText{font: font, cells: map[int]*glyphTextCell{}}
}
func (g *glyphText) reset() {
	if g != nil {
		g.cells = map[int]*glyphTextCell{}
		g.call = nil
	}
}
func (g *glyphText) before(v StepCPU) {
	if g == nil {
		return
	}
	at := uint32(v.CS())*16 + uint32(v.IP())
	if c := g.call; c != nil && at == c.ret && v.SS() == c.ss && v.SP() == c.sp {
		c.cell.pending = false
		g.call = nil
	}
	if at != 0x1c486 || v.Read8(0x449) != 0x13 {
		return
	}
	if g.call != nil && g.call.step == v.Steps() && g.call.ss == v.SS() && g.call.sp == v.SP()+14 {
		return
	}
	if g.call != nil {
		delete(g.cells, g.call.key)
		g.call = nil
	}
	stack := uint32(v.SS())*16 + uint32(v.SP())
	if uint32(v.SP())+14 > 65536 || stack < 0x500 || stack+14 > 0xa0000 {
		return
	}
	ch := v.Read8(stack + 4)
	if ch < 0x20 || ch > 0x7e {
		return
	}
	var code [93]byte
	for i := range code {
		code[i] = v.Read8(0x1c486 + uint32(i))
	}
	hash := sha256.Sum256(code[:])
	if hex.EncodeToString(hash[:]) != glyphCallCodeSHA {
		return
	}
	y, x := v.Read16(stack+10), v.Read16(stack+12)
	if x >= 40 || y >= 25 {
		return
	}
	table := uint32(v.DS())*16 + 0x9d86
	if table+4 > 0xa0000 {
		return
	}
	base, seg := v.Read16(table), v.Read16(table+2)
	src := uint32(seg)*16 + uint32(base)
	if uint32(base)+1016 > 65536 || src+1016 > 0xa0000 {
		return
	}
	var original [760]byte
	for i := range original {
		original[i] = v.Read8(src + 256 + uint32(i))
	}
	hash = sha256.Sum256(original[:])
	if hex.EncodeToString(hash[:]) != glyphTextSourceSHA {
		return
	}
	cell := &glyphTextCell{x: int(x) * 8, y: int(y) * 8, ch: ch, bg: v.Read8(stack + 6), fg: v.Read8(stack + 8), pending: true}
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			value := cell.bg
			if original[(int(ch)-32)*8+row]&(0x80>>col) != 0 {
				value = cell.fg
			}
			cell.source[row*8+col] = value
		}
	}
	g.call = &glyphTextCall{step: v.Steps(), ret: uint32(v.Read16(stack+2))*16 + uint32(v.Read16(stack)), ss: v.SS(), sp: v.SP() + 14, key: int(y)*40 + int(x), cell: cell}
}
func (g *glyphText) write(offset uint32) {
	if g == nil || offset >= 64000 || g.call == nil {
		return
	}
	key := int(offset/320/8)*40 + int(offset%320/8)
	if key == g.call.key {
		g.cells[key] = g.call.cell
	}
}
func (g *glyphText) frame(indexed []byte) {
	if g == nil {
		return
	}
	if len(indexed) != 64000 {
		g.reset()
		return
	}
	for key, c := range g.cells {
		if c.pending {
			if g.call == nil || g.call.cell != c {
				delete(g.cells, key)
			}
			continue
		}
		valid := true
		for row := 0; row < 8 && valid; row++ {
			for col := 0; col < 8; col++ {
				if indexed[(c.y+row)*320+c.x+col] != c.source[row*8+col] {
					valid = false
					break
				}
			}
		}
		if !valid {
			delete(g.cells, key)
		}
	}
}
func (g *glyphText) draw(dst []byte, scale int, pal [256][3]byte) int {
	font := g.nativeFont(scale)
	if font == nil || len(dst) != 320*200*scale*scale*4 {
		return 0
	}
	layer := &xlate.Layer{W: 320, H: 200}
	for key := 0; key < 1000; key++ {
		if c := g.cells[key]; c != nil {
			bitmap, known := font.Glyphs[rune(c.ch)]
			if !known || len(bitmap) != (font.W+7)/8*font.H {
				continue
			}
			stamp := &xlate.Stamp{X: c.x, Y: c.y, Cells: 1, CellW: 8, CellH: 8, State: xlate.Shown, FG: pal[c.fg], BG: pal[c.bg]}
			if !c.pending {
				stamp.Font = font
				stamp.GlyphScale = 1
				stamp.Text = []rune{rune(c.ch)}
			}
			layer.Stamps = append(layer.Stamps, stamp)
		}
	}
	layer.Draw(dst, scale, nil)
	return len(layer.Stamps)
}
