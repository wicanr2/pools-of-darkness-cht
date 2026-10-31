# 字型重建來源

由[跨機重建入口](../SETUP.md)產生正式字型。來源檔與固定字集已存Git，不需前作或舊`workplace/fonts-src/`。

| 來源 | 版本／授權 | 用途 |
|---|---|---|
| [源泉圓體](https://github.com/ButTaiwan/gensen-font) | GenSenRounded2 TW v2.100，[OFL 1.1](source/rounded/SIL_Open_Font_License_1.1.txt) | Medium：12／16像素；Bold：22／24像素 |
| [Courier Prime](https://github.com/google/fonts/tree/main/ofl/courierprime) | Bold固定檔案SHA，[OFL 1.1](source/courier/OFL.txt) | 英文16／24像素 |
| [GNU Unifont](https://unifoundry.com/unifont/index.html) | 17.0.05台灣字源，[OFL 1.1](source/unifont/OFL-1.1.txt)／[完整授權](source/unifont/COPYING) | 源泉圓體缺字補字來源 |

所有來源完整SHA與十份成品SHA固定在[重建清單](../review/hd/runtime/manifest.json)。字型依原授權提供，專案LICENSE不取代字型授權。
`charset-pod-rounded*.json`保存已驗字型的唯一遞增碼點；[rounded_font.py](../tools/rounded_font.py)的`--charset-file`依這份字集重生，避免被本機翻譯草稿改變。
繁中4份圓體字型皆不需Unifont補字，仍保留同版補字來源與授權，供後續正式字集使用。

簡中另有16／24及工具列12／22四份源泉圓體，2,783個遊戲字元與331個工具列字元，均無Unifont補字。[字集來源](../tools/zhcn_font.py)只讀正式簡中顯示欄，固定碼點保存於[遊戲字集](charset-pod-zh-CN.json)與[工具列字集](charset-pod-zh-CN-bar.json)。繁中與英文六份原字型保持原SHA。
