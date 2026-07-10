const assert = require("node:assert/strict");
const test = require("node:test");

const {
  PublicationPreflightError,
  ensureNpmVersionIsAvailable,
  expectedReleaseAssets,
  isDryRun,
  main,
  publishedPackageMetadata,
  runPublicationPreflight,
} = require("./publish");

const packageName = "@nk02/shine";
const version = "0.1.2";
const releaseRepo = "Nithish-Yenaganti/shine";

test("expectedReleaseAssets matches every platform supported by install.js", () => {
  assert.deepEqual(expectedReleaseAssets(version), [
    "shine_0.1.2_checksums.txt",
    "shine_0.1.2_darwin_amd64.tar.gz",
    "shine_0.1.2_darwin_arm64.tar.gz",
    "shine_0.1.2_linux_amd64.tar.gz",
    "shine_0.1.2_linux_arm64.tar.gz",
  ]);
});

test("isDryRun recognizes npm arguments and configuration", () => {
  assert.equal(isDryRun(["--dry-run"], {}), true);
  assert.equal(isDryRun(["--dry-run=true"], {}), true);
  assert.equal(isDryRun([], { npm_config_dry_run: "true" }), true);
  assert.equal(isDryRun(["--dry-run=false"], { npm_config_dry_run: "true" }), false);
  assert.equal(isDryRun(["--access", "public"], {}), false);
});

test("main skips remote preflight for dry runs and still invokes npm publishing", async () => {
  const events = [];
  const args = ["--dry-run", "--access", "public"];

  const status = await main({
    args,
    env: {},
    log: (message) => events.push(["log", message]),
    preflight: async () => events.push(["preflight"]),
    publish: (publishArgs) => {
      events.push(["publish", publishArgs]);
      return 0;
    },
  });

  assert.equal(status, 0);
  assert.deepEqual(events, [
    ["log", "Skipping remote publication preflight for npm --dry-run."],
    ["publish", args],
  ]);
});

test("main completes preflight before a real publish", async () => {
  const events = [];

  const status = await main({
    args: ["--access", "public"],
    env: { GITHUB_TOKEN: "test-token" },
    packageJson: { name: packageName, version },
    log: () => {},
    preflight: async (options) => {
      events.push("preflight");
      assert.equal(options.packageName, packageName);
      assert.equal(options.version, version);
      assert.equal(options.githubToken, "test-token");
    },
    publish: () => {
      events.push("publish");
      return 0;
    },
  });

  assert.equal(status, 0);
  assert.deepEqual(events, ["preflight", "publish"]);
});

test("published package metadata excludes maintainer-only scripts", () => {
  const metadata = publishedPackageMetadata({
    name: packageName,
    version,
    scripts: {
      postinstall: "node npm/install.js",
      "publish:npm": "node npm/publish.js",
      "test:npm": "SHINE_BINARY_PATH=bin/shine node npm/shine.js version",
      "test:publish": "node --test npm/publish.test.js",
    },
  });

  assert.deepEqual(metadata.scripts, {
    postinstall: "node npm/install.js",
  });
});

test("publication preflight accepts an unpublished version with a complete release", async () => {
  const requests = [];
  const assets = expectedReleaseAssets(version).map((name) => ({ name }));

  await runPublicationPreflight({
    packageName,
    version,
    releaseRepo,
    githubToken: "test-token",
    request: async (url, options) => {
      requests.push({ url, options });
      if (requests.length === 1) {
        return { statusCode: 404, body: "{}" };
      }
      return { statusCode: 200, body: JSON.stringify({ assets }) };
    },
  });

  assert.equal(requests.length, 2);
  assert.equal(requests[0].url, "https://registry.npmjs.org/%40nk02%2Fshine/0.1.2");
  assert.equal(
    requests[1].url,
    "https://api.github.com/repos/Nithish-Yenaganti/shine/releases/tags/v0.1.2",
  );
  assert.equal(requests[1].options.headers.authorization, "Bearer test-token");
});

test("publication preflight rejects a version that already exists on npm", async () => {
  let requestCount = 0;

  await assert.rejects(
    runPublicationPreflight({
      packageName,
      version,
      releaseRepo,
      request: async () => {
        requestCount += 1;
        return { statusCode: 200, body: "{}" };
      },
    }),
    (error) => {
      assert.ok(error instanceof PublicationPreflightError);
      assert.match(error.message, /@nk02\/shine@0\.1\.2 is already published on npm/);
      assert.match(error.message, /Update package\.json to a new version/);
      return true;
    },
  );

  assert.equal(requestCount, 1, "GitHub should not be queried after the npm check fails");
});

test("publication preflight lists every missing GitHub release asset", async () => {
  const availableAsset = "shine_0.1.2_darwin_amd64.tar.gz";

  await assert.rejects(
    runPublicationPreflight({
      packageName,
      version,
      releaseRepo,
      request: async (url) => {
        if (url.includes("registry.npmjs.org")) {
          return { statusCode: 404, body: "{}" };
        }
        return {
          statusCode: 200,
          body: JSON.stringify({ assets: [{ name: availableAsset }] }),
        };
      },
    }),
    (error) => {
      assert.ok(error instanceof PublicationPreflightError);
      assert.match(error.message, /missing required npm installer assets/);
      assert.match(error.message, /shine_0\.1\.2_checksums\.txt/);
      assert.match(error.message, /shine_0\.1\.2_darwin_arm64\.tar\.gz/);
      assert.match(error.message, /shine_0\.1\.2_linux_amd64\.tar\.gz/);
      assert.match(error.message, /shine_0\.1\.2_linux_arm64\.tar\.gz/);
      assert.doesNotMatch(error.message, new RegExp(availableAsset.replaceAll(".", "\\.")));
      assert.match(error.message, /Upload the complete GoReleaser output/);
      return true;
    },
  );
});

test("publication preflight explains how to recover from a missing GitHub release", async () => {
  await assert.rejects(
    runPublicationPreflight({
      packageName,
      version,
      releaseRepo,
      request: async () => ({ statusCode: 404, body: "{}" }),
    }),
    (error) => {
      assert.match(error.message, /GitHub release v0\.1\.2 was not found/);
      assert.match(error.message, /Create and publish the release with GoReleaser/);
      return true;
    },
  );
});

test("publication preflight rejects a draft GitHub release", async () => {
  await assert.rejects(
    runPublicationPreflight({
      packageName,
      version,
      releaseRepo,
      request: async (url) => {
        if (url.includes("registry.npmjs.org")) {
          return { statusCode: 404, body: "{}" };
        }
        return {
          statusCode: 200,
          body: JSON.stringify({ draft: true, assets: expectedReleaseAssets(version) }),
        };
      },
    }),
    (error) => {
      assert.match(error.message, /is still a draft/);
      assert.match(error.message, /Publish the GitHub release/);
      return true;
    },
  );
});

test("publication preflight reports network failures before npm publish starts", async () => {
  await assert.rejects(
    ensureNpmVersionIsAvailable({
      packageName,
      version,
      request: async () => {
        throw new Error("socket unavailable");
      },
    }),
    (error) => {
      assert.match(error.message, /socket unavailable/);
      assert.match(error.message, /npm publish was not started/);
      return true;
    },
  );
});
