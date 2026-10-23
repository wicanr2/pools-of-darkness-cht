#!/usr/bin/env bash
# 由正式譯文（text/*.tsv）與工作區譯文批次重建開發用圓體字型（AGENTS.md 第 1 節：源泉圓體，2 倍 Medium、3 倍 Bold）。
# 改寫自銀色匕首之謎 tools/font-dev.sh；切換列字級（規格 020）與多語系字型等對應的表建立後再加。
# 輸出 workplace/font/pod-rounded.golemfnt（16×16）與 pod-rounded-3x.golemfnt（24×24）。
# 字型來源：workplace/fonts-src/rounded（GenSenRounded2 TW v2.100，OFL）、workplace/fonts-src/unifont（GNU Unifont 17.0.05，補字），
# 由銀色匕首之謎 workplace/fonts-src 複製，雜湊記在 WORKLOG。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC="$ROOT/workplace/fonts-src"
UNI="$SRC/unifont/unifont-17.0.05/font/precompiled"
for f in "$SRC/rounded/GenSenRounded2TW-M.otf" "$SRC/rounded/GenSenRounded2TW-B.otf" "$UNI/unifont_t-17.0.05.hex"; do
  test -f "$f" || { echo "缺字型來源 $f" >&2; exit 1; }
done
mkdir -p "$ROOT/workplace/font"
# 繁中字集只收玩家看得到的表（明確清單；新表要顯示中文時加在這裡）
cats=()
for f in journal.zh-TW; do
  test -f "$ROOT/text/$f.tsv" || { echo "缺 text/$f.tsv" >&2; exit 1; }
  cats+=("/text/$f.tsv")
done
for f in ecl-text.zh-TW ui-text.zh-TW name.zh-TW composed-templates host-text.zh-TW menu-masters translit-chars.zh-TW; do
  test -f "$ROOT/text/$f.tsv" && cats+=("/text/$f.tsv")
done
# 工作區的譯文批次與覆繪驗證用的試譯表（workplace/text-src/trial，含前作佔位）
while IFS= read -r f; do cats+=("/w/${f#"$ROOT/workplace/"}"); done < <(find "$ROOT/workplace/tr" "$ROOT/workplace/text-src/trial" \
  -name '*.tsv' \( -path '*/done/*' -o -path '*/trial/*' \) 2>/dev/null | sort)
timeout 900 docker run --rm --network none -u "$(id -u):$(id -g)" --memory 2g --cpus 2 --pids-limit 32 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -v "$ROOT/tools:/tools:ro" -v "$ROOT/text:/text:ro" -v "$ROOT/workplace:/w" -v "$SRC/rounded:/f:ro" -v "$UNI:/uni:ro" \
  sosb-img:r1 sh -c "
  python -I /tools/rounded_font.py ${cats[*]} --otf /f/GenSenRounded2TW-M.otf --cell 16 --fallback /uni/unifont_t-17.0.05.hex --out /w/font/pod-rounded.golemfnt &&
  python -I /tools/rounded_font.py ${cats[*]} --otf /f/GenSenRounded2TW-B.otf --cell 24 --fallback /uni/unifont_t-17.0.05.hex --out /w/font/pod-rounded-3x.golemfnt"
ls -la "$ROOT/workplace/font"/pod-*.golemfnt
