# Dockerized Release Builds

Build all local release artifacts into `dist-release/`:

```powershell
pwsh -File tools/release/build_release.ps1
```

What it produces:
- `npm/` for npm
- `pypi/` for PyPI
- `crates/` for crates.io
- `maven/java/` for the Java JAR bundle
- `maven/kotlin/` for the Kotlin JAR bundle
- `nuget/` for NuGet
- `rubygems/` for RubyGems
- `packagist/` source archive for review
- `go/` source archive for review
- `cpp/` source archive for review

Notes:
- Java and Kotlin artifacts are local build outputs only; Maven Central publishing still needs signing and Central-specific metadata.
- Go and Packagist publish from a tagged repository rather than from a package upload.
