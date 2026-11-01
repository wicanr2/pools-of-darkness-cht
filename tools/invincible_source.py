#!/usr/bin/env python3
"""從既有正常快照的診斷RAM與IDA匯出保存本作欄位證據，不含原版素材。"""
import argparse,hashlib,json
from pathlib import Path

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def main():
 ap=argparse.ArgumentParser(description=__doc__)
 for n in ('root','case','receipt'):ap.add_argument('--'+n,type=Path,required=True)
 a=ap.parse_args();r=a.root;o=a.case;original=r/'workplace/input/dos/pod';ovr=(original/'GAME.OVR').read_bytes()
 assert sha(original/'GAME.OVR')=='885d6a07a3658f0688dc9d0630dfa984683a3d1d1e515a8e9909d29f62012485'
 ranges={'ida-final-fields.json':[(0x16afd,0x16b1b)],'ida-final-health.json':[(0x4e19,0x4e3b),(0x4ef8,0x4f0b),(0x6d01,0x6d29)],'ida-final-effects.json':[(0x3665f,0x366cf),(0x2392f,0x2394b)],'ida-final-harmful.json':[(0x366ef,0x367a0),(0x1d850,0x1d8e0)]}
 ida=[]
 for name,spans in ranges.items():
  d=json.loads((o/name).read_text());assert d['ida_version']=='9.4'and d['bitness']==0 and d['sha256']==sha(original/'GAME.OVR')
  rows=[x for x in d['rows']if any(lo<=x['file_offset']<hi for lo,hi in spans)]
  assert rows
  for row in rows:
   off=row['file_offset'];b=bytes.fromhex(row['bytes']);assert ovr[off:off+len(b)]==b
  ida.append(dict(path=name,sha256=sha(o/name),ida_version=d['ida_version'],ida_base=d['ida_base'],address_space=d['address_space'],rows=rows))
 scenes={}
 for name in ('town','battle'):
  p=o/(name+'.ram');b=p.read_bytes();assert len(b)==655360
  assert b[0x16ca0:0x16ca0+12].hex()=='803ede7d017505c6060e8801'
  far=lambda at:int.from_bytes(b[at:at+2],'little')+16*int.from_bytes(b[at+2:at+4],'little')
  recs=[];at=far(0x20960+0x9cf2);seen=set()
  while at:
   assert 0x10000<=at<0xa0000-0x1fd and at not in seen and len(seen)<64;seen.add(at)
   n=b[at];assert 1<=n<=15
   effects=[];node=far(at+0x143);visited=set()
   while node:
    assert node+9<=len(b)and node not in visited and len(visited)<64;visited.add(node)
    effects.append(dict(linear=node,type=b[node],next=far(node+5)));node=far(node+5)
   recs.append(dict(linear=at,name=b[at+1:at+1+n].decode('ascii'),HP=b[at+0xb2],combat_HP=b[at+0x1fc],status=b[at+0x1ed],active=b[at+0x1ee],side=b[at+0x1ef],next=far(at+0x1e3),effects=effects))
   at=far(at+0x1e3)
  scenes[name]=dict(ram_sha256=sha(p),records=recs)
 assert len(scenes['town']['records'])==1 and scenes['town']['records'][0]['HP']==40
 assert [x['HP']for x in scenes['battle']['records']if x['side']==0]==[162,124,80]
 b=(o/'town.ram').read_bytes();labels=[]
 for i in range(9):
  at=0x20960+0x1144+13*i;labels.append(b[at+1:at+1+b[at]].decode('ascii'))
 calls=[(0x2cbd1,0x21),(0x2e725,0x2a),(0x1d886,0x34),(0x2c16e,0x35)]
 for off,typ in calls:assert ovr[off:off+3]==bytes([0xb0,typ,0x50])
 poison=ovr.find(bytes.fromhex('b037508d7efc16579aa700ec00'),0x3c40,0x3cb0);assert poison>=0
 calls.append((poison,0x37))
 inputs={n:sha(original/n)for n in ('GAME.EXE','GAME.OVR')}
 proof=dict(date='2026-11-01',issue=72,level='confirmed field structure and consumers; semantic limits in029 note',inputs=inputs,
  address_spaces='runtime segment:offset for DS and RAM; OVR file offsets for instruction evidence; IDA EA plus ida_base retained separately',ida=ida,scenes=scenes,health_labels=labels,
  harmful_calls=[dict(OVR_offset=off,type=typ,bytes=ovr[off:off+3].hex())for off,typ in calls],
  snapshot_inputs={n:sha(r/n)for n in ('workplace/probe/g6-town/town.state','workplace/probe/demo-control-v1/snapshot/shots/battle.state')},
  limits=['diagnostic RAM from existing normal snapshots; not fresh normal gameplay acceptance','old0x1F and0x33 semantic labels unverified and excluded','no original EXE,OVR,RAM or state embedded in this receipt'])
 a.receipt.write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n');print(json.dumps(dict(status='PASS',town_party=1,battle_party=3,battle_records=len(scenes['battle']['records']),poison_call=hex(poison)),indent=2))
if __name__=='__main__':main()
