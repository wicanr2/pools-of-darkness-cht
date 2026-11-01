package hdimport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Verify 核對已產生快取的全部檔案；未知檔案與缺項均拒絕。
func Verify(raw []byte, dir string) error {
	var m Manifest
	if len(raw) > 8<<20 {
		return fmt.Errorf("HD清單過大")
	}
	if e := json.Unmarshal(raw, &m); e != nil {
		return e
	}
	if m.Version != 1 || len(m.Theme) < 1 || len(m.Theme) > 4096 {
		return fmt.Errorf("HD清單版本／數量不合")
	}
	root, e := filepath.EvalSymlinks(dir)
	if e != nil {
		return e
	}
	root, e = filepath.Abs(root)
	if e != nil {
		return e
	}
	expected := map[string]string{}
	for _, row := range m.Theme {
		if !cleanName(row.Path) || !validHash(row.SHA) {
			return fmt.Errorf("HD快取清單不合")
		}
		if _, ok := expected[row.Path]; ok {
			return fmt.Errorf("HD快取路徑重複")
		}
		expected[row.Path] = row.SHA
	}
	count := 0
	e = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, e := filepath.Rel(root, p)
		if e != nil {
			return e
		}
		name := filepath.ToSlash(rel)
		want, ok := expected[name]
		if !ok {
			return fmt.Errorf("HD快取多出檔案：%s", name)
		}
		data, e := read(root, name)
		if e != nil {
			return e
		}
		if Sum(data) != want {
			return fmt.Errorf("HD快取SHA不符：%s", name)
		}
		count++
		return nil
	})
	if e != nil {
		return e
	}
	if count != len(expected) {
		return fmt.Errorf("HD快取缺少檔案")
	}
	return nil
}
