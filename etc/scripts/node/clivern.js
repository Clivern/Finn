#!/usr/bin/env node

const { spawnSync } = require("node:child_process");

const HOST = "clivern.com";
const DURATION_SECONDS = 60;

const result = spawnSync("ping", ["-c", String(DURATION_SECONDS), HOST], {
  stdio: "inherit",
});

if (result.error) {
  throw result.error;
}

process.exitCode = result.status ?? 1;
if (process.exitCode === 0) {
  console.log("Done.");
}
