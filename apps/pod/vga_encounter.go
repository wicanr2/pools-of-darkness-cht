package pod

// 主專案規格010 b41、docs/re/017-b41-source-proof.json。
// 已證實SPRIT041近景在正常DEMO的[24,24,88,88]；中遠景僅為來源候選。
import (
	"crypto/sha256"
	"fmt"
	"strings"
)

var encounterVGASources = map[string]localizedSource{
	"sprit1-041@0": {"e3a835776c06b3fbc49aecd405af7e03ac2b43527b28ee377dce1152dbba9607", "9061c36f48b0d095009e3cb060df0a767c2cdd5b898167e43ccf98d05c9a9d70"},
	"sprit1-041@1": {"d30de9a7a7f70b2880687daab1ea2d4d20f74c1af797768b1cb3adde871e6f7b", "573485ad3d86e8bcc281b8cbeb8c2cdced091accf5921ccacf17d78ff4f8034b"},
	"sprit1-041@2": {"a5e4da506c2eb631448b12197b56f12b24aef4617a6cb4fd17c25209acc401d1", "c73595adf793f677f2f22c29e302d68c1d7c330bf77db0197001cf4fc7ae4312"},
}

func encounterVGAPalette(key string, ref []byte, size, position [2]int) (string, error) {
	vga := strings.HasPrefix(key, "sprit1-")
	for _, v := range ref {
		vga = vga || v > 15
	}
	if !vga {
		return "", nil
	}
	source, ok := encounterVGASources[key]
	if !ok || size != [2]int{88, 88} || position != [2]int{24, 24} || fmt.Sprintf("%x", sha256.Sum256(ref)) != source.indexed {
		return "", fmt.Errorf("%s：VGA遭遇圖來源、尺寸或位置不符", key)
	}
	for _, v := range ref {
		if v != 0 && v < 176 {
			return "", fmt.Errorf("%s：VGA遭遇圖色號不符", key)
		}
	}
	return source.palette, nil
}
