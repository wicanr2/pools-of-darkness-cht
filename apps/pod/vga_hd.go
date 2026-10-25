package pod

// 主專案規格010／研究017：僅核對原始DAX提供的色盤區段，不核對其他圖層的保留色。
// 常數為原始資料區段的SHA-256，不包含原版色盤或像素資料。
import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

const cpicVGAPalette = "f565f6507750cb76102763cde24ce198cfcbbc37e1b08edcc7dc121f03299151"

var terrainVGAPalettes = map[string]string{
	"dungcom-001": "d10ff3c9b87208a6d05cede1114360fa8d1dbf7c8eb3eb7cd200c701079c9a5f",
	"dungcom-002": "43ba359d37b0bb9be6b7f22c9d2034ec522df97f104a3bb53e775aefa2c906c3",
	"dungcom-003": "a8f9ba1517e6201f9a01b8dc39c3c7806423b059e0ae39b618e30f9471346a20",
	"dungcom-004": "1e4309c18ac862d4b78bf1c1f04c4236b87b1dc2fa28f800e10ab7c3fcbf6830",
	"dungcom-005": "072944b66eed642a372f2e70e094649cad7b53b7574b0cad471dc7e26164dd63",
	"dungcom-006": "f9437fb038ecd486af5fa4fd22f9e11253c3128fecbbaaa52c6111d3d6405be6",
	"wildcom-000": "d5bdbb077fba30cd4ab4ac314cf316d08bcdd4a00fb3b15030c426a894074998",
	"wildcom-001": "2fa59ed27afe7410987e5db7e653bfad517795058520d27e31ad5e88cc1082d4",
	"wildcom-002": "84e32cd2ac203bd725d48591de3aa501c86e593829fc5258293faadfeded556d",
}

func paletteBankHash(pal [256][3]uint8, first, end int) string {
	var data [768]byte
	for i := first; i < end; i++ {
		copy(data[(i-first)*3:], pal[i][:])
	}
	sum := sha256.Sum256(data[:(end-first)*3])
	return hex.EncodeToString(sum[:])
}

func hasVGAPixels(ref []byte) bool {
	for _, c := range ref {
		if c > 15 {
			return true
		}
	}
	return false
}

// 主專案研究017：CPIC1.DAX的244筆實際目錄；ID不連續，不能以筆數代替最大ID。
func cpicVGAIDExists(n int) bool {
	if n >= 128 {
		n -= 128
	}
	return n >= 1 && n <= 108 || n >= 112 && n <= 122 || n >= 125 && n <= 127
}

func vgaActorReference(key string, ref []byte) (bool, error) {
	if !hasVGAPixels(ref) {
		return false, nil
	}
	n, err := strconv.Atoi(strings.TrimPrefix(key, "cpic1-"))
	if err != nil || key != fmt.Sprintf("cpic1-%03d", n) || !cpicVGAIDExists(n) {
		return false, fmt.Errorf("%s：未知VGA怪物來源", key)
	}
	for _, c := range ref {
		if c > 15 && c < 144 {
			return false, fmt.Errorf("%s：色號%d不在CPIC色盤範圍", key, c)
		}
	}
	return true, nil
}

func vgaTerrainReference(key string, ref []byte) (string, error) {
	if !hasVGAPixels(ref) {
		return "", nil
	}
	base, frame, ok := strings.Cut(key, "@")
	n, err := strconv.Atoi(frame)
	want := terrainVGAPalettes[base]
	limit := 26
	if strings.HasPrefix(base, "wildcom-") {
		limit = 40
	}
	if !ok || err != nil || n < 0 || n >= limit || strconv.Itoa(n) != frame || want == "" {
		return "", fmt.Errorf("%s：未知VGA地形來源", key)
	}
	for _, c := range ref {
		if c > 15 && (c < 32 || c >= 64) {
			return "", fmt.Errorf("%s：色號%d不在地形色盤範圍", key, c)
		}
	}
	return want, nil
}

type vgaTerrainSignature struct {
	pixels  [32]byte
	palette string
}
