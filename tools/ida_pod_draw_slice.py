"""匯出已觀測GAME繪圖段；IDA offset與runtime CS:IP分別保留。"""
import hashlib,json
from pathlib import Path
import ida_auto,ida_bytes,ida_funcs,ida_kernwin,ida_segment,idautils,idc
name=Path(idc.get_input_file_path()).name;segment=int(name.split('-')[1].split('.')[0],16)
source=Path('/case')/name;data=source.read_bytes();proof=next(i for i in json.loads(Path('/case/slices.json').read_text())if i['file']==name);assert hashlib.sha256(data).hexdigest()==proof['sha256']
s=ida_segment.get_first_seg();ida_segment.set_segm_addressing(s,0);base=s.start_ea
lo,hi={0x1ed4:(0x19b0,0x1ab0),0x1d1f:(0x3e0,0x490),0x1b2e:(0x1160,0x1210)}[segment]
ida_bytes.del_items(base+lo,ida_bytes.DELIT_SIMPLE,hi-lo)
at=lo
while at<hi:
 size=idc.create_insn(base+at)
 if size<=0:at+=1
 else:at+=size
ida_auto.auto_wait();rows=[]
for at in idautils.Heads(base+lo,base+hi):
 size=idc.get_item_size(at);f=ida_funcs.get_func(at)
 rows.append(dict(file_offset=at-base,ida_ea=at,runtime_cs=segment,runtime_ip=at-base,bytes=ida_bytes.get_bytes(at,size).hex(),mnemonic=idc.print_insn_mnem(at),operands=[idc.print_operand(at,n)for n in range(3)if idc.print_operand(at,n)],original_function=ida_funcs.get_func_name(f.start_ea)if f else None))
report=dict(schema=1,input=name,sha256=proof['sha256'],ida_version=ida_kernwin.get_kernel_version(),runtime_cs=segment,ida_base=base,address_space='binary file offsets == runtime IP of recorded CS; IDA EA separately retained',rows=rows)
Path('/out',name+'.json').write_text(json.dumps(report,indent=2)+'\n');idc.qexit(0)
