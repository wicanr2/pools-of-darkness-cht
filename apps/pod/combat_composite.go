package pod

// 主專案規格010：完整特效前景、可見CPIC及地形組成當幀完整來源。
import "bytes"

type compositeEffect struct {
	pattern, ax, ay int
	anchor          []byte
}
type compositePrefix struct {
	compositeEffect
	rows   int
	source []byte
}
type compositeSprite struct {
	pattern, x, y int
	conflict      bool
}
type compositeTerrain struct{ tile, x, y int }
type combatComposite struct {
	effects               []compositeEffect
	prefixes              []compositePrefix
	deferCells            [21 * 21]bool
	tiles                 map[string][]int
	source                []byte
	effectMask, cover     []bool
	owner                 []int
	effectsNow, actorsNow []compositeSprite
	terrainNow            []compositeTerrain
}

func newCombatComposite(c *combatHD, a *actorHD) *combatComposite {
	if c == nil || a == nil {
		return nil
	}
	h := &combatComposite{tiles: map[string][]int{}, effectMask: make([]bool, 64000), cover: make([]bool, 64000), owner: make([]int, 64000)}
	for sig, n := range c.vgaLookup {
		h.tiles[sig.palette] = append(h.tiles[sig.palette], n)
	}
	for n, p := range a.patterns {
		if !p.vgaEffect || !p.freeFull {
			continue
		}
		// 高色號錨點不能由地形或CPIC的色盤誤命中，之後仍驗完整前景。
		best, ax, ay := 0, 0, 0
		for y := 0; y < p.h; y++ {
			for x := 0; x < p.w; {
				start := x
				for x < p.w && p.full[y*p.w+x] >= 64 && p.full[y*p.w+x] < 96 {
					x++
				}
				if x-start > best {
					best, ax, ay = x-start, start, y
				}
				if x == start {
					x++
				}
			}
		}
		if best > 0 {
			h.effects = append(h.effects, compositeEffect{n, ax, ay, append([]byte(nil), p.full[ay*p.w+ax:ay*p.w+ax+best]...)})
		}
	}
	// 小型效果的已完成8／16列。完整來源前綴需在全部效果及鏡像中唯一。
	counts := map[string]int{}
	for _, p := range a.patterns {
		if !p.vgaEffect || len(p.full) != 576 {
			continue
		}
		for _, rows := range []int{8, 16} {
			prefix := append([]byte(nil), p.full[:rows*24]...)
			for n, v := range prefix {
				if v == 8 {
					prefix[n] = 32
				}
			}
			counts[string(prefix)]++
		}
	}
	for n, p := range a.patterns {
		if !p.vgaEffect || len(p.full) != 576 || len(p.ink) >= 144 {
			continue
		}
		for _, rows := range []int{8, 16} {
			prefix := append([]byte(nil), p.full[:rows*24]...)
			for i, v := range prefix {
				if v == 8 {
					prefix[i] = 32
				}
			}
			if counts[string(prefix)] != 1 {
				continue
			}
			best, ax, ay := 0, 0, 0
			var used [256]bool
			colors := 0
			for _, v := range prefix {
				if !used[v] {
					used[v] = true
					colors++
				}
			}
			if colors < 3 {
				continue
			}
			for y := 0; y < rows; y++ {
				for x := 0; x < 24; {
					start := x
					for x < 24 && prefix[y*24+x] >= 64 && prefix[y*24+x] < 96 {
						x++
					}
					if x-start > best {
						best, ax, ay = x-start, start, y
					}
					if x == start {
						x++
					}
				}
			}
			if best > 0 {
				h.prefixes = append(h.prefixes, compositePrefix{compositeEffect{n, ax, ay, append([]byte(nil), prefix[ay*24+ax:ay*24+ax+best]...)}, rows, prefix})
			}
		}
	}
	return h
}

func foregroundPresent(p *actorPattern, idx []byte, x, y int, mask []bool) (visible, hidden, colors int, ok bool) {
	var used [256]bool
	for _, pt := range p.ink {
		at := (y+pt.y)*320 + x + pt.x
		if mask != nil && mask[at] {
			hidden++
			continue
		}
		if idx[at] != pt.color {
			return 0, 0, 0, false
		}
		visible++
		if !used[pt.color] {
			used[pt.color] = true
			colors++
		}
	}
	return visible, hidden, colors, true
}

