#!/usr/bin/env python3

import time
from pathlib import Path

HOST = "clivern.com"
DURATION_SECONDS = 10

for i in range(1, DURATION_SECONDS + 1):
    print(f"tick {i}/{DURATION_SECONDS} host={HOST}", flush=True)
    time.sleep(1)

print(Path("hello.txt").read_text(encoding="utf-8").strip())
print("Done.")
