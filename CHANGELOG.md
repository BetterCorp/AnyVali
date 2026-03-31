# Changelog

## [0.3.1](https://github.com/BetterCorp/AnyVali/compare/anyvali-v0.3.0...anyvali-v0.3.1) (2026-03-31)


### Bug Fixes

* always auto-merge release PR when one exists ([70c15a7](https://github.com/BetterCorp/AnyVali/commit/70c15a7b67d5ead20f12378ce89bb9ab5ce7e25f))
* enable auto-merge on release PR after CI passes ([7bf170c](https://github.com/BetterCorp/AnyVali/commit/7bf170cb4df8328338d8be61e15147d78e49ff34))
* lock release PR immediately, close on merge failure ([8f97a10](https://github.com/BetterCorp/AnyVali/commit/8f97a10abe927f57fd8b9a9404412d1a1d65eb8f))
* resolve merge conflict, keep simplified build-release ([7807802](https://github.com/BetterCorp/AnyVali/commit/780780259da32288e030bb046571e4f37fd69ec6))
* restructure CI and release flow ([880eb22](https://github.com/BetterCorp/AnyVali/commit/880eb220318c1123137a1510d3349db71b772c5e))
* simplify build-release, remove auto-merge ([02ac7d8](https://github.com/BetterCorp/AnyVali/commit/02ac7d8aeb269caae85decbcd1ccad8b8419ed4e))
* use env vars for commit message in CI skip check ([331dc94](https://github.com/BetterCorp/AnyVali/commit/331dc946aae47af1e2507cb140a18fe1903ceba4))

## [0.3.0](https://github.com/BetterCorp/AnyVali/compare/anyvali-v0.2.2...anyvali-v0.3.0) (2026-03-30)


### Features

* add SchemaAny type alias to Python, Kotlin, Rust, and C++ ([a5222a4](https://github.com/BetterCorp/AnyVali/commit/a5222a48c7bb40a56f7c8488a6876bb3fae1c52c))
* **js:** add SchemaAny type alias for generic constraints ([6c1dbe5](https://github.com/BetterCorp/AnyVali/commit/6c1dbe5e4ed1ddfc71e92657ab44a1177857d6ac))

## [0.2.2](https://github.com/BetterCorp/AnyVali/compare/anyvali-v0.2.1...anyvali-v0.2.2) (2026-03-30)


### Bug Fixes

* **ci:** only publish packages that have new releases ([fcdd045](https://github.com/BetterCorp/AnyVali/commit/fcdd04579b729750accf154d15ec2f68c9bcd258))

## [0.2.1](https://github.com/BetterCorp/AnyVali/compare/anyvali-v0.2.0...anyvali-v0.2.1) (2026-03-30)


### Bug Fixes

* auto-merge release PRs and document release process ([ed464a8](https://github.com/BetterCorp/AnyVali/commit/ed464a8091329860de6949b0dbf7eb941046b4ce))
* **kotlin,ruby:** fix remaining CI failures from type inference ([0677331](https://github.com/BetterCorp/AnyVali/commit/0677331e61b728081b141753485d1861ef20d62f))
* use GitHub API for auto-merge instead of gh CLI ([0486d85](https://github.com/BetterCorp/AnyVali/commit/0486d85f1798e088353744cdab6fde827bda4723))

## [0.2.0](https://github.com/BetterCorp/AnyVali/compare/anyvali-v0.1.8...anyvali-v0.2.0) (2026-03-30)


### Features

* add static type inference to all 10 SDKs ([f2902ec](https://github.com/BetterCorp/AnyVali/commit/f2902ecf9cdd1153da355b8e27fae07c36486138))


### Bug Fixes

* ensure CI runs on PRs targeting master ([2b950c2](https://github.com/BetterCorp/AnyVali/commit/2b950c2e7ed632834be0689ee67c7614605b6750))
* **java,kotlin:** handle numeric type conversion in typed parse ([6273111](https://github.com/BetterCorp/AnyVali/commit/6273111c893296cb12c90d4096e7329403dcac6b))
* **kotlin:** use gradle instead of ./gradlew (no wrapper in repo) ([32e41fa](https://github.com/BetterCorp/AnyVali/commit/32e41fa287f49d5a8d0b5ebc8111a25999467bb4))

## [0.1.8](https://github.com/BetterCorp/AnyVali/compare/anyvali-v0.1.7...anyvali-v0.1.8) (2026-03-30)


### Bug Fixes

* add id-token permission for PyPI trusted publishing ([8d88fb0](https://github.com/BetterCorp/AnyVali/commit/8d88fb0bb80debfd1310e312ef4a1e4ae611c6e3))
* add workflow_call trigger to all release workflows ([f465315](https://github.com/BetterCorp/AnyVali/commit/f465315858c6468fe863ebd584c2a679ee1f545f))
* **cli:** fetch tags in upload job for GitHub Release ([ceed6e5](https://github.com/BetterCorp/AnyVali/commit/ceed6e5ed813e573b65cc502cb5cee44a62e1dbf))
* dispatch all publish workflows on any release ([0ad5655](https://github.com/BetterCorp/AnyVali/commit/0ad56553d5e0a0ab5f5fa11a5d9bb943e49bcb68))
* force version bump across all packages ([959aaee](https://github.com/BetterCorp/AnyVali/commit/959aaeee8617156d7b531fa3fbfb588d99081ea7))
* **java:** use central-publishing-maven-plugin for Sonatype Central ([987a6e2](https://github.com/BetterCorp/AnyVali/commit/987a6e232ea07f66463af040235e1efd4a7eb7aa))
* **js:** update repository URL to BetterCorp/AnyVali ([3dae779](https://github.com/BetterCorp/AnyVali/commit/3dae779cd07c1aeb307ea9b181542073f6dba8bc))
* **kotlin:** add Maven Central publishing via vanniktech plugin ([987a6e2](https://github.com/BetterCorp/AnyVali/commit/987a6e232ea07f66463af040235e1efd4a7eb7aa))
* **php:** remove split-repo workflow, Packagist pulls from monorepo directly ([94e4079](https://github.com/BetterCorp/AnyVali/commit/94e4079b1d0a3e203ab895b8a149aa609a550ddd))

## [0.1.7](https://github.com/BetterCorp/AnyVali/compare/anyvali-v0.1.6...anyvali-v0.1.7) (2026-03-30)


### Bug Fixes

* use workflow_dispatch for all release workflows ([801f098](https://github.com/BetterCorp/AnyVali/commit/801f098f093d413f8a2d5726d498c5fcd507b3a3))

## [0.1.6](https://github.com/BetterCorp/AnyVali/compare/anyvali-v0.1.5...anyvali-v0.1.6) (2026-03-30)


### Bug Fixes

* use on:release trigger instead of on:push:tags ([62b76a5](https://github.com/BetterCorp/AnyVali/commit/62b76a59fa9aa97a6a2d0a6eb61ec832e4e5d9a0))

## [0.1.5](https://github.com/BetterCorp/AnyVali/compare/anyvali-v0.1.4...anyvali-v0.1.5) (2026-03-30)


### Bug Fixes

* **cli:** add CLI to release-please and build workflow ([2682c44](https://github.com/BetterCorp/AnyVali/commit/2682c44ca71a8488271457722c952ee3ca3956e9))
* match release workflow tags to release-please component tags ([af5d15b](https://github.com/BetterCorp/AnyVali/commit/af5d15b540e37437f731c6ff8110b8dca5954d8e))
* resolve merge conflict in release-please manifest ([5c8098c](https://github.com/BetterCorp/AnyVali/commit/5c8098ce9ee5ab0891a830c4fa82f7db20e5d9cf))

## [0.1.4](https://github.com/BetterCorp/AnyVali/compare/anyvali-v0.1.3...anyvali-v0.1.4) (2026-03-30)


### Bug Fixes

* synchronize all SDK releases ([904dda4](https://github.com/BetterCorp/AnyVali/commit/904dda427fe45aeb2321ceaf32f71683d71e96d2))

## [0.1.3](https://github.com/BetterCorp/AnyVali/compare/anyvali-v0.1.2...anyvali-v0.1.3) (2026-03-30)


### Bug Fixes

* **java:** change release-type from java to simple ([3e10859](https://github.com/BetterCorp/AnyVali/commit/3e10859e02c4be6f12663ed2594971a3a058d742))

## [0.1.2](https://github.com/BetterCorp/AnyVali/compare/anyvali-v0.1.1...anyvali-v0.1.2) (2026-03-30)


### Bug Fixes

* add scope to release-please title pattern ([43e6b88](https://github.com/BetterCorp/AnyVali/commit/43e6b88c947bf41145a7ea3d2a4ebaa4a640d6b8))
* **java:** bootstrap initial release changelog ([c9bd624](https://github.com/BetterCorp/AnyVali/commit/c9bd624231b5ba752bf4e1b2a99771689a70aa60))
* **python:** bootstrap initial release changelog ([00b3484](https://github.com/BetterCorp/AnyVali/commit/00b34849d1c49d38ac9798c2d5f9aedbd38026a3))
* **ruby:** bootstrap initial release changelog ([282910d](https://github.com/BetterCorp/AnyVali/commit/282910d85f5326597ed49e6e55cb7a8018c57a62))

## [0.1.1](https://github.com/BetterCorp/AnyVali/compare/anyvali-v0.1.0...anyvali-v0.1.1) (2026-03-30)


### Bug Fixes

* **java:** sync release-please manifest version ([afdedce](https://github.com/BetterCorp/AnyVali/commit/afdedcee2aed72751a7f61c14db7e057104e7b2f))
* restore @anyvali/js scoped package name ([6875213](https://github.com/BetterCorp/AnyVali/commit/68752137d9de7079654f04fec7a1520601b06deb))
* **ruby:** sync release-please manifest version ([afdedce](https://github.com/BetterCorp/AnyVali/commit/afdedcee2aed72751a7f61c14db7e057104e7b2f))

## [0.1.0](https://github.com/BetterCorp/AnyVali/compare/anyvali-v0.0.1...anyvali-v0.1.0) (2026-03-30)


### Features

* **js:** add forms bindings and DOM enhancement tests ([1cb7fa1](https://github.com/BetterCorp/AnyVali/commit/1cb7fa1493282c5db3207f1e552d86ac3c57b4b0))


### Bug Fixes

* **cli:** treat bare "-" as stdin positional arg, not a flag ([e3fb7ec](https://github.com/BetterCorp/AnyVali/commit/e3fb7ec8869c47c95fbd59f558b7eca77048c221))
* **cpp:** add missing stdexcept include in types.hpp ([ca9b706](https://github.com/BetterCorp/AnyVali/commit/ca9b706c1a1a70eeaa3270e0c578a43a5494d59e))
* **cpp:** reject unsupported schema kinds in string_to_kind ([4948b74](https://github.com/BetterCorp/AnyVali/commit/4948b74de35058e9c9f6ab147a910c62c5eef885))
* resolve build and test failures across all 10 SDKs ([ea46c39](https://github.com/BetterCorp/AnyVali/commit/ea46c39a36b24e6babb1d78013005463badda03e))
* update CI for Node 22/24 and fix CLI test runner ([43938a4](https://github.com/BetterCorp/AnyVali/commit/43938a4fb8f2e4867784b6729b71581e9e0dc8bf))
* update Dockerfiles to match current project structure ([2ac8c9f](https://github.com/BetterCorp/AnyVali/commit/2ac8c9ffb99fb67908618c3388a0c8861d5201ff))

## [Unreleased] - 2026-03-30

### Features
- feat(js): add forms bindings and DOM enhancement tests (1cb7fa1)

### Bug Fixes
- fix: resolve build and test failures across all 10 SDKs (ea46c39)
- fix: update Dockerfiles to match current project structure (2ac8c9f)

### Chores
- chore: add build_test to gitignore (9580c50)
- chore: configure package publishing and README distribution (ab04b17)
- chore: update docs, JS/Java SDK enhancements, and release tooling (5fde263)
- chore: update GitHub references to BetterCorp/AnyVali (24e7643)



All notable changes to this repository will be documented in this file.

This project uses [Release Please](https://github.com/googleapis/release-please) and conventional commits to manage release PRs, version bumps, and changelog generation.

## Unreleased

- Initial repository setup for AnyVali multi-SDK development.
