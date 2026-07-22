"""Headless stream-level verification of streakboard's kitty graphics.

Usage: python3 ptydump.py <demo-binary> <outdir>

Runs the demo on a pty (118x14 cells), answers its startup terminal
queries, sends q to quit, and reconstructs every kitty graphics
transmission from the raw byte stream. Writes:

  <outdir>/pty-raw.bin              the full captured stream
  <outdir>/board-from-pty-id<N>.png the exact PNG the app transmitted

Prints placement geometry, per-image byte counts, and the number of
U+10EEEE placeholder cells seen in the text stream.
"""
import base64
import fcntl
import os
import re
import select
import struct
import subprocess
import sys
import termios
import time

COLS, ROWS = 118, 14

demo, outdir = sys.argv[1], sys.argv[2]

master, slave = os.openpty()
fcntl.ioctl(slave, termios.TIOCSWINSZ,
            struct.pack("HHHH", ROWS, COLS, COLS * 8, ROWS * 16))

env = dict(os.environ, TERM="xterm-ghostty", COLORTERM="truecolor")
proc = subprocess.Popen([demo], stdin=slave, stdout=slave, stderr=slave,
                        env=env, close_fds=True)
os.close(slave)


def respond(buf: bytes):
    """Answer common startup queries so the app doesn't stall."""
    replies = []
    if b"\x1b[c" in buf or b"\x1b[0c" in buf:
        replies.append(b"\x1b[?62;4c")           # DA1
    for m in re.finditer(rb"\x1b\]1[01];\?(\x07|\x1b\\)", buf):
        n = m.group(0)[2:4]
        replies.append(b"\x1b]" + n + b";rgb:1e1e/1e1e/2e2e\x1b\\")  # OSC 10/11
    if b"\x1b[?u" in buf:
        replies.append(b"\x1b[?0u")              # kitty keyboard
    if b"\x1b[6n" in buf:
        replies.append(b"\x1b[1;1R")             # cursor position
    if b"\x1b[?996n" in buf:
        replies.append(b"\x1b[?997;1n")          # color scheme: dark
    for r in replies:
        os.write(master, r)


raw = bytearray()
deadline = time.time() + 4.0
sent_quit = False
answered = 0
while time.time() < deadline:
    r, _, _ = select.select([master], [], [], 0.25)
    if master in r:
        try:
            chunk = os.read(master, 65536)
        except OSError:
            break
        if not chunk:
            break
        raw.extend(chunk)
        if answered < 20:  # only during startup
            respond(chunk)
            answered += 1
    elif not sent_quit and raw and time.time() > deadline - 1.5:
        os.write(master, b"q")
        sent_quit = True
drain = time.time() + 1.0
while time.time() < drain:
    r, _, _ = select.select([master], [], [], 0.2)
    if master not in r:
        continue
    try:
        chunk = os.read(master, 65536)
    except OSError:
        break
    if not chunk:
        break
    raw.extend(chunk)
proc.terminate()

open(os.path.join(outdir, "pty-raw.bin"), "wb").write(raw)
print(f"captured {len(raw)} raw bytes from the pty")

apcs = re.findall(rb"\x1b_G([^\x1b]*)\x1b\\", bytes(raw))
print(f"kitty APC sequences found: {len(apcs)}")

images: dict[str, bytearray] = {}
order: list[str] = []
cur_id = None
for apc in apcs:
    ctrl, _, payload = apc.partition(b";")
    keys = dict(kv.split(b"=", 1) for kv in ctrl.split(b",") if b"=" in kv)
    action = keys.get(b"a", b"")
    if action == b"t":
        cur_id = keys.get(b"i", b"?").decode()
        images.setdefault(cur_id, bytearray()).extend(payload)
        if cur_id not in order:
            order.append(cur_id)
    elif action == b"" and cur_id is not None:  # m=1/m=0 continuation
        images[cur_id].extend(payload)
    elif action == b"p":
        print(f"  placement: id={keys.get(b'i', b'?').decode()} "
              f"rows={keys.get(b'r', b'?').decode()} "
              f"cols={keys.get(b'c', b'?').decode()}")
    elif action == b"d":
        print(f"  delete: id={keys.get(b'i', b'?').decode()}")

for img_id in order:
    png = base64.b64decode(images[img_id])
    path = os.path.join(outdir, f"board-from-pty-id{img_id}.png")
    open(path, "wb").write(png)
    magic = png[:8] == b"\x89PNG\r\n\x1a\n"
    print(f"  image id={img_id}: {len(png)} bytes -> {path} "
          f"(valid PNG magic: {magic})")

text = raw.decode("utf-8", errors="replace")
print(f"U+10EEEE placeholder cells in text stream: {text.count(chr(0x10EEEE))}")
