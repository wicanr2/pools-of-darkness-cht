package main

// 規格012的HD patch匯入分支。海報定案前不從正式啟動流程呼叫。
import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/dosgolem/apps/pod/hdimport"
)

func prepareHDSource(raw []byte, bundle, game, data string) (string, error) {
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
