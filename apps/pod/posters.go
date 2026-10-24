package pod

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
)

// PosterSet 僅合成遊戲畫布外側，素材為主專案規格 023 的既有 PNG；不讀寫機器狀態。
type PosterSet struct {
	base  map[int][2]*image.NRGBA
	text  map[string]map[int][2]*image.NRGBA
	cache map[string][]byte
}

func LoadPosters(dir string) (*PosterSet, error) {
	p := &PosterSet{base: map[int][2]*image.NRGBA{}, text: map[string]map[int][2]*image.NRGBA{}, cache: map[string][]byte{}}
	load := func(name string, scale int) (*image.NRGBA, error) {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		if len(data) > 8<<20 {
			return nil, fmt.Errorf("%s：海報檔案過大", name)
		}
		c, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if c.Width != 80*scale || c.Height != 200*scale {
			return nil, fmt.Errorf("%s：海報尺寸 %d×%d 不合", name, c.Width, c.Height)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		n := image.NewNRGBA(image.Rect(0, 0, c.Width, c.Height))
		draw.Draw(n, n.Bounds(), im, im.Bounds().Min, draw.Src)
		return n, nil
	}
	for _, scale := range []int{2, 3} {
		var pair [2]*image.NRGBA
		for side, name := range []string{"left", "right"} {
			im, err := load(fmt.Sprintf("poster-%s-%dx.png", name, scale), scale)
			if err != nil {
				return nil, err
			}
			for i := 3; i < len(im.Pix); i += 4 {
				if im.Pix[i] != 255 {
					return nil, fmt.Errorf("%s：海報底圖須全不透明", name)
				}
			}
			pair[side] = im
		}
		p.base[scale] = pair
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	pattern := regexp.MustCompile(`^poster-text-(?:left|right)\.([A-Za-z0-9-]+)-[23]x\.png$`)
	for _, file := range files {
		m := pattern.FindStringSubmatch(file.Name())
		if m == nil {
			continue
		}
		code := m[1]
		if p.text[code] != nil {
			continue
		}
		variants := map[int][2]*image.NRGBA{}
		for _, scale := range []int{2, 3} {
			var pair [2]*image.NRGBA
			for side, name := range []string{"left", "right"} {
				im, err := load(fmt.Sprintf("poster-text-%s.%s-%dx.png", name, code, scale), scale)
				if err != nil {
					return nil, err
				}
				pair[side] = im
			}
			variants[scale] = pair
		}
		p.text[code] = variants
	}
	return p, nil
}

// Compose 的中央 bytes 逐列原樣複製。未知語系只用無字底圖，不沿用上一個語系。
func (p *PosterSet) Compose(dst, game []byte, scale int, locale string) []byte {
	if p == nil {
		return game
	}
	if (scale != 2 && scale != 3) || len(game) != 320*200*scale*scale*4 {
		return nil
	}
	if p.text[locale] == nil {
		locale = ""
	}
	key := fmt.Sprintf("%d:%s", scale, locale)
	background := p.cache[key]
	w, h := 480*scale, 200*scale
	if background == nil {
		im := image.NewNRGBA(image.Rect(0, 0, w, h))
		for i := 3; i < len(im.Pix); i += 4 {
			im.Pix[i] = 255
		}
		for side, x := range []int{0, 400 * scale} {
			r := image.Rect(x, 0, x+80*scale, h)
			draw.Draw(im, r, p.base[scale][side], image.Point{}, draw.Src)
			if text := p.text[locale]; text != nil {
				draw.Draw(im, r, text[scale][side], image.Point{}, draw.Over)
			}
		}
		background = im.Pix
		p.cache[key] = background
	}
	if cap(dst) < len(background) {
		dst = make([]byte, len(background))
	} else {
		dst = dst[:len(background)]
	}
	copy(dst, background)
	for y := 0; y < h; y++ {
		copy(dst[(y*w+80*scale)*4:(y*w+400*scale)*4], game[y*320*scale*4:(y+1)*320*scale*4])
	}
	return dst
}