func (h *combatComposite) frame(idx []byte, pal [256][3]byte, c *combatHD, a *actorHD) {
	if h == nil {
		return
	}
	h.effectsNow = h.effectsNow[:0]
	h.actorsNow = h.actorsNow[:0]
	h.terrainNow = h.terrainNow[:0]
	h.source = h.source[:0]
	clear(h.effectMask)
	clear(h.cover)
	clear(h.deferCells[:])
	if len(idx) != 64000 || c == nil || a == nil || len(c.matches) < 3 {
		return
	}
	if a.effectPaletteOK && paletteBankHash(pal, 32, 64) == terrainVGAPalettes["dungcom-005"] {
		for _, p := range h.prefixes {
			for from := 0; from+len(p.anchor) <= len(idx); {
				off := bytes.Index(idx[from:], p.anchor)
				if off < 0 {
					break
				}
				at := from + off
				from = at + 1
				x, y := at%320-p.ax, at/320-p.ay
				if x < 8 || y < 8 || x+24 > 176 || y+24 > 176 || (x-8)%8 != 0 || (y-8)%8 != 0 {
					continue
				}
				ok := true
				for row := 0; row < p.rows; row++ {
					if !bytes.Equal(idx[(y+row)*320+x:(y+row)*320+x+24], p.source[row*24:(row+1)*24]) {
						ok = false
						break
					}
				}
				if !ok {
					continue
				}
				for yy := 0; yy < p.rows; yy++ {
					for xx := 0; xx < 24; xx++ {
						if p.source[yy*24+xx] != 32 {
							h.deferCells[((y-8+yy)/8)*21+(x-8+xx)/8] = true
						}
					}
				}
			}
		}
	}
	// 未見特效高色號時不走此分支。
	high := false
	for y := 8; y < 176 && !high; y++ {
		for x := 8; x < 176; x++ {
			v := idx[y*320+x]
			if v >= 64 && v < 96 {
				high = true
				break
			}
		}
	}
	for _, e := range h.effects {
		if !high || !a.effectPaletteOK {
			break
		}
		p := &a.patterns[e.pattern]
		for from := 0; from+len(e.anchor) <= len(idx); {
			off := bytes.Index(idx[from:], e.anchor)
			if off < 0 {
				break
			}
			at := from + off
			from = at + 1
			x, y := at%320-e.ax, at/320-e.ay
			if x < 8 || y < 8 || x+p.w > 176 || y+p.h > 176 {
				continue
			}
			if _, _, _, ok := foregroundPresent(p, idx, x, y, nil); ok {
				h.effectsNow = append(h.effectsNow, compositeSprite{pattern: e.pattern, x: x, y: y})
			}
		}
	}
	// 所有相交前景都拒絕，不以列舉順序選第一個。
	owner := h.owner
	clear(owner)
	for n, m := range h.effectsNow {
		for _, pt := range a.patterns[m.pattern].ink {
			at := (m.y+pt.y)*320 + m.x + pt.x
			if prev := owner[at]; prev != 0 {
				h.effectsNow[n].conflict = true
				h.effectsNow[prev-1].conflict = true
			} else {
				owner[at] = n + 1
			}
		}
	}
	for _, m := range h.effectsNow {
		if m.conflict {
			continue
		}
		for _, pt := range a.patterns[m.pattern].ink {
			h.effectMask[(m.y+pt.y)*320+m.x+pt.x] = true
		}
	}
	copy(h.cover, h.effectMask)
	// 已完整匹配的CPIC也屬已證實前景。
	for _, m := range a.matches {
		p := &a.patterns[m.pattern]
		if m.conflict || !p.vga || !a.vgaPaletteOK {
			continue
		}
		for _, pt := range p.ink {
			h.cover[(m.y+pt.y)*320+m.x+pt.x] = true
		}
	}
	if a.vgaPaletteOK {
		for n, p := range a.patterns {
			if !p.vga || len(p.full) != 0 {
				continue
			}
			for y := 8; y+p.h <= 176; y += 24 {
				for x := 8; x+p.w <= 176; x += 24 {
					visible, hidden, colors, ok := foregroundPresent(&p, idx, x, y, h.effectMask)
					if !ok || hidden == 0 || visible < 144 || colors < 3 {
						continue
					}
					witness := false
					for _, pt := range p.ink {
						at := (y+pt.y)*320 + x + pt.x
						if h.effectMask[at] && idx[at] != pt.color {
							witness = true
							break
						}
					}
					if !witness {
						continue
					} // 原版像素須證明效果在CPIC上方。
					blocked := false
					for _, pt := range p.ink {
						at := (y+pt.y)*320 + x + pt.x
						if h.cover[at] && !h.effectMask[at] {
							blocked = true
							break
						}
					}
					if !blocked {
						h.actorsNow = append(h.actorsNow, compositeSprite{pattern: n, x: x, y: y})
					}
				}
			}
		}
		clear(owner)
		for n, m := range h.actorsNow {
			for _, pt := range a.patterns[m.pattern].ink {
				at := (m.y+pt.y)*320 + m.x + pt.x
				if h.effectMask[at] {
					continue
				}
				if prev := owner[at]; prev != 0 {
					h.actorsNow[n].conflict = true
					h.actorsNow[prev-1].conflict = true
				} else {
					owner[at] = n + 1
				}
			}
		}
		for _, m := range h.actorsNow {
			if m.conflict {
				continue
			}
			for _, pt := range a.patterns[m.pattern].ink {
				h.cover[(m.y+pt.y)*320+m.x+pt.x] = true
			}
		}
	}
	var complete [49]bool
	for _, m := range c.matches {
		complete[(m.y-8)/24*7+(m.x-8)/24] = true
	}
	candidates := h.tiles[paletteBankHash(pal, 32, 64)]
	for y := 8; y <= 152; y += 24 {
		for x := 8; x <= 152; x += 24 {
			if complete[(y-8)/24*7+(x-8)/24] {
				continue
			}
			chosen := -1
			for _, n := range candidates {
				p := &c.tiles[n]
				visible, hidden, colors := 0, 0, 0
				var used [256]bool
				ok := true
				for yy := 0; yy < 24 && ok; yy++ {
					for xx := 0; xx < 24; xx++ {
						at := (y+yy)*320 + x + xx
						if h.cover[at] {
							hidden++
							continue
						}
						v := idx[at]
						if v != p.reference[yy*24+xx] {
							ok = false
							break
						}
						visible++
						if !used[v] {
							used[v] = true
							colors++
						}
					}
				}
				if !ok || hidden == 0 || visible < 144 || colors < 3 {
					continue
				}
				if chosen >= 0 {
					chosen = -2
					break
				}
				chosen = n
			}
			if chosen >= 0 {
				h.terrainNow = append(h.terrainNow, compositeTerrain{chosen, x, y})
			}
		}
	}
	h.source = append(h.source, idx...)
}

