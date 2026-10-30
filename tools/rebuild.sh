#!/usr/bin/env bash
# 主機只作Docker/Git編排；所有解碼、字型、編譯與執行都在容器。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-assets}"
ORIGINAL="${2:-$ROOT/workplace/input/dos/pod}"
DEST="${3:-$ROOT/workplace/rebuild/current}"
IMAGE=dosgolem-go-ebiten:1.24-r1
IMG=sosb-img:r1
read -r COMMIT < "$ROOT/tools/rebuild-engine.commit"
case "$MODE" in assets|build|smoke) ;; *) echo '用法：tools/rebuild.sh assets|build|smoke [原版目錄] [重建目錄]' >&2; exit 2;; esac
test -d "$ORIGINAL" || { echo "缺自備原版目錄：$ORIGINAL" >&2; exit 1; }
ORIGINAL="$(cd "$ORIGINAL" && pwd)"
case "$DEST" in "$ROOT"/workplace/*) ;; *) echo '重建輸出須放本專案workplace內' >&2; exit 2;; esac
case "$DEST/" in *'/../'*|*'/./'*) echo '重建路徑不能含相對跳轉' >&2; exit 2;; esac
mkdir -p "$(dirname "$DEST")"
opts=(--rm --init --label pod.task=rebuild --user "$(id -u):$(id -g)" --memory 3g --cpus 2 --pids-limit 256 --log-opt max-size=10m --log-opt max-file=3 -e HOME=/tmp -e GOMAXPROCS=2)
if [[ "$MODE" = assets ]]; then
  test ! -e "$DEST" || { echo '輸出已存在；請選新路徑，不覆寫既有驗收' >&2; exit 1; }
  docker image inspect "$IMG" >/dev/null 2>&1 || timeout 900 docker build -t "$IMG" -f "$ROOT/tools/docker/img.Dockerfile" "$ROOT/tools/docker"
  timeout 300 docker run "${opts[@]}" --network none -v "$ROOT:/project:ro" -v "$ORIGINAL:/original:ro" \
    -v "$(dirname "$DEST"):/output:rw" "$IMG" python3 -I /project/tools/rebuild_assets.py --root /project --original /original --out "/output/$(basename "$DEST")"
elif [[ "$MODE" = build ]]; then
  test -f "$DEST/assets-receipt.json"
  test ! -e "$DEST/engine" && test ! -e "$DEST/bin"
  git -C "$ROOT" cat-file -e "$COMMIT^{commit}" 2>/dev/null || git -C "$ROOT" fetch --no-tags origin dosgolem/pod
  docker image inspect "$IMAGE" >/dev/null 2>&1 || timeout 1200 docker build -t "$IMAGE" -f "$ROOT/tools/docker/rebuild-go.Dockerfile" "$ROOT/tools/docker"
  timeout 120 docker run "${opts[@]}" --network none -v "$ROOT:/project:ro" -v "$DEST:/out:rw" "$IMAGE" bash -euo pipefail -c '
    python3 -I -c "import json,pathlib,sys; assert json.loads(pathlib.Path(\"/project/review/hd/runtime/manifest.json\").read_text())[\"engine\"][\"commit\"] == sys.argv[1], \"引擎commit與素材清單不合\"" "$1"
    mkdir /out/engine /out/bin /out/gomodcache
    git -C /project archive "$1" | tar -x -C /out/engine
  ' -- "$COMMIT"
  timeout 300 docker run "${opts[@]}" --network bridge -e GOMODCACHE=/out/gomodcache -v "$DEST:/out:rw" -w /out/engine "$IMAGE" go mod download
  timeout 600 docker run "${opts[@]}" --network none -e GOCACHE=/tmp/gocache -e GOMODCACHE=/out/gomodcache -e GOFLAGS=-mod=readonly \
    -v "$DEST:/out:rw" -w /out/engine "$IMAGE" bash -euo pipefail -c '
      test "$(go env GOVERSION)" = go1.24.13
      go build -trimpath -buildvcs=false -o /out/bin/pod-play ./cmd/pod-play
      go version -m /out/bin/pod-play > /out/bin/build-info.txt
      sha256sum /out/bin/pod-play > /out/bin/SHA256SUMS
      find /out -user root -o -type d -name "*.md"
    '
else
  for p in "$DEST/bin/pod-play" "$DEST/font/pod-rounded.golemfnt" "$DEST/theme/theme.json"; do test -f "$p"; done
  test ! -e "$DEST/smoke" || { echo 'smoke收據已存在；保留後再用新目錄驗證' >&2; exit 1; }
  mkdir "$DEST/smoke"
  timeout 900 docker run "${opts[@]}" --network none -e POD_NO_DIALOG=1 \
    -v "$DEST/bin:/app:ro" -v "$ORIGINAL:/orig:ro" -v "$ROOT/text:/text:ro" -v "$DEST/font:/font:ro" \
    -v "$DEST/theme:/theme:ro" -v "$DEST/smoke:/out:rw" "$IMAGE" bash -euo pipefail -c '
      xvfb-run -a /app/pod-play -original /orig -save /out/save -text-dir /text -font /font/pod-rounded.golemfnt \
        -scale 3 -audio=false -frames 7400 -script "900: ,1800:d" -hd /theme -posters /theme/posters -flash-audit \
        -shot-frame /out/frame.png -shot-original /out/original.rgba -shot-source /out/source > /out/run.log 2>&1
      tail -n 9 /out/run.log
      find /out -user root -o -type d -name "*.md"
    '
  timeout 60 docker run "${opts[@]}" --network none -v "$ROOT:/project:ro" -v "$DEST:/out:rw" \
    "$IMAGE" python3 -I /project/tools/rebuild_verify.py --root /project --out /out
fi
