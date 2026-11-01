"""IDA9.4一次性16-bit GAME.OVR資料庫的窄範圍匯出，不覆寫正式DB。"""
import hashlib,json
from pathlib import Path
import ida_auto,ida_bytes,ida_funcs,ida_kernwin,ida_segment,idautils,idc

source=Path(idc.get_input_file_path());raw=source.read_bytes()
assert hashlib.sha256(raw).hexdigest()=='885d6a07a3658f0688dc9d0630dfa984683a3d1d1e515a8e9909d29f62012485'
segment=ida_segment.get_first_seg();ida_segment.set_segm_addressing(segment,0);base=segment.start_ea
groups={'fields':[(0x16afd,0x16b1b)],'health':[(0x4e19,0x4e3b),(0x4ef8,0x4f0b),(0x6d01,0x6d29)],'effects':[(0x3665f,0x366cf),(0x2392f,0x2394b)],'harmful':[(0x366ef,0x36800),(0x1d6b4,0x1d8e0)]}
for name,spans in groups.items():
 rows=[]
 for lo,hi in spans:
  ida_bytes.del_items(base+lo,ida_bytes.DELIT_SIMPLE,hi-lo);at=lo
  while at<hi:
   size=idc.create_insn(base+at);at+=max(size,1)
  ida_auto.auto_wait()
  for at in idautils.Heads(base+lo,base+hi):
   size=idc.get_item_size(at);f=ida_funcs.get_func(at)
   rows.append(dict(file_offset=at-base,ida_ea=at,bytes=ida_bytes.get_bytes(at,size).hex(),text=idc.generate_disasm_line(at,0),original_function=ida_funcs.get_func_name(f.start_ea)if f else None,xrefs=[dict(source=x.frm,type=x.type)for x in idautils.XrefsTo(at)]))
 proof=dict(input=source.name,sha256=hashlib.sha256(raw).hexdigest(),ida_version=ida_kernwin.get_kernel_version(),ida_base=base,bitness=segment.bitness,address_space='GAME.OVR file offset; IDA EA separately retained; no runtime CS claim',rows=rows)
 Path('/out','ida-final-'+name+'.json').write_text(json.dumps(proof,indent=2)+'\n')
idc.qexit(0)
