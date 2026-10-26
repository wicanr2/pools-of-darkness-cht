"""Generate fixed Go drawing effects from IDA94-verified instructions.

Run only in sc-cht-disasm:capstone5.0.3-r1. Inputs remain private; output
contains effects and hashes, not executable bytes. Unknown/system operations
end a block and execute in the original CPU. See research019/spec025.
"""
import argparse, hashlib, json
from pathlib import Path
from capstone import Cs, CS_ARCH_X86, CS_MODE_16
from capstone.x86 import X86_OP_REG, X86_OP_MEM, X86_OP_IMM

p=argparse.ArgumentParser();p.add_argument('--project',type=Path,required=True);p.add_argument('--out',type=Path,required=True);p.add_argument('--package',choices=['drawblocks','drawall'],default='drawblocks');a=p.parse_args()
root=a.project; research=root/'workplace/probe/draw-intercept-v1'
md=Cs(CS_ARCH_X86,CS_MODE_16);md.detail=True
vga=(root/'workplace/input/dos/pod/VGA.DRV').read_bytes()
assert hashlib.sha256(vga).hexdigest()=='eef4e14b54201696e779c4a35fb71582825a5569e921e8196c536672ee9cc19f'
j=json.loads((research/'vga-ida.json').read_text());assert j['ida_version']=='9.4'
entries=[];table={};functions=[]
for f in j['functions']:
    # Admit the original driver's drawing entries and local helpers. Mixed
    # hardware entries stop at in/out/int/POPF; those boundaries stay original.
    lo,hi=f['ida_start'],f['ida_end']
    functions.append((0x48a1,lo,hi,f['original_name']))
    for r in f['instructions']:
        ip=r['file_offset'];raw=bytes.fromhex(r['bytes']);assert vga[ip:ip+len(raw)]==raw
        ins=list(md.disasm(raw,ip,count=1));assert len(ins)==1 and ins[0].size==len(raw),(hex(ip),r)
        table[(0x48a1,ip)]=(ins[0],f['original_name'])
for seg,lo,hi in [(0x1ed4,0x1a5c,0x1a7f),(0x1ed4,0x1a7f,0x1a93),(0x1d1f,0x46e,0x487)]:
    j=json.loads((research/f'ida-game/game-{seg:04x}.bin.json').read_text()); rows=[r for r in j['rows'] if lo<=r['runtime_ip']<hi]
    functions.append((seg,lo,hi,f'sub_{lo:X}'))
    for r in rows:
        raw=bytes.fromhex(r['bytes']);ins=list(md.disasm(raw,r['runtime_ip'],count=1));assert len(ins)==1 and ins[0].size==len(raw)
        table[(seg,r['runtime_ip'])]=(ins[0],f'sub_{lo:X}')

regs={n:f'e.s.R[{i}]' for i,n in enumerate(['ax','cx','dx','bx','sp','bp','si','di'])}
regs.update({n:f'e.s.Seg[{i}]' for i,n in enumerate(['es','cs','ss','ds'])})
reg8={n:i for i,n in enumerate(['al','cl','dl','bl','ah','ch','dh','bh'])}
def off(o):
    m=o.mem; terms=[]
    if m.base:terms.append(regs[md.reg_name(m.base)])
    if m.index:terms.append(regs[md.reg_name(m.index)])
    terms.append(str(m.disp&65535));return '('+'+'.join(terms)+')'
def seg(o):
    m=o.mem
    return regs[md.reg_name(m.segment)] if m.segment else ('e.s.Seg[2]' if md.reg_name(m.base)=='bp' else 'e.s.Seg[3]')
def get(o):
    if o.type==X86_OP_IMM:return str(o.imm&65535)
    if o.type==X86_OP_REG:
        name=md.reg_name(o.reg);return regs[name] if name in regs else f'e.reg8({reg8[name]})'
    if o.type==X86_OP_MEM:return f'e.read({seg(o)},{off(o)},{o.size*8})'
    raise ValueError('operand')
def put(o,v):
    if o.type==X86_OP_REG:
        name=md.reg_name(o.reg)
        return f'{regs[name]}=uint16({v})' if name in regs else f'e.set8({reg8[name]},uint16({v}))'
    return f'e.write({seg(o)},{off(o)},uint16({v}),{o.size*8})'
