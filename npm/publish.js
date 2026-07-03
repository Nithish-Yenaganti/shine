#!/usr/bin/env node

const childProcess = require("node:child_process");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

const root = path.resolve(__dirname, "..");
const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "shine-npm-publish-"));

try {
  copy("package.json", "package.json");
  copy("LICENSE", "LICENSE");
  copy("CHANGELOG.md", "CHANGELOG.md");
  copy("npm/README.md", "README.md");
  copy("npm/install.js", "npm/install.js");
  copy("npm/shine.js", "npm/shine.js");

  const args = ["publish", tempDir, ...process.argv.slice(2)];
  const result = childProcess.spawnSync("npm", args, { stdio: "inherit" });
  if (result.error) {
    throw result.error;
  }
  process.exit(result.status ?? 0);
} finally {
  fs.rmSync(tempDir, { recursive: true, force: true });
}

function copy(from, to) {
  const source = path.join(root, from);
  const destination = path.join(tempDir, to);
  fs.mkdirSync(path.dirname(destination), { recursive: true });
  fs.copyFileSync(source, destination);
}
