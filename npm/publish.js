#!/usr/bin/env node

const childProcess = require("node:child_process");
const fs = require("node:fs");
const https = require("node:https");
const os = require("node:os");
const path = require("node:path");

const packageJson = require("../package.json");

const root = path.resolve(__dirname, "..");
const npmRegistry = "https://registry.npmjs.org";
const githubApi = "https://api.github.com";
const releaseRepo = "Nithish-Yenaganti/shine";
const releaseTargets = [
  ["darwin", "amd64"],
  ["darwin", "arm64"],
  ["linux", "amd64"],
  ["linux", "arm64"],
];
const packageFiles = [
  ["LICENSE", "LICENSE"],
  ["CHANGELOG.md", "CHANGELOG.md"],
  ["npm/README.md", "README.md"],
  ["npm/install.js", "npm/install.js"],
  ["npm/shine.js", "npm/shine.js"],
];

class PublicationPreflightError extends Error {
  constructor(message) {
    super(message);
    this.name = "PublicationPreflightError";
  }
}

async function main(options = {}) {
  const args = options.args ?? process.argv.slice(2);
  const env = options.env ?? process.env;
  const metadata = options.packageJson ?? packageJson;
  const log = options.log ?? console.log;
  const preflight = options.preflight ?? runPublicationPreflight;
  const publish = options.publish ?? publishPackage;

  if (isDryRun(args, env)) {
    log("Skipping remote publication preflight for npm --dry-run.");
  } else {
    await preflight({
      packageName: metadata.name,
      version: metadata.version,
      releaseRepo: options.releaseRepo ?? releaseRepo,
      npmRegistry: options.npmRegistry ?? npmRegistry,
      githubApi: options.githubApi ?? githubApi,
      githubToken: options.githubToken ?? env.GITHUB_TOKEN ?? env.GH_TOKEN,
      request: options.request ?? requestText,
    });
    log(`Publication preflight passed for ${metadata.name}@${metadata.version}.`);
  }

  return publish(args);
}

async function runPublicationPreflight(options) {
  await ensureNpmVersionIsAvailable(options);
  await ensureGitHubReleaseIsComplete(options);
}

async function ensureNpmVersionIsAvailable({
  packageName,
  version,
  npmRegistry: registry = npmRegistry,
  request = requestText,
}) {
  const url = `${stripTrailingSlash(registry)}/${encodeURIComponent(packageName)}/${encodeURIComponent(version)}`;
  const response = await preflightRequest(
    request,
    url,
    {
      headers: {
        accept: "application/json",
        "user-agent": "shine-npm-publisher",
      },
    },
    `npm registry lookup for ${packageName}@${version}`,
  );

  if (response.statusCode === 404) {
    return;
  }
  if (response.statusCode >= 200 && response.statusCode < 300) {
    throw new PublicationPreflightError(
      `${packageName}@${version} is already published on npm. ` +
        "Update package.json to a new version and create its matching GitHub release before retrying.",
    );
  }

  throw new PublicationPreflightError(
    `Could not verify ${packageName}@${version} on npm: the registry returned HTTP ${response.statusCode}. ` +
      "Check npm registry access and retry; npm publish was not started.",
  );
}

async function ensureGitHubReleaseIsComplete({
  version,
  releaseRepo: repo = releaseRepo,
  githubApi: api = githubApi,
  githubToken,
  request = requestText,
}) {
  const tag = `v${version}`;
  const repoPath = repo.split("/").map(encodeURIComponent).join("/");
  const url = `${stripTrailingSlash(api)}/repos/${repoPath}/releases/tags/${encodeURIComponent(tag)}`;
  const headers = {
    accept: "application/vnd.github+json",
    "user-agent": "shine-npm-publisher",
    "x-github-api-version": "2022-11-28",
  };
  if (githubToken) {
    headers.authorization = `Bearer ${githubToken}`;
  }

  const response = await preflightRequest(
    request,
    url,
    { headers },
    `GitHub release lookup for ${repo}@${tag}`,
  );

  if (response.statusCode === 404) {
    throw new PublicationPreflightError(
      `GitHub release ${tag} was not found in ${repo}. ` +
        "Create and publish the release with GoReleaser before retrying npm publish.",
    );
  }
  if (response.statusCode < 200 || response.statusCode >= 300) {
    const authenticationHint = response.statusCode === 401 || response.statusCode === 403
      ? " Check GitHub API access or set GITHUB_TOKEN, then retry;"
      : " Check GitHub API access and retry;";
    throw new PublicationPreflightError(
      `Could not inspect GitHub release ${repo}@${tag}: the API returned HTTP ${response.statusCode}.` +
        `${authenticationHint} npm publish was not started.`,
    );
  }

  let release;
  try {
    release = JSON.parse(response.body);
  } catch (error) {
    throw new PublicationPreflightError(
      `Could not inspect GitHub release ${repo}@${tag}: the API returned invalid JSON. ` +
        "Retry the request; npm publish was not started.",
    );
  }

  if (release.draft) {
    throw new PublicationPreflightError(
      `GitHub release ${repo}@${tag} is still a draft. ` +
        "Publish the GitHub release before retrying npm publish.",
    );
  }

  if (!Array.isArray(release.assets)) {
    throw new PublicationPreflightError(
      `Could not inspect GitHub release ${repo}@${tag}: the API response did not contain an asset list. ` +
        "Retry the request; npm publish was not started.",
    );
  }

  const availableAssets = new Set(release.assets.map((asset) => asset && asset.name));
  const missingAssets = expectedReleaseAssets(version).filter((name) => !availableAssets.has(name));
  if (missingAssets.length > 0) {
    throw new PublicationPreflightError(
      `GitHub release ${repo}@${tag} is missing required npm installer assets:\n` +
        `${missingAssets.map((name) => `  - ${name}`).join("\n")}\n` +
        "Upload the complete GoReleaser output before retrying npm publish.",
    );
  }
}

