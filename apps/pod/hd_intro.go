package pod

// 主專案規格010「BMA片頭動畫接入」。僅觀測來源，客體不寫入。
import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

const introPixels = 320 * 168
const introImageCodeSHA = "d9ebc163f106be514229472090f2497771df8e36ab5aeca61681f758d32baa54"

type introFile struct {
	Path string `json:"path"`
	SHA  string `json:"sha256"`
}
type introPalette struct {
	SHA   string `json:"sha256"`
	Light uint32 `json:"light"`
}
type introFrameMeta struct {
	Block  int       `json:"block"`
	Frame  int       `json:"frame"`
	Index  introFile `json:"index"`
	Before introFile `json:"before"`
	Writes introFile `json:"writes"`
	PNG    introFile `json:"png"`
}
type introBlockMeta struct {
	SHA      string         `json:"sha256"`
	Size     int            `json:"size"`
	Count    int            `json:"count"`
	Palettes []introPalette `json:"palettes"`
}
type introMeta struct {
	Version int              `json:"version"`
	Scale   int              `json:"scale"`
	Blocks  []introBlockMeta `json:"blocks"`
	Frames  []introFrameMeta `json:"frames"`
}
type introFrame struct {
	block, frame        int
	ref, before, pixels []byte
	writes              []uint16
	previous            *introFrame
}
type IntroHDStats struct {
	Calls, KnownCalls, Full, Prefix, Clear, Fade, Rejected uint64
}
type introHD struct {
	meta                             introMeta
	frames                           []*introFrame
	byCall                           map[[2]int]*introFrame
	blocks                           map[string]int
	active                           *introFrame
	lastStep                         uint64
	seenStep                         bool
	generation, classifiedGeneration uint64
	lastIndex                        []byte
	lastPalette                      [256][3]uint8
	rgba                             []byte
	stats                            IntroHDStats
}

func loadIntroHD(dir string) (*introHD, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "intro.json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	h := &introHD{byCall: map[[2]int]*introFrame{}, blocks: map[string]int{}}
	if err = json.Unmarshal(raw, &h.meta); err != nil {
		return nil, err
	}
	if h.meta.Version != 1 || h.meta.Scale != 3 || len(h.meta.Blocks) != 14 || len(h.meta.Frames) != 133 {
		return nil, fmt.Errorf("intro.json：片頭格式／數量不符")
	}
	checked := func(f introFile) ([]byte, error) {
		p, ok := themeFile(dir, f.Path)
		if !ok {
			return nil, fmt.Errorf("片頭路徑不合法")
		}
		return readChecked(p, f.SHA)
	}
	counts := 0
	for n, b := range h.meta.Blocks {
		if len(b.SHA) != 64 || b.Size < 10 || b.Size > 65535 || b.Count < 1 || b.Count > 32 || len(b.Palettes) > 128 {
			return nil, fmt.Errorf("片頭來源區塊不合法")
		}
		if _, ok := h.blocks[b.SHA]; ok {
			return nil, fmt.Errorf("片頭來源重複")
		}
		h.blocks[b.SHA] = n
		counts += b.Count
		pals := map[string]bool{}
		for _, p := range b.Palettes {
			if len(p.SHA) != 64 || p.Light > 65535 || pals[p.SHA] {
				return nil, fmt.Errorf("片頭色盤不合法")
			}
			pals[p.SHA] = true
		}
	}
	if counts != 133 {
		return nil, fmt.Errorf("片頭來源圖數不符")
	}
	images := map[string][]byte{}
	for n, m := range h.meta.Frames {
		if m.Block < 0 || m.Block >= 14 || m.Frame < 0 || m.Frame >= h.meta.Blocks[m.Block].Count {
			return nil, fmt.Errorf("片頭相位越界")
		}
		key := [2]int{m.Block, m.Frame}
		if h.byCall[key] != nil {
			return nil, fmt.Errorf("片頭相位重複")
		}
		ref, e := checked(m.Index)
		if e != nil || len(ref) != introPixels {
			return nil, fmt.Errorf("片頭索引不合：%v", e)
		}
		before, e := checked(m.Before)
		if e != nil || len(before) != introPixels {
			return nil, fmt.Errorf("片頭前圖不合：%v", e)
		}
		writes, e := checked(m.Writes)
		if e != nil || len(writes)%2 != 0 || len(writes) > introPixels*2 {
			return nil, fmt.Errorf("片頭順序表不合：%v", e)
		}
		f := &introFrame{block: m.Block, frame: m.Frame, ref: ref, before: before}
		seen := make([]bool, introPixels)
		for at := 0; at < len(writes); at += 2 {
			p := binary.LittleEndian.Uint16(writes[at:])
			if int(p) >= introPixels || seen[p] {
				return nil, fmt.Errorf("片頭寫入越界或重複")
			}
			seen[p] = true
			f.writes = append(f.writes, p)
		}
		for at := range ref {
			if ref[at] != before[at] && !seen[at] {
				return nil, fmt.Errorf("片頭來源變更不在寫入表")
			}
		}
		if m.PNG.Path != "" {
			pixels, ok := images[m.PNG.SHA]
			if !ok {
				data, e := checked(m.PNG)
				if e != nil {
					return nil, e
				}
				if len(data) < 26 || data[24] != 8 || data[25] != 6 {
					return nil, fmt.Errorf("片頭PNG需RGBA")
				}
				im, e := png.Decode(bytes.NewReader(data))
				if e != nil {
					return nil, e
				}
				if im.Bounds().Size() != (image.Point{X: 960, Y: 504}) {
					return nil, fmt.Errorf("片頭PNG尺寸不符")
				}
				rgba := image.NewNRGBA(image.Rect(0, 0, 960, 504))
				draw.Draw(rgba, rgba.Bounds(), im, im.Bounds().Min, draw.Src)
				for at := 3; at < len(rgba.Pix); at += 4 {
					if rgba.Pix[at] != 255 {
						return nil, fmt.Errorf("片頭PNG不可透明")
					}
				}
				pixels = rgba.Pix
				images[m.PNG.SHA] = pixels
			}
			f.pixels = pixels
		}
		if n > 0 {
			f.previous = h.frames[n-1]
			if !bytes.Equal(before, f.previous.ref) {
				return nil, fmt.Errorf("片頭前後索引不連續")
			}
		}
		h.frames = append(h.frames, f)
		h.byCall[key] = f
	}
	return h, nil
}

