# Release Process

This document describes how code gets from a developer's machine to published packages.

## Flow

```
branch (develop/feature/hotfix)
  │
  └─ PR to master
       │
       ├─ CI runs all SDK tests
       │
       └─ Merge to master
            │
            ├─ CI runs again (verifies merge)
            │
            └─ CI passes → Build Release triggers
                 │
                 ├─ Version bump + changelog (committed directly, skips CI)
                 ├─ GitHub Releases + tags created
                 └─ Publish workflows dispatched to registries
```

1. **Create a branch** off `master` (`develop`, `feature/*`, `hotfix/*`, or any branch)
2. **Write code** using [conventional commits](#commit-messages)
3. **Open a PR** targeting `master` — CI runs all SDK tests
4. **Merge the PR** — CI runs again on master to verify the merge
5. **CI passes** — Build Release triggers automatically:
   - Calculates version bumps from conventional commits
   - Generates changelogs
   - Commits version/changelog updates directly to master (this commit skips CI)
   - Creates GitHub Releases with tags
   - Dispatches publish workflows to package registries

Hotfixes follow the same flow — branch off `master`, PR back to `master`.

## Commit Messages

Version bumps are determined by [Conventional Commits](https://www.conventionalcommits.org/):

| Prefix | Version bump | Example |
|--------|-------------|---------|
| `fix:` | Patch (0.0.x) | `fix(js): handle null input in parser` |
| `feat:` | Minor (0.x.0) | `feat(python): add date-time coercion` |
| `feat!:` or `BREAKING CHANGE:` | Major (x.0.0) | `feat!: rename SafeParse to Validate` |
| `chore:`, `docs:`, `test:`, `ci:` | No bump | `docs: update CLI examples` |

### Scoping

Use the component name in parentheses to scope a commit to a specific SDK:

```
fix(js): handle null input in parser       → bumps @anyvali/js only
feat(python): add date-time coercion       → bumps anyvali (Python) only
fix(cli): treat bare "-" as stdin arg       → bumps CLI only
fix: update all SDK releases                → bumps root package + touched paths
```

Valid scopes: `js`, `python`, `go`, `java`, `rust`, `php`, `ruby`, `kotlin`, `csharp`, `cpp`, `cli`

## What Gets Published

| Tag | Registry | Package |
|-----|----------|---------|
| `js-v*` | npm | `@anyvali/js` |
| `python-v*` | PyPI | `anyvali` |
| `go-v*` | Go modules | `github.com/BetterCorp/AnyVali/sdk/go` |
| `rust-v*` | crates.io | `anyvali` |
| `java-v*` | Maven Central | `com.anyvali:anyvali` |
| `kotlin-v*` | Maven Central | `com.anyvali:anyvali-kotlin` |
| `csharp-v*` | NuGet | `AnyVali` |
| `ruby-v*` | RubyGems | `anyvali` |
| `php-v*` | Packagist | `anyvali/anyvali` (via webhook) |
| `cpp-v*` | — | Source release only |
| `cli-v*` | GitHub Releases | Binary downloads (linux/mac/windows) |

## Independent Versioning

Each SDK versions independently. A `fix(js):` commit only bumps the JS SDK version. Other SDKs stay at their current version.

The root package (`anyvali-v*`) tracks cross-cutting changes that don't belong to a single SDK.

## Automation Details

### Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | Push to master/develop, PRs to master | Run tests for all SDKs |
| `build-release.yml` | After CI passes on master (`workflow_run`) | Version bump, changelog, create releases, publish |
| `release-*.yml` | Called by build-release (`workflow_call`) | Publish to individual registries |

### How Build Release Works

The build-release workflow uses [Release Please](https://github.com/googleapis/release-please) under the hood and triggers automatically after CI passes on master:

1. **Your PR merges to master** — CI runs and passes.
2. **Build Release triggers** — Release Please reads conventional commits, creates a release PR with version bumps and changelog updates, and auto-merges it immediately.
3. **The release PR merge pushes to master** — CI detects the `chore: release` commit message and **skips** (no redundant test run).
4. **Build Release triggers again** (from the push) — Release Please detects the merged release PR, creates GitHub Releases with tags, and dispatches the relevant publish workflows.

From a developer's perspective: merge your PR, packages appear on registries. No manual steps.

### Why CI Skips Release Commits

When Build Release auto-merges its version bump PR, the commit message starts with `chore: release`. CI checks for this prefix and skips all test jobs to avoid a redundant build cycle. Build Release itself is not affected by this skip — it triggers independently via `workflow_run`.

### Configuration Files

| File | Purpose |
|------|---------|
| `.release-please-config.json` | Package definitions, release types, changelog paths |
| `.release-please-manifest.json` | Current version for each package (updated automatically) |

## Hotfixes

Hotfixes follow the exact same flow:

```bash
git checkout -b hotfix/critical-bug master
# fix the issue
git commit -m "fix(go): prevent panic on nil schema input"
git push origin hotfix/critical-bug
# open PR to master, get CI green, merge
# → auto version bump → auto release → auto publish
```

## Manual Publishing

Each publish workflow can be triggered manually from the GitHub Actions tab:

1. Go to **Actions** → select the publish workflow (e.g., "Publish npm")
2. Click **Run workflow** → select `master` branch → click **Run**

This is useful for retrying a failed publish without creating a new release.

## Branch Protection

Recommended settings for `master`:

- **Require pull request before merging** — no direct pushes
- **Require status checks to pass** — CI must be green
- **Require linear history** — optional, keeps history clean
- **Allow auto-merge** — required for the release PR auto-merge to work

## Secrets Required

| Secret | Used by | Purpose |
|--------|---------|---------|
| `NPM2_TOKEN_PUB` | release-npm | npm publish token |
| `CODECOV_TOKEN` | CI | Coverage upload |
| `CRATES_IO_TOKEN` | release-crates | crates.io API token |
| `MAVEN_USERNAME` | release-maven, release-kotlin | Sonatype Central username |
| `MAVEN_PASSWORD` | release-maven, release-kotlin | Sonatype Central password |
| `MAVEN_GPG_PRIVATE_KEY` | release-maven, release-kotlin | GPG key for artifact signing |
| `MAVEN_GPG_PASSPHRASE` | release-maven, release-kotlin | GPG key passphrase |
| `NUGET_API_KEY` | release-nuget | NuGet API key |
| `RUBYGEMS_API_KEY` | release-rubygems | RubyGems API key |
