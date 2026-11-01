#!/usr/bin/env python3
"""BMA研究用唯讀解碼器；原版依據與位址基準見docs/re/031-intro-bma.md。"""
from dataclasses import dataclass
import hashlib
import struct


@dataclass(frozen=True)
class Span:
    file_offset: int
    y: int
    x: int
    pixels: bytes


@dataclass(frozen=True)
class Frame:
    block: int
    index: int
    file_offset: int
    size: int
    width: int
    height: int
    spans: tuple[Span, ...]


@dataclass(frozen=True)
class Block:
    index: int
    file_offset: int
    size: int
    sha256: str
    frames: tuple[Frame, ...]


def digest(data):
    return hashlib.sha256(data).hexdigest()


def decode_image(data, file_offset, block, index):
    if len(data) < 5:
        raise ValueError('圖面標頭或終止符不完整')
    width, height = struct.unpack_from('<HH', data)
    if not 1 <= width <= 320 or not 1 <= height <= 200:
        raise ValueError('本次已驗畫布範圍外')
    pos, spans = 4, []
    while pos < len(data):
        y = data[pos]
        pos += 1
        if y == 255:
            if pos != len(data):
                raise ValueError('終止符之後仍有資料')
            return Frame(block, index, file_offset, len(data), width, height, tuple(spans))
        if y >= height:
            raise ValueError('列超出圖面')
        while True:
            if pos + 4 > len(data):
                raise ValueError('跨度標頭不完整')
            start = pos
            raw, count, packed = struct.unpack_from('<HBB', data, pos)
            pos += 4
            x = raw & 0x7fff
            if count == 0 or packed == 0 or x + count > width:
                raise ValueError('跨度邊界或長度不合法')
            end = pos + packed
            if end > len(data):
                raise ValueError('壓縮資料超出圖面')
            pixels = bytearray()
            while pos < end:
                code = data[pos]
                pos += 1
                length = (code >> 1) + 1
                if len(pixels) + length > count:
                    raise ValueError('RLE展開超過跨度')
                if code & 1:
                    if pos >= end:
                        raise ValueError('重複色封包缺色號')
                    pixels.extend([data[pos]] * length)
                    pos += 1
                else:
                    if pos + length > end:
                        raise ValueError('連續色封包不完整')
                    pixels.extend(data[pos:pos+length])
                    pos += length
            if len(pixels) != count:
                raise ValueError('RLE展開長度不同')
            spans.append(Span(file_offset + start, y, x, bytes(pixels)))
            if raw & 0x8000:
                break
    raise ValueError('缺少圖面終止符')


def decode(data):
    if not data or len(data) > 16 * 1024 * 1024:
        raise ValueError('空檔或超出研究上限')
    at, blocks, images = 0, [], 0
    while at < len(data):
        if len(blocks) >= 1024 or at + 6 > len(data):
            raise ValueError('區塊標頭不完整或區塊過多')
        size, count = struct.unpack_from('<IH', data, at)
        table = 6 + 4 * count
        if not 1 <= count <= 4096 or not table <= size <= len(data) - at:
            raise ValueError('區塊長度或圖數不合法')
        images += count
        if images > 4096:
            raise ValueError('圖數超出研究上限')
        offsets = list(struct.unpack_from('<' + 'I' * count, data, at + 6)) + [size]
        if offsets[0] != table or any(a >= b for a, b in zip(offsets, offsets[1:])):
            raise ValueError('目錄位移不連續或倒置')
        if any(off < table or off >= size for off in offsets[:-1]):
            raise ValueError('目錄位移越界')
        frames = tuple(decode_image(data[at+lo:at+hi], at+lo, len(blocks), n)
                       for n, (lo, hi) in enumerate(zip(offsets, offsets[1:])))
        blocks.append(Block(len(blocks), at, size, digest(data[at:at+size]), frames))
        at += size
    return tuple(blocks)


def states(blocks):
    """本次正常INTRO依序呼叫各圖面；不宣稱ACF通用播放器。"""
    canvas = bytearray(64000)
    for block in blocks:
        for frame in block.frames:
            before = bytes(canvas)
            for span in frame.spans:
                start = span.y * 320 + span.x
                canvas[start:start+len(span.pixels)] = span.pixels
            yield frame, before, bytes(canvas)


def prefix(frame, base, target):
    """只接受與原版有序寫入前綴逐點相同的來源，不略過任意異色。"""
    if len(base) != len(target) or len(base) != 53760:
        raise ValueError('只核對上方168列')
    points, values = [], []
    for span in frame.spans:
        start = span.y * 320 + span.x
        for at, color in enumerate(span.pixels, start):
            if at < 53760:
                points.append(at)
                values.append(color)
    if len(set(points)) != len(points):
        return None  # 未驗重疊寫入的中途狀態，不能自行挑一個前綴。
    cut = 0
    for n, (at, color) in enumerate(zip(points, values)):
        if target[at] != base[at]:
            if target[at] != color:
                return None
            cut = n + 1
    got = bytearray(base)
    for at, color in zip(points[:cut], values[:cut]):
        got[at] = color
    return cut if got == target else None