function expectedReleaseAssets(version) {
  return [
    `shine_${version}_checksums.txt`,
    ...releaseTargets.map(([goos, goarch]) => `shine_${version}_${goos}_${goarch}.tar.gz`),
  ];
}

function isDryRun(args, env = {}) {
  let dryRun = truthy(env.npm_config_dry_run);
  for (const arg of args) {
    if (arg === "--dry-run") {
      dryRun = true;
    } else if (arg.startsWith("--dry-run=")) {
      dryRun = truthy(arg.slice("--dry-run=".length));
    }
  }
  return dryRun;
}

function publishPackage(args) {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "shine-npm-publish-"));

  try {
    fs.writeFileSync(
      path.join(tempDir, "package.json"),
      `${JSON.stringify(publishedPackageMetadata(), null, 2)}\n`,
    );
    for (const [from, to] of packageFiles) {
      copy(from, to, tempDir);
    }

    const result = childProcess.spawnSync("npm", ["publish", tempDir, ...args], { stdio: "inherit" });
    if (result.error) {
      throw result.error;
    }
    return result.status ?? 0;
  } finally {
    fs.rmSync(tempDir, { recursive: true, force: true });
  }
}

function publishedPackageMetadata(metadata = packageJson) {
  return {
    ...metadata,
    scripts: {
      postinstall: metadata.scripts.postinstall,
    },
  };
}

function copy(from, to, tempDir) {
  const source = path.join(root, from);
  const destination = path.join(tempDir, to);
  fs.mkdirSync(path.dirname(destination), { recursive: true });
  fs.copyFileSync(source, destination);
}

async function preflightRequest(request, url, options, description) {
  try {
    return await request(url, options);
  } catch (error) {
    throw new PublicationPreflightError(
      `Could not complete the ${description}: ${error.message}. ` +
        "Check network access and retry; npm publish was not started.",
    );
  }
}

function requestText(url, options = {}, redirectsRemaining = 5) {
  return new Promise((resolve, reject) => {
    const request = https.get(url, { headers: options.headers }, (response) => {
      if ([301, 302, 303, 307, 308].includes(response.statusCode) && response.headers.location) {
        response.resume();
        if (redirectsRemaining === 0) {
          reject(new Error(`too many redirects while requesting ${url}`));
          return;
        }
        const redirectUrl = new URL(response.headers.location, url).toString();
        requestText(redirectUrl, options, redirectsRemaining - 1).then(resolve, reject);
        return;
      }

      response.setEncoding("utf8");
      let body = "";
      response.on("data", (chunk) => {
        body += chunk;
      });
      response.on("end", () => {
        resolve({ statusCode: response.statusCode, body });
      });
      response.on("error", reject);
    });

    request.setTimeout(options.timeoutMs ?? 10_000, () => {
      request.destroy(new Error(`request timed out while requesting ${url}`));
    });
    request.on("error", reject);
  });
}

function stripTrailingSlash(value) {
  return value.replace(/\/+$/, "");
}

function truthy(value) {
  return value === true || ["1", "true", "yes"].includes(String(value).toLowerCase());
}

if (require.main === module) {
  main().then(
    (status) => {
      process.exitCode = status;
    },
    (error) => {
      console.error(`shine npm publish failed: ${error.message}`);
      process.exitCode = 1;
    },
  );
}

module.exports = {
  PublicationPreflightError,
  ensureGitHubReleaseIsComplete,
  ensureNpmVersionIsAvailable,
  expectedReleaseAssets,
  isDryRun,
  main,
  publishedPackageMetadata,
  requestText,
  runPublicationPreflight,
};
