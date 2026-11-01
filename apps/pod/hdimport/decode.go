package hdimport

import (
	"encoding/binary"
	"fmt"
)

func dax(data []byte) (map[int][]byte, error) {
	if len(data) < 2 || len(data) > 16<<20 {
		return nil, fmt.Errorf("DAX長度不合")
	}
	head := int(binary.LittleEndian.Uint16(data)) + 2
	if head > len(data) || (head-2)%9 != 0 {
		return nil, fmt.Errorf("DAX目錄不合")
	}
	blocks := map[int][]byte{}
	for p := 2; p < head; p += 9 {
		id := int(data[p])
		off := uint64(binary.LittleEndian.Uint32(data[p+1:]))
		count := int(binary.LittleEndian.Uint16(data[p+5:]))
		packed := uint64(binary.LittleEndian.Uint16(data[p+7:]))
		base := uint64(head) + off
		if base+packed > uint64(len(data)) {
			return nil, fmt.Errorf("DAX資料越界")
		}
		if _, ok := blocks[id]; ok {
			return nil, fmt.Errorf("DAX區塊重複")
		}
		src := data[base : base+packed]
		out := make([]byte, 0, count)
		for q := 0; q < len(src) && len(out) < count; {
			c := int(src[q])
			q++
			if c < 128 {
				n := c + 1
				if q+n > len(src) || len(out)+n > count {
					return nil, fmt.Errorf("DAX literal越界")
				}
				out = append(out, src[q:q+n]...)
				q += n
			} else {
				n := 256 - c
				if q >= len(src) || len(out)+n > count {
					return nil, fmt.Errorf("DAX repeat越界")
				}
				for i := 0; i < n; i++ {
					out = append(out, src[q])
				}
				q++
			}
		}
		if len(out) != count {
			return nil, fmt.Errorf("DAX展開長度不同")
		}
		blocks[id] = out
	}
	return blocks, nil
}

func decodeFrames(data []byte, decoder string) ([][]byte, error) {
	if decoder == "wall-8x8" {
		if len(data) < 434 || (len(data)-434)%64 != 0 {
			return nil, fmt.Errorf("牆面表頭／長度不合")
		}
		frames := [][]byte{}
		for n := 434; n < len(data); n += 64 {
			frames = append(frames, data[n:n+64])
		}
		return frames, nil
	}
	if decoder == "block-vga" {
		if len(data) < 10 {
			return nil, fmt.Errorf("VGA表頭不足")
		}
		h, w, n := int(data[0]), int(data[1])*8, int(binary.LittleEndian.Uint16(data[6:]))
		base, colors := int(data[8]), int(data[9])+1
		paletteEnd := 10 + colors*3
		start := len(data) - w*h*n
		if w < 1 || w > 320 || h < 1 || h > 200 || n < 1 || n > 64 || base+colors > 256 || paletteEnd > start {
			return nil, fmt.Errorf("VGA尺寸或範圍不合")
		}
		for _, v := range data[10:paletteEnd] {
			if v > 63 {
				return nil, fmt.Errorf("VGA色盤超界")
			}
		}
		frames := [][]byte{}
		for i := 0; i < n; i++ {
			frames = append(frames, data[start+i*w*h:start+(i+1)*w*h])
		}
		return frames, nil
	}
	if decoder != "sprite-vga" {
		return nil, fmt.Errorf("未知索引解碼器")
	}
	if len(data) < 11 {
		return nil, fmt.Errorf("sprite表頭不足")
	}
	h, w := int(binary.LittleEndian.Uint16(data)), int(binary.LittleEndian.Uint16(data[2:]))*8
	extra, base, colors := int(data[8]), int(data[9]), int(data[10])
	start := colors*3 + 132 + extra*3
	paletteEnd := 11 + colors*3
	if w < 1 || w > 320 || h < 1 || h > 200 || colors < 1 || base+colors > 256 || paletteEnd >= start || start >= len(data) {
		return nil, fmt.Errorf("sprite尺寸或範圍不合")
	}
	for _, v := range data[11:paletteEnd] {
		if v > 63 {
			return nil, fmt.Errorf("sprite色盤超界")
		}
	}
	out := []byte{}
	size := w * h
	for q := start; q < len(data); {
		c := int(data[q])
		q++
		if c < 128 {
			n := c + 1
			if q+n > len(data) {
				return nil, fmt.Errorf("sprite literal越界")
			}
			out = append(out, data[q:q+n]...)
			q += n
		} else {
			n := 257 - c
			if q >= len(data) {
				return nil, fmt.Errorf("sprite repeat越界")
			}
			for i := 0; i < n; i++ {
				out = append(out, data[q])
			}
			q++
		}
		if len(out) > 64*size {
			return nil, fmt.Errorf("sprite展開超過上限")
		}
	}
	if len(out) == 0 || len(out)%size != 0 {
		return nil, fmt.Errorf("sprite展開長度不合")
	}
	count := len(out) / size
	main := int(data[colors*3+127])
	if main >= count {
		return nil, fmt.Errorf("sprite基底越界")
	}
	frames := [][]byte{}
	baseline := out[main*size : (main+1)*size]
	for i := 0; i < count; i++ {
		frame := append([]byte(nil), out[i*size:(i+1)*size]...)
		if i != main {
			for j := range frame {
				frame[j] ^= baseline[j]
			}
		}
		frames = append(frames, frame)
	}
	return frames, nil
}
