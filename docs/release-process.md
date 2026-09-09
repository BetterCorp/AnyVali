# Release Process

## Publish a unified release

1. Branch from `master`, implement a focused fix, and open a PR with conventional commits and regression tests.
2. Wait for CI and CodeQL to pass and address review findings before merging.
3. After the merged commit passes checks, select the next semantic version and push a unified `v<version>` tag on that commit.
4. The **Build Release** workflow reads the version from the tag, publishes packages, builds CLI binaries, and creates the GitHub release. Verify every publishing job succeeds and update the release notes with the fixes and relevant compatibility changes.

For example, once the chosen commit is verified:

```bash
git tag v1.2.3 <verified-commit>
git push origin v1.2.3
```

Choose the version based on the public changes: `fix:` normally means a patch, `feat:` a minor, and breaking API changes a major. Commit scopes identify the SDK, for example `fix(go): isolate concurrent imports`. The active workflow uses one unified tag; historical SDK-specific tags and release-please configuration files do not drive the current release process.

## Distribution

| SDK | Distribution from unified `v*` tags |
|-----|------------------------------------|
| JavaScript | npm: `anyvali` |
| Python | PyPI: `anyvali` |
| Go | Repository module source; use a commit or matching module version |
| Rust | crates.io: `anyvali` |
| Java | Maven Central: `com.anyvali:anyvali` |
| Kotlin | Maven Central: `com.anyvali:anyvali-kotlin` |
| C# | NuGet: `AnyVali` |
| Ruby | RubyGems: `anyvali` |
| PHP | [Pinned Composer path installation](sdk-php.md#installation) from source |
| C++ | Repository source |
| CLI | GitHub release binaries for Linux, macOS, and Windows, with SHA-256 checksums |

The public PHP package `anyvali/anyvali` is unavailable on Packagist. There is no active Packagist publishing job or webhook configuration in this repository. Historical `php-v*` tags do not prove a Composer package was published. Use the unified source tag and an explicit Composer path version mapping; the old manifest `0.0.1` field and legacy `sdk/php/VERSION` file are not release identifiers.

The `php-source` job installs that tag's PHP SDK into an empty Composer consumer, with Packagist disabled, and verifies version mapping, platform requirements, autoloading, parsing, and interchange. GitHub release publication depends on this check. Local validation:

```bash
bash tools/release/smoke_php_install.sh sdk/php 1.2.3
```

The optional Docker build script produces a PHP source archive under `dist-release/packagist/` for review. This is not a registry upload.

## Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | Branch pushes, merge groups, manual dispatch | SDK tests and conformance; PHP consumer installation |
| `codeql.yml` | Configured code scanning events | Static security analysis |
| `build-release.yml` | Unified `v*` tag push, manual dispatch on a version tag | Registry publishing, PHP source validation, CLI builds and GitHub release |

Build Release requires a version tag as its ref; dispatching it on `master` fails version validation. Publishing credentials are supplied through GitHub Actions secrets and environment configuration. Inspect the current workflow for required names.

Published releases are immutable. Do not delete an existing release or move its tag to retry publication. Diagnose failed jobs first; retry only where the corresponding registry and job support it, otherwise prepare a new version. A successful CLI release upload alone does not establish that every registry publish succeeded.
