package main

// 規格012的HD patch匯入分支。既有完整HD優先；B版來源由首次匯入重建。
import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/dosgolem/apps/pod/hdimport"
)

func prepareHDSource(raw []byte, bundle, game, data string) (string, error) {
	if err := hdimport.CheckOriginals(raw, game); err != nil {
		return "", err
	}
	root := filepath.Join(data, "hd-cache")
	if err := os.MkdirAll(root, 0755); err != nil {
		return "", err
	}
	dest := filepath.Join(root, hdimport.Sum(raw))
	if _, err := os.Stat(dest); err == nil {
		if err := hdimport.Verify(raw, dest); err != nil {
			return "", fmt.Errorf("HD快取核對失敗：%s：%w", dest, err)
		}
		return dest, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if _, err := hdimport.Build(raw, bundle, game, dest); err != nil {
		return "", fmt.Errorf("從自備原版重建HD辨識資料：%w", err)
	}
	return dest, nil
}

func releaseHDSource(game, data string) (string, error) {
	meta, err := resourcePath(filepath.Join("hd-source", "manifest.json"))
	if err != nil {
		return "", nil
	} // 基本包與既有完整HD沒有此來源。
	info, err := os.Stat(meta)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 8<<20 {
		return "", fmt.Errorf("HD來源清單不合法")
	}
	raw, err := os.ReadFile(meta)
	if err != nil {
		return "", err
	}
	return prepareHDSource(raw, filepath.Dir(meta), game, data)
}