def op(i):return i.mnemonic.split()[-1]
def cost(i):
    raw=bytes(i.bytes);n=0
    while raw[n] in (0x26,0x2e,0x36,0x3e,0xf2,0xf3):n+=1
    c=2+n
    if i.modrm_offset and any(o.type==X86_OP_MEM for o in i.operands):c+=4
    if op(i) in ('les','lds')or(op(i)=='mov'and i.operands[0].type==X86_OP_REG and md.reg_name(i.operands[0].reg)in ('es','ds','ss')):c+=18
    return c
alu={'add':0,'or':1,'adc':2,'sbb':3,'and':4,'sub':5,'xor':6,'cmp':7,'test':4}
shift={'rol':0,'ror':1,'rcl':2,'rcr':3,'shl':4,'shr':5,'sar':7}
def body(i):
    name=op(i);o=i.operands;nxt=(i.address+i.size)&65535;s=[];end=False
    if any(x in i.prefix for x in (0x66,0x67,0xf0)):return None
    if any(x.size>2 for x in o) and name not in ('les','lds'):return None
    if name=='mov':s=[put(o[0],get(o[1]))]
    elif name in alu:
        value=f'e.alu({alu[name]},uint16({get(o[0])}),uint16({get(o[1])}),{o[0].size*8})'
        s=[value if name in ('cmp','test') else put(o[0],value)]
    elif name in ('inc','dec'):s=[put(o[0],f'e.inc({get(o[0])},{o[0].size*8},{str(name=="dec").lower()})')]
    elif name=='neg':s=[put(o[0],f'e.neg({get(o[0])},{o[0].size*8})')]
    elif name in shift:s=[put(o[0],f'e.shift({shift[name]},{get(o[0])},uint16({get(o[1])}),{o[0].size*8})')]
    elif name in ('mul','div'):s=[f'e.{name}({get(o[0])},{o[0].size*8})']
    elif name=='lea':s=[put(o[0],off(o[1]))]
    elif name=='xchg':s=['{t:='+get(o[0])+';'+put(o[0],get(o[1]))+';'+put(o[1],'t')+'}']
    elif name in ('les','lds'):
        target=0 if name=='les'else 3
        s=['{s,o:='+seg(o[1])+',uint16('+off(o[1])+');v:=e.read(s,o,16);t:=e.read(s,o+2,16);'+put(o[0],'v')+f';e.s.Seg[{target}]=t'+'}']
    elif name=='push':
        if o[0].type==X86_OP_REG and md.reg_name(o[0].reg)=='sp':return None
        s=[f'e.push(uint16({get(o[0])}))']
    elif name=='pop':s=[put(o[0],'e.pop()')]
    elif name=='pushf':s=['e.push(e.s.Flags)']
    # POPF depends on CPU model and may enable tracing/interrupts. Keep the
    # original instruction as an observation/event boundary.
    elif name in ('cld','std','clc','stc'):
        flag=0x400 if name in ('cld','std')else 1;s=[f'e.set({flag},{str(name in ("std","stc")).lower()})']
    elif name=='cwd':s=['e.s.R[2]=0;if e.s.R[0]&0x8000!=0{e.s.R[2]=0xffff}']
    elif name=='xlatb':s=['e.set8(0,e.read(e.s.Seg[3],e.s.R[3]+e.reg8(0),8))']
    elif name in ('movsb','movsw','stosb','stosw','lodsb','lodsw','cmpsw','scasb'):
        raw=bytes(i.bytes);opcode=raw[-1];rp=next((x for x in raw[:-1] if x in (0xf2,0xf3)),0);src=next(({0x26:0,0x2e:1,0x36:2,0x3e:3}[x]for x in raw[:-1]if x in (0x26,0x2e,0x36,0x3e)),3)
        s=[f'e.str({opcode},{rp},{src})']
    elif name=='call':
        if len(o)!=1 or o[0].type!=X86_OP_IMM:return None
        s=[f'e.charge(7);e.push({nxt});e.s.IP={get(o[0])}'];end=True
    elif name in ('retf','ret'):
        extra=get(o[0]) if o else '0'
        s=[f'e.charge({18 if name=="retf"else 8});e.s.IP=e.pop()']
        if name=='retf':s+=['e.s.Seg[1]=e.pop()']
        s+=[f'e.s.R[4]+={extra}'];end=True
    elif name=='jmp':
        if len(o)!=1 or o[0].type!=X86_OP_IMM:return None
        s=[f'e.charge(5);e.s.IP={get(o[0])}'];end=True
    elif name.startswith('j') and len(o)==1 and o[0].type==X86_OP_IMM:
        if name=='jcxz':cond='e.s.R[1]==0';more=''
        else:cond=f'e.cond({i.opcode[0]&15})';more='e.charge(5);'
        s=[f'e.s.IP={nxt};if {cond}'+'{'+more+f'e.s.IP={get(o[0])}'+'}'];end=True
    elif name=='loop':s=[f'e.s.R[1]--;e.s.IP={nxt};if e.s.R[1]!=0'+'{'+f'e.s.IP={get(o[0])}'+'}'];end=True
    elif name=='nop':s=[]
    else:return None
    if not end:s+=[f'e.s.IP={nxt}']
    return s,end

