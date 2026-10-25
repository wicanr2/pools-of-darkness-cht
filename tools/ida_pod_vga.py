"""IDA9.4研究匯出：保留VGA.DRV原始位址與名稱，不改語意名稱。"""
import hashlib,json
from pathlib import Path
import ida_auto,ida_bytes,ida_funcs,ida_kernwin,ida_segment,idautils,idc

out=Path('/out');source=Path('/orig/VGA.DRV');raw=source.read_bytes()
assert hashlib.sha256(raw).hexdigest()=='eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f'
segment=ida_segment.get_first_seg();assert segment is not None
ida_segment.set_segm_addressing(segment,0)
base=segment.start_ea;entries=[]
for off in range(0,len(raw)-3,3):
 if raw[off]!=0xE9:break
 displacement=int.from_bytes(raw[off+1:off+3],'little',signed=True);target=(off+3+displacement)&0xFFFF
 assert 0<=target<len(raw)
 idc.create_insn(base+off);idc.create_insn(base+target);ida_funcs.add_func(base+target)
 entries.append(dict(slot=off//3,file_offset=off,target_file_offset=target,bytes=raw[off:off+3].hex()))
ida_auto.auto_wait();functions=[]
for address in idautils.Functions():
 f=ida_funcs.get_func(address);instructions=[]
 for ea in idautils.FuncItems(address):
  size=idc.get_item_size(ea);data=ida_bytes.get_bytes(ea,size)
  instructions.append(dict(ida_ea=ea,file_offset=ea-base,bytes=data.hex()if data else'',mnemonic=idc.print_insn_mnem(ea),operands=[idc.print_operand(ea,n)for n in range(3)if idc.print_operand(ea,n)],xrefs_to=[dict(from_ea=x.frm,type=x.type)for x in idautils.XrefsTo(ea)]))
 functions.append(dict(original_name=ida_funcs.get_func_name(address),ida_start=address,ida_end=f.end_ea,file_offset=address-base,instructions=instructions))
report=dict(schema=1,input=source.name,input_sha256=hashlib.sha256(raw).hexdigest(),ida_version=ida_kernwin.get_kernel_version(),address_space='IDA binary segment base; file_offset=ida_ea-segment_base; not runtime CS:IP',segment_base=base,bitness=ida_segment.get_first_seg().bitness,jump_entries=entries,functions=functions,evidence_level='confirmed bytes and IDA boundaries; semantic unknown')
(out/'vga-ida.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n');idc.qexit(0)
