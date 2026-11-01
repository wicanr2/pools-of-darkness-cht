#!/usr/bin/env bash
# HD patch檔案重建核心；不接正式入口，不產生候選包，不改海報。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG=${1:-hd-import-core-replay}
ENGINE="$ROOT/workplace/dosgolem"
ORIGINAL="$ROOT/workplace/input/dos/pod"
COMMIT=f09cf683c06098ed3d48703afc1ecac489df44f6
MAIN=461b274523e15b84b2251927b16ecc5a68ef195c
case "$TAG" in *[!A-Za-z0-9_-]*|'') exit 2;; esac
OUT="$ROOT/workplace/$TAG"
for d in "$ENGINE" "$ORIGINAL" "$ENGINE/workplace/gomodcache";do test -d "$d";done
test ! -e "$OUT"
mkdir "$OUT"
opts=(--rm --init --label pod.task=hd-import --network none --memory 3g --cpus 2 --pids-limit 256
 --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 -e HOME=/tmp -e GOMAXPROCS=2)
timeout 600 docker run "${opts[@]}" --name "pod-hd-import-$TAG" \
 --mount "type=bind,source=$ROOT,target=/project,readonly" \
 --mount "type=bind,source=$ENGINE,target=/source,readonly" \
 --mount "type=bind,source=$ORIGINAL,target=/original,readonly" \
 --mount "type=bind,source=$ENGINE/workplace/gomodcache,target=/gomodcache,readonly" \
 --mount "type=bind,source=$OUT,target=/out" -e GOCACHE=/tmp/gocache -e GOMODCACHE=/gomodcache \
 -e GOPROXY=off -e GOTOOLCHAIN=local dosgolem-go-ebiten:1.24-r1 bash -euo pipefail -c '
 test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
 test "$(go env GOVERSION)" = go1.24.13
 mkdir /out/engine /out/main /out/test
 git -C /project rev-parse "$2" > /out/main-commit.txt
 git -C /project archive "$2" | tar -x -C /out/main
 git -C /source archive "$1" | tar -x -C /out/engine
 cd /out/engine
 POD_HD_ROOT=/out/main POD_HD_ORIGINAL=/original POD_HD_OUT=/out/test \
  go test -p 2 ./apps/pod/hdimport -v > /out/core-test.log 2>&1
 POD_HD_ROOT=/out/main POD_HD_ORIGINAL=/original POD_HD_OUT=/out/test \
  xvfb-run -a go test -p 2 ./cmd/pod-play -run TestHDSourceCache -v > /out/cache-test.log 2>&1
 xvfb-run -a go test -p 2 ./apps/pod/hdimport ./cmd/pod-play > /out/regression.log 2>&1
 GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -p 2 -c -o /out/import-windows.test.exe ./apps/pod/hdimport
 GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go test -p 2 -c -o /out/import-darwin-amd64.test ./apps/pod/hdimport
 GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go test -p 2 -c -o /out/import-darwin-arm64.test ./apps/pod/hdimport
 tail -3 /out/core-test.log
 tail -4 /out/cache-test.log
 cat /out/regression.log
 find /out -user root -o -type d -name "*.md"
 ' -- "$COMMIT" "$MAIN"
docker ps -a --filter label=pod.task=hd-import --format '{{.Names}} {{.Status}}'
