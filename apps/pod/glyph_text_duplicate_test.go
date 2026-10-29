package pod

import "testing"

func TestGlyphTextDuplicateEntry(t *testing.T) {
	base := glyphViews(t)[0]
	v := *base
	v.memory = append([]byte(nil), base.memory...)
	stack := uint32(v.SS())*16 + uint32(v.SP())
	v.memory[stack+4] = 'A'
	g := testGlyphTextFont(t)
	g.before(&v)
	if g.call == nil {
		t.Fatal("need known source")
	}
	key := g.call.key
	old := g.call.cell
	old.pending = false
	g.cells[key] = old
	g.call = nil
	g.before(&v)
	next := g.call
	g.before(&v)
	if g.cells[key] != old || g.call != next {
		t.Fatal("duplicate same-step entry discarded visible text before first pixel")
	}
}
