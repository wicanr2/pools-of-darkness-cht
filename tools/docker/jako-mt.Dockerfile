# 本作日韓離線翻譯。base ID及四個wheel SHA由工具來源清單核對。
FROM dosgolem-go-ebiten:1.24-r1
COPY wheels/ /tmp/pod-mt-wheels/
COPY manifest.json /opt/pod-mt-lock.json
RUN python3 -I -c 'from pathlib import Path; import hashlib,json,zipfile; lock=json.loads(Path("/opt/pod-mt-lock.json").read_text()); out=Path("/opt/pod-mt"); out.mkdir(); [(lambda p,row: (None if hashlib.sha256(p.read_bytes()).hexdigest()==row["sha256"] else (_ for _ in ()).throw(ValueError("wheel SHA mismatch")),zipfile.ZipFile(p).extractall(out)))(Path("/tmp/pod-mt-wheels")/row["file"],row) for row in lock["files"]]'
ENV OMP_NUM_THREADS=2 OPENBLAS_NUM_THREADS=1
