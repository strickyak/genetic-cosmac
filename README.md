genetic-cosmac
==============

RCA 1802 Cosmac subset emulator, for genetic programming.

Requires https://github.com/strickyak/rye

Try this:

    python ../rye/rye.py run gene1.py

One small opcode change:
  0x00 is supposed to be IDL.  I made it NOP.
  0x68 is not supposed to be assigned.  I made it STOP (like IDL, but there's nothing to wait for).
