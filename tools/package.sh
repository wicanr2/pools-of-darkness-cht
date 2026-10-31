#!/usr/bin/env bash
# 主機只編排Docker與Git；候選固定於規格012，不建立tag或Release。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE=${1:?build|smoke|verify}
ORIGINAL=${2:?合法DOS原版目錄}
DEST=${3:?本專案workplace內的新輸出目錄}
VERSION=v.0.1.0-20261101
COMMIT=4a8ef9f4ff18c82d8c98595faf19fb1ee803c609
GO_IMAGE=dosgolem-go-ebiten:1.24-r1
IMG=sosb-img:r1
MAC=cok-osxcross:go1.24.13-sdk15.5-r1
PKG=kgb-cht-pkg:1.99.1-r2
WINE=yuan-wine-verify:noble-r2
case "$MODE" in build|smoke|verify) ;; *) exit 2;; esac
test -f "$ORIGINAL/GAME.EXE"
ORIGINAL="$(cd "$ORIGINAL" && pwd)"
case "$DEST" in "$ROOT"/workplace/*) ;; *) echo '輸出須在本專案workplace內' >&2;exit 2;; esac
case "$DEST/" in *'/../'*|*'/./'*) exit 2;; esac
opts=(--rm --init --label pod.task=package --user "$(id -u):$(id -g)" --memory 3g --cpus 2 --pids-limit 256 --log-opt max-size=10m --log-opt max-file=3 -e HOME=/tmp -e GOMAXPROCS=2)
if [[ "$MODE" = build ]]; then
 test ! -e "$DEST"
 for image in "$GO_IMAGE" "$IMG" "$MAC" "$PKG"; do docker image inspect "$image" >/dev/null;done
 git -C "$ROOT" cat-file -e "$COMMIT^{commit}" 2>/dev/null || git -C "$ROOT" fetch --no-tags origin dosgolem/pod
 mkdir -p "$DEST"
 # 候選既有九包的來源固定；後續語系／HD更新不能悄悄改變此版內容。
 git -C "$ROOT" archive 4bdd711 > "$DEST/main-source.tar"
 git -C "$ROOT" rev-parse '4bdd711^{tree}' > "$DEST/main-tree.txt"
 git -C "$ROOT" archive "$COMMIT" > "$DEST/engine-source.tar"
 timeout 60 docker run "${opts[@]}" --network none -v "$DEST:/out" "$GO_IMAGE" sh -ec '
  mkdir /out/main /out/engine /out/test /out/bin /out/gomodcache
  tar -xf /out/main-source.tar -C /out/main
  tar -xf /out/engine-source.tar -C /out/engine
 '
 timeout 300 docker run "${opts[@]}" --network none -v "$DEST:/out" -v "$ORIGINAL:/original:ro" "$IMG" \
  python3 -I /out/main/tools/rebuild_assets.py --root /out/main --original /original --out /out/assets
 timeout 300 docker run "${opts[@]}" --network bridge -v "$DEST/engine:/engine:ro" -v "$DEST/gomodcache:/gomodcache" \
  -w /engine -e GOMODCACHE=/gomodcache "$GO_IMAGE" go mod download
 for platform in linux windows macos; do
  image=$GO_IMAGE; [[ "$platform" != macos ]] || image=$MAC
  timeout 1200 docker run "${opts[@]}" --network none -v "$DEST/engine:/engine:ro" -v "$DEST/main/tools:/tools:ro" \
   -v "$DEST/gomodcache:/gomodcache:ro" -v "$DEST/bin:/out" -w /engine "$image" \
   sh /tools/package_build.sh "$platform" "$VERSION" "$COMMIT" > "$DEST/test/build-$platform.log" 2>&1
 done
 timeout 300 docker run "${opts[@]}" --network none -v "$DEST:/out" -v "$ORIGINAL:/original:ro" "$IMG" \
  python3 -I /out/main/tools/package_payload.py --root /out/main --engine /out/engine --original /original \
  --assets /out/assets --build /out/bin --stage /out/stage --modcache /out/gomodcache --version "$VERSION" --commit "$COMMIT"
 timeout 300 docker run "${opts[@]}" --network none -v "$DEST:/out" "$PKG" \
  python3 -I /out/main/tools/package_appimage.py --stage /out/stage --runtime /out/main/tools/vendor/appimage/runtime-x86_64 --out /out/packages --logs /out/test
elif [[ "$MODE" = smoke ]]; then
 # 槽A必須由正常玩家路徑建立，不能用測試狀態注入取代。
 SEED=${4:?正常自建槽A的SAVE目錄}
 test -f "$SEED/savgamA.pty";test ! -e "$SEED/savgamB.pty"
 SEED="$(cd "$SEED" && pwd)"
 test -d "$DEST/packages";test ! -e "$DEST/smoke";test ! -e "$DEST/test/base-flash-packaged"
 docker image inspect "$WINE" >/dev/null
 mkdir "$DEST/smoke" "$DEST/test/base-flash-packaged"
 for mode in base hd save missing flash; do
  out="$DEST/smoke/linux-$mode"; [[ "$mode" != flash ]] || out="$DEST/test/base-flash-packaged"
  [[ "$mode" = flash ]] || mkdir "$out"
  timeout 1500 docker run "${opts[@]}" --network none -e LIBGL_ALWAYS_SOFTWARE=1 -v "$ROOT/tools:/tools:ro" \
   -v "$DEST/packages:/packages:ro" -v "$SEED:/seed:ro" -v "$out:/out" "$GO_IMAGE" \
   sh /tools/package_smoke_linux.sh "$VERSION" "$mode"
 done
 mkdir "$DEST/smoke/windows"
 timeout 1200 docker run "${opts[@]}" --network none -v "$ROOT/tools:/tools:ro" -v "$DEST/packages:/packages:ro" \
  -v "$DEST/smoke/windows:/out" "$WINE" python3 -I /tools/package_smoke_windows.py --version "$VERSION"
else
 test -f "$DEST/packages/packages.json"
 timeout 600 docker run "${opts[@]}" --network none -v "$ROOT:/project:ro" -v "$DEST:/out" -v "$ORIGINAL:/original:ro" \
  "$IMG" python3 -I /project/tools/package_verify.py --root /project --work /out --original /original --receipt /out/package-runtime.json
fi
timeout 30 docker run "${opts[@]}" --network none -v "$DEST:/out:ro" "$IMG" find /out -user root -o -type d -name '*.md'
docker ps -a --filter label=pod.task=package --format '{{.Names}} {{.Status}}'