func (h *introHD) reset() {
	if h != nil {
		h.active = nil
		h.rgba = nil
		h.lastIndex = nil
		h.seenStep = false
		h.generation++
	}
}
func (h *introHD) observe(v StepCPU) {
	if h == nil || v.CS() != 0x48a1 || v.IP() != 0x6733 {
		return
	}
	if h.seenStep && h.lastStep == v.Steps() {
		return
	}
	h.seenStep = true
	h.lastStep = v.Steps()
	h.stats.Calls++
	h.generation++
	h.active = nil
	read := func(at uint32, size int) []byte {
		b := make([]byte, size)
		for n := range b {
			b[n] = v.Read8(at + uint32(n))
		}
		return b
	}
	if fmt.Sprintf("%x", sha256.Sum256(read(0x48a10+0x6733, 255))) != introImageCodeSHA {
		return
	}
	stack := linear(v.SS(), v.SP())
	arg := func(off uint32) uint16 { return v.Read16(stack + off) }
	if arg(4) != 2 || arg(6) != 0 || arg(8) != 0 || arg(10) != 0 {
		return
	}
	head := linear(arg(14), arg(12))
	if head < 0x500 || head >= 0xa0000 {
		return
	}
	size := uint32(v.Read16(head)) | uint32(v.Read16(head+2))<<16
	if size < 10 || size > 65535 || head+size > 0xa0000 {
		return
	}
	b, ok := h.blocks[fmt.Sprintf("%x", sha256.Sum256(read(head, int(size))))]
	if !ok || int(size) != h.meta.Blocks[b].Size {
		return
	}
	h.active = h.byCall[[2]int{b, int(arg(16))}]
	if h.active != nil {
		h.stats.KnownCalls++
	}
}

