#!/usr/bin/env bash
# 只啟動隔離工作；片頭來源探針與雜湊皆在容器內執行。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG="${1:-intro-hd-source}"
SEED="${2:-$ROOT/workplace/probe/g2-receipt/scratch}"
case "$TAG" in *[!A-Za-z0-9_-]*|'') echo '收據名稱不合法' >&2; exit 2;; esac
for d in "$ROOT/workplace/dosgolem" "$ROOT/workplace/dosgolem/workplace/gomodcache" "$ROOT/workplace/input/dos/pod" "$SEED"; do test -d "$d"; done
for f in pool4.cfg snd.com; do test -f "$SEED/$f"; done
test -f "$ROOT/tools/movie_hd_probe.go"
git -C "$ROOT/workplace/dosgolem" diff --quiet
git -C "$ROOT/workplace/dosgolem" diff --cached --quiet
ENGINE_COMMIT="$(git -C "$ROOT/workplace/dosgolem" rev-parse HEAD)"
OUT="$ROOT/workplace/probe/$TAG"
test ! -e "$OUT" || { echo '保留舊收據，請另定名稱' >&2; exit 2; }
mkdir -p "$OUT"
timeout 900s docker run --rm --label pod.task=movie-hd-source --network none \
 --memory 2g --cpus 2 --pids-limit 128 --user "$(id -u):$(id -g)" \
 --log-opt max-size=10m --log-opt max-file=3 \
 -v "$ROOT/workplace/dosgolem:/engine:ro" -v "$ROOT/tools:/project/tools:ro" \
 -v "$ROOT/workplace/dosgolem/workplace/gomodcache:/gomodcache:ro" \
 -v "$ROOT/workplace/input/dos/pod:/orig:ro" -v "$SEED:/seed:ro" -v "$OUT:/out:rw" \
 -w /engine -e GOMAXPROCS=2 -e GOCACHE=/tmp/gocache -e GOMODCACHE=/gomodcache \
 -e GOPROXY=off -e GOTOOLCHAIN=local -e POD_ENGINE_COMMIT="$ENGINE_COMMIT" \
 dosgolem-go-ebiten:1.24-r1 sh -eu -c '
 python3 -I -c '\''import json;from pathlib import Path;Path("/out/overlay.json").write_text(json.dumps({"Replace":{"/engine/workplace/movieprobe/main.go":"/project/tools/movie_hd_probe.go"}}))'\''
 go build -p 2 -trimpath -buildvcs=false -overlay /out/overlay.json -o /out/movie-hd-probe ./workplace/movieprobe
 python3 -I -c '\''import json,hashlib,os;from pathlib import Path
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
record={"original":{str(p.relative_to("/orig")):sha(p) for p in sorted(Path("/orig").rglob("*")) if p.is_file()},"seed":{n:sha(Path("/seed")/n) for n in ("pool4.cfg","snd.com")},"probe_sha256":sha(Path("/out/movie-hd-probe")),"source_sha256":sha(Path("/project/tools/movie_hd_probe.go")),"engine_commit":os.environ["POD_ENGINE_COMMIT"]}
Path("/out/inputs.json").write_text(json.dumps(record,indent=2)+"\n")'\''
 /out/movie-hd-probe -orig /orig -seed /seed -out /out -frames 5100 > /out/run.log 2>&1
 tail -n 4 /out/run.log
 find /out -user root -o -type d -name "*.md"
 '
