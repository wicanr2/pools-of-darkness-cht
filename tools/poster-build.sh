#!/usr/bin/env bash
# 本機海報圖層建置；原版商店插畫不加入一般 patch。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
for d in review/hd/poster workplace/fonts-src/rounded workplace/hd/theme; do test -d "$ROOT/$d"; done
test -f "$ROOT/review/hd/poster/prototype.html"
OUT="$ROOT/workplace/hd/poster-build"
mkdir -p "$OUT" "$ROOT/workplace/hd/theme/posters"
timeout 90s docker run --rm --init --network none --memory 1g --cpus 2 --pids-limit 128 \
  --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 -e HOME=/tmp \
  -v "$ROOT:/project:ro" -v "$OUT:/out" mcr.microsoft.com/playwright:v1.56.1-noble \
  node /project/tools/poster_capture.mjs /out assets
timeout 45s docker run --rm --network none --memory 1g --cpus 2 --pids-limit 64 \
  --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 \
  -v "$ROOT/tools:/tools:ro" -v "$ROOT:/project:ro" -v "$OUT:/out" \
  -v "$ROOT/workplace/hd/theme/posters:/theme" sosb-img:r1 sh -eu -c '
  test "$(stat -c %u:%g /theme)" = "$(id -u):$(id -g)"
  python -I /tools/poster_layers.py --dir /out
  python -I - <<"PY"
import hashlib,json,shutil
from pathlib import Path
src,dst=Path("/out"),Path("/theme")
files=sorted(src.glob("poster-left-*.png"))+sorted(src.glob("poster-right-*.png"))+sorted(src.glob("poster-text-*.png"))
if len(files)!=12:raise ValueError("需要四張底圖與八張文字層")
for f in files:shutil.copyfile(f,dst/f.name)
meta=json.loads((src/"layers.json").read_text())
meta.update(private_only=True,original_art_sha256="7bfb0f23a9d09f5046c0b79b96421f4b1c841f8e97b63c0ec0c248ab7c73d9f0",
 files={f.name:hashlib.sha256(f.read_bytes()).hexdigest()for f in files},
 text_sha256={code:hashlib.sha256((Path("/project/text")/("poster."+code+".tsv")).read_bytes()).hexdigest()for code in ("zh-TW","en")})
(dst/"posters.manifest.json").write_text(json.dumps(meta,ensure_ascii=False,indent=2)+"\n")
print("本機海報主題已重建，12 PNG；原版插畫僅供本機與私人審閱。")
PY
  find /out /theme -user root -o -type d -name "*.md"
  '
