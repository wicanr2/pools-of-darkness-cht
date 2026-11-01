# 字型重建來源

由[跨機重建入口](../SETUP.md)產生正式字型。來源檔與固定字集已存Git，不需前作或舊`workplace/fonts-src/`。

| 來源 | 版本／授權 | 用途 |
|---|---|---|
| [源泉圓體](https://github.com/ButTaiwan/gensen-font) | GenSenRounded2 TW／JP v2.100，[OFL 1.1](source/rounded/SIL_Open_Font_License_1.1.txt) | Medium：12／16像素；Bold：22／24像素 |
| [Courier Prime](https://github.com/google/fonts/tree/main/ofl/courierprime) | Bold固定檔案SHA，[OFL 1.1](source/courier/OFL.txt) | 英文16／24像素 |
| [GNU Unifont](https://unifoundry.com/unifont/index.html) | 17.0.05台灣與日本字源，[OFL 1.1](source/unifont/OFL-1.1.txt)／[完整授權](source/unifont/COPYING) | 源泉圓體缺字補字來源 |

所有現行來源完整SHA與十八份成品SHA固定在[重建清單](../review/hd/runtime/manifest.json)。字型依原授權提供，專案LICENSE不取代字型授權。

日文使用同版本[源泉圓體JP來源](source/rounded/JP-manifest.json)的Medium／Bold，韓文使用TW字源內的原生韓文字模。補字分別使用[Unifont JP 17.0.05](source/unifont/unifont_jp-17.0.05.hex)與[Unifont TW 17.0.05](source/unifont/unifont_t-17.0.05.hex)，來源SHA由重建清單核對。

日韓各有16／24像素遊戲字模及12／22像素工具列字模，共新增八份，原十份正式字型SHA不變。[字模工具](../tools/jako_font.py)由正式日韓顯示欄產生字集，固定碼點與補字範圍如下。

| 語系 | 遊戲字集 | 工具列字集 | Unifont補字 |
|---|---|---|---|
| 日文 | 2,022字，[16像素](charset-pod-ja.json)／[24像素](charset-pod-ja-3x.json) | 252字，[12像素](charset-pod-ja-bar12.json)／[22像素](charset-pod-ja-bar22.json) | 無 |
| 韓文 | 1,182字，[16像素](charset-pod-ko.json)／[24像素](charset-pod-ko-3x.json) | 240字，[12像素](charset-pod-ko-bar12.json)／[22像素](charset-pod-ko-bar22.json) | 遊戲字模各1字，工具列無 |

十八份字型已由乾淨Git來源重建並逐檔核對SHA，日韓正式譯文的字型覆蓋與手札／攻略全文分頁檢查通過，範圍見[034收據](../docs/re/034-jako-runtime.json)。日韓譯文為機器翻譯，經語境校讀，未經母語者審校。

`charset-pod-rounded*.json`保存已驗字型的唯一遞增碼點；[rounded_font.py](../tools/rounded_font.py)的`--charset-file`依這份字集重生，避免被本機翻譯草稿改變。
繁中4份圓體字型皆不需Unifont補字，仍保留同版補字來源與授權，供後續正式字集使用。

簡中另有16／24及工具列12／22四份源泉圓體，2,783個遊戲字元與331個工具列字元，均無Unifont補字。[字集來源](../tools/zhcn_font.py)只讀正式簡中顯示欄，固定碼點保存於[遊戲字集](charset-pod-zh-CN.json)與[工具列字集](charset-pod-zh-CN-bar.json)。繁中與英文六份原字型保持原SHA。
