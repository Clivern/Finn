#!/usr/bin/env node

const { setTimeout: sleep } = require("node:timers/promises");

const HOST = "clivern.com";
const DURATION_SECONDS = 60;

(async () => {
  for (let i = 1; i <= DURATION_SECONDS; i++) {
    console.log(`tick ${i}/${DURATION_SECONDS} host=${HOST}`);
    await sleep(1000);
  }
  console.log("Done.");
})().catch((err) => {
  console.error(err);
  process.exit(1);
});
