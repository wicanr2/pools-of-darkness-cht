"""G1 原始輸入盤點：原版 zip、珍157 rar、攻略掃描的雜湊、成員清單與解壓。

用法（Docker sosb-tools 內，原始輸入唯讀掛在 /in，workplace 掛在 /work）：
  python3 -I tools/inventory.py /in /work/input
原版 zip 是已展開的遊戲目錄（不是安裝器），頂層目錄 pools-of-darkness/ 展開到 <輸出根>/dos/pod/。
珍157 rar 以 unar -e cp950 展開到 <輸出根>/manual-zh/。攻略掃描只計雜湊，不複製。
"""
import hashlib, json, shutil, subprocess, sys, zipfile
from pathlib import Path

src, out = Path(sys.argv[1]), Path(sys.argv[2])
ZIP, RAR, WALK = src / "pools-of-darkness.zip", src / "珍157-黑暗之池.rar", src / "walkthrough"

def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def tree(root):
    return [{"path": str(p.relative_to(root)), "size": p.stat().st_size, "sha256": sha(p)}
            for p in sorted(root.rglob("*")) if p.is_file()]

report = {}

# 原版遊戲：zipfile 可直接讀（deflate），成員日期與 CRC 一併記錄。
game = out / "dos" / "pod"
if game.exists():
    raise SystemExit(f"{game} 已存在；確認後手動移除再重跑")
with zipfile.ZipFile(ZIP) as z:
    members = [{"name": i.filename, "size": i.file_size, "crc": f"{i.CRC:08x}",
                "date": "%04d-%02d-%02d %02d:%02d" % i.date_time[:5]} for i in z.infolist()]
    stage = out / "dos" / ".stage"
    z.extractall(stage)
shutil.move(stage / "pools-of-darkness", game)
stage.rmdir()
report["game"] = {"archive": ZIP.name, "sha256": sha(ZIP), "members": members, "files": tree(game)}

# 珍157：RAR，檔名 CP950。
manual = out / "manual-zh"
if manual.exists():
    raise SystemExit(f"{manual} 已存在；確認後手動移除再重跑")
subprocess.run(["unar", "-q", "-e", "cp950", "-o", str(manual), str(RAR)], check=True)
report["manual"] = {"archive": RAR.name, "sha256": sha(RAR), "files": tree(manual)}

# 攻略掃描：只記雜湊。
report["walkthrough"] = {"files": tree(WALK)}

json.dump(report, open(out / "inventory.json", "w"), ensure_ascii=False, indent=1)
for k in ("game", "manual"):
    print(report[k]["archive"], report[k]["sha256"], len(report[k]["files"]), "files")
print("walkthrough", len(report["walkthrough"]["files"]), "files")
