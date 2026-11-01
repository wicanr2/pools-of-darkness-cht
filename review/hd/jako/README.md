# 日文與韓文文字圖層

沿已核准的三張開場與B版海報，只重排日韓正式TSV文字。日文使用源泉圓體JP，韓文使用同家族韓文字模；背景及互動保持既定版本。譯文為機器翻譯並經語境校讀，尚未經母語者審校。

本目錄保存六張開場PNG、八張透明海報文字PNG及兩份主題JSON。六份原版TITLE辨識索引由自備DAX重生，不放入Git。

逐檔來源、SHA與字型入口在[正式素材清單](../runtime/manifest.json)。[規格015](../../../docs/spec/015-multilingual.md#日文與韓文收尾)記錄驗證契約，[SETUP](../../../SETUP.md)說明乾淨Git重建。生成入口為[jako_assets.py](../../../tools/jako_assets.py)與[固定瀏覽器擷取](../../../tools/jako_capture.mjs)，素材核對與登錄入口為[jako_runtime_manifest.py](../../../tools/jako_runtime_manifest.py)。
