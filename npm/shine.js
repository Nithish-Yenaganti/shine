#!/usr/bin/env node

const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

function binaryPath() {
  if (process.env.SHINE_BINARY_PATH) {
    return path.resolve(process.env.SHINE_BINARY_PATH);
  }
  return path.join(__dirname, "bin", process.platform === "win32" ? "shine.exe" : "shine");
}

const bin = binaryPath();
if (!fs.existsSync(bin)) {
  console.error("shine binary was not found.");
  console.error("Run `npm rebuild @nithish-yenaganti/shine` or reinstall the package.");
  console.error("For local development, set SHINE_BINARY_PATH=bin/shine.");
  process.exit(1);
}

const result = spawnSync(bin, process.argv.slice(2), { stdio: "inherit" });

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}

if (result.signal) {
  process.kill(process.pid, result.signal);
}

process.exit(result.status ?? 0);
