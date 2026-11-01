// Package hdimport 重生本作HD辨識資料。主專案規格012；不寫入原版或DOS存檔。
package hdimport

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Row struct {
	Path    string `json:"path"`
	SHA     string `json:"sha256"`
	Kind    string `json:"kind"`
	Source  string `json:"source"`
	Archive string `json:"archive"`
	Block   int    `json:"block"`
	Frame   int    `json:"frame"`
	Decoder string `json:"decoder"`
	Side    string `json:"side"`
}
type Manifest struct {
	Version   int               `json:"version"`
	Theme     []Row             `json:"theme"`
	Originals map[string]string `json:"original_archives"`
}
type Report struct {
	ManifestSHA   string            `json:"manifest_sha256"`
	Files         map[string]string `json:"files"`
	OriginalFiles int               `json:"original_files"`
}

func Sum(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func validHash(h string) bool {
	b, e := hex.DecodeString(h)
	return e == nil && len(b) == 32 && strings.ToLower(h) == h
}
func cleanName(name string) bool {
	return name != "" && !path.IsAbs(name) && path.Clean(name) == name && name != ".." && !strings.HasPrefix(name, "../") && !strings.ContainsAny(name, `\:`)
}
func under(root, child string) bool {
	rel, e := filepath.Rel(root, child)
	return e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
func read(root, name string) ([]byte, error) {
	if !cleanName(name) {
		return nil, fmt.Errorf("不合法相對路徑：%s", name)
	}
	full := filepath.Join(root, filepath.FromSlash(name))
	resolved, e := filepath.EvalSymlinks(full)
	if e != nil || !under(root, resolved) {
		return nil, fmt.Errorf("來源不存在或超出根目錄：%s", name)
	}
	info, e := os.Stat(resolved)
	if e != nil || !info.Mode().IsRegular() || info.Size() > 64<<20 {
		return nil, fmt.Errorf("來源不是一般檔或過大：%s", name)
	}
	return os.ReadFile(resolved)
}

// Build 沿正式清單建立新主題。output須不存在，且不在bundle／original裡。
// manifest保留既有版本1；其餘字型、版本與包裝欄位由封包層處理。
func Build(raw []byte, bundle, original, output string) (report Report, err error) {
	var m Manifest
	if len(raw) > 8<<20 {
		return report, fmt.Errorf("HD清單過大")
	}
	if e := json.Unmarshal(raw, &m); e != nil {
		return report, e
	}
	if m.Version != 1 || len(m.Theme) == 0 || len(m.Theme) > 4096 || len(m.Originals) == 0 || len(m.Originals) > 32 {
		return report, fmt.Errorf("HD清單版本或數量不合")
	}
	bundle, err = filepath.EvalSymlinks(bundle)
	if err != nil {
		return report, err
	}
	bundle, err = filepath.Abs(bundle)
	if err != nil {
		return report, err
	}
	original, err = filepath.EvalSymlinks(original)
	if err != nil {
		return report, err
	}
	original, err = filepath.Abs(original)
	if err != nil {
		return report, err
	}
	parent, e := filepath.EvalSymlinks(filepath.Dir(output))
	if e != nil {
		return report, e
	}
	parent, e = filepath.Abs(parent)
	if e != nil {
		return report, e
	}
	output = filepath.Join(parent, filepath.Base(output))
	if under(bundle, output) || under(original, output) {
		return report, fmt.Errorf("輸出不可位於來源根目錄")
	}
	if _, e = os.Lstat(output); !os.IsNotExist(e) {
		return report, fmt.Errorf("HD輸出已存在")
	}
	entries, e := os.ReadDir(original)
	if e != nil {
		return report, e
	}
	originalNames := map[string]string{}
	for _, entry := range entries {
		key := strings.ToUpper(entry.Name())
		if _, ok := originalNames[key]; ok {
			return report, fmt.Errorf("原版大小寫檔名重複：%s", key)
		}
		originalNames[key] = entry.Name()
	}
	archives := map[string]map[int][]byte{}
	movies := map[string]map[[2]int]bmaFrame{}
	for name, want := range m.Originals {
		if !cleanName(name) || path.Base(name) != name || !validHash(want) {
			return report, fmt.Errorf("原版清單不合法")
		}
		actual, ok := originalNames[strings.ToUpper(name)]
		if !ok {
			return report, fmt.Errorf("缺少原版%s", name)
		}
		data, e := read(original, actual)
		if e != nil {
			return report, e
		}
		if Sum(data) != want {
			return report, fmt.Errorf("原版%s SHA-256不符", name)
		}
		switch {
		case strings.HasSuffix(name, ".DAX"):
			archives[name], e = dax(data)
			if e != nil {
				return report, fmt.Errorf("%s：%w", name, e)
			}
		case name == "INTRO.BMA" || name == "FINAL.BMA":
			movies[name], e = bma(data)
			if e != nil {
				return report, e
			}
		case name == "INTRO.ACF" || name == "FINAL.ACF":
		default:
			return report, fmt.Errorf("未知原版來源%s", name)
		}
	}
	planned := map[string][]byte{}
	decoded := map[string][][]byte{}
	for _, row := range m.Theme {
		if !cleanName(row.Path) || !validHash(row.SHA) {
			return report, fmt.Errorf("HD輸出路徑或SHA不合法")
		}
		if _, ok := planned[row.Path]; ok {
			return report, fmt.Errorf("HD輸出重複：%s", row.Path)
		}
		var data []byte
		switch row.Kind {
		case "repo-file":
			data, e = read(bundle, row.Source)
		case "original-index":
			if row.Frame < 0 || row.Block < 0 {
				return report, fmt.Errorf("負來源位置")
			}
			cache := fmt.Sprintf("%s:%d:%s", row.Archive, row.Block, row.Decoder)
			frames, ok := decoded[cache]
			if !ok {
				block, ok := archives[row.Archive][row.Block]
				if !ok {
					return report, fmt.Errorf("缺少來源區塊：%s", cache)
				}
				frames, e = decodeFrames(block, row.Decoder)
				if e != nil {
					return report, e
				}
				decoded[cache] = frames
			}
			if row.Frame >= len(frames) {
				return report, fmt.Errorf("來源影格越界")
			}
			data = frames[row.Frame]
		case "original-intro":
			archive := row.Archive
			if archive == "" {
				archive = "INTRO.BMA"
			}
			if archive != "INTRO.BMA" && archive != "FINAL.BMA" {
				return report, fmt.Errorf("未知BMA來源")
			}
			frame, ok := movies[archive][[2]int{row.Block, row.Frame}]
			if !ok {
				return report, fmt.Errorf("BMA來源影格越界")
			}
			switch row.Decoder {
			case "bma-index":
				switch row.Side {
				case "before":
					data = frame.before
				case "after":
					data = frame.after
				default:
					return report, fmt.Errorf("BMA前後狀態不合")
				}
			case "bma-writes":
				data = frame.writes
			default:
				return report, fmt.Errorf("未知BMA解碼器")
			}
		default:
			return report, fmt.Errorf("未知HD來源kind")
		}
		if e != nil {
			return report, e
		}
		if Sum(data) != row.SHA {
			return report, fmt.Errorf("HD來源重生SHA不符：%s", row.Path)
		}
		planned[row.Path] = data
	}
	// 全部來源已核對；只發布完整的新目錄，不碰既有快取。
	stage, e := os.MkdirTemp(parent, ".pod-hd-import-")
	if e != nil {
		return report, e
	}
	defer os.RemoveAll(stage)
	for name, data := range planned {
		dest := filepath.Join(stage, filepath.FromSlash(name))
		if e = os.MkdirAll(filepath.Dir(dest), 0755); e != nil {
			return report, e
		}
		if e = os.WriteFile(dest, data, 0644); e != nil {
			return report, e
		}
	}
	if e = os.Rename(stage, output); e != nil {
		return report, e
	}
	report = Report{ManifestSHA: Sum(raw), Files: map[string]string{}, OriginalFiles: len(m.Originals)}
	for name, data := range planned {
		report.Files[name] = Sum(data)
	}
	return report, nil
}
