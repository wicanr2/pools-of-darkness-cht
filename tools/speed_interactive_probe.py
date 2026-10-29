#!/usr/bin/env python3
"""隔離副本只附加讀值；用私人Xvfb的正常關窗事件結束互動程式。"""
import argparse
import ctypes as c
import ctypes.util
import json
import os
from pathlib import Path
import shutil
import subprocess
import time


def prepare(module):
    src = module / 'cmd/pod-play'
    dst = module / 'workplace/speed-interactive-cmd'
    if dst.exists():
        shutil.rmtree(dst)
    dst.mkdir()
    for p in src.glob('*.go'):
        shutil.copy2(p, dst / p.name)
    p = dst / 'main.go'
    text = p.read_text()
    replacements = {
        'func (g *game) Update() error {': 'func (g *game) Update() error {\n\tdefer g.receiptUpdate(time.Now())',
        'start, capped, stopped := time.Now(), false, false': 'start, capped, stopped := time.Now(), false, false\n\treceiptBudget, receiptClock = g.budget, g.owner.ClockPercent()\n\tdefer func(){receiptCapped=capped}()',
        'func (g *game) Draw(screen *ebiten.Image) {': 'func (g *game) Draw(screen *ebiten.Image) {\n\tdefer g.receiptDraw(time.Now())',
        'if g.frames > 0 { // 自動模式：以畫格命名，收據重跑得到同名檔': 'if g.frames > 0 || os.Getenv("POD_SPEED_RECEIPT") != "" { // 研究截圖使用畫格名稱',
        'err = ebiten.RunGame(g)': 'err = ebiten.RunGame(g)\n\tg.receiptFinish()',
    }
    for a, b in replacements.items():
        assert text.count(a) == 1, a
        text = text.replace(a, b, 1)
    p.write_text(text)
    (dst / 'speed_receipt.go').write_text(r'''package main
import("encoding/json";"fmt";"os";"path/filepath";"time")
var receiptStart=time.Now()
var receiptUpdates []map[string]any
var receiptDraws []map[string]any
var receiptPreviousSteps uint64
var receiptBudget uint64
var receiptClock int
var receiptCapped bool
func receiptWrite(name string,v any){
 dir:=os.Getenv("POD_SPEED_RECEIPT");if dir==""{return}
 b,e:=json.Marshal(v);if e!=nil{panic(e)}
 p:=filepath.Join(dir,name);if e=os.WriteFile(p+".tmp",b,0600);e!=nil{panic(e)}
 if e=os.Rename(p+".tmp",p);e!=nil{panic(e)}
}
func(g *game)receiptUpdate(start time.Time){
 if os.Getenv("POD_SPEED_RECEIPT")==""{return}
 advanced:=int64(0);if n:=len(g.perf.adv);n>0{advanced=int64(g.perf.adv[n-1])}
 receiptUpdates=append(receiptUpdates,map[string]any{"frame":g.frame,"start_ns":start.Sub(receiptStart).Nanoseconds(),"total_ns":time.Since(start).Nanoseconds(),"advance_ns":advanced,"set":g.speed.set,"clock_used":receiptClock,"clock_next":g.owner.ClockPercent(),"budget":receiptBudget,"steps":g.lastSteps-receiptPreviousSteps,"capped":receiptCapped,"ready":g.speed.ready,"focused":g.focused,"scale":g.scale,"draw_interval_ns":int64(g.speed.drawInterval)})
 receiptPreviousSteps=g.lastSteps
 if g.frame==4800{
  d,e:=g.owner.Digest();if e!=nil{panic(e)}
  receiptWrite("checkpoint-4800.json",map[string]any{"steps":d.Steps,"cpu":fmt.Sprintf("%x",d.CPUSHA256),"memory":fmt.Sprintf("%x",d.MemorySHA256),"indexed":fmt.Sprintf("%x",d.IndexedSHA256),"palette":fmt.Sprintf("%x",d.PaletteSHA256)})
 }
 if g.frame>=4800 && g.frame%200==0 {receiptWrite("timing.partial.json",map[string]any{"updates":receiptUpdates,"draws":receiptDraws,"frames":g.frame,"interactive":g.frames==0})}
 if g.frame==6200 {receiptWrite("ready-to-close.json",map[string]any{"frame":g.frame})}
}
func(g *game)receiptDraw(start time.Time){
 if os.Getenv("POD_SPEED_RECEIPT")==""{return}
 receiptDraws=append(receiptDraws,map[string]any{"frame":g.frame,"start_ns":start.Sub(receiptStart).Nanoseconds(),"total_ns":time.Since(start).Nanoseconds()})
}
func(g *game)receiptFinish(){
 if os.Getenv("POD_SPEED_RECEIPT")==""{return}
 receiptWrite("timing.json",map[string]any{"updates":receiptUpdates,"draws":receiptDraws,"frames":g.frame,"interactive":g.frames==0})
 g.perf.report(g.speed.base)
}
''')
    # ff7dca8的修改前對照尚無繪製回饋欄位，仍可用同一入口重生其量測。
    if 'drawInterval' not in (src / 'speed.go').read_text():
        probe = dst / 'speed_receipt.go'
        probe.write_text(probe.read_text().replace(',"scale":g.scale,"draw_interval_ns":int64(g.speed.drawInterval)', ''))
    return dst


