#!/usr/bin/env python3
"""量測既有Nuked OPL擷取的PCM，不產生或替換音源。"""
import argparse,json
from pathlib import Path
import numpy as np

def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('capture',type=Path);a=ap.parse_args()
    raw=np.fromfile(a.capture/'audio.s16le',dtype='<i2').astype(np.int32);assert len(raw)%(48000*2)==0
    peak=int(np.abs(raw).max());rms=float(np.sqrt(np.mean((raw.astype(np.float64)/32768)**2)))
    seconds=len(raw)/(48000*2);events=sum(1 for line in (a.capture/'opl.jsonl').open() if line.strip())
    assert seconds==85 and peak<32767 and rms>.01 and events==6660
    per=[]
    for n in range(int(seconds)):
        part=raw[n*96000:(n+1)*96000];per.append([n,round(float(np.sqrt(np.mean((part.astype(np.float64)/32768)**2))),6),int(np.abs(part).max())])
    result=dict(rate=48000,channels=2,seconds=seconds,samples=len(raw)//2,peak=peak,rms=rms,opl_events=events,per_second=per)
    (a.capture/'audio-metrics.json').write_text(json.dumps(result,indent=2)+'\n');print('PASS',seconds,'seconds, peak',peak,'OPL',events)
if __name__=='__main__':main()