supported={};unsupported=[]
for key,(ins,fn) in table.items():
    try:r=body(ins)
    except (KeyError,ValueError):r=None
    if r is None:unsupported.append({'cs':key[0],'ip':key[1],'mnemonic':ins.mnemonic,'operands':ins.op_str,'function':fn})
    else:supported[key]=(ins,fn,r)
starts=set((cs,lo)for cs,lo,hi,name in functions)
for (cs,ip),(ins,fn)in table.items():
    if op(ins).startswith('j')or op(ins) in ('loop','call'):
        for o in ins.operands:
            if o.type==X86_OP_IMM:starts.add((cs,o.imm&65535))
    if (cs,ip)not in supported or supported[(cs,ip)][2][1]:starts.add((cs,(ip+ins.size)&65535))

groups=[];claimed=set()
for key in sorted(supported):
    if key in claimed:continue
    cs,ip=key;seq=[];size=0
    while key in supported and key not in claimed:
        ins,fn,(stmts,end)=supported[key];seq.append((ins,stmts));claimed.add(key);size+=ins.size
        nxt=(cs,(ins.address+ins.size)&65535)
        if end or len(seq)>=16 or nxt in starts or size>=96:break
        key=nxt
    groups.append((cs,ip,seq))

a.out.mkdir(parents=True,exist_ok=True)
lines=['// Code generated by tools/draw_blocks_gen.py from IDA94-verified drawing instructions. DO NOT EDIT.',f'package {a.package}','var blocks=map[uint32]block{']
manifest=[]
for block_id,(cs,ip,seq) in enumerate(groups):
    code=b''.join(bytes(i.bytes)for i,s in seq);sha=hashlib.sha256(code).digest();addr=cs*16+ip;func=f'block_{cs:04x}_{ip:04x}'
    lines.append(f'{addr}:{{id:{block_id},cs:{cs},n:{len(seq)},size:{len(code)},sha:[32]byte{{'+','.join(str(x)for x in sha)+f'}},run:{func}'+'},')
    manifest.append({'cs':cs,'ip':ip,'instructions':len(seq),'bytes':len(code),'sha256':sha.hex(),'function':table[(cs,ip)][1]})
lines+=['}',f'const blockCount={len(groups)}']
for cs,ip,seq in groups:
    lines.append(f'func block_{cs:04x}_{ip:04x}(e *effect)'+'{')
    for ins,stmts in seq:
        lines.append(f'e.mark({ins.address},{cost(ins)})')
        lines.extend(stmts)
        lines.append('if e.bad{return}')
    lines.append('}')
(a.out/'blocks_generated.go').write_text('\n'.join(lines)+'\n')
(a.out/'blocks.json').write_text(json.dumps({'schema':1,'functions':functions,'blocks':manifest,'original_only':unsupported,'generator_sha256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest()},indent=2)+'\n')
print('functions',len(functions),'blocks',len(groups),'instructions',len(claimed),'original_only',len(unsupported))
print('original-only mnemonics',sorted(set(x['mnemonic']for x in unsupported)))