// introPrefix 比對所有點，接受原版有序寫入的一個前綴，不接受任意遮罩。
func introPrefix(idx, base, target []byte, ops []uint16) (int, bool) {
	if len(idx) != introPixels || len(base) != introPixels || len(target) != introPixels {
		return 0, false
	}
	cut := 0
	for n, p := range ops {
		if idx[p] != base[p] {
			if idx[p] != target[p] {
				return 0, false
			}
			cut = n + 1
		}
	}
	expected := append([]byte(nil), base...)
	for _, p := range ops[:cut] {
		expected[p] = target[p]
	}
	return cut, bytes.Equal(idx, expected)
}
func (h *introHD) frame(idx []byte, pal [256][3]uint8) {
	if h == nil {
		return
	}
	if h.classifiedGeneration == h.generation && pal == h.lastPalette && bytes.Equal(idx, h.lastIndex) {
		return
	}
	h.classifiedGeneration = h.generation
	h.lastPalette = pal
	h.lastIndex = append(h.lastIndex[:0], idx...)
	h.rgba = nil
	f := h.active
	if f == nil || len(f.pixels) == 0 || len(idx) != 64000 {
		return
	}
	flat := make([]byte, 0, 768)
	for _, c := range pal {
		flat = append(flat, c[:]...)
	}
	ph := fmt.Sprintf("%x", sha256.Sum256(flat))
	light, valid := uint32(0), false
	for _, p := range h.meta.Blocks[f.block].Palettes {
		if p.SHA == ph {
			light, valid = p.Light, true
			break
		}
	}
	if !valid {
		h.stats.Rejected++
		return
	}
	view := idx[:introPixels]
	mask := []bool(nil)
	var basePixels []byte
	clearPixels := f.pixels
	clear := 0
	if bytes.Equal(view, f.ref) {
		h.stats.Full++
	} else {
		cut, ok := introPrefix(view, f.before, f.ref, f.writes)
		if ok {
			if f.previous != nil {
				basePixels = f.previous.pixels
			}
		} else {
			cut, ok = introPrefix(view, make([]byte, introPixels), f.ref, f.writes)
		}
		if ok {
			mask = make([]bool, introPixels)
			for _, p := range f.writes[:cut] {
				mask[p] = true
			}
			h.stats.Prefix++
		} else {
			for clear < introPixels && view[clear] == 0 {
				clear++
			}
			if clear == 0 {
				h.stats.Rejected++
				return
			}
			if !bytes.Equal(view[clear:], f.ref[clear:]) {
				if !bytes.Equal(view[clear:], f.before[clear:]) || f.previous == nil || len(f.previous.pixels) == 0 {
					h.stats.Rejected++
					return
				}
				clearPixels = f.previous.pixels
			}
			h.stats.Clear++
		}
	}
	out := make([]byte, 960*504*4)
	for p := 0; p < introPixels; p++ {
		pixels := f.pixels
		if clear > 0 {
			pixels = clearPixels
		}
		if mask != nil && !mask[p] {
			pixels = basePixels
		}
		for dy := 0; dy < 3; dy++ {
			at := ((p/320*3+dy)*960 + p%320*3) * 4
			for dx := 0; dx < 3; dx++ {
				q := at + dx*4
				for ch := 0; ch < 3; ch++ {
					if p < clear {
						out[q+ch] = 0
					} else if len(pixels) > 0 {
						out[q+ch] = byte(uint32(pixels[q+ch]) * light / 65535)
					} else if mask != nil && !mask[p] {
						out[q+ch] = pal[idx[p]][ch]
					}
				}
				out[q+3] = 255
			}
		}
	}
	if light < 65535 {
		h.stats.Fade++
	}
	h.rgba = out
}
func (h *introHD) draw(out []byte, scale int) {
	if h == nil || scale != 3 || len(out) != 960*600*4 || len(h.rgba) != 960*504*4 {
		return
	}
	copy(out, h.rgba)
}
func (h *HDLayers) observeIntro(v StepCPU) {
	if h != nil {
		h.intro.observe(v)
	}
}
func (h *HDLayers) drawIntroArt(out []byte, scale int) {
	if h != nil {
		h.intro.draw(out, scale)
	}
}
func (l *Live) IntroHDStats() IntroHDStats {
	if l.HDLayers == nil || l.intro == nil {
		return IntroHDStats{}
	}
	return l.intro.stats
}
