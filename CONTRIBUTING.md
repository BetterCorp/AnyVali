# Contributing

## Ground Rules

AnyVali is a portability-first project. Contributions are expected to preserve deterministic behavior across SDKs and across schema import/export boundaries.

If a change is only correct in one language, it is usually incomplete.

## Before You Start

- Open an issue for large features, portability changes, or public API changes.
- Keep pull requests focused. Avoid mixing refactors, feature work, and unrelated formatting.
- Prefer spec-first changes for any behavior that affects multiple SDKs.

## Required Conditions For Contributions

All contributions must satisfy these conditions:

1. Behavior changes must be reflected in the spec, docs, or corpus when applicable.
2. New features must include tests.
3. Bug fixes must include a regression test when practical.
4. Portable behavior must not silently diverge between SDKs.
5. New non-portable behavior must be explicitly documented as local-only or extension-based.
6. Breaking changes must be called out clearly in the pull request.
7. Generated artifacts, local caches, and dependency folders should not be committed unless intentionally tracked.

## Commit Style

This repository uses conventional commits for release automation and changelog generation.

Examples:

```text
feat(js): add record schema import
fix(python): validate default before returning parsed output
docs: tighten portability guide wording
chore(ci): add kotlin and cpp jobs
```

Use scopes when the change is specific to one SDK or subsystem.

## Development Workflow

Use the root runner where possible:

```bash
./runner.sh install
./runner.sh build
./runner.sh test
./runner.sh ci
```

Run a single SDK if needed:

```bash
./runner.sh test go
./runner.sh build js
```

## Testing Expectations

At minimum, contributors should run the relevant local test suite for the code they changed.

If a change affects the portable format, conformance corpus, parse semantics, or import/export behavior, run all impacted SDK suites before submitting.

## Pull Request Checklist

- The change is scoped and explained clearly.
- Tests were added or updated.
- Relevant docs were updated.
- Public behavior changes are called out.
- Commit messages follow conventional commits.
- CI is expected to pass on supported platforms.

## Coding Expectations

- Match existing style in the touched SDK.
- Keep APIs small and explicit.
- Prefer declarative, portable semantics over clever runtime-specific behavior.
- Do not introduce hidden coercions or lossy numeric conversions in portable flows.

## License

By contributing, you agree that your contributions will be licensed under the MIT License in this repository.