def close_window():
    """XSendEvent只送WM_DELETE_WINDOW；私人DISPLAY沒有其他專案視窗。"""
    lib = c.CDLL(ctypes.util.find_library('X11'))
    lib.XOpenDisplay.argtypes = [c.c_char_p]
    lib.XOpenDisplay.restype = c.c_void_p
    d = lib.XOpenDisplay(None)
    assert d, 'XOpenDisplay'
    lib.XDefaultRootWindow.argtypes = [c.c_void_p]
    lib.XDefaultRootWindow.restype = c.c_ulong
    lib.XQueryTree.argtypes = [c.c_void_p, c.c_ulong, c.POINTER(c.c_ulong), c.POINTER(c.c_ulong), c.POINTER(c.POINTER(c.c_ulong)), c.POINTER(c.c_uint)]
    lib.XFetchName.argtypes = [c.c_void_p, c.c_ulong, c.POINTER(c.c_char_p)]
    lib.XInternAtom.argtypes = [c.c_void_p, c.c_char_p, c.c_int]
    lib.XInternAtom.restype = c.c_ulong
    lib.XFree.argtypes = [c.c_void_p]
    lib.XGetWMProtocols.argtypes = [c.c_void_p, c.c_ulong, c.POINTER(c.POINTER(c.c_ulong)), c.POINTER(c.c_int)]
    lib.XFlush.argtypes = [c.c_void_p]
    lib.XCloseDisplay.argtypes = [c.c_void_p]
    root, parent = c.c_ulong(), c.c_ulong()
    children, n = c.POINTER(c.c_ulong)(), c.c_uint()
    assert lib.XQueryTree(d, lib.XDefaultRootWindow(d), c.byref(root), c.byref(parent), c.byref(children), c.byref(n))
    window = None
    names = []
    delete = lib.XInternAtom(d, b'WM_DELETE_WINDOW', 0)
    for i in range(n.value):
        name = c.c_char_p()
        if lib.XFetchName(d, children[i], c.byref(name)) and name.value:
            title = name.value.decode('utf-8', errors='replace')
            names.append(title)
            lib.XFree(name)
        protocols, count = c.POINTER(c.c_ulong)(), c.c_int()
        if lib.XGetWMProtocols(d, children[i], c.byref(protocols), c.byref(count)):
            if delete in [protocols[k] for k in range(count.value)]:
                assert window is None, '私人DISPLAY中有多個可關閉視窗'
                window = children[i]
            lib.XFree(protocols)
    lib.XFree(children)
    assert window, names
    class Data(c.Union):
        _fields_ = [('l', c.c_long * 5)]
    class Client(c.Structure):
        _fields_ = [('type', c.c_int), ('serial', c.c_ulong), ('send_event', c.c_int), ('display', c.c_void_p), ('window', c.c_ulong), ('message_type', c.c_ulong), ('format', c.c_int), ('data', Data)]
    class Event(c.Union):
        _fields_ = [('client', Client), ('pad', c.c_long * 24)]
    event = Event()
    event.client.type, event.client.send_event = 33, 1
    event.client.display, event.client.window, event.client.format = d, window, 32
    event.client.message_type = lib.XInternAtom(d, b'WM_PROTOCOLS', 0)
    event.client.data.l[0] = delete
    lib.XSendEvent.argtypes = [c.c_void_p, c.c_ulong, c.c_int, c.c_long, c.POINTER(Event)]
    assert lib.XSendEvent(d, window, 0, 0, c.byref(event))
    lib.XFlush(d)
    lib.XCloseDisplay(d)
    return {'title': '黑暗之池', 'event': 'WM_DELETE_WINDOW'}


def run(out, command):
    assert command and not (out / 'timing.json').exists()
    def resources():
        cg = Path('/sys/fs/cgroup')
        return {name: (cg / name).read_text() for name in ['cpu.max', 'cpu.stat', 'memory.max', 'pids.max']}
    (out / 'resources-start.json').write_text(json.dumps(resources()) + '\n')
    env = dict(os.environ, POD_SPEED_RECEIPT=str(out))
    with (out / 'run.log').open('w') as log:
        process = subprocess.Popen(command, env=env, stdout=log, stderr=subprocess.STDOUT)
        try:
            deadline = time.monotonic() + 1000
            while not (out / 'ready-to-close.json').exists():
                assert process.poll() is None, f'提前退出：{process.returncode}'
                assert time.monotonic() < deadline, '互動收據逾時'
                time.sleep(0.1)
            event = close_window()
            process.wait(timeout=20)
            assert process.returncode == 0, process.returncode
            (out / 'close.json').write_text(json.dumps(event) + '\n')
            (out / 'resources-end.json').write_text(json.dumps(resources()) + '\n')
        finally:
            if process.poll() is None:
                process.terminate()
                try:
                    process.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.wait()


if __name__ == '__main__':
    ap = argparse.ArgumentParser()
    ap.add_argument('mode', choices=['prepare', 'run'])
    ap.add_argument('path', type=Path)
    ap.add_argument('command', nargs=argparse.REMAINDER)
    args = ap.parse_args()
    if args.mode == 'prepare':
        prepare(args.path)
    else:
        run(args.path, args.command)
