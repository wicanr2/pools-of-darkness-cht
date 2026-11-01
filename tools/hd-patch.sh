#!/usr/bin/env bash
# #75：B版海報與首次辨識資料匯入；僅建立三平台HD patch候選。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE=${1:?build|payload}
DEST=${2:?本專案workplace內的新輸出目錄}
ORIGINAL="$ROOT/workplace/input/dos/pod"
ENGINE="$ROOT/workplace/dosgolem"
VERSION=v.0.1.1-20261101
MAIN=7ce48d7ebcec6189ca757cd153a71c5800ae2da5
read -r COMMIT < "$ROOT/tools/rebuild-engine.commit"
[[ "$COMMIT" = 661f7dc527d4d719047fcf8dac87f2e2cb3f0cbf ]]
case "$DEST" in "$ROOT"/workplace/*) ;; *) exit 2;; esac
case "$DEST/" in *'/../'*|*'/./'*) exit 2;; esac
for d in "$ENGINE" "$ORIGINAL" "$ENGINE/workplace/gomodcache";do test -d "$d";done
opts=(--rm --init --label pod.task=hd-patch --network none --memory 3g --cpus 2 --pids-limit 256
 --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 -e HOME=/tmp -e GOMAXPROCS=2)
if [[ "$MODE" = build ]];then
 test ! -e "$DEST"
 for img in dosgolem-go-ebiten:1.24-r1 sosb-img:r1 cok-osxcross:go1.24.13-sdk15.5-r1 kgb-cht-pkg:1.99.1-r2;do docker image inspect "$img" >/dev/null;done
 mkdir -p "$DEST"
 timeout 120 docker run "${opts[@]}" --mount "type=bind,source=$ROOT,target=/project,readonly" \
  --mount "type=bind,source=$ENGINE,target=/source,readonly" --mount "type=bind,source=$DEST,target=/out" \
  dosgolem-go-ebiten:1.24-r1 bash -euo pipefail -c '
  mkdir /out/main /out/engine /out/bin /out/test
  git -C /project rev-parse "$2" > /out/main-commit.txt
  git -C /project archive "$2" > /out/main-source.tar
  git -C /source archive "$1" > /out/engine-source.tar
  tar -xf /out/main-source.tar -C /out/main
  tar -xf /out/engine-source.tar -C /out/engine
  ' -- "$COMMIT" "$MAIN"
 timeout 300 docker run "${opts[@]}" --mount "type=bind,source=$DEST,target=/out" \
  --mount "type=bind,source=$ORIGINAL,target=/original,readonly" sosb-img:r1 \
  python3 -I /out/main/tools/rebuild_assets.py --root /out/main --original /original --out /out/assets > "$DEST/test/assets.log" 2>&1
 # 獨立平台只讀相同來源，各寫自己的bin子目錄；每個容器各2 CPU／3 GiB。
 pids=()
 for platform in linux windows macos;do
  image=dosgolem-go-ebiten:1.24-r1;[[ "$platform" != macos ]] || image=cok-osxcross:go1.24.13-sdk15.5-r1
  timeout 1200 docker run "${opts[@]}" --name "pod-hd-build-$platform" \
   --mount "type=bind,source=$DEST/engine,target=/engine,readonly" \
   --mount "type=bind,source=$DEST/main/tools,target=/tools,readonly" \
   --mount "type=bind,source=$ENGINE/workplace/gomodcache,target=/gomodcache,readonly" \
   --mount "type=bind,source=$DEST/bin,target=/out" -w /engine "$image" \
   sh /tools/package_build.sh "$platform" "$VERSION" "$COMMIT" > "$DEST/test/build-$platform.log" 2>&1 &
  pids+=("$!")
 done
 failure=0
 for pid in "${pids[@]}";do wait "$pid" || failure=1;done
 [[ "$failure" = 0 ]]
elif [[ "$MODE" != payload ]];then exit 2
fi
timeout 300 docker run "${opts[@]}" --mount "type=bind,source=$DEST,target=/out" \
 --mount "type=bind,source=$ORIGINAL,target=/original,readonly" \
 --mount "type=bind,source=$ENGINE/workplace/gomodcache,target=/gomodcache,readonly" sosb-img:r1 \
 python3 -I /out/main/tools/package_payload.py --root /out/main --engine /out/engine --original /original \
 --assets /out/assets --build /out/bin --stage /out/stage --modcache /gomodcache --version "$VERSION" --commit "$COMMIT" --hd-patch-only
timeout 300 docker run "${opts[@]}" --mount "type=bind,source=$DEST,target=/out" kgb-cht-pkg:1.99.1-r2 \
 python3 -I /out/main/tools/package_appimage.py --stage /out/stage --runtime /out/main/tools/vendor/appimage/runtime-x86_64 --out /out/packages --logs /out/test
docker ps -a --filter label=pod.task=hd-patch --format '{{.Names}} {{.Status}}'
