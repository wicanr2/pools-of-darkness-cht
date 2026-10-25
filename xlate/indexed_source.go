package xlate

// IndexedSource is an immutable original frame. Alpha zero in a deferred
// surface means "use this source", not a transparent final output pixel.
type IndexedSource struct {
	w, h, scale int
	indexed     []byte
	palette     [256][3]uint8
}

func NewIndexedSource(indexed []byte, palette [256][3]uint8, w, h, scale int) *IndexedSource {
	if w <= 0 || h <= 0 || scale <= 0 || w > 8192 || h > 8192 || scale > 16 || len(indexed) != w*h {
		return nil
	}
	return &IndexedSource{w: w, h: h, scale: scale, indexed: append([]byte(nil), indexed...), palette: palette}
}
func (s *IndexedSource) valid(out []byte) bool {
	return s != nil && len(out) == s.w*s.h*s.scale*s.scale*4
}
func (s *IndexedSource) colour(x, y int) [3]uint8 {
	return s.palette[s.indexed[(y/s.scale)*s.w+x/s.scale]]
}

func (s *IndexedSource) CellIsOriginal(out []byte, cx, cy int) bool {
	if !s.valid(out) || cx < 0 || cy < 0 || (cx+1)*8 > s.w || (cy+1)*8 > s.h {
		return false
	}
	w, cell := s.w*s.scale, 8*s.scale
	for y := cy * cell; y < (cy+1)*cell; y++ {
		for x := cx * cell; x < (cx+1)*cell; x++ {
			o := (y*w + x) * 4
			if out[o+3] == 0 {
				continue
			}
			c := s.colour(x, y)
			if out[o] != c[0] || out[o+1] != c[1] || out[o+2] != c[2] || out[o+3] != 255 {
				return false
			}
		}
	}
	return true
}

// FillOriginal only paints still-unselected pixels, returning their count.
// A partial-alpha surface is rejected before any mutation.
func (s *IndexedSource) FillOriginal(out []byte) int {
	if !s.valid(out) {
		return -1
	}
	for n := 3; n < len(out); n += 4 {
		if out[n] != 0 && out[n] != 255 {
			return -1
		}
	}
	count := 0
	w := s.w * s.scale
	for y := 0; y < s.h*s.scale; y++ {
		for x := 0; x < w; x++ {
			o := (y*w + x) * 4
			if out[o+3] != 0 {
				continue
			}
			c := s.colour(x, y)
			out[o], out[o+1], out[o+2], out[o+3] = c[0], c[1], c[2], 255
			count++
		}
	}
	return count
}

func (s *IndexedSource) CopyCellTo(dst, src []byte, cx, cy int) {
	if !s.valid(dst) || !s.valid(src) {
		return
	}
	copyRecordedCell(dst, src, s.w, s.scale, cx, cy, true, s)
}

func sameOriginalCell(base, out []byte, w, scale, cx, cy int, source *IndexedSource) bool {
	if source != nil {
		return source.CellIsOriginal(out, cx, cy)
	}
	return cellSame(base, out, w, cx, cy, scale)
}

func copyRecordedCell(dst, src []byte, w, scale, cx, cy int, whole bool, source *IndexedSource) {
	width, cell := w*scale, 8*scale
	for row := 0; row < cell; row++ {
		y := cy*cell + row
		off := (y*width + cx*cell) * 4
		from := row * cell * 4
		if whole {
			from = off
		}
		copy(dst[off:off+cell*4], src[from:from+cell*4])
		if source == nil {
			continue
		}
		for col := 0; col < cell; col++ {
			o := off + col*4
			if dst[o+3] == 0 {
				c := source.colour(cx*cell+col, y)
				dst[o], dst[o+1], dst[o+2], dst[o+3] = c[0], c[1], c[2], 255
			}
		}
	}
}
