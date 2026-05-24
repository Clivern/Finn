#!/usr/bin/env python3

import time

HOST = "clivern.com"
DURATION_SECONDS = 60

for i in range(1, DURATION_SECONDS + 1):
    print(f"tick {i}/{DURATION_SECONDS} host={HOST}", flush=True)
    time.sleep(1)

print("Done.")
