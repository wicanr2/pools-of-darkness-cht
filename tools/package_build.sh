#!/bin/sh
# 容器內建置入口；外層須掛固定乾淨引擎、唯讀模組快取與獨立輸出。
set -eu
platform=${1:?linux|windows|macos}
version=${2:?完整候選版號}
commit=${3:?固定引擎commit}
export GOCACHE=${POD_BUILD_CACHE:-/tmp/gocache} GOMODCACHE=/gomodcache GOPROXY=off GOTOOLCHAIN=local GOMAXPROCS=2
test "$(go env GOVERSION)" = go1.24.13
flags="-s -w -X main.version=$version -X main.dosgolemCommit=$commit"
mkdir -p "/out/$platform"
case "$platform" in
 linux)
  go build -p 2 -trimpath -buildvcs=false -ldflags "$flags" -o /out/linux/pod-play ./cmd/pod-play
  go version -m /out/linux/pod-play > /out/linux/build-info.txt
  cp "$(go env GOROOT)/LICENSE" /out/linux/Go-LICENSE
  ;;
 windows)
  GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -p 2 -trimpath -buildvcs=false -ldflags "$flags -H windowsgui" -o /out/windows/PODCHT.exe ./cmd/pod-play
  go version -m /out/windows/PODCHT.exe > /out/windows/build-info.txt
  ;;
 macos)
  export GOOS=darwin CGO_ENABLED=1 MACOSX_DEPLOYMENT_TARGET=11.0
  export CGO_CFLAGS=-mmacosx-version-min=11.0 CGO_LDFLAGS=-mmacosx-version-min=11.0
  GOARCH=arm64 CC=oa64-clang CXX=oa64-clang++ go build -p 2 -trimpath -buildvcs=false -ldflags "$flags" -o /out/macos/pod-play-arm64 ./cmd/pod-play
  GOARCH=amd64 CC=o64-clang CXX=o64-clang++ go build -p 2 -trimpath -buildvcs=false -ldflags "$flags" -o /out/macos/pod-play-amd64 ./cmd/pod-play
  lipo -create /out/macos/pod-play-arm64 /out/macos/pod-play-amd64 -output /out/macos/pod-play
  lipo -archs /out/macos/pod-play > /out/macos/archs.txt
  x86_64-apple-darwin24.5-otool -L /out/macos/pod-play > /out/macos/libraries.txt
  go version -m /out/macos/pod-play-arm64 > /out/macos/build-info-arm64.txt
  go version -m /out/macos/pod-play-amd64 > /out/macos/build-info-amd64.txt
  ;;
 *) exit 2;;
esac
find "/out/$platform" -user root -o -type d -name '*.md'
