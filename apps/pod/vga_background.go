package pod

// 主專案規格010 b40、docs/re/017-b40-source-proof.json。
// 已證實完整88×88來源與使用RGB；正常位置及遮擋未知，只接既有完整矩形辨識。
import (
	"crypto/sha256"
	"fmt"
	"strings"
)

var exactBackgroundSources = map[string]localizedSource{
	"back1-001":   {"6e6497b68401ffa64001e15770ec2ee123e49f19eb60110e3c2d6461fb3804be", "adb917b88a493112d133bac4e314539572d6bfba778ddb2c3a53e22b6d5a6baf"},
	"back1-002":   {"07ff7762725474441a1173a794f2481353aa61af37f715eda25b81603a080c66", "8720264943a68137da0cab3bbfd0f3df75fcd67c2cab0d8edc17acaf0282f25d"},
	"back1-003":   {"969788ff0fa3a51c8847608d42832b9ac630b4ecb510854de5a119ccbcefec93", "a30e3ab18fe9affacf9166e642a37044cfe65efc4fbd05e65a24bbf911de143f"},
	"back1-004":   {"838d3c29e5d63ca0939efa6aeb38716b16b599d2c1eafcf79977f11aa1db3941", "4ce74f66db8bb002b55a83de400a790792fe9426722376c25af8e05320d0d410"},
	"back1-005":   {"bdb163a5f0fe0f042baad21fa36e145363789276a8ac15c57f4bf3b1f7c20b9a", "2780d5312de4ba2339660ce6883da810e38f09220d13132fc954374d9a726cc4"},
	"back1-006":   {"1ad27eb247fa05ac10158fe575f803c47d98ac9c147431d8fb1ae60b3212da1f", "01ade3d90c3382017dace8260bc653faa3013efb196cbdc7bbd2ab115295fb2b"},
	"skygrnd-002": {"3b7758d528493c836059053a9f64243c9907c75743d77028f9cc5d9afcdd79cb", "67b2cabf18e5f3c0e5bb29a1020957adc1bc36b121031f885a3ef63d9468cc9e"},
	"skygrnd-003": {"3b7758d528493c836059053a9f64243c9907c75743d77028f9cc5d9afcdd79cb", "9f14f27a6a1645c4fb2041550d6518641552eb7f8a32ac4c4fffca1d1d16e2ef"},
	"skygrnd-004": {"0c036b72e3c6ff1ced410c5171df87ac9abc4aaceb7cb891b500594be56176fa", "231d9bcc0266fcc4beda26b6b807a9e9f80d210c52ba8b9729f3dd5eee43d6f8"},
}

func exactBackgroundPalette(key string, ref []byte, size [2]int, algorithm string) (string, error) {
	_, known := exactBackgroundSources[key]
	vga := known || strings.HasPrefix(key, "back1-") || strings.HasPrefix(key, "skygrnd-")
	for _, v := range ref {
		vga = vga || v > 15
	}
	if !vga {
		return "", nil
	}
	source, ok := exactBackgroundSources[key]
	if !ok || size != [2]int{88, 88} || algorithm != "" || fmt.Sprintf("%x", sha256.Sum256(ref)) != source.indexed {
		return "", fmt.Errorf("%s：完整VGA背景來源或尺寸不符", key)
	}
	for _, v := range ref {
		if v < 144 || v >= 176 {
			return "", fmt.Errorf("%s：完整VGA背景色號不符", key)
		}
	}
	return source.palette, nil
}
