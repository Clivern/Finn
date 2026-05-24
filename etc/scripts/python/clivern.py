#!/usr/bin/env python3

import subprocess


HOST = "clivern.com"
DURATION_SECONDS = 60

subprocess.run(["ping", "-c", str(DURATION_SECONDS), HOST], check=True)
print("Done.")