func (h *combatComposite) drawTerrain(out, idx []byte, scale int, c *combatHD) {
	if h == nil || scale != 3 || len(out) != 960*600*4 || !bytes.Equal(h.source, idx) || len(idx) != 64000 {
		return
	}
	for _, m := range h.terrainNow {
		p := &c.tiles[m.tile]
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				if h.cover[(m.y+y)*320+m.x+x] {
					continue
				}
				for dy := 0; dy < 3; dy++ {
					src := ((y*3+dy)*72 + x*3) * 4
					dst := (((m.y+y)*3+dy)*960 + (m.x+x)*3) * 4
					copy(out[dst:dst+12], p.pixels[src:src+12])
				}
			}
		}
	}
}

func (h *combatComposite) drawActors(out, idx []byte, scale int, a *actorHD) {
	if h == nil || scale != 3 || len(out) != 960*600*4 || !bytes.Equal(h.source, idx) || len(idx) != 64000 {
		return
	}
	for group, matches := range [][]compositeSprite{h.actorsNow, h.effectsNow} {
		for _, m := range matches {
			if m.conflict {
				continue
			}
			p := &a.patterns[m.pattern]
			for _, pt := range p.ink {
				if group == 0 && h.effectMask[(m.y+pt.y)*320+m.x+pt.x] {
					continue
				}
				for dy := 0; dy < 3; dy++ {
					src := ((pt.y*3+dy)*p.w*3 + pt.x*3) * 4
					dst := (((m.y+pt.y)*3+dy)*960 + (m.x+pt.x)*3) * 4
					copy(out[dst:dst+12], p.pixels[src:src+12])
				}
			}
		}
	}
}
