#!/usr/bin/env bash
# 繪圖收尾：乾淨建置、正式回歸及既有正常路線。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG="${1:?指定新的收據名稱}"
case "$TAG" in *[!A-Za-z0-9_-]*|'') exit 2;; esac
for d in workplace/dosgolem workplace/dosgolem/workplace workplace/dosgolem/workplace/bin workplace/probe workplace/hd/theme tools; do test -d "$ROOT/$d"; done
test ! -e "$ROOT/workplace/probe/g9-hd/$TAG"
test ! -e "$ROOT/workplace/probe/g9-hd/$TAG-plain"
test ! -e "$ROOT/workplace/probe/g9-hd/$TAG-save"
timeout 1500s docker run --rm --init --network none --user "$(id -u):$(id -g)" --memory 2g --cpus 2 --pids-limit 128 \
 --log-opt max-size=10m --log-opt max-file=3 --label pod.task=draw-closure-build \
 -v "$ROOT:/project:ro" -v "$ROOT/workplace/dosgolem/workplace:/project/workplace/dosgolem/workplace:rw" \
 -e POD_ROOT=/project -e POD_DRAW_CALLS=/project/workplace/probe/draw-intercept-v1/native-fade-v1/trace-original-entries \
 -e GOCACHE=/project/workplace/dosgolem/workplace/gocache -e GOMODCACHE=/project/workplace/dosgolem/workplace/gomodcache -e GOMAXPROCS=2 \
 -w /project/workplace/dosgolem dosgolem-go-ebiten:1.24-r1 bash -eu -c '
 test "$(stat -c %u:%g workplace/bin)" = "$(id -u):$(id -g)"
 test -z "$(git status --porcelain)"
 go test -count=1 -run "TestNativeFlood|TestNativePointReadBoundaryOriginal|TestFixedCharacterRecordFramebufferAlias|TestFixedFramebufferGuard" -v ./apps/pod/drawblocks > workplace/draw-closure-guard-targeted.log 2>&1
 go test -timeout 20m -count=1 ./apps/pod/drawblocks > workplace/draw-closure-regression.log 2>&1
 go test -count=1 ./internal/machine ./session ./apps/pod ./apps/pod/bgm ./xlate > workplace/draw-closure-related.log 2>&1
 xvfb-run -a go test -count=1 ./cmd/pod-play > workplace/draw-closure-frontend.log 2>&1
 go vet ./apps/pod/drawblocks ./apps/pod ./apps/pod/bgm ./xlate ./cmd/pod-play > workplace/draw-closure-vet.log 2>&1
 python3 -I /project/tools/draw_closure_probe.py --module /project/workplace/dosgolem --replace
 go build -o workplace/bin/pod-play-drawclosure ./workplace/draw-closure-audit-cmd
 go build -o workplace/bin/pod-play-drawfinal ./cmd/pod-play
 go version -m workplace/bin/pod-play-drawfinal > workplace/draw-closure-build.txt
 test -z "$(git status --porcelain)"
 '
for route in title movie wall; do
 bash "$ROOT/tools/direct-ui-receipt.sh" "$route" direct "$TAG" pod-play-drawclosure
done
POD_DIRECT_REPORT=1 POD_BIGPIC_SAMPLE=1 POD_RECEIPT_TIMEOUT=1800 \
 bash "$ROOT/tools/vga-battle-play-receipt.sh" hd "$TAG" workplace/hd/theme pod-play-drawclosure
bash "$ROOT/tools/whole-call-save-receipt.sh" native pod-play-drawclosure "$TAG-save"
bash "$ROOT/tools/direct-ui-receipt.sh" title direct "$TAG-plain" pod-play-drawfinal
timeout 180s docker run --rm --network none --user "$(id -u):$(id -g)" --memory 512m --cpus 1 --pids-limit 64 \
 --log-opt max-size=10m --log-opt max-file=3 --label pod.task=draw-closure-check \
 -v "$ROOT:/project:ro" -v "$ROOT/workplace/dosgolem/workplace:/project/workplace/dosgolem/workplace:rw" \
 -w /project dosgolem-go-ebiten:1.24-r1 bash -eu -c '
 python3 -I tools/draw_closure_check.py --tag "$1" --out "workplace/dosgolem/workplace/draw-closure-$1.json"
 find /project -user root -o -type d -name "*.md"
 ' _ "$TAG"
