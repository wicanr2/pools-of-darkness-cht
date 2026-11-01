#!/usr/bin/env bash
# 主機只編排Docker與Git；解碼、建置、原版與Go重播均在容器內。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG=${1:-intro-bma-source-v1}
SEED=${2:-$ROOT/workplace/probe/g2-receipt/scratch}
case "$TAG" in *[!A-Za-z0-9_-]*|'') echo '收據名稱不合法' >&2;exit 2;; esac
ENGINE="$ROOT/workplace/dosgolem"
ORIGINAL="$ROOT/workplace/input/dos/pod"
MODULES="$ENGINE/workplace/gomodcache"
OUT="$ROOT/workplace/probe/$TAG"
for path in "$ENGINE" "$ORIGINAL" "$MODULES" "$SEED";do test -d "$path";done
for name in pool4.cfg snd.com;do test -f "$SEED/$name";done
test ! -e "$OUT"
git -C "$ENGINE" diff --quiet
git -C "$ENGINE" diff --cached --quiet
COMMIT=$(git -C "$ENGINE" rev-parse HEAD)
GO=dosgolem-go-ebiten:1.24-r1
IMG=sosb-img:r1
docker image inspect "$GO" >/dev/null
docker image inspect "$IMG" >/dev/null
mkdir "$OUT"
opts=(--rm --init --label pod.task=intro-bma --network none --memory 2g --cpus 2 --pids-limit 128
 --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 -e HOME=/tmp -e GOMAXPROCS=2)
timeout 300s docker run "${opts[@]}" -v "$ENGINE:/engine:ro" -v "$ROOT/tools:/tools:ro" \
 -v "$MODULES:/gomodcache:ro" -v "$OUT:/out:rw" -w /engine \
 -e GOCACHE=/tmp/gocache -e GOMODCACHE=/gomodcache -e GOPROXY=off -e GOTOOLCHAIN=local \
 "$GO" bash -euo pipefail -c '
 test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
 python3 -I -c "import json;from pathlib import Path;Path(\"/out/overlay.json\").write_text(json.dumps({\"Replace\":{\"/engine/workplace/movieprobe/main.go\":\"/tools/movie_bma_probe.go\"}}))"
 go build -p 2 -trimpath -buildvcs=false -overlay /out/overlay.json -o /out/movie-bma-probe ./workplace/movieprobe
 go version -m /out/movie-bma-probe > /out/build-info.txt
 '
for mode in raw go;do
 timeout 900s docker run "${opts[@]}" -v "$ORIGINAL:/original:ro" -v "$SEED:/seed:ro" -v "$OUT:/out:rw" \
 "$GO" bash -euo pipefail -c '
 args=();if [[ $1 == go ]];then args=(-go-draw);fi
 /out/movie-bma-probe "${args[@]}" -orig /original -seed /seed -out "/out/$1" -frames 5100 > "/out/$1.log" 2>&1
 tail -n 3 "/out/$1.log"
 ' -- "$mode"
done
timeout 120s docker run "${opts[@]}" -v "$ROOT:/project:ro" -v "$SEED:/seed:ro" -v "$OUT:/out:rw" "$IMG" \
 bash -euo pipefail -c '
 python3 -I /project/tools/intro_bma_test.py -v > /out/codec.log 2>&1
 python3 -I /project/tools/intro_bma_inventory.py --root /project --seed /seed --engine "$1" \
  --raw /out/raw --fast /out/go --probe /out/movie-bma-probe --out /out/requests --proof /out/source-proof.json
 find /out -user root -o -type d -name "*.md"
 ' -- "$COMMIT"
docker ps -a --filter label=pod.task=intro-bma --format '{{.Names}} {{.Status}}'
