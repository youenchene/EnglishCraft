"""Generate original Alex pixel-art PWA icons with no licensed dependencies."""
import struct
import zlib
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1] / "web" / "public"
SHAPES = [
    (7, 3, 18, 6, (24, 42, 73)), (5, 7, 22, 10, (24, 42, 73)),
    (7, 9, 18, 13, (214, 164, 120)), (10, 12, 3, 3, (24, 42, 73)),
    (20, 12, 3, 3, (24, 42, 73)), (13, 18, 7, 2, (135, 62, 67)),
    (8, 22, 16, 12, (39, 150, 131)), (3, 23, 5, 10, (39, 150, 131)),
    (24, 23, 5, 10, (39, 150, 131)), (14, 23, 4, 6, (241, 202, 101)),
    (8, 34, 7, 6, (24, 42, 73)), (18, 34, 6, 6, (24, 42, 73)),
]


def chunk(kind, data):
    return struct.pack(">I", len(data)) + kind + data + struct.pack(">I", zlib.crc32(kind + data))


def generate(size, name):
    pixels = bytearray((239, 246, 255) * size * size)
    scale = size / 64
    for x, y, width, height, color in SHAPES:
        for row in range(int((y + 12) * scale), int((y + 12 + height) * scale)):
            for col in range(int((x + 16) * scale), int((x + 16 + width) * scale)):
                offset = (row * size + col) * 3
                pixels[offset:offset + 3] = bytes(color)
    raw = b"".join(b"\0" + pixels[row * size * 3:(row + 1) * size * 3] for row in range(size))
    png = b"\x89PNG\r\n\x1a\n"
    png += chunk(b"IHDR", struct.pack(">IIBBBBB", size, size, 8, 2, 0, 0, 0))
    png += chunk(b"IDAT", zlib.compress(raw)) + chunk(b"IEND", b"")
    (ROOT / name).write_bytes(png)


for dimension, filename in [(192, "icon-192.png"), (512, "icon-512.png"), (180, "apple-touch-icon.png")]:
    generate(dimension, filename)
