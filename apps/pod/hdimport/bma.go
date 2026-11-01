package hdimport

import (
	"encoding/binary"
	"fmt"
)

const artSize = 320 * 168

type bmaFrame struct{ before, after, writes []byte }

func bma(data []byte) (map[[2]int]bmaFrame, error) {
	if len(data) == 0 || len(data) > 16<<20 {
		return nil, fmt.Errorf("BMA長度不合")
	}
	frames := map[[2]int]bmaFrame{}
	canvas := make([]byte, artSize)
	block := 0
	for at := 0; at < len(data); block++ {
		if at+6 > len(data) || block >= 1024 {
			return nil, fmt.Errorf("BMA區塊表頭不合")
		}
		size := int(binary.LittleEndian.Uint32(data[at:]))
		count := int(binary.LittleEndian.Uint16(data[at+4:]))
		table := 6 + count*4
		if count < 1 || count > 4096 || len(frames)+count > 4096 || size < table || size > len(data)-at {
			return nil, fmt.Errorf("BMA區塊範圍不合")
		}
		offsets := make([]int, count+1)
		for n := 0; n < count; n++ {
			offsets[n] = int(binary.LittleEndian.Uint32(data[at+6+n*4:]))
		}
		offsets[count] = size
		if offsets[0] != table {
			return nil, fmt.Errorf("BMA目錄起點不合")
		}
		for n := 0; n < count; n++ {
			if offsets[n] < table || offsets[n] >= offsets[n+1] || offsets[n+1] > size {
				return nil, fmt.Errorf("BMA圖面越界")
			}
			src := data[at+offsets[n] : at+offsets[n+1]]
			if len(src) < 5 {
				return nil, fmt.Errorf("BMA影像表頭不足")
			}
			w, h := int(binary.LittleEndian.Uint16(src)), int(binary.LittleEndian.Uint16(src[2:]))
			if w < 1 || w > 320 || h < 1 || h > 200 {
				return nil, fmt.Errorf("BMA尺寸越界")
			}
			before := append([]byte(nil), canvas...)
			writes := []byte{}
			seen := make([]bool, artSize)
			q, ended := 4, false
			for q < len(src) {
				y := int(src[q])
				q++
				if y == 255 {
					if q != len(src) {
						return nil, fmt.Errorf("BMA終止後仍有資料")
					}
					ended = true
					break
				}
				if y >= h {
					return nil, fmt.Errorf("BMA列越界")
				}
				for {
					if q+4 > len(src) {
						return nil, fmt.Errorf("BMA跨度表頭不足")
					}
					raw := binary.LittleEndian.Uint16(src[q:])
					x := int(raw & 0x7fff)
					width, packed := int(src[q+2]), int(src[q+3])
					q += 4
					if width == 0 || packed == 0 || x+width > w || q+packed > len(src) {
						return nil, fmt.Errorf("BMA跨度越界")
					}
					end := q + packed
					pixels := []byte{}
					for q < end {
						code := int(src[q])
						q++
						length := (code >> 1) + 1
						if len(pixels)+length > width {
							return nil, fmt.Errorf("BMA RLE超界")
						}
						if code&1 != 0 {
							if q >= end {
								return nil, fmt.Errorf("BMA repeat缺色號")
							}
							for i := 0; i < length; i++ {
								pixels = append(pixels, src[q])
							}
							q++
						} else {
							if q+length > end {
								return nil, fmt.Errorf("BMA literal越界")
							}
							pixels = append(pixels, src[q:q+length]...)
							q += length
						}
					}
					if len(pixels) != width {
						return nil, fmt.Errorf("BMA跨度展開長度不同")
					}
					for i, c := range pixels {
						pos := y*320 + x + i
						if pos >= artSize {
							continue
						}
						if seen[pos] {
							return nil, fmt.Errorf("BMA重複寫入尚未支援")
						}
						seen[pos] = true
						canvas[pos] = c
						writes = binary.LittleEndian.AppendUint16(writes, uint16(pos))
					}
					if raw&0x8000 != 0 {
						break
					}
				}
			}
			if !ended {
				return nil, fmt.Errorf("BMA缺終止符")
			}
			frames[[2]int{block, n}] = bmaFrame{before: before, after: append([]byte(nil), canvas...), writes: writes}
		}
		at += size
	}
	return frames, nil
}
