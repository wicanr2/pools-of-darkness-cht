#!/usr/bin/env python3
"""研究031的格式邊界及真實輸入檢查；在Docker內執行。"""
from pathlib import Path
import struct
import sys
import unittest

sys.path.insert(0, str(Path(__file__).resolve().parent))
from intro_bma import decode, decode_image, digest, prefix, states


def archive(*images):
    at = 6 + 4 * len(images)
    offsets = []
    for image in images:
        offsets.append(at)
        at += len(image)
    return struct.pack('<IH', at, len(images)) + struct.pack('<'+'I'*len(images), *offsets) + b''.join(images)


class BMA(unittest.TestCase):
    def test_literal_repeat_and_rows(self):
        image = struct.pack('<HH', 8, 2) + bytes.fromhex('00 00 00 03 04 04 01 02 03 04 80 02 02 03 09 ff')
        block = decode(archive(image))[0]
        self.assertEqual([(s.y, s.x, s.pixels) for s in block.frames[0].spans], [(0, 0, b'\1\2\3'), (0, 4, b'\11\11')])
        self.assertEqual(list(states((block,)))[0][2][:8], b'\1\2\3\0\11\11\0\0')

    def test_noop_preserves_canvas(self):
        block = decode(archive(struct.pack('<HH', 8, 2)+b'\xff'))[0]
        self.assertFalse(block.frames[0].spans)
        self.assertEqual(list(states((block,)))[0][2], bytes(64000))

    def test_archive_rejections(self):
        good = archive(struct.pack('<HH', 8, 2)+b'\xff')
        cases = [b'', b'\0'*5, struct.pack('<IH', 1000, 1)+good[6:],
                 struct.pack('<IH', len(good), 0)+good[6:],
                 good[:6]+struct.pack('<I', 11)+good[10:],
                 good[:6]+struct.pack('<I', 999)+good[10:], good[:-1], good+b'\0']
        for bad in cases:
            with self.subTest(data=bad.hex()), self.assertRaises(ValueError):
                decode(bad)

    def test_image_rejections(self):
        head = struct.pack('<HH', 8, 2)
        cases = [b'', struct.pack('<HH', 0, 2)+b'\xff', head+b'\x02\xff',
                 head+b'\x00\x80', head+bytes.fromhex('00 08 80 01 02 01 07 ff'),
                 head+bytes.fromhex('00 00 80 00 02 01 07 ff'),
                 head+bytes.fromhex('00 00 80 01 00 ff'),
                 head+bytes.fromhex('00 00 80 01 20 ff'),
                 head+bytes.fromhex('00 00 80 01 01 01 ff'),
                 head+bytes.fromhex('00 00 80 03 02 04 07 ff'),
                 head+bytes.fromhex('00 00 80 01 02 03 07 ff'),
                 head+bytes.fromhex('00 00 80 02 02 01 07 ff'),
                 head+bytes.fromhex('00 00 80 01 02 01 07'), head+b'\xff\x00']
        for bad in cases:
            with self.subTest(data=bad.hex()), self.assertRaises(ValueError):
                decode_image(bad, 0, 0, 0)

    def test_prefix_requires_all_other_points(self):
        image = struct.pack('<HH', 8, 2)+bytes.fromhex('00 00 80 03 04 04 01 02 03 ff')
        frame = decode(archive(image))[0].frames[0]
        base = bytes(53760)
        target = bytearray(base)
        target[:2] = b'\1\2'
        self.assertEqual(prefix(frame, base, bytes(target)), 2)
        target[30000] = 1
        self.assertIsNone(prefix(frame, base, bytes(target)))

    def test_original(self):
        source = Path('/project/workplace/input/dos/pod/INTRO.BMA')
        if not source.is_file():
            self.skipTest('缺自備原版，不算原版來源驗收')
        data = source.read_bytes()
        self.assertEqual(digest(data), '12f3944652b2ae77feb3c0b16435cb2a3d855c26c388e6d6f97c277abedcb94f')
        blocks = decode(data)
        self.assertEqual((len(blocks), sum(len(b.frames) for b in blocks)), (14, 133))
        self.assertEqual(len({digest(after[:53760]) for _, _, after in states(blocks)}), 98)


if __name__ == '__main__':
    unittest.main()
