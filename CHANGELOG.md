# Changelog

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
