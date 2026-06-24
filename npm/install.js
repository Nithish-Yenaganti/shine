#!/usr/bin/env node

const childProcess = require("node:child_process");
const crypto = require("node:crypto");
const fs = require("node:fs");
const https = require("node:https");
const os = require("node:os");
const path = require("node:path");

const packageJson = require("../package.json");

const repo = process.env.SHINE_RELEASE_REPO || "Nithish-Yenaganti/shine";
const version = process.env.SHINE_RELEASE_VERSION || packageJson.version;
const tag = process.env.SHINE_RELEASE_TAG || `v${version}`;
const binaryName = process.platform === "win32" ? "shine.exe" : "shine";
const installDir = path.join(__dirname, "bin");
const installPath = path.join(installDir, binaryName);

main().catch((error) => {
  console.error(`shine install failed: ${error.message}`);
  process.exit(1);
});

async function main() {
  if (truthy(process.env.SHINE_SKIP_DOWNLOAD)) {
    console.log("Skipping shine binary download because SHINE_SKIP_DOWNLOAD is set.");
    return;
  }

  if (process.env.SHINE_BINARY_PATH) {
    copyLocalBinary(process.env.SHINE_BINARY_PATH);
    return;
  }

  const target = releaseTarget();
  const archiveName = `shine_${version}_${target.goos}_${target.goarch}.tar.gz`;
  const baseUrl = `https://github.com/${repo}/releases/download/${tag}`;
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "shine-npm-"));
  const archivePath = path.join(tmpDir, archiveName);
  const checksumsPath = path.join(tmpDir, "checksums.txt");

  try {
    await download(`${baseUrl}/${archiveName}`, archivePath);
    if (!truthy(process.env.SHINE_SKIP_CHECKSUM)) {
      await download(`${baseUrl}/shine_${version}_checksums.txt`, checksumsPath);
      verifyChecksum(archivePath, archiveName, checksumsPath);
    }
    extractArchive(archivePath, tmpDir);
    copyLocalBinary(path.join(tmpDir, binaryName));
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

function releaseTarget() {
  const osMap = {
    darwin: "darwin",
    linux: "linux",
  };
  const archMap = {
    x64: "amd64",
    arm64: "arm64",
  };

  const goos = osMap[process.platform];
  const goarch = archMap[process.arch];
  if (!goos || !goarch) {
    throw new Error(`unsupported platform: ${process.platform}/${process.arch}`);
  }
  return { goos, goarch };
}

function copyLocalBinary(source) {
  const resolved = path.resolve(source);
  if (!fs.existsSync(resolved)) {
    throw new Error(`binary not found: ${resolved}`);
  }
  fs.mkdirSync(installDir, { recursive: true });
  fs.copyFileSync(resolved, installPath);
  fs.chmodSync(installPath, 0o755);
  console.log(`Installed shine binary to ${installPath}`);
}

function download(url, destination) {
  return new Promise((resolve, reject) => {
    const request = https.get(url, (response) => {
      if ([301, 302, 303, 307, 308].includes(response.statusCode)) {
        response.resume();
        download(response.headers.location, destination).then(resolve, reject);
        return;
      }

      if (response.statusCode !== 200) {
        response.resume();
        reject(new Error(`download failed: ${url} returned ${response.statusCode}`));
        return;
      }

      const file = fs.createWriteStream(destination);
      response.pipe(file);
      file.on("finish", () => file.close(resolve));
      file.on("error", reject);
    });
    request.on("error", reject);
  });
}

function verifyChecksum(archivePath, archiveName, checksumsPath) {
  const expected = fs
    .readFileSync(checksumsPath, "utf8")
    .split(/\r?\n/)
    .map((line) => line.trim().split(/\s+/))
    .find((parts) => parts[1] === archiveName)?.[0];

  if (!expected) {
    throw new Error(`checksum not found for ${archiveName}`);
  }

  const actual = crypto.createHash("sha256").update(fs.readFileSync(archivePath)).digest("hex");
  if (actual !== expected) {
    throw new Error(`checksum mismatch for ${archiveName}`);
  }
}

function extractArchive(archivePath, destination) {
  const result = childProcess.spawnSync("tar", ["-xzf", archivePath, "-C", destination], {
    stdio: "inherit",
  });
  if (result.error) {
    throw result.error;
  }
  if (result.status !== 0) {
    throw new Error(`tar exited with status ${result.status}`);
  }
}

function truthy(value) {
  return value === "1" || value === "true" || value === "yes";
}
