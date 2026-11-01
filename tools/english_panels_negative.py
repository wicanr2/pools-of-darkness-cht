#!/usr/bin/env python3
"""固定小型資料負對照：更動寶物加值或攻略座標必須拒絕；只改隔離副本。"""
import argparse
import json
from pathlib import Path
import shutil
import sys

sys.path.insert(0,str(Path(__file__).resolve().parent))
from english_panels import check

def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('--root',type=Path,required=True);ap.add_argument('--out',type=Path,required=True);a=ap.parse_args();d=a.out/'negative';d.mkdir();(d/'text').mkdir();(d/'reference').mkdir()
    for n in ('journal.en.tsv','walkthrough.en.tsv','map-titles.en.tsv','map-guide.en.tsv','journal.zh-TW.tsv','walkthrough.zh-TW.tsv','map-geo.tsv','map-guide.zh-TW.tsv'):shutil.copyfile(a.root/'text'/n,d/'text'/n)
    shutil.copyfile(a.root/'reference/journal-en.tsv',d/'reference/journal-en.tsv');check(d)
    checks=[];p=d/'text/walkthrough.en.tsv';original=p.read_text();assert 'Trident +3' in original;p.write_text(original.replace('Trident +3','Trident +9',1))
    try:check(d)
    except AssertionError:checks.append(dict(case='altered_item_bonus',rejected=True))
    else:raise ValueError('更動加值未拒絕')
    p.write_text(original);p=d/'text/map-guide.en.tsv';original=p.read_text();assert '1/10\t3\t6\t3' in original;p.write_text(original.replace('1/10\t3\t6\t3','1/10\t3\t99\t3',1))
    try:check(d)
    except AssertionError:checks.append(dict(case='altered_map_coordinate',rejected=True))
    else:raise ValueError('更動座標未拒絕')
    (d/'receipt.json').write_text(json.dumps(checks,indent=2)+'\n');print(checks)

if __name__=='__main__':main()
